package admin

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"absensi-cn-api/internal/modules/academic"
	studentModule "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/password"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// ─── DTOs ────────────────────────────────────────────────────────────────────

type ImportError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ImportResult struct {
	Imported int           `json:"imported"`
	Skipped  int           `json:"skipped"`
	Errors   []ImportError `json:"errors"`
}

// ─── Style constants ─────────────────────────────────────────────────────────

const (
	colorHeaderBg      = "059669" // emerald-600
	colorHeaderReqBg   = "047857" // emerald-700
	colorHeaderText    = "FFFFFF"
	colorExampleBg     = "ECFDF5" // emerald-50
	colorExampleText   = "065F46" // emerald-900
	colorInstructBg    = "064E3B"
	colorInstructText  = "FFFFFF"
	colorSectionBg     = "D1FAE5" // emerald-100
	colorSectionText   = "064E3B"
	colorRefHeaderBg   = "6EE7B7" // emerald-300
	colorRefHeaderText = "064E3B"
)

// ─── Template generators ─────────────────────────────────────────────────────

func (s *Service) GenerateTeacherTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Data Guru"
	f.SetSheetName("Sheet1", sheet)

	styles, err := buildTemplateStyles(f)
	if err != nil {
		return nil, err
	}

	// Column definitions: header, required, width
	cols := []struct {
		Header   string
		Required bool
		Width    float64
	}{
		{"Nama Lengkap", true, 30},
		{"Username", true, 22},
		{"Password", true, 22},
		{"NIP", false, 22},
		{"NUPTK", false, 22},
		{"Jenis Kelamin", false, 18},
		{"No. Telepon", false, 20},
		{"Alamat", false, 35},
		{"Wali Kelas", false, 22}, // col I (index 8) — homeroom class assignment
	}

	writeTemplateHeaders(f, sheet, cols, styles)

	// Example row (row 2) — name is prefixed with "(Contoh)" so the importer
	// skips it by content regardless of its row position.
	examples := []interface{}{
		"(Contoh) Budi Santoso", "budi.santoso", "budi1234",
		"198501012010011001", "1234567890123456", "MALE",
		"08123456789", "Jl. Merdeka No. 1, Jakarta",
		"", // Wali Kelas — kosongkan jika tidak ada
	}
	writeExampleRow(f, sheet, 2, examples, styles.exampleCell)

	// Freeze header row
	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	// Data Kelas sheet + dropdown validation on Wali Kelas column (I)
	if s.db != nil {
		numClasses, _ := writeClassReferenceSheet(f, s.db, styles) // non-fatal
		if numClasses > 0 {
			dv := excelize.NewDataValidation(true)
			dv.Sqref = "I3:I10000"
			dv.Type = "list"
			dv.Formula1 = fmt.Sprintf("'Data Kelas'!$A$2:$A$%d", numClasses+1)
			errStyle := "warning"
			dv.ErrorStyle = &errStyle
			errTitle := "Kelas Tidak Dikenal"
			dv.ErrorTitle = &errTitle
			errMsg := "Nama kelas tidak ada di sheet 'Data Kelas'. Tambahkan baris di sheet tersebut (isi A–E), lalu ketik nama yang sama di sini."
			dv.Error = &errMsg
			_ = f.AddDataValidation(sheet, dv)
		}
	}

	// Petunjuk sheet
	writeTeacherInstructionSheet(f, styles)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("generate teacher template: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *Service) GenerateStudentTemplate() ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheet = "Data Siswa"
	f.SetSheetName("Sheet1", sheet)

	styles, err := buildTemplateStyles(f)
	if err != nil {
		return nil, err
	}

	cols := []struct {
		Header   string
		Required bool
		Width    float64
	}{
		{"Nama Lengkap", true, 30},
		{"NIS (10 digit)", true, 18},
		{"Password", true, 22},
		{"Kelas", false, 22},
		{"NISN", false, 18},
		{"Jenis Kelamin", false, 18},
		{"Tempat Lahir", false, 22},
		{"Tanggal Lahir (DD/MM/YYYY)", false, 28},
		{"No. Telepon", false, 20},
		{"Nama Orang Tua", false, 25},
		{"Telepon Orang Tua", false, 20},
		{"Angkatan (Tahun)", false, 20},
	}

	writeTemplateHeaders(f, sheet, cols, styles)

	// Example row (row 2) — name is prefixed with "(Contoh)" so the importer
	// skips it by content regardless of its row position.
	examples := []interface{}{
		"(Contoh) Andi Pratama", "2024001001", "andi1234",
		"X RPL 1", "0012345678", "MALE",
		"Jakarta", "15/03/2008", "08987654321",
		"Pak Pratama", "08111222333", "2024",
	}
	writeExampleRow(f, sheet, 2, examples, styles.exampleCell)

	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	// Data Kelas sheet + dropdown validation on the Kelas column (D)
	if s.db != nil {
		numClasses, _ := writeClassReferenceSheet(f, s.db, styles) // non-fatal if error
		if numClasses > 0 {
			dv := excelize.NewDataValidation(true)
			dv.Sqref = "D3:D10000"
			dv.Type = "list"
			dv.Formula1 = fmt.Sprintf("'Data Kelas'!$A$2:$A$%d", numClasses+1)
			errStyle := "warning"
			dv.ErrorStyle = &errStyle
			errTitle := "Kelas Tidak Dikenal"
			dv.ErrorTitle = &errTitle
			errMsg := "Nama kelas tidak ada di sheet 'Data Kelas'. Tambahkan baris di sheet tersebut (isi A–E), lalu ketik nama yang sama di sini."
			dv.Error = &errMsg
			_ = f.AddDataValidation(sheet, dv)
		}
	}

	writeStudentInstructionSheet(f, styles)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("generate student template: %w", err)
	}
	return buf.Bytes(), nil
}

