package database

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"absensi-cn-api/internal/config"
	"absensi-cn-api/internal/modules/academic"
	attendanceModule "absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/internal/modules/counseling"
	leaveModule "absensi-cn-api/internal/modules/leave"
	studentModule "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/password"

	"github.com/google/uuid"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type seedTeacherData struct {
	Name     string
	Username string
	Password string
	NIP      string
	NUPTK    string
	Gender   string
	Phone    string
	Address  string
}

type seedStudentData struct {
	Name        string
	NIS         string
	NISN        string
	Password    string
	Gender      string
	BirthPlace  string
	BirthDate   string
	Address     string
	Phone       string
	ParentName  string
	ParentPhone string
	EntryYear   int
	ClassKey    string
}

type seedAcademicContext struct {
	SchoolYear *academic.SchoolYear
	ClassByKey map[string]*academic.SchoolClass
}

func NewMySQL(cfg *config.Config) (*gorm.DB, error) {
	if !cfg.Database.Enabled {
		slog.Warn("database connection skipped because DB_ENABLED=false")
		return nil, nil
	}

	if err := ensureDatabaseExists(cfg); err != nil {
		return nil, err
	}

	dsn := buildDatabaseDSN(cfg)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newGormLogger(cfg),
	})
	if err != nil {
		return nil, fmt.Errorf("connect mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	slog.Info("database connected", "driver", "mysql")

	if err := autoMigrate(db); err != nil {
		return nil, err
	}

	if err := migrateAcademicColumns(db); err != nil {
		return nil, err
	}

	if err := migrateLegacyClassrooms(db); err != nil {
		return nil, err
	}

	if err := migrateLegacyRoles(db); err != nil {
		return nil, err
	}

	if err := seedInitialData(db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureDatabaseExists(cfg *config.Config) error {
	sqlDB, err := sql.Open("mysql", buildServerDSN(cfg))
	if err != nil {
		return fmt.Errorf("connect mysql server: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping mysql server: %w", err)
	}

	databaseName := strings.ReplaceAll(cfg.Database.Name, "`", "")
	query := fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		databaseName,
	)

	if _, err := sqlDB.Exec(query); err != nil {
		return fmt.Errorf("create database %s: %w", cfg.Database.Name, err)
	}

	return nil
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&user.User{},
		&academic.Teacher{},
		&academic.Subject{},
		&academic.Major{},
		&academic.SchoolYear{},
		&academic.SchoolClass{},
		&academic.TeacherSubjectAssignment{},
		&academic.HomeroomAssignment{},
		&studentModule.Student{},
		&studentModule.StudentClassMembership{},
		&attendanceModule.AttendanceRule{},
		&attendanceModule.AttendanceRecord{},
		&leaveModule.Submission{},
		&counseling.Note{},
	); err != nil {
		return fmt.Errorf("auto migrate schema: %w", err)
	}

	return nil
}

func seedInitialData(db *gorm.DB) error {
	defaultUsers := []struct {
		ID       string
		Name     string
		Role     user.Role
		NIS      *string
		Username *string
		Password string
	}{
		{
			ID:       uuid.NewString(),
			Name:     "Rafi Akbar",
			Role:     user.RoleStudent,
			NIS:      stringPtr("1234567890"),
			Username: stringPtr("siswa"),
			Password: "121212",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Administrator",
			Role:     user.RoleAdmin,
			Username: stringPtr("admin"),
			Password: "121212",
		},
		{
			ID:       uuid.NewString(),
			Name:     "Guru BK",
			Role:     user.RoleBK,
			Username: stringPtr("bk"),
			Password: "121212",
		},
	}

	for _, seededUser := range defaultUsers {
		if _, err := ensureSeedUser(db, seededUser.Name, seededUser.Role, seededUser.NIS, seededUser.Username, seededUser.Password); err != nil {
			return err
		}
	}

	defaultTeachers := []seedTeacherData{
		{
			Name:     "Wali Kelas",
			Username: "walas",
			Password: "121212",
			NIP:      "198402102008011000",
			NUPTK:    "1122334455999",
			Gender:   "FEMALE",
			Phone:    "081234567899",
			Address:  "Jl. Pendidikan No. 10",
		},
		{
			Name:     "Ahmad Fauzi",
			Username: "guru1",
			Password: "121212",
			NIP:      "198706152010011001",
			NUPTK:    "1122334455001",
			Gender:   "MALE",
			Phone:    "081234567801",
			Address:  "Jl. Melati No. 12",
		},
		{
			Name:     "Siti Rahmawati",
			Username: "guru2",
			Password: "121212",
			NIP:      "198809202011012002",
			NUPTK:    "1122334455002",
			Gender:   "FEMALE",
			Phone:    "081234567802",
			Address:  "Jl. Kenanga No. 8",
		},
		{
			Name:     "Budi Santoso",
			Username: "guru3",
			Password: "121212",
			NIP:      "198512052009011003",
			NUPTK:    "1122334455003",
			Gender:   "MALE",
			Phone:    "081234567803",
			Address:  "Jl. Anggrek No. 21",
		},
		{
			Name:     "Dewi Lestari",
			Username: "guru4",
			Password: "121212",
			NIP:      "199001172014022004",
			NUPTK:    "1122334455004",
			Gender:   "FEMALE",
			Phone:    "081234567804",
			Address:  "Jl. Flamboyan No. 5",
		},
		{
			Name:     "Rizky Pratama",
			Username: "guru5",
			Password: "121212",
			NIP:      "199203112015031005",
			NUPTK:    "1122334455005",
			Gender:   "MALE",
			Phone:    "081234567805",
			Address:  "Jl. Mawar No. 30",
		},
	}

	for _, seededTeacher := range defaultTeachers {
		account, err := ensureSeedUser(
			db,
			seededTeacher.Name,
			user.RoleTeacher,
			nil,
			stringPtr(seededTeacher.Username),
			seededTeacher.Password,
		)
		if err != nil {
			return err
		}

		if err := ensureSeedTeacherProfile(db, account.ID, seededTeacher); err != nil {
			return err
		}
	}

	academicContext, err := seedAcademicReferenceData(db)
	if err != nil {
		return err
	}

	defaultStudents := []seedStudentData{
		{
			Name:        "Rafi Akbar",
			NIS:         "1234567890",
			NISN:        "998877660001",
			Password:    "121212",
			Gender:      "MALE",
			BirthPlace:  "Bekasi",
			BirthDate:   "2010-01-15",
			Address:     "Jl. Pendidikan No. 1",
			Phone:       "081300000001",
			ParentName:  "Bapak Akbar",
			ParentPhone: "081311111111",
			EntryYear:   2026,
			ClassKey:    "X-PPLG-1",
		},
		{
			Name:        "Salsa Aulia",
			NIS:         "1234567896",
			NISN:        "998877660007",
			Password:    "121212",
			Gender:      "FEMALE",
			BirthPlace:  "Bekasi",
			BirthDate:   "2010-08-11",
			Address:     "Jl. Flamboyan No. 18",
			Phone:       "081300000007",
			ParentName:  "Ibu Salsa",
			ParentPhone: "081311111117",
			EntryYear:   2026,
			ClassKey:    "X-PPLG-1",
		},
		{
			Name:        "Dimas Prakoso",
			NIS:         "1234567897",
			NISN:        "998877660008",
			Password:    "121212",
			Gender:      "MALE",
			BirthPlace:  "Jakarta",
			BirthDate:   "2010-10-03",
			Address:     "Jl. Sawo No. 6",
			Phone:       "081300000008",
			ParentName:  "Bapak Dimas",
			ParentPhone: "081311111118",
			EntryYear:   2026,
			ClassKey:    "X-PPLG-1",
		},
		{
			Name:        "Nabila Putri",
			NIS:         "1234567891",
			NISN:        "998877660002",
			Password:    "121212",
			Gender:      "FEMALE",
			BirthPlace:  "Jakarta",
			BirthDate:   "2010-03-02",
			Address:     "Jl. Cendana No. 9",
			Phone:       "081300000002",
			ParentName:  "Ibu Nabila",
			ParentPhone: "081311111112",
			EntryYear:   2026,
			ClassKey:    "X-TJKT-1",
		},
		{
			Name:        "Farhan Ramadhan",
			NIS:         "1234567892",
			NISN:        "998877660003",
			Password:    "121212",
			Gender:      "MALE",
			BirthPlace:  "Bogor",
			BirthDate:   "2009-09-18",
			Address:     "Jl. Pahlawan No. 4",
			Phone:       "081300000003",
			ParentName:  "Bapak Farhan",
			ParentPhone: "081311111113",
			EntryYear:   2025,
			ClassKey:    "XI-DKV-1",
		},
		{
			Name:        "Putri Maharani",
			NIS:         "1234567893",
			NISN:        "998877660004",
			Password:    "121212",
			Gender:      "FEMALE",
			BirthPlace:  "Depok",
			BirthDate:   "2009-11-27",
			Address:     "Jl. Karya No. 7",
			Phone:       "081300000004",
			ParentName:  "Ibu Putri",
			ParentPhone: "081311111114",
			EntryYear:   2025,
			ClassKey:    "XI-MPLB-1",
		},
		{
			Name:        "Raka Saputra",
			NIS:         "1234567894",
			NISN:        "998877660005",
			Password:    "121212",
			Gender:      "MALE",
			BirthPlace:  "Tangerang",
			BirthDate:   "2008-07-06",
			Address:     "Jl. Nusantara No. 15",
			Phone:       "081300000005",
			ParentName:  "Bapak Raka",
			ParentPhone: "081311111115",
			EntryYear:   2024,
			ClassKey:    "XII-PM-1",
		},
		{
			Name:        "Citra Lestari",
			NIS:         "1234567895",
			NISN:        "998877660006",
			Password:    "121212",
			Gender:      "FEMALE",
			BirthPlace:  "Bandung",
			BirthDate:   "2010-05-22",
			Address:     "Jl. Angsana No. 11",
			Phone:       "081300000006",
			ParentName:  "Ibu Citra",
			ParentPhone: "081311111116",
			EntryYear:   2026,
			ClassKey:    "X-PH-1",
		},
	}

	for _, seededStudent := range defaultStudents {
		account, err := ensureSeedUser(
			db,
			seededStudent.Name,
			user.RoleStudent,
			stringPtr(seededStudent.NIS),
			nil,
			seededStudent.Password,
		)
		if err != nil {
			return err
		}

		studentRecord, err := ensureSeedStudentProfile(db, account, seededStudent)
		if err != nil {
			return err
		}

		classRecord := academicContext.ClassByKey[seededStudent.ClassKey]
		if classRecord == nil {
			return fmt.Errorf("seed class %s is not available for student %s", seededStudent.ClassKey, seededStudent.Name)
		}

		if err := ensureSeedStudentMembership(db, studentRecord.ID, classRecord.ID, academicContext.SchoolYear.ID, true); err != nil {
			return err
		}
	}

	if err := ensureSeedAttendanceRule(db, academicContext.SchoolYear.ID, attendanceModule.AttendanceRule{
		CheckInStart: "06:30:00",
		OnTimeUntil:  "07:00:00",
		LateUntil:    "07:30:00",
		IsActive:     true,
	}); err != nil {
		return err
	}

	if err := seedSupportCaseData(db); err != nil {
		return err
	}

	return nil
}

func ensureSeedUser(
	db *gorm.DB,
	name string,
	role user.Role,
	nis *string,
	username *string,
	rawPassword string,
) (*user.User, error) {
	var existing user.User

	normalizedNIS := normalizeOptionalString(nis)
	normalizedUsername := normalizeOptionalString(username)

	query := db.Model(&user.User{})
	switch {
	case normalizedUsername != nil && normalizedNIS != nil:
		query = query.Where("username = ? OR nis = ?", *normalizedUsername, *normalizedNIS)
	case normalizedUsername != nil:
		query = query.Where("username = ?", *normalizedUsername)
	case normalizedNIS != nil:
		query = query.Where("nis = ?", *normalizedNIS)
	default:
		query = query.Where("role = ? AND name = ?", role, strings.TrimSpace(name))
	}

	err := query.First(&existing).Error
	if err == nil {
		existing.Name = strings.TrimSpace(name)
		existing.Role = role
		existing.NIS = normalizedNIS
		existing.Username = normalizedUsername
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("sync seed user %s: %w", name, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed user %s: %w", name, err)
	}

	hashedPassword, err := password.Hash(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("hash seed password for %s: %w", name, err)
	}

	record := user.User{
		ID:           uuid.NewString(),
		Name:         strings.TrimSpace(name),
		NIS:          normalizedNIS,
		Username:     normalizedUsername,
		PasswordHash: hashedPassword,
		Role:         role,
	}

	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed user %s: %w", name, err)
	}

	return &record, nil
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func ensureSeedTeacherProfile(
	db *gorm.DB,
	userID string,
	teacher seedTeacherData,
) error {
	var existing academic.Teacher
	err := db.Where("user_id = ?", userID).First(&existing).Error
	if err == nil {
		existing.NIP = stringPtr(teacher.NIP)
		existing.NUPTK = stringPtr(teacher.NUPTK)
		existing.Gender = stringPtr(teacher.Gender)
		existing.Phone = stringPtr(teacher.Phone)
		existing.Address = stringPtr(teacher.Address)
		existing.IsActive = true

		if err := db.Save(&existing).Error; err != nil {
			return fmt.Errorf("update seed teacher profile for user %s: %w", userID, err)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed teacher profile for user %s: %w", userID, err)
	}

	record := academic.Teacher{
		ID:       uuid.NewString(),
		UserID:   userID,
		NIP:      stringPtr(teacher.NIP),
		NUPTK:    stringPtr(teacher.NUPTK),
		Gender:   stringPtr(teacher.Gender),
		Phone:    stringPtr(teacher.Phone),
		Address:  stringPtr(teacher.Address),
		IsActive: true,
	}

	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed teacher profile for user %s: %w", userID, err)
	}

	return nil
}

func seedAcademicReferenceData(db *gorm.DB) (*seedAcademicContext, error) {
	schoolYear, err := ensureSeedSchoolYear(db, "2026/2027", 2026, 2027, true)
	if err != nil {
		return nil, err
	}

	majors := []struct {
		Code string
		Name string
	}{
		{Code: "PPLG", Name: "Pengembangan Perangkat Lunak dan Gim"},
		{Code: "TJKT", Name: "Teknik Jaringan Komputer dan Telekomunikasi"},
		{Code: "DKV", Name: "Desain Komunikasi Visual"},
		{Code: "MPLB", Name: "Manajemen Perkantoran dan Layanan Bisnis"},
		{Code: "PM", Name: "Pemasaran"},
		{Code: "PH", Name: "Perhotelan"},
	}

	majorByCode := make(map[string]*academic.Major, len(majors))
	for _, item := range majors {
		record, err := ensureSeedMajor(db, item.Code, item.Name, true)
		if err != nil {
			return nil, err
		}
		majorByCode[item.Code] = record
	}

	subjects := []struct {
		Code  string
		Name  string
		Group string
	}{
		{Code: "MTK", Name: "Matematika", Group: "Umum"},
		{Code: "BIND", Name: "Bahasa Indonesia", Group: "Umum"},
		{Code: "PWEB", Name: "Pemrograman Web", Group: "Kejuruan"},
		{Code: "TJAR", Name: "Administrasi Jaringan", Group: "Kejuruan"},
		{Code: "DKV1", Name: "Desain Grafis", Group: "Kejuruan"},
	}

	subjectByCode := make(map[string]*academic.Subject, len(subjects))
	for _, item := range subjects {
		record, err := ensureSeedSubject(db, item.Code, item.Name, item.Group, true)
		if err != nil {
			return nil, err
		}
		subjectByCode[item.Code] = record
	}

	classes := []struct {
		Grade     string
		MajorCode string
		Name      string
	}{
		{Grade: "X", MajorCode: "PPLG", Name: "1"},
		{Grade: "X", MajorCode: "TJKT", Name: "1"},
		{Grade: "XI", MajorCode: "DKV", Name: "1"},
		{Grade: "XI", MajorCode: "MPLB", Name: "1"},
		{Grade: "XII", MajorCode: "PM", Name: "1"},
		{Grade: "X", MajorCode: "PH", Name: "1"},
	}

	classByKey := make(map[string]*academic.SchoolClass, len(classes))
	for _, item := range classes {
		majorRecord := majorByCode[item.MajorCode]
		record, err := ensureSeedClass(db, item.Grade, majorRecord.ID, item.Name, schoolYear.ID, true)
		if err != nil {
			return nil, err
		}
		classByKey[item.Grade+"-"+item.MajorCode+"-"+item.Name] = record
	}

	assignments := []struct {
		Username    string
		SubjectCode string
		ClassKey    string
		IsHomeroom  bool
	}{
		{Username: "guru1", SubjectCode: "MTK", ClassKey: "X-PPLG-1", IsHomeroom: false},
		{Username: "walas", SubjectCode: "PWEB", ClassKey: "X-PPLG-1", IsHomeroom: true},
		{Username: "guru2", SubjectCode: "BIND", ClassKey: "X-TJKT-1", IsHomeroom: true},
		{Username: "guru3", SubjectCode: "PWEB", ClassKey: "XI-DKV-1", IsHomeroom: true},
		{Username: "guru4", SubjectCode: "TJAR", ClassKey: "XI-MPLB-1", IsHomeroom: true},
		{Username: "guru5", SubjectCode: "DKV1", ClassKey: "XII-PM-1", IsHomeroom: true},
	}

	for _, item := range assignments {
		var (
			teacherRecord *academic.Teacher
			err           error
		)
		if item.Username == "walas" {
			teacherRecord, err = findPrimaryHomeroomTeacher(db)
		} else {
			teacherRecord, err = findTeacherByUsername(db, item.Username)
		}
		if err != nil {
			return nil, err
		}

		subjectRecord := subjectByCode[item.SubjectCode]
		classRecord := classByKey[item.ClassKey]

		if err := ensureSeedTeacherSubjectAssignment(db, teacherRecord.ID, subjectRecord.ID, classRecord.ID, schoolYear.ID, true); err != nil {
			return nil, err
		}

		if item.IsHomeroom {
			if err := ensureSeedHomeroomAssignment(db, teacherRecord.ID, classRecord.ID, schoolYear.ID, true); err != nil {
				return nil, err
			}
		}
	}

	return &seedAcademicContext{
		SchoolYear: schoolYear,
		ClassByKey: classByKey,
	}, nil
}

func ensureSeedMajor(db *gorm.DB, code, name string, isActive bool) (*academic.Major, error) {
	var existing academic.Major
	err := db.Where("code = ?", strings.ToUpper(strings.TrimSpace(code))).First(&existing).Error
	if err == nil {
		existing.Name = strings.TrimSpace(name)
		existing.IsActive = isActive
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update seed major %s: %w", code, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed major %s: %w", code, err)
	}

	record := academic.Major{
		ID:       uuid.NewString(),
		Code:     strings.ToUpper(strings.TrimSpace(code)),
		Name:     strings.TrimSpace(name),
		IsActive: isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed major %s: %w", code, err)
	}
	return &record, nil
}

func ensureSeedSubject(db *gorm.DB, code, name, group string, isActive bool) (*academic.Subject, error) {
	var existing academic.Subject
	err := db.Where("code = ?", strings.ToUpper(strings.TrimSpace(code))).First(&existing).Error
	if err == nil {
		existing.Name = strings.TrimSpace(name)
		existing.SubjectGroup = stringPtr(group)
		existing.IsActive = isActive
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update seed subject %s: %w", code, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed subject %s: %w", code, err)
	}

	record := academic.Subject{
		ID:           uuid.NewString(),
		Code:         strings.ToUpper(strings.TrimSpace(code)),
		Name:         strings.TrimSpace(name),
		SubjectGroup: stringPtr(group),
		IsActive:     isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed subject %s: %w", code, err)
	}
	return &record, nil
}

