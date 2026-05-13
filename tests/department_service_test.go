package tests

import (
	"errors"
	"testing"

	"org-structure-api/internal/department"

	"gorm.io/gorm"
)

type fakeRepo struct {
	departments     map[int64]*department.Department
	children        map[int64][]department.Department
	employees       map[int64][]department.Employee
	reassignCalls   []reassignCall
	deletedDeptIDs  []int64
	deletedEmpCalls []int64
}

type reassignCall struct {
	from int64
	to   int64
}

func (f *fakeRepo) CreateDepartment(dep *department.Department) error { dep.ID = 100; return nil }
func (f *fakeRepo) GetDepartment(id int64) (*department.Department, error) {
	if dep, ok := f.departments[id]; ok {
		return dep, nil
	}
	return nil, errors.New("record not found")
}
func (f *fakeRepo) UpdateDepartment(dep *department.Department) error {
	f.departments[dep.ID] = dep
	return nil
}
func (f *fakeRepo) DeleteDepartment(id int64) error {
	f.deletedDeptIDs = append(f.deletedDeptIDs, id)
	delete(f.departments, id)
	return nil
}
func (f *fakeRepo) CreateEmployee(emp *department.Employee) error { return nil }
func (f *fakeRepo) GetEmployeesByDepartmentIDs(ids []int64) ([]department.Employee, error) {
	var out []department.Employee
	for _, id := range ids {
		out = append(out, f.employees[id]...)
	}
	return out, nil
}
func (f *fakeRepo) ListChildren(parentID int64) ([]department.Department, error) {
	return f.children[parentID], nil
}
func (f *fakeRepo) ListAllDepartments() ([]department.Department, error) { return nil, nil }
func (f *fakeRepo) ListDepartmentsByIDs(ids []int64) ([]department.Department, error) {
	var out []department.Department
	for _, id := range ids {
		if dep, ok := f.departments[id]; ok {
			out = append(out, *dep)
		}
	}
	return out, nil
}
func (f *fakeRepo) ReassignEmployees(fromDepartmentID, toDepartmentID int64) error {
	f.reassignCalls = append(f.reassignCalls, reassignCall{from: fromDepartmentID, to: toDepartmentID})
	return nil
}
func (f *fakeRepo) DB() *gorm.DB { return nil }

func TestCreateDepartmentValidation(t *testing.T) {
	svc := department.NewService(&fakeRepo{departments: map[int64]*department.Department{}})
	if _, err := svc.CreateDepartment(department.CreateDepartmentRequest{Name: "   "}); !errors.Is(err, department.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetDepartmentTreeDepth(t *testing.T) {
	repo := &fakeRepo{
		departments: map[int64]*department.Department{
			1: &department.Department{ID: 1, Name: "IT"},
			2: &department.Department{ID: 2, Name: "Backend", ParentID: ptr(1)},
			3: &department.Department{ID: 3, Name: "PHP", ParentID: ptr(2)},
		},
		children: map[int64][]department.Department{
			1: []department.Department{
				{ID: 2, Name: "Backend", ParentID: ptr(1)},
			},
			2: []department.Department{
				{ID: 3, Name: "PHP", ParentID: ptr(2)},
			},
		},
		employees: map[int64][]department.Employee{
			1: []department.Employee{
				{ID: 11, DepartmentID: 1, FullName: "Ivan Ivanov", Position: "Lead"},
			},
		},
	}
	svc := department.NewService(repo)
	out, err := svc.GetDepartment(1, 2, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	children, ok := out["children"].([]department.DepartmentNode)
	if !ok {
		t.Fatalf("children type mismatch: %T", out["children"])
	}
	if len(children) != 1 || children[0].Department.ID != 2 {
		t.Fatalf("unexpected tree: %#v", out)
	}
	if len(children[0].Children) != 1 || children[0].Children[0].Department.ID != 3 {
		t.Fatalf("depth not expanded: %#v", out)
	}
}

func TestUpdateDepartmentRejectsCycle(t *testing.T) {
	repo := &fakeRepo{
		departments: map[int64]*department.Department{
			1: &department.Department{ID: 1, Name: "IT"},
			2: &department.Department{ID: 2, Name: "Backend", ParentID: ptr(1)},
			3: &department.Department{ID: 3, Name: "PHP", ParentID: ptr(2)},
		},
		children: map[int64][]department.Department{
			1: []department.Department{
				{ID: 2, Name: "Backend", ParentID: ptr(1)},
			},
			2: []department.Department{
				{ID: 3, Name: "PHP", ParentID: ptr(2)},
			},
		},
	}
	svc := department.NewService(repo)
	if _, err := svc.UpdateDepartment(1, department.UpdateDepartmentRequest{ParentID: department.NullableInt64{Set: true, Valid: true, Value: 3}}); !errors.Is(err, department.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestDeleteDepartmentReassign(t *testing.T) {
	repo := &fakeRepo{
		departments: map[int64]*department.Department{
			1: &department.Department{ID: 1, Name: "IT"},
			2: &department.Department{ID: 2, Name: "Backend", ParentID: ptr(1)},
			5: &department.Department{ID: 5, Name: "Support"},
		},
		children: map[int64][]department.Department{
			1: []department.Department{
				{ID: 2, Name: "Backend", ParentID: ptr(1)},
			},
		},
	}
	svc := department.NewService(repo)
	target := int64(5)
	if err := svc.DeleteDepartment(1, "reassign", &target); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.reassignCalls) != 1 {
		t.Fatalf("expected one reassign call, got %#v", repo.reassignCalls)
	}
}

func TestDeleteDepartmentRejectsReassignIntoSubtree(t *testing.T) {
	repo := &fakeRepo{
		departments: map[int64]*department.Department{
			1: &department.Department{ID: 1, Name: "IT"},
			2: &department.Department{ID: 2, Name: "Backend", ParentID: ptr(1)},
		},
		children: map[int64][]department.Department{
			1: []department.Department{
				{ID: 2, Name: "Backend", ParentID: ptr(1)},
			},
		},
	}
	svc := department.NewService(repo)
	target := int64(2)
	if err := svc.DeleteDepartment(1, "reassign", &target); !errors.Is(err, department.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func ptr(v int64) *int64 { return &v }
