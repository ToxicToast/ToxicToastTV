package entity

import "time"

// UserRoleEntity represents the user-role relationship in the database
type UserRoleEntity struct {
	UserID    string    `gorm:"type:uuid;primaryKey;index:idx_user_role"`
	RoleID    string    `gorm:"type:uuid;primaryKey;index:idx_user_role"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for UserRoleEntity
func (UserRoleEntity) TableName() string {
	return "azkaban_user_roles"
}
