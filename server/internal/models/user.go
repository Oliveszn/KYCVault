package models

type User struct {
	ID           uint   `gorm:"primaryKey;column:id" json:"id"`
	Email        string `gorm:"column:email;not null;unique" json:"email"`
	PasswordHash string `gorm:"column:passwordHash;not null" json:"-"` // Using json:"-" to exclude from JSON responses
}

func (User) TableName() string {
	return "users"
}
