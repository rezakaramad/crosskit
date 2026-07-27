package resources

import (
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantargo/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	xtenantargo "github.com/rezakaramad/crosskit/types/xtenantargo"
)

// XContext is the concrete FunctionContext for this function.
type XContext = composer.FunctionContext[*xtenantargo.XTenantArgo, *inputv1beta1.Input]

// XComposer is the concrete BaseComposer for this function.
type XComposer = composer.BaseComposer[*xtenantargo.XTenantArgo, *inputv1beta1.Input]
