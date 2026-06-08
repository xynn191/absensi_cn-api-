package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
)

type PhotoUploaderConfig struct {
	CloudName    string
	APIKey       string
	APISecret    string
	UploadFolder string
	LocalRoot    string
}

type PhotoUploader struct {
	cloudinary   *cloudinary.Cloudinary
	uploadFolder string
	localRoot    string
}

func NewPhotoUploader(cfg PhotoUploaderConfig) (*PhotoUploader, error) {
	localRoot := strings.TrimSpace(cfg.LocalRoot)
	if localRoot == "" {
		localRoot = filepath.Join("storage", "uploads")
	}

	uploadFolder := strings.Trim(strings.TrimSpace(cfg.UploadFolder), "/")
	if uploadFolder == "" {
		uploadFolder = "absensi-cn"
	}

	photoUploader := &PhotoUploader{
		uploadFolder: uploadFolder,
		localRoot:    localRoot,
	}

	if cfg.CloudName != "" && cfg.APIKey != "" && cfg.APISecret != "" {
		cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
		if err != nil {
			return nil, fmt.Errorf("initialize cloudinary uploader: %w", err)
		}
		photoUploader.cloudinary = cld
	}

	return photoUploader, nil
}

func (u *PhotoUploader) Store(ctx context.Context, photo *multipart.FileHeader, category string, submittedAt time.Time) (string, string, error) {
	if photo == nil {
		return "", "", fmt.Errorf("photo is required")
	}

	extension := strings.ToLower(filepath.Ext(photo.Filename))
	fileName := uuid.NewString() + extension
	category = strings.Trim(strings.TrimSpace(category), "/")
	if category == "" {
		category = "general"
	}

	if u.cloudinary != nil {
		return u.storeCloudinary(ctx, photo, category, submittedAt, fileName)
	}

	return u.storeLocal(photo, category, submittedAt, fileName)
}

func (u *PhotoUploader) storeCloudinary(ctx context.Context, photo *multipart.FileHeader, category string, submittedAt time.Time, fileName string) (string, string, error) {
	source, err := photo.Open()
	if err != nil {
		return "", "", fmt.Errorf("open photo for cloudinary upload: %w", err)
	}
	defer source.Close()

	folder := path.Join(u.uploadFolder, category, submittedAt.Format("2006"), submittedAt.Format("01"), submittedAt.Format("02"))
	publicID := strings.TrimSuffix(fileName, filepath.Ext(fileName))

	result, err := u.cloudinary.Upload.Upload(ctx, source, uploader.UploadParams{
		Folder:       folder,
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return "", "", fmt.Errorf("upload photo to cloudinary: %w", err)
	}

	return result.SecureURL, fileName, nil
}

func (u *PhotoUploader) storeLocal(photo *multipart.FileHeader, category string, submittedAt time.Time, fileName string) (string, string, error) {
	relativeDir := filepath.Join(category, submittedAt.Format("2006"), submittedAt.Format("01"), submittedAt.Format("02"))
	absoluteDir := filepath.Join(u.localRoot, relativeDir)

	if err := os.MkdirAll(absoluteDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create upload directory: %w", err)
	}

	source, err := photo.Open()
	if err != nil {
		return "", "", fmt.Errorf("open photo for local upload: %w", err)
	}
	defer source.Close()

	absolutePath := filepath.Join(absoluteDir, fileName)
	destination, err := os.Create(absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("create local photo file: %w", err)
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return "", "", fmt.Errorf("save local photo file: %w", err)
	}

	return filepath.ToSlash(filepath.Join("/uploads", relativeDir, fileName)), fileName, nil
}
