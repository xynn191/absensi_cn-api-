package admin

import (
	"errors"
	"fmt"
	"strings"

	"absensi-cn-api/internal/modules/academic"
	attendanceModule "absensi-cn-api/internal/modules/attendance"
	studentModule "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTeacherProfileNotFound     = errors.New("teacher profile not found")
	ErrSubjectNotFound            = errors.New("subject not found")
	ErrMajorNotFound              = errors.New("major not found")
	ErrSchoolYearNotFound         = errors.New("school year not found")
	ErrClassNotFound              = errors.New("class not found")
	ErrTeacherAssignmentNotFound  = errors.New("teacher subject assignment not found")
	ErrHomeroomAssignmentNotFound = errors.New("homeroom assignment not found")
)

type teacherProfileRow struct {
	ID       string
	UserID   string
	Name     string
	Username *string
	NIP      *string
	NUPTK    *string
	Gender   *string
	Phone    *string
	Address  *string
	IsActive bool
}

type subjectRow struct {
	ID           string
	Code         string
	Name         string
	SubjectGroup *string
	IsActive     bool
}

type majorRow struct {
	ID       string
	Code     string
	Name     string
	IsActive bool
}

type schoolYearRow struct {
	ID        string
	Name      string
	StartYear int
	EndYear   int
	IsActive  bool
}

type classRow struct {
	ID             string
	Grade          string
	Name           string
	MajorID        string
	MajorCode      string
	MajorName      string
	SchoolYearID   string
	SchoolYearName string
	IsActive       bool
}

type teacherSubjectAssignmentRow struct {
	ID             string
	TeacherID      string
	TeacherName    string
	SubjectID      string
	SubjectCode    string
	SubjectName    string
	ClassID        string
	ClassName      string
	SchoolYearID   string
	SchoolYearName string
	IsActive       bool
}

type homeroomAssignmentRow struct {
	ID             string
	TeacherID      string
	TeacherName    string
	ClassID        string
	ClassName      string
	SchoolYearID   string
	SchoolYearName string
	IsActive       bool
}

func (s *Service) ListTeacherProfiles() ([]TeacherProfileResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []teacherProfileRow
	if err := s.db.Table("teachers").
		Select("teachers.id, teachers.user_id, users.name, users.username, teachers.nip, teachers.nuptk, teachers.gender, teachers.phone, teachers.address, teachers.is_active").
		Joins("join users on users.id = teachers.user_id").
		Order("users.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teacher profiles: %w", err)
	}

	return mapTeacherProfiles(rows), nil
}

func (s *Service) CreateTeacherProfile(input UpsertTeacherRequest) (*TeacherProfileResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	account, err := s.requireTeacherUser(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, err
	}

	if err := s.ensureTeacherProfileUniqueFields("", input); err != nil {
		return nil, err
	}

	var existing int64
	if err := s.db.Model(&academic.Teacher{}).Where("user_id = ?", account.ID).Count(&existing).Error; err != nil {
		return nil, fmt.Errorf("check existing teacher profile: %w", err)
	}
	if existing > 0 {
		return nil, fmt.Errorf("teacher profile for user_id already exists")
	}

	record := academic.Teacher{
		ID:       uuid.NewString(),
		UserID:   account.ID,
		NIP:      optionalString(input.NIP),
		NUPTK:    optionalString(input.NUPTK),
		Gender:   optionalUpperString(input.Gender),
		Phone:    optionalString(input.Phone),
		Address:  optionalString(input.Address),
		IsActive: boolValueOrDefault(input.IsActive, true),
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create teacher profile: %w", err)
	}

	response := mapTeacherProfileRow(teacherProfileRow{
		ID:       record.ID,
		UserID:   record.UserID,
		Name:     account.Name,
		Username: account.Username,
		NIP:      record.NIP,
		NUPTK:    record.NUPTK,
		Gender:   record.Gender,
		Phone:    record.Phone,
		Address:  record.Address,
		IsActive: record.IsActive,
	})

	return &response, nil
}

