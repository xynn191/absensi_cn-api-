package router

import (
	"absensi-cn-api/internal/config"
	"absensi-cn-api/internal/middleware"
	"absensi-cn-api/internal/modules/admin"
	"absensi-cn-api/internal/modules/attendance"
	"absensi-cn-api/internal/modules/auth"
	"absensi-cn-api/internal/modules/health"
	"absensi-cn-api/internal/modules/staff"
	studentPortal "absensi-cn-api/internal/modules/student"
	"absensi-cn-api/internal/modules/user"
	"absensi-cn-api/pkg/token"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(cfg.GinMode())

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
	}))
	engine.Static("/uploads", "./storage/uploads")

	jwtManager := token.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpiresInHours)

	healthHandler := health.NewHandler(db)
	authHandler := auth.NewHandler(auth.NewService(jwtManager, db))
	attendanceService := attendance.NewService(db)
	attendanceHandler := attendance.NewHandler(attendanceService)
	adminHandler := admin.NewHandler(admin.NewService(db, attendanceService))
	staffHandler := staff.NewHandler(staff.NewService(db, attendanceService))
	studentHandler := studentPortal.NewHandler(studentPortal.NewService(db, attendanceService))

	api := engine.Group(cfg.APIPrefix)
	{
		api.GET("/health", healthHandler.Check)
		api.POST("/auth/login", authHandler.Login)

		attendanceGroup := api.Group("/attendance")
		attendanceGroup.Use(middleware.RequireAuth(jwtManager), middleware.RequireRoles(string(user.RoleStudent)))
		{
			attendanceGroup.GET("", attendanceHandler.History)
			attendanceGroup.GET("/today", attendanceHandler.Today)
			attendanceGroup.POST("/check-in", attendanceHandler.CheckIn)
			attendanceGroup.POST("/check-out", attendanceHandler.CheckOut)
		}

		studentGroup := api.Group("/student")
		studentGroup.Use(middleware.RequireAuth(jwtManager), middleware.RequireRoles(string(user.RoleStudent)))
		{
			studentGroup.GET("/dashboard", studentHandler.Dashboard)
			studentGroup.GET("/profile", studentHandler.Profile)
			studentGroup.GET("/me", studentHandler.Profile)
			studentGroup.GET("/today", studentHandler.Today)
			studentGroup.GET("/history", studentHandler.History)
			studentGroup.POST("/daily-report", studentHandler.SubmitDailyReport)
			studentGroup.GET("/attendance/today", studentHandler.Today)
			studentGroup.GET("/attendance/history", studentHandler.AttendanceHistory)
			studentGroup.POST("/attendance/check-in", attendanceHandler.CheckIn)
			studentGroup.GET("/submissions", studentHandler.Submissions)
			studentGroup.POST("/submissions", staffHandler.StudentCreateSubmission)
		}

		teacherGroup := api.Group("/teacher")
		teacherGroup.Use(middleware.RequireAuth(jwtManager), middleware.RequireRoles(string(user.RoleTeacher)))
		{
			teacherGroup.GET("/me", staffHandler.TeacherMe)
			teacherGroup.GET("/subject-assignments", staffHandler.TeacherSubjectAssignments)
			teacherGroup.GET("/homeroom", staffHandler.TeacherHomeroom)
			teacherGroup.GET("/homeroom/dashboard", staffHandler.TeacherHomeroomDashboard)
			teacherGroup.GET("/homeroom/students", staffHandler.TeacherHomeroomStudents)
			teacherGroup.GET("/homeroom/students/:student_id", staffHandler.TeacherHomeroomStudentDetail)
			teacherGroup.GET("/homeroom/attendance", staffHandler.TeacherHomeroomAttendance)
			teacherGroup.GET("/homeroom/attendance-overview", staffHandler.TeacherHomeroomAttendanceOverview)
			teacherGroup.GET("/homeroom/attendance-summary", staffHandler.TeacherHomeroomAttendanceSummary)
			teacherGroup.PATCH("/homeroom/attendance/:id/review", staffHandler.TeacherHomeroomAttendanceReview)
			teacherGroup.GET("/homeroom/students/:student_id/attendance-history", staffHandler.TeacherStudentAttendanceHistory)
			teacherGroup.GET("/homeroom/submissions", staffHandler.TeacherHomeroomSubmissions)
			teacherGroup.GET("/homeroom/submissions-overview", staffHandler.TeacherHomeroomSubmissionsOverview)
			teacherGroup.PATCH("/homeroom/submissions/:id/review", staffHandler.TeacherHomeroomSubmissionReview)
		}

		bkGroup := api.Group("/bk")
		bkGroup.Use(middleware.RequireAuth(jwtManager), middleware.RequireRoles(string(user.RoleBK)))
		{
			bkGroup.GET("/dashboard", staffHandler.BKDashboard)
			bkGroup.GET("/students", staffHandler.BKStudents)
			bkGroup.GET("/students-overview", staffHandler.BKStudentsOverview)
			bkGroup.GET("/students/:student_id", staffHandler.BKStudent)
			bkGroup.GET("/attendance", staffHandler.BKAttendance)
			bkGroup.GET("/attendance-overview", staffHandler.BKAttendanceOverview)
			bkGroup.GET("/attendance/summary", staffHandler.BKAttendanceSummary)
			bkGroup.PATCH("/attendance/:id/review", staffHandler.BKAttendanceReview)
			bkGroup.GET("/students/:student_id/attendance-history", staffHandler.BKStudentAttendanceHistory)
			bkGroup.GET("/students/:student_id/counseling-notes", staffHandler.BKStudentCounselingNotes)
			bkGroup.GET("/counseling-overview", staffHandler.BKCounselingOverview)
			bkGroup.POST("/students/:student_id/counseling-notes", staffHandler.BKCreateCounselingNote)
			bkGroup.PATCH("/counseling-notes/:id", staffHandler.BKUpdateCounselingNote)
			bkGroup.DELETE("/counseling-notes/:id", staffHandler.BKDeleteCounselingNote)
			bkGroup.GET("/submissions", staffHandler.BKSubmissions)
			bkGroup.GET("/submissions-overview", staffHandler.BKSubmissionsOverview)
			bkGroup.PATCH("/submissions/:id/review", staffHandler.BKSubmissionReview)
		}

		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.RequireAuth(jwtManager), middleware.RequireRoles(string(user.RoleAdmin)))
		{
			adminGroup.GET("/dashboard", adminHandler.Dashboard)
			adminGroup.GET("/students", adminHandler.ListStudents)
			adminGroup.GET("/students/:id", adminHandler.GetStudent)
			adminGroup.POST("/students", adminHandler.CreateStudent)
			adminGroup.PATCH("/students/:id", adminHandler.UpdateStudent)
			adminGroup.DELETE("/students/:id", adminHandler.DeleteStudent)
			adminGroup.GET("/student-class-memberships", adminHandler.ListStudentClassMemberships)
			adminGroup.GET("/student-class-memberships/:id", adminHandler.GetStudentClassMembership)
			adminGroup.POST("/student-class-memberships", adminHandler.CreateStudentClassMembership)
			adminGroup.PATCH("/student-class-memberships/:id", adminHandler.UpdateStudentClassMembership)
			adminGroup.DELETE("/student-class-memberships/:id", adminHandler.DeleteStudentClassMembership)
			adminGroup.GET("/attendance-rules", adminHandler.ListAttendanceRules)
			adminGroup.GET("/attendance-rules/:id", adminHandler.GetAttendanceRule)
			adminGroup.POST("/attendance-rules", adminHandler.CreateAttendanceRule)
			adminGroup.PATCH("/attendance-rules/:id", adminHandler.UpdateAttendanceRule)
			adminGroup.DELETE("/attendance-rules/:id", adminHandler.DeleteAttendanceRule)
			adminGroup.GET("/attendance", adminHandler.ListAttendanceRecords)
			adminGroup.PATCH("/attendance/:id/review", adminHandler.ReviewAttendance)
			adminGroup.GET("/teachers", adminHandler.ListTeachers)
			adminGroup.GET("/teacher-profiles", adminHandler.ListTeacherProfiles)
			adminGroup.POST("/teacher-profiles", adminHandler.CreateTeacherProfile)
			adminGroup.PATCH("/teacher-profiles/:id", adminHandler.UpdateTeacherProfile)
			adminGroup.DELETE("/teacher-profiles/:id", adminHandler.DeleteTeacherProfile)
			adminGroup.GET("/subjects", adminHandler.ListSubjects)
			adminGroup.POST("/subjects", adminHandler.CreateSubject)
			adminGroup.PATCH("/subjects/:id", adminHandler.UpdateSubject)
			adminGroup.DELETE("/subjects/:id", adminHandler.DeleteSubject)
			adminGroup.GET("/majors", adminHandler.ListMajors)
			adminGroup.POST("/majors", adminHandler.CreateMajor)
			adminGroup.PATCH("/majors/:id", adminHandler.UpdateMajor)
			adminGroup.DELETE("/majors/:id", adminHandler.DeleteMajor)
			adminGroup.GET("/school-years", adminHandler.ListSchoolYears)
			adminGroup.POST("/school-years", adminHandler.CreateSchoolYear)
			adminGroup.PATCH("/school-years/:id", adminHandler.UpdateSchoolYear)
			adminGroup.DELETE("/school-years/:id", adminHandler.DeleteSchoolYear)
			adminGroup.GET("/classes", adminHandler.ListClasses)
			adminGroup.POST("/classes", adminHandler.CreateClass)
			adminGroup.PATCH("/classes/:id", adminHandler.UpdateClass)
			adminGroup.DELETE("/classes/:id", adminHandler.DeleteClass)
			adminGroup.GET("/teacher-subject-assignments", adminHandler.ListTeacherSubjectAssignments)
			adminGroup.POST("/teacher-subject-assignments", adminHandler.CreateTeacherSubjectAssignment)
			adminGroup.PATCH("/teacher-subject-assignments/:id", adminHandler.UpdateTeacherSubjectAssignment)
			adminGroup.DELETE("/teacher-subject-assignments/:id", adminHandler.DeleteTeacherSubjectAssignment)
			adminGroup.GET("/homeroom-assignments", adminHandler.ListHomeroomAssignments)
			adminGroup.POST("/homeroom-assignments", adminHandler.CreateHomeroomAssignment)
			adminGroup.PATCH("/homeroom-assignments/:id", adminHandler.UpdateHomeroomAssignment)
			adminGroup.DELETE("/homeroom-assignments/:id", adminHandler.DeleteHomeroomAssignment)
			adminGroup.GET("/users", adminHandler.ListUsers)
			adminGroup.GET("/users/:id", adminHandler.GetUser)
			adminGroup.POST("/users", adminHandler.CreateUser)
			adminGroup.PATCH("/users/:id", adminHandler.UpdateUser)
			adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
		}
	}

	return engine
}
