package resources

import (
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenant/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	xtenantv1beta1 "github.com/rezakaramad/crosskit/types/xtenant/v1beta1"
)

// XContext is the concrete FunctionContext for this function.
type XContext = composer.FunctionContext[*xtenantv1beta1.XTenant, *inputv1beta1.Input]

// XComposer is the concrete BaseComposer for this function.
type XComposer = composer.BaseComposer[*xtenantv1beta1.XTenant, *inputv1beta1.Input]
