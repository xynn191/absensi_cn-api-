package staff

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/internal/modules/counseling"
	leaveModule "absensi-cn-api/internal/modules/leave"
	studentModule "absensi-cn-api/internal/modules/student"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrStaffDataUnavailable   = errors.New("staff database is not available")
	ErrTeacherProfileNotFound = errors.New("teacher profile not found")
	ErrHomeroomNotFound       = errors.New("active homeroom assignment not found")
	ErrStudentNotFound        = errors.New("student not found")
	ErrSubmissionNotFound     = errors.New("submission not found")
	ErrCounselingNoteNotFound = errors.New("counseling note not found")
)

type Service struct {
	db                *gorm.DB
	attendanceService *attendance.Service
}

type teacherMeRow struct {
	UserID    string
	TeacherID string
	Name      string
	Username  *string
	NIP       *string
	NUPTK     *string
	Gender    *string
	Phone     *string
}

type subjectAssignmentRow struct {
	ID             string
	SubjectID      string
	SubjectCode    string
	SubjectName    string
	ClassID        string
	ClassName      string
	SchoolYearID   string
	SchoolYearName string
	IsActive       bool
}

type homeroomRow struct {
	AssignmentID   string
	TeacherID      string
	ClassID        string
	ClassName      string
	SchoolYearID   string
	SchoolYearName string
	IsActive       bool
}

