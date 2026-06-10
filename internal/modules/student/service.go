package student

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	attendanceModule "absensi-cn-api/internal/modules/attendance"
	leaveModule "absensi-cn-api/internal/modules/leave"
	"absensi-cn-api/pkg/storage"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultCheckInStart = "06:30:00"
	defaultOnTimeUntil  = "07:00:00"
	defaultLateUntil    = "07:30:00"
	maxPhotoSizeBytes   = 5 * 1024 * 1024
)

var (
	ErrStudentDataUnavailable = errors.New("student database is not available")
	ErrStudentNotFound        = errors.New("student profile not found")
	ErrAttendanceNotOpen      = errors.New("attendance has not opened yet")
	ErrAlreadySubmittedToday  = errors.New("attendance for today has already been submitted")
	ErrPhotoRequired          = errors.New("attendance photo is required")
	ErrPhotoTooLarge          = errors.New("attendance photo must be 5MB or smaller")
	ErrPhotoInvalid           = errors.New("attendance photo must be jpg, jpeg, png, or webp")
	ErrReportTypeInvalid      = errors.New("report type must be HADIR, IZIN, or SAKIT")
	ErrReasonRequired         = errors.New("reason is required for izin or sakit")
)

type Service struct {
	db                *gorm.DB
	attendanceService *attendanceModule.Service
	photoUploader     *storage.PhotoUploader
}