// ─── Import teachers ─────────────────────────────────────────────────────────

func (s *Service) ImportTeachers(data []byte) (*ImportResult, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("buka file Excel gagal: pastikan file adalah format .xlsx yang valid")
	}
	defer f.Close()

	rows, err := f.GetRows("Data Guru")
	if err != nil {
		return nil, fmt.Errorf("sheet 'Data Guru' tidak ditemukan. Gunakan template resmi yang telah disediakan")
	}

	// Build class lookup from Data Kelas sheet, fall back to live DB scan.
	classMap := buildClassMapFromExcel(f)
	if len(classMap) == 0 {
		dbMap, err := buildClassLookupMap(s.db)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat data kelas: %w", err)
		}
		classMap = dbMap
	}
	resolvedCache := make(map[string]classInfo)

	result := &ImportResult{Errors: []ImportError{}}
	for i, row := range rows {
		if i == 0 {
			continue // selalu skip baris header (row 1)
		}
		excelRow := i + 1

		name := cellStr(row, 0)
		username := cellStr(row, 1)
		pwd := cellStr(row, 2)

		// Skip baris contoh berdasarkan konten — tidak bergantung pada posisi baris
		if isExampleRow(name) {
			continue
		}

		if name == "" && username == "" && pwd == "" {
			continue // skip baris kosong
		}

		rowErrs := validateTeacherRow(name, username, pwd, excelRow)
		if len(rowErrs) > 0 {
			result.Errors = append(result.Errors, rowErrs...)
			result.Skipped++
			continue
		}

		// Check username duplicate
		var count int64
		if err := s.db.Model(&user.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "username", Message: "gagal memeriksa duplikat username"})
			result.Skipped++
			continue
		}
		if count > 0 {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "username", Message: fmt.Sprintf("username '%s' sudah digunakan", username)})
			result.Skipped++
			continue
		}

		nip := cellStr(row, 3)
		nuptk := cellStr(row, 4)
		gender := strings.ToUpper(cellStr(row, 5))
		phone := cellStr(row, 6)
		address := cellStr(row, 7)
		waliKelas := cellStr(row, 8) // col I — homeroom class (optional)

		// Resolve Wali Kelas class BEFORE the main transaction
		var waliClassID, waliSchoolYearID string
		if waliKelas != "" {
			normalized := normalizeClassName(waliKelas)
			info, found := resolvedCache[normalized]
			if !found {
				info, found = classMap[normalized]
			}
			if !found {
				result.Errors = append(result.Errors, ImportError{
					Row:     excelRow,
					Field:   "Wali Kelas",
					Message: fmt.Sprintf("kelas '%s' tidak ditemukan. Buka sheet 'Data Kelas', tambahkan baris baru (A: nama kelas, B: tingkat, C: kode jurusan, D: rombel, E: tahun ajaran), lalu upload ulang.", waliKelas),
				})
				result.Skipped++
				continue
			}
			if info.classID != "" {
				waliClassID = info.classID
				waliSchoolYearID = info.schoolYearID
			} else {
				cid, syid, resolveErr := lookupOrCreateClass(s.db, info.grade, info.majorCode, info.className, info.schoolYear)
				if resolveErr != nil {
					result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "Wali Kelas", Message: resolveErr.Error()})
					result.Skipped++
					continue
				}
				waliClassID = cid
				waliSchoolYearID = syid
				resolvedCache[normalized] = classInfo{classID: cid, schoolYearID: syid}
			}
		}

		hashedPwd, err := password.Hash(pwd)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "password", Message: "gagal memproses password"})
			result.Skipped++
			continue
		}

		var homeroomConflict bool
		txErr := s.db.Transaction(func(tx *gorm.DB) error {
			account := user.User{
				ID:           uuid.NewString(),
				Name:         name,
				Username:     &username,
				PasswordHash: hashedPwd,
				Role:         user.RoleTeacher,
			}
			if err := tx.Create(&account).Error; err != nil {
				return err
			}
			teacher := academic.Teacher{
				ID:       uuid.NewString(),
				UserID:   account.ID,
				NIP:      optionalString(nip),
				NUPTK:    optionalString(nuptk),
				Gender:   optionalUpperString(gender),
				Phone:    optionalString(phone),
				Address:  optionalString(address),
				IsActive: true,
			}
			if err := tx.Create(&teacher).Error; err != nil {
				return err
			}

			if waliClassID != "" {
				// Guard against duplicate homeroom (one per class per school year)
				var existing int64
				if err := tx.Model(&academic.HomeroomAssignment{}).
					Where("class_id = ? AND school_year_id = ?", waliClassID, waliSchoolYearID).
					Count(&existing).Error; err != nil {
					return err
				}
				if existing > 0 {
					homeroomConflict = true
					return nil // teacher created successfully; homeroom skipped
				}
				homeroom := academic.HomeroomAssignment{
					ID:           uuid.NewString(),
					TeacherID:    teacher.ID,
					ClassID:      waliClassID,
					SchoolYearID: waliSchoolYearID,
					IsActive:     true,
				}
				if err := tx.Create(&homeroom).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if txErr != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "-", Message: fmt.Sprintf("gagal menyimpan data: %s", txErr.Error())})
			result.Skipped++
			continue
		}

		if homeroomConflict {
			result.Errors = append(result.Errors, ImportError{
				Row:     excelRow,
				Field:   "Wali Kelas",
				Message: fmt.Sprintf("guru berhasil diimport, tapi kelas '%s' sudah memiliki wali kelas lain (penugasan wali kelas dilewati)", waliKelas),
			})
		}

		result.Imported++
	}

	return result, nil
}

