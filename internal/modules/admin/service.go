package admin

import (
	"errors"
	"fmt"
	"strings"

	"absensi-cn-api/internal/modules/academic"
	"absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/internal/modules/counseling"
	leaveModule "absensi-cn-api/internal/modules/leave"
	studentModule "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/password"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrAdminDataUnavailable = errors.New("admin database is not available")
	ErrUserNotFound         = errors.New("user not found")
	ErrCannotDeleteSelf     = errors.New("admin cannot delete the active account")
)

type Service struct {
	db                *gorm.DB
	attendanceService *attendance.Service
}

func NewService(db *gorm.DB, attendanceService *attendance.Service) *Service {
	return &Service{
		db:                db,
		attendanceService: attendanceService,
	}
}

func (s *Service) GetDashboard() (*DashboardResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	counts, err := s.getCounts()
	if err != nil {
		return nil, err
	}

	todayStatus, err := s.buildTodayStatus()
	if err != nil {
		return nil, err
	}

	attendancePercentage, err := s.calculateAttendancePercentage()
	if err != nil {
		return nil, err
	}

	return &DashboardResponse{
		AttendancePercentage: attendancePercentage,
		Counts:               counts,
		TodayStatus:          todayStatus,
		SemesterTrend: []TrendPoint{
			{Label: "Jul", Present: 88, Late: 12, Alpha: 4},
			{Label: "Agu", Present: 92, Late: 10, Alpha: 3},
			{Label: "Sep", Present: 86, Late: 15, Alpha: 5},
			{Label: "Okt", Present: 94, Late: 9, Alpha: 2},
			{Label: "Nov", Present: 90, Late: 11, Alpha: 4},
			{Label: "Des", Present: 96, Late: 7, Alpha: 1},
		},
		ClassPerformance: []ClassPerformance{
			{ClassName: "10 RPL 2", Percentage: 91, PresentText: "31/34 hadir"},
			{ClassName: "11 MPLB 1", Percentage: 88, PresentText: "29/33 hadir"},
			{ClassName: "10 DKV 1", Percentage: 84, PresentText: "26/31 hadir"},
			{ClassName: "11 TJKT 3", Percentage: 79, PresentText: "24/30 hadir"},
			{ClassName: "12 PM 2", Percentage: 93, PresentText: "28/30 hadir"},
		},
		Announcements: []Announcement{
			{
				ID:          "ann-001",
				Title:       "Perlu tindak lanjut keterlambatan",
				Description: "Beberapa siswa tercatat telat lebih dari tiga kali minggu ini. Admin dapat koordinasi dengan walas dan BK.",
				Tone:        "warning",
			},
			{
				ID:          "ann-002",
				Title:       "Data absensi terbaik bulan ini",
				Description: "Kelas 10 RPL 2 memimpin persentase kehadiran dan menjadi acuan monitoring untuk kelas lain.",
				Tone:        "success",
			},
			{
				ID:          "ann-003",
				Title:       "Sinkronisasi akun staff",
				Description: "Pastikan akun walas, BK, dan admin baru memiliki username yang unik sebelum dibagikan ke pengguna.",
				Tone:        "info",
			},
		},
	}, nil
}

func (s *Service) ListUsers() ([]UserResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	var users []user.User
	if err := s.db.Order("name asc").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return mapUsers(users), nil
}

func (s *Service) ListTeachers() ([]TeacherResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	type teacherListRow struct {
		UserID       string
		Name         string
		NUPTK        *string
		Phone        *string
		HomeroomName *string
	}

	var rows []teacherListRow
	if err := s.db.Table("users").
		Select("users.id as user_id, users.name, teachers.nuptk, teachers.phone, concat(classes.grade, ' ', majors.code, ' ', classes.name) as homeroom_name").
		Joins("left join teachers on teachers.user_id = users.id").
		Joins("left join homeroom_assignments on homeroom_assignments.teacher_id = teachers.id and homeroom_assignments.is_active = ?", true).
		Joins("left join classes on classes.id = homeroom_assignments.class_id").
		Joins("left join majors on majors.id = classes.major_id").
		Where("users.role = ?", user.RoleTeacher).
		Order("users.name asc").
		Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list teachers: %w", err)
	}

	result := make([]TeacherResponse, 0, len(rows))
	for index, row := range rows {
		teacherNumber := index + 1
		displayID := fmt.Sprintf("%03d", teacherNumber)
		avatarLabel := strings.ToUpper(strings.TrimSpace(firstCharacter(row.Name)))
		if avatarLabel == "" {
			avatarLabel = "G"
		}

		roleLabel := "Guru"
		if row.HomeroomName != nil && strings.TrimSpace(*row.HomeroomName) != "" {
			roleLabel = "Guru/Walas"
		}

		result = append(result, TeacherResponse{
			No:          displayID,
			ID:          row.UserID,
			Name:        row.Name,
			Role:        roleLabel,
			Class:       dereference(row.HomeroomName),
			NUPTK:       dereference(row.NUPTK),
			Contact:     dereference(row.Phone),
			AvatarLabel: avatarLabel,
		})
	}

	return result, nil
}

