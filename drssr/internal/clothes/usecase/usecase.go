package usecase

import (
	"context"
	"drssr/config"
	"drssr/internal/clothes/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/classifier"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/cutter"
	"drssr/internal/pkg/rollback"
	"drssr/internal/pkg/similarity"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

type IClothesUsecase interface {
	AddFile(ctx context.Context, args AddFileArgs) (models.Clothes, int, error)
	UpdateClothes(ctx context.Context, uid uint64, newClothersData models.Clothes) (models.Clothes, int, error)
	DeleteClothes(ctx context.Context, uid uint64, cid uint64) (int, error)
	GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error)
	GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error)
}

type clothesUsecase struct {
	psql             repository.IPostgresqlRepository
	cutterClient     cutter.Client
	classifierClient classifier.RecognizeAPIClient
	similarityClient similarity.Client
	logger           logrus.Logger
}

func NewClothesUsecase(
	pr repository.IPostgresqlRepository,
	cc cutter.Client,
	cfc classifier.RecognizeAPIClient,
	sc similarity.Client,
	logger logrus.Logger,
) IClothesUsecase {
	return &clothesUsecase{
		psql:             pr,
		cutterClient:     cc,
		classifierClient: cfc,
		similarityClient: sc,
		logger:           logger,
	}
}

type AddFileArgs struct {
	UID        uint64
	FileHeader *multipart.FileHeader
	File       multipart.File
}

// when uploading file
func (cu *clothesUsecase) AddFile(
	ctx context.Context,
	args AddFileArgs,
) (models.Clothes, int, error) {
	buf := make([]byte, args.FileHeader.Size)
	_, err := args.File.Read(buf)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to read file: %w", err)
	}

	fileType := http.DetectContentType(buf)
	if !common.IsEnabledFileType(fileType) {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: not enabled file type")
	}

	res, err := cu.cutterClient.UploadImg(ctx, &cutter.UploadImgArgs{
		FileHeader: *args.FileHeader,
		File:       buf,
	})
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to upload img in cutter: %w", err)
	}
	// TODO: add rollback for cutter

	clothesType, err := cu.classifierClient.RecognizePhoto(ctx, buf)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to determine type of clothes: %w", err)
	}

	createdClothes, err := cu.psql.AddClothes(ctx, models.Clothes{
		OwnerID:  args.UID,
		Type:     clothesType,
		ImgPath:  res.ImgPath,
		MaskPath: res.MaskPath,
	})
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to save clothes in db: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() {
		err := cu.psql.DeleteClothes(ctx, createdClothes.ID)
		if err != nil {
			cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback adding of clothes: %w", err)
		}
	})

	// TODO: change this hack
	createdClothes.ImgPath = strings.ReplaceAll(consts.HomeDirectory, createdClothes.ImgPath, "")
	createdClothes.MaskPath = strings.ReplaceAll(consts.HomeDirectory, createdClothes.MaskPath, "")

	// TODO: delete after testing
	// createdClothes.Img = res.Img
	// createdClothes.Mask = res.Mask

	return createdClothes, http.StatusOK, nil
}

// after click save button
func (cu *clothesUsecase) UpdateClothes(
	ctx context.Context,
	uid uint64,
	newClothersData models.Clothes,
) (models.Clothes, int, error) {
	clothes, err := cu.psql.GetClothesByID(ctx, newClothersData.ID)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.UpdateClothes: failed to get clothes form db: %w", err)
	}

	if clothes.OwnerID != uid {
		return models.Clothes{},
			http.StatusForbidden,
			fmt.Errorf("ClothesUsecase.DeleUpdateClothesteClothes: user try to update not his own clothes: %w", err)
	}

	updatedClothes, err := cu.psql.UpdateClothes(ctx, newClothersData)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.UpdateClothes: failed to update clothes in db: %w", err)
	}

	// TODO: change this hack
	updatedClothes.ImgPath = strings.ReplaceAll(consts.HomeDirectory, updatedClothes.ImgPath, "")
	updatedClothes.MaskPath = strings.ReplaceAll(consts.HomeDirectory, updatedClothes.MaskPath, "")

	// TODO: delete after testing
	// img, err := ioutil.ReadFile(updatedClothes.ImgPath)
	// if err != nil {
	// 	return models.Clothes{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("ClothesUsecase.UpdateClothes: failed to read img file: %w", err)
	// }

	// mask, err := ioutil.ReadFile(updatedClothes.MaskPath)
	// if err != nil {
	// 	return models.Clothes{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("ClothesUsecase.UpdateClothes: failed to read mask file: %w", err)
	// }

	// updatedClothes.Img = base64.StdEncoding.EncodeToString(img)
	// updatedClothes.Mask = base64.StdEncoding.EncodeToString(mask)

	// processing similarity
	go cu.processingSimilarity(ctx, updatedClothes)

	return updatedClothes, http.StatusOK, nil
}