// ─── Import students ─────────────────────────────────────────────────────────

func (s *Service) ImportStudents(data []byte) (*ImportResult, error) {
	if s.db == nil {
		return nil, ErrAdminDataUnavailable
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("buka file Excel gagal: pastikan file adalah format .xlsx yang valid")
	}
	defer f.Close()

	rows, err := f.GetRows("Data Siswa")
	if err != nil {
		return nil, fmt.Errorf("sheet 'Data Siswa' tidak ditemukan. Gunakan template resmi yang telah disediakan")
	}

	// Build class lookup from the uploaded file's Data Kelas sheet.
	// Fall back to a live DB query when the sheet is missing entirely.
	classMap := buildClassMapFromExcel(f)
	if len(classMap) == 0 {
		dbMap, err := buildClassLookupMap(s.db)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat data kelas: %w", err)
		}
		classMap = dbMap
	}
	// Cache for classes that were auto-created during this import run so we
	// don't try to create the same class twice for multiple students.
	resolvedCache := make(map[string]classInfo)

	result := &ImportResult{Errors: []ImportError{}}

	for i, row := range rows {
		if i == 0 {
			continue // selalu skip baris header (row 1)
		}
		excelRow := i + 1

		name := cellStr(row, 0)
		nis := cellStr(row, 1)
		pwd := cellStr(row, 2)

		// Skip baris contoh berdasarkan konten
		if isExampleRow(name) {
			continue
		}

		if name == "" && nis == "" && pwd == "" {
			continue
		}

		rowErrs := validateStudentRow(name, nis, pwd, excelRow)
		if len(rowErrs) > 0 {
			result.Errors = append(result.Errors, rowErrs...)
			result.Skipped++
			continue
		}

		// Check NIS duplicate
		var nisCount int64
		if err := s.db.Model(&studentModule.Student{}).Where("nis = ?", nis).Count(&nisCount).Error; err != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "NIS", Message: "gagal memeriksa duplikat NIS"})
			result.Skipped++
			continue
		}
		if nisCount > 0 {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "NIS", Message: fmt.Sprintf("NIS '%s' sudah terdaftar", nis)})
			result.Skipped++
			continue
		}

		kelasName := cellStr(row, 3)
		nisn := cellStr(row, 4)
		gender := strings.ToUpper(cellStr(row, 5))
		birthPlace := cellStr(row, 6)
		birthDateStr := cellStr(row, 7)
		phone := cellStr(row, 8)
		parentName := cellStr(row, 9)
		parentPhone := cellStr(row, 10)
		entryYearStr := cellStr(row, 11)

		// Validate optional fields
		var birthDate *time.Time
		if birthDateStr != "" {
			parsed, parseErr := parseFlexibleDate(birthDateStr)
			if parseErr != nil {
				result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "Tanggal Lahir", Message: "format tanggal tidak dikenali. Gunakan DD/MM/YYYY (contoh: 15/03/2008)"})
				result.Skipped++
				continue
			}
			birthDate = parsed
		}

		entryYear := 0
		if entryYearStr != "" {
			var yr int
			if _, scanErr := fmt.Sscanf(entryYearStr, "%d", &yr); scanErr != nil || yr < 2000 || yr > 2100 {
				result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "Angkatan", Message: "angkatan harus berupa tahun (contoh: 2024)"})
				result.Skipped++
				continue
			}
			entryYear = yr
		}

		// Resolve class
		var classID, schoolYearID string
		if kelasName != "" {
			normalized := normalizeClassName(kelasName)

			// Check resolved-this-run cache first (avoids duplicate creates)
			info, found := resolvedCache[normalized]
			if !found {
				info, found = classMap[normalized]
			}

			if !found {
				result.Errors = append(result.Errors, ImportError{
					Row:   excelRow,
					Field: "Kelas",
					Message: fmt.Sprintf(
						"kelas '%s' tidak ditemukan. Buka sheet 'Data Kelas', tambahkan baris baru (A: nama kelas, B: tingkat, C: kode jurusan, D: rombel, E: tahun ajaran), lalu upload ulang.",
						kelasName,
					),
				})
				result.Skipped++
				continue
			}

			if info.classID != "" {
				// IDs already embedded — use directly
				classID = info.classID
				schoolYearID = info.schoolYearID
			} else {
				// No IDs: look up or auto-create using the form fields
				cid, syid, resolveErr := lookupOrCreateClass(s.db, info.grade, info.majorCode, info.className, info.schoolYear)
				if resolveErr != nil {
					result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "Kelas", Message: resolveErr.Error()})
					result.Skipped++
					continue
				}
				classID = cid
				schoolYearID = syid
				resolvedCache[normalized] = classInfo{classID: cid, schoolYearID: syid}
			}
		}

		hashedPwd, err := password.Hash(pwd)
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "Password", Message: "gagal memproses password"})
			result.Skipped++
			continue
		}

		txErr := s.db.Transaction(func(tx *gorm.DB) error {
			account := user.User{
				ID:           uuid.NewString(),
				Name:         name,
				NIS:          &nis,
				PasswordHash: hashedPwd,
				Role:         user.RoleStudent,
			}
			if err := tx.Create(&account).Error; err != nil {
				return err
			}

			student := studentModule.Student{
				ID:          uuid.NewString(),
				UserID:      account.ID,
				NIS:         nis,
				NISN:        optionalString(nisn),
				Gender:      optionalUpperString(gender),
				BirthPlace:  optionalString(birthPlace),
				BirthDate:   birthDate,
				Phone:       optionalString(phone),
				ParentName:  optionalString(parentName),
				ParentPhone: optionalString(parentPhone),
				EntryYear:   entryYear,
				IsActive:    true,
			}
			if err := tx.Create(&student).Error; err != nil {
				return err
			}

			if classID != "" && schoolYearID != "" {
				membership := studentModule.StudentClassMembership{
					ID:           uuid.NewString(),
					StudentID:    student.ID,
					ClassID:      classID,
					SchoolYearID: schoolYearID,
					Status:       studentModule.MembershipStatusActive,
					IsActive:     true,
				}
				if err := tx.Create(&membership).Error; err != nil {
					return err
				}
			}

			return nil
		})
		if txErr != nil {
			result.Errors = append(result.Errors, ImportError{Row: excelRow, Field: "-", Message: fmt.Sprintf("gagal menyimpan data: %s", txErr.Error())})
			result.Skipped++
			continue
		}

		result.Imported++
	}

	return result, nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// normalizeClassName uppercases and collapses extra whitespace so minor
