package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"gitlab.smartbet.am/golang/smart-image/internal/service/processor"
)

// Image holds the schema definition for the Image entity.
type Image struct {
	ent.Schema
}

// Fields of the Image.
func (Image) Fields() []ent.Field {
	return []ent.Field{
		field.String("uuid").Unique(),
		field.String("url").Optional(),
		field.String("object_id").Optional(),
		field.String("tmp_url"),
		field.String("service").Optional(),
		field.String("type").Optional(),
		field.Bool("is_proceed").Default(false),
		field.Bool("is_deleted").Default(false),
		field.String("content_type"),
		field.JSON("size", processor.Size{}).Optional(),
	}
}

// Edges of the Image.
func (Image) Edges() []ent.Edge {
	return nil
}