func (s *Service) UpdateTeacherProfile(id string, input UpsertTeacherRequest) (*TeacherProfileResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	record, err := s.findTeacherProfileByID(id)
	if err != nil {
		return nil, err
	}

	account, err := s.requireTeacherUser(strings.TrimSpace(input.UserID))
	if err != nil {
		return nil, err
	}

	if err := s.ensureTeacherProfileUniqueFields(id, input); err != nil {
		return nil, err
	}

	record.UserID = account.ID
	record.NIP = optionalString(input.NIP)
	record.NUPTK = optionalString(input.NUPTK)
	record.Gender = optionalUpperString(input.Gender)
	record.Phone = optionalString(input.Phone)
	record.Address = optionalString(input.Address)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update teacher profile: %w", err)
	}

	response := mapTeacherProfileRow(teacherProfileRow{
		ID:       record.ID,
		UserID:   record.UserID,
		Name:     account.Name,
		Username: account.Username,
		NIP:      record.NIP,
		NUPTK:    record.NUPTK,
		Gender:   record.Gender,
		Phone:    record.Phone,
		Address:  record.Address,
		IsActive: record.IsActive,
	})

	return &response, nil
}

func (s *Service) DeleteTeacherProfile(id string) error {
	if s.db == nil {
		return ErrAdminDataUnavailable
	}

	record, err := s.findTeacherProfileByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("teacher_id = ?", record.ID).Delete(&academic.TeacherSubjectAssignment{}).Error; err != nil {
			return fmt.Errorf("delete teacher subject assignments: %w", err)
		}
		if err := tx.Where("teacher_id = ?", record.ID).Delete(&academic.HomeroomAssignment{}).Error; err != nil {
			return fmt.Errorf("delete homeroom assignments: %w", err)
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete teacher profile: %w", err)
		}
		return nil
	})
}

func (s *Service) ListSubjects() ([]SubjectResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []subjectRow
	if err := s.db.Table("subjects").Order("code asc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list subjects: %w", err)
	}

	result := make([]SubjectResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, SubjectResponse{
			ID:       row.ID,
			Code:     row.Code,
			Name:     row.Name,
			Group:    dereference(row.SubjectGroup),
			IsActive: row.IsActive,
		})
	}

	return result, nil
}

func (s *Service) CreateSubject(input UpsertSubjectRequest) (*SubjectResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	if err := s.ensureUniqueRecord("subjects", "code", strings.TrimSpace(input.Code), ""); err != nil {
		return nil, err
	}

	record := academic.Subject{
		ID:           uuid.NewString(),
		Code:         strings.ToUpper(strings.TrimSpace(input.Code)),
		Name:         strings.TrimSpace(input.Name),
		SubjectGroup: optionalString(input.Group),
		IsActive:     boolValueOrDefault(input.IsActive, true),
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create subject: %w", err)
	}

	response := SubjectResponse{
		ID:       record.ID,
		Code:     record.Code,
		Name:     record.Name,
		Group:    dereference(record.SubjectGroup),
		IsActive: record.IsActive,
	}
	return &response, nil
}

func (s *Service) UpdateSubject(id string, input UpsertSubjectRequest) (*SubjectResponse, error) {
	record, err := s.findSubjectByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueRecord("subjects", "code", strings.TrimSpace(input.Code), id); err != nil {
		return nil, err
	}

	record.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	record.Name = strings.TrimSpace(input.Name)
	record.SubjectGroup = optionalString(input.Group)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update subject: %w", err)
	}

	response := SubjectResponse{
		ID:       record.ID,
		Code:     record.Code,
		Name:     record.Name,
		Group:    dereference(record.SubjectGroup),
		IsActive: record.IsActive,
	}
	return &response, nil
}

func (s *Service) DeleteSubject(id string) error {
	record, err := s.findSubjectByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("subject_id = ?", record.ID).Delete(&academic.TeacherSubjectAssignment{}).Error; err != nil {
			return fmt.Errorf("delete teacher subject assignments by subject: %w", err)
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete subject: %w", err)
		}
		return nil
	})
}

func (s *Service) ListMajors() ([]MajorResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []majorRow
	if err := s.db.Table("majors").Order("code asc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list majors: %w", err)
	}

	result := make([]MajorResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, MajorResponse{
			ID:       row.ID,
			Code:     row.Code,
			Name:     row.Name,
			IsActive: row.IsActive,
		})
	}

	return result, nil
}

func (s *Service) CreateMajor(input UpsertMajorRequest) (*MajorResponse, error) {
	if err := s.ensureUniqueRecord("majors", "code", strings.TrimSpace(input.Code), ""); err != nil {
		return nil, err
	}

	record := academic.Major{
		ID:       uuid.NewString(),
		Code:     strings.ToUpper(strings.TrimSpace(input.Code)),
		Name:     strings.TrimSpace(input.Name),
		IsActive: boolValueOrDefault(input.IsActive, true),
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create major: %w", err)
	}

	response := MajorResponse{
		ID:       record.ID,
		Code:     record.Code,
		Name:     record.Name,
		IsActive: record.IsActive,
	}
	return &response, nil
}

