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