func (s *Service) GetUser(id string) (*UserResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	account, err := s.findUserByID(id)
	if err != nil {
		return nil, err
	}

	response := mapUser(*account)
	return &response, nil
}

func (s *Service) CreateUser(input UpsertUserRequest) (*UserResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	if err := s.ensureUniqueUserFields("", input); err != nil {
		return nil, err
	}

	hashedPassword, err := password.Hash(strings.TrimSpace(input.Password))
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	account := user.User{
		ID:           uuid.NewString(),
		Name:         strings.TrimSpace(input.Name),
		Role:         user.Role(strings.TrimSpace(input.Role)),
		PasswordHash: hashedPassword,
	}

	if trimmedNIS := strings.TrimSpace(input.NIS); trimmedNIS != "" {
		account.NIS = stringPtr(trimmedNIS)
	}

	if trimmedUsername := strings.TrimSpace(input.Username); trimmedUsername != "" {
		account.Username = stringPtr(trimmedUsername)
	}

	if err := s.db.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	response := mapUser(account)
	return &response, nil
}

func (s *Service) UpdateUser(id string, input UpsertUserRequest) (*UserResponse, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	account, err := s.findUserByID(id)
	if err != nil {
		return nil, err
	}

	if err := s.ensureUniqueUserFields(id, input); err != nil {
		return nil, err
	}

	account.Name = strings.TrimSpace(input.Name)
	account.Role = user.Role(strings.TrimSpace(input.Role))

	if trimmedNIS := strings.TrimSpace(input.NIS); trimmedNIS != "" {
		account.NIS = stringPtr(trimmedNIS)
	} else {
		account.NIS = nil
	}

	if trimmedUsername := strings.TrimSpace(input.Username); trimmedUsername != "" {
		account.Username = stringPtr(trimmedUsername)
	} else {
		account.Username = nil
	}

	if trimmedPassword := strings.TrimSpace(input.Password); trimmedPassword != "" {
		hashedPassword, err := password.Hash(trimmedPassword)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		account.PasswordHash = hashedPassword
	}

	if err := s.db.Save(account).Error; err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	response := mapUser(*account)
	return &response, nil
}

func (s *Service) DeleteUser(id, actorID string) error {
	if s.db == nil {
		return ErrAdminDataUnavailable
	}

	if id == actorID {
		return ErrCannotDeleteSelf
	}

	account, err := s.findUserByID(id)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if account.Role == user.RoleTeacher {
			var teacherProfile academic.Teacher
			err := tx.First(&teacherProfile, "user_id = ?", account.ID).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("find teacher profile by user: %w", err)
			}
			if err == nil {
				if err := tx.Where("teacher_id = ?", teacherProfile.ID).Delete(&academic.TeacherSubjectAssignment{}).Error; err != nil {
					return fmt.Errorf("delete teacher subject assignments by user: %w", err)
				}
				if err := tx.Where("teacher_id = ?", teacherProfile.ID).Delete(&academic.HomeroomAssignment{}).Error; err != nil {
					return fmt.Errorf("delete homeroom assignments by user: %w", err)
				}
				if err := tx.Delete(&teacherProfile).Error; err != nil {
					return fmt.Errorf("delete teacher profile by user: %w", err)
				}
			}
		}

		if account.Role == user.RoleStudent {
			var studentProfile studentModule.Student
			err := tx.First(&studentProfile, "user_id = ?", account.ID).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("find student profile by user: %w", err)
			}
			if err == nil {
				if err := tx.Where("student_id = ?", studentProfile.ID).Delete(&studentModule.StudentClassMembership{}).Error; err != nil {
					return fmt.Errorf("delete student memberships by user: %w", err)
				}
				if err := tx.Where("student_id = ?", studentProfile.ID).Delete(&attendance.AttendanceRecord{}).Error; err != nil {
					return fmt.Errorf("delete attendance records by user: %w", err)
				}
				if err := tx.Where("student_id = ?", studentProfile.ID).Delete(&leaveModule.Submission{}).Error; err != nil {
					return fmt.Errorf("delete submissions by user: %w", err)
				}
				if err := tx.Where("student_id = ?", studentProfile.ID).Delete(&counseling.Note{}).Error; err != nil {
					return fmt.Errorf("delete counseling notes by user: %w", err)
				}
				if err := tx.Delete(&studentProfile).Error; err != nil {
					return fmt.Errorf("delete student profile by user: %w", err)
				}
			}
		}

		if err := tx.Delete(account).Error; err != nil {
			return fmt.Errorf("delete user: %w", err)
		}
		return nil
	})
}

