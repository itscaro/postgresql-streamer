package dev

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type TestModel struct {
	ID        uint           `gorm:"primary_key"`
	Source    postgres.Jsonb `json:"source" gorm:"index:idx_source"`
	Content   postgres.Jsonb `json:"content"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
