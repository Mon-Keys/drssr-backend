package usecase

import (
	"context"
	"drssr/internal/clothes/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/cutter"
	"drssr/internal/pkg/rollback"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/sirupsen/logrus"
)

type IClothesUsecase interface {
	AddFile(ctx context.Context, uid uint64, fileHeader *multipart.FileHeader, file multipart.File) (models.Clothes, int, error)
}

type clothesUsecase struct {
	psql         repository.IPostgresqlRepository
	cutterClient cutter.Client
	logger       logrus.Logger
}

func NewClothesUsecase(
	pr repository.IPostgresqlRepository,
	cc cutter.Client,
	logger logrus.Logger,
) IClothesUsecase {
	return &clothesUsecase{
		psql:         pr,
		cutterClient: cc,
		logger:       logger,
	}
}

func (cu *clothesUsecase) AddFile(
	ctx context.Context,
	uid uint64,
	fileHeader *multipart.FileHeader,
	file multipart.File,
) (models.Clothes, int, error) {
	buf := make([]byte, fileHeader.Size)
	_, err := file.Read(buf)
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to read file: %w", err)
	}

	fileType := http.DetectContentType(buf)
	if !isEnabledFileType(fileType) {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: not enabled file type")
	}

	res, err := cu.cutterClient.UploadImg(ctx, &cutter.UploadImgArgs{
		FileHeader: *fileHeader,
		File:       buf,
	})
	if err != nil {
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to upload img in cutter: %w", err)
	}
	// TODO: add rollback for cutter

	// TODO: сходить в классификатор и похожесть

	createdClothes, err := cu.psql.AddClothes(ctx, models.Clothes{
		Type:     "хуета",
		Color:    "rerwf",
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

	_, err = cu.psql.AddClothesUserBind(ctx, uid, createdClothes.ID)
	if err != nil {
		rb.Run()

		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("ClothesUsecase.AddFile: failed to create clothes-user bind in db: %w", err)
	}

	createdClothes.Img = res.Img
	createdClothes.Mask = res.Mask

	return createdClothes, http.StatusOK, nil
}

func isEnabledFileType(fileType string) bool {
	imgTypes := map[string]bool{
		"image/jpg":  true,
		"image/jpeg": true,
		"image/png":  true,
	}
	if imgTypes[fileType] {
		return true
	}
	return false
}

func isEnabledExt(fileType string) bool {
	imgTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	if imgTypes[fileType] {
		return true
	}
	return false
}
