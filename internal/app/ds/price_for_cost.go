package ds

type Price_request_for_cost struct {
	ID_request   uint64       `gorm:"primaryKey"`
	ID_cost      uint64       `gorm:"primaryKey"`
	Cost_request Cost_request `gorm:"foreignKey:ID_request"`
	Cost         Cost         `gorm:"foreignKey:ID_cost"`
	Cost_price   float64      `gorm:"type:double precision"`
	Main         bool         `gorm:"type:boolean; default:False "`
}
