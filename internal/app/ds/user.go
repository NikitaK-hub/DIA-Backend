package ds

import (
	"DIA_Backend/internal/app/role"

	"github.com/google/uuid"
)

type User struct {
	ID       uint64    `gorm:"primary_key"`
	UUID     uuid.UUID `gorm:"type:uuid; unique; not null; default:gen_random_uuid()"`
	Username string    `gorm:"type:varchar(50); unique; not null"`
	Password string    `gorm:"type:varchar(50); not null"`
	Role     role.Role `gorm:"type:integer; not null; default:0"`
}
