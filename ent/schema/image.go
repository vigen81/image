package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
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
		field.String("tmp_url"),
		field.String("service").Optional(),
		field.String("type").Optional(),
		field.Bool("is_proceed").Default(false),
		field.Bool("is_deleted").Default(false),
		field.String("content_type"),
	}
}

func (Image) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the Image.
func (Image) Edges() []ent.Edge {
	return nil
}