func (cu *clothesUsecase) processingSimilarity(ctx context.Context, clothes models.Clothes) {
	ctx, rb := rollback.NewCtxRollback(ctx)
	logger := cu.logger.WithContext(ctx)

	clothesWithSameType, err := cu.psql.GetClothesMaskByTypeAndSex(ctx, clothes.Type, clothes.Sex)
	if err != nil {
		rb.Run()

		logger.Errorf("ClothesUsecase.processingSimilarity: failed to get clothes with same type: %w", err)
		return
	}

	clothesWithSameTypeMap := make(map[uint64]string, len(clothesWithSameType)-1)
	for _, clothes := range clothesWithSameType {
		if clothes.ID != clothes.ID {
			clothesWithSameTypeMap[clothes.ID] = clothes.MaskPath
		}
	}
	if len(clothesWithSameTypeMap) == 0 {
		logger.Infof("ClothesUsecase.processingSimilarity: clothes haven't similar clothes by type and sex: %w", err)
		return
	}

	// we compare images by mask
	similarityRes, err := cu.similarityClient.CheckSimilarity(ctx, &similarity.CheckSimilarityArgs{
		CheckedImage:   clothes.MaskPath,
		CheckingImages: clothesWithSameTypeMap,
	})
	if err != nil {
		rb.Run()

		logger.Errorf("ClothesUsecase.processingSimilarity: failed to check similarity: %w", err)
		return
	}

	for key, value := range similarityRes.Similarity {
		if value >= config.WellSimilarityPercent {
			similarityBindID, err := cu.psql.AddSimilarityBind(ctx, clothes.ID, key, value)
			if err != nil {
				rb.Run()

				logger.Errorf("ClothesUsecase.processingSimilarity: failed to save similarity bind in db: %w", err)
				return
			}
			rb.Add(func() {
				err := cu.psql.DeleteSimilarityBind(ctx, similarityBindID)
				if err != nil {
					logger.Errorf("ClothesUsecase.processingSimilarity: failed to rollback adding of similarity bind: %w", err)
				}
			})
		}
	}
}

func (cu *clothesUsecase) DeleteClothes(ctx context.Context, uid uint64, cid uint64) (int, error) {
	clothes, err := cu.psql.GetClothesByID(ctx, cid)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ClothesUsecase.DeleteClothes: failed to get clothes form db: %w", err)
	}

	if clothes.OwnerID != uid {
		return http.StatusForbidden, fmt.Errorf("ClothesUsecase.DeleteClothes: user try to delete not his own clothes: %w", err)
	}

	err = cu.psql.DeleteClothes(ctx, cid)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ClothesUsecase.DeleteClothes: failed to delete clothes: %w", err)
	}

	// TODO: need to iterate by every bind in go and delete it to do rollback
	// very dangerous now
	err = cu.psql.DeleteSimilarityBindByClothesID(ctx, cid)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("ClothesUsecase.DeleteClothes: failed to delete similarity binds: %w", err)
	}

	return http.StatusOK, nil
}

func (cu *clothesUsecase) GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error) {
	clothes, err := cu.psql.GetAllClothes(ctx, limit, offset)
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.GetAllClothes: failed to get clothes from db: %w", err)
	}

	// TODO: change this hack
	for i := range clothes {
		clothes[i].ImgPath = strings.ReplaceAll(consts.HomeDirectory, clothes[i].ImgPath, "")
		clothes[i].MaskPath = strings.ReplaceAll(consts.HomeDirectory, clothes[i].MaskPath, "")
	}

	// TODO: delete after testing
	// for i, v := range clothes {
	// 	img, err := ioutil.ReadFile(v.ImgPath)
	// 	if err != nil {
	// 		return nil,
	// 			http.StatusInternalServerError,
	// 			fmt.Errorf("ClothesUsecase.GetAllClothes: failed to open img %s file: %w", v.ImgPath, err)
	// 	}

	// 	clothes[i].Img = base64.StdEncoding.EncodeToString(img)

	// 	mask, err := ioutil.ReadFile(v.MaskPath)
	// 	if err != nil {
	// 		return nil,
	// 			http.StatusInternalServerError,
	// 			fmt.Errorf("ClothesUsecase.GetAllClothes: failed to open mask %s file: %w", v.MaskPath, err)
	// 	}

	// 	clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)
	// }

	return clothes, http.StatusOK, nil
}

func (cu *clothesUsecase) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error) {
	clothes, err := cu.psql.GetUsersClothes(ctx, limit, offset, uid)
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.GetUsersClothes: failed to get clothes from db: %w", err)
	}

	// TODO: change this hack
	for i := range clothes {
		clothes[i].ImgPath = strings.ReplaceAll(consts.HomeDirectory, clothes[i].ImgPath, "")
		clothes[i].MaskPath = strings.ReplaceAll(consts.HomeDirectory, clothes[i].MaskPath, "")
	}

	// for i, v := range clothes {
	// 	img, err := ioutil.ReadFile(v.ImgPath)
	// 	if err != nil {
	// 		return nil,
	// 			http.StatusInternalServerError,
	// 			fmt.Errorf("ClothesUsecase.GetUsersClothes: failed to open img %s file: %w", v.ImgPath, err)
	// 	}

	// 	clothes[i].Img = base64.StdEncoding.EncodeToString(img)

	// 	mask, err := ioutil.ReadFile(v.MaskPath)
	// 	if err != nil {
	// 		return nil,
	// 			http.StatusInternalServerError,
	// 			fmt.Errorf("ClothesUsecase.GetUsersClothes: failed to open mask %s file: %w", v.MaskPath, err)
	// 	}

	// 	clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)
	// }

	return clothes, http.StatusOK, nil
}
