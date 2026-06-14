package validate

import (
	"github.com/crossplane/function-sdk-go/resource"
	xtenant "github.com/rezakaramad/crossplane-toolkit/types/xtenant"
)

// SetPhase writes status.phase onto the XR so Crossplane surfaces it.
func SetPhase(xr *resource.Composite, phase xtenant.Phase) {
	_ = xr.Resource.SetValue("status.phase", string(phase))
}
