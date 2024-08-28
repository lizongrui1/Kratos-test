package model

import (
	"time"
)

type Student struct {
	ID        int32     `json:"id" gorm:"column:id;primaryKey"`
	Name      string    `json:"name" gorm:"column:name"`
	Info      string    `json:"info" gorm:"column:info"`
	Status    int32     `json:"status" gorm:"column:status"`
	Score     int32     `json:"score" gorm:"column:score"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (Student) TableName() string {
	return "students"
}