type studentSummaryRow struct {
	ID               string
	UserID           string
	Name             string
	NIS              string
	NISN             *string
	Gender           *string
	Phone            *string
	ClassID          *string
	ClassName        *string
	SchoolYearID     *string
	SchoolYearName   *string
	MembershipID     *string
	MembershipStatus *string
	EntryYear        int
	IsActive         bool
	PresentCount     int
	LateCount        int
	PermissionCount  int
	SickCount        int
	AlphaCount       int
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

type counselingNoteRow struct {
	ID            string
	StudentID     string
	StudentName   string
	NIS           string
	ClassID       *string
	ClassName     *string
	CreatedBy     string
	CreatedByName *string
	Title         string
	Note          string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type riskStudentRow struct {
	StudentID   string
	StudentName string
	NIS         string
	ClassName   string
	Occurrences int
}

type bkClassSummaryRow struct {
	ClassID        string
	ClassName      string
	SchoolYearID   *string
	SchoolYearName *string
}

func NewService(db *gorm.DB, attendanceService *attendance.Service) *Service {
	return &Service{
		db:                db,
		attendanceService: attendanceService,
	}
}

func (s *Service) GetTeacherMe(userID string) (*TeacherMeResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	profile, err := s.findTeacherMeByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := &TeacherMeResponse{
		UserID:    profile.UserID,
		TeacherID: profile.TeacherID,
		Name:      profile.Name,
		Username:  dereference(profile.Username),
		NIP:       dereference(profile.NIP),
		NUPTK:     dereference(profile.NUPTK),
		Gender:    dereference(profile.Gender),
		Phone:     dereference(profile.Phone),
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err == nil {
		mapped := mapHomeroomRow(*homeroom)
		response.IsHomeroomTeacher = true
		response.ActiveHomeroom = &mapped
		return response, nil
	}
	if !errors.Is(err, ErrHomeroomNotFound) {
		return nil, err
	}

	return response, nil
}

func (s *Service) ListTeacherSubjectAssignments(userID string) ([]TeacherSubjectAssignmentResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	rows, err := s.queryTeacherSubjectAssignments(userID)
	if err != nil {
		return nil, err
	}

	result := make([]TeacherSubjectAssignmentResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, TeacherSubjectAssignmentResponse{
			ID:             row.ID,
			SubjectID:      row.SubjectID,
			SubjectCode:    row.SubjectCode,
			SubjectName:    row.SubjectName,
			ClassID:        row.ClassID,
			ClassName:      row.ClassName,
			SchoolYearID:   row.SchoolYearID,
			SchoolYearName: row.SchoolYearName,
			IsActive:       row.IsActive,
		})
	}

	return result, nil
}

func (s *Service) GetTeacherHomeroom(userID string) (*HomeroomContextResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	row, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	response := mapHomeroomRow(*row)
	return &response, nil
}

func (s *Service) ListHomeroomStudents(userID string) ([]StudentSummaryResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queryStudentSummaries(homeroom.ClassID, "")
	if err != nil {
		return nil, err
	}

	return mapStudentSummaries(rows), nil
}

func (s *Service) GetHomeroomDashboard(userID string) (*HomeroomDashboardResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	students, err := s.queryStudentSummaries(homeroom.ClassID, "")
	if err != nil {
		return nil, err
	}

	records, err := s.attendanceService.ListAdminAttendance("", "", homeroom.ClassID)
	if err != nil {
		return nil, err
	}

	todaySummary := summarizeAttendance(records)
	recentSubmissions, err := s.querySubmissions(homeroom.ClassID, "", "")
	if err != nil {
		return nil, err
	}

	return &HomeroomDashboardResponse{
		Homeroom:                 mapHomeroomRow(*homeroom),
		TotalStudents:            len(students),
		Today:                    todaySummary,
		StudentsNeedingAttention: buildAttentionStudents(records),
		RecentSubmissions:        limitSubmissions(mapSubmissions(recentSubmissions), 5),
	}, nil
}

func (s *Service) ListHomeroomAttendance(userID, date, status string) ([]attendance.AttendanceRecordResponse, error) {
	if s.db == nil { 
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	return s.attendanceService.ListAdminAttendance(date, status, homeroom.ClassID)
}

func (s *Service) GetHomeroomAttendanceOverview(userID, date, status, query string) (*HomeroomAttendanceOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	summaryRecords, err := s.attendanceService.ListAdminAttendance(date, "", homeroom.ClassID)
	if err != nil {
		return nil, err
	}

	filteredRecords, err := s.attendanceService.ListAdminAttendance(date, status, homeroom.ClassID)
	if err != nil {
		return nil, err
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery != "" {
		filteredRecords = filterAttendanceRecords(filteredRecords, normalizedQuery)
	}

	targetDate := currentHomeroomDate(date)
	return &HomeroomAttendanceOverviewResponse{
		Homeroom:     mapHomeroomRow(*homeroom),
		Date:         targetDate,
		StatusFilter: strings.TrimSpace(status),
		Query:        strings.TrimSpace(query),
		Summary:      summarizeAttendance(summaryRecords),
		Records:      filteredRecords,
	}, nil
}

func (s *Service) GetHomeroomAttendanceSummary(userID, date string) (*AttendanceSummaryResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	records, err := s.attendanceService.ListAdminAttendance(date, "", homeroom.ClassID)
	if err != nil {
		return nil, err
	}

	summary := summarizeAttendance(records)
	return &summary, nil
}

func (s *Service) ReviewHomeroomAttendance(userID, attendanceID, actorID string, input attendance.ReviewAttendanceRequest) (*attendance.AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	recordClassID, err := s.findAttendanceRecordClassID(attendanceID)
	if err != nil {
		return nil, err
	}
	if recordClassID != homeroom.ClassID {
		return nil, ErrStudentNotFound
	}

	return s.attendanceService.ReviewAttendance(attendanceID, actorID, input)
}

func (s *Service) ListHomeroomStudentAttendanceHistory(userID, studentID string) ([]attendance.AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	studentUserID, err := s.ensureHomeroomStudent(userID, studentID)
	if err != nil {
		return nil, err
	}

	return s.attendanceService.ListStudentAttendanceHistory(studentUserID)
}

func (s *Service) GetHomeroomStudentDetail(userID, studentID string) (*HomeroomStudentDetailResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	_, err := s.ensureHomeroomStudent(userID, studentID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queryStudentSummaries("", studentID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrStudentNotFound
	}

	attendanceHistory, err := s.ListHomeroomStudentAttendanceHistory(userID, studentID)
	if err != nil {
		return nil, err
	}

	submissionRows, err := s.querySubmissions("", studentID, "")
	if err != nil {
		return nil, err
	}

	summary := summarizeAttendance(attendanceHistory)

	return &HomeroomStudentDetailResponse{
		Student:           mapStudentSummaryRow(rows[0]),
		AttendanceSummary: summary,
		RecentAttendance:  limitAttendanceRecords(attendanceHistory, 20),
		RecentSubmissions: limitSubmissions(mapSubmissions(submissionRows), 10),
	}, nil
}

func (s *Service) ListHomeroomSubmissions(userID, status string) ([]SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.querySubmissions(homeroom.ClassID, "", status)
	if err != nil {
		return nil, err
	}

	return mapSubmissions(rows), nil
}

func (s *Service) GetHomeroomSubmissionsOverview(userID, status, submissionType, query string) (*HomeroomSubmissionOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.querySubmissions(homeroom.ClassID, "", "")
	if err != nil {
		return nil, err
	}

	filteredRows := filterSubmissionRows(rows, status, submissionType, query)

	return &HomeroomSubmissionOverviewResponse{
		Homeroom:     mapHomeroomRow(*homeroom),
		StatusFilter: strings.TrimSpace(status),
		TypeFilter:   strings.ToUpper(strings.TrimSpace(submissionType)),
		Query:        strings.TrimSpace(query),
		Counts:       summarizeSubmissions(rows),
		Records:      mapSubmissions(filteredRows),
	}, nil
}

func (s *Service) ReviewHomeroomSubmission(userID, submissionID, actorID string, input ReviewSubmissionRequest) (*SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return nil, err
	}

	row, err := s.findSubmissionResponseByID(submissionID)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(row.ClassID) == "" || row.ClassID != homeroom.ClassID {
		return nil, ErrSubmissionNotFound
	}

	return s.reviewSubmission(submissionID, actorID, input)
}

func (s *Service) ListStudentSubmissions(userID string) ([]SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	studentID, err := s.findStudentIDByUserID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := s.querySubmissions("", studentID, "")
	if err != nil {
		return nil, err
	}

	return mapSubmissions(rows), nil
}

func (s *Service) CreateStudentSubmission(userID string, input CreateSubmissionRequest) (*SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	context, err := s.attendanceService.GetStudentContextByUserID(userID)
	if err != nil {
		return nil, err
	}

	submissionType := normalizeSubmissionType(input.Type)
	if submissionType == "" {
		return nil, fmt.Errorf("submission type must be one of IZIN, SAKIT, or DISPENSASI")
	}
	if strings.TrimSpace(input.Reason) == "" {
		return nil, fmt.Errorf("reason is required")
	}

	record := leaveModule.Submission{
		ID:         uuid.NewString(),
		StudentID:  context.ID,
		Type:       submissionType,
		Reason:     strings.TrimSpace(input.Reason),
		Attachment: strings.TrimSpace(input.Attachment),
		Status:     leaveModule.StatusPending,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	return s.findSubmissionResponseByID(record.ID)
}

func (s *Service) ListBKStudents(classID string) ([]StudentSummaryResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	rows, err := s.queryStudentSummaries(strings.TrimSpace(classID), "")
	if err != nil {
		return nil, err
	}

	return mapStudentSummaries(rows), nil
}

func (s *Service) GetBKDashboard() (*BKDashboardResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	students, err := s.queryStudentSummaries("", "")
	if err != nil {
		return nil, err
	}

	todayDate := currentHomeroomDate("")
	todayRecords, err := s.attendanceService.ListAdminAttendance(todayDate, "", "")
	if err != nil {
		return nil, err
	}

	counselingRows, err := s.queryCounselingNotes("", "", "")
	if err != nil {
		return nil, err
	}

	submissionRows, err := s.querySubmissions("", "", "")
	if err != nil {
		return nil, err
	}

	classRows, err := s.queryBKClasses()
	if err != nil {
		return nil, err
	}

	todaySummary := summarizeAttendance(todayRecords)
	repeatedLate, err := s.queryRiskStudents(string(attendance.StatusTelat), "", 30, 3)
	if err != nil {
		return nil, err
	}
	repeatedAlpha, err := s.queryRiskStudents(string(attendance.StatusAlfa), "", 30, 3)
	if err != nil {
		return nil, err
	}
	todaySummary.RepeatedLate = mapRiskStudents(repeatedLate)
	todaySummary.RepeatedAlpha = mapRiskStudents(repeatedAlpha)

	return &BKDashboardResponse{
		TotalStudents:         len(students),
		StudentsNeedAttention: countStudentSummaryRowsNeedingAttention(students),
		TotalCounselingNotes:  len(counselingRows),
		PendingSubmissions:    summarizeSubmissions(submissionRows).Pending,
		Today:                 todaySummary,
		TopRiskStudents:       buildBKDashboardRiskStudents(todaySummary, 6),
		RecentSubmissions:     limitSubmissions(mapSubmissions(submissionRows), 5),
		RecentCounselingNotes: limitCounselingNotes(mapCounselingNotes(counselingRows), 5),
		Classes:               mapBKClassSummaries(classRows),
	}, nil
}

func (s *Service) GetBKStudentsOverview(classID, riskFilter, query string) (*BKStudentsOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	rows, err := s.queryStudentSummaries(strings.TrimSpace(classID), "")
	if err != nil {
		return nil, err
	}

	counselingRows, err := s.queryCounselingNotes(strings.TrimSpace(classID), "", "")
	if err != nil {
		return nil, err
	}

	classRows, err := s.queryBKClasses()
	if err != nil {
		return nil, err
	}

	noteStudentIDs := collectUniqueStudentIDsFromNotes(counselingRows)
	filteredRows := filterStudentSummaryRows(rows, strings.TrimSpace(riskFilter), strings.TrimSpace(query), noteStudentIDs)

	return &BKStudentsOverviewResponse{
		ClassFilter: strings.TrimSpace(classID),
		RiskFilter:  strings.TrimSpace(riskFilter),
		Query:       strings.TrimSpace(query),
		Counts: StudentOverviewCounts{
			Total:               len(rows),
			Active:              countActiveStudents(rows),
			NeedAttention:       countStudentSummaryRowsNeedingAttention(rows),
			TotalLate:           sumStudentSummaryField(rows, func(row studentSummaryRow) int { return row.LateCount }),
			TotalAlpha:          sumStudentSummaryField(rows, func(row studentSummaryRow) int { return row.AlphaCount }),
			WithCounselingNotes: len(noteStudentIDs),
		},
		Students: mapStudentSummaries(filteredRows),
		Classes:  mapBKClassSummaries(classRows),
	}, nil
}

func (s *Service) GetBKStudent(studentID string) (*BKStudentDetailResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	rows, err := s.queryStudentSummaries("", studentID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, ErrStudentNotFound
	}

	attendanceHistory, err := s.ListBKStudentAttendanceHistory(studentID)
	if err != nil {
		return nil, err
	}

	submissionRows, err := s.querySubmissions("", studentID, "")
	if err != nil {
		return nil, err
	}

	counselingRows, err := s.queryCounselingNotes("", studentID, "")
	if err != nil {
		return nil, err
	}

	return &BKStudentDetailResponse{
		Student:           mapStudentSummaryRow(rows[0]),
		AttendanceSummary: summarizeAttendance(attendanceHistory),
		RecentAttendance:  limitAttendanceRecords(attendanceHistory, 20),
		RecentSubmissions: limitSubmissions(mapSubmissions(submissionRows), 10),
		CounselingNotes:   limitCounselingNotes(mapCounselingNotes(counselingRows), 12),
	}, nil
}

func (s *Service) ListBKAttendance(date, status, classID string) ([]attendance.AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	return s.attendanceService.ListAdminAttendance(date, status, classID)
}

func (s *Service) GetBKAttendanceSummary(date string) (*AttendanceSummaryResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	records, err := s.attendanceService.ListAdminAttendance(date, "", "")
	if err != nil {
		return nil, err
	}

	summary := &AttendanceSummaryResponse{}
	for _, record := range records {
		switch record.Status {
		case string(attendance.StatusHadir):
			summary.Present++
		case string(attendance.StatusTelat):
			summary.Late++
		case string(attendance.StatusIzin):
			summary.Permission++
		case string(attendance.StatusSakit):
			summary.Sick++
		case string(attendance.StatusAlfa):
			summary.Alpha++
		}
	}

	repeatedLate, err := s.queryRiskStudents(string(attendance.StatusTelat), "", 30, 3)
	if err != nil {
		return nil, err
	}
	repeatedAlpha, err := s.queryRiskStudents(string(attendance.StatusAlfa), "", 30, 3)
	if err != nil {
		return nil, err
	}

	summary.RepeatedLate = mapRiskStudents(repeatedLate)
	summary.RepeatedAlpha = mapRiskStudents(repeatedAlpha)
	return summary, nil
}

func (s *Service) GetBKAttendanceOverview(date, status, classID, query string) (*BKAttendanceOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	trimmedClassID := strings.TrimSpace(classID)
	summaryRecords, err := s.attendanceService.ListAdminAttendance(date, "", trimmedClassID)
	if err != nil {
		return nil, err
	}

	filteredRecords, err := s.attendanceService.ListAdminAttendance(date, status, trimmedClassID)
	if err != nil {
		return nil, err
	}

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	if normalizedQuery != "" {
		filteredRecords = filterAttendanceRecords(filteredRecords, normalizedQuery)
	}

	classRows, err := s.queryBKClasses()
	if err != nil {
		return nil, err
	}

	summary := summarizeAttendance(summaryRecords)
	repeatedLate, err := s.queryRiskStudents(string(attendance.StatusTelat), trimmedClassID, 30, 3)
	if err != nil {
		return nil, err
	}
	repeatedAlpha, err := s.queryRiskStudents(string(attendance.StatusAlfa), trimmedClassID, 30, 3)
	if err != nil {
		return nil, err
	}
	summary.RepeatedLate = mapRiskStudents(repeatedLate)
	summary.RepeatedAlpha = mapRiskStudents(repeatedAlpha)

	return &BKAttendanceOverviewResponse{
		Date:         currentHomeroomDate(date),
		StatusFilter: strings.TrimSpace(status),
		ClassFilter:  trimmedClassID,
		Query:        strings.TrimSpace(query),
		Summary:      summary,
		Records:      filteredRecords,
		Classes:      mapBKClassSummaries(classRows),
	}, nil
}

func (s *Service) ReviewBKAttendance(attendanceID, actorID string, input attendance.ReviewAttendanceRequest) (*attendance.AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	return s.attendanceService.ReviewAttendance(attendanceID, actorID, input)
}

func (s *Service) ListBKStudentAttendanceHistory(studentID string) ([]attendance.AttendanceRecordResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	studentUserID, err := s.findStudentUserIDByStudentID(studentID)
	if err != nil {
		return nil, err
	}

	return s.attendanceService.ListStudentAttendanceHistory(studentUserID)
}

func (s *Service) ListBKStudentCounselingNotes(studentID string) ([]CounselingNoteResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	if _, err := s.findStudentUserIDByStudentID(studentID); err != nil {
		return nil, err
	}

	rows, err := s.queryCounselingNotes("", studentID, "")
	if err != nil {
		return nil, err
	}

	return mapCounselingNotes(rows), nil
}

func (s *Service) GetBKCounselingOverview(classID, studentID, query string) (*BKCounselingOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	trimmedClassID := strings.TrimSpace(classID)
	trimmedStudentID := strings.TrimSpace(studentID)
	trimmedQuery := strings.TrimSpace(query)

	rows, err := s.queryCounselingNotes(trimmedClassID, trimmedStudentID, trimmedQuery)
	if err != nil {
		return nil, err
	}

	classRows, err := s.queryBKClasses()
	if err != nil {
		return nil, err
	}

	students, err := s.queryStudentSummaries(trimmedClassID, "")
	if err != nil {
		return nil, err
	}

	return &BKCounselingOverviewResponse{
		ClassFilter: trimmedClassID,
		StudentID:   trimmedStudentID,
		Query:       trimmedQuery,
		Counts: CounselingOverviewCount{
			TotalNotes:      len(rows),
			StudentsCovered: countUniqueStudentIDsInNotes(rows),
			ClassesCovered:  countUniqueClassIDsInNotes(rows),
			RecentWeekNotes: countRecentWeekNotes(rows),
		},
		Records:  mapCounselingNotes(rows),
		Classes:  mapBKClassSummaries(classRows),
		Students: mapStudentSummaries(students),
	}, nil
}

func (s *Service) CreateBKStudentCounselingNote(studentID, actorID string, input UpsertCounselingNoteRequest) (*CounselingNoteResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	if _, err := s.findStudentUserIDByStudentID(studentID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(input.Note) == "" {
		return nil, fmt.Errorf("note is required")
	}

	record := counseling.Note{
		ID:        uuid.NewString(),
		StudentID: studentID,
		CreatedBy: actorID,
		Title:     strings.TrimSpace(input.Title),
		Note:      strings.TrimSpace(input.Note),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create counseling note: %w", err)
	}

	return s.findCounselingNoteResponseByID(record.ID)
}

func (s *Service) UpdateBKCounselingNote(id string, input UpsertCounselingNoteRequest) (*CounselingNoteResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	record, err := s.findCounselingNoteByID(id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(input.Note) == "" {
		return nil, fmt.Errorf("note is required")
	}

	record.Title = strings.TrimSpace(input.Title)
	record.Note = strings.TrimSpace(input.Note)
	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update counseling note: %w", err)
	}

	return s.findCounselingNoteResponseByID(record.ID)
}

func (s *Service) DeleteBKCounselingNote(id string) error {
	if s.db == nil {
		return ErrStaffDataUnavailable
	}

	record, err := s.findCounselingNoteByID(id)
	if err != nil {
		return err
	}

	if err := s.db.Delete(record).Error; err != nil {
		return fmt.Errorf("delete counseling note: %w", err)
	}
	return nil
}

func (s *Service) ListBKSubmissions(status, classID string) ([]SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	rows, err := s.querySubmissions(strings.TrimSpace(classID), "", status)
	if err != nil {
		return nil, err
	}

	return mapSubmissions(rows), nil
}

func (s *Service) GetBKSubmissionsOverview(status, submissionType, classID, query string) (*BKSubmissionOverviewResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	trimmedClassID := strings.TrimSpace(classID)
	rows, err := s.querySubmissions(trimmedClassID, "", "")
	if err != nil {
		return nil, err
	}

	classRows, err := s.queryBKClasses()
	if err != nil {
		return nil, err
	}

	filteredRows := filterSubmissionRows(rows, status, submissionType, query)

	return &BKSubmissionOverviewResponse{
		StatusFilter: strings.TrimSpace(status),
		TypeFilter:   strings.ToUpper(strings.TrimSpace(submissionType)),
		ClassFilter:  trimmedClassID,
		Query:        strings.TrimSpace(query),
		Counts:       summarizeSubmissions(rows),
		Records:      mapSubmissions(filteredRows),
		Classes:      mapBKClassSummaries(classRows),
	}, nil
}

func (s *Service) ReviewBKSubmission(submissionID, actorID string, input ReviewSubmissionRequest) (*SubmissionResponse, error) {
	if s.db == nil {
		return nil, ErrStaffDataUnavailable
	}

	return s.reviewSubmission(submissionID, actorID, input)
}

func (s *Service) reviewSubmission(submissionID, actorID string, input ReviewSubmissionRequest) (*SubmissionResponse, error) {
	record, err := s.findSubmissionByID(submissionID)
	if err != nil {
		return nil, err
	}

	status := normalizeReviewSubmissionStatus(input.Status)
	if status == "" {
		return nil, fmt.Errorf("submission review status must be menunggu, diterima, or ditolak")
	}

	now := time.Now()
	record.Status = status
	record.ReviewedBy = stringPtr(actorID)
	record.ReviewedAt = &now
	record.ReviewNote = optionalString(input.ReviewNote)
	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("review submission: %w", err)
	}

	return s.findSubmissionResponseByID(record.ID)
}

func (s *Service) findTeacherMeByUserID(userID string) (*teacherMeRow, error) {
	var row teacherMeRow
	if err := s.db.Table("teachers").
		Select("users.id as user_id, teachers.id as teacher_id, users.name, users.username, teachers.nip, teachers.nuptk, teachers.gender, teachers.phone").
		Joins("join users on users.id = teachers.user_id").
		Where("teachers.user_id = ? AND teachers.is_active = ?", userID, true).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("find teacher profile by user: %w", err)
	}
	if row.TeacherID == "" {
		return nil, ErrTeacherProfileNotFound
	}
	return &row, nil
}

func (s *Service) queryTeacherSubjectAssignments(userID string) ([]subjectAssignmentRow, error) {
	if _, err := s.findTeacherMeByUserID(userID); err != nil {
		return nil, err
	}

	var rows []subjectAssignmentRow
	if err := s.db.Table("teacher_subject_assignments as tsa").
		Select("tsa.id, tsa.subject_id, subjects.code as subject_code, subjects.name as subject_name, tsa.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, tsa.school_year_id, school_years.name as school_year_name, tsa.is_active").
		Joins("join teachers on teachers.id = tsa.teacher_id").
		Joins("join subjects on subjects.id = tsa.subject_id").
		Joins("join classes on classes.id = tsa.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = tsa.school_year_id").
		Where("teachers.user_id = ?", userID).
		Order("school_years.start_year desc, class_name asc, subjects.code asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teacher subject assignments: %w", err)
	}
	return rows, nil
}

func (s *Service) findActiveHomeroomByUserID(userID string) (*homeroomRow, error) {
	var row homeroomRow
	if err := s.db.Table("homeroom_assignments as ha").
		Select("ha.id as assignment_id, ha.teacher_id, ha.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, ha.school_year_id, school_years.name as school_year_name, ha.is_active").
		Joins("join teachers on teachers.id = ha.teacher_id").
		Joins("join classes on classes.id = ha.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = ha.school_year_id").
		Where("teachers.user_id = ? AND ha.is_active = ?", userID, true).
		Order("school_years.is_active desc, school_years.start_year desc").
		Limit(1).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("find active homeroom assignment: %w", err)
	}
	if row.AssignmentID == "" {
		return nil, ErrHomeroomNotFound
	}
	return &row, nil
}

func (s *Service) queryStudentSummaries(classID, studentID string) ([]studentSummaryRow, error) {
	query := s.db.Table("students").
		Select("students.id, students.user_id, users.name, students.nis, students.nisn, students.gender, students.phone, classes.id as class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, school_years.id as school_year_id, school_years.name as school_year_name, scm.id as membership_id, scm.status as membership_status, students.entry_year, students.is_active").
		Joins("join users on users.id = students.user_id").
		Joins("join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("join classes on classes.id = scm.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = scm.school_year_id").
		Where("students.is_active = ?", true)

	if strings.TrimSpace(classID) != "" {
		query = query.Where("classes.id = ?", strings.TrimSpace(classID))
	}
	if strings.TrimSpace(studentID) != "" {
		query = query.Where("students.id = ?", strings.TrimSpace(studentID))
	}

	var rows []studentSummaryRow
	if err := query.Order("users.name asc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	if err := s.enrichStudentSummaryCounts(rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) enrichStudentSummaryCounts(rows []studentSummaryRow) error {
	if len(rows) == 0 {
		return nil
	}

	studentIDs := make([]string, 0, len(rows))
	indexByID := make(map[string]int, len(rows))
	for index, row := range rows {
		studentIDs = append(studentIDs, row.ID)
		indexByID[row.ID] = index
	}

	var counts []struct {
		StudentID string
		Status    string
		Count     int
	}
	if err := s.db.Table("attendance_records").
		Select("student_id, status, count(*) as count").
		Where("student_id IN ?", studentIDs).
		Group("student_id, status").
		Scan(&counts).Error; err != nil {
		return fmt.Errorf("aggregate student attendance counts: %w", err)
	}

	for _, count := range counts {
		index, exists := indexByID[count.StudentID]
		if !exists {
			continue
		}
		switch count.Status {
		case string(attendance.StatusHadir):
			rows[index].PresentCount = count.Count
		case string(attendance.StatusTelat):
			rows[index].LateCount = count.Count
		case string(attendance.StatusIzin):
			rows[index].PermissionCount = count.Count
		case string(attendance.StatusSakit):
			rows[index].SickCount = count.Count
		case string(attendance.StatusAlfa):
			rows[index].AlphaCount = count.Count
		}
	}

	return nil
}

func (s *Service) ensureHomeroomStudent(userID, studentID string) (string, error) {
	homeroom, err := s.findActiveHomeroomByUserID(userID)
	if err != nil {
		return "", err
	}

	rows, err := s.queryStudentSummaries(homeroom.ClassID, studentID)
	if err != nil {
		return "", err
	}
	if len(rows) == 0 {
		return "", ErrStudentNotFound
	}

	return rows[0].UserID, nil
}

func (s *Service) findAttendanceRecordClassID(id string) (string, error) {
	var row struct {
		ClassID string
	}
	if err := s.db.Table("attendance_records").Select("class_id").Where("id = ?", id).Scan(&row).Error; err != nil {
		return "", fmt.Errorf("find attendance record class: %w", err)
	}
	if row.ClassID == "" {
		return "", attendance.ErrAttendanceRecordNotFound
	}
	return row.ClassID, nil
}

func (s *Service) querySubmissions(classID, studentID, status string) ([]submissionRow, error) {
	query := s.db.Table("submissions").
		Select("submissions.id, submissions.student_id, users.name as student_name, students.nis, classes.id as class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, submissions.type, submissions.reason, submissions.attachment, submissions.status, submissions.reviewed_by, reviewer.name as reviewed_by_name, submissions.review_note, submissions.reviewed_at, submissions.created_at, submissions.updated_at").
		Joins("join students on students.id = submissions.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("left join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("left join classes on classes.id = scm.class_id").
		Joins("left join majors on majors.id = classes.major_id").
		Joins("left join users as reviewer on reviewer.id = submissions.reviewed_by")

	if strings.TrimSpace(classID) != "" {
		query = query.Where("classes.id = ?", strings.TrimSpace(classID))
	}
	if strings.TrimSpace(studentID) != "" {
		query = query.Where("submissions.student_id = ?", strings.TrimSpace(studentID))
	}
	if strings.TrimSpace(status) != "" {
		query = query.Where("submissions.status = ?", strings.TrimSpace(status))
	}

	var rows []submissionRow
	if err := query.Order("submissions.created_at desc, users.name asc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	return rows, nil
}

func (s *Service) findSubmissionByID(id string) (*leaveModule.Submission, error) {
	var record leaveModule.Submission
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, fmt.Errorf("find submission: %w", err)
	}
	return &record, nil
}

func (s *Service) findSubmissionResponseByID(id string) (*SubmissionResponse, error) {
	rows, err := s.querySubmissions("", "", "")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == id {
			response := mapSubmissionRow(row)
			return &response, nil
		}
	}
	return nil, ErrSubmissionNotFound
}

func (s *Service) findStudentIDByUserID(userID string) (string, error) {
	var row struct {
		ID string
	}
	if err := s.db.Model(&studentModule.Student{}).Select("id").Where("user_id = ? AND is_active = ?", userID, true).Scan(&row).Error; err != nil {
		return "", fmt.Errorf("find student by user: %w", err)
	}
	if row.ID == "" {
		return "", ErrStudentNotFound
	}
	return row.ID, nil
}

func (s *Service) findStudentUserIDByStudentID(studentID string) (string, error) {
	var row struct {
		UserID string
	}
	if err := s.db.Model(&studentModule.Student{}).Select("user_id").Where("id = ? AND is_active = ?", studentID, true).Scan(&row).Error; err != nil {
		return "", fmt.Errorf("find student user id: %w", err)
	}
	if row.UserID == "" {
		return "", ErrStudentNotFound
	}
	return row.UserID, nil
}

func (s *Service) queryCounselingNotes(classID, studentID, searchQuery string) ([]counselingNoteRow, error) {
	dbQuery := s.db.Table("notes").
		Select("notes.id, notes.student_id, users.name as student_name, students.nis, classes.id as class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, notes.created_by, creator.name as created_by_name, notes.title, notes.note, notes.created_at, notes.updated_at").
		Joins("join students on students.id = notes.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("left join student_class_memberships as scm on scm.student_id = students.id and scm.is_active = ?", true).
		Joins("left join classes on classes.id = scm.class_id").
		Joins("left join majors on majors.id = classes.major_id").
		Joins("left join users as creator on creator.id = notes.created_by")

	if strings.TrimSpace(classID) != "" {
		dbQuery = dbQuery.Where("classes.id = ?", strings.TrimSpace(classID))
	}
	if strings.TrimSpace(studentID) != "" {
		dbQuery = dbQuery.Where("notes.student_id = ?", strings.TrimSpace(studentID))
	}
	if strings.TrimSpace(searchQuery) != "" {
		like := "%" + strings.ToLower(strings.TrimSpace(searchQuery)) + "%"
		dbQuery = dbQuery.Where(
			"lower(users.name) LIKE ? OR lower(students.nis) LIKE ? OR lower(notes.title) LIKE ? OR lower(notes.note) LIKE ? OR lower(concat(classes.grade, ' ', majors.code, ' ', classes.name)) LIKE ?",
			like, like, like, like, like,
		)
	}

	var rows []counselingNoteRow
	if err := dbQuery.Order("notes.created_at desc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list counseling notes: %w", err)
	}
	return rows, nil
}

func (s *Service) findCounselingNoteByID(id string) (*counseling.Note, error) {
	var record counseling.Note
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCounselingNoteNotFound
		}
		return nil, fmt.Errorf("find counseling note: %w", err)
	}
	return &record, nil
}

func (s *Service) findCounselingNoteResponseByID(id string) (*CounselingNoteResponse, error) {
	rows, err := s.queryCounselingNotes("", "", "")
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == id {
			response := mapCounselingNoteRow(row)
			return &response, nil
		}
	}
	return nil, ErrCounselingNoteNotFound
}

func (s *Service) queryRiskStudents(status, classID string, daysWindow, minimumOccurrences int) ([]riskStudentRow, error) {
	since := time.Now().AddDate(0, 0, -daysWindow)

	query := s.db.Table("attendance_records as ar").
		Select("ar.student_id, users.name as student_name, students.nis, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, count(*) as occurrences").
		Joins("join students on students.id = ar.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("join classes on classes.id = ar.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Where("ar.status = ? AND ar.attendance_date >= ?", status, since.Format("2006-01-02"))

	if strings.TrimSpace(classID) != "" {
		query = query.Where("classes.id = ?", strings.TrimSpace(classID))
	}

	var rows []riskStudentRow
	if err := query.Group("ar.student_id, users.name, students.nis, class_name").
		Having("count(*) >= ?", minimumOccurrences).
		Order("occurrences desc, users.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list risk students: %w", err)
	}
	return rows, nil
}

func (s *Service) queryBKClasses() ([]bkClassSummaryRow, error) {
	var rows []bkClassSummaryRow
	if err := s.db.Table("classes").
		Select("classes.id as class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, school_years.id as school_year_id, school_years.name as school_year_name").
		Joins("join majors on majors.id = classes.major_id").
		Joins("left join school_years on school_years.id = classes.school_year_id").
		Where("classes.is_active = ?", true).
		Order("school_years.start_year desc, class_name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list bk classes: %w", err)
	}
	return rows, nil
}

