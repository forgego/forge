package work

import "github.com/forgego/forge/schema"

type Member struct{ MemberGenerated }

func (Member) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.StringField("email", schema.Required(), schema.MaxLength(80), schema.Unique()),
		schema.StringField("api_token", schema.Required(), schema.WriteOnly(), schema.Editable(false)),
		schema.BoolField("is_staff", schema.Default(false), schema.WriteOnly(), schema.Editable(false)),
	}
}
func (Member) Meta() schema.Meta { return schema.Meta{TableName: "members", VerboseName: "Member"} }
func (Member) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Member) Hooks() *schema.ModelHooks { return nil }

type Employee struct{ EmployeeGenerated }

func (Employee) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required(), schema.MaxLength(80)),
		schema.EmailField("email", schema.Required(), schema.Unique()),
		schema.StringField("department", schema.Required(), schema.MaxLength(80), schema.Default("General")),
		schema.StringField("internal_token", schema.Default("system-only"), schema.WriteOnly(), schema.Editable(false)),
		schema.DateTimeField("created_at", schema.AutoNowAdd()),
		schema.DateTimeField("updated_at", schema.AutoNow()),
		// ACCEPTANCE EXTRA FIELD
	}
}
func (Employee) Meta() schema.Meta {
	return schema.Meta{TableName: "employees", VerboseName: "Employee", Indexes: []schema.Index{{Name: "idx_employees_name", Fields: []string{"name"}}}, OrderBy: []string{"name"}}
}
func (Employee) Relations() []schema.Relation {
	return []schema.Relation{}
}
func (Employee) Hooks() *schema.ModelHooks { return nil }

type ExpenseReport struct{ ExpenseReportGenerated }

func (ExpenseReport) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("employee_id", schema.Required()),
		schema.Int64Field("owner_id", schema.Required()),
		schema.StringField("title", schema.Required(), schema.MaxLength(80)),
		schema.StringField("status", schema.Required(), schema.MaxLength(80), schema.Default("draft")),
		schema.Int64Field("total_cents", schema.Required(), schema.Default(0), schema.MinValue(0)),
	}
}
func (ExpenseReport) Meta() schema.Meta {
	return schema.Meta{TableName: "expense_reports", VerboseName: "ExpenseReport"}
}
func (ExpenseReport) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("employee_id", "Employee", schema.OnDelete(schema.CascadePROTECT)),
		schema.ForeignKeyField("owner_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (ExpenseReport) Hooks() *schema.ModelHooks { return nil }

type Expense struct{ ExpenseGenerated }

func (Expense) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("report_id", schema.Required()),
		schema.Int64Field("amount_cents", schema.Required(), schema.MinValue(0)),
		schema.StringField("currency", schema.Required(), schema.MaxLength(80), schema.Default("USD")),
		schema.StringField("description", schema.Required(), schema.MaxLength(80)),
		schema.DateTimeField("incurred_at", schema.Optional()),
	}
}
func (Expense) Meta() schema.Meta { return schema.Meta{TableName: "expenses", VerboseName: "Expense"} }
func (Expense) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("report_id", "ExpenseReport", schema.OnDelete(schema.CascadeCASCADE)),
	}
}
func (Expense) Hooks() *schema.ModelHooks { return nil }

type Approval struct{ ApprovalGenerated }

func (Approval) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.Int64Field("report_id", schema.Required()),
		schema.Int64Field("approver_id", schema.Required()),
		schema.StringField("decision", schema.Required(), schema.MaxLength(80)),
		schema.TextField("note", schema.Optional()),
	}
}
func (Approval) Meta() schema.Meta {
	return schema.Meta{TableName: "approvals", VerboseName: "Approval"}
}
func (Approval) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("report_id", "ExpenseReport", schema.OnDelete(schema.CascadeCASCADE)),
		schema.ForeignKeyField("approver_id", "Member", schema.OnDelete(schema.CascadePROTECT)),
	}
}
func (Approval) Hooks() *schema.ModelHooks { return nil }
