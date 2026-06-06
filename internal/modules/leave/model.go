package leave

import "time"

type Status string

const (
	StatusPending  Status = "menunggu"
	StatusApproved Status = "diterima"
	StatusRejected Status = "ditolak"
)

type Submission struct {
	ID         string  `gorm:"type:char(36);primaryKey"`
	StudentID  string  `gorm:"type:char(36);not null;index"`
	Type       string  `gorm:"size:20;not null"`
	Reason     string  `gorm:"type:text;not null"`
	Attachment string  `gorm:"size:255"`
	Status     Status  `gorm:"size:20;not null;default:menunggu"`
	ReviewedBy *string `gorm:"type:char(36)"`
	ReviewNote *string `gorm:"type:text"`
	ReviewedAt *time.Time
	CreatedAt  time.Time `gorm:"not null"`
	UpdatedAt  time.Time `gorm:"not null"`
}