func (s *Service) UpdateMajor(id string, input UpsertMajorRequest) (*MajorResponse, error) {
	record, err := s.findMajorByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueRecord("majors", "code", strings.TrimSpace(input.Code), id); err != nil {
		return nil, err
	}

	record.Code = strings.ToUpper(strings.TrimSpace(input.Code))
	record.Name = strings.TrimSpace(input.Name)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update major: %w", err)
	}

	response := MajorResponse{
		ID:       record.ID,
		Code:     record.Code,
		Name:     record.Name,
		IsActive: record.IsActive,
	}
	return &response, nil
}

func (s *Service) DeleteMajor(id string) error {
	record, err := s.findMajorByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		classIDs, err := s.classIDsByField(tx, "major_id", record.ID)
		if err != nil {
			return err
		}
		if err := s.deleteClassDependencies(tx, classIDs); err != nil {
			return err
		}
		if len(classIDs) > 0 {
			if err := tx.Where("id IN ?", classIDs).Delete(&academic.SchoolClass{}).Error; err != nil {
				return fmt.Errorf("delete classes by major: %w", err)
			}
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete major: %w", err)
		}
		return nil
	})
}

func (s *Service) ListSchoolYears() ([]SchoolYearResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var rows []schoolYearRow
	if err := s.db.Table("school_years").Order("start_year desc").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list school years: %w", err)
	}

	result := make([]SchoolYearResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, SchoolYearResponse{
			ID:        row.ID,
			Name:      row.Name,
			StartYear: row.StartYear,
			EndYear:   row.EndYear,
			IsActive:  row.IsActive,
		})
	}

	return result, nil
}

func (s *Service) CreateSchoolYear(input UpsertSchoolYearRequest) (*SchoolYearResponse, error) {
	if err := s.ensureUniqueRecord("school_years", "name", strings.TrimSpace(input.Name), ""); err != nil {
		return nil, err
	}

	record := academic.SchoolYear{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(input.Name),
		StartYear: input.StartYear,
		EndYear:   input.EndYear,
		IsActive:  boolValueOrDefault(input.IsActive, true),
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create school year: %w", err)
	}

	response := SchoolYearResponse{
		ID:        record.ID,
		Name:      record.Name,
		StartYear: record.StartYear,
		EndYear:   record.EndYear,
		IsActive:  record.IsActive,
	}
	return &response, nil
}

func (s *Service) UpdateSchoolYear(id string, input UpsertSchoolYearRequest) (*SchoolYearResponse, error) {
	record, err := s.findSchoolYearByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueRecord("school_years", "name", strings.TrimSpace(input.Name), id); err != nil {
		return nil, err
	}

	record.Name = strings.TrimSpace(input.Name)
	record.StartYear = input.StartYear
	record.EndYear = input.EndYear
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update school year: %w", err)
	}

	response := SchoolYearResponse{
		ID:        record.ID,
		Name:      record.Name,
		StartYear: record.StartYear,
		EndYear:   record.EndYear,
		IsActive:  record.IsActive,
	}
	return &response, nil
}

func (s *Service) DeleteSchoolYear(id string) error {
	record, err := s.findSchoolYearByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		classIDs, err := s.classIDsByField(tx, "school_year_id", record.ID)
		if err != nil {
			return err
		}
		if err := s.deleteClassDependencies(tx, classIDs); err != nil {
			return err
		}
		if len(classIDs) > 0 {
			if err := tx.Where("id IN ?", classIDs).Delete(&academic.SchoolClass{}).Error; err != nil {
				return fmt.Errorf("delete classes by school year: %w", err)
			}
		}
		if err := tx.Where("school_year_id = ?", record.ID).Delete(&academic.TeacherSubjectAssignment{}).Error; err != nil {
			return fmt.Errorf("delete teacher subject assignments by school year: %w", err)
		}
		if err := tx.Where("school_year_id = ?", record.ID).Delete(&academic.HomeroomAssignment{}).Error; err != nil {
			return fmt.Errorf("delete homeroom assignments by school year: %w", err)
		}
		if err := tx.Where("school_year_id = ?", record.ID).Delete(&studentModule.StudentClassMembership{}).Error; err != nil {
			return fmt.Errorf("delete student class memberships by school year: %w", err)
		}
		if err := tx.Where("school_year_id = ?", record.ID).Delete(&attendanceModule.AttendanceRecord{}).Error; err != nil {
			return fmt.Errorf("delete attendance records by school year: %w", err)
		}
		if err := tx.Where("school_year_id = ?", record.ID).Delete(&attendanceModule.AttendanceRule{}).Error; err != nil {
			return fmt.Errorf("delete attendance rules by school year: %w", err)
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete school year: %w", err)
		}
		return nil
	})
}

