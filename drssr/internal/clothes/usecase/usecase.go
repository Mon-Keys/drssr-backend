package usecase

import (
	"context"
	"crypto/sha1"
	"drssr/config"
	"drssr/internal/clothes/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/classifier"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/cutter"
	"drssr/internal/pkg/file_utils"
	"drssr/internal/pkg/rollback"
	"drssr/internal/pkg/similarity"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type IClothesUsecase interface {
	AddFile(ctx context.Context, args AddFileArgs) (models.Clothes, int, error)
	UpdateClothes(ctx context.Context, uid uint64, newClothersData models.Clothes) (models.Clothes, int, error)
	DeleteClothes(ctx context.Context, uid uint64, cid uint64) (int, error)
	GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error)
	GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error)
	GetClothesByID(ctx context.Context, cid uint64) (models.Clothes, int, error)
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
	UserEmail  string
	FileHeader *multipart.FileHeader
	File       multipart.File
}

// when uploading file
func (cu *clothesUsecase) AddFile(
	ctx context.Context,
	args AddFileArgs,
) (models.Clothes, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	buf := make([]byte, args.FileHeader.Size)
	_, err := args.File.Read(buf)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to read file: %w", err)
	}

	fileType := http.DetectContentType(buf)
	if !file_utils.IsEnabledFileType(fileType) {
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

	folderNameByte := sha1.New().Sum([]byte(args.UserEmail))
	folderName := fmt.Sprintf(hex.EncodeToString(folderNameByte))

	// saving clothes file
	clothesFileName := file_utils.GenerateFileName("clothes", consts.FileExt)
	clothesFolderPath := fmt.Sprintf("%s/%s", consts.ClothesBaseFolderPath, folderName)
	clothesFilePath := fmt.Sprintf("%s/%s/%s", consts.ClothesBaseFolderPath, folderName, clothesFileName)

	err = file_utils.SaveBase64ToFile(clothesFolderPath, clothesFilePath, res.Img)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to save clothes's file: %w", err)
	}

	rb.Add(func() {
		err := os.Remove(clothesFilePath)
		if err != nil {
			cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback creating of clothes's file: %w", err)
		}
	})

	// saving masks file
	maskFileName := file_utils.GenerateFileName("mask", consts.FileExt)
	masksFolderPath := fmt.Sprintf("%s/%s", consts.MasksBaseFolderPath, folderName)
	masksFilePath := fmt.Sprintf("%s/%s/%s", consts.MasksBaseFolderPath, folderName, maskFileName)

	err = file_utils.SaveBase64ToFile(masksFolderPath, masksFilePath, res.Mask)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to save mask's file: %w", err)
	}

	rb.Add(func() {
		err := os.Remove(masksFilePath)
		if err != nil {
			cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback creating of mask's file: %w", err)
		}
	})

	clothesType, err := cu.classifierClient.RecognizePhoto(ctx, buf)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to determine type of clothes: %w", err)
	}

	createdClothes, err := cu.psql.AddClothes(ctx, models.Clothes{
		OwnerID:  args.UID,
		Type:     clothesType,
		ImgPath:  clothesFilePath,
		MaskPath: masksFilePath,
	})
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to save clothes in db: %w", err)
	}

	rb.Add(func() {
		err := cu.psql.DeleteClothes(ctx, createdClothes.ID)
		if err != nil {
			cu.logger.Errorf("ClothesUsecase.AddFile: failed to rollback adding of clothes: %w", err)
		}
	})

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
		if err == pgx.ErrNoRows {
			return models.Clothes{},
				http.StatusNotFound,
				fmt.Errorf("ClothesUsecase.UpdateClothes: failed to found clothes for update: %w", err)
		}
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

	return clothes, http.StatusOK, nil
}

func (cu *clothesUsecase) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error) {
	clothes, err := cu.psql.GetUsersClothes(ctx, limit, offset, uid)
	if err != nil {
		return nil,
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.GetUsersClothes: failed to get clothes from db: %w", err)
	}

	return clothes, http.StatusOK, nil
}

func (cu *clothesUsecase) GetClothesByID(ctx context.Context, cid uint64) (models.Clothes, int, error) {
	clothes, err := cu.psql.GetClothesByID(ctx, cid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Clothes{},
				http.StatusNotFound,
				fmt.Errorf("ClothesUsecase.GetClothesByID: failed to found clothes for update: %w", err)
		}
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.GetClothesByID: failed to get clothes form db: %w", err)
	}

	return clothes, http.StatusOK, nil
}
