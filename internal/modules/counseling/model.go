package counseling

import "time"

type Note struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	StudentID string    `gorm:"type:char(36);not null;index"`
	CreatedBy string    `gorm:"type:char(36);not null"`
	Title     string    `gorm:"size:150;not null"`
	Note      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
