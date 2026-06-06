package admin

type DashboardResponse struct {
	AttendancePercentage int                `json:"attendance_percentage"`
	Counts               DashboardCounts    `json:"counts"`
	TodayStatus          TodayStatus        `json:"today_status"`
	SemesterTrend        []TrendPoint       `json:"semester_trend"`
	ClassPerformance     []ClassPerformance `json:"class_performance"`
	Announcements        []Announcement     `json:"announcements"`
}

type DashboardCounts struct {
	TotalUsers    int64 `json:"total_users"`
	TotalStudents int64 `json:"total_students"`
	TotalTeachers int64 `json:"total_teachers"`
	TotalBK       int64 `json:"total_bk"`
	TotalAdmins   int64 `json:"total_admins"`
}

type TrendPoint struct {
	Label   string `json:"label"`
	Present int    `json:"present"`
	Late    int    `json:"late"`
	Alpha   int    `json:"alpha"`
}

type TodayStatus struct {
	Present    int `json:"present"`
	Late       int `json:"late"`
	Permission int `json:"permission"`
	Sick       int `json:"sick"`
	Alpha      int `json:"alpha"`
}

type ClassPerformance struct {
	ClassName   string `json:"class_name"`
	Percentage  int    `json:"percentage"`
	PresentText string `json:"present_text"`
}

type Announcement struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tone        string `json:"tone"`
}

type UserResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	NIS      string `json:"nis,omitempty"`
	Username string `json:"username,omitempty"`
}

type TeacherResponse struct {
	No          string `json:"no"`
	ID          string `json:"id"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Class       string `json:"class"`
	NUPTK       string `json:"nuptk"`
	Contact     string `json:"contact"`
	AvatarLabel string `json:"avatar_label"`
}

type UpsertUserRequest struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	NIS      string `json:"nis"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type TeacherProfileResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Name     string `json:"name"`
	Username string `json:"username,omitempty"`
	NIP      string `json:"nip,omitempty"`
	NUPTK    string `json:"nuptk,omitempty"`
	Gender   string `json:"gender,omitempty"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	IsActive bool   `json:"is_active"`
}

type UpsertTeacherRequest struct {
	UserID   string `json:"user_id"`
	NIP      string `json:"nip"`
	NUPTK    string `json:"nuptk"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

type SubjectResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Group    string `json:"group,omitempty"`
	IsActive bool   `json:"is_active"`
}

type UpsertSubjectRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Group    string `json:"group"`
	IsActive *bool  `json:"is_active"`
}

type MajorResponse struct {
	ID       string `json:"id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

type UpsertMajorRequest struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsActive *bool  `json:"is_active"`
}

type SchoolYearResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
	IsActive  bool   `json:"is_active"`
}

type UpsertSchoolYearRequest struct {
	Name      string `json:"name"`
	StartYear int    `json:"start_year"`
	EndYear   int    `json:"end_year"`
	IsActive  *bool  `json:"is_active"`
}

type ClassResponse struct {
	ID             string `json:"id"`
	Grade          string `json:"grade"`
	Name           string `json:"name"`
	MajorID        string `json:"major_id"`
	MajorCode      string `json:"major_code"`
	MajorName      string `json:"major_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	DisplayName    string `json:"display_name"`
	IsActive       bool   `json:"is_active"`
}

type UpsertClassRequest struct {
	Grade        string `json:"grade"`
	Name         string `json:"name"`
	MajorID      string `json:"major_id"`
	SchoolYearID string `json:"school_year_id"`
	IsActive     *bool  `json:"is_active"`
}

type TeacherSubjectAssignmentResponse struct {
	ID             string `json:"id"`
	TeacherID      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	SubjectID      string `json:"subject_id"`
	SubjectCode    string `json:"subject_code"`
	SubjectName    string `json:"subject_name"`
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	IsActive       bool   `json:"is_active"`
}

type UpsertTeacherSubjectAssignmentRequest struct {
	TeacherID    string `json:"teacher_id"`
	SubjectID    string `json:"subject_id"`
	ClassID      string `json:"class_id"`
	SchoolYearID string `json:"school_year_id"`
	IsActive     *bool  `json:"is_active"`
}

type HomeroomAssignmentResponse struct {
	ID             string `json:"id"`
	TeacherID      string `json:"teacher_id"`
	TeacherName    string `json:"teacher_name"`
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	IsActive       bool   `json:"is_active"`
}

type UpsertHomeroomAssignmentRequest struct {
	TeacherID    string `json:"teacher_id"`
	ClassID      string `json:"class_id"`
	SchoolYearID string `json:"school_year_id"`
	IsActive     *bool  `json:"is_active"`
}

type StudentResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Name           string `json:"name"`
	NIS            string `json:"nis"`
	NISN           string `json:"nisn,omitempty"`
	Gender         string `json:"gender,omitempty"`
	BirthPlace     string `json:"birth_place,omitempty"`
	BirthDate      string `json:"birth_date,omitempty"`
	BirthPlaceDate string `json:"birth_place_date,omitempty"`
	Address        string `json:"address,omitempty"`
	Phone          string `json:"phone,omitempty"`
	ParentName     string `json:"parent_name,omitempty"`
	ParentPhone    string `json:"parent_phone,omitempty"`
	EntryYear      int    `json:"entry_year"`
	IsActive       bool   `json:"is_active"`
}

type UpsertStudentRequest struct {
	Name        string `json:"name"`
	NIS         string `json:"nis"`
	NISN        string `json:"nisn"`
	Password    string `json:"password"`
	Gender      string `json:"gender"`
	BirthPlace  string `json:"birth_place"`
	BirthDate   string `json:"birth_date"`
	Address     string `json:"address"`
	Phone       string `json:"phone"`
	ParentName  string `json:"parent_name"`
	ParentPhone string `json:"parent_phone"`
	EntryYear   int    `json:"entry_year"`
	IsActive    *bool  `json:"is_active"`
}

type StudentClassMembershipResponse struct {
	ID             string `json:"id"`
	StudentID      string `json:"student_id"`
	StudentName    string `json:"student_name"`
	NIS            string `json:"nis"`
	ClassID        string `json:"class_id"`
	ClassName      string `json:"class_name"`
	SchoolYearID   string `json:"school_year_id"`
	SchoolYearName string `json:"school_year_name"`
	Status         string `json:"status"`
	JoinedAt       string `json:"joined_at,omitempty"`
	LeftAt         string `json:"left_at,omitempty"`
	IsActive       bool   `json:"is_active"`
}

type UpsertStudentClassMembershipRequest struct {
	StudentID    string `json:"student_id"`
	ClassID      string `json:"class_id"`
	SchoolYearID string `json:"school_year_id"`
	Status       string `json:"status"`
	JoinedAt     string `json:"joined_at"`
	LeftAt       string `json:"left_at"`
	IsActive     *bool  `json:"is_active"`
}

type AttendanceRuleResponse struct {
	ID           string `json:"id"`
	SchoolYearID string `json:"school_year_id"`
	SchoolYear   string `json:"school_year"`
	CheckInStart string `json:"check_in_start"`
	OnTimeUntil  string `json:"on_time_until"`
	LateUntil    string `json:"late_until"`
	IsActive     bool   `json:"is_active"`
}

type UpsertAttendanceRuleRequest struct {
	SchoolYearID string `json:"school_year_id"`
	CheckInStart string `json:"check_in_start"`
	OnTimeUntil  string `json:"on_time_until"`
	LateUntil    string `json:"late_until"`
	IsActive     *bool  `json:"is_active"`
}