// spacing/casing differences between template and user input don't break lookup.
func normalizeClassName(name string) string {
	parts := strings.Fields(strings.ToUpper(strings.TrimSpace(name)))
	return strings.Join(parts, " ")
}

// parseFlexibleDate handles dates regardless of how Excel serializes them:
// user-typed DD/MM/YYYY, Excel-reformatted ISO, US format, or serial number.
func parseFlexibleDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	formats := []string{
		"02/01/2006",          // DD/MM/YYYY  ← target
		"2/1/2006",            // D/M/YYYY
		"2006-01-02",          // YYYY-MM-DD  (ISO — excelize raw)
		"2006-01-02 00:00:00", // ISO with time
		"01/02/2006",          // MM/DD/YYYY  (US Excel default)
		"1/2/2006",            // M/D/YYYY
		"02-01-2006",          // DD-MM-YYYY
		"2006/01/02",          // YYYY/MM/DD
	}
	for _, f := range formats {
		if t, err := time.Parse(f, value); err == nil {
			return &t, nil
		}
	}

	// Excel date serial number (e.g. "39462")
	if serial, err := strconv.ParseFloat(value, 64); err == nil && serial > 1 {
		if t, err := excelize.ExcelDateToTime(serial, false); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("format tanggal tidak dikenali")
}

// isExampleRow returns true when the name cell carries the "(Contoh)" prefix
// that is embedded in every template example row. This lets the importer skip
// the example regardless of its physical row position — so it stays safe even
// when users insert/delete rows above their data.
func isExampleRow(name string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(name)), "(CONTOH)")
}

func cellStr(row []string, idx int) string {
	if idx >= len(row) {
		return ""
	}
	return strings.TrimSpace(row[idx])
}

type classInfo struct {
	classID      string
	schoolYearID string
	// Populated from Data Kelas form columns when IDs are absent (new/user-added rows).
	// Used by lookupOrCreateClass to find or create the class in the DB.
	grade      string
	majorCode  string
	className  string
	schoolYear string
}

func buildClassLookupMap(db *gorm.DB) (map[string]classInfo, error) {
	type classLookupRow struct {
		ID           string
		SchoolYearID string
		DisplayName  string
	}
	var rows []classLookupRow
	if err := db.Table("classes").
		Select("classes.id, classes.school_year_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as display_name").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = classes.school_year_id").
		Where("classes.is_active = ?", true).
		Order("school_years.start_year desc").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	m := make(map[string]classInfo, len(rows))
	for _, row := range rows {
		key := normalizeClassName(row.DisplayName)
		// Only add if not already present (first = most recent school year wins due to ORDER BY)
		if _, exists := m[key]; !exists {
			m[key] = classInfo{classID: row.ID, schoolYearID: row.SchoolYearID}
		}
	}
	return m, nil
}

