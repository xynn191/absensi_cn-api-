package admin

import (
	"strings"
	"time"
)

func ValidateUpsertUserRequest(input UpsertUserRequest, isCreate bool) map[string]string {
	errors := map[string]string{}

	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}

	switch strings.TrimSpace(input.Role) {
	case "STUDENT":
		if len(strings.TrimSpace(input.NIS)) != 10 {
			errors["nis"] = "nis must contain exactly 10 digits"
		}
		for _, char := range strings.TrimSpace(input.NIS) {
			if char < '0' || char > '9' {
				errors["nis"] = "nis must contain exactly 10 digits"
				break
			}
		}
	case "TEACHER", "BK", "ADMIN":
		if strings.TrimSpace(input.Username) == "" {
			errors["username"] = "username is required for staff and admin roles"
		}
		if strings.TrimSpace(input.NIS) != "" {
			errors["nis"] = "nis is only allowed for student role"
		}
	default:
		errors["role"] = "role must be one of STUDENT, TEACHER, BK, or ADMIN"
	}

	if isCreate && len(strings.TrimSpace(input.Password)) < 6 {
		errors["password"] = "password must be at least 6 characters"
	}

	if !isCreate && strings.TrimSpace(input.Password) != "" && len(strings.TrimSpace(input.Password)) < 6 {
		errors["password"] = "password must be at least 6 characters"
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func ValidateUpsertTeacherRequest(input UpsertTeacherRequest) map[string]string {
	errors := map[string]string{}

	if strings.TrimSpace(input.UserID) == "" {
		errors["user_id"] = "user_id is required"
	}

	if input.Gender != "" {
		switch strings.ToUpper(strings.TrimSpace(input.Gender)) {
		case "MALE", "FEMALE", "L", "P":
		default:
			errors["gender"] = "gender must be one of MALE, FEMALE, L, or P"
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

func ValidateUpsertSubjectRequest(input UpsertSubjectRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.Code) == "" {
		errors["code"] = "code is required"
	}
	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertMajorRequest(input UpsertMajorRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.Code) == "" {
		errors["code"] = "code is required"
	}
	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertSchoolYearRequest(input UpsertSchoolYearRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}
	if input.StartYear <= 0 {
		errors["start_year"] = "start_year must be greater than 0"
	}
	if input.EndYear <= input.StartYear {
		errors["end_year"] = "end_year must be greater than start_year"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertClassRequest(input UpsertClassRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.Grade) == "" {
		errors["grade"] = "grade is required"
	}
	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}
	if strings.TrimSpace(input.MajorID) == "" {
		errors["major_id"] = "major_id is required"
	}
	if strings.TrimSpace(input.SchoolYearID) == "" {
		errors["school_year_id"] = "school_year_id is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertTeacherSubjectAssignmentRequest(input UpsertTeacherSubjectAssignmentRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.TeacherID) == "" {
		errors["teacher_id"] = "teacher_id is required"
	}
	if strings.TrimSpace(input.SubjectID) == "" {
		errors["subject_id"] = "subject_id is required"
	}
	if strings.TrimSpace(input.ClassID) == "" {
		errors["class_id"] = "class_id is required"
	}
	if strings.TrimSpace(input.SchoolYearID) == "" {
		errors["school_year_id"] = "school_year_id is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertHomeroomAssignmentRequest(input UpsertHomeroomAssignmentRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.TeacherID) == "" {
		errors["teacher_id"] = "teacher_id is required"
	}
	if strings.TrimSpace(input.ClassID) == "" {
		errors["class_id"] = "class_id is required"
	}
	if strings.TrimSpace(input.SchoolYearID) == "" {
		errors["school_year_id"] = "school_year_id is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertStudentRequest(input UpsertStudentRequest, isCreate bool) map[string]string {
	errors := map[string]string{}

	if strings.TrimSpace(input.Name) == "" {
		errors["name"] = "name is required"
	}

	trimmedNIS := strings.TrimSpace(input.NIS)
	if len(trimmedNIS) != 10 {
		errors["nis"] = "nis must contain exactly 10 digits"
	} else {
		for _, char := range trimmedNIS {
			if char < '0' || char > '9' {
				errors["nis"] = "nis must contain exactly 10 digits"
				break
			}
		}
	}

	if isCreate && len(strings.TrimSpace(input.Password)) < 6 {
		errors["password"] = "password must be at least 6 characters"
	}
	if !isCreate && strings.TrimSpace(input.Password) != "" && len(strings.TrimSpace(input.Password)) < 6 {
		errors["password"] = "password must be at least 6 characters"
	}

	if input.EntryYear <= 0 {
		errors["entry_year"] = "entry_year must be greater than 0"
	}

	if input.Gender != "" {
		switch strings.ToUpper(strings.TrimSpace(input.Gender)) {
		case "MALE", "FEMALE", "L", "P":
		default:
			errors["gender"] = "gender must be one of MALE, FEMALE, L, or P"
		}
	}

	if input.BirthDate != "" {
		if len(strings.TrimSpace(input.BirthDate)) != len("2006-01-02") {
			errors["birth_date"] = "birth_date must be in YYYY-MM-DD format"
		}
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertStudentClassMembershipRequest(input UpsertStudentClassMembershipRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.StudentID) == "" {
		errors["student_id"] = "student_id is required"
	}
	if strings.TrimSpace(input.ClassID) == "" {
		errors["class_id"] = "class_id is required"
	}
	if strings.TrimSpace(input.SchoolYearID) == "" {
		errors["school_year_id"] = "school_year_id is required"
	}
	if strings.TrimSpace(input.Status) == "" {
		errors["status"] = "status is required"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func ValidateUpsertAttendanceRuleRequest(input UpsertAttendanceRuleRequest) map[string]string {
	errors := map[string]string{}
	if strings.TrimSpace(input.SchoolYearID) == "" {
		errors["school_year_id"] = "school_year_id is required"
	}
	if strings.TrimSpace(input.CheckInStart) == "" {
		errors["check_in_start"] = "check_in_start is required"
	}
	if strings.TrimSpace(input.OnTimeUntil) == "" {
		errors["on_time_until"] = "on_time_until is required"
	}
	if strings.TrimSpace(input.LateUntil) == "" {
		errors["late_until"] = "late_until is required"
	}

	checkInStart, startErr := parseClockValue(input.CheckInStart)
	if strings.TrimSpace(input.CheckInStart) != "" && startErr != nil {
		errors["check_in_start"] = "check_in_start must use HH:MM or HH:MM:SS"
	}

	onTimeUntil, onTimeErr := parseClockValue(input.OnTimeUntil)
	if strings.TrimSpace(input.OnTimeUntil) != "" && onTimeErr != nil {
		errors["on_time_until"] = "on_time_until must use HH:MM or HH:MM:SS"
	}

	lateUntil, lateErr := parseClockValue(input.LateUntil)
	if strings.TrimSpace(input.LateUntil) != "" && lateErr != nil {
		errors["late_until"] = "late_until must use HH:MM or HH:MM:SS"
	}

	if startErr == nil && onTimeErr == nil && !checkInStart.Before(onTimeUntil) {
		errors["on_time_until"] = "on_time_until must be later than check_in_start"
	}
	if onTimeErr == nil && lateErr == nil && !onTimeUntil.Before(lateUntil) {
		errors["late_until"] = "late_until must be later than on_time_until"
	}

	if len(errors) == 0 {
		return nil
	}
	return errors
}

func parseClockValue(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	parsed, err := time.Parse("15:04:05", trimmed)
	if err == nil {
		return parsed, nil
	}

	return time.Parse("15:04", trimmed)
}
