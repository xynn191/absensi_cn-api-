package attendance

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"absensi-cn-api/pkg/storage"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultCheckInStart = "06:30:00"
	defaultOnTimeUntil  = "07:00:00"
	defaultLateUntil    = "07:30:00"
	maxPhotoSizeBytes   = 5 * 1024 * 1024
)

var (
	ErrAttendanceDataUnavailable = errors.New("attendance database is not available")
	ErrStudentContextNotFound    = errors.New("student context not found")
	ErrAttendanceRecordNotFound  = errors.New("attendance record not found")
	ErrAttendanceNotOpen         = errors.New("attendance check-in has not opened yet")
	ErrAttendanceAlreadyRecorded = errors.New("attendance for today has already been recorded")
	ErrAttendancePhotoRequired   = errors.New("attendance photo is required")
	ErrAttendancePhotoTooLarge   = errors.New("attendance photo must be 5MB or smaller")
	ErrAttendancePhotoInvalid    = errors.New("attendance photo must be jpg, jpeg, png, or webp")
	ErrAttendanceReviewInvalid   = errors.New("attendance review status is invalid")
)

type Service struct {
	db            *gorm.DB
	photoUploader *storage.PhotoUploader
}

type TodayStatusCounts struct {
	Present    int
	Late       int
	Permission int
	Sick       int
	Alpha      int
}

type studentContextRow struct {
	StudentID        string
	UserID           string
	Name             string
	NIS              string
	NISN             *string
	Gender           *string
	Phone            *string
	StudentActive    bool
	MembershipID     string
	MembershipStatus string
	ClassID          string
	Grade            string
	ClassLabel       string
	MajorCode        string
	MajorName        string
	SchoolYearID     string
	SchoolYearName   string
}

type attendanceRow struct {
	ID               string
	StudentID        string
	StudentName      string
	NIS              string
	ClassID          string
	ClassName        string
	SchoolYearID     string
	SchoolYearName   string
	AttendanceDate   time.Time
	CheckInAt        *time.Time
	Status           string
	PhotoURL         *string
	Notes            *string
	VerifiedByName   *string
	VerifiedAt       *time.Time
	VerificationNote *string
}

func NewService(db *gorm.DB, photoUploader *storage.PhotoUploader) *Service {
	return &Service{
		db:            db,
		photoUploader: photoUploader,
	}
}

func (s *Service) GetHistory() []Record {
	if s.db == nil {
		return nil
	}

	rows, err := s.queryAttendanceRows("", "", "", 100)
	if err != nil {
		return nil
	}

	result := make([]Record, 0, len(rows))
	for _, row := range rows {
		record := Record{
			ID:          row.ID,
			StudentName: row.StudentName,
			ClassName:   row.ClassName,
			Date:        row.AttendanceDate.Format("2006-01-02"),
			CheckOut:    "",
			Status:      AttendanceStatus(row.Status),
		}
		if row.CheckInAt != nil {
			record.CheckIn = row.CheckInAt.In(attendanceLocation()).Format("15:04")
		}
		result = append(result, record)
	}

	return result
}

func (s *Service) GetTodayStatusCounts() (TodayStatusCounts, error) {
	if s.db == nil {
		return TodayStatusCounts{}, ErrAttendanceDataUnavailable
	}

	today := attendanceDateOnly(currentAttendanceTime())
	if err := s.ensureAlphaRecordsForDate(today); err != nil {
		return TodayStatusCounts{}, err
	}

	var rows []struct {
		Status string
		Count  int
	}
	if err := s.db.Model(&AttendanceRecord{}).
		Select("status, count(*) as count").
		Where("attendance_date = ?", today.Format("2006-01-02")).
		Group("status").
		Scan(&rows).Error; err != nil {
		return TodayStatusCounts{}, fmt.Errorf("count today attendance statuses: %w", err)
	}

	status := TodayStatusCounts{}
	for _, row := range rows {
		switch AttendanceStatus(row.Status) {
		case StatusHadir:
			status.Present = row.Count
		case StatusTelat:
			status.Late = row.Count
		case StatusIzin:
			status.Permission = row.Count
		case StatusSakit:
			status.Sick = row.Count
		case StatusAlfa:
			status.Alpha = row.Count
		}
	}

	return status, nil
}

