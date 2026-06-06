package staff

import attendanceModule "absensi-cn-api/internal/modules/attendance"

type TeacherMeResponse struct {
	UserID            string                   `json:"user_id"`
	TeacherID         string                   `json:"teacher_id"`
	Name              string                   `json:"name"`
	Username          string                   `json:"username,omitempty"`
	NIP               string                   `json:"nip,omitempty"`
	NUPTK             string                   `json:"nuptk,omitempty"`
	Gender            string                   `json:"gender,omitempty"`
	Phone             string                   `json:"phone,omitempty"`
	IsHomeroomTeacher bool                     `json:"is_homeroom_teacher"`
	ActiveHomeroom    *HomeroomContextResponse `json:"active_homeroom,omitempty"`
}

type TeacherSubjectAssignmentResponse struct {
	ID             string `json:"id"`
	SubjectID      string `json:"subject_id"`
	SubjectCode    string `json:"subject_code"`
	SubjectName    string `json:"subject_name"`
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	IsActive       bool   `json:"is_active"`
}

type HomeroomContextResponse struct {
	AssignmentID   string `json:"assignment_id"`
	TeacherID      string `json:"teacher_id"`
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	IsActive       bool   `json:"is_active"`
}

type StudentSummaryResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Name             string `json:"name"`
	NIS              string `json:"nis"`
	NISN             string `json:"nisn,omitempty"`
	Gender           string `json:"gender,omitempty"`
	Phone            string `json:"phone,omitempty"`
	ClassID          string `json:"class_id,omitempty"`
	ClassName        string `json:"class_name,omitempty"`
	SchoolYearID     string `json:"school_year_id,omitempty"`
	SchoolYearName   string `json:"school_year_name,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
	MembershipStatus string `json:"membership_status,omitempty"`
	EntryYear        int    `json:"entry_year"`
	IsActive         bool   `json:"is_active"`
	PresentCount     int    `json:"present_count"`
	LateCount        int    `json:"late_count"`
	PermissionCount  int    `json:"permission_count"`
	SickCount        int    `json:"sick_count"`
	AlphaCount       int    `json:"alpha_count"`
}

type CreateSubmissionRequest struct {
	Type       string `json:"type"`
	Reason     string `json:"reason"`
	Attachment string `json:"attachment"`
}

type ReviewSubmissionRequest struct {
	Status     string `json:"status"`
	ReviewNote string `json:"review_note"`
}

type SubmissionResponse struct {
	ID             string `json:"id"`
	StudentID      string `json:"student_id"`
	StudentName    string `json:"student_name"`
	NIS            string `json:"nis"`
	ClassID        string `json:"class_id,omitempty"`
	ClassName      string `json:"class_name,omitempty"`
	Type           string `json:"type"`
	Reason         string `json:"reason"`
	Attachment     string `json:"attachment,omitempty"`
	Status         string `json:"status"`
	ReviewedBy     string `json:"reviewed_by,omitempty"`
	ReviewedByName string `json:"reviewed_by_name,omitempty"`
	ReviewNote     string `json:"review_note,omitempty"`
	ReviewedAt     string `json:"reviewed_at,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

type SubmissionCountsResponse struct {
	Total          int `json:"total"`
	Pending        int `json:"pending"`
	Approved       int `json:"approved"`
	Rejected       int `json:"rejected"`
	WithAttachment int `json:"with_attachment"`
}

type UpsertCounselingNoteRequest struct {
	Title string `json:"title"`
	Note  string `json:"note"`
}

type CounselingNoteResponse struct {
	ID            string `json:"id"`
	StudentID     string `json:"student_id"`
	StudentName   string `json:"student_name"`
	NIS           string `json:"nis"`
	ClassID       string `json:"class_id,omitempty"`
	ClassName     string `json:"class_name,omitempty"`
	CreatedBy     string `json:"created_by"`
	CreatedByName string `json:"created_by_name,omitempty"`
	Title         string `json:"title"`
	Note          string `json:"note"`
	CreatedAt     string `json:"created_at,omitempty"`
	UpdatedAt     string `json:"updated_at,omitempty"`
}

type AttendanceSummaryResponse struct {
	Present       int                 `json:"present"`
	Late          int                 `json:"late"`
	Permission    int                 `json:"permission"`
	Sick          int                 `json:"sick"`
	Alpha         int                 `json:"alpha"`
	RepeatedLate  []RiskStudentRecord `json:"repeated_late"`
	RepeatedAlpha []RiskStudentRecord `json:"repeated_alpha"`
}

type RiskStudentRecord struct {
	StudentID   string `json:"student_id"`
	StudentName string `json:"student_name"`
	NIS         string `json:"nis"`
	ClassName   string `json:"class_name"`
	Occurrences int    `json:"occurrences"`
}

type TeacherAttendanceReviewRequest struct {
	Status           string `json:"status"`
	VerificationNote string `json:"verification_note"`
}

type AttendanceListResponse struct {
	Records []attendanceModule.AttendanceRecordResponse `json:"records"`
}

