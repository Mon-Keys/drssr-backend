package usecase

import (
	"context"
	"drssr/config"
	"drssr/internal/clothes/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/classifier"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/cutter"
	"drssr/internal/pkg/rollback"
	"drssr/internal/pkg/similarity"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/sirupsen/logrus"
)

type IClothesUsecase interface {
	AddFile(ctx context.Context, args AddFileArgs) (models.Clothes, int, error)
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
	UID          uint64
	FileHeader   *multipart.FileHeader
	File         multipart.File
	ClothesBrand string
	ClothesSex   string
}

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
		Type:     clothesType,
		Color:    "rerwf",
		ImgPath:  res.ImgPath,
		MaskPath: res.MaskPath,
		Brand:    args.ClothesBrand,
		Sex:      args.ClothesSex,
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

	bindID, err := cu.psql.AddClothesUserBind(ctx, args.UID, createdClothes.ID)
	if err != nil {
		rb.Run()

		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to create clothes-user bind in db: %w", err)
	}
	rb.Add(func() {
		err := cu.psql.DeleteClothesUserBind(ctx, bindID)
		if err != nil {
			cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback adding of clothes-user bind: %w", err)
		}
	})

	createdClothes.Img = res.Img
	createdClothes.Mask = res.Mask

	clothesWithSameType, err := cu.psql.GetClothesMaskByTypeAndSex(ctx, createdClothes.Type, createdClothes.Sex)
	if err != nil {
		rb.Run()

		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to get clothes with same type: %w", err)
	}

	clothesWithSameTypeMap := make(map[uint64]string, len(clothesWithSameType)-1)
	for _, clothes := range clothesWithSameType {
		if clothes.ID != createdClothes.ID {
			clothesWithSameTypeMap[clothes.ID] = clothes.MaskPath
		}
	}
	if len(clothesWithSameTypeMap) == 0 {
		return createdClothes, http.StatusOK, nil
	}

	// we compare images by mask
	similarityRes, err := cu.similarityClient.CheckSimilarity(ctx, &similarity.CheckSimilarityArgs{
		CheckedImage:   createdClothes.MaskPath,
		CheckingImages: clothesWithSameTypeMap,
	})
	if err != nil {
		rb.Run()

		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to check similarity: %w", err)
	}

	for key, value := range similarityRes.Similarity {
		if value >= config.WellSimilarityPercent {
			similarityBindID, err := cu.psql.AddSimilarityBind(ctx, createdClothes.ID, key, value)
			if err != nil {
				rb.Run()

				return models.Clothes{},
					http.StatusInternalServerError,
					fmt.Errorf("ClothesUsecase.AddFile: failed to save similarity bind in db: %w", err)
			}
			rb.Add(func() {
				err := cu.psql.DeleteSimilarityBind(ctx, similarityBindID)
				if err != nil {
					cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback adding of similarity bind: %w", err)
				}
			})
		}
	}

	return createdClothes, http.StatusOK, nil
}

func (cu *clothesUsecase) GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error) {
	clothes, err := cu.psql.GetAllClothes(ctx, limit, offset)
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			err
	}

	for i, v := range clothes {
		img, err := ioutil.ReadFile(v.ImgPath)
		if err != nil {
			log.Fatal(err)
		}

		clothes[i].Img = base64.StdEncoding.EncodeToString(img)

		mask, err := ioutil.ReadFile(v.MaskPath)
		if err != nil {
			log.Fatal(err)
		}
		clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)

	}

	return clothes, http.StatusOK, nil
}

func (cu *clothesUsecase) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error) {
	clothes, err := cu.psql.GetUsersClothes(ctx, limit, offset, uid)
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			err
	}

	for i, v := range clothes {
		img, err := ioutil.ReadFile(v.ImgPath)
		if err != nil {
			log.Fatal(err)
		}

		clothes[i].Img = base64.StdEncoding.EncodeToString(img)

		mask, err := ioutil.ReadFile(v.MaskPath)
		if err != nil {
			log.Fatal(err)
		}
		clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)

	}

	return clothes, http.StatusOK, nil
}