func ensureSeedSchoolYear(db *gorm.DB, name string, startYear, endYear int, isActive bool) (*academic.SchoolYear, error) {
	var existing academic.SchoolYear
	err := db.Where("name = ?", strings.TrimSpace(name)).First(&existing).Error
	if err == nil {
		existing.StartYear = startYear
		existing.EndYear = endYear
		existing.IsActive = isActive
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update seed school year %s: %w", name, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed school year %s: %w", name, err)
	}

	record := academic.SchoolYear{
		ID:        uuid.NewString(),
		Name:      strings.TrimSpace(name),
		StartYear: startYear,
		EndYear:   endYear,
		IsActive:  isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed school year %s: %w", name, err)
	}
	return &record, nil
}

func ensureSeedClass(db *gorm.DB, grade, majorID, name, schoolYearID string, isActive bool) (*academic.SchoolClass, error) {
	var existing academic.SchoolClass
	err := db.Where("grade = ? AND major_id = ? AND name = ? AND school_year_id = ?", grade, majorID, name, schoolYearID).First(&existing).Error
	if err == nil {
		existing.IsActive = isActive
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update seed class %s %s: %w", grade, name, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed class %s %s: %w", grade, name, err)
	}

	record := academic.SchoolClass{
		ID:           uuid.NewString(),
		Grade:        strings.ToUpper(strings.TrimSpace(grade)),
		MajorID:      majorID,
		Name:         strings.TrimSpace(name),
		SchoolYearID: schoolYearID,
		IsActive:     isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed class %s %s: %w", grade, name, err)
	}
	return &record, nil
}

