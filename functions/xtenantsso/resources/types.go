package resources

import (
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantentra/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	xtenantentra "github.com/rezakaramad/crosskit/types/xtenantentra"
)

// XContext is the concrete FunctionContext for this function.
type XContext = composer.FunctionContext[*xtenantentra.XTenantEntra, *inputv1beta1.Input]

// XComposer is the concrete BaseComposer for this function.
type XComposer = composer.BaseComposer[*xtenantentra.XTenantEntra, *inputv1beta1.Input]
