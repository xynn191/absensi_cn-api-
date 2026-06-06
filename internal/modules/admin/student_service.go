package admin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	attendanceModule "absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/internal/modules/counseling"
	leaveModule "absensi-cn-api/internal/modules/leave"
	studentModule "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/password"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrStudentNotFound           = errors.New("student not found")
	ErrStudentMembershipNotFound = errors.New("student class membership not found")
	ErrAttendanceRuleNotFound    = errors.New("attendance rule not found")
)

type studentRow struct {
	ID          string
	UserID      string
	Name        string
	NIS         string
	NISN        *string
	Gender      *string
	BirthPlace  *string
	BirthDate   *time.Time
	Address     *string
	Phone       *string
	ParentName  *string
	ParentPhone *string
	EntryYear   int
	IsActive    bool
}

type studentMembershipRow struct {
	ID             string
	StudentID      string
	StudentName    string
	NIS            string
	ClassID        string
	ClassName      string
	SchoolYearID   string
	SchoolYearName string
	Status         string
	JoinedAt       *time.Time
	LeftAt         *time.Time
	IsActive       bool
}

type attendanceRuleRow struct {
	ID           string
	SchoolYearID string
	SchoolYear   string
	CheckInStart string
	OnTimeUntil  string
	LateUntil    string
	IsActive     bool
}

func (s *Service) ListStudents() ([]StudentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []studentRow
	if err := s.db.Table("students").
		Select("students.id, students.user_id, users.name, students.nis, students.nisn, students.gender, students.birth_place, students.birth_date, students.address, students.phone, students.parent_name, students.parent_phone, students.entry_year, students.is_active").
		Joins("join users on users.id = students.user_id").
		Order("users.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}

	result := make([]StudentResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapStudentRow(row))
	}
	return result, nil
}

func (s *Service) GetStudent(id string) (*StudentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var row studentRow
	if err := s.db.Table("students").
		Select("students.id, students.user_id, users.name, students.nis, students.nisn, students.gender, students.birth_place, students.birth_date, students.address, students.phone, students.parent_name, students.parent_phone, students.entry_year, students.is_active").
		Joins("join users on users.id = students.user_id").
		Where("students.id = ?", id).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("get student: %w", err)
	}
	if row.ID == "" {
		return nil, ErrStudentNotFound
	}

	response := mapStudentRow(row)
	return &response, nil
}

