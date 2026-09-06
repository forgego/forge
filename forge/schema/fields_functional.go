package schema

type FieldOpt func(*Field)

func Int64Field(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeInt64, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func Int32Field(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeInt32, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func StringField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeString, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func TextField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeText, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func EmailField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeEmail, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func URLField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeURL, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func BoolField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeBool, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func TimeField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeTime, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DateField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeDate, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DateTimeField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeDateTime, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func Float64Field(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeFloat64, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

// FloatField is an alias for Float64Field.
func FloatField(name string, opts ...FieldOpt) Field {
	return Float64Field(name, opts...)
}

func Float32Field(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeFloat32, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func DecimalField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeDecimal, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func JSONField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeJSON, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func BytesField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeBytes, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func UUIDField(name string, opts ...FieldOpt) Field {
	f := Field{Name: name, Type: TypeUUID, Editable: true, Serialize: true}
	for _, opt := range opts {
		opt(&f)
	}
	return f
}

func Default(value interface{}) FieldOpt {
	return func(f *Field) { f.Default = value }
}

func DBColumn(name string) FieldOpt {
	return func(f *Field) { f.DBColumn = name }
}

func DBType(t string) FieldOpt {
	return func(f *Field) { f.DBType = t }
}

func DBCollation(c string) FieldOpt {
	return func(f *Field) { f.DBCollation = c }
}

func DBComment(c string) FieldOpt {
	return func(f *Field) { f.DBComment = c }
}

func DBTablespace(sp string) FieldOpt {
	return func(f *Field) { f.DBTablespace = sp }
}

func DBDefault(expr string) FieldOpt {
	return func(f *Field) { f.DBDefault = expr }
}

func DBIndex() FieldOpt {
	return func(f *Field) { f.DBIndex = true }
}

func GeneratedColumn(expr string, stored bool) FieldOpt {
	return func(f *Field) {
		f.Generated = true
		f.GeneratedExpr = expr
		f.IsStored = stored
	}
}

func Required() FieldOpt {
	return func(f *Field) { f.Required = true }
}

func Optional() FieldOpt {
	return func(f *Field) { f.Required = false }
}

func Unique() FieldOpt {
	return func(f *Field) { f.Unique = true }
}

func Blank() FieldOpt {
	return func(f *Field) { f.Blank = true }
}

func MaxLength(n int) FieldOpt {
	return func(f *Field) { f.MaxLength = &n }
}

func MinLength(n int) FieldOpt {
	return func(f *Field) { f.MinLength = &n }
}

func MaxValue(v float64) FieldOpt {
	return func(f *Field) { f.MaxValue = &v }
}

func MinValue(v float64) FieldOpt {
	return func(f *Field) { f.MinValue = &v }
}

func MaxDigits(n int) FieldOpt {
	return func(f *Field) { f.MaxDigits = &n }
}

func DecimalPlaces(n int) FieldOpt {
	return func(f *Field) { f.DecimalPlaces = &n }
}

func ChoicesOpts(choices ...Choice) FieldOpt {
	return func(f *Field) { f.Choices = choices }
}

func ChoicesFromPairsOpts(pairs ...string) FieldOpt {
	return func(f *Field) { f.Choices = Choices(pairs...) }
}

func Validators(v ...Validator) FieldOpt {
	return func(f *Field) { f.Validators = append(f.Validators, v...) }
}

func ValidationTag(tag string) FieldOpt {
	return func(f *Field) { f.ValidationTag = tag }
}

func ErrorMessages(msg map[string]string) FieldOpt {
	return func(f *Field) { f.ErrorMessages = msg }
}

func HelpText(text string) FieldOpt {
	return func(f *Field) { f.HelpText = text }
}

func VerboseName(name string) FieldOpt {
	return func(f *Field) { f.VerboseName = name }
}

func UniqueForDate(field string) FieldOpt {
	return func(f *Field) { f.UniqueForDate = field }
}

func UniqueForMonth(field string) FieldOpt {
	return func(f *Field) { f.UniqueForMonth = field }
}

func UniqueForYear(field string) FieldOpt {
	return func(f *Field) { f.UniqueForYear = field }
}

func Editable(editable bool) FieldOpt {
	return func(f *Field) { f.Editable = editable }
}

func Serialize(serialize bool) FieldOpt {
	return func(f *Field) { f.Serialize = serialize }
}

func Primary() FieldOpt {
	return func(f *Field) { f.PrimaryKey = true }
}

func AutoIncrement() FieldOpt {
	return func(f *Field) { f.AutoIncrement = true }
}

func AutoNow() FieldOpt {
	return func(f *Field) { f.AutoNow = true }
}

func AutoNowAdd() FieldOpt {
	return func(f *Field) { f.AutoNowAdd = true }
}

func WriteOnly() FieldOpt {
	return func(f *Field) { f.Serialize = false }
}