func mapHomeroomRow(row homeroomRow) HomeroomContextResponse {
	return HomeroomContextResponse{
		AssignmentID:   row.AssignmentID,
		TeacherID:      row.TeacherID,
		ClassID:        row.ClassID,
		ClassName:      row.ClassName,
		SchoolYearID:   row.SchoolYearID,
		SchoolYearName: row.SchoolYearName,
		IsActive:       row.IsActive,
	}
}

func mapStudentSummaries(rows []studentSummaryRow) []StudentSummaryResponse {
	result := make([]StudentSummaryResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapStudentSummaryRow(row))
	}
	return result
}

func mapStudentSummaryRow(row studentSummaryRow) StudentSummaryResponse {
	return StudentSummaryResponse{
		ID:               row.ID,
		UserID:           row.UserID,
		Name:             row.Name,
		NIS:              row.NIS,
		NISN:             dereference(row.NISN),
		Gender:           dereference(row.Gender),
		Phone:            dereference(row.Phone),
		ClassID:          dereference(row.ClassID),
		ClassName:        dereference(row.ClassName),
		SchoolYearID:     dereference(row.SchoolYearID),
		SchoolYearName:   dereference(row.SchoolYearName),
		MembershipID:     dereference(row.MembershipID),
		MembershipStatus: dereference(row.MembershipStatus),
		EntryYear:        row.EntryYear,
		IsActive:         row.IsActive,
		PresentCount:     row.PresentCount,
		LateCount:        row.LateCount,
		PermissionCount:  row.PermissionCount,
		SickCount:        row.SickCount,
		AlphaCount:       row.AlphaCount,
	}
}

