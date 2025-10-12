package ds

import "time"

type Cost_request struct {
	ID                           uint64                   `gorm:"primaryKey"`
	Status                       uint64                   `gorm:"not null;default:1"` // 1 - draft, 2 - deleted, 3 - pending, 4 - resolved, 5 - rejected
	ID_user                      uint64                   `gorm:"not null"`
	User                         User                     `gorm:"foreignKey:ID_user"`
	ID_moderator                 uint64                   `gorm:"default:null"`
	Moderator                    User                     `gorm:"foreignKey:ID_moderator"`
	Min_volume                   uint64                   `gorm:"not null"`
	Max_volume                   uint64                   `gorm:"not null"`
	CalculationResultFixCosts    float64                  `gorm:"default: 0"`
	CalculationResultChangeCosts float64                  `gorm:"default: 0"`
	Ratio                        float64                  `gorm:"not null"`
	Price_request_for_cost       []Price_request_for_cost `gorm:"foreignKey:ID_request"`
	CreatedAt                    time.Time                `gorm:"not null; default:now()"`
	FormedAt                     time.Time                `gorm:"default:null"`
	ClosedAt                     time.Time                `gorm:"default:null"`
}
