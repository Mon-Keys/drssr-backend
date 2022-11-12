package common

import (
	"bytes"
	"drssr/internal/pkg/consts"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

func ReadFileIntoBase64(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file")
	}

	bytesImg, err := ioutil.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("failed to read file")
	}

	decodedFile := base64.StdEncoding.EncodeToString(bytesImg)

	return decodedFile, nil
}

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

func SaveFile(dirPath string, filePath string, fileBytes []byte) error {
	// create dst directory
	err := os.Mkdir(dirPath, 0777)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed to create a new dir: %w", err)
	}
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

func DeleteFile(path string) error {
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
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

func GetExtFromFileType(fileType string) string {
	_, after, _ := strings.Cut(fileType, "/")
	return after
}