func mapSubmissions(rows []submissionRow) []SubmissionResponse {
	result := make([]SubmissionResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapSubmissionRow(row))
	}
	return result
}

func mapSubmissionRow(row submissionRow) SubmissionResponse {
	response := SubmissionResponse{
		ID:             row.ID,
		StudentID:      row.StudentID,
		StudentName:    row.StudentName,
		NIS:            row.NIS,
		ClassID:        dereference(row.ClassID),
		ClassName:      dereference(row.ClassName),
		Type:           row.Type,
		Reason:         row.Reason,
		Attachment:     row.Attachment,
		Status:         row.Status,
		ReviewedBy:     dereference(row.ReviewedBy),
		ReviewedByName: dereference(row.ReviewedByName),
		ReviewNote:     dereference(row.ReviewNote),
	}
	if row.ReviewedAt != nil {
		response.ReviewedAt = row.ReviewedAt.Format(time.RFC3339)
	}
	if !row.CreatedAt.IsZero() {
		response.CreatedAt = row.CreatedAt.Format(time.RFC3339)
	}
	if !row.UpdatedAt.IsZero() {
		response.UpdatedAt = row.UpdatedAt.Format(time.RFC3339)
	}
	return response
}

func mapCounselingNotes(rows []counselingNoteRow) []CounselingNoteResponse {
	result := make([]CounselingNoteResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapCounselingNoteRow(row))
	}
	return result
}

