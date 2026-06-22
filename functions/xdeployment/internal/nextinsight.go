package render

import (
	"context"
	"maps"

	"github.com/rezakaramad/crosskit/modules/nextinsight"

	"github.com/crossplane/function-sdk-go/resource/composed"
)

// FetchApplicationLabels returns workload labels from Next-Insight for the given appID.
// Returns empty labels (no error) when the client, appID, or labelPrefix is missing.
func FetchApplicationLabels(ctx context.Context, client nextinsight.Client, appID, labelPrefix string) (map[string]string, error) {
	if client == nil || appID == "" || labelPrefix == "" {
		return map[string]string{}, nil
	}
	meta, err := client.FetchApplicationMetadata(ctx, appID)
	if err != nil {
		return nil, err
	}
	return meta.WorkloadLabels(labelPrefix), nil
}

// ApplyLabels merges extra labels onto each composed resource.
// It is a no-op when extra is empty.
func ApplyLabels(extra map[string]string, resources ...*composed.Unstructured) {
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
