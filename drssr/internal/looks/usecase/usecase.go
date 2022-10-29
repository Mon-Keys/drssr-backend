package usecase

import (
	"context"
	"crypto/sha1"
	clothes_repository "drssr/internal/clothes/repository"
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
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ILooksUsecase interface {
	AddLook(ctx context.Context, look models.Look) (models.Look, int, error)
	// GetAllClothes(ctx context.Context, limit, offset int) (models.ArrayClothes, int, error)
	// GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) (models.ArrayClothes, int, error)
}

type looksUsecase struct {
	psql             repository.IPostgresqlRepository
	clothesPsql      clothes_repository.IPostgresqlRepository
	cutterClient     cutter.Client
	classifierClient classifier.RecognizeAPIClient
	similarityClient similarity.Client
	logger           logrus.Logger
}

func NewLooksUsecase(
	pr repository.IPostgresqlRepository,
	cpr clothes_repository.IPostgresqlRepository,
	logger logrus.Logger,
) ILooksUsecase {
	return &looksUsecase{
		psql:        pr,
		clothesPsql: cpr,
		logger:      logger,
	}
}

func (lu *looksUsecase) AddLook(
	ctx context.Context,
	look models.Look,
) (models.Look, int, error) {
	// checking file ext
	splitedFilename := strings.Split(look.Filename, ".")
	ext := fmt.Sprintf(".%s", splitedFilename[len(splitedFilename)-1])
	if !common.IsEnabledExt(ext) {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: not enabled file extension")
	}

	// decoding base64 img into []byte
	decodedImg, err := base64.StdEncoding.DecodeString(look.Img)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to decode base64 into byte slice: %w", err)
	}

	today := time.Now().Format("02-24-2006")
	folderNameByte := sha1.New().Sum([]byte(today))
	folderName := hex.EncodeToString(folderNameByte)

	folderPath := fmt.Sprintf("%s/%s", consts.LooksBaseFolderPath, folderName)
	filePath := fmt.Sprintf("%s/%s/%s", consts.LooksBaseFolderPath, folderName, look.Filename)

	err = common.SaveFile(folderPath, filePath, decodedImg)
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

	createdLook, err := lu.psql.AddLook(ctx, look)
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

	for _, clothes := range look.Clothes {
		createdBind, err := lu.psql.AddLookClothesBind(ctx, clothes, createdLook.ID)
		if err != nil {
			rb.Run()

			return models.Look{},
				http.StatusInternalServerError,
				fmt.Errorf("LooksUsecase.AddLook: failed to save look-clothes bind in db: %w", err)
		}

		rb.Add(func() {
			err := lu.psql.DeleteLookClothesBind(ctx, createdBind.ID)
			if err != nil {
				lu.logger.Errorf("LooksUsecase.AddLook: failed to rollback saving of look-clothes bind in db: %w", err)
			}
		})

		clothedFromDB, err := lu.clothesPsql.GetClothesByID(ctx, clothes.ID)
		if err != nil {
			rb.Run()

			return models.Look{},
				http.StatusInternalServerError,
				fmt.Errorf("LooksUsecase.AddLook: failed to get clothes from db: %w", err)
		}

		createdLook.Clothes = append(createdLook.Clothes, models.ClothesStruct{
			ID:     clothes.ID,
			Label:  fmt.Sprintf("%s %s", clothedFromDB.Type, clothedFromDB.Brand),
			Coords: createdBind.Coords,
		})
	}

	createdLook.Img = look.Img

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
