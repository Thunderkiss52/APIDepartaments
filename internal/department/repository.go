package department

import (
	"sort"

	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) DB() *gorm.DB { return r.db }

func (r *Repository) CreateDepartment(dep *Department) error {
	if r.db == nil {
		dep.ID = 1
		return nil
	}
	return r.db.Create(dep).Error
}
func (r *Repository) GetDepartment(id int64) (*Department, error) {
	if r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var dep Department
	if err := r.db.First(&dep, id).Error; err != nil {
		return nil, err
	}
	return &dep, nil
}
func (r *Repository) UpdateDepartment(dep *Department) error {
	if r.db == nil {
		return nil
	}
	return r.db.Save(dep).Error
}
func (r *Repository) DeleteDepartment(id int64) error {
	if r.db == nil {
		return nil
	}
	return r.db.Delete(&Department{}, id).Error
}
func (r *Repository) CreateEmployee(emp *Employee) error {
	if r.db == nil {
		emp.ID = 1
		return nil
	}
	return r.db.Create(emp).Error
}
func (r *Repository) GetEmployeesByDepartmentIDs(ids []int64) ([]Employee, error) {
	if r.db == nil {
		return []Employee{}, nil
	}
	var out []Employee
	if err := r.db.Where("department_id IN ?", ids).Find(&out).Error; err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt.Equal(out[j].CreatedAt) {
			return out[i].FullName < out[j].FullName
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out, nil
}

func (r *Repository) ListChildren(parentID int64) ([]Department, error) {
	if r.db == nil {
		return []Department{}, nil
	}
	var deps []Department
	if err := r.db.Where("parent_id = ?", parentID).Order("id ASC").Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *Repository) ListAllDepartments() ([]Department, error) {
	if r.db == nil {
		return []Department{}, nil
	}
	var deps []Department
	if err := r.db.Order("id ASC").Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *Repository) ListDepartmentsByIDs(ids []int64) ([]Department, error) {
	if r.db == nil {
		return []Department{}, nil
	}
	var deps []Department
	if len(ids) == 0 {
		return deps, nil
	}
	if err := r.db.Where("id IN ?", ids).Find(&deps).Error; err != nil {
		return nil, err
	}
	return deps, nil
}

func (r *Repository) ReassignEmployees(fromDepartmentID, toDepartmentID int64) error {
	if r.db == nil {
		return nil
	}
	return r.db.Model(&Employee{}).
		Where("department_id = ?", fromDepartmentID).
		Update("department_id", toDepartmentID).Error
}
