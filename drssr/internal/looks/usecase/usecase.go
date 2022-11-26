package usecase

import (
	"context"
	"crypto/sha1"
	clothes_repository "drssr/internal/clothes/repository"
	"drssr/internal/looks/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/file_utils"
	"drssr/internal/pkg/rollback"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type ILooksUsecase interface {
	AddLook(ctx context.Context, userEmail string, look models.Look) (models.Look, int, error)
	UpdateLook(ctx context.Context, userEmail string, newLook models.Look, lid uint64, uid uint64) (models.Look, int, error)
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
	userEmail string,
	look models.Look,
) (models.Look, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	// TODO: delete if not needed
	// // checking file ext
	// if !file_utils.IsEnabledExt(look.FileExt) {
	// 	return models.Look{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("LooksUsecase.AddLook: not enabled file extension")
	// }

	folderNameByte := sha1.New().Sum([]byte(userEmail))
	folderName := fmt.Sprintf(hex.EncodeToString(folderNameByte))

	fileName := file_utils.GenerateFileName("look", consts.FileExt)
	folderPath := fmt.Sprintf("%s/%s", consts.LooksBaseFolderPath, folderName)
	filePath := fmt.Sprintf("%s/%s/%s", consts.LooksBaseFolderPath, folderName, fileName)
	look.ImgPath = filePath

	err := file_utils.SaveBase64ToFile(folderPath, filePath, look.Img)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.AddLook: failed to save look's file: %w", err)
	}

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
			Type:     clothedFromDB.Type,
			Name:     clothedFromDB.Name,
			Desc:     clothedFromDB.Desc,
			Brand:    clothedFromDB.Brand,
			Coords:   createdBind.Coords,
			Rotation: createdBind.Rotation,
			Scaling:  createdBind.Scaling,
			ImgPath:  clothedFromDB.ImgPath,
			MaskPath: clothedFromDB.MaskPath,
		})
	}

	// createdLook.Img = look.Img

	return createdLook, http.StatusOK, nil
}

func (lu *looksUsecase) UpdateLook(
	ctx context.Context,
	userEmail string,
	newLook models.Look,
	lid uint64,
	uid uint64,
) (models.Look, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

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

	// TODO: delete if not needed
	// // checking file ext
	// if !file_utils.IsEnabledExt(newLook.FileExt) {
	// 	return models.Look{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("LooksUsecase.UpdateLook: not enabled file extension")
	// }

	folderNameByte := sha1.New().Sum([]byte(userEmail))
	folderName := hex.EncodeToString(folderNameByte)

	fileName := file_utils.GenerateFileName("look", consts.FileExt)
	folderPath := fmt.Sprintf("%s/%s", consts.LooksBaseFolderPath, folderName)
	filePath := fmt.Sprintf("%s/%s/%s", consts.LooksBaseFolderPath, folderName, fileName)
	newLook.ImgPath = filePath

	err = file_utils.SaveBase64ToFile(folderPath, filePath, newLook.Img)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to save look's file: %w", err)
	}

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
			Type:     clothedFromDB.Type,
			Name:     clothedFromDB.Name,
			Desc:     clothedFromDB.Desc,
			Brand:    clothedFromDB.Brand,
			Coords:   createdBind.Coords,
			Rotation: createdBind.Rotation,
			Scaling:  createdBind.Scaling,
			ImgPath:  clothedFromDB.ImgPath,
			MaskPath: clothedFromDB.MaskPath,
		})
	}

	// deleting old img
	err = file_utils.DeleteFile(foundingLook.ImgPath)
	if err != nil {
		rb.Run()

		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("LooksUsecase.UpdateLook: failed to delete old look's img file: %w", err)
	}

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
	}

	return looks, http.StatusOK, nil
}
