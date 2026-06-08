package public

import (
	"fmt"

	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type attendanceWindowRow struct {
	SchoolYear   string
	CheckInStart string
	OnTimeUntil  string
	LateUntil    string
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) AttendanceWindow() (*AttendanceWindowResponse, error) {
	fallback := &AttendanceWindowResponse{
		SchoolYear:   "",
		CheckInStart: "06:30:00",
		OnTimeUntil:  "07:00:00",
		LateUntil:    "07:30:00",
	}

	if s.db == nil {
		return fallback, nil
	}

	var row attendanceWindowRow
	if err := s.db.Table("attendance_rules").
		Select("school_years.name as school_year, attendance_rules.check_in_start, attendance_rules.on_time_until, attendance_rules.late_until").
		Joins("join school_years on school_years.id = attendance_rules.school_year_id").
		Where("attendance_rules.is_active = ?", true).
		Order("school_years.start_year desc, attendance_rules.updated_at desc").
		Limit(1).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("get public attendance window: %w", err)
	}

	if row.CheckInStart == "" || row.OnTimeUntil == "" || row.LateUntil == "" {
		return fallback, nil
	}

	return &AttendanceWindowResponse{
		SchoolYear:   row.SchoolYear,
		CheckInStart: row.CheckInStart,
		OnTimeUntil:  row.OnTimeUntil,
		LateUntil:    row.LateUntil,
	}, nil
}