func mapCounselingNoteRow(row counselingNoteRow) CounselingNoteResponse {
	response := CounselingNoteResponse{
		ID:            row.ID,
		StudentID:     row.StudentID,
		StudentName:   row.StudentName,
		NIS:           row.NIS,
		ClassID:       dereference(row.ClassID),
		ClassName:     dereference(row.ClassName),
		CreatedBy:     row.CreatedBy,
		CreatedByName: dereference(row.CreatedByName),
		Title:         row.Title,
		Note:          row.Note,
	}
	if !row.CreatedAt.IsZero() {
		response.CreatedAt = row.CreatedAt.Format(time.RFC3339)
	}
	if !row.UpdatedAt.IsZero() {
		response.UpdatedAt = row.UpdatedAt.Format(time.RFC3339)
	}
	return response
}

func mapRiskStudents(rows []riskStudentRow) []RiskStudentRecord {
	result := make([]RiskStudentRecord, 0, len(rows))
	for _, row := range rows {
		result = append(result, RiskStudentRecord{
			StudentID:   row.StudentID,
			StudentName: row.StudentName,
			NIS:         row.NIS,
			ClassName:   row.ClassName,
			Occurrences: row.Occurrences,
		})
	}
	return result
}

func mapBKClassSummaries(rows []bkClassSummaryRow) []BKClassSummaryResponse {
	result := make([]BKClassSummaryResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, BKClassSummaryResponse{
			ClassID:        row.ClassID,
			ClassName:      row.ClassName,
			SchoolYearID:   dereference(row.SchoolYearID),
			SchoolYearName: dereference(row.SchoolYearName),
		})
	}
	return result
}

