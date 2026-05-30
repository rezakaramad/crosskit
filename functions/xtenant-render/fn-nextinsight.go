package main

import (
	"context"
	"maps"

	"github.com/rezakaramad/crossplane-toolkit/modules/nextinsight"

	"github.com/crossplane/function-sdk-go/resource/composed"
)

// fetchTenantLabels returns labels for namespace-boundary resources.
// If the client, teamID, or labelPrefix is missing, it returns no labels and no error.
func fetchTenantLabels(ctx context.Context, client nextinsight.Client, teamID, labelPrefix string) (map[string]string, error) {
	if client == nil || teamID == "" || labelPrefix == "" {
		return map[string]string{}, nil
	}
	return client.FetchTenantLabels(ctx, teamID, labelPrefix)
}

// applyNextInsightLabels merges the extra labels onto each composed resource.
// It is a no-op when extra is empty, so callers don't need to guard the call.
func applyNextInsightLabels(extra map[string]string, resources ...*composed.Unstructured) {
	if len(extra) == 0 {
		return
	}

	for _, res := range resources {
		if res == nil {
			continue
		}

		existing := res.GetLabels()
		if existing == nil {
			existing = make(map[string]string, len(extra))
		}
		maps.Copy(existing, extra)
		res.SetLabels(existing)
	}
}
