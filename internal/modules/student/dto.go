package student

import attendanceModule "absensi-cn-api/internal/modules/attendance"

type AttendanceWindowResponse struct {
	CheckInStart string `json:"check_in_start"`
	OnTimeUntil  string `json:"on_time_until"`
	LateUntil    string `json:"late_until"`
}

type StudentProfileResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Name             string `json:"name"`
	NIS              string `json:"nis"`
	NISN             string `json:"nisn,omitempty"`
	Gender           string `json:"gender,omitempty"`
	BirthPlace       string `json:"birth_place,omitempty"`
	BirthDate        string `json:"birth_date,omitempty"`
	BirthPlaceDate   string `json:"birth_place_date,omitempty"`
	Address          string `json:"address,omitempty"`
	Phone            string `json:"phone,omitempty"`
	ParentName       string `json:"parent_name,omitempty"`
	ParentPhone      string `json:"parent_phone,omitempty"`
	EntryYear        int    `json:"entry_year"`
	IsActive         bool   `json:"is_active"`
	ClassID          string `json:"class_id,omitempty"`
	ClassName        string `json:"class_name,omitempty"`
	MajorCode        string `json:"major_code,omitempty"`
	MajorName        string `json:"major_name,omitempty"`
	SchoolYearID     string `json:"school_year_id,omitempty"`
	SchoolYearName   string `json:"school_year_name,omitempty"`
	MembershipID     string `json:"membership_id,omitempty"`
	MembershipStatus string `json:"membership_status,omitempty"`
}

type StudentSubmissionResponse struct {
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

type StudentStatsResponse struct {
	TotalAttendance int `json:"total_attendance"`
	Present         int `json:"present"`
	Late            int `json:"late"`
	Permission      int `json:"permission"`
	Sick            int `json:"sick"`
	Alpha           int `json:"alpha"`
	PendingReviews  int `json:"pending_reviews"`
	PendingRequests int `json:"pending_requests"`
}

type StudentTodayResponse struct {
	Profile       StudentProfileResponse                     `json:"profile"`
	Window        AttendanceWindowResponse                   `json:"window"`
	Attendance    *attendanceModule.AttendanceRecordResponse `json:"attendance,omitempty"`
	CanSubmit     bool                                       `json:"can_submit"`
	CooldownUntil string                                     `json:"cooldown_until,omitempty"`
	CurrentStatus string                                     `json:"current_status"`
	CurrentTime   string                                     `json:"current_time"`
	Message       string                                     `json:"message"`
}

type StudentDashboardResponse struct {
	Today             StudentTodayResponse                        `json:"today"`
	Stats             StudentStatsResponse                        `json:"stats"`
	RecentAttendance  []attendanceModule.AttendanceRecordResponse `json:"recent_attendance"`
	RecentSubmissions []StudentSubmissionResponse                 `json:"recent_submissions"`
	Notifications     []StudentNotificationResponse               `json:"notifications"`
}

type StudentHistoryResponse struct {
	Profile     StudentProfileResponse                      `json:"profile"`
	Stats       StudentStatsResponse                        `json:"stats"`
	Attendance  []attendanceModule.AttendanceRecordResponse `json:"attendance"`
	Submissions []StudentSubmissionResponse                 `json:"submissions"`
}

type StudentNotificationResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tone        string `json:"tone"`
	CreatedAt   string `json:"created_at,omitempty"`
}