func summarizeAttendance(records []attendance.AttendanceRecordResponse) AttendanceSummaryResponse {
	summary := AttendanceSummaryResponse{}
	for _, record := range records {
		switch record.Status {
		case string(attendance.StatusHadir):
			summary.Present++
		case string(attendance.StatusTelat):
			summary.Late++
		case string(attendance.StatusIzin):
			summary.Permission++
		case string(attendance.StatusSakit):
			summary.Sick++
		case string(attendance.StatusAlfa):
			summary.Alpha++
		}
	}
	return summary
}

func buildAttentionStudents(records []attendance.AttendanceRecordResponse) []RiskStudentRecord {
	result := make([]RiskStudentRecord, 0)
	for _, record := range records {
		if record.Status != string(attendance.StatusTelat) && record.Status != string(attendance.StatusAlfa) {
			continue
		}
		result = append(result, RiskStudentRecord{
			StudentID:   record.StudentID,
			StudentName: record.StudentName,
			NIS:         record.NIS,
			ClassName:   record.ClassName,
			Occurrences: 1,
		})
	}
	return result
}

func buildBKDashboardRiskStudents(summary AttendanceSummaryResponse, limit int) []RiskStudentRecord {
	combined := make([]RiskStudentRecord, 0, len(summary.RepeatedAlpha)+len(summary.RepeatedLate))
	seen := make(map[string]struct{})

	for _, item := range summary.RepeatedAlpha {
		if _, exists := seen[item.StudentID]; exists {
			continue
		}
		seen[item.StudentID] = struct{}{}
		combined = append(combined, item)
	}
	for _, item := range summary.RepeatedLate {
		if _, exists := seen[item.StudentID]; exists {
			continue
		}
		seen[item.StudentID] = struct{}{}
		combined = append(combined, item)
	}

	if limit <= 0 || len(combined) <= limit {
		return combined
	}
	return combined[:limit]
}