func findTeacherByUsername(db *gorm.DB, username string) (*academic.Teacher, error) {
	var account user.User
	if err := db.Where("username = ?", strings.TrimSpace(username)).First(&account).Error; err != nil {
		return nil, fmt.Errorf("find teacher user by username %s: %w", username, err)
	}

	var teacherRecord academic.Teacher
	if err := db.Where("user_id = ?", account.ID).First(&teacherRecord).Error; err != nil {
		return nil, fmt.Errorf("find teacher profile by username %s: %w", username, err)
	}

	return &teacherRecord, nil
}

func findPrimaryHomeroomTeacher(db *gorm.DB) (*academic.Teacher, error) {
	var teacherRecord academic.Teacher
	err := db.Table("teachers").
		Select("teachers.*").
		Joins("join users on users.id = teachers.user_id").
		Where("users.role = ?", user.RoleTeacher).
		Where("users.username = ? OR users.name LIKE ?", "walas", "Wali Kelas%").
		Order("case when users.name like 'Wali Kelas%' then 0 when users.username = 'walas' then 1 else 2 end").
		Order("teachers.id").
		First(&teacherRecord).Error
	if err != nil {
		return nil, fmt.Errorf("find primary homeroom teacher: %w", err)
	}

	return &teacherRecord, nil
}

