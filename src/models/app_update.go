package models

import "time"

type AppUpdate struct {
	ID        uint      `gorm:"primarykey"`
	Version   string    `gorm:"type:varchar(20);unique;not null"`
	Applied   bool      `gorm:"default:false"`
	AppliedAt time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}