// lookupOrCreateClass resolves a classInfo that has form fields but no embedded IDs.
// It finds the major by code, the school year by name, then finds or creates the class.
// Returns the resolved classID and schoolYearID.
func lookupOrCreateClass(db *gorm.DB, grade, majorCode, className, schoolYearName string) (classID, schoolYearID string, err error) {
	var major academic.Major
	if err = db.Where("UPPER(code) = UPPER(?) AND is_active = ?", majorCode, true).First(&major).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("jurusan dengan kode '%s' tidak ditemukan di sistem", majorCode)
		}
		return "", "", fmt.Errorf("gagal mencari jurusan: %w", err)
	}

	var sy academic.SchoolYear
	if err = db.Where("name = ? AND is_active = ?", schoolYearName, true).First(&sy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("tahun ajaran '%s' tidak ditemukan di sistem", schoolYearName)
		}
		return "", "", fmt.Errorf("gagal mencari tahun ajaran: %w", err)
	}

	var cls academic.SchoolClass
	err = db.Where("grade = ? AND major_id = ? AND name = ? AND school_year_id = ?",
		grade, major.ID, className, sy.ID).First(&cls).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("gagal mencari kelas: %w", err)
		}
		// Class does not exist yet — create it
		cls = academic.SchoolClass{
			ID:           uuid.NewString(),
			Grade:        strings.ToUpper(strings.TrimSpace(grade)),
			MajorID:      major.ID,
			Name:         strings.TrimSpace(className),
			SchoolYearID: sy.ID,
			IsActive:     true,
		}
		if err = db.Create(&cls).Error; err != nil {
			return "", "", fmt.Errorf("gagal membuat kelas baru '%s %s %s': %w", grade, majorCode, className, err)
		}
	}

	return cls.ID, sy.ID, nil
}

func validateTeacherRow(name, username, pwd string, row int) []ImportError {
	var errs []ImportError
	if name == "" {
		errs = append(errs, ImportError{Row: row, Field: "Nama Lengkap", Message: "nama lengkap wajib diisi"})
	}
	if username == "" {
		errs = append(errs, ImportError{Row: row, Field: "Username", Message: "username wajib diisi"})
	}
	if len(pwd) < 6 {
		errs = append(errs, ImportError{Row: row, Field: "Password", Message: "password minimal 6 karakter"})
	}
	return errs
}

func validateStudentRow(name, nis, pwd string, row int) []ImportError {
	var errs []ImportError
	if name == "" {
		errs = append(errs, ImportError{Row: row, Field: "Nama Lengkap", Message: "nama lengkap wajib diisi"})
	}
	if len(nis) != 10 {
		errs = append(errs, ImportError{Row: row, Field: "NIS", Message: "NIS harus tepat 10 digit angka"})
	} else {
		for _, c := range nis {
			if c < '0' || c > '9' {
				errs = append(errs, ImportError{Row: row, Field: "NIS", Message: "NIS hanya boleh berisi angka"})
				break
			}
		}
	}
	if len(pwd) < 6 {
		errs = append(errs, ImportError{Row: row, Field: "Password", Message: "password minimal 6 karakter"})
	}
	return errs
}

// ─── Template styling helpers ─────────────────────────────────────────────────

type templateStyles struct {
	headerRequired int
	headerOptional int
	exampleCell    int
	labelCell      int
	instructTitle  int
	sectionHeader  int
	bodyCell       int
	refHeader      int
	refCell        int
}

func buildTemplateStyles(f *excelize.File) (*templateStyles, error) {
	makeStyle := func(s *excelize.Style) (int, error) {
		return f.NewStyle(s)
	}

	headerRequired, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorHeaderReqBg}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: colorHeaderText, Size: 10.5, Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 2},
			{Type: "right", Color: "FFFFFF", Style: 2},
			{Type: "bottom", Color: "FFFFFF", Style: 2},
		},
	})
	if err != nil {
		return nil, err
	}

	headerOptional, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorHeaderBg}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: colorHeaderText, Size: 10.5, Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "FFFFFF", Style: 1},
			{Type: "right", Color: "FFFFFF", Style: 1},
			{Type: "bottom", Color: "FFFFFF", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	exampleCell, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorExampleBg}, Pattern: 1},
		Font: &excelize.Font{Color: colorExampleText, Size: 10, Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "A7F3D0", Style: 1},
			{Type: "right", Color: "A7F3D0", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	labelCell, err := makeStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "374151", Size: 10, Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	instructTitle, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorInstructBg}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: colorInstructText, Size: 14, Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	sectionHeader, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorSectionBg}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: colorSectionText, Size: 11, Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	bodyCell, err := makeStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "374151", Size: 10, Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "bottom", Color: "E5E7EB", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	refHeader, err := makeStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{colorRefHeaderBg}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Color: colorRefHeaderText, Size: 10.5, Family: "Calibri"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	refCell, err := makeStyle(&excelize.Style{
		Font:      &excelize.Font{Color: "065F46", Size: 10, Family: "Calibri"},
		Alignment: &excelize.Alignment{Vertical: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "D1FAE5", Style: 1},
			{Type: "right", Color: "D1FAE5", Style: 1},
		},
	})
	if err != nil {
		return nil, err
	}

	return &templateStyles{
		headerRequired: headerRequired,
		headerOptional: headerOptional,
		exampleCell:    exampleCell,
		labelCell:      labelCell,
		instructTitle:  instructTitle,
		sectionHeader:  sectionHeader,
		bodyCell:       bodyCell,
		refHeader:      refHeader,
		refCell:        refCell,
	}, nil
}

func writeTemplateHeaders(f *excelize.File, sheet string, cols []struct {
	Header   string
	Required bool
	Width    float64
}, styles *templateStyles) {
	f.SetRowHeight(sheet, 1, 36)

	for i, col := range cols {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s1", colLetter)

		label := col.Header
		if col.Required {
			label = "★ " + col.Header
		}
		f.SetCellValue(sheet, cell, label)

		styleID := styles.headerOptional
		if col.Required {
			styleID = styles.headerRequired
		}
		f.SetCellStyle(sheet, cell, cell, styleID)
		f.SetColWidth(sheet, colLetter, colLetter, col.Width)
	}
}