func ensureSeedTeacherSubjectAssignment(db *gorm.DB, teacherID, subjectID, classID, schoolYearID string, isActive bool) error {
	var existing academic.TeacherSubjectAssignment
	err := db.Where("teacher_id = ? AND subject_id = ? AND class_id = ? AND school_year_id = ?", teacherID, subjectID, classID, schoolYearID).First(&existing).Error
	if err == nil {
		existing.IsActive = isActive
		return db.Save(&existing).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed teacher subject assignment: %w", err)
	}

	record := academic.TeacherSubjectAssignment{
		ID:           uuid.NewString(),
		TeacherID:    teacherID,
		SubjectID:    subjectID,
		ClassID:      classID,
		SchoolYearID: schoolYearID,
		IsActive:     isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed teacher subject assignment: %w", err)
	}
	return nil
}

func ensureSeedHomeroomAssignment(db *gorm.DB, teacherID, classID, schoolYearID string, isActive bool) error {
	var existing academic.HomeroomAssignment
	err := db.Where("class_id = ? AND school_year_id = ?", classID, schoolYearID).First(&existing).Error
	if err == nil {
		existing.TeacherID = teacherID
		existing.IsActive = isActive
		return db.Save(&existing).Error
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed homeroom assignment: %w", err)
	}

	record := academic.HomeroomAssignment{
		ID:           uuid.NewString(),
		TeacherID:    teacherID,
		ClassID:      classID,
		SchoolYearID: schoolYearID,
		IsActive:     isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed homeroom assignment: %w", err)
	}
	return nil
}

