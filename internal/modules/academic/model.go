package academic

import "time"

type Teacher struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	UserID    string    `gorm:"type:char(36);not null;uniqueIndex"`
	NIP       *string   `gorm:"column:nip;size:30;uniqueIndex"`
	NUPTK     *string   `gorm:"column:nuptk;size:30;uniqueIndex"`
	Gender    *string   `gorm:"size:20"`
	Phone     *string   `gorm:"size:30"`
	Address   *string   `gorm:"type:text"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Subject struct {
	ID           string    `gorm:"type:char(36);primaryKey"`
	Code         string    `gorm:"size:30;not null;uniqueIndex"`
	Name         string    `gorm:"size:150;not null"`
	SubjectGroup *string   `gorm:"column:subject_group;size:100"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

type Major struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	Code      string    `gorm:"size:30;not null;uniqueIndex"`
	Name      string    `gorm:"size:150;not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type SchoolYear struct {
	ID        string    `gorm:"type:char(36);primaryKey"`
	Name      string    `gorm:"size:30;not null;uniqueIndex"`
	StartYear int       `gorm:"not null"`
	EndYear   int       `gorm:"not null"`
	IsActive  bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type SchoolClass struct {
	ID           string    `gorm:"type:char(36);primaryKey"`
	Grade        string    `gorm:"size:10;not null;index:idx_class_identity,unique"`
	MajorID      string    `gorm:"type:char(36);not null;index:idx_class_identity,unique"`
	Name         string    `gorm:"size:30;not null;index:idx_class_identity,unique"`
	SchoolYearID string    `gorm:"type:char(36);not null;index:idx_class_identity,unique"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

func (SchoolClass) TableName() string {
	return "classes"
}

type TeacherSubjectAssignment struct {
	ID           string    `gorm:"type:char(36);primaryKey"`
	TeacherID    string    `gorm:"type:char(36);not null;index:idx_teacher_subject_assignment,unique"`
	SubjectID    string    `gorm:"type:char(36);not null;index:idx_teacher_subject_assignment,unique"`
	ClassID      string    `gorm:"type:char(36);not null;index:idx_teacher_subject_assignment,unique"`
	SchoolYearID string    `gorm:"type:char(36);not null;index:idx_teacher_subject_assignment,unique"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}

type HomeroomAssignment struct {
	ID           string    `gorm:"type:char(36);primaryKey"`
	TeacherID    string    `gorm:"type:char(36);not null;index:idx_homeroom_teacher_class,unique"`
	ClassID      string    `gorm:"type:char(36);not null;uniqueIndex:idx_homeroom_class_year"`
	SchoolYearID string    `gorm:"type:char(36);not null;uniqueIndex:idx_homeroom_class_year;index:idx_homeroom_teacher_class,unique"`
	IsActive     bool      `gorm:"not null;default:true"`
	CreatedAt    time.Time `gorm:"not null"`
	UpdatedAt    time.Time `gorm:"not null"`
}