func (s *Service) ListClasses() ([]ClassResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	rows, err := s.queryClasses()
	if err != nil {
		return nil, err
	}

	return mapClasses(rows), nil
}

func (s *Service) CreateClass(input UpsertClassRequest) (*ClassResponse, error) {
	majorRecord, err := s.findMajorByID(strings.TrimSpace(input.MajorID))
	if err != nil {
		return nil, err
	}
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}

	record := academic.SchoolClass{
		ID:           uuid.NewString(),
		Grade:        strings.ToUpper(strings.TrimSpace(input.Grade)),
		Name:         strings.TrimSpace(input.Name),
		MajorID:      majorRecord.ID,
		SchoolYearID: schoolYearRecord.ID,
		IsActive:     boolValueOrDefault(input.IsActive, true),
	}

	if err := s.ensureUniqueClassIdentity(record, ""); err != nil {
		return nil, err
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create class: %w", err)
	}

	response := mapClassRow(classRow{
		ID:             record.ID,
		Grade:          record.Grade,
		Name:           record.Name,
		MajorID:        record.MajorID,
		MajorCode:      majorRecord.Code,
		MajorName:      majorRecord.Name,
		SchoolYearID:   record.SchoolYearID,
		SchoolYearName: schoolYearRecord.Name,
		IsActive:       record.IsActive,
	})
	return &response, nil
}

func (s *Service) UpdateClass(id string, input UpsertClassRequest) (*ClassResponse, error) {
	record, err := s.findClassByID(id)
	if err != nil {
		return nil, err
	}
	majorRecord, err := s.findMajorByID(strings.TrimSpace(input.MajorID))
	if err != nil {
		return nil, err
	}
	schoolYearRecord, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID))
	if err != nil {
		return nil, err
	}

	record.Grade = strings.ToUpper(strings.TrimSpace(input.Grade))
	record.Name = strings.TrimSpace(input.Name)
	record.MajorID = majorRecord.ID
	record.SchoolYearID = schoolYearRecord.ID
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.ensureUniqueClassIdentity(*record, id); err != nil {
		return nil, err
	}

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update class: %w", err)
	}

	response := mapClassRow(classRow{
		ID:             record.ID,
		Grade:          record.Grade,
		Name:           record.Name,
		MajorID:        record.MajorID,
		MajorCode:      majorRecord.Code,
		MajorName:      majorRecord.Name,
		SchoolYearID:   record.SchoolYearID,
		SchoolYearName: schoolYearRecord.Name,
		IsActive:       record.IsActive,
	})
	return &response, nil
}

func (s *Service) DeleteClass(id string) error {
	record, err := s.findClassByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.deleteClassDependencies(tx, []string{record.ID}); err != nil {
			return err
		}
		if err := tx.Delete(record).Error; err != nil {
			return fmt.Errorf("delete class: %w", err)
		}
		return nil
	})
}