func ensureSeedStudentProfile(db *gorm.DB, account *user.User, seededStudent seedStudentData) (*studentModule.Student, error) {
	var existing studentModule.Student
	err := db.Where("user_id = ?", account.ID).First(&existing).Error

	birthDate, parseErr := time.Parse("2006-01-02", seededStudent.BirthDate)
	if parseErr != nil {
		return nil, fmt.Errorf("parse seed student birth date for %s: %w", seededStudent.Name, parseErr)
	}

	if err == nil {
		existing.NIS = seededStudent.NIS
		existing.NISN = stringPtr(seededStudent.NISN)
		existing.Gender = stringPtr(seededStudent.Gender)
		existing.BirthPlace = stringPtr(seededStudent.BirthPlace)
		existing.BirthDate = &birthDate
		existing.Address = stringPtr(seededStudent.Address)
		existing.Phone = stringPtr(seededStudent.Phone)
		existing.ParentName = stringPtr(seededStudent.ParentName)
		existing.ParentPhone = stringPtr(seededStudent.ParentPhone)
		existing.EntryYear = seededStudent.EntryYear
		existing.IsActive = true
		if err := db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("update seed student profile for user %s: %w", account.ID, err)
		}
		return &existing, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("find seed student profile for user %s: %w", account.ID, err)
	}

	record := studentModule.Student{
		ID:          uuid.NewString(),
		UserID:      account.ID,
		NIS:         seededStudent.NIS,
		NISN:        stringPtr(seededStudent.NISN),
		Gender:      stringPtr(seededStudent.Gender),
		BirthPlace:  stringPtr(seededStudent.BirthPlace),
		BirthDate:   &birthDate,
		Address:     stringPtr(seededStudent.Address),
		Phone:       stringPtr(seededStudent.Phone),
		ParentName:  stringPtr(seededStudent.ParentName),
		ParentPhone: stringPtr(seededStudent.ParentPhone),
		EntryYear:   seededStudent.EntryYear,
		IsActive:    true,
	}

	if err := db.Create(&record).Error; err != nil {
		return nil, fmt.Errorf("create seed student profile for user %s: %w", account.ID, err)
	}

	return &record, nil
}