func (s *Service) CreateStudent(input UpsertStudentRequest) (*StudentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	if err := s.ensureUniqueStudentFields("", "", input); err != nil {
		return nil, err
	}

	hashedPassword, err := password.Hash(strings.TrimSpace(input.Password))
	if err != nil {
		return nil, fmt.Errorf("hash student password: %w", err)
	}

	var response StudentResponse
	err = s.db.Transaction(func(tx *gorm.DB) error {
		account := user.User{
			ID:           uuid.NewString(),
			Name:         strings.TrimSpace(input.Name),
			NIS:          stringPtr(strings.TrimSpace(input.NIS)),
			PasswordHash: hashedPassword,
			Role:         user.RoleStudent,
		}
		if err := tx.Create(&account).Error; err != nil {
			return fmt.Errorf("create student user: %w", err)
		}

		birthDate, err := parseOptionalDate(input.BirthDate)
		if err != nil {
			return err
		}

		record := studentModule.Student{
			ID:          uuid.NewString(),
			UserID:      account.ID,
			NIS:         strings.TrimSpace(input.NIS),
			NISN:        optionalString(input.NISN),
			Gender:      optionalUpperString(input.Gender),
			BirthPlace:  optionalString(input.BirthPlace),
			BirthDate:   birthDate,
			Address:     optionalString(input.Address),
			Phone:       optionalString(input.Phone),
			ParentName:  optionalString(input.ParentName),
			ParentPhone: optionalString(input.ParentPhone),
			EntryYear:   input.EntryYear,
			IsActive:    boolValueOrDefault(input.IsActive, true),
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("create student profile: %w", err)
		}

		response = mapStudentRow(studentRow{
			ID:          record.ID,
			UserID:      record.UserID,
			Name:        account.Name,
			NIS:         record.NIS,
			NISN:        record.NISN,
			Gender:      record.Gender,
			BirthPlace:  record.BirthPlace,
			BirthDate:   record.BirthDate,
			Address:     record.Address,
			Phone:       record.Phone,
			ParentName:  record.ParentName,
			ParentPhone: record.ParentPhone,
			EntryYear:   record.EntryYear,
			IsActive:    record.IsActive,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *Service) UpdateStudent(id string, input UpsertStudentRequest) (*StudentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	record, err := s.findStudentByID(id)
	if err != nil {
		return nil, err
	}
	account, err := s.findUserByID(record.UserID)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueStudentFields(id, account.ID, input); err != nil {
		return nil, err
	}

	birthDate, err := parseOptionalDate(input.BirthDate)
	if err != nil {
		return nil, err
	}

	account.Name = strings.TrimSpace(input.Name)
	account.NIS = stringPtr(strings.TrimSpace(input.NIS))
	if trimmedPassword := strings.TrimSpace(input.Password); trimmedPassword != "" {
		hashedPassword, err := password.Hash(trimmedPassword)
		if err != nil {
			return nil, fmt.Errorf("hash student password: %w", err)
		}
		account.PasswordHash = hashedPassword
	}

	record.NIS = strings.TrimSpace(input.NIS)
	record.NISN = optionalString(input.NISN)
	record.Gender = optionalUpperString(input.Gender)
	record.BirthPlace = optionalString(input.BirthPlace)
	record.BirthDate = birthDate
	record.Address = optionalString(input.Address)
	record.Phone = optionalString(input.Phone)
	record.ParentName = optionalString(input.ParentName)
	record.ParentPhone = optionalString(input.ParentPhone)
	record.EntryYear = input.EntryYear
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(account).Error; err != nil {
			return fmt.Errorf("update student user: %w", err)
		}
		if err := tx.Save(record).Error; err != nil {
			return fmt.Errorf("update student profile: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	response := mapStudentRow(studentRow{
		ID:          record.ID,
		UserID:      record.UserID,
		Name:        account.Name,
		NIS:         record.NIS,
		NISN:        record.NISN,
		Gender:      record.Gender,
		BirthPlace:  record.BirthPlace,
		BirthDate:   record.BirthDate,
		Address:     record.Address,
		Phone:       record.Phone,
		ParentName:  record.ParentName,
		ParentPhone: record.ParentPhone,
		EntryYear:   record.EntryYear,
		IsActive:    record.IsActive,
	})

	return &response, nil
}

func (s *Service) DeleteStudent(id string) error {
	if s.db == nil {
		return ErrAdminDataUnavailable
	}

	record, err := s.findStudentByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("student_id = ?", record.ID).Delete(&studentModule.StudentClassMembership{}).Error; err != nil {
			return fmt.Errorf("delete student memberships: %w", err)
		}
		if err := tx.Where("student_id = ?", record.ID).Delete(&attendanceModule.AttendanceRecord{}).Error; err != nil {
			return fmt.Errorf("delete student attendance records: %w", err)
		}
		if err := tx.Where("student_id = ?", record.ID).Delete(&leaveModule.Submission{}).Error; err != nil {
			return fmt.Errorf("delete student submissions: %w", err)
		}
		if err := tx.Where("student_id = ?", record.ID).Delete(&counseling.Note{}).Error; err != nil {
			return fmt.Errorf("delete student counseling notes: %w", err)
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete student profile: %w", err)
		}
		if err := tx.Where("id = ?", record.UserID).Delete(&user.User{}).Error; err != nil {
			return fmt.Errorf("delete student user: %w", err)
		}
		return nil
	})
}

func (s *Service) ListStudentClassMemberships() ([]StudentClassMembershipResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []studentMembershipRow
	if err := s.db.Table("student_class_memberships as scm").
		Select("scm.id, scm.student_id, users.name as student_name, students.nis, scm.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, scm.school_year_id, school_years.name as school_year_name, scm.status, scm.joined_at, scm.left_at, scm.is_active").
		Joins("join students on students.id = scm.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("join classes on classes.id = scm.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = scm.school_year_id").
		Order("school_years.start_year desc, users.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list student class memberships: %w", err)
	}

	result := make([]StudentClassMembershipResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapStudentMembershipRow(row))
	}
	return result, nil
}

func (s *Service) GetStudentClassMembership(id string) (*StudentClassMembershipResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	return s.findStudentMembershipResponse(id)
}

func (s *Service) CreateStudentClassMembership(input UpsertStudentClassMembershipRequest) (*StudentClassMembershipResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	studentRecord, err := s.findStudentByID(strings.TrimSpace(input.StudentID))
	if err != nil {
		return nil, err
	}
	classRecord, err := s.findClassByID(strings.TrimSpace(input.ClassID))
	if err != nil {
		return nil, err
	}
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}
	if classRecord.SchoolYearID != schoolYearRecord.ID {
		return nil, fmt.Errorf("class does not belong to the provided school year")
	}

	status := normalizeMembershipStatus(input.Status)
	if status == "" {
		return nil, fmt.Errorf("membership status is invalid")
	}

	if err := s.ensureUniqueStudentMembership("", studentRecord.ID, classRecord.ID, schoolYearRecord.ID); err != nil {
		return nil, err
	}
	isActive := status == studentModule.MembershipStatusActive
	if err := s.ensureSingleActiveMembership("", studentRecord.ID, schoolYearRecord.ID, isActive); err != nil {
		return nil, err
	}

	joinedAt, err := parseOptionalDateTime(input.JoinedAt)
	if err != nil {
		return nil, err
	}
	leftAt, err := parseOptionalDateTime(input.LeftAt)
	if err != nil {
		return nil, err
	}

	record := studentModule.StudentClassMembership{
		ID:           uuid.NewString(),
		StudentID:    studentRecord.ID,
		ClassID:      classRecord.ID,
		SchoolYearID: schoolYearRecord.ID,
		Status:       status,
		JoinedAt:     joinedAt,
		LeftAt:       leftAt,
		IsActive:     isActive,
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create student class membership: %w", err)
	}

	return s.findStudentMembershipResponse(record.ID)
}

