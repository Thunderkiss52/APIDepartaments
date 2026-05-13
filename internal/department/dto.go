package department

import "encoding/json"

type CreateDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
}

type UpdateDepartmentRequest struct {
	Name     *string       `json:"name"`
	ParentID NullableInt64 `json:"parent_id"`
}

type CreateEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

type NullableInt64 struct {
	Set   bool
	Valid bool
	Value int64
}

func (n *NullableInt64) UnmarshalJSON(data []byte) error {
	n.Set = true
	if string(data) == "null" {
		n.Valid = false
		n.Value = 0
		return nil
	}
	n.Valid = true
	return json.Unmarshal(data, &n.Value)
}
