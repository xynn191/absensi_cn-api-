package user

type Role string

const (
	RoleStudent Role = "STUDENT"
	RoleTeacher Role = "TEACHER"
	RoleBK      Role = "BK"
	RoleAdmin   Role = "ADMIN"
)

type User struct {
	ID           string  `gorm:"type:char(36);primaryKey"`
	Name         string  `gorm:"size:150;not null"`
	NIS          *string `gorm:"size:10;uniqueIndex"`
	Username     *string `gorm:"size:50;uniqueIndex"`
	PasswordHash string  `gorm:"size:255;not null"`
	Role         Role    `gorm:"size:30;not null"`
}
