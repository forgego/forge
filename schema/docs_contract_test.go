package schema

import "testing"

func TestDocumentedFunctionalFieldAndRelationOptions(t *testing.T) {
	field := DecimalField("price_with_tax", GeneratedColumn("price * 1.1", true), DBIndex())
	if !field.Generated || field.GeneratedExpr != "price * 1.1" || !field.IsStored || !field.DBIndex {
		t.Fatalf("functional options lost field metadata: %+v", field)
	}
	status := StringField("status", ChoicesOpts(NewChoice("draft", "Draft")))
	if len(status.Choices) != 1 || status.Choices[0].Value != "draft" {
		t.Fatalf("choices option lost metadata: %+v", status.Choices)
	}
	relation := ManyToManyField("tags", "Tag", Through("order_tags"), RelatedName("orders"))
	if relation.Type != RelationManyToMany || relation.Through != "order_tags" || relation.RelatedName != "orders" {
		t.Fatalf("relation options lost metadata: %+v", relation)
	}
}

func TestFieldReferenceConstructors(t *testing.T) {
	cases := []struct {
		row    string
		fields []Field
		want   []FieldType
	}{
		{"A-153", []Field{Int64Field("id"), Int32Field("age")}, []FieldType{TypeInt64, TypeInt32}},
		{"A-154", []Field{Float64Field("score"), Float32Field("ratio")}, []FieldType{TypeFloat64, TypeFloat32}},
		{"A-155", []Field{DecimalField("amount", MaxDigits(12), DecimalPlaces(2))}, []FieldType{TypeDecimal}},
		{"A-156", []Field{StringField("name"), TextField("body")}, []FieldType{TypeString, TypeText}},
		{"A-157", []Field{EmailField("email"), URLField("website")}, []FieldType{TypeEmail, TypeURL}},
		{"A-158", []Field{UUIDField("identifier")}, []FieldType{TypeUUID}},
		{"A-159", []Field{TimeField("seen_at"), DateField("day"), DateTimeField("timestamp")}, []FieldType{TypeTime, TypeDate, TypeDateTime}},
		{"A-160", []Field{BoolField("active"), JSONField("payload"), BytesField("raw")}, []FieldType{TypeBool, TypeJSON, TypeBytes}},
	}
	for _, tc := range cases {
		t.Run(tc.row, func(t *testing.T) {
			for i, field := range tc.fields {
				if field.Type != tc.want[i] {
					t.Errorf("field %q: type = %v, want %v", field.Name, field.Type, tc.want[i])
				}
			}
		})
	}
}

func TestRelationReferenceConstructors(t *testing.T) {
	cases := []struct {
		row      string
		relation Relation
		wantType RelationType
	}{
		{"A-170", ForeignKeyField("author_id", "User", OnDelete(CascadeCASCADE), RelatedName("posts"), RelatedQueryName("post")), RelationForeignKey},
		{"A-172", OneToOneField("profile_id", "UserProfile", OnDelete(CascadeCASCADE), RelatedName("user")), RelationOneToOne},
		{"A-173", ManyToManyField("tags", "Tag", Through("post_tags"), RelatedName("posts")), RelationManyToMany},
	}
	for _, tc := range cases {
		t.Run(tc.row, func(t *testing.T) {
			if tc.relation.Type != tc.wantType || tc.relation.To == "" {
				t.Fatalf("unexpected relation metadata: %+v", tc.relation)
			}
		})
	}
	t.Run("A-178", func(t *testing.T) {
		relation := ForeignKeyField("author_id", "User", RelatedName("posts"))
		if relation.RelatedName != "posts" {
			t.Fatalf("reverse name not stored: %+v", relation)
		}
	})
}