func (s *Service) ListTeacherSubjectAssignments() ([]TeacherSubjectAssignmentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	rows, err := s.queryTeacherSubjectAssignments()
	if err != nil {
		return nil, err
	}

	result := make([]TeacherSubjectAssignmentResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, TeacherSubjectAssignmentResponse{
			ID:             row.ID,
			TeacherID:      row.TeacherID,
			TeacherName:    row.TeacherName,
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

func (s *Service) CreateTeacherSubjectAssignment(input UpsertTeacherSubjectAssignmentRequest) (*TeacherSubjectAssignmentResponse, error) {
	teacherRecord, err := s.findTeacherProfileByID(strings.TrimSpace(input.TeacherID))
	if err != nil {
		return nil, err
	}
	subjectRecord, err := s.findSubjectByID(strings.TrimSpace(input.SubjectID))
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

	record := academic.TeacherSubjectAssignment{
		ID:           uuid.NewString(),
		TeacherID:    teacherRecord.ID,
		SubjectID:    subjectRecord.ID,
		ClassID:      classRecord.ID,
		SchoolYearID: schoolYearRecord.ID,
		IsActive:     boolValueOrDefault(input.IsActive, true),
	}

	if err := s.ensureUniqueTeacherSubjectAssignment(record, ""); err != nil {
		return nil, err
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create teacher subject assignment: %w", err)
	}

	result, err := s.findTeacherSubjectAssignmentResponse(record.ID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) UpdateTeacherSubjectAssignment(id string, input UpsertTeacherSubjectAssignmentRequest) (*TeacherSubjectAssignmentResponse, error) {
	record, err := s.findTeacherSubjectAssignmentByID(id)
	if err != nil {
		return nil, err
	}
	teacherRecord, err := s.findTeacherProfileByID(strings.TrimSpace(input.TeacherID))
	if err != nil {
		return nil, err
	}
	if _, err := s.findSubjectByID(strings.TrimSpace(input.SubjectID)); err != nil {
		return nil, err
	}
	classRecord, err := s.findClassByID(strings.TrimSpace(input.ClassID))
	if err != nil {
		return nil, err
	}
	if _, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID)); err != nil {
		return nil, err
	}
	if classRecord.SchoolYearID != strings.TrimSpace(input.SchoolYearID) {
		return nil, fmt.Errorf("class does not belong to the provided school year")
	}

	record.TeacherID = teacherRecord.ID
	record.SubjectID = strings.TrimSpace(input.SubjectID)
	record.ClassID = classRecord.ID
	record.SchoolYearID = strings.TrimSpace(input.SchoolYearID)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.ensureUniqueTeacherSubjectAssignment(*record, id); err != nil {
		return nil, err
	}

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update teacher subject assignment: %w", err)
	}

	return s.findTeacherSubjectAssignmentResponse(record.ID)
}

func (s *Service) DeleteTeacherSubjectAssignment(id string) error {
	record, err := s.findTeacherSubjectAssignmentByID(id)
	if err != nil {
		return err
	}
	if err := s.db.Delete(record).Error; err != nil {
		return fmt.Errorf("delete teacher subject assignment: %w", err)
	}
	return nil
}

func (s *Service) ListHomeroomAssignments() ([]HomeroomAssignmentResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	rows, err := s.queryHomeroomAssignments()
	if err != nil {
		return nil, err
	}

	result := make([]HomeroomAssignmentResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, HomeroomAssignmentResponse{
			ID:             row.ID,
			TeacherID:      row.TeacherID,
			TeacherName:    row.TeacherName,
			ClassID:        row.ClassID,
			ClassName:      row.ClassName,
			SchoolYearID:   row.SchoolYearID,
			SchoolYearName: row.SchoolYearName,
			IsActive:       row.IsActive,
		})
	}

	return result, nil
}

func (s *Service) CreateHomeroomAssignment(input UpsertHomeroomAssignmentRequest) (*HomeroomAssignmentResponse, error) {
	if _, err := s.findTeacherProfileByID(strings.TrimSpace(input.TeacherID)); err != nil {
		return nil, err
	}
	classRecord, err := s.findClassByID(strings.TrimSpace(input.ClassID))
	if err != nil {
		return nil, err
	}
	if _, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID)); err != nil {
		return nil, err
	}
	if classRecord.SchoolYearID != strings.TrimSpace(input.SchoolYearID) {
		return nil, fmt.Errorf("class does not belong to the provided school year")
	}

	record := academic.HomeroomAssignment{
		ID:           uuid.NewString(),
		TeacherID:    strings.TrimSpace(input.TeacherID),
		ClassID:      classRecord.ID,
		SchoolYearID: strings.TrimSpace(input.SchoolYearID),
		IsActive:     boolValueOrDefault(input.IsActive, true),
	}

	if err := s.ensureUniqueHomeroomAssignment(record, ""); err != nil {
		return nil, err
	}

	if err := s.db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create homeroom assignment: %w", err)
	}

	return s.findHomeroomAssignmentResponse(record.ID)
}

