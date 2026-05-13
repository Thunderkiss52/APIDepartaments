package department

import "time"

type Department struct {
	ID        int64        `gorm:"primaryKey" json:"id"`
	Name      string       `gorm:"size:200;not null" json:"name"`
	ParentID  *int64       `json:"parent_id"`
	CreatedAt time.Time    `json:"created_at"`
	Employees []Employee   `gorm:"foreignKey:DepartmentID" json:"-"`
	Children  []Department `gorm:"foreignKey:ParentID" json:"-"`
}

type Employee struct {
	ID           int64      `gorm:"primaryKey" json:"id"`
	DepartmentID int64      `json:"department_id"`
	FullName     string     `gorm:"size:200;not null" json:"full_name"`
	Position     string     `gorm:"size:200;not null" json:"position"`
	HiredAt      *time.Time `gorm:"type:date" json:"hired_at"`
	CreatedAt    time.Time  `json:"created_at"`
}