func (s *Service) UpdateStudentClassMembership(id string, input UpsertStudentClassMembershipRequest) (*StudentClassMembershipResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	record, err := s.findStudentMembershipByID(id)
	if err != nil {
		return nil, err
	}
	studentRecord, err := s.findStudentByID(strings.TrimSpace(input.StudentID))
	if err != nil {
		return nil, err
	}
	classRecord, err := s.findClassByID(strings.TrimSpace(input.ClassID))
	if err != nil {
		return nil, err
	}
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}
	if classRecord.SchoolYearID != schoolYearRecord.ID {
		return nil, fmt.Errorf("class does not belong to the provided school year")
	}

	status := normalizeMembershipStatus(input.Status)
	if status == "" {
		return nil, fmt.Errorf("membership status is invalid")
	}
	if err := s.ensureUniqueStudentMembership(id, studentRecord.ID, classRecord.ID, schoolYearRecord.ID); err != nil {
		return nil, err
	}
	isActive := status == studentModule.MembershipStatusActive
	if err := s.ensureSingleActiveMembership(id, studentRecord.ID, schoolYearRecord.ID, isActive); err != nil {
		return nil, err
	}

	joinedAt, err := parseOptionalDateTime(input.JoinedAt)
	if err != nil {
		return nil, err
	}
	leftAt, err := parseOptionalDateTime(input.LeftAt)
	if err != nil {
		return nil, err
	}

	record.StudentID = studentRecord.ID
	record.ClassID = classRecord.ID
	record.SchoolYearID = schoolYearRecord.ID
	record.Status = status
	record.JoinedAt = joinedAt
	record.LeftAt = leftAt
	record.IsActive = isActive
	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update student class membership: %w", err)
	}
	if err := s.db.Model(&attendanceModule.AttendanceRecord{}).
		Where("student_class_membership_id = ?", record.ID).
		Updates(map[string]interface{}{
			"class_id":       record.ClassID,
			"school_year_id": record.SchoolYearID,
		}).Error; err != nil {
		return nil, fmt.Errorf("sync attendance records for membership update: %w", err)
	}

	return s.findStudentMembershipResponse(record.ID)
}

