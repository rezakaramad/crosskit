package runner

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/resource"
)

// Composer models the full lifecycle of one composed child resource.
type Composer[XR any, Input runtime.Object, Observed runtime.Object, Desired runtime.Object] interface {
	// The name of the status condition that gets set on the XR to report whether this child resource is ready.
	ConditionType() string
	// The stable name used to track this resource across reconciliations.
	ResourceName(ctx Context[XR, Input]) resource.Name
	// Compose builds the desired resource based on the input context.
	Compose(ctx Context[XR, Input]) (Desired, error)
	// IsReady checks if the observed resource is ready.
	IsReady(ctx Context[XR, Input], observed Observed) bool
	// ConnectionDetails extracts connection details from the observed resource.
	ConnectionDetails(ctx Context[XR, Input], observed Observed) map[string]string
}
