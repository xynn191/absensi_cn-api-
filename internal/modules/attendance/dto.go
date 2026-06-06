package attendance

type StudentContextResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Name             string `json:"name"`
	NIS              string `json:"nis"`
	NISN             string `json:"nisn,omitempty"`
	Gender           string `json:"gender,omitempty"`
	Phone            string `json:"phone,omitempty"`
	ClassID          string `json:"class_id"`
	ClassName        string `json:"class_name"`
	MajorCode        string `json:"major_code"`
	MajorName        string `json:"major_name"`
	SchoolYearID     string `json:"school_year_id"`
	SchoolYearName   string `json:"school_year_name"`
	MembershipID     string `json:"membership_id"`
	MembershipStatus string `json:"membership_status"`
	IsActive         bool   `json:"is_active"`
}

type AttendanceWindowResponse struct {
	CheckInStart string `json:"check_in_start"`
	OnTimeUntil  string `json:"on_time_until"`
	LateUntil    string `json:"late_until"`
}

type AttendanceRecordResponse struct {
	ID               string `json:"id"`
	StudentID        string `json:"student_id"`
	StudentName      string `json:"student_name"`
	NIS              string `json:"nis"`
	ClassID          string `json:"class_id"`
	ClassName        string `json:"class_name"`
	SchoolYearID     string `json:"school_year_id"`
	SchoolYearName   string `json:"school_year_name"`
	AttendanceDate   string `json:"attendance_date"`
	CheckInAt        string `json:"check_in_at,omitempty"`
	Status           string `json:"status"`
	PhotoURL         string `json:"photo_url,omitempty"`
	Notes            string `json:"notes,omitempty"`
	VerifiedBy       string `json:"verified_by,omitempty"`
	VerifiedAt       string `json:"verified_at,omitempty"`
	VerificationNote string `json:"verification_note,omitempty"`
}

type TodayAttendanceResponse struct {
	Student        StudentContextResponse    `json:"student"`
	Window         AttendanceWindowResponse  `json:"window"`
	Attendance     *AttendanceRecordResponse `json:"attendance,omitempty"`
	CanCheckIn     bool                      `json:"can_check_in"`
	CurrentStatus  string                    `json:"current_status"`
	CurrentTime    string                    `json:"current_time"`
	Message        string                    `json:"message"`
}

type ReviewAttendanceRequest struct {
	Status           string `json:"status"`
	VerificationNote string `json:"verification_note"`
}