func ensureSeedStudentMembership(db *gorm.DB, studentID, classID, schoolYearID string, isActive bool) error {
	var existing studentModule.StudentClassMembership
	err := db.Where("student_id = ? AND school_year_id = ? AND is_active = ?", studentID, schoolYearID, true).First(&existing).Error
	if err == nil {
		existing.ClassID = classID
		existing.Status = studentModule.MembershipStatusActive
		existing.IsActive = isActive
		existing.LeftAt = nil
		if err := db.Save(&existing).Error; err != nil {
			return fmt.Errorf("update seed student membership: %w", err)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed student membership: %w", err)
	}

	record := studentModule.StudentClassMembership{
		ID:           uuid.NewString(),
		StudentID:    studentID,
		ClassID:      classID,
		SchoolYearID: schoolYearID,
		Status:       studentModule.MembershipStatusActive,
		IsActive:     isActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed student membership: %w", err)
	}
	return nil
}

func ensureSeedAttendanceRule(db *gorm.DB, schoolYearID string, rule attendanceModule.AttendanceRule) error {
	var existing attendanceModule.AttendanceRule
	err := db.Where("school_year_id = ?", schoolYearID).First(&existing).Error
	if err == nil {
		existing.CheckInStart = rule.CheckInStart
		existing.OnTimeUntil = rule.OnTimeUntil
		existing.LateUntil = rule.LateUntil
		existing.IsActive = rule.IsActive
		if err := db.Save(&existing).Error; err != nil {
			return fmt.Errorf("update seed attendance rule: %w", err)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed attendance rule: %w", err)
	}

	record := attendanceModule.AttendanceRule{
		ID:           uuid.NewString(),
		SchoolYearID: schoolYearID,
		CheckInStart: rule.CheckInStart,
		OnTimeUntil:  rule.OnTimeUntil,
		LateUntil:    rule.LateUntil,
		IsActive:     rule.IsActive,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed attendance rule: %w", err)
	}
	return nil
}

func seedSupportCaseData(db *gorm.DB) error {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return fmt.Errorf("load jakarta location for seed support data: %w", err)
	}

	studentDemo, err := findStudentByNIS(db, "1234567890")
	if err != nil {
		return err
	}
	studentSalsa, err := findStudentByNIS(db, "1234567896")
	if err != nil {
		return err
	}
	studentDimas, err := findStudentByNIS(db, "1234567897")
	if err != nil {
		return err
	}
	studentNabila, err := findStudentByNIS(db, "1234567891")
	if err != nil {
		return err
	}

	adminUser, err := findUserByUsername(db, "admin")
	if err != nil {
		return err
	}
	bkUser, err := findUserByUsername(db, "bk")
	if err != nil {
		return err
	}

	if err := ensureSeedSubmission(db, studentDemo.ID, "SAKIT", "Demam dan perlu istirahat di rumah", "", leaveModule.StatusPending, nil, nil, nil); err != nil {
		return err
	}
	if err := ensureSeedSubmission(db, studentSalsa.ID, "IZIN", "Mengikuti keperluan keluarga pada pagi hari", "", leaveModule.StatusPending, nil, nil, nil); err != nil {
		return err
	}
	if err := ensureSeedSubmission(db, studentNabila.ID, "IZIN", "Menghadiri acara keluarga pada pagi hari", "", leaveModule.StatusApproved, &adminUser.ID, stringPtr("Disetujui untuk hari ini"), timePtr(time.Now())); err != nil {
		return err
	}
	if err := ensureSeedCounselingNote(db, studentDemo.ID, bkUser.ID, "Monitoring Keterlambatan", "Siswa perlu dipantau karena beberapa kali datang mepet batas waktu absensi."); err != nil {
		return err
	}

	now := time.Now().In(location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	yesterday := today.AddDate(0, 0, -1)
	twoDaysAgo := today.AddDate(0, 0, -2)

	if err := ensureSeedStudentAttendance(db, studentDemo.ID, today, attendanceModule.StatusHadir, timePtr(time.Date(today.Year(), today.Month(), today.Day(), 6, 42, 0, 0, location)), stringPtr("Hadir tepat waktu")); err != nil {
		return err
	}
	if err := ensureSeedStudentAttendance(db, studentDemo.ID, yesterday, attendanceModule.StatusTelat, timePtr(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 7, 11, 0, 0, location)), stringPtr("Datang melewati batas hadir")); err != nil {
		return err
	}
	if err := ensureSeedStudentAttendance(db, studentDemo.ID, twoDaysAgo, attendanceModule.StatusHadir, timePtr(time.Date(twoDaysAgo.Year(), twoDaysAgo.Month(), twoDaysAgo.Day(), 6, 48, 0, 0, location)), stringPtr("Check-in normal")); err != nil {
		return err
	}

	if err := ensureSeedStudentAttendance(db, studentSalsa.ID, today, attendanceModule.StatusTelat, timePtr(time.Date(today.Year(), today.Month(), today.Day(), 7, 14, 0, 0, location)), stringPtr("Masuk setelah jam 07.00")); err != nil {
		return err
	}
	if err := ensureSeedStudentAttendance(db, studentSalsa.ID, yesterday, attendanceModule.StatusHadir, timePtr(time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 6, 39, 0, 0, location)), stringPtr("Datang lebih awal")); err != nil {
		return err
	}

	if err := ensureSeedStudentAttendance(db, studentDimas.ID, today, attendanceModule.StatusAlfa, nil, stringPtr("Belum melakukan check-in")); err != nil {
		return err
	}
	if err := ensureSeedStudentAttendance(db, studentDimas.ID, yesterday, attendanceModule.StatusSakit, nil, stringPtr("Izin sakit dari orang tua")); err != nil {
		return err
	}

	return nil
}

