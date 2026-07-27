package resources

import "fmt"

// labelRef returns an Argo CD Go-template expression that resolves a generator
// label value at render time, e.g. labelRef("argocd.rezakara.demo/environment")
// yields `{{ index .metadata.labels "argocd.rezakara.demo/environment" }}`.
func labelRef(key string) string {
	return fmt.Sprintf("{{ index .metadata.labels %q }}", key)
}
