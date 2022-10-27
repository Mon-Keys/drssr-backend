package usecase

import (
	"context"
	"crypto/sha1"
	"drssr/internal/looks/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/classifier"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/cutter"
	"drssr/internal/pkg/rollback"
	"drssr/internal/pkg/similarity"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type ILooksUsecase interface {
	AddLook(ctx context.Context, args AddLookArgs) (models.Look, int, error)
	// GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error)
	// GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error)
}

type looksUsecase struct {
	psql             repository.IPostgresqlRepository
	cutterClient     cutter.Client
	classifierClient classifier.RecognizeAPIClient
	similarityClient similarity.Client
	logger           logrus.Logger
}

func NewLooksUsecase(
	pr repository.IPostgresqlRepository,
	logger logrus.Logger,
) ILooksUsecase {
	return &looksUsecase{
		psql:   pr,
		logger: logger,
	}
}

type AddLookArgs struct {
	UID               uint64
	FileHeader        *multipart.FileHeader
	File              multipart.File
	PreviewFileHeader *multipart.FileHeader
	PreviewFile       multipart.File
	Clothes           []uint64
	Description       string
}

func (lu *looksUsecase) AddLook(
	ctx context.Context,
	args AddLookArgs,
) (models.Look, int, error) {
	fileBuf := make([]byte, args.FileHeader.Size)
	_, err := args.File.Read(fileBuf)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to read look's file: %w", err)
	}

	fileType := http.DetectContentType(fileBuf)
	if !common.IsEnabledFileType(fileType) {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: not enabled look's file type")
	}

	filePreviewBuf := make([]byte, args.PreviewFileHeader.Size)
	_, err = args.PreviewFile.Read(filePreviewBuf)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to read preview's file: %w", err)
	}

	filePreviewType := http.DetectContentType(filePreviewBuf)
	if !common.IsEnabledFileType(filePreviewType) {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: not enabled preview's file type")
	}

	today := time.Now().Format("02-24-2006")
	folderNameByte := sha1.New().Sum([]byte(today))
	folderName := hex.EncodeToString(folderNameByte)

	filePath := fmt.Sprintf("%s/%s%s", consts.LooksBaseFolderPath, folderName, args.FileHeader.Filename)
	filePreviewPath := fmt.Sprintf("%s/%s%s", consts.LooksBaseFolderPath, folderName, args.PreviewFileHeader.Filename)

	err = common.SaveFile(filePath, fileBuf)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to save look's file: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() {
		err := os.Remove(filePath)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.AddLook: failed to rollback creating of look's file: %w", err)
		}
	})

	err = common.SaveFile(filePreviewPath, filePreviewBuf)
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to save preview's file: %w", err)
	}

	rb.Add(func() {
		err := os.Remove(filePreviewPath)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.AddLook: failed to rollback creating of preview's file: %w", err)
		}
	})

	createdLook, err := lu.psql.AddLook(ctx, models.Look{
		PreviewPath: filePreviewPath,
		ImgPath:     filePath,
		Description: args.Description,
		CreatorID:   args.UID,
	})
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to save look in db: %w", err)
	}

	rb.Add(func() {
		err := lu.psql.DeleteLook(ctx, createdLook.ID)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.AddLook: failed to rollback saving of look in db: %w", err)
		}
	})

	for _, clothesID := range args.Clothes {
		createdBindID, err := lu.psql.AddLookClothesBind(ctx, createdLook.ID, clothesID)
		if err != nil {
			rb.Run()

			return models.Look{},
				http.StatusInternalServerError,
				fmt.Errorf("LooksUsecase.AddLook: failed to save look-clothes bind in db: %w", err)
		}

		rb.Add(func() {
			err := lu.psql.DeleteLookClothesBind(ctx, createdBindID)
			if err != nil {
				lu.logger.Errorf("LooksUsecase.AddLook: failed to rollback saving of look-clothes bind in db: %w", err)
			}
		})
	}

	createdLook.Clothes = args.Clothes
	createdLook.Img = base64.StdEncoding.EncodeToString(fileBuf)
	createdLook.Preview = base64.StdEncoding.EncodeToString(filePreviewBuf)

	return createdLook, http.StatusOK, nil
}

// func (cu *clothesUsecase) GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error) {
// 	clothes, err := cu.psql.GetAllClothes(ctx, limit, offset)
// 	if err != nil {
// 		return nil,
// 			http.StatusInternalServerError,
// 			err
// 	}

// 	for i, v := range clothes {
// 		img, err := ioutil.ReadFile(v.ImgPath)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		clothes[i].Img = base64.StdEncoding.EncodeToString(img)

// 		mask, err := ioutil.ReadFile(v.MaskPath)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)

// 	}

// 	return clothes, http.StatusOK, nil
// }

// func (cu *clothesUsecase) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error) {
// 	clothes, err := cu.psql.GetUsersClothes(ctx, limit, offset, uid)
// 	if err != nil {
// 		return nil,
// 			http.StatusInternalServerError,
// 			err
// 	}

// 	for i, v := range clothes {
// 		img, err := ioutil.ReadFile(v.ImgPath)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		clothes[i].Img = base64.StdEncoding.EncodeToString(img)

// 		mask, err := ioutil.ReadFile(v.MaskPath)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		clothes[i].Mask = base64.StdEncoding.EncodeToString(mask)

// 	}

// 	return clothes, http.StatusOK, nil
// }

// func isEnabledFileType(fileType string) bool {
// 	imgTypes := map[string]bool{
// 		"image/jpg":  true,
// 		"image/jpeg": true,
// 		"image/png":  true,
// 		"image/webp": true,
// 	}

// 	return imgTypes[fileType]
// }

// func isEnabledExt(fileType string) bool {
// 	imgTypes := map[string]bool{
// 		".jpg":  true,
// 		".jpeg": true,
// 		".png":  true,
// 	}

// 	return imgTypes[fileType]
// }