func writeExampleRow(f *excelize.File, sheet string, rowNum int, values []interface{}, styleID int) {
	f.SetRowHeight(sheet, rowNum, 22)
	for i, val := range values {
		colLetter, _ := excelize.ColumnNumberToName(i + 1)
		cell := fmt.Sprintf("%s%d", colLetter, rowNum)
		f.SetCellValue(sheet, cell, val)
		f.SetCellStyle(sheet, cell, cell, styleID)
	}
}

func writeTeacherInstructionSheet(f *excelize.File, styles *templateStyles) {
	const sheet = "Petunjuk"
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "A", 28)
	f.SetColWidth(sheet, "B", "B", 55)

	rows := []struct {
		rowType string // "title", "section", "label", "body", "empty"
		colA    string
		colB    string
		height  float64
	}{
		{"title", "PANDUAN IMPORT DATA GURU", "", 40},
		{"empty", "", "", 8},
		{"section", "CARA PENGGUNAAN", "", 26},
		{"body", "1.", "Isi sheet 'Data Guru' mulai dari baris ke-3 (baris 1 = header, baris 2 = contoh)", 40},
		{"body", "2.", "Kolom bertanda ★ bersifat WAJIB diisi", 24},
		{"body", "3.", "Kolom Wali Kelas: klik sel untuk memilih dari dropdown, atau ketik nama kelas (misal: XI PPLG 2)", 48},
		{"body", "4.", "Wali Kelas BARU: tambah baris di sheet 'Data Kelas' (isi A–E), lalu tulis nama yang sama di kolom Wali Kelas", 56},
		{"body", "5.", "Simpan file dalam format .xlsx lalu upload", 24},
		{"empty", "", "", 8},
		{"section", "SHEET DATA KELAS — CARA TAMBAH KELAS BARU", "", 26},
		{"body", "A", "Nama Kelas — isi sama persis dengan kolom Wali Kelas di Data Guru (misal: XI PPLG 2)", 40},
		{"body", "B", "Tingkat — X, XI, atau XII", 24},
		{"body", "C", "Kode Jurusan — sesuai kode jurusan di sistem (misal: RPL, TKJ, PPLG)", 36},
		{"body", "D", "Rombel — nomor rombel kelas (misal: 1, 2, 3, 4, 5)", 32},
		{"body", "E", "Tahun Ajaran — sesuai nama tahun ajaran di sistem (misal: 2024/2025)", 36},
		{"body", "⚠", "Kelas baru OTOMATIS dibuat saat import; guru langsung jadi wali kelas kelas tersebut", 40},
		{"empty", "", "", 8},
		{"section", "KETERANGAN KOLOM DATA GURU", "", 26},
		{"label", "★ Nama Lengkap", "Nama lengkap guru. Contoh: Budi Santoso", 0},
		{"label", "★ Username", "Username untuk login guru. Contoh: budi.santoso (harus unik)", 0},
		{"label", "★ Password", "Password login, minimal 6 karakter", 0},
		{"label", "NIP", "Nomor Induk Pegawai (opsional)", 0},
		{"label", "NUPTK", "Nomor Unik Pendidik dan Tenaga Kependidikan (opsional)", 0},
		{"label", "Jenis Kelamin", "Isi dengan: MALE atau FEMALE (opsional)", 0},
		{"label", "No. Telepon", "Nomor telepon guru (opsional)", 0},
		{"label", "Alamat", "Alamat lengkap guru (opsional)", 0},
		{"label", "Wali Kelas", "Nama kelas dari sheet Data Kelas; guru otomatis jadi wali kelas (opsional)", 0},
		{"empty", "", "", 8},
		{"section", "CATATAN PENTING", "", 26},
		{"body", "•", "Baris duplikat (username sudah ada) akan dilewati dan dilaporkan sebagai error", 24},
		{"body", "•", "Jika kelas sudah punya wali kelas lain, penugasan dilewati tapi guru tetap diimport", 36},
		{"body", "•", "Pastikan username belum pernah digunakan sebelumnya di sistem", 24},
	}

	for rowIdx, r := range rows {
		excelRow := rowIdx + 1
		switch r.rowType {
		case "title":
			f.MergeCell(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.instructTitle)
			f.SetRowHeight(sheet, excelRow, r.height)
		case "section":
			f.MergeCell(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.sectionHeader)
			f.SetRowHeight(sheet, excelRow, r.height)
		case "label":
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), r.colB)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), styles.labelCell)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.bodyCell)
			f.SetRowHeight(sheet, excelRow, 22)
		case "body":
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), r.colB)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), styles.bodyCell)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.bodyCell)
			if r.height > 0 {
				f.SetRowHeight(sheet, excelRow, r.height)
			} else {
				f.SetRowHeight(sheet, excelRow, 22)
			}
		case "empty":
			f.SetRowHeight(sheet, excelRow, r.height)
		}
	}
}

