package entity

import (
	"time"

	"gorm.io/gorm"
)

type DeliveryAttemptEntity struct {
	ID              string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeliveryID      string         `gorm:"type:uuid;not null;index"`
	AttemptNumber   int            `gorm:"not null"`
	RequestURL      string         `gorm:"type:varchar(500);not null"`
	RequestHeaders  string         `gorm:"type:text"`
	RequestBody     string         `gorm:"type:text"`
	ResponseStatus  int
	ResponseHeaders string    `gorm:"type:text"`
	ResponseBody    string    `gorm:"type:text"`
	Success         bool
	Error           string    `gorm:"type:text"`
	DurationMs      int
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	// Relations
	Delivery *DeliveryEntity `gorm:"foreignKey:DeliveryID"`
}

func (DeliveryAttemptEntity) TableName() string {
	return "webhook_delivery_attempts"
}
