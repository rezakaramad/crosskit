package generator

import (
	"testing"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func TestHideSchemaPath(t *testing.T) {
	newSchema := func() *apiextv1.JSONSchemaProps {
		return &apiextv1.JSONSchemaProps{
			Properties: map[string]apiextv1.JSONSchemaProps{
				"spec": {
					Properties: map[string]apiextv1.JSONSchemaProps{
						"containers": {
							Type: "array",
							Items: &apiextv1.JSONSchemaPropsOrArray{Schema: &apiextv1.JSONSchemaProps{
								Required: []string{"name", "image"},
								Properties: map[string]apiextv1.JSONSchemaProps{
									"name":  {Type: "string"},
									"image": {Type: "string"},
								},
							}},
						},
						"tolerations": {Type: "array"},
					},
				},
			},
		}
	}

	t.Run("removes a nested property", func(t *testing.T) {
		s := newSchema()
		if err := hideSchemaPath(s, []string{"spec", "tolerations"}); err != nil {
			t.Fatal(err)
		}
		if _, ok := s.Properties["spec"].Properties["tolerations"]; ok {
			t.Error("tolerations should be removed")
		}
		if _, ok := s.Properties["spec"].Properties["containers"]; !ok {
			t.Error("containers should be kept")
		}
	})

	t.Run("traverses arrays and drops required entry", func(t *testing.T) {
		s := newSchema()
		if err := hideSchemaPath(s, []string{"spec", "containers", "image"}); err != nil {
			t.Fatal(err)
		}
		item := s.Properties["spec"].Properties["containers"].Items.Schema
		if _, ok := item.Properties["image"]; ok {
			t.Error("image should be removed")
		}
		if len(item.Required) != 1 || item.Required[0] != "name" {
			t.Errorf("unexpected required list: %v", item.Required)
		}
	})

	t.Run("errors on a missing path", func(t *testing.T) {
		if err := hideSchemaPath(newSchema(), []string{"spec", "nope"}); err == nil {
			t.Error("expected an error")
		}
	})
}
