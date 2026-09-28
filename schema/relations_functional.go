package schema

type RelationOpt func(*Relation)

func ForeignKeyField(name, to string, opts ...RelationOpt) Relation {
	r := Relation{
		Name:         name,
		Type:         RelationForeignKey,
		To:           to,
		DBConstraint: true,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func OneToOneField(name, to string, opts ...RelationOpt) Relation {
	r := Relation{
		Name:         name,
		Type:         RelationOneToOne,
		To:           to,
		DBConstraint: true,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

func ManyToManyField(name, to string, opts ...RelationOpt) Relation {
	r := Relation{
		Name: name,
		Type: RelationManyToMany,
		To:   to,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return r
}

// Relation Options

func RelatedName(name string) RelationOpt {
	return func(r *Relation) {
		r.RelatedName = name
	}
}

func RelatedQueryName(name string) RelationOpt {
	return func(r *Relation) {
		r.RelatedQueryName = name
	}
}

func OnDelete(cascade CascadeType) RelationOpt {
	return func(r *Relation) {
		r.OnDelete = cascade
	}
}

func OnUpdate(cascade CascadeType) RelationOpt {
	return func(r *Relation) {
		r.OnUpdate = cascade
	}
}

func LimitChoicesTo(limit interface{}) RelationOpt {
	return func(r *Relation) {
		r.LimitChoicesTo = limit
	}
}

func ParentLink() RelationOpt {
	return func(r *Relation) {
		r.ParentLink = true
	}
}

func DBConstraint(enabled bool) RelationOpt {
	return func(r *Relation) {
		r.DBConstraint = enabled
	}
}

func ConstraintName(name string) RelationOpt {
	return func(r *Relation) {
		r.ConstraintName = name
	}
}

func Deferrable(deferrable DeferrableType) RelationOpt {
	return func(r *Relation) {
		r.Deferrable = deferrable
	}
}

func Match(matchType FKMatchType) RelationOpt {
	return func(r *Relation) {
		r.Match = matchType
	}
}

func Swappable(swappable string) RelationOpt {
	return func(r *Relation) {
		r.Swappable = swappable
	}
}

func Through(throughTable string) RelationOpt {
	return func(r *Relation) {
		r.Through = throughTable
	}
}

func Symmetrical() RelationOpt {
	return func(r *Relation) {
		r.Symmetrical = true
	}
}
