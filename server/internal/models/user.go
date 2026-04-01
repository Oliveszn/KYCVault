package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `gorm:"column:email;not null;unique" json:"email"`
	FirstName    string    `gorm:"column:firstname;not null" json:"firstname"`
	LastName     string    `gorm:"column:lastname;not null" json:"lastname"`
	Role         string    `gorm:"column:role;not null;default:user" json:"role"`
	PasswordHash string    `gorm:"column:passwordHash;not null" json:"-"` // Using json:"-" to exclude from JSON responses
}

func (User) TableName() string {
	return "users"
}
