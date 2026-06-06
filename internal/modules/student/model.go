package student

import "time"

type Student struct {
	ID          string     `gorm:"type:char(36);primaryKey"`
	UserID      string     `gorm:"type:char(36);not null;uniqueIndex"`
	NIS         string     `gorm:"size:10;not null;uniqueIndex"`
	NISN        *string    `gorm:"size:20;uniqueIndex"`
	Gender      *string    `gorm:"size:20"`
	BirthPlace  *string    `gorm:"size:100"`
	BirthDate   *time.Time `gorm:"type:date"`
	Address     *string    `gorm:"type:text"`
	Phone       *string    `gorm:"size:30"`
	ParentName  *string    `gorm:"size:150"`
	ParentPhone *string    `gorm:"size:30"`
	EntryYear   int        `gorm:"not null"`
	IsActive    bool       `gorm:"not null;default:true"`
	CreatedAt   time.Time  `gorm:"not null"`
	UpdatedAt   time.Time  `gorm:"not null"`
}

type MembershipStatus string

const (
	MembershipStatusActive      MembershipStatus = "ACTIVE"
	MembershipStatusTransferred MembershipStatus = "TRANSFERRED"
	MembershipStatusGraduated   MembershipStatus = "GRADUATED"
	MembershipStatusInactive    MembershipStatus = "INACTIVE"
)

type StudentClassMembership struct {
	ID           string           `gorm:"type:char(36);primaryKey"`
	StudentID    string           `gorm:"type:char(36);not null;index:idx_student_membership_unique,unique"`
	ClassID      string           `gorm:"type:char(36);not null;index:idx_student_membership_unique,unique"`
	SchoolYearID string           `gorm:"type:char(36);not null;index:idx_student_membership_unique,unique"`
	Status       MembershipStatus `gorm:"size:20;not null;default:ACTIVE"`
	JoinedAt     *time.Time       `gorm:"type:datetime"`
	LeftAt       *time.Time       `gorm:"type:datetime"`
	IsActive     bool             `gorm:"not null;default:true"`
	CreatedAt    time.Time        `gorm:"not null"`
	UpdatedAt    time.Time        `gorm:"not null"`
}