func (s *Service) DeleteStudentClassMembership(id string) error {
	record, err := s.findStudentMembershipByID(id)
	if err != nil {
		return err
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("student_class_membership_id = ?", record.ID).Delete(&attendanceModule.AttendanceRecord{}).Error; err != nil {
			return fmt.Errorf("delete attendance records by membership: %w", err)
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete student class membership: %w", err)
		}
		return nil
	})
}

func (s *Service) ListAttendanceRules() ([]AttendanceRuleResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []attendanceRuleRow
	if err := s.db.Table("attendance_rules").
		Select("attendance_rules.id, attendance_rules.school_year_id, school_years.name as school_year, attendance_rules.check_in_start, attendance_rules.on_time_until, attendance_rules.late_until, attendance_rules.is_active").
		Joins("join school_years on school_years.id = attendance_rules.school_year_id").
		Order("school_years.start_year desc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list attendance rules: %w", err)
	}

	result := make([]AttendanceRuleResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, AttendanceRuleResponse{
			ID:           row.ID,
			SchoolYearID: row.SchoolYearID,
			SchoolYear:   row.SchoolYear,
			CheckInStart: row.CheckInStart,
			OnTimeUntil:  row.OnTimeUntil,
			LateUntil:    row.LateUntil,
			IsActive:     row.IsActive,
		})
	}
	return result, nil
}

func (s *Service) GetAttendanceRule(id string) (*AttendanceRuleResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var row attendanceRuleRow
	if err := s.db.Table("attendance_rules").
		Select("attendance_rules.id, attendance_rules.school_year_id, school_years.name as school_year, attendance_rules.check_in_start, attendance_rules.on_time_until, attendance_rules.late_until, attendance_rules.is_active").
		Joins("join school_years on school_years.id = attendance_rules.school_year_id").
		Where("attendance_rules.id = ?", id).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("get attendance rule: %w", err)
	}
	if row.ID == "" {
		return nil, ErrAttendanceRuleNotFound
	}

	response := AttendanceRuleResponse{
		ID:           row.ID,
		SchoolYearID: row.SchoolYearID,
		SchoolYear:   row.SchoolYear,
		CheckInStart: row.CheckInStart,
		OnTimeUntil:  row.OnTimeUntil,
		LateUntil:    row.LateUntil,
		IsActive:     row.IsActive,
	}
	return &response, nil
}

func (s *Service) CreateAttendanceRule(input UpsertAttendanceRuleRequest) (*AttendanceRuleResponse, error) {
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueRecord("attendance_rules", "school_year_id", schoolYearRecord.ID, ""); err != nil {
		return nil, err
	}

	record := attendanceModule.AttendanceRule{
		ID:           uuid.NewString(),
		SchoolYearID: schoolYearRecord.ID,
		CheckInStart: strings.TrimSpace(input.CheckInStart),
		OnTimeUntil:  strings.TrimSpace(input.OnTimeUntil),
		LateUntil:    strings.TrimSpace(input.LateUntil),
		IsActive:     boolValueOrDefault(input.IsActive, true),
	}
	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create attendance rule: %w", err)
	}

	response := AttendanceRuleResponse{
		ID:           record.ID,
		SchoolYearID: record.SchoolYearID,
		SchoolYear:   schoolYearRecord.Name,
		CheckInStart: record.CheckInStart,
		OnTimeUntil:  record.OnTimeUntil,
		LateUntil:    record.LateUntil,
		IsActive:     record.IsActive,
	}
	return &response, nil
}