func limitSubmissions(items []SubmissionResponse, limit int) []SubmissionResponse {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitCounselingNotes(items []CounselingNoteResponse, limit int) []CounselingNoteResponse {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitAttendanceRecords(items []attendance.AttendanceRecordResponse, limit int) []attendance.AttendanceRecordResponse {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

func summarizeSubmissions(rows []submissionRow) SubmissionCountsResponse {
	summary := SubmissionCountsResponse{
		Total: len(rows),
	}

	for _, row := range rows {
		if strings.TrimSpace(row.Attachment) != "" {
			summary.WithAttachment++
		}

		switch strings.ToLower(strings.TrimSpace(row.Status)) {
		case string(leaveModule.StatusPending):
			summary.Pending++
		case string(leaveModule.StatusApproved):
			summary.Approved++
		case string(leaveModule.StatusRejected):
			summary.Rejected++
		}
	}

	return summary
}

func countStudentSummaryRowsNeedingAttention(rows []studentSummaryRow) int {
	total := 0
	for _, row := range rows {
		if row.LateCount > 0 || row.AlphaCount > 0 {
			total++
		}
	}
	return total
}

func countActiveStudents(rows []studentSummaryRow) int {
	total := 0
	for _, row := range rows {
		if row.IsActive {
			total++
		}
	}
	return total
}

func sumStudentSummaryField(rows []studentSummaryRow, selector func(row studentSummaryRow) int) int {
	total := 0
	for _, row := range rows {
		total += selector(row)
	}
	return total
}

func collectUniqueStudentIDsFromNotes(rows []counselingNoteRow) map[string]struct{} {
	result := make(map[string]struct{})
	for _, row := range rows {
		if strings.TrimSpace(row.StudentID) == "" {
			continue
		}
		result[row.StudentID] = struct{}{}
	}
	return result
}

func filterStudentSummaryRows(rows []studentSummaryRow, riskFilter, query string, noteStudentIDs map[string]struct{}) []studentSummaryRow {
	normalizedRisk := strings.ToLower(strings.TrimSpace(riskFilter))
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	filtered := make([]studentSummaryRow, 0, len(rows))
	for _, row := range rows {
		if normalizedRisk != "" {
			switch normalizedRisk {
			case "need_attention":
				if row.LateCount == 0 && row.AlphaCount == 0 {
					continue
				}
			case "stable":
				if row.LateCount > 0 || row.AlphaCount > 0 {
					continue
				}
			case "late":
				if row.LateCount == 0 {
					continue
				}
			case "alpha":
				if row.AlphaCount == 0 {
					continue
				}
			case "counseling":
				if _, exists := noteStudentIDs[row.ID]; !exists {
					continue
				}
			}
		}

		if normalizedQuery != "" {
			searchSpace := strings.ToLower(strings.Join([]string{
				row.Name,
				row.NIS,
				dereference(row.NISN),
				dereference(row.Phone),
				dereference(row.ClassName),
			}, " "))
			if !strings.Contains(searchSpace, normalizedQuery) {
				continue
			}
		}

		filtered = append(filtered, row)
	}

	return filtered
}

func filterAttendanceRecords(items []attendance.AttendanceRecordResponse, query string) []attendance.AttendanceRecordResponse {
	if query == "" {
		return items
	}

	filtered := make([]attendance.AttendanceRecordResponse, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.StudentName), query) ||
			strings.Contains(strings.ToLower(item.NIS), query) ||
			strings.Contains(strings.ToLower(item.Status), query) ||
			strings.Contains(strings.ToLower(item.Notes), query) {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func filterSubmissionRows(rows []submissionRow, status, submissionType, query string) []submissionRow {
	normalizedStatus := strings.ToLower(strings.TrimSpace(status))
	normalizedType := strings.ToUpper(strings.TrimSpace(submissionType))
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	filtered := make([]submissionRow, 0, len(rows))
	for _, row := range rows {
		if normalizedStatus != "" && strings.ToLower(strings.TrimSpace(row.Status)) != normalizedStatus {
			continue
		}
		if normalizedType != "" && strings.ToUpper(strings.TrimSpace(row.Type)) != normalizedType {
			continue
		}
		if normalizedQuery != "" {
			className := dereference(row.ClassName)
			searchSpace := strings.ToLower(strings.Join([]string{
				row.StudentName,
				row.NIS,
				row.Type,
				row.Reason,
				row.Status,
				className,
			}, " "))
			if !strings.Contains(searchSpace, normalizedQuery) {
				continue
			}
		}

		filtered = append(filtered, row)
	}

	return filtered
}

func countUniqueStudentIDsInNotes(rows []counselingNoteRow) int {
	return len(collectUniqueStudentIDsFromNotes(rows))
}

func countUniqueClassIDsInNotes(rows []counselingNoteRow) int {
	seen := make(map[string]struct{})
	for _, row := range rows {
		classID := strings.TrimSpace(dereference(row.ClassID))
		if classID == "" {
			continue
		}
		seen[classID] = struct{}{}
	}
	return len(seen)
}

func countRecentWeekNotes(rows []counselingNoteRow) int {
	cutoff := time.Now().AddDate(0, 0, -7)
	total := 0
	for _, row := range rows {
		if !row.CreatedAt.IsZero() && (row.CreatedAt.After(cutoff) || row.CreatedAt.Equal(cutoff)) {
			total++
		}
	}
	return total
}

func currentHomeroomDate(raw string) string {
	if strings.TrimSpace(raw) != "" {
		return strings.TrimSpace(raw)
	}

	return time.Now().In(time.FixedZone("WIB", 7*60*60)).Format("2006-01-02")
}

func normalizeSubmissionType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "IZIN":
		return "IZIN"
	case "SAKIT":
		return "SAKIT"
	case "DISPENSASI":
		return "DISPENSASI"
	default:
		return ""
	}
}

func normalizeReviewSubmissionStatus(value string) leaveModule.Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(leaveModule.StatusPending):
		return leaveModule.StatusPending
	case string(leaveModule.StatusApproved):
		return leaveModule.StatusApproved
	case string(leaveModule.StatusRejected):
		return leaveModule.StatusRejected
	default:
		return ""
	}
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
