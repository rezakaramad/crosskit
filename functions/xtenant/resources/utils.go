package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/functions/xtenant/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenant/input/v1beta1"
)

var syncOptions = []string{
	"ServerSideApply=true",
	"ApplyOutOfSyncOnly=true",
	"RespectIgnoreDifferences=true",
}

// labelRef returns an Argo CD Go-template expression that resolves a generator
// label value at render time, e.g. labelRef("argocd.rezakara.demo/environment")
// yields `{{ index .metadata.labels "argocd.rezakara.demo/environment" }}`.
func labelRef(key string) string {
	return fmt.Sprintf("{{ index .metadata.labels %q }}", key)
}

// applicationSource builds an ApplicationSource with tenant identity and Helm configuration.
func applicationSource(
	repoURL string,
	path string,
	targetRevision string,
	helmConfig inputv1beta1.HelmConfig,
	tenant string,
	shortName string,
) *argocd.ApplicationSource {
	helm := &argocd.HelmSource{
		ValueFiles: helmConfig.ValueFiles,
		Parameters: []argocd.HelmParameter{
			{Name: "tenant.name", Value: tenant},
			{Name: "tenant.tenantShortName", Value: shortName},
		},
	}
	for _, p := range helmConfig.Parameters {
		helm.Parameters = append(helm.Parameters, argocd.HelmParameter{
			Name:  p.Name,
			Value: p.Value,
		})
	}
	return &argocd.ApplicationSource{
		RepoURL:        repoURL,
		Path:           path,
		TargetRevision: targetRevision,
		Helm:           helm,
	}
}
