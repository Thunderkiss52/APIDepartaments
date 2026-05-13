-- +goose Up
CREATE TABLE departments (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id INT NULL REFERENCES departments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX departments_parent_name_unique
ON departments (parent_id, lower(name))
WHERE parent_id IS NOT NULL;

CREATE UNIQUE INDEX departments_root_name_unique
ON departments (lower(name))
WHERE parent_id IS NULL;

CREATE INDEX departments_parent_id_idx ON departments(parent_id);

CREATE TABLE employees (
    id SERIAL PRIMARY KEY,
    department_id INT NOT NULL REFERENCES departments(id) ON DELETE CASCADE,
    full_name VARCHAR(200) NOT NULL,
    position VARCHAR(200) NOT NULL,
    hired_at DATE NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX employees_department_id_idx ON employees(department_id);

-- +goose Down
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS departments;