func (s *Service) GetAttendancePercentage() (int, error) {
	if s.db == nil {
		return 0, ErrAttendanceDataUnavailable
	}

	var total int64
	if err := s.db.Model(&AttendanceRecord{}).Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count attendance records: %w", err)
	}
	if total == 0 {
		return 0, nil
	}

	var positive int64
	if err := s.db.Model(&AttendanceRecord{}).Where("status <> ?", StatusAlfa).Count(&positive).Error; err != nil {
		return 0, fmt.Errorf("count positive attendance records: %w", err)
	}

	return int(float64(positive) / float64(total) * 100), nil
}

func (s *Service) GetStudentContextByUserID(userID string) (*StudentContextResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}

	contextRow, err := s.findStudentContext(userID)
	if err != nil {
		return nil, err
	}

	response := mapStudentContext(*contextRow)
	return &response, nil
}

func (s *Service) GetTodayAttendance(userID string) (*TodayAttendanceResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}

	contextRow, err := s.findStudentContext(userID)
	if err != nil {
		return nil, err
	}

	rule, err := s.findAttendanceRule(contextRow.SchoolYearID)
	if err != nil {
		return nil, err
	}

	now := currentAttendanceTime()
	record, err := s.findAttendanceRecordByStudentAndDate(contextRow.StudentID, now)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	canCheckIn := record == nil || (record.PhotoURL == nil && record.CheckInAt == nil)
	currentStatus := determineAttendanceStatus(now, rule.CheckInStart, rule.OnTimeUntil, rule.LateUntil)
	message := buildAttendanceMessage(now, rule.CheckInStart, rule.LateUntil, record != nil)

	response := &TodayAttendanceResponse{
		Student:       mapStudentContext(*contextRow),
		Window:        mapAttendanceWindow(rule),
		CanCheckIn:    canCheckIn,
		CurrentStatus: string(currentStatus),
		CurrentTime:   now.Format(time.RFC3339),
		Message:       message,
	}

	if record != nil {
		mapped, err := s.mapAttendanceRecord(*record)
		if err != nil {
			return nil, err
		}
		response.Attendance = mapped
		response.CurrentStatus = mapped.Status
		response.CanCheckIn = canCheckIn
		response.Message = "attendance for today has already been submitted"
	}

	return response, nil
}

func (s *Service) CheckIn(userID string, photo *multipart.FileHeader, notes string) (*AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}
	if photo == nil {
		return nil, ErrAttendancePhotoRequired
	}
	if photo.Size > maxPhotoSizeBytes {
		return nil, ErrAttendancePhotoTooLarge
	}
	if !isAllowedPhotoExtension(photo.Filename) {
		return nil, ErrAttendancePhotoInvalid
	}

	contextRow, err := s.findStudentContext(userID)
	if err != nil {
		return nil, err
	}

	rule, err := s.findAttendanceRule(contextRow.SchoolYearID)
	if err != nil {
		return nil, err
	}

	now := currentAttendanceTime()
	if now.Before(attendanceMoment(now, rule.CheckInStart)) {
		return nil, ErrAttendanceNotOpen
	}

	existingRecord, err := s.findAttendanceRecordByStudentAndDate(contextRow.StudentID, now)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingRecord != nil && existingRecord.PhotoURL != nil && existingRecord.CheckInAt != nil {
		return s.mapAttendanceRecord(*existingRecord)
	}

	photoURL, photoFilename, err := s.storeAttendancePhoto(photo, now)
	if err != nil {
		return nil, err
	}

	status := determineAttendanceStatus(now, rule.CheckInStart, rule.OnTimeUntil, rule.LateUntil)
	trimmedNotes := strings.TrimSpace(notes)

	var savedRecord AttendanceRecord
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if existingRecord != nil {
			existingRecord.StudentClassMembershipID = contextRow.MembershipID
			existingRecord.ClassID = contextRow.ClassID
			existingRecord.SchoolYearID = contextRow.SchoolYearID
			existingRecord.CheckInAt = &now
			existingRecord.Status = status
			existingRecord.PhotoURL = stringPtr(photoURL)
			existingRecord.PhotoFilename = stringPtr(photoFilename)
			existingRecord.Notes = optionalString(trimmedNotes)
			existingRecord.VerifiedBy = nil
			existingRecord.VerifiedAt = nil
			existingRecord.VerificationNote = nil
			if err := tx.Save(existingRecord).Error; err != nil {
				return fmt.Errorf("update attendance record: %w", err)
			}
			savedRecord = *existingRecord
			return nil
		}

		record := AttendanceRecord{
			ID:                       uuid.NewString(),
			StudentID:                contextRow.StudentID,
			StudentClassMembershipID: contextRow.MembershipID,
			ClassID:                  contextRow.ClassID,
			SchoolYearID:             contextRow.SchoolYearID,
			AttendanceDate:           attendanceDateOnly(now),
			CheckInAt:                &now,
			Status:                   status,
			PhotoURL:                 stringPtr(photoURL),
			PhotoFilename:            stringPtr(photoFilename),
			Notes:                    optionalString(trimmedNotes),
		}
		if err := tx.Create(&record).Error; err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				return ErrAttendanceAlreadyRecorded
			}
			return fmt.Errorf("create attendance record: %w", err)
		}
		savedRecord = record
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrAttendanceAlreadyRecorded) {
			existing, _ := s.findAttendanceRecordByStudentAndDate(contextRow.StudentID, now)
			if existing != nil {
				return s.mapAttendanceRecord(*existing)
			}
		}
		return nil, err
	}

	return s.mapAttendanceRecord(savedRecord)
}

