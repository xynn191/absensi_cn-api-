package attendance

import "time"

type AttendanceStatus string

const (
	StatusHadir AttendanceStatus = "hadir"
	StatusTelat AttendanceStatus = "telat"
	StatusIzin  AttendanceStatus = "izin"
	StatusSakit AttendanceStatus = "sakit"
	StatusAlfa  AttendanceStatus = "alfa"
)

type AttendanceRule struct {
	ID           string    `gorm:"type:char(36);primaryKey"`
	SchoolYearID string    `gorm:"type:char(36);not null;uniqueIndex"`
	CheckInStart string    `gorm:"size:8;not null"`
	OnTimeUntil  string    `gorm:"size:8;not null"`
	LateUntil    string    `gorm:"size:8;not null"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

type AttendanceRecord struct {
	ID                       string           `gorm:"type:char(36);primaryKey"`
	StudentID                string           `gorm:"type:char(36);not null;index:idx_student_attendance_date,unique"`
	StudentClassMembershipID string           `gorm:"type:char(36);not null"`
	ClassID                  string           `gorm:"type:char(36);not null"`
	SchoolYearID             string           `gorm:"type:char(36);not null"`
	AttendanceDate           time.Time        `gorm:"type:date;not null;index:idx_student_attendance_date,unique"`
	CheckInAt                *time.Time       `gorm:"type:datetime"`
	Status                   AttendanceStatus `gorm:"size:20;not null"`
	PhotoURL                 *string          `gorm:"size:255"`
	PhotoFilename            *string          `gorm:"size:255"`
	Notes                    *string          `gorm:"type:text"`
	VerifiedBy               *string          `gorm:"type:char(36)"`
	VerifiedAt               *time.Time       `gorm:"type:datetime"`
	VerificationNote         *string          `gorm:"type:text"`
	CreatedAt                time.Time        `gorm:"not null"`
	UpdatedAt                time.Time        `gorm:"not null"`
}

type Record struct {
	ID          string           `json:"id"`
	StudentName string           `json:"student_name"`
	ClassName   string           `json:"class_name"`
	Date        string           `json:"date"`
	CheckIn     string           `json:"check_in"`
	CheckOut    string           `json:"check_out"`
	Status      AttendanceStatus `json:"status"`
}
