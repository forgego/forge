package testdata

import (
	"database/sql"
	"time"

	"github.com/forgego/forge/schema"
)

// Organization represents a company with hierarchical structure
type Organization struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Code        string         `db:"code"` // Unique
	Description sql.NullString `db:"description"`
	ParentID    sql.NullInt64  `db:"parent_id"` // Self-referential FK for hierarchy
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// Employee represents employees with manager relationship (self-referential)
type Employee struct {
	schema.BaseSchema
	ID             int64          `db:"id"`
	OrganizationID int64          `db:"organization_id"` // FK to Organization
	ManagerID      sql.NullInt64  `db:"manager_id"`      // Self-referential FK
	EmployeeNumber string         `db:"employee_number"` // Unique
	FirstName      string         `db:"first_name"`
	LastName       string         `db:"last_name"`
	Email          string         `db:"email"` // Unique
	Phone          sql.NullString `db:"phone"`
	HireDate       time.Time      `db:"hire_date"`
	IsActive       bool           `db:"is_active"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

// Department represents departments with many-to-many employees
type Department struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Code        string         `db:"code"` // Unique
	Description sql.NullString `db:"description"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// EmployeeDepartment represents the many-to-many relationship between Employee and Department
type EmployeeDepartment struct {
	schema.BaseSchema
	ID           int64        `db:"id"`
	EmployeeID   int64        `db:"employee_id"`   // FK to Employee
	DepartmentID int64        `db:"department_id"` // FK to Department
	Role         string       `db:"role"`          // Extra field in join table
	StartDate    time.Time    `db:"start_date"`
	EndDate      sql.NullTime `db:"end_date"`
	CreatedAt    time.Time    `db:"created_at"`
}

// Project represents projects with polymorphic relationships
type Project struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Code        string         `db:"code"` // Unique
	Description sql.NullString `db:"description"`
	Status      string         `db:"status"` // planning, active, completed, cancelled
	StartDate   sql.NullTime   `db:"start_date"`
	EndDate     sql.NullTime   `db:"end_date"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// Task represents tasks with dependencies (self-referential many-to-many)
type Task struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	ProjectID   int64          `db:"project_id"`  // FK to Project
	AssigneeID  sql.NullInt64  `db:"assignee_id"` // FK to Employee
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	Status      string         `db:"status"`   // todo, in_progress, done, blocked
	Priority    string         `db:"priority"` // low, medium, high, critical
	DueDate     sql.NullTime   `db:"due_date"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// TaskDependency represents the many-to-many self-referential relationship for task dependencies
type TaskDependency struct {
	schema.BaseSchema
	ID             int64     `db:"id"`
	TaskID         int64     `db:"task_id"`         // FK to Task
	DependencyID   int64     `db:"dependency_id"`   // FK to Task (self-referential)
	DependencyType string    `db:"dependency_type"` // blocks, blocked_by, related
	CreatedAt      time.Time `db:"created_at"`
}

// ProjectEmployee represents the many-to-many relationship between Project and Employee
type ProjectEmployee struct {
	schema.BaseSchema
	ID         int64        `db:"id"`
	ProjectID  int64        `db:"project_id"`  // FK to Project
	EmployeeID int64        `db:"employee_id"` // FK to Employee
	Role       string       `db:"role"`        // project_manager, developer, designer, etc.
	StartDate  time.Time    `db:"start_date"`
	EndDate    sql.NullTime `db:"end_date"`
	CreatedAt  time.Time    `db:"created_at"`
}
