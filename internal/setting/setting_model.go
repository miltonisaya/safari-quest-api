package setting

import "time"

// Setting is one stored system setting value. Which keys exist, their types
// and ranges are defined in code (Definitions); only values live here.
type Setting struct {
	Key       string `gorm:"primaryKey;type:varchar(64)"`
	Value     string `gorm:"not null"`
	UpdatedBy *uint
	UpdatedAt time.Time
}

func (Setting) TableName() string { return "system_settings" }