func writeStudentInstructionSheet(f *excelize.File, styles *templateStyles) {
	const sheet = "Petunjuk"
	f.NewSheet(sheet)
	f.SetColWidth(sheet, "A", "A", 32)
	f.SetColWidth(sheet, "B", "B", 60)

	rows := []struct {
		rowType string
		colA    string
		colB    string
		height  float64
	}{
		{"title", "PANDUAN IMPORT DATA SISWA", "", 40},
		{"empty", "", "", 8},
		{"section", "CARA PENGGUNAAN", "", 26},
		{"body", "1.", "Isi sheet 'Data Siswa' mulai dari baris ke-3 (baris 1 = header, baris 2 = contoh)", 40},
		{"body", "2.", "Kolom bertanda ★ bersifat WAJIB diisi", 24},
		{"body", "3.", "Kolom Kelas: klik sel untuk memilih dari dropdown, atau ketik nama kelas (misal: XI PPLG 2)", 40},
		{"body", "4.", "Kelas BARU: tambah baris di sheet 'Data Kelas' (isi A Nama Kelas, B Tingkat, C Kode Jurusan, D Rombel, E Tahun Ajaran), lalu tulis nama yang sama di kolom Kelas", 56},
		{"body", "5.", "Simpan file dalam format .xlsx lalu upload", 24},
		{"empty", "", "", 8},
		{"section", "SHEET DATA KELAS — CARA TAMBAH KELAS BARU", "", 26},
		{"body", "A", "Nama Kelas — isi sama persis dengan kolom Kelas di Data Siswa (misal: XI PPLG 2)", 40},
		{"body", "B", "Tingkat — X, XI, atau XII", 24},
		{"body", "C", "Kode Jurusan — sesuai kode jurusan di sistem (misal: RPL, TKJ, PPLG)", 36},
		{"body", "D", "Rombel — nomor rombel kelas (misal: 1, 2, 3, 4, 5)", 32},
		{"body", "E", "Tahun Ajaran — sesuai nama tahun ajaran di sistem (misal: 2024/2025)", 36},
		{"body", "⚠", "Kelas baru OTOMATIS dibuat saat import (tidak perlu buat manual di Manajemen Kelas)", 40},
		{"empty", "", "", 8},
		{"section", "KETERANGAN KOLOM DATA SISWA", "", 26},
		{"label", "★ Nama Lengkap", "Nama lengkap siswa", 0},
		{"label", "★ NIS (10 digit)", "Nomor Induk Siswa, tepat 10 digit angka. Contoh: 2024001001", 0},
		{"label", "★ Password", "Password login, minimal 6 karakter", 0},
		{"label", "Kelas", "Nama kelas dari Data Kelas. Kelas baru dibuat otomatis saat import (opsional)", 0},
		{"label", "NISN", "Nomor Induk Siswa Nasional (opsional)", 0},
		{"label", "Jenis Kelamin", "MALE atau FEMALE (opsional)", 0},
		{"label", "Tempat Lahir", "Kota tempat lahir (opsional)", 0},
		{"label", "Tanggal Lahir", "Format: DD/MM/YYYY. Contoh: 15/03/2008 (opsional)", 0},
		{"label", "No. Telepon", "Nomor telepon siswa (opsional)", 0},
		{"label", "Nama Orang Tua", "Nama orang tua/wali (opsional)", 0},
		{"label", "Telepon Orang Tua", "Nomor telepon orang tua (opsional)", 0},
		{"label", "Angkatan (Tahun)", "Tahun masuk siswa. Contoh: 2024 (opsional)", 0},
		{"empty", "", "", 8},
		{"section", "CATATAN PENTING", "", 26},
		{"body", "•", "NIS duplikat akan dilewati dan dilaporkan sebagai error", 24},
		{"body", "•", "Kelas yang tidak ditemukan di Data Kelas akan menyebabkan error pada baris tersebut", 40},
		{"body", "•", "Kelas baru (tanpa ID) dibuat otomatis — Kode Jurusan dan Tahun Ajaran harus valid di sistem", 48},
		{"body", "•", "Siswa tanpa kelas tetap dibuat, namun tidak di-assign ke kelas manapun", 32},
	}

	for rowIdx, r := range rows {
		excelRow := rowIdx + 1
		switch r.rowType {
		case "title":
			f.MergeCell(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.instructTitle)
			f.SetRowHeight(sheet, excelRow, r.height)
		case "section":
			f.MergeCell(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow))
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.sectionHeader)
			f.SetRowHeight(sheet, excelRow, r.height)
		case "label":
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), r.colB)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), styles.labelCell)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.bodyCell)
			f.SetRowHeight(sheet, excelRow, 22)
		case "body":
			f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), r.colA)
			f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), r.colB)
			f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("A%d", excelRow), styles.bodyCell)
			f.SetCellStyle(sheet, fmt.Sprintf("B%d", excelRow), fmt.Sprintf("B%d", excelRow), styles.bodyCell)
			if r.height > 0 {
				f.SetRowHeight(sheet, excelRow, r.height)
			} else {
				f.SetRowHeight(sheet, excelRow, 22)
			}
		case "empty":
			f.SetRowHeight(sheet, excelRow, r.height)
		}
	}
}

