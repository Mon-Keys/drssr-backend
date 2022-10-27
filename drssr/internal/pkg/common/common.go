package common

import (
	"bytes"
	"drssr/internal/pkg/consts"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func OpenFileFromReq(r *http.Request, fileName string) (*multipart.File, *multipart.FileHeader, int, error) {
	files := r.MultipartForm.File[fileName]

	// we upload only 1 file
	if len(files) == 0 {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("no file in request")
	}

	fileHeader := files[0]

	if fileHeader.Size > consts.MaxUploadFileSize {
		return nil, nil, http.StatusBadRequest, fmt.Errorf("file is too big")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, nil, http.StatusInternalServerError, fmt.Errorf("failed to open file")
	}

	return &file, fileHeader, http.StatusOK, nil
}

func SaveFile(filePath string, fileBytes []byte) error {
	// create a new file in the dst directory
	dst, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create a new file: %w", err)
	}
	defer dst.Close()

	// copy the uploaded file to the filesystem
	// at the specified destination
	_, err = io.Copy(dst, bytes.NewReader(fileBytes))
	if err != nil {
		// deleting of created file
		removingErr := os.Remove(filePath)
		if removingErr != nil {
			return fmt.Errorf("failed to save data in created file: %w and failed to remove empty created file: %v", err, removingErr)
		}

		return fmt.Errorf("failed to save data in created file: %w", err)
	}

	return nil
}

func IsEnabledFileType(fileType string) bool {
	imgTypes := map[string]bool{
		"image/jpg":  true,
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
	}

	return imgTypes[fileType]
}

func IsEnabledExt(fileType string) bool {
	imgTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}

	return imgTypes[fileType]
}