type HomeroomAttendanceOverviewResponse struct {
	Homeroom     HomeroomContextResponse                     `json:"homeroom"`
	Date         string                                      `json:"date"`
	StatusFilter string                                      `json:"status_filter,omitempty"`
	Query        string                                      `json:"query,omitempty"`
	Summary      AttendanceSummaryResponse                   `json:"summary"`
	Records      []attendanceModule.AttendanceRecordResponse `json:"records"`
}

type HomeroomDashboardResponse struct {
	Homeroom                 HomeroomContextResponse   `json:"homeroom"`
	TotalStudents            int                       `json:"total_students"`
	Today                    AttendanceSummaryResponse `json:"today"`
	StudentsNeedingAttention []RiskStudentRecord       `json:"students_needing_attention"`
	RecentSubmissions        []SubmissionResponse      `json:"recent_submissions"`
}

type HomeroomStudentDetailResponse struct {
	Student           StudentSummaryResponse                      `json:"student"`
	AttendanceSummary AttendanceSummaryResponse                   `json:"attendance_summary"`
	RecentAttendance  []attendanceModule.AttendanceRecordResponse `json:"recent_attendance"`
	RecentSubmissions []SubmissionResponse                        `json:"recent_submissions"`
}

type HomeroomSubmissionOverviewResponse struct {
	Homeroom     HomeroomContextResponse  `json:"homeroom"`
	StatusFilter string                   `json:"status_filter,omitempty"`
	TypeFilter   string                   `json:"type_filter,omitempty"`
	Query        string                   `json:"query,omitempty"`
	Counts       SubmissionCountsResponse `json:"counts"`
	Records      []SubmissionResponse     `json:"records"`
}

type BKClassSummaryResponse struct {
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id,omitempty"`
	SchoolYearName string `json:"school_year_name,omitempty"`
}

type BKDashboardResponse struct {
	TotalStudents         int                       `json:"total_students"`
	StudentsNeedAttention int                       `json:"students_need_attention"`
	TotalCounselingNotes  int                       `json:"total_counseling_notes"`
	PendingSubmissions    int                       `json:"pending_submissions"`
	Today                 AttendanceSummaryResponse `json:"today"`
	TopRiskStudents       []RiskStudentRecord       `json:"top_risk_students"`
	RecentSubmissions     []SubmissionResponse      `json:"recent_submissions"`
	RecentCounselingNotes []CounselingNoteResponse  `json:"recent_counseling_notes"`
	Classes               []BKClassSummaryResponse  `json:"classes"`
}

type BKStudentsOverviewResponse struct {
	ClassFilter string                   `json:"class_filter,omitempty"`
	RiskFilter  string                   `json:"risk_filter,omitempty"`
	Query       string                   `json:"query,omitempty"`
	Counts      StudentOverviewCounts    `json:"counts"`
	Students    []StudentSummaryResponse `json:"students"`
	Classes     []BKClassSummaryResponse `json:"classes"`
}

type StudentOverviewCounts struct {
	Total               int `json:"total"`
	Active              int `json:"active"`
	NeedAttention       int `json:"need_attention"`
	TotalLate           int `json:"total_late"`
	TotalAlpha          int `json:"total_alpha"`
	WithCounselingNotes int `json:"with_counseling_notes"`
}

type BKStudentDetailResponse struct {
	Student           StudentSummaryResponse                      `json:"student"`
	AttendanceSummary AttendanceSummaryResponse                   `json:"attendance_summary"`
	RecentAttendance  []attendanceModule.AttendanceRecordResponse `json:"recent_attendance"`
	RecentSubmissions []SubmissionResponse                        `json:"recent_submissions"`
	CounselingNotes   []CounselingNoteResponse                    `json:"counseling_notes"`
}

type BKAttendanceOverviewResponse struct {
	Date         string                                      `json:"date"`
	StatusFilter string                                      `json:"status_filter,omitempty"`
	ClassFilter  string                                      `json:"class_filter,omitempty"`
	Query        string                                      `json:"query,omitempty"`
	Summary      AttendanceSummaryResponse                   `json:"summary"`
	Records      []attendanceModule.AttendanceRecordResponse `json:"records"`
	Classes      []BKClassSummaryResponse                    `json:"classes"`
}

type BKCounselingOverviewResponse struct {
	ClassFilter string                   `json:"class_filter,omitempty"`
	StudentID   string                   `json:"student_id,omitempty"`
	Query       string                   `json:"query,omitempty"`
	Counts      CounselingOverviewCount  `json:"counts"`
	Records     []CounselingNoteResponse `json:"records"`
	Classes     []BKClassSummaryResponse `json:"classes"`
	Students    []StudentSummaryResponse `json:"students"`
}

type CounselingOverviewCount struct {
	TotalNotes      int `json:"total_notes"`
	StudentsCovered int `json:"students_covered"`
	ClassesCovered  int `json:"classes_covered"`
	RecentWeekNotes int `json:"recent_week_notes"`
}

type BKSubmissionOverviewResponse struct {
	StatusFilter string                   `json:"status_filter,omitempty"`
	TypeFilter   string                   `json:"type_filter,omitempty"`
	ClassFilter  string                   `json:"class_filter,omitempty"`
	Query        string                   `json:"query,omitempty"`
	Counts       SubmissionCountsResponse `json:"counts"`
	Records      []SubmissionResponse     `json:"records"`
	Classes      []BKClassSummaryResponse `json:"classes"`
}