func (s *Service) UpdateAttendanceRule(id string, input UpsertAttendanceRuleRequest) (*AttendanceRuleResponse, error) {
	record, err := s.findAttendanceRuleByID(id)
	if err != nil {
		return nil, err
	}
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueRecord("attendance_rules", "school_year_id", schoolYearRecord.ID, id); err != nil {
		return nil, err
	}

	record.SchoolYearID = schoolYearRecord.ID
	record.CheckInStart = strings.TrimSpace(input.CheckInStart)
	record.OnTimeUntil = strings.TrimSpace(input.OnTimeUntil)
	record.LateUntil = strings.TrimSpace(input.LateUntil)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)
	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update attendance rule: %w", err)
	}

	response := AttendanceRuleResponse{
		ID:           record.ID,
		SchoolYearID: record.SchoolYearID,
		SchoolYear:   schoolYearRecord.Name,
		CheckInStart: record.CheckInStart,
		OnTimeUntil:  record.OnTimeUntil,
		LateUntil:    record.LateUntil,
		IsActive:     record.IsActive,
	}
	return &response, nil
}

func (s *Service) DeleteAttendanceRule(id string) error {
	record, err := s.findAttendanceRuleByID(id)
	if err != nil {
		return err
	}
	if err := s.db.Delete(record).Error; err != nil {
		return fmt.Errorf("delete attendance rule: %w", err)
	}
	return nil
}

func (s *Service) ListAdminAttendance(date, status, classID string) ([]attendanceModule.AttendanceRecordResponse, error) {
	return s.attendanceService.ListAdminAttendance(date, status, classID)
}

func (s *Service) ReviewAttendance(id, actorID string, input attendanceModule.ReviewAttendanceRequest) (*attendanceModule.AttendanceRecordResponse, error) {
	return s.attendanceService.ReviewAttendance(id, actorID, input)
}

func (s *Service) ensureUniqueStudentFields(studentID, userID string, input UpsertStudentRequest) error {
	if err := s.checkDuplicateField("nis", strings.TrimSpace(input.NIS), userID); err != nil {
		return err
	}
	if err := s.ensureUniqueRecord("students", "nis", strings.TrimSpace(input.NIS), studentID); err != nil {
		return err
	}
	if trimmedNISN := strings.TrimSpace(input.NISN); trimmedNISN != "" {
		if err := s.ensureUniqueRecord("students", "nisn", trimmedNISN, studentID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureUniqueStudentMembership(excludedID, studentID, classID, schoolYearID string) error {
	query := s.db.Model(&studentModule.StudentClassMembership{}).
		Where("student_id = ? AND class_id = ? AND school_year_id = ?", studentID, classID, schoolYearID)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check duplicate student membership: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("student class membership already exists")
	}
	return nil
}

func (s *Service) ensureSingleActiveMembership(excludedID, studentID, schoolYearID string, isActive bool) error {
	if !isActive {
		return nil
	}
	query := s.db.Model(&studentModule.StudentClassMembership{}).
		Where("student_id = ? AND school_year_id = ? AND is_active = ?", studentID, schoolYearID, true)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check active student membership: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("student already has an active class membership for this school year")
	}
	return nil
}

func (s *Service) findStudentByID(id string) (*studentModule.Student, error) {
	var record studentModule.Student
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentNotFound
		}
		return nil, fmt.Errorf("find student: %w", err)
	}
	return &record, nil
}

func (s *Service) findStudentMembershipByID(id string) (*studentModule.StudentClassMembership, error) {
	var record studentModule.StudentClassMembership
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudentMembershipNotFound
		}
		return nil, fmt.Errorf("find student class membership: %w", err)
	}
	return &record, nil
}

