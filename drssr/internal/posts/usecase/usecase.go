package usecase

import (
	"context"
	"crypto/sha1"
	clothes_repository "drssr/internal/clothes/repository"
	looks_repository "drssr/internal/looks/repository"
	"drssr/internal/models"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/rollback"
	"drssr/internal/posts/repository"
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

type IPostUsecase interface {
	AddPost(ctx context.Context, look models.Post) (models.Post, int, error)
	DeletePost(ctx context.Context, uid uint64, pid uint64) (int, error)

	GetPostByID(ctx context.Context, pid uint64) (models.Post, int, error)
	GetUserPosts(ctx context.Context, uid uint64, limit int, offset int) (models.ArrayPosts, int, error)
	GetAllPosts(ctx context.Context, limit int, offset int) (models.ArrayPosts, int, error)
}

type postsUsecase struct {
	psql        repository.IPostgresqlRepository
	clothesPsql clothes_repository.IPostgresqlRepository
	looksPsql   looks_repository.IPostgresqlRepository
	logger      logrus.Logger
}

func NewPostsUsecase(
	pr repository.IPostgresqlRepository,
	cpr clothes_repository.IPostgresqlRepository,
	lpr looks_repository.IPostgresqlRepository,
	logger logrus.Logger,
) IPostUsecase {
	return &postsUsecase{
		psql:        pr,
		clothesPsql: cpr,
		looksPsql:   lpr,
		logger:      logger,
	}
}

func (pu *postsUsecase) generateClothesElement(
	ctx context.Context,
	cid uint64,
	postCreatorID uint64,
) (models.Clothes, int, error) {
	postClothes, err := pu.clothesPsql.GetClothesByID(ctx, cid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Clothes{},
				http.StatusNotFound,
				fmt.Errorf("clothes with such id not found")
		}
		return models.Clothes{},
			http.StatusInternalServerError,
			fmt.Errorf("failed to get clothes from db: %w", err)
	}

	// checking owner
	if postClothes.OwnerID != postCreatorID {
		return models.Clothes{},
			http.StatusForbidden,
			fmt.Errorf("can't create post with not owner clothes")
	}

	// // decoding img
	// decodedImg, err := common.ReadFileIntoBase64(postClothes.ImgPath)
	// if err != nil {
	// 	return models.Clothes{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("failed to read clothes img into base64: %w", err)
	// }

	// postClothes.Img = decodedImg

	// // decoding mask
	// decodedMask, err := common.ReadFileIntoBase64(postClothes.MaskPath)
	// if err != nil {
	// 	return models.Clothes{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("failed to read clothes mask into base64: %w", err)
	// }

	// postClothes.Mask = decodedMask

	return postClothes, http.StatusOK, nil
}

func (pu *postsUsecase) generateLookElement(
	ctx context.Context,
	lid uint64,
	postCreatorID uint64,
) (models.Look, int, error) {
	// getting from db
	postLook, err := pu.looksPsql.GetLookByID(ctx, lid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Look{},
				http.StatusNotFound,
				fmt.Errorf("look with such id not found")
		}
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("failed to get look from db: %w", err)
	}

	// checking owner
	if postLook.CreatorID != postCreatorID {
		return models.Look{},
			http.StatusForbidden,
			fmt.Errorf("can't create post with not owner look")
	}

	// get look's clothes
	clothes, err := pu.looksPsql.GetLookClothes(ctx, postLook.ID)
	if err != nil {
		return models.Look{},
			http.StatusInternalServerError,
			fmt.Errorf("failed to get look's clothes from db: %w", err)
	}

	postLook.Clothes = clothes

	// // decoding look's img
	// decodedLookImg, err := common.ReadFileIntoBase64(postLook.ImgPath)
	// if err != nil {
	// 	return models.Look{},
	// 		http.StatusInternalServerError,
	// 		fmt.Errorf("failed to read look's img file into base64: %w", err)
	// }

	// postLook.Img = decodedLookImg

	return postLook, http.StatusOK, nil
}

func (pu *postsUsecase) generateElement(
	ctx context.Context,
	post models.Post,
) (models.Post, int, error) {
	var postClothes models.Clothes
	var postLook models.Look
	var status int
	var err error

	switch post.Type {

	// post with clothes
	case models.PostTypeClothes:
		// generating clothes element
		postClothes, status, err = pu.generateClothesElement(ctx, post.ElementID, post.CreatorID)
		if err != nil || status != http.StatusOK {
			return models.Post{},
				status,
				fmt.Errorf("PostsUsecase.AddPost: failed to generate post's clothes element: %w", err)
		}

	// post with look
	case models.PostTypeLook:
		// generating look element
		postLook, status, err = pu.generateLookElement(ctx, post.ElementID, post.CreatorID)
		if err != nil || status != http.StatusOK {
			return models.Post{},
				status,
				fmt.Errorf("PostsUsecase.AddPost: failed to generate post's look element: %w", err)
		}

	default:
		return models.Post{},
			http.StatusBadRequest,
			fmt.Errorf("PostsUsecase.AddPost: unknown post type %s", post.Type)
	}

	post.Clothes = postClothes
	post.Look = postLook

	return post, http.StatusOK, nil
}