// writeClassReferenceSheet generates the "Data Kelas" sheet.
// Columns A–E are visible form-like fields; F and G store embedded IDs (hidden).
// Returns the number of class rows written so the caller can size dropdown ranges.
//
// Column layout:
//   A: Nama Kelas  (display name, used as dropdown key in Data Siswa)
//   B: Tingkat     (X / XI / XII)
//   C: Kode Jurusan
//   D: Rombel      (rombel number, e.g. "1", "2", "3")
//   E: Tahun Ajaran
//   F: _class_id        ← hidden, embedded for reliable lookup
//   G: _school_year_id  ← hidden, embedded for reliable lookup
//
// For existing classes all seven columns are filled.
// Users may add NEW rows by filling A–E only (leave F–G blank); the importer
// will look up or auto-create the class using those form fields.
func writeClassReferenceSheet(f *excelize.File, db *gorm.DB, styles *templateStyles) (int, error) {
	type refRow struct {
		ID             string
		SchoolYearID   string
		DisplayName    string
		Grade          string
		MajorCode      string
		ClassName      string
		SchoolYearName string
	}

	var rows []refRow
	if err := db.Table("classes").
		Select("classes.id, classes.school_year_id, concat(classes.grade, ' ', majors.code, ' ', classes.name) as display_name, classes.grade, majors.code as major_code, classes.name as class_name, school_years.name as school_year_name").
		Joins("join majors on majors.id = classes.major_id").
		Joins("join school_years on school_years.id = classes.school_year_id").
		Where("classes.is_active = ?", true).
		Order("school_years.start_year desc, classes.grade asc, majors.code asc, classes.name asc").
		Scan(&rows).Error; err != nil {
		return 0, err
	}

	const sheet = "Data Kelas"
	f.NewSheet(sheet)

	// Column widths
	f.SetColWidth(sheet, "A", "A", 22) // Nama Kelas
	f.SetColWidth(sheet, "B", "B", 10) // Tingkat
	f.SetColWidth(sheet, "C", "C", 16) // Kode Jurusan
	f.SetColWidth(sheet, "D", "D", 10) // Rombel
	f.SetColWidth(sheet, "E", "E", 20) // Tahun Ajaran
	// Hidden ID columns
	f.SetColWidth(sheet, "F", "F", 0.1)
	f.SetColWidth(sheet, "G", "G", 0.1)
	_ = f.SetColVisible(sheet, "F", false)
	_ = f.SetColVisible(sheet, "G", false)

	// Header row
	f.SetRowHeight(sheet, 1, 32)
	headers := []struct{ cell, value string }{
		{"A1", "Nama Kelas (isi di kolom Kelas Data Siswa)"},
		{"B1", "Tingkat"},
		{"C1", "Kode Jurusan"},
		{"D1", "Rombel"},
		{"E1", "Tahun Ajaran"},
		{"F1", "_class_id"},
		{"G1", "_school_year_id"},
	}
	for _, h := range headers {
		f.SetCellValue(sheet, h.cell, h.value)
	}
	f.SetCellStyle(sheet, "A1", "E1", styles.refHeader)

	for i, row := range rows {
		excelRow := i + 2
		f.SetRowHeight(sheet, excelRow, 20)
		f.SetCellValue(sheet, fmt.Sprintf("A%d", excelRow), row.DisplayName)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", excelRow), row.Grade)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", excelRow), row.MajorCode)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", excelRow), row.ClassName)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", excelRow), row.SchoolYearName)
		f.SetCellValue(sheet, fmt.Sprintf("F%d", excelRow), row.ID)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", excelRow), row.SchoolYearID)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", excelRow), fmt.Sprintf("E%d", excelRow), styles.refCell)
	}

	if len(rows) == 0 {
		f.SetCellValue(sheet, "A2", "Belum ada kelas. Tambahkan baris baru di sini: isi A (nama), B (tingkat: X/XI/XII), C (kode jurusan), D (rombel: 1/2/3), E (tahun ajaran).")
		f.SetRowHeight(sheet, 2, 28)
	}

	_ = f.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	return len(rows), nil
}

// buildClassMapFromExcel reads the "Data Kelas" sheet from the uploaded file.
//
// Column layout expected (matches writeClassReferenceSheet):
//   col 0 (A): Nama Kelas  — used as the lookup key
//   col 1 (B): Tingkat
//   col 2 (C): Kode Jurusan
//   col 3 (D): Rombel
//   col 4 (E): Tahun Ajaran
//   col 5 (F): _class_id        (empty for user-added new-class rows)
//   col 6 (G): _school_year_id  (empty for user-added new-class rows)
//
// Rows with IDs (cols F & G) are resolved directly.
// Rows without IDs but with form fields (cols B–E) are flagged for
// lookupOrCreateClass during import.
func buildClassMapFromExcel(f *excelize.File) map[string]classInfo {
	const sheet = "Data Kelas"
	rows, err := f.GetRows(sheet)
	if err != nil || len(rows) < 2 {
		return nil
	}

	m := make(map[string]classInfo)
	for _, row := range rows[1:] { // skip header
		displayName := cellStr(row, 0)
		if displayName == "" {
			continue
		}
		grade := cellStr(row, 1)
		majorCode := cellStr(row, 2)
		className := cellStr(row, 3)
		schoolYear := cellStr(row, 4)
		classID := cellStr(row, 5)
		schoolYearID := cellStr(row, 6)

		// Skip rows that have neither IDs nor usable form fields
		hasIDs := classID != "" && schoolYearID != ""
		hasFormFields := grade != "" && majorCode != "" && className != "" && schoolYear != ""
		if !hasIDs && !hasFormFields {
			continue
		}

		key := normalizeClassName(displayName)
		if _, exists := m[key]; !exists {
			m[key] = classInfo{
				classID:      classID,
				schoolYearID: schoolYearID,
				grade:        grade,
				majorCode:    majorCode,
				className:    className,
				schoolYear:   schoolYear,
			}
		}
	}
	return m
}