func (s *Service) UpdateHomeroomAssignment(id string, input UpsertHomeroomAssignmentRequest) (*HomeroomAssignmentResponse, error) {
	record, err := s.findHomeroomAssignmentByID(id)
	if err != nil {
		return nil, err
	}
	if _, err := s.findTeacherProfileByID(strings.TrimSpace(input.TeacherID)); err != nil {
		return nil, err
	}
	classRecord, err := s.findClassByID(strings.TrimSpace(input.ClassID))
	if err != nil {
		return nil, err
	}
	if _, err := s.findSchoolYearByID(strings.TrimSpace(input.SchoolYearID)); err != nil {
		return nil, err
	}
	if classRecord.SchoolYearID != strings.TrimSpace(input.SchoolYearID) {
		return nil, fmt.Errorf("class does not belong to the provided school year")
	}

	record.TeacherID = strings.TrimSpace(input.TeacherID)
	record.ClassID = classRecord.ID
	record.SchoolYearID = strings.TrimSpace(input.SchoolYearID)
	record.IsActive = boolValueOrDefault(input.IsActive, record.IsActive)

	if err := s.ensureUniqueHomeroomAssignment(*record, id); err != nil {
		return nil, err
	}

	if err := s.db.Save(record).Error; err != nil {
		return nil, fmt.Errorf("update homeroom assignment: %w", err)
	}

	return s.findHomeroomAssignmentResponse(record.ID)
}

func (s *Service) DeleteHomeroomAssignment(id string) error {
	record, err := s.findHomeroomAssignmentByID(id)
	if err != nil {
		return err
	}
	if err := s.db.Delete(record).Error; err != nil {
		return fmt.Errorf("delete homeroom assignment: %w", err)
	}
	return nil
}

func (s *Service) ensureTeacherProfileUniqueFields(excludedID string, input UpsertTeacherRequest) error {
	if trimmedNIP := strings.TrimSpace(input.NIP); trimmedNIP != "" {
		if err := s.ensureUniqueRecord("teachers", "nip", trimmedNIP, excludedID); err != nil {
			return err
		}
	}
	if trimmedNUPTK := strings.TrimSpace(input.NUPTK); trimmedNUPTK != "" {
		if err := s.ensureUniqueRecord("teachers", "nuptk", trimmedNUPTK, excludedID); err != nil {
			return err
		}
	}
	if trimmedUserID := strings.TrimSpace(input.UserID); trimmedUserID != "" {
		if err := s.ensureUniqueRecord("teachers", "user_id", trimmedUserID, excludedID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ensureUniqueRecord(table, field, value, excludedID string) error {
	query := s.db.Table(table).Where(field+" = ?", value)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check duplicate %s.%s: %w", table, field, err)
	}
	if count > 0 {
		return fmt.Errorf("%s already exists", field)
	}
	return nil
}

func (s *Service) ensureUniqueClassIdentity(record academic.SchoolClass, excludedID string) error {
	query := s.db.Model(&academic.SchoolClass{}).
		Where("grade = ? AND major_id = ? AND name = ? AND school_year_id = ?", record.Grade, record.MajorID, record.Name, record.SchoolYearID)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check duplicate class: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("class already exists for the same major and school year")
	}
	return nil
}

func (s *Service) ensureUniqueTeacherSubjectAssignment(record academic.TeacherSubjectAssignment, excludedID string) error {
	query := s.db.Model(&academic.TeacherSubjectAssignment{}).
		Where("teacher_id = ? AND subject_id = ? AND class_id = ? AND school_year_id = ?", record.TeacherID, record.SubjectID, record.ClassID, record.SchoolYearID)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check duplicate teacher subject assignment: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("teacher subject assignment already exists")
	}
	return nil
}

func (s *Service) ensureUniqueHomeroomAssignment(record academic.HomeroomAssignment, excludedID string) error {
	var classYearCount int64
	classYearQuery := s.db.Model(&academic.HomeroomAssignment{}).
		Where("class_id = ? AND school_year_id = ?", record.ClassID, record.SchoolYearID)
	if excludedID != "" {
		classYearQuery = classYearQuery.Where("id <> ?", excludedID)
	}
	if err := classYearQuery.Count(&classYearCount).Error; err != nil {
		return fmt.Errorf("check duplicate homeroom class/year: %w", err)
	}
	if classYearCount > 0 {
		return fmt.Errorf("homeroom assignment for this class and school year already exists")
	}

	var teacherClassCount int64
	teacherClassQuery := s.db.Model(&academic.HomeroomAssignment{}).
		Where("teacher_id = ? AND class_id = ? AND school_year_id = ?", record.TeacherID, record.ClassID, record.SchoolYearID)
	if excludedID != "" {
		teacherClassQuery = teacherClassQuery.Where("id <> ?", excludedID)
	}
	if err := teacherClassQuery.Count(&teacherClassCount).Error; err != nil {
		return fmt.Errorf("check duplicate homeroom teacher/class: %w", err)
	}
	if teacherClassCount > 0 {
		return fmt.Errorf("homeroom assignment already exists")
	}

	return nil
}

