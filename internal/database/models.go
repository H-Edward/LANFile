package database

import "time"

type File struct {
	ID           string `gorm:"primaryKey"`
	StorageKey   string `gorm:"uniqueIndex;not null"`
	OriginalName string `gorm:"not null"`
	ContentType  string
	Size         int64 `gorm:"not null"`
	CreatedAt    time.Time
	Encrypted    string `gorm:"not null;default:none"`

	NeedsAuth         bool   `gorm:"not null;default:false"`
	AuthorisationHash string `gorm:"not null;default:none" json:"-"`
}