func (s *Service) ListStudentAttendanceHistory(userID string) ([]AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}

	contextRow, err := s.findStudentContext(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queryAttendanceRows(contextRow.StudentID, "", "", 90)
	if err != nil {
		return nil, err
	}

	return mapAttendanceRows(rows), nil
}

func (s *Service) ListAdminAttendance(date, status, classID string) ([]AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}

	targetDate := currentAttendanceTime()
	if strings.TrimSpace(date) != "" {
		parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(date), attendanceLocation())
		if err != nil {
			return nil, fmt.Errorf("attendance date must be in YYYY-MM-DD format")
		}
		targetDate = parsed
	}

	if err := s.ensureAlphaRecordsForDate(targetDate); err != nil {
		return nil, err
	}

	rows, err := s.queryAttendanceRows("", targetDate.Format("2006-01-02"), strings.TrimSpace(status), 0)
	if err != nil {
		return nil, err
	}

	result := make([]AttendanceRecordResponse, 0, len(rows))
	for _, row := range rows {
		if classID != "" && row.ClassID != classID {
			continue
		}
		result = append(result, mapAttendanceRow(row))
	}

	return result, nil
}

func (s *Service) ReviewAttendance(id, actorID string, input ReviewAttendanceRequest) (*AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrAttendanceDataUnavailable
	}

	normalizedStatus := AttendanceStatus(strings.ToLower(strings.TrimSpace(input.Status)))
	switch normalizedStatus {
	case StatusHadir, StatusTelat, StatusIzin, StatusSakit, StatusAlfa:
	default:
		return nil, ErrAttendanceReviewInvalid
	}

	record, err := s.findAttendanceRecordByID(id)
	if err != nil {
		return nil, err
	}

	now := currentAttendanceTime()
	record.Status = normalizedStatus
	record.VerifiedBy = stringPtr(actorID)
	record.VerifiedAt = &now
	record.VerificationNote = optionalString(input.VerificationNote)

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("review attendance record: %w", err)
	}

	return s.mapAttendanceRecord(*record)
}