func (pu *postsUsecase) AddPost(
	ctx context.Context,
	post models.Post,
) (models.Post, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	// generate post's element
	post, status, err := pu.generateElement(ctx, post)
	if err != nil || status != http.StatusOK {
		return models.Post{},
			status,
			fmt.Errorf("PostsUsecase.AddPost: failed to generate post's element: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	folderNameByte := sha1.New().Sum([]byte(today))
	folderName := hex.EncodeToString(folderNameByte)
	folderPath := fmt.Sprintf("%s/%s", consts.PostsBaseFolderPath, folderName)

	for fileName, preview := range post.Previews {
		// checking file ext
		splitedFilename := strings.Split(fileName, ".")
		ext := fmt.Sprintf(".%s", splitedFilename[len(splitedFilename)-1])
		if !common.IsEnabledExt(ext) {
			rb.Run()

			return models.Post{},
				http.StatusInternalServerError,
				fmt.Errorf("PostsUsecase.AddPost: not enabled file extension")
		}

		decodedPreview, err := base64.StdEncoding.DecodeString(preview)
		if err != nil {
			rb.Run()

			return models.Post{},
				http.StatusInternalServerError,
				fmt.Errorf("PostsUsecase.AddPost: failed to decode base64 into byte slice: %w", err)
		}

		filePath := fmt.Sprintf("%s/%s/%s", consts.PostsBaseFolderPath, folderName, fileName)

		err = common.SaveFile(folderPath, filePath, decodedPreview)
		if err != nil {
			rb.Run()

			return models.Post{},
				http.StatusInternalServerError,
				fmt.Errorf("PostsUsecase.AddPost: failed to save post's preview file: %w", err)
		}

		rb.Add(func() {
			err := os.Remove(filePath)
			if err != nil {
				pu.logger.Errorf("PostsUsecase.AddPost: failed to rollback creating of posts's preview file: %w", err)
			}
		})

		post.PreviewsPaths = append(post.PreviewsPaths, filePath)
	}

	createdPost, err := pu.psql.AddPost(ctx, post)
	if err != nil {
		rb.Run()

		return models.Post{},
			http.StatusInternalServerError,
			fmt.Errorf("PostsUsecase.AddPost: failed to save post in db: %w", err)
	}

	rb.Add(func() {
		_, err := pu.psql.DeletePost(ctx, createdPost.ID)
		if err != nil {
			pu.logger.Errorf("PostsUsecase.AddPost: failed to rollback saving post in db: %w", err)
		}
	})

	// createdPost.Previews = post.Previews
	createdPost.Clothes = post.Clothes
	createdPost.Look = post.Look

	return createdPost, http.StatusOK, nil
}

func (pu *postsUsecase) DeletePost(ctx context.Context, uid uint64, pid uint64) (int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	// checking post in db
	foundingPost, err := pu.psql.GetPostByID(ctx, pid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusNotFound, fmt.Errorf("PostsUsecase.DeletePost: post not found")
		}
		return http.StatusInternalServerError, fmt.Errorf("PostsUsecase.DeletePost: failed to getting post from db: %w", err)
	}

	if uid != foundingPost.CreatorID {
		return http.StatusForbidden, fmt.Errorf("PostsUsecase.DeleteLook: can't delete not user's post")
	}

	// deleting post
	deletedLook, err := pu.psql.DeletePost(ctx, pid)
	if err != nil {
		rb.Run()

		return http.StatusInternalServerError, fmt.Errorf("PostsUsecase.DeleteLook: failed to delete post from db: %w", err)
	}

	rb.Add(func() {
		_, err := pu.psql.AddPost(ctx, deletedLook)
		if err != nil {
			pu.logger.Errorf("PostsUsecase.DeleteLook: failed to rollback deleting of post from db: %w", err)
		}
	})

	for _, preview := range deletedLook.PreviewsPaths {
		err = common.DeleteFile(preview)
		if err != nil {
			rb.Run()

			return http.StatusInternalServerError, fmt.Errorf("PostsUsecase.DeleteLook: failed to delete post's preview from disk: %w", err)
		}

		rb.Add(func() {
			encodedPreview, err := common.ReadFileIntoBase64(preview)
			if err != nil {
				pu.logger.Errorf("PostsUsecase.DeleteLook: failed to rollback deleting of post's preview from disk: %w", err)
			}

			decodedPreview, err := base64.StdEncoding.DecodeString(encodedPreview)
			if err != nil {
				pu.logger.Errorf("PostsUsecase.DeleteLook: failed to rollback deleting of post's preview from disk: %w", err)
			}

			lastSlashIndex := strings.LastIndex(preview, "/")
			dirPath := preview[:lastSlashIndex-1]

			err = common.SaveFile(dirPath, preview, decodedPreview)
			if err != nil {
				pu.logger.Errorf("PostsUsecase.DeleteLook: failed to rollback deleting of post's preview from disk: %w", err)
			}
		})
	}

	return http.StatusOK, nil
}

func (pu *postsUsecase) GetUserPosts(ctx context.Context, uid uint64, limit int, offset int) (models.ArrayPosts, int, error) {
	posts, err := pu.psql.GetUserPosts(ctx, limit, offset, uid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, http.StatusNotFound, fmt.Errorf("PostsUsecase.GetUserPosts: user don't have any looks")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("PostsUsecase.GetUserPosts: failed to get user looks from db: %w", err)
	}

	for i := range posts {
		var status int
		posts[i], status, err = pu.generateElement(ctx, posts[i])
		if err != nil || status != http.StatusOK {
			return nil, status, fmt.Errorf("PostsUsecase.GetUserPosts: failed to generate post's element: %w", err)
		}

		// posts[i].Previews = make(map[string]string, len(posts[i].PreviewsPaths))

		// for _, previewPath := range posts[i].PreviewsPaths {
		// 	decodedPreview, err := common.ReadFileIntoBase64(previewPath)
		// 	if err != nil {
		// 		return nil, http.StatusInternalServerError, fmt.Errorf("PostsUsecase.GetUserPosts: failed to read preview file into base64: %w", err)
		// 	}

		// 	lastSlashIndex := strings.LastIndex(previewPath, "/")
		// 	fileName := previewPath[lastSlashIndex+1:]

		// 	posts[i].Previews[fileName] = decodedPreview
		// }
	}

	return posts, http.StatusOK, nil
}

func (pu *postsUsecase) GetPostByID(ctx context.Context, pid uint64) (models.Post, int, error) {
	// checking post in db
	foundingPost, err := pu.psql.GetPostByID(ctx, pid)
	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Post{},
				http.StatusNotFound,
				fmt.Errorf("PostsUsecase.GetPostByID: look not found")
		}
		return models.Post{},
			http.StatusInternalServerError,
			fmt.Errorf("PostsUsecase.GetPostByID: failed to found look in db")
	}

	var status int
	foundingPost, status, err = pu.generateElement(ctx, foundingPost)
	if err != nil || status != http.StatusOK {
		return models.Post{}, status, fmt.Errorf("PostsUsecase.GetPostByID: failed to generate post's element: %w", err)
	}

	// foundingPost.Previews = make(map[string]string, len(foundingPost.PreviewsPaths))

	// for _, previewPath := range foundingPost.PreviewsPaths {
	// 	decodedPreview, err := common.ReadFileIntoBase64(previewPath)
	// 	if err != nil {
	// 		return models.Post{}, http.StatusInternalServerError, fmt.Errorf("PostsUsecase.GetPostByID: failed to read preview file into base64: %w", err)
	// 	}

	// 	lastSlashIndex := strings.LastIndex(previewPath, "/")
	// 	fileName := previewPath[lastSlashIndex+1:]

	// 	foundingPost.Previews[fileName] = decodedPreview
	// }

	return foundingPost, http.StatusOK, nil
}

