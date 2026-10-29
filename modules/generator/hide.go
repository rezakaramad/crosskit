package generator

import (
	"fmt"
	"slices"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// hideSchemaPath removes the property at path from the schema, so the API server
// rejects it as an unknown field. Array levels are traversed transparently.
// It errors when a segment is missing so upstream renames can't silently un-hide a field.
func hideSchemaPath(s *apiextv1.JSONSchemaProps, path []string) error {
	if s.Items != nil && s.Items.Schema != nil {
		return hideSchemaPath(s.Items.Schema, path)
	}

	name := path[0]
	prop, ok := s.Properties[name]
	if !ok {
		return fmt.Errorf("property %q not found in schema", name)
	}

	if len(path) == 1 {
		delete(s.Properties, name)
		s.Required = slices.DeleteFunc(s.Required, func(r string) bool { return r == name })
		return nil
	}

	if err := hideSchemaPath(&prop, path[1:]); err != nil {
		return err
	}
	s.Properties[name] = prop
	return nil
}