func (s *Service) queryAttendanceRows(studentID, attendanceDate, status string, limit int) ([]attendanceRow, error) {
	var rows []attendanceRow

	query := s.db.Table("attendance_records as ar").
		Select("ar.id, ar.student_id, users.name as student_name, students.nis, ar.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, ar.school_year_id, school_years.name as school_year_name, ar.attendance_date, ar.check_in_at, ar.status, ar.photo_url, ar.notes, verifier.name as verified_by_name, ar.verified_at, ar.verification_note").
		Joins("join students on students.id = ar.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("join classes on classes.id = ar.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = ar.school_year_id").
		Joins("left join users as verifier on verifier.id = ar.verified_by")

	if studentID != "" {
		query = query.Where("ar.student_id = ?", studentID)
	}
	if attendanceDate != "" {
		query = query.Where("ar.attendance_date = ?", attendanceDate)
	}
	if status != "" {
		query = query.Where("ar.status = ?", strings.ToLower(status))
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("ar.attendance_date desc, ar.check_in_at desc, users.name asc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list attendance rows: %w", err)
	}

	return rows, nil
}

func (s *Service) findStudentContext(userID string) (*studentContextRow, error) {
	var row studentContextRow
	if err := s.db.Table("students").
		Select("students.id as student_id, students.user_id, users.name, students.nis, students.nisn, students.gender, students.phone, students.is_active as student_active, scm.id as membership_id, scm.status as membership_status, scm.class_id, classes.grade, classes.name as class_label, majors.code as major_code, majors.name as major_name, scm.school_year_id, school_years.name as school_year_name").
		Joins("join users on users.id = students.user_id").
		Joins("join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("join classes on classes.id = scm.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = scm.school_year_id").
		Where("students.user_id = ? AND students.is_active = ?", userID, true).
		Order("school_years.start_year desc").
		Limit(1).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("find student context: %w", err)
	}
	if row.StudentID == "" {
		return nil, ErrStudentContextNotFound
	}
	return &row, nil
}

func (s *Service) findAttendanceRule(schoolYearID string) (*AttendanceRule, error) {
	var rule AttendanceRule
	err := s.db.Where("school_year_id = ? AND is_active = ?", schoolYearID, true).First(&rule).Error
	if err == nil {
		return &rule, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find attendance rule: %w", err)
	}

	return &AttendanceRule{
		SchoolYearID: schoolYearID,
		CheckInStart: defaultCheckInStart,
		OnTimeUntil:  defaultOnTimeUntil,
		LateUntil:    defaultLateUntil,
		IsActive:     true,
	}, nil
}

func (s *Service) findAttendanceRecordByStudentAndDate(studentID string, date time.Time) (*AttendanceRecord, error) {
	var record AttendanceRecord
	err := s.db.Where("student_id = ? AND attendance_date = ?", studentID, attendanceDateOnly(date).Format("2006-01-02")).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find attendance record by student/date: %w", err)
	}
	return &record, nil
}

func (s *Service) findAttendanceRecordByID(id string) (*AttendanceRecord, error) {
	var record AttendanceRecord
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceRecordNotFound
		}
		return nil, fmt.Errorf("find attendance record: %w", err)
	}
	return &record, nil
}

func (s *Service) ensureAlphaRecordsForDate(date time.Time) error {
	if s.db == nil {
		return ErrAttendanceDataUnavailable
	}

	targetDate := attendanceDateOnly(date)
	now := currentAttendanceTime()
	today := attendanceDateOnly(now)
	if targetDate.After(today) {
		return nil
	}

	type membershipRow struct {
		StudentID    string
		MembershipID string
		ClassID      string
		SchoolYearID string
	}

	var memberships []membershipRow
	if err := s.db.Table("student_class_memberships as scm").
		Select("scm.student_id, scm.id as membership_id, scm.class_id, scm.school_year_id").
		Joins("join students on students.id = scm.student_id").
		Where("scm.is_active = ? AND students.is_active = ?", true, true).
		Scan(&memberships).Error; err != nil {
		return fmt.Errorf("list active memberships for alpha sync: %w", err)
	}

	for _, membership := range memberships {
		if targetDate.Equal(today) {
			rule, err := s.findAttendanceRule(membership.SchoolYearID)
			if err != nil {
				return err
			}
			if now.Before(attendanceMoment(now, rule.LateUntil)) {
				continue
			}
		}

		record := AttendanceRecord{
			ID:                       uuid.NewString(),
			StudentID:                membership.StudentID,
			StudentClassMembershipID: membership.MembershipID,
			ClassID:                  membership.ClassID,
			SchoolYearID:             membership.SchoolYearID,
			AttendanceDate:           targetDate,
			Status:                   StatusAlfa,
		}
		if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&record).Error; err != nil {
			var mysqlErr *mysql.MySQLError
			if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
				continue
			}
			return fmt.Errorf("create alpha attendance record: %w", err)
		}
	}

	return nil
}

func (s *Service) storeAttendancePhoto(photo *multipart.FileHeader, submittedAt time.Time) (string, string, error) {
	if s.photoUploader == nil {
		return "", "", fmt.Errorf("photo uploader is not configured")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	photoURL, fileName, err := s.photoUploader.Store(ctx, photo, "attendance", submittedAt)
	if err != nil {
		return "", "", fmt.Errorf("store attendance photo: %w", err)
	}
	return photoURL, fileName, nil
}

func (s *Service) mapAttendanceRecord(record AttendanceRecord) (*AttendanceRecordResponse, error) {
	rows, err := s.queryAttendanceRows("", record.AttendanceDate.Format("2006-01-02"), "", 0)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == record.ID {
			response := mapAttendanceRow(row)
			return &response, nil
		}
	}
	return nil, ErrAttendanceRecordNotFound
}