type studentProfileRow struct {
	StudentID        string
	UserID           string
	Name             string
	NIS              string
	NISN             *string
	Gender           *string
	BirthPlace       *string
	BirthDate        *time.Time
	Address          *string
	Phone            *string
	ParentName       *string
	ParentPhone      *string
	EntryYear        int
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

type submissionRow struct {
	ID             string
	StudentID      string
	StudentName    string
	NIS            string
	ClassID        *string
	ClassName      *string
	Type           string
	Reason         string
	Attachment     string
	Status         string
	ReviewedBy     *string
	ReviewedByName *string
	ReviewNote     *string
	ReviewedAt     *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewService(db *gorm.DB, attendanceService *attendanceModule.Service, photoUploader *storage.PhotoUploader) *Service {
	return &Service{
		db:                db,
		attendanceService: attendanceService,
		photoUploader:     photoUploader,
	}
}

func (s *Service) GetDashboard(userID string) (*StudentDashboardResponse, error) {
	today, err := s.GetToday(userID)
	if err != nil {
		return nil, err
	}

	attendanceHistory, err := s.ListAttendanceHistory(userID)
	if err != nil {
		return nil, err
	}
	submissions, err := s.ListSubmissions(userID)
	if err != nil {
		return nil, err
	}

	stats := buildStudentStats(attendanceHistory, submissions)
	return &StudentDashboardResponse{
		Today:             *today,
		Stats:             stats,
		RecentAttendance:  limitAttendance(attendanceHistory, 6),
		RecentSubmissions: limitSubmissions(submissions, 5),
		Notifications:     buildNotifications(today, attendanceHistory, submissions),
	}, nil
}

func (s *Service) GetProfile(userID string) (*StudentProfileResponse, error) {
	row, err := s.findProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	profile := mapProfileRow(*row)
	return &profile, nil
}

func (s *Service) GetToday(userID string) (*StudentTodayResponse, error) {
	if s.db == nil {
		return nil, ErrStudentDataUnavailable
	}

	row, err := s.findProfileByUserID(userID)
	if err != nil {
		return nil, err
	}

	rule, err := s.findAttendanceRule(row.SchoolYearID)
	if err != nil {
		return nil, err
	}

	now := currentStudentTime()
	record, err := s.findAttendanceRecord(row.StudentID, now)
	if err != nil {
		return nil, err
	}

	hasSubmittedToday := isSubmittedAttendanceRecord(record)
	canSubmit := !hasSubmittedToday
	message := "Absensi hari ini masih bisa dikirim."
	if now.Before(studentMoment(now, rule.CheckInStart)) {
		message = "Absensi dibuka pukul " + rule.CheckInStart + "."
		canSubmit = false
	}
	if hasSubmittedToday && !canSubmit {
		message = "Absensi hari ini sudah dikirim. Kamu bisa absen lagi besok."
	}

	response := &StudentTodayResponse{
		Profile:       mapProfileRow(*row),
		Window:        mapWindow(rule),
		CanSubmit:     canSubmit,
		CooldownUntil: attendanceDateOnly(now).AddDate(0, 0, 1).Format(time.RFC3339),
		CurrentStatus: string(determineAttendanceStatus(now, rule.CheckInStart, rule.OnTimeUntil, rule.LateUntil)),
		CurrentTime:   now.Format(time.RFC3339),
		Message:       message,
	}

	if hasSubmittedToday && record != nil {
		mapped, err := s.mapAttendanceRecord(*record)
		if err != nil {
			return nil, err
		}
		response.Attendance = mapped
		response.CurrentStatus = mapped.Status
	}

	return response, nil
}

func (s *Service) GetHistory(userID string) (*StudentHistoryResponse, error) {
	profile, err := s.GetProfile(userID)
	if err != nil {
		return nil, err
	}
	attendanceHistory, err := s.ListAttendanceHistory(userID)
	if err != nil {
		return nil, err
	}
	submissions, err := s.ListSubmissions(userID)
	if err != nil {
		return nil, err
	}
	return &StudentHistoryResponse{
		Profile:     *profile,
		Stats:       buildStudentStats(attendanceHistory, submissions),
		Attendance:  attendanceHistory,
		Submissions: submissions,
	}, nil
}

func (s *Service) SubmitDailyReport(userID, reportType, reason string, photo *multipart.FileHeader) (*StudentTodayResponse, error) {
	if s.db == nil {
		return nil, ErrStudentDataUnavailable
	}
	if photo == nil {
		return nil, ErrPhotoRequired
	}
	if photo.Size > maxPhotoSizeBytes {
		return nil, ErrPhotoTooLarge
	}
	if !isAllowedPhotoExtension(photo.Filename) {
		return nil, ErrPhotoInvalid
	}

	row, err := s.findProfileByUserID(userID)
	if err != nil {
		return nil, err
	}

	rule, err := s.findAttendanceRule(row.SchoolYearID)
	if err != nil {
		return nil, err
	}

	normalizedType := normalizeReportType(reportType)
	if normalizedType == "" {
		return nil, ErrReportTypeInvalid
	}
	trimmedReason := strings.TrimSpace(reason)
	if (normalizedType == "IZIN" || normalizedType == "SAKIT") && trimmedReason == "" {
		return nil, ErrReasonRequired
	}

	now := currentStudentTime()
	if normalizedType == "HADIR" && now.Before(studentMoment(now, rule.CheckInStart)) {
		return nil, ErrAttendanceNotOpen
	}

	existing, err := s.findAttendanceRecord(row.StudentID, now)
	if err != nil {
		return nil, err
	}
	if isSubmittedAttendanceRecord(existing) {
		return nil, ErrAlreadySubmittedToday
	}

	photoURL, photoFilename, err := s.storePhoto(photo, now)
	if err != nil {
		return nil, err
	}

	status := determineAttendanceStatus(now, rule.CheckInStart, rule.OnTimeUntil, rule.LateUntil)
	if normalizedType == "IZIN" {
		status = attendanceModule.StatusIzin
	}
	if normalizedType == "SAKIT" {
		status = attendanceModule.StatusSakit
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		checkInAt := &now
		if normalizedType == "IZIN" || normalizedType == "SAKIT" {
			checkInAt = nil
		}

		if existing != nil {
			existing.StudentClassMembershipID = row.MembershipID
			existing.ClassID = row.ClassID
			existing.SchoolYearID = row.SchoolYearID
			existing.Status = status
			existing.CheckInAt = checkInAt
			existing.PhotoURL = stringPtr(photoURL)
			existing.PhotoFilename = stringPtr(photoFilename)
			existing.Notes = optionalString(buildAttendanceNote(normalizedType, trimmedReason))
			existing.VerifiedBy = nil
			existing.VerifiedAt = nil
			existing.VerificationNote = nil
			if err := tx.Save(existing).Error; err != nil {
				return fmt.Errorf("update student attendance report: %w", err)
			}
		} else {
			record := attendanceModule.AttendanceRecord{
				ID:                       uuid.NewString(),
				StudentID:                row.StudentID,
				StudentClassMembershipID: row.MembershipID,
				ClassID:                  row.ClassID,
				SchoolYearID:             row.SchoolYearID,
				AttendanceDate:           attendanceDateOnly(now),
				CheckInAt:                checkInAt,
				Status:                   status,
				PhotoURL:                 stringPtr(photoURL),
				PhotoFilename:            stringPtr(photoFilename),
				Notes:                    optionalString(buildAttendanceNote(normalizedType, trimmedReason)),
			}
			if err := tx.Create(&record).Error; err != nil {
				var mysqlErr *mysql.MySQLError
				if !errors.As(err, &mysqlErr) || mysqlErr.Number != 1062 {
					return fmt.Errorf("create student attendance report: %w", err)
				}
				// A concurrent alpha-sync inserted a placeholder between our pre-check
				// and this insert. Fetch it outside the transaction — REPEATABLE READ
				// isolation prevents tx from seeing rows inserted by other transactions.
				var conflict attendanceModule.AttendanceRecord
				if fetchErr := s.db.Where("student_id = ? AND attendance_date = ?",
					row.StudentID,
					attendanceDateOnly(now).Format("2006-01-02"),
				).First(&conflict).Error; fetchErr != nil {
					return fmt.Errorf("fetch conflicting attendance record: %w", fetchErr)
				}
				if isSubmittedAttendanceRecord(&conflict) {
					return ErrAlreadySubmittedToday
				}
				conflict.StudentClassMembershipID = row.MembershipID
				conflict.ClassID = row.ClassID
				conflict.SchoolYearID = row.SchoolYearID
				conflict.Status = status
				conflict.CheckInAt = checkInAt
				conflict.PhotoURL = stringPtr(photoURL)
				conflict.PhotoFilename = stringPtr(photoFilename)
				conflict.Notes = optionalString(buildAttendanceNote(normalizedType, trimmedReason))
				conflict.VerifiedBy = nil
				conflict.VerifiedAt = nil
				conflict.VerificationNote = nil
				if updateErr := tx.Save(&conflict).Error; updateErr != nil {
					return fmt.Errorf("update conflicting attendance record: %w", updateErr)
				}
			}
		}

		if normalizedType == "IZIN" || normalizedType == "SAKIT" {
			submission := leaveModule.Submission{
				ID:         uuid.NewString(),
				StudentID:  row.StudentID,
				Type:       normalizedType,
				Reason:     trimmedReason,
				Attachment: photoURL,
				Status:     leaveModule.StatusPending,
			}
			if err := tx.Create(&submission).Error; err != nil {
				return fmt.Errorf("create student leave submission: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.GetToday(userID)
}

func (s *Service) ListAttendanceHistory(userID string) ([]attendanceModule.AttendanceRecordResponse, error) {
	if s.attendanceService == nil {
		return nil, ErrStudentDataUnavailable
	}
	return s.attendanceService.ListStudentAttendanceHistory(userID)
}

func (s *Service) ListSubmissions(userID string) ([]StudentSubmissionResponse, error) {
	row, err := s.findProfileByUserID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.querySubmissions(row.StudentID)
	if err != nil {
		return nil, err
	}
	return mapSubmissions(rows), nil
}

func (s *Service) findProfileByUserID(userID string) (*studentProfileRow, error) {
	var row studentProfileRow
	if err := s.db.Table("students").
		Select("students.id as student_id, students.user_id, users.name, students.nis, students.nisn, students.gender, students.birth_place, students.birth_date, students.address, students.phone, students.parent_name, students.parent_phone, students.entry_year, students.is_active as student_active, scm.id as membership_id, scm.status as membership_status, scm.class_id, classes.grade, classes.name as class_label, majors.code as major_code, majors.name as major_name, scm.school_year_id, school_years.name as school_year_name").
		Joins("join users on users.id = students.user_id").
		Joins("left join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("left join classes on classes.id = scm.class_id").
		Joins("left join majors on majors.id = classes.major_id").
		Joins("left join school_years on school_years.id = scm.school_year_id").
		Where("students.user_id = ? AND students.is_active = ?", userID, true).
		Order("school_years.start_year desc").
		Limit(1).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("find student profile: %w", err)
	}
	if row.StudentID == "" {
		return nil, ErrStudentNotFound
	}
	return &row, nil
}

func (s *Service) findAttendanceRule(schoolYearID string) (*attendanceModule.AttendanceRule, error) {
	var rule attendanceModule.AttendanceRule
	err := s.db.Where("school_year_id = ? AND is_active = ?", schoolYearID, true).First(&rule).Error
	if err == nil {
		return &rule, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("find student attendance rule: %w", err)
	}
	return &attendanceModule.AttendanceRule{
		SchoolYearID: schoolYearID,
		CheckInStart: defaultCheckInStart,
		OnTimeUntil:  defaultOnTimeUntil,
		LateUntil:    defaultLateUntil,
		IsActive:     true,
	}, nil
}

func (s *Service) findAttendanceRecord(studentID string, date time.Time) (*attendanceModule.AttendanceRecord, error) {
	var record attendanceModule.AttendanceRecord
	err := s.db.Where("student_id = ? AND attendance_date = ?", studentID, attendanceDateOnly(date).Format("2006-01-02")).First(&record).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find student attendance record: %w", err)
	}
	return &record, nil
}

func (s *Service) mapAttendanceRecord(record attendanceModule.AttendanceRecord) (*attendanceModule.AttendanceRecordResponse, error) {
	records, err := s.attendanceService.ListAdminAttendance(record.AttendanceDate.Format("2006-01-02"), "", record.ClassID)
	if err != nil {
		return nil, err
	}
	for _, item := range records {
		if item.ID == record.ID {
			return &item, nil
		}
	}
	return nil, fmt.Errorf("attendance record not found")
}

func (s *Service) querySubmissions(studentID string) ([]submissionRow, error) {
	var rows []submissionRow
	if err := s.db.Table("submissions").
		Select("submissions.id, submissions.student_id, users.name as student_name, students.nis, classes.id as class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, submissions.type, submissions.reason, submissions.attachment, submissions.status, submissions.reviewed_by, reviewer.name as reviewed_by_name, submissions.review_note, submissions.reviewed_at, submissions.created_at, submissions.updated_at").
		Joins("join students on students.id = submissions.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("left join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("left join classes on classes.id = scm.class_id").
		Joins("left join majors on majors.id = classes.major_id").
		Joins("left join users as reviewer on reviewer.id = submissions.reviewed_by").
		Where("submissions.student_id = ?", studentID).
		Order("submissions.created_at desc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list student submissions: %w", err)
	}
	return rows, nil
}

func (s *Service) storePhoto(photo *multipart.FileHeader, submittedAt time.Time) (string, string, error) {
	if s.photoUploader == nil {
		return "", "", fmt.Errorf("photo uploader is not configured")
	}
	photoURL, fileName, err := s.photoUploader.Store(context.Background(), photo, "student-attendance", submittedAt)
	if err != nil {
		return "", "", fmt.Errorf("store student attendance photo: %w", err)
	}
	return photoURL, fileName, nil
}

func mapProfileRow(row studentProfileRow) StudentProfileResponse {
	birthPlace := dereference(row.BirthPlace)
	birthDate := formatDate(row.BirthDate)
	return StudentProfileResponse{
		ID:               row.StudentID,
		UserID:           row.UserID,
		Name:             row.Name,
		NIS:              row.NIS,
		NISN:             dereference(row.NISN),
		Gender:           dereference(row.Gender),
		BirthPlace:       birthPlace,
		BirthDate:        birthDate,
		BirthPlaceDate:   formatBirthPlaceDate(birthPlace, row.BirthDate),
		Address:          dereference(row.Address),
		Phone:            dereference(row.Phone),
		ParentName:       dereference(row.ParentName),
		ParentPhone:      dereference(row.ParentPhone),
		EntryYear:        row.EntryYear,
		IsActive:         row.StudentActive,
		ClassID:          row.ClassID,
		ClassName:        strings.TrimSpace(fmt.Sprintf("%s %s %s", row.Grade, row.MajorCode, row.ClassLabel)),
		MajorCode:        row.MajorCode,
		MajorName:        row.MajorName,
		SchoolYearID:     row.SchoolYearID,
		SchoolYearName:   row.SchoolYearName,
		MembershipID:     row.MembershipID,
		MembershipStatus: row.MembershipStatus,
	}
}

func formatBirthPlaceDate(place string, date *time.Time) string {
	trimmedPlace := strings.TrimSpace(place)
	if date == nil {
		return trimmedPlace
	}
	formattedDate := formatIndonesianDate(*date)
	if trimmedPlace == "" {
		return formattedDate
	}
	return fmt.Sprintf("%s, %s", trimmedPlace, formattedDate)
}

func formatIndonesianDate(value time.Time) string {
	months := []string{
		"Januari", "Februari", "Maret", "April", "Mei", "Juni",
		"Juli", "Agustus", "September", "Oktober", "November", "Desember",
	}
	month := months[int(value.Month())-1]
	return fmt.Sprintf("%d %s %d", value.Day(), month, value.Year())
}

func mapWindow(rule *attendanceModule.AttendanceRule) AttendanceWindowResponse {
	return AttendanceWindowResponse{
		CheckInStart: rule.CheckInStart,
		OnTimeUntil:  rule.OnTimeUntil,
		LateUntil:    rule.LateUntil,
	}
}

func mapSubmissions(rows []submissionRow) []StudentSubmissionResponse {
	result := make([]StudentSubmissionResponse, 0, len(rows))
	for _, row := range rows {
		item := StudentSubmissionResponse{
			ID:          row.ID,
			StudentID:   row.StudentID,
			StudentName: row.StudentName,
			NIS:         row.NIS,
			Type:        row.Type,
			Reason:      row.Reason,
			Attachment:  row.Attachment,
			Status:      row.Status,
			CreatedAt:   row.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   row.UpdatedAt.Format(time.RFC3339),
		}
		if row.ClassID != nil {
			item.ClassID = *row.ClassID
		}
		if row.ClassName != nil {
			item.ClassName = *row.ClassName
		}
		if row.ReviewedBy != nil {
			item.ReviewedBy = *row.ReviewedBy
		}
		if row.ReviewedByName != nil {
			item.ReviewedByName = *row.ReviewedByName
		}
		if row.ReviewNote != nil {
			item.ReviewNote = *row.ReviewNote
		}
		if row.ReviewedAt != nil {
			item.ReviewedAt = row.ReviewedAt.Format(time.RFC3339)
		}
		result = append(result, item)
	}
	return result
}

func buildStudentStats(attendanceHistory []attendanceModule.AttendanceRecordResponse, submissions []StudentSubmissionResponse) StudentStatsResponse {
	stats := StudentStatsResponse{TotalAttendance: len(attendanceHistory)}
	for _, item := range attendanceHistory {
		switch strings.ToLower(item.Status) {
		case string(attendanceModule.StatusHadir):
			stats.Present++
		case string(attendanceModule.StatusTelat):
			stats.Late++
		case string(attendanceModule.StatusIzin):
			stats.Permission++
		case string(attendanceModule.StatusSakit):
			stats.Sick++
		case string(attendanceModule.StatusAlfa):
			stats.Alpha++
		}
		if strings.TrimSpace(item.VerifiedAt) == "" {
			stats.PendingReviews++
		}
	}
	for _, item := range submissions {
		if strings.ToLower(item.Status) == string(leaveModule.StatusPending) {
			stats.PendingRequests++
		}
	}
	return stats
}

func buildNotifications(today *StudentTodayResponse, attendanceHistory []attendanceModule.AttendanceRecordResponse, submissions []StudentSubmissionResponse) []StudentNotificationResponse {
	notifications := make([]StudentNotificationResponse, 0, 5)
	tone := "emerald"
	title := "Absensi siap dikirim"
	description := today.Message
	if !today.CanSubmit {
		tone = "slate"
		title = "Absensi hari ini terkunci"
	}
	notifications = append(notifications, StudentNotificationResponse{
		ID:          "today-attendance",
		Title:       title,
		Description: description,
		Tone:        tone,
		CreatedAt:   today.CurrentTime,
	})

	for _, item := range submissions {
		if strings.ToLower(item.Status) == string(leaveModule.StatusPending) {
			notifications = append(notifications, StudentNotificationResponse{
				ID:          "submission-" + item.ID,
				Title:       "Pengajuan menunggu review",
				Description: fmt.Sprintf("%s kamu sedang menunggu validasi walas atau BK.", formatDisplayLabel(item.Type)),
				Tone:        "amber",
				CreatedAt:   item.CreatedAt,
			})
			break
		}
	}

	for _, item := range attendanceHistory {
		if strings.TrimSpace(item.VerificationNote) != "" {
			notifications = append(notifications, StudentNotificationResponse{
				ID:          "attendance-note-" + item.ID,
				Title:       "Catatan validasi terbaru",
				Description: item.VerificationNote,
				Tone:        "sky",
				CreatedAt:   item.VerifiedAt,
			})
			break
		}
	}
	return notifications
}

func limitAttendance(records []attendanceModule.AttendanceRecordResponse, limit int) []attendanceModule.AttendanceRecordResponse {
	if len(records) <= limit {
		return records
	}
	return records[:limit]
}

func limitSubmissions(records []StudentSubmissionResponse, limit int) []StudentSubmissionResponse {
	if len(records) <= limit {
		return records
	}
	return records[:limit]
}

func normalizeReportType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "HADIR", "IZIN", "SAKIT":
		return strings.ToUpper(strings.TrimSpace(value))
	default:
		return ""
	}
}

func determineAttendanceStatus(now time.Time, checkInStart, onTimeUntil, lateUntil string) attendanceModule.AttendanceStatus {
	if now.Before(studentMoment(now, checkInStart)) {
		return attendanceModule.StatusAlfa
	}
	if now.Before(studentMoment(now, onTimeUntil)) {
		return attendanceModule.StatusHadir
	}
	if now.Before(studentMoment(now, lateUntil)) || now.Equal(studentMoment(now, lateUntil)) {
		return attendanceModule.StatusTelat
	}
	return attendanceModule.StatusAlfa
}

func isSubmittedAttendanceRecord(record *attendanceModule.AttendanceRecord) bool {
	if record == nil {
		return false
	}
	return record.PhotoURL != nil || record.CheckInAt != nil || record.Status != attendanceModule.StatusAlfa
}

func buildAttendanceNote(reportType, reason string) string {
	if reportType == "HADIR" {
		return "Absen mandiri siswa"
	}
	return reason
}

func studentMoment(base time.Time, clockValue string) time.Time {
	parsed, err := time.ParseInLocation("15:04:05", clockValue, studentLocation())
	if err != nil {
		parsed, _ = time.ParseInLocation("15:04", clockValue, studentLocation())
	}
	current := base.In(studentLocation())
	return time.Date(current.Year(), current.Month(), current.Day(), parsed.Hour(), parsed.Minute(), parsed.Second(), 0, studentLocation())
}

func attendanceDateOnly(value time.Time) time.Time {
	current := value.In(studentLocation())
	return time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, studentLocation())
}

func currentStudentTime() time.Time {
	return time.Now().In(studentLocation())
}

func studentLocation() *time.Location {
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

func formatDate(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02")
}

func formatDisplayLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(strings.ReplaceAll(trimmed, "_", " "))
	return strings.ToUpper(lower[:1]) + lower[1:]
}

func isAllowedPhotoExtension(fileName string) bool {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
