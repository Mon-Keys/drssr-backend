package usecase

import (
	"context"
	"crypto/sha1"
	clothes_repository "drssr/internal/clothes/repository"
	"drssr/internal/looks/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/rollback"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type ILooksUsecase interface {
	AddLook(ctx context.Context, look models.Look) (models.Look, int, error)
	UpdateLook(ctx context.Context, newLook models.Look, lid uint64, uid uint64) (models.Look, int, error)
	DeleteLook(ctx context.Context, uid uint64, lid uint64) (int, error)

	GetLookByID(ctx context.Context, lid uint64) (models.Look, int, error)
	GetUserLooks(ctx context.Context, uid uint64, limit int, offset int) (models.ArrayLooks, int, error)
	GetAllLooks(ctx context.Context, limit int, offset int) (models.ArrayLooks, int, error)
}

type looksUsecase struct {
	psql        repository.IPostgresqlRepository
	clothesPsql clothes_repository.IPostgresqlRepository
	logger      logrus.Logger
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

	today := time.Now().Format("2006-01-02")
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
		_, err := lu.psql.DeleteLook(ctx, createdLook.ID)
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
			ID:       clothes.ID,
			Label:    fmt.Sprintf("%s %s", clothedFromDB.Type, clothedFromDB.Brand),
			Coords:   createdBind.Coords,
			ImgPath:  clothedFromDB.ImgPath,
			MaskPath: clothedFromDB.MaskPath,
		})
	}

	// createdLook.Img = look.Img

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
		if err == pgx.ErrNoRows {
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

	today := time.Now().Format("2006-01-02")
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
			ID:       clothes.ID,
			Label:    fmt.Sprintf("%s %s", clothedFromDB.Type, clothedFromDB.Brand),
			Coords:   createdBind.Coords,
			ImgPath:  clothedFromDB.ImgPath,
			MaskPath: clothedFromDB.MaskPath,
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

	// updatedLook.Img = newLook.Img

	return updatedLook, http.StatusOK, nil
}

func (lu *looksUsecase) DeleteLook(ctx context.Context, uid uint64, lid uint64) (int, error) {
	// checking look in db
	foundingLook, err := lu.psql.GetLookByID(ctx, lid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, fmt.Errorf("LooksUsecase.DeleteLook: look not found")
		}
		return http.StatusInternalServerError, fmt.Errorf("LooksUsecase.DeleteLook: failed to found look in db")
	}

	if uid != foundingLook.CreatorID {
		return http.StatusForbidden, fmt.Errorf("LooksUsecase.DeleteLook: can't delete not user's look")
	}

	// deleting look-clothes binds
	deletedBinds, err := lu.psql.DeleteLookClothesBindsByID(ctx, lid)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("LooksUsecase.DeleteLook: failed to delete old look's binds: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)

	rb.Add(func() {
		for _, clothes := range deletedBinds {
			_, err := lu.psql.AddLookClothesBind(ctx, clothes, lid)
			if err != nil {
				rb.Run()

				lu.logger.Errorf("LooksUsecase.DeleteLook: failed to rollback deleting of old look's binds: %w", err)
			}
		}
	})

	// deleting look
	deletedLook, err := lu.psql.DeleteLook(ctx, lid)
	if err != nil {
		rb.Run()

		return http.StatusInternalServerError, fmt.Errorf("LooksUsecase.DeleteLook: failed to delete look from db: %w", err)
	}

	rb.Add(func() {
		_, err := lu.psql.AddLook(ctx, deletedLook)
		if err != nil {
			lu.logger.Errorf("LooksUsecase.DeleteLook: failed to rollback deleting of look from db: %w", err)
		}
	})

	return http.StatusOK, nil
}

func (lu *looksUsecase) GetUserLooks(ctx context.Context, uid uint64, limit int, offset int) (models.ArrayLooks, int, error) {
	looks, err := lu.psql.GetUserLooks(ctx, limit, offset, uid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, http.StatusNotFound, fmt.Errorf("LooksUsecase.GetUserLooks: user don't have any looks")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetUserLooks: failed to get user looks from db: %w", err)
	}

	for i := range looks {
		clothes, err := lu.psql.GetLookClothes(ctx, looks[i].ID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetUserLooks: failed to get look's clothes from db: %w", err)
		}

		looks[i].Clothes = clothes

		// decodedLookImg, err := common.ReadFileIntoBase64(looks[i].ImgPath)
		// if err != nil {
		// 	return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetUserLooks: failed to read img file into base64: %w", err)
		// }

		// looks[i].Img = decodedLookImg
	}

	return looks, http.StatusOK, nil
}

func (lu *looksUsecase) GetLookByID(ctx context.Context, lid uint64) (models.Look, int, error) {
	// checking look in db
	foundingLook, err := lu.psql.GetLookByID(ctx, lid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Look{},
				http.StatusNotFound,
				fmt.Errorf("LooksUsecase.GetLookByID: look not found")
		}
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to found look in db: %w", err)
	}

	clothes, err := lu.psql.GetLookClothes(ctx, lid)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.GetLookByID: failed to get clothes from db: %w", err)
	}

	foundingLook.Clothes = clothes

	// decodedLookImg, err := common.ReadFileIntoBase64(foundingLook.ImgPath)
	// if err != nil {
	// 	return models.Look{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("LooksUsecase.GetLookByID: failed to read img file into base64: %w", err)
	// }

	// foundingLook.Img = decodedLookImg

	return foundingLook, http.StatusOK, nil
}

func (lu *looksUsecase) GetAllLooks(ctx context.Context, limit int, offset int) (models.ArrayLooks, int, error) {
	looks, err := lu.psql.GetAllLooks(ctx, limit, offset)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, http.StatusNotFound, fmt.Errorf("LooksUsecase.GetAllLooks: not found any looks")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetAllLooks: failed to get looks from db: %w", err)
	}

	for i := range looks {
		clothes, err := lu.psql.GetLookClothes(ctx, looks[i].ID)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetAllLooks: failed to get look's clothes from db: %w", err)
		}

		looks[i].Clothes = clothes

		// decodedLookImg, err := common.ReadFileIntoBase64(looks[i].ImgPath)
		// if err != nil {
		// 	return nil, http.StatusInternalServerError, fmt.Errorf("LooksUsecase.GetAllLooks: failed to read img file into base64: %w", err)
		// }

		// looks[i].Img = decodedLookImg
	}

	return looks, http.StatusOK, nil
}