func mapAttendanceRows(rows []attendanceRow) []AttendanceRecordResponse {
	result := make([]AttendanceRecordResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapAttendanceRow(row))
	}
	return result
}

func mapAttendanceRow(row attendanceRow) AttendanceRecordResponse {
	response := AttendanceRecordResponse{
		ID:             row.ID,
		StudentID:      row.StudentID,
		StudentName:    row.StudentName,
		NIS:            row.NIS,
		ClassID:        row.ClassID,
		ClassName:      row.ClassName,
		SchoolYearID:   row.SchoolYearID,
		SchoolYearName: row.SchoolYearName,
		AttendanceDate: row.AttendanceDate.Format("2006-01-02"),
		Status:         row.Status,
	}

	if row.CheckInAt != nil {
		response.CheckInAt = row.CheckInAt.In(attendanceLocation()).Format(time.RFC3339)
	}
	if row.PhotoURL != nil {
		response.PhotoURL = *row.PhotoURL
	}
	if row.Notes != nil {
		response.Notes = *row.Notes
	}
	if row.VerifiedByName != nil {
		response.VerifiedBy = *row.VerifiedByName
	}
	if row.VerifiedAt != nil {
		response.VerifiedAt = row.VerifiedAt.In(attendanceLocation()).Format(time.RFC3339)
	}
	if row.VerificationNote != nil {
		response.VerificationNote = *row.VerificationNote
	}

	return response
}

func mapStudentContext(row studentContextRow) StudentContextResponse {
	className := strings.TrimSpace(fmt.Sprintf("%s %s %s", row.Grade, row.MajorCode, row.ClassLabel))
	response := StudentContextResponse{
		ID:               row.StudentID,
		UserID:           row.UserID,
		Name:             row.Name,
		NIS:              row.NIS,
		Gender:           dereference(row.Gender),
		Phone:            dereference(row.Phone),
		ClassID:          row.ClassID,
		ClassName:        className,
		MajorCode:        row.MajorCode,
		MajorName:        row.MajorName,
		SchoolYearID:     row.SchoolYearID,
		SchoolYearName:   row.SchoolYearName,
		MembershipID:     row.MembershipID,
		MembershipStatus: row.MembershipStatus,
		IsActive:         row.StudentActive,
	}
	if row.NISN != nil {
		response.NISN = *row.NISN
	}
	return response
}

func mapAttendanceWindow(rule *AttendanceRule) AttendanceWindowResponse {
	return AttendanceWindowResponse{
		CheckInStart: rule.CheckInStart,
		OnTimeUntil:  rule.OnTimeUntil,
		LateUntil:    rule.LateUntil,
	}
}

func determineAttendanceStatus(now time.Time, checkInStart, onTimeUntil, lateUntil string) AttendanceStatus {
	checkInOpenAt := attendanceMoment(now, checkInStart)
	onTimeUntilAt := attendanceMoment(now, onTimeUntil)
	lateUntilAt := attendanceMoment(now, lateUntil)

	switch {
	case now.Before(checkInOpenAt):
		return StatusAlfa
	case now.Before(onTimeUntilAt):
		return StatusHadir
	case now.Before(lateUntilAt) || now.Equal(lateUntilAt):
		return StatusTelat
	default:
		return StatusAlfa
	}
}

func buildAttendanceMessage(now time.Time, checkInStart, lateUntil string, alreadyRecorded bool) string {
	if alreadyRecorded {
		return "attendance for today has already been submitted"
	}
	if now.Before(attendanceMoment(now, checkInStart)) {
		return "attendance opens at " + checkInStart
	}
	if now.After(attendanceMoment(now, lateUntil)) {
		return "attendance window has passed; new submission will be recorded as alfa"
	}
	return "attendance is open"
}

func attendanceMoment(base time.Time, clockValue string) time.Time {
	location := attendanceLocation()
	parsed, err := time.ParseInLocation("15:04:05", clockValue, location)
	if err != nil {
		parsed, _ = time.ParseInLocation("15:04", clockValue, location)
	}
	current := base.In(location)
	return time.Date(current.Year(), current.Month(), current.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, location)
}

func attendanceDateOnly(value time.Time) time.Time {
	current := value.In(attendanceLocation())
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, attendanceLocation())
}

func currentAttendanceTime() time.Time {
	return time.Now().In(attendanceLocation())
}

func attendanceLocation() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func stringPtr(value string) *string {
	return &value
}

func dereference(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func isAllowedPhotoExtension(fileName string) bool {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
