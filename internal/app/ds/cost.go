package ds

type Cost struct {
	ID          uint64 `gorm:"primaryKey"`
	Title       string `gorm:"type:varchar(40);not null"`
	Img         string `gorm:"type:varchar"`
	Info        string `gorm:"type:varchar"`
	Type_change bool   `gorm:"type:boolean not null"`
	IsDeleted   bool   `gorm:"boolean; not null; default: false"`
}