func (pu *postsUsecase) GetAllPosts(ctx context.Context, limit int, offset int) (models.ArrayPosts, int, error) {
	posts, err := pu.psql.GetAllPosts(ctx, limit, offset)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, http.StatusNotFound, fmt.Errorf("PostsUsecase.GetUserPosts: user don't have any looks")
		}
		return nil, http.StatusInternalServerError, fmt.Errorf("PostsUsecase.GetUserPosts: failed to get user looks from db: %w", err)
	}

	for i := range posts {
		var status int
		posts[i], status, err = pu.generateElement(ctx, posts[i])
		if err != nil || status != http.StatusOK {
			return nil, status, fmt.Errorf("PostsUsecase.GetUserPosts: failed to generate post's element: %w", err)
		}

		// posts[i].Previews = make(map[string]string, len(posts[i].PreviewsPaths))

		// for _, previewPath := range posts[i].PreviewsPaths {
		// 	decodedPreview, err := common.ReadFileIntoBase64(previewPath)
		// 	if err != nil {
		// 		return nil, http.StatusInternalServerError, fmt.Errorf("PostsUsecase.GetUserPosts: failed to read preview file into base64: %w", err)
		// 	}

		// 	lastSlashIndex := strings.LastIndex(previewPath, "/")
		// 	fileName := previewPath[lastSlashIndex+1:]

		// 	posts[i].Previews[fileName] = decodedPreview
		// }
	}

	return posts, http.StatusOK, nil
}