func (s *Service) getCounts() (DashboardCounts, error) {
	var counts DashboardCounts

	if err := s.db.Model(&user.User{}).Count(&counts.TotalUsers).Error; err != nil {
		return counts, fmt.Errorf("count total users: %w", err)
	}

	roleCounts := []struct {
		role  user.Role
		store *int64
	}{
		{role: user.RoleStudent, store: &counts.TotalStudents},
		{role: user.RoleTeacher, store: &counts.TotalTeachers},
		{role: user.RoleBK, store: &counts.TotalBK},
		{role: user.RoleAdmin, store: &counts.TotalAdmins},
	}

	for _, item := range roleCounts {
		if err := s.db.Model(&user.User{}).Where("role = ?", item.role).Count(item.store).Error; err != nil {
			return counts, fmt.Errorf("count role %s: %w", item.role, err)
		}
	}

	return counts, nil
}

func (s *Service) calculateAttendancePercentage() (int, error) {
	percentage, err := s.attendanceService.GetAttendancePercentage()
	if err != nil {
		return 0, err
	}

	return percentage, nil
}

func (s *Service) buildTodayStatus() (TodayStatus, error) {
	counts, err := s.attendanceService.GetTodayStatusCounts()
	if err != nil {
		return TodayStatus{}, err
	}

	return TodayStatus{
		Present:    counts.Present,
		Late:       counts.Late,
		Permission: counts.Permission,
		Sick:       counts.Sick,
		Alpha:      counts.Alpha,
	}, nil
}

func (s *Service) ensureUniqueUserFields(excludedID string, input UpsertUserRequest) error {
	if trimmedNIS := strings.TrimSpace(input.NIS); trimmedNIS != "" {
		if err := s.checkDuplicateField("nis", trimmedNIS, excludedID); err != nil {
			return err
		}
	}

	if trimmedUsername := strings.TrimSpace(input.Username); trimmedUsername != "" {
		if err := s.checkDuplicateField("username", trimmedUsername, excludedID); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) checkDuplicateField(field, value, excludedID string) error {
	query := s.db.Model(&user.User{}).Where(field+" = ?", value)
	if excludedID != "" {
		query = query.Where("id <> ?", excludedID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("check duplicate %s: %w", field, err)
	}

	if count > 0 {
		return fmt.Errorf("%s already exists", field)
	}

	return nil
}

func (s *Service) findUserByID(id string) (*user.User, error) {
	var account user.User
	if err := s.db.First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user: %w", err)
	}

	return &account, nil
}

func mapUsers(accounts []user.User) []UserResponse {
	result := make([]UserResponse, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, mapUser(account))
	}

	return result
}

func mapUser(account user.User) UserResponse {
	response := UserResponse{
		ID:   account.ID,
		Name: account.Name,
		Role: string(account.Role),
	}

	if account.NIS != nil {
		response.NIS = *account.NIS
	}

	if account.Username != nil {
		response.Username = *account.Username
	}

	return response
}

func firstCharacter(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	for _, char := range trimmed {
		return string(char)
	}

	return ""
}

func stringPtr(value string) *string {
	return &value
}