func findStudentByNIS(db *gorm.DB, nis string) (*studentModule.Student, error) {
	var record studentModule.Student
	if err := db.Where("nis = ?", strings.TrimSpace(nis)).First(&record).Error; err != nil {
		return nil, fmt.Errorf("find student by nis %s: %w", nis, err)
	}
	return &record, nil
}

func findUserByUsername(db *gorm.DB, username string) (*user.User, error) {
	var account user.User
	if err := db.Where("username = ?", strings.TrimSpace(username)).First(&account).Error; err != nil {
		return nil, fmt.Errorf("find user by username %s: %w", username, err)
	}
	return &account, nil
}

func ensureSeedSubmission(db *gorm.DB, studentID, submissionType, reason, attachment string, status leaveModule.Status, reviewedBy *string, reviewNote *string, reviewedAt *time.Time) error {
	var existing leaveModule.Submission
	err := db.Where("student_id = ? AND type = ? AND reason = ?", studentID, strings.TrimSpace(submissionType), strings.TrimSpace(reason)).First(&existing).Error
	if err == nil {
		existing.Attachment = strings.TrimSpace(attachment)
		existing.Status = status
		existing.ReviewedBy = reviewedBy
		existing.ReviewNote = reviewNote
		existing.ReviewedAt = reviewedAt
		if err := db.Save(&existing).Error; err != nil {
			return fmt.Errorf("update seed submission: %w", err)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed submission: %w", err)
	}

	record := leaveModule.Submission{
		ID:         uuid.NewString(),
		StudentID:  studentID,
		Type:       strings.TrimSpace(submissionType),
		Reason:     strings.TrimSpace(reason),
		Attachment: strings.TrimSpace(attachment),
		Status:     status,
		ReviewedBy: reviewedBy,
		ReviewNote: reviewNote,
		ReviewedAt: reviewedAt,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed submission: %w", err)
	}
	return nil
}

func ensureSeedStudentAttendance(
	db *gorm.DB,
	studentID string,
	attendanceDate time.Time,
	status attendanceModule.AttendanceStatus,
	checkInAt *time.Time,
	notes *string,
) error {
	var membership studentModule.StudentClassMembership
	if err := db.
		Where("student_id = ? AND is_active = ?", studentID, true).
		Order("updated_at desc").
		First(&membership).Error; err != nil {
		return fmt.Errorf("find active membership for seeded attendance student %s: %w", studentID, err)
	}

	targetDate := time.Date(attendanceDate.Year(), attendanceDate.Month(), attendanceDate.Day(), 0, 0, 0, 0, attendanceDate.Location())

	var existing attendanceModule.AttendanceRecord
	err := db.Where("student_id = ? AND attendance_date = ?", studentID, targetDate).First(&existing).Error
	if err == nil {
		existing.StudentClassMembershipID = membership.ID
		existing.ClassID = membership.ClassID
		existing.SchoolYearID = membership.SchoolYearID
		existing.CheckInAt = checkInAt
		existing.Status = status
		existing.Notes = notes
		if saveErr := db.Save(&existing).Error; saveErr != nil {
			return fmt.Errorf("update seed attendance record for student %s: %w", studentID, saveErr)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed attendance record for student %s: %w", studentID, err)
	}

	record := attendanceModule.AttendanceRecord{
		ID:                       uuid.NewString(),
		StudentID:                studentID,
		StudentClassMembershipID: membership.ID,
		ClassID:                  membership.ClassID,
		SchoolYearID:             membership.SchoolYearID,
		AttendanceDate:           targetDate,
		CheckInAt:                checkInAt,
		Status:                   status,
		Notes:                    notes,
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed attendance record for student %s: %w", studentID, err)
	}

	return nil
}

func ensureSeedCounselingNote(db *gorm.DB, studentID, createdBy, title, note string) error {
	var existing counseling.Note
	err := db.Where("student_id = ? AND created_by = ? AND title = ?", studentID, createdBy, strings.TrimSpace(title)).First(&existing).Error
	if err == nil {
		existing.Note = strings.TrimSpace(note)
		if err := db.Save(&existing).Error; err != nil {
			return fmt.Errorf("update seed counseling note: %w", err)
		}
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find seed counseling note: %w", err)
	}

	record := counseling.Note{
		ID:        uuid.NewString(),
		StudentID: studentID,
		CreatedBy: createdBy,
		Title:     strings.TrimSpace(title),
		Note:      strings.TrimSpace(note),
	}
	if err := db.Create(&record).Error; err != nil {
		return fmt.Errorf("create seed counseling note: %w", err)
	}
	return nil
}

func timePtr(value time.Time) *time.Time {
	return &value
}

func migrateLegacyRoles(db *gorm.DB) error {
	if err := db.Model(&user.User{}).
		Where("role = ?", "HOMEROOM_TEACHER").
		Update("role", string(user.RoleTeacher)).Error; err != nil {
		return fmt.Errorf("migrate legacy teacher roles: %w", err)
	}

	return nil
}

func migrateAcademicColumns(db *gorm.DB) error {
	migrator := db.Migrator()

	if migrator.HasTable(&academic.Teacher{}) {
		hasNIP := migrator.HasColumn(&academic.Teacher{}, "nip")
		hasLegacyNIP := migrator.HasColumn(&academic.Teacher{}, "n_ip")
		if !hasNIP && hasLegacyNIP {
			if err := migrator.RenameColumn(&academic.Teacher{}, "n_ip", "nip"); err != nil {
				return fmt.Errorf("rename teachers.n_ip to nip: %w", err)
			}
		}
		if migrator.HasColumn(&academic.Teacher{}, "nip") && migrator.HasColumn(&academic.Teacher{}, "n_ip") {
			if err := db.Exec("UPDATE teachers SET nip = COALESCE(nip, n_ip) WHERE n_ip IS NOT NULL").Error; err != nil {
				return fmt.Errorf("migrate teachers.n_ip to nip: %w", err)
			}
			if err := migrator.DropColumn(&academic.Teacher{}, "n_ip"); err != nil {
				return fmt.Errorf("drop legacy teachers.n_ip: %w", err)
			}
		}

		hasNUPTK := migrator.HasColumn(&academic.Teacher{}, "nuptk")
		hasLegacyNUPTK := migrator.HasColumn(&academic.Teacher{}, "n_u_p_t_k")
		if !hasNUPTK && hasLegacyNUPTK {
			if err := migrator.RenameColumn(&academic.Teacher{}, "n_u_p_t_k", "nuptk"); err != nil {
				return fmt.Errorf("rename teachers.n_u_p_t_k to nuptk: %w", err)
			}
		}
		if migrator.HasColumn(&academic.Teacher{}, "nuptk") && migrator.HasColumn(&academic.Teacher{}, "n_u_p_t_k") {
			if err := db.Exec("UPDATE teachers SET nuptk = COALESCE(nuptk, n_u_p_t_k) WHERE n_u_p_t_k IS NOT NULL").Error; err != nil {
				return fmt.Errorf("migrate teachers.n_u_p_t_k to nuptk: %w", err)
			}
			if err := migrator.DropColumn(&academic.Teacher{}, "n_u_p_t_k"); err != nil {
				return fmt.Errorf("drop legacy teachers.n_u_p_t_k: %w", err)
			}
		}

		if migrator.HasColumn(&academic.Teacher{}, "status_kepegawaian") {
			if err := migrator.DropColumn(&academic.Teacher{}, "status_kepegawaian"); err != nil {
				return fmt.Errorf("drop teachers.status_kepegawaian: %w", err)
			}
		}
	}

	return nil
}

func migrateLegacyClassrooms(db *gorm.DB) error {
	migrator := db.Migrator()
	if !migrator.HasTable("classrooms") {
		return nil
	}

	if err := migrator.DropTable("classrooms"); err != nil {
		return fmt.Errorf("drop legacy classrooms table: %w", err)
	}

	return nil
}

func buildServerDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/?%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Params,
	)
}

func buildDatabaseDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.Params,
	)
}

func stringPtr(value string) *string {
	return &value
}

func newGormLogger(cfg *config.Config) logger.Interface {
	logLevel := logger.Warn
	slowThreshold := 200 * time.Millisecond

	if cfg.IsProduction() {
		logLevel = logger.Error
		slowThreshold = 5 * time.Second
	}

	return logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             slowThreshold,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)
}
