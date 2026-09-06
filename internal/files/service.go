package files

import (
	"context"
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
	OriginalName string
	ContentType  string
	Size         int64
	StorageKey   string
	Encrypted    string
}

func (s *Service) GetByID(ctx context.Context, id string) (*database.File, error) {
	var file database.File
	if err := s.db.WithContext(ctx).First(&file, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *Service) GetByName(ctx context.Context, originalName string) ([]database.File, error) {
	var files []database.File
	if err := s.db.WithContext(ctx).Where("original_name = ?", originalName).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (s *Service) Create(ctx context.Context, input CreateFileInput) (*database.File, error) {
	file := &database.File{
		ID:           uuid.New().String(),
		StorageKey:   input.StorageKey,
		OriginalName: input.OriginalName,
		ContentType:  input.ContentType,
		Size:         input.Size,
		Encrypted:    input.Encrypted,
	}

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		return nil, err
	}

	return file, nil
}

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

	if err := s.db.WithContext(ctx).Save(file).Error; err != nil {
		return nil, err
	}
	return file, nil
}

// func (s *Service) DeleteByID(ctx context.Context, id string) error {
// 	result := s.db.WithContext(ctx).Delete(&database.File{}, "id = ?", id)
// 	if result.Error != nil {
// 		return result.Error
// 	}
// 	if result.RowsAffected == 0 {
// 		return fmt.Errorf("file not found")
// 	}
// 	return nil
// }