func (s *Service) requireTeacherUser(userID string) (*user.User, error) {
	account, err := s.findUserByID(userID)
	if err != nil {
		return nil, err
	}
	if account.Role != user.RoleTeacher {
		return nil, fmt.Errorf("user must have TEACHER role")
	}
	return account, nil
}

func (s *Service) findTeacherProfileByID(id string) (*academic.Teacher, error) {
	var record academic.Teacher
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeacherProfileNotFound
		}
		return nil, fmt.Errorf("find teacher profile: %w", err)
	}
	return &record, nil
}

func (s *Service) findSubjectByID(id string) (*academic.Subject, error) {
	var record academic.Subject
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubjectNotFound
		}
		return nil, fmt.Errorf("find subject: %w", err)
	}
	return &record, nil
}

func (s *Service) findMajorByID(id string) (*academic.Major, error) {
	var record academic.Major
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMajorNotFound
		}
		return nil, fmt.Errorf("find major: %w", err)
	}
	return &record, nil
}

func (s *Service) findSchoolYearByID(id string) (*academic.SchoolYear, error) {
	var record academic.SchoolYear
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSchoolYearNotFound
		}
		return nil, fmt.Errorf("find school year: %w", err)
	}
	return &record, nil
}

func (s *Service) findClassByID(id string) (*academic.SchoolClass, error) {
	var record academic.SchoolClass
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, fmt.Errorf("find class: %w", err)
	}
	return &record, nil
}

func (s *Service) findTeacherSubjectAssignmentByID(id string) (*academic.TeacherSubjectAssignment, error) {
	var record academic.TeacherSubjectAssignment
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTeacherAssignmentNotFound
		}
		return nil, fmt.Errorf("find teacher subject assignment: %w", err)
	}
	return &record, nil
}

func (s *Service) findHomeroomAssignmentByID(id string) (*academic.HomeroomAssignment, error) {
	var record academic.HomeroomAssignment
	if err := s.db.First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrHomeroomAssignmentNotFound
		}
		return nil, fmt.Errorf("find homeroom assignment: %w", err)
	}
	return &record, nil
}

func (s *Service) queryClasses() ([]classRow, error) {
	var rows []classRow
	if err := s.db.Table("classes").
		Select("classes.id, classes.grade, classes.name, classes.major_id, majors.code as major_code, majors.name as major_name, classes.school_year_id, school_years.name as school_year_name, classes.is_active").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = classes.school_year_id").
		Order("school_years.start_year desc, classes.grade asc, majors.code asc, classes.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list classes: %w", err)
	}
	return rows, nil
}

func (s *Service) queryTeacherSubjectAssignments() ([]teacherSubjectAssignmentRow, error) {
	var rows []teacherSubjectAssignmentRow
	if err := s.db.Table("teacher_subject_assignments as tsa").
		Select("tsa.id, tsa.teacher_id, users.name as teacher_name, tsa.subject_id, subjects.code as subject_code, subjects.name as subject_name, tsa.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, tsa.school_year_id, school_years.name as school_year_name, tsa.is_active").
		Joins("join teachers on teachers.id = tsa.teacher_id").
		Joins("join users on users.id = teachers.user_id").
		Joins("join subjects on subjects.id = tsa.subject_id").
		Joins("join classes on classes.id = tsa.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = tsa.school_year_id").
		Order("school_years.start_year desc, users.name asc, subjects.code asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teacher subject assignments: %w", err)
	}
	return rows, nil
}

func (s *Service) queryHomeroomAssignments() ([]homeroomAssignmentRow, error) {
	var rows []homeroomAssignmentRow
	if err := s.db.Table("homeroom_assignments as ha").
		Select("ha.id, ha.teacher_id, users.name as teacher_name, ha.class_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as class_name, ha.school_year_id, school_years.name as school_year_name, ha.is_active").
		Joins("join teachers on teachers.id = ha.teacher_id").
		Joins("join users on users.id = teachers.user_id").
		Joins("join classes on classes.id = ha.class_id").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = ha.school_year_id").
		Order("school_years.start_year desc, class_name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list homeroom assignments: %w", err)
	}
	return rows, nil
}