func (s *Service) findAttendanceRuleByID(id string) (*attendanceModule.AttendanceRule, error) {
	var record attendanceModule.AttendanceRule
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAttendanceRuleNotFound
		}
		return nil, fmt.Errorf("find attendance rule: %w", err)
	}
	return &record, nil
}

func (s *Service) findStudentMembershipResponse(id string) (*StudentClassMembershipResponse, error) {
	var row studentMembershipRow
	if err := s.db.Table("student_class_memberships as scm").
		Select("scm.id, scm.student_id, users.name as student_name, students.nis, scm.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, scm.school_year_id, school_years.name as school_year_name, scm.status, scm.joined_at, scm.left_at, scm.is_active").
		Joins("join students on students.id = scm.student_id").
		Joins("join users on users.id = students.user_id").
		Joins("join classes on classes.id = scm.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = scm.school_year_id").
		Where("scm.id = ?", id).
		Scan(&row).Error; err != nil {
		return nil, fmt.Errorf("find student membership response: %w", err)
	}
	if row.ID == "" {
		return nil, ErrStudentMembershipNotFound
	}
	response := mapStudentMembershipRow(row)
	return &response, nil
}

func mapStudentRow(row studentRow) StudentResponse {
	response := StudentResponse{
		ID:        row.ID,
		UserID:    row.UserID,
		Name:      row.Name,
		NIS:       row.NIS,
		EntryYear: row.EntryYear,
		IsActive:  row.IsActive,
	}
	if row.NISN != nil {
		response.NISN = *row.NISN
	}
	if row.Gender != nil {
		response.Gender = *row.Gender
	}
	if row.BirthPlace != nil {
		response.BirthPlace = *row.BirthPlace
	}
	if row.BirthDate != nil {
		response.BirthDate = row.BirthDate.Format("2006-01-02")
	}
	response.BirthPlaceDate = formatBirthPlaceDate(response.BirthPlace, row.BirthDate)
	if row.Address != nil {
		response.Address = *row.Address
	}
	if row.Phone != nil {
		response.Phone = *row.Phone
	}
	if row.ParentName != nil {
		response.ParentName = *row.ParentName
	}
	if row.ParentPhone != nil {
		response.ParentPhone = *row.ParentPhone
	}
	return response
}

func mapStudentMembershipRow(row studentMembershipRow) StudentClassMembershipResponse {
	response := StudentClassMembershipResponse{
		ID:             row.ID,
		StudentID:      row.StudentID,
		StudentName:    row.StudentName,
		NIS:            row.NIS,
		ClassID:        row.ClassID,
		ClassName:      row.ClassName,
		SchoolYearID:   row.SchoolYearID,
		SchoolYearName: row.SchoolYearName,
		Status:         row.Status,
		IsActive:       row.IsActive,
	}
	if row.JoinedAt != nil {
		response.JoinedAt = row.JoinedAt.Format(time.RFC3339)
	}
	if row.LeftAt != nil {
		response.LeftAt = row.LeftAt.Format(time.RFC3339)
	}
	return response
}

func normalizeMembershipStatus(value string) studentModule.MembershipStatus {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case string(studentModule.MembershipStatusActive):
		return studentModule.MembershipStatusActive
	case string(studentModule.MembershipStatusTransferred):
		return studentModule.MembershipStatusTransferred
	case string(studentModule.MembershipStatusGraduated):
		return studentModule.MembershipStatusGraduated
	case string(studentModule.MembershipStatusInactive):
		return studentModule.MembershipStatusInactive
	default:
		return ""
	}
}

func parseOptionalDate(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", trimmed)
	if err != nil {
		return nil, fmt.Errorf("birth_date must be in YYYY-MM-DD format")
	}
	return &parsed, nil
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

func parseOptionalDateTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		return nil, fmt.Errorf("date time value must be in RFC3339 format")
	}
	return &parsed, nil
}
