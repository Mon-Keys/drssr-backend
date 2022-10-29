package usecase

import (
	"context"
	"crypto/sha1"
	"database/sql"
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
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type ILooksUsecase interface {
	AddLook(ctx context.Context, look models.Look) (models.Look, int, error)
	UpdateLook(ctx context.Context, newLook models.Look, lid uint64, uid uint64) (models.Look, int, error)
	GetLookByID(ctx context.Context, lid uint64, uid uint64) (models.Look, int, error)
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
	look.ImgPath = filePath

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

func (lu *looksUsecase) UpdateLook(
	ctx context.Context,
	newLook models.Look,
	lid uint64,
	uid uint64,
) (models.Look, int, error) {
	// checking look in db
	foundingLook, err := lu.psql.GetLookByID(ctx, lid)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Look{},
				http.StatusNotFound,
				fmt.Errorf("LooksUsecase.UpdateLook: look not found")
		}
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to found look in db")
	}

	if uid != foundingLook.CreatorID {
		return models.Look{},
			http.StatusForbidden,
			fmt.Errorf("LooksUsecase.UpdateLook: can't update not user's look")
	}

	// checking file ext
	splitedFilename := strings.Split(newLook.Filename, ".")
	ext := fmt.Sprintf(".%s", splitedFilename[len(splitedFilename)-1])
	if !common.IsEnabledExt(ext) {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: not enabled file extension")
	}

	// decoding base64 img into []byte
	decodedImg, err := base64.StdEncoding.DecodeString(newLook.Img)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to decode base64 into byte slice: %w", err)
	}

	today := time.Now().Format("02-24-2006")
	folderNameByte := sha1.New().Sum([]byte(today))
	folderName := hex.EncodeToString(folderNameByte)

	folderPath := fmt.Sprintf("%s/%s", consts.LooksBaseFolderPath, folderName)
	filePath := fmt.Sprintf("%s/%s/%s", consts.LooksBaseFolderPath, folderName, newLook.Filename)
	newLook.ImgPath = filePath

	err = common.SaveFile(folderPath, filePath, decodedImg)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to save look's file: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() {
		err := os.Remove(filePath)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.UpdateLook: failed to rollback creating of look's file: %w", err)
		}
	})

	updatedLook, err := lu.psql.UpdateLookByID(ctx, lid, newLook)
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to save look in db: %w", err)
	}

	rb.Add(func() {
		_, err := lu.psql.UpdateLookByID(ctx, lid, foundingLook)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.UpdateLook: failed to rollback saving of look in db: %w", err)
		}
	})

	deletedBinds, err := lu.psql.DeleteLookClothesBindsByID(ctx, lid)
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to delete old look's binds: %w", err)
	}

	rb.Add(func() {
		for _, clothes := range deletedBinds {
			_, err := lu.psql.AddLookClothesBind(ctx, clothes, lid)
			if err != nil {
				rb.Run()

				lu.logger.Errorf("LooksUsecase.UpdateLook: failed to rollback deleting of old look's binds: %w", err)
			}
		}
	})

	// set new clothes for look
	for _, clothes := range newLook.Clothes {
		createdBind, err := lu.psql.AddLookClothesBind(ctx, clothes, updatedLook.ID)
		if err != nil {
			rb.Run()

			return models.Look{},
				http.StatusInternalServerError,
				fmt.Errorf("LooksUsecase.UpdateLook: failed to save look-clothes bind in db: %w", err)
		}

		rb.Add(func() {
			err := lu.psql.DeleteLookClothesBind(ctx, createdBind.ID)
			if err != nil {
				lu.logger.Errorf(
					"LooksUsecase.UpdateLook: failed to rollback saving of look-clothes bind in db: %w",
					err,
				)
			}
		})

		clothedFromDB, err := lu.clothesPsql.GetClothesByID(ctx, clothes.ID)
		if err != nil {
			rb.Run()

			return models.Look{},
				http.StatusInternalServerError,
				fmt.Errorf("LooksUsecase.UpdateLook: failed to get clothes from db: %w", err)
		}

		updatedLook.Clothes = append(updatedLook.Clothes, models.ClothesStruct{
			ID:     clothes.ID,
			Label:  fmt.Sprintf("%s %s", clothedFromDB.Type, clothedFromDB.Brand),
			Coords: createdBind.Coords,
		})
	}

	// deleting old img
	err = common.DeleteFile(foundingLook.ImgPath)
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to delete old look's img file: %w", err)
	}

	updatedLook.Img = newLook.Img

	return updatedLook, http.StatusOK, nil
}

func (lu *looksUsecase) GetLookByID(ctx context.Context, lid uint64, uid uint64) (models.Look, int, error) {
	// checking look in db
	foundingLook, err := lu.psql.GetLookByID(ctx, lid)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.Look{},
				http.StatusNotFound,
				fmt.Errorf("LooksUsecase.GetLookByID: look not found")
		}
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to found look in db")
	}

	if uid != foundingLook.CreatorID {
		return models.Look{},
			http.StatusForbidden,
			fmt.Errorf("LooksUsecase.GetLookByID: can't update not user's look")
	}

	clothes, err := lu.psql.GetLookClothes(ctx, lid)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to get clothes from db")
	}

	foundingLook.Clothes = append(foundingLook.Clothes, clothes...)

	f, err := os.Open(foundingLook.ImgPath)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to open look's img file")
	}

	bytesImg, err := ioutil.ReadAll(f)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to read img file")
	}

	foundingLook.Img = base64.StdEncoding.EncodeToString(bytesImg)

	return foundingLook, http.StatusOK, nil
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
