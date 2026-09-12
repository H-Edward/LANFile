package files

import (
	"context"
	"fmt"
	"uuid"

	"github.com/H-Edward/LANFile/internal/database"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateFileInput struct {
	OriginalName      string
	ContentType       string
	Size              int64
	StorageKey        string
	Encrypted         string
	NeedsAuth         bool
	AuthorisationHash string
}

// GetAll retrieves a list of all files from the database.
func (s *Service) GetAll(ctx context.Context) ([]database.File, error) {
	var files []database.File
	if err := s.db.WithContext(ctx).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// GetByID retrieves a file by its ID. Returns an error if the file is not found.
func (s *Service) GetByID(ctx context.Context, id string) (*database.File, error) {
	var file database.File
	if err := s.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("file not found")
		}
		return nil, err
	}
	return &file, nil
}

// GetByName retrieves a list of files by their exact original name. Returns an error if no files are found.
func (s *Service) GetByName(ctx context.Context, originalName string) ([]database.File, error) {
	var files []database.File
	if err := s.db.WithContext(ctx).Where("original_name = ?", originalName).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// SearchByName retrieves a list of files whose original name contains the specified substring. Returns an error if no files are found.
func (s *Service) SearchByName(ctx context.Context, originalName string) ([]database.File, error) {
	var files []database.File
	if err := s.db.WithContext(ctx).Where("original_name LIKE ?", "%"+originalName+"%").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

// Create creates a new file record in the database with the provided input data.
func (s *Service) Create(ctx context.Context, input CreateFileInput) (*database.File, error) {
	file := &database.File{
		ID:                uuid.New().String(),
		StorageKey:        input.StorageKey,
		OriginalName:      input.OriginalName,
		ContentType:       input.ContentType,
		Size:              input.Size,
		Encrypted:         input.Encrypted,
		NeedsAuth:         input.NeedsAuth,
		AuthorisationHash: input.AuthorisationHash,
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		return nil, err
	}

	return file, nil
}

// OverwriteByID updates an existing file record in the database with the provided input data. Returns an error if the file is not found.
func (s *Service) OverwriteByID(ctx context.Context, id string, input CreateFileInput) (*database.File, error) {
	file, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	file.OriginalName = input.OriginalName
	file.ContentType = input.ContentType
	file.Size = input.Size
	file.StorageKey = input.StorageKey
	file.Encrypted = input.Encrypted
	file.NeedsAuth = input.NeedsAuth
	file.AuthorisationHash = input.AuthorisationHash
	if err := s.db.WithContext(ctx).Save(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// DeleteByID deletes a file record from the database by its ID. Returns an error if the file is not found.
func (s *Service) DeleteByID(ctx context.Context, id string) error {
	result := s.db.WithContext(ctx).Delete(&database.File{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("file not found")
	}
	return nil
}
