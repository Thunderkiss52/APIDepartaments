package department

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")
var ErrValidation = errors.New("validation error")

type DepartmentRepository interface {
	CreateDepartment(*Department) error
	GetDepartment(int64) (*Department, error)
	UpdateDepartment(*Department) error
	DeleteDepartment(int64) error
	CreateEmployee(*Employee) error
	GetEmployeesByDepartmentIDs([]int64) ([]Employee, error)
	ListChildren(int64) ([]Department, error)
	ListAllDepartments() ([]Department, error)
	ListDepartmentsByIDs([]int64) ([]Department, error)
	ReassignEmployees(int64, int64) error
	DB() *gorm.DB
}

type Service struct{ repo DepartmentRepository }

func NewService(repo DepartmentRepository) *Service { return &Service{repo: repo} }

type DepartmentNode struct {
	Department Department       `json:"department"`
	Employees  []Employee       `json:"employees"`
	Children   []DepartmentNode `json:"children"`
}

func (s *Service) ListDepartments() ([]Department, error) {
	deps, err := s.repo.ListAllDepartments()
	if err != nil {
		return nil, mapDBError(err)
	}
	return deps, nil
}

func (s *Service) CreateDepartment(req CreateDepartmentRequest) (*Department, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) < 1 || len(name) > 200 {
		return nil, ErrValidation
	}
	dep := &Department{Name: name, ParentID: req.ParentID}
	if req.ParentID != nil {
		if *req.ParentID <= 0 {
			return nil, ErrValidation
		}
		if _, err := s.repo.GetDepartment(*req.ParentID); err != nil {
			return nil, ErrNotFound
		}
	}
	if err := s.repo.CreateDepartment(dep); err != nil {
		return nil, mapDBError(err)
	}
	return dep, nil
}

func (s *Service) CreateEmployee(deptID int64, req CreateEmployeeRequest) (*Employee, error) {
	fn := strings.TrimSpace(req.FullName)
	pos := strings.TrimSpace(req.Position)
	if len(fn) < 1 || len(fn) > 200 || len(pos) < 1 || len(pos) > 200 {
		return nil, ErrValidation
	}
	if _, err := s.repo.GetDepartment(deptID); err != nil {
		return nil, ErrNotFound
	}
	emp := &Employee{DepartmentID: deptID, FullName: fn, Position: pos}
	if req.HiredAt != nil && strings.TrimSpace(*req.HiredAt) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.HiredAt))
		if err != nil {
			return nil, ErrValidation
		}
		emp.HiredAt = &t
	}
	if err := s.repo.CreateEmployee(emp); err != nil {
		return nil, mapDBError(err)
	}
	return emp, nil
}

func (s *Service) GetDepartment(id int64, depth int, includeEmployees bool) (map[string]any, error) {
	dep, err := s.repo.GetDepartment(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if depth < 0 || depth > 5 {
		return nil, ErrValidation
	}
	node, err := s.buildNode(*dep, depth, includeEmployees)
	if err != nil {
		return nil, err
	}
	out := map[string]any{
		"department": node.Department,
		"employees":  node.Employees,
		"children":   node.Children,
	}
	return out, nil
}

func (s *Service) UpdateDepartment(id int64, req UpdateDepartmentRequest) (*Department, error) {
	dep, err := s.repo.GetDepartment(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if req.Name != nil {
		v := strings.TrimSpace(*req.Name)
		if len(v) < 1 || len(v) > 200 {
			return nil, ErrValidation
		}
		dep.Name = v
	}
	if req.ParentID.Set {
		if !req.ParentID.Valid {
			dep.ParentID = nil
		} else {
			parentID := req.ParentID.Value
			if parentID == id {
				return nil, ErrConflict
			}
			if parentID <= 0 {
				return nil, ErrValidation
			}
			if _, err := s.repo.GetDepartment(parentID); err != nil {
				return nil, ErrNotFound
			}
			if err := s.ensureNoCycle(id, parentID); err != nil {
				return nil, err
			}
			dep.ParentID = &parentID
		}
	}
	if err := s.repo.UpdateDepartment(dep); err != nil {
		return nil, mapDBError(err)
	}
	return dep, nil
}

func (s *Service) DeleteDepartment(id int64, mode string, reassignTo *int64) error {
	if mode != "cascade" && mode != "reassign" {
		return ErrValidation
	}
	dep, err := s.repo.GetDepartment(id)
	if err != nil {
		return ErrNotFound
	}
	subtree, err := s.collectSubtreeIDs(id)
	if err != nil {
		return err
	}
	if mode == "reassign" {
		if reassignTo == nil {
			return ErrValidation
		}
		if *reassignTo == id {
			return ErrConflict
		}
		if containsID(subtree, *reassignTo) {
			return ErrConflict
		}
		if _, err := s.repo.GetDepartment(*reassignTo); err != nil {
			return ErrNotFound
		}
		if err := s.repo.ReassignEmployees(id, *reassignTo); err != nil {
			return mapDBError(err)
		}
	}
	if err := s.deleteSubtree(dep.ID, subtree); err != nil {
		return err
	}
	return nil
}

func isNotFound(err error) bool { return errors.Is(err, gorm.ErrRecordNotFound) }

func (s *Service) buildNode(dep Department, depth int, includeEmployees bool) (DepartmentNode, error) {
	node := DepartmentNode{Department: dep, Employees: []Employee{}, Children: []DepartmentNode{}}
	if includeEmployees {
		emps, err := s.repo.GetEmployeesByDepartmentIDs([]int64{dep.ID})
		if err != nil {
			return node, mapDBError(err)
		}
		node.Employees = emps
	}
	if depth == 0 {
		return node, nil
	}
	children, err := s.repo.ListChildren(dep.ID)
	if err != nil {
		return node, mapDBError(err)
	}
	for _, child := range children {
		childNode, err := s.buildNode(child, depth-1, includeEmployees)
		if err != nil {
			return node, err
		}
		node.Children = append(node.Children, childNode)
	}
	return node, nil
}

func (s *Service) ensureNoCycle(id, newParentID int64) error {
	subtree, err := s.collectSubtreeIDs(id)
	if err != nil {
		return err
	}
	if containsID(subtree, newParentID) {
		return ErrConflict
	}
	return nil
}

func (s *Service) collectSubtreeIDs(rootID int64) ([]int64, error) {
	ids := []int64{rootID}
	queue := []int64{rootID}
	seen := map[int64]bool{rootID: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		children, err := s.repo.ListChildren(current)
		if err != nil {
			return nil, mapDBError(err)
		}
		for _, child := range children {
			if seen[child.ID] {
				continue
			}
			seen[child.ID] = true
			ids = append(ids, child.ID)
			queue = append(queue, child.ID)
		}
	}
	return ids, nil
}

func (s *Service) deleteSubtree(rootID int64, subtree []int64) error {
	if len(subtree) == 0 {
		subtree = []int64{rootID}
	}
	if s.repo.DB() == nil {
		return nil
	}
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("department_id IN ?", subtree).Delete(&Employee{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id IN ?", subtree).Delete(&Department{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case strings.Contains(msg, "duplicate key"), strings.Contains(msg, "unique constraint"), strings.Contains(msg, "violates unique constraint"):
		return ErrConflict
	default:
		return err
	}
}

func containsID(ids []int64, target int64) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