func (s *Service) findTeacherSubjectAssignmentResponse(id string) (*TeacherSubjectAssignmentResponse, error) {
	rows, err := s.queryTeacherSubjectAssignments()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == id {
			response := TeacherSubjectAssignmentResponse{
				ID:             row.ID,
				TeacherID:      row.TeacherID,
				TeacherName:    row.TeacherName,
				SubjectID:      row.SubjectID,
				SubjectCode:    row.SubjectCode,
				SubjectName:    row.SubjectName,
				ClassID:        row.ClassID,
				ClassName:      row.ClassName,
				SchoolYearID:   row.SchoolYearID,
				SchoolYearName: row.SchoolYearName,
				IsActive:       row.IsActive,
			}
			return &response, nil
		}
	}
	return nil, ErrTeacherAssignmentNotFound
}

func (s *Service) findHomeroomAssignmentResponse(id string) (*HomeroomAssignmentResponse, error) {
	rows, err := s.queryHomeroomAssignments()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ID == id {
			response := HomeroomAssignmentResponse{
				ID:             row.ID,
				TeacherID:      row.TeacherID,
				TeacherName:    row.TeacherName,
				ClassID:        row.ClassID,
				ClassName:      row.ClassName,
				SchoolYearID:   row.SchoolYearID,
				SchoolYearName: row.SchoolYearName,
				IsActive:       row.IsActive,
			}
			return &response, nil
		}
	}
	return nil, ErrHomeroomAssignmentNotFound
}

func (s *Service) classIDsByField(tx *gorm.DB, field, value string) ([]string, error) {
	var classIDs []string
	if err := tx.Model(&academic.SchoolClass{}).Where(field+" = ?", value).Pluck("id", &classIDs).Error; err != nil {
		return nil, fmt.Errorf("find class ids by %s: %w", field, err)
	}
	return classIDs, nil
}

func (s *Service) deleteClassDependencies(tx *gorm.DB, classIDs []string) error {
	if len(classIDs) == 0 {
		return nil
	}
	if err := tx.Where("class_id IN ?", classIDs).Delete(&academic.TeacherSubjectAssignment{}).Error; err != nil {
		return fmt.Errorf("delete teacher subject assignments by class: %w", err)
	}
	if err := tx.Where("class_id IN ?", classIDs).Delete(&academic.HomeroomAssignment{}).Error; err != nil {
		return fmt.Errorf("delete homeroom assignments by class: %w", err)
	}
	if err := tx.Where("class_id IN ?", classIDs).Delete(&studentModule.StudentClassMembership{}).Error; err != nil {
		return fmt.Errorf("delete student class memberships by class: %w", err)
	}
	if err := tx.Where("class_id IN ?", classIDs).Delete(&attendanceModule.AttendanceRecord{}).Error; err != nil {
		return fmt.Errorf("delete attendance records by class: %w", err)
	}
	return nil
}

func mapTeacherProfiles(rows []teacherProfileRow) []TeacherProfileResponse {
	result := make([]TeacherProfileResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapTeacherProfileRow(row))
	}
	return result
}

func mapTeacherProfileRow(row teacherProfileRow) TeacherProfileResponse {
	return TeacherProfileResponse{
		ID:       row.ID,
		UserID:   row.UserID,
		Name:     row.Name,
		Username: dereference(row.Username),
		NIP:      dereference(row.NIP),
		NUPTK:    dereference(row.NUPTK),
		Gender:   dereference(row.Gender),
		Phone:    dereference(row.Phone),
		Address:  dereference(row.Address),
		IsActive: row.IsActive,
	}
}

func mapClasses(rows []classRow) []ClassResponse {
	result := make([]ClassResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapClassRow(row))
	}
	return result
}

func mapClassRow(row classRow) ClassResponse {
	displayName := strings.TrimSpace(fmt.Sprintf("%s %s %s", row.Grade, row.MajorCode, row.Name))
	return ClassResponse{
		ID:             row.ID,
		Grade:          row.Grade,
		Name:           row.Name,
		MajorID:        row.MajorID,
		MajorCode:      row.MajorCode,
		MajorName:      row.MajorName,
		SchoolYearID:   row.SchoolYearID,
		SchoolYearName: row.SchoolYearName,
		DisplayName:    displayName,
		IsActive:       row.IsActive,
	}
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalUpperString(value string) *string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func dereference(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func boolValueOrDefault(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}
