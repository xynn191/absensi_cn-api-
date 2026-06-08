package public

type AttendanceWindowResponse struct {
	SchoolYear   string `json:"school_year"`
	CheckInStart string `json:"check_in_start"`
	OnTimeUntil  string `json:"on_time_until"`
	LateUntil    string `json:"late_until"`
}
