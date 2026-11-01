package resources

import (
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	xtenantssov1beta1 "github.com/rezakaramad/crosskit/types/xtenantsso/v1beta1"
)

// XContext is the concrete FunctionContext for this function.
type XContext = composer.FunctionContext[*xtenantssov1beta1.XTenantSSO, *inputv1beta1.Input]

// XComposer is the concrete BaseComposer for this function.
type XComposer = composer.BaseComposer[*xtenantssov1beta1.XTenantSSO, *inputv1beta1.Input]
