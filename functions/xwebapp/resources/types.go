// Package resources implements the composers that build the child resources of an XWebApp.
package resources

import (
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xwebapp/input/v1beta1"
	"github.com/rezakaramad/crosskit/modules/composer"
	"github.com/rezakaramad/crosskit/types/xwebapp"
)

// XContext is the concrete FunctionContext for this function.
type XContext = composer.FunctionContext[*xwebapp.XWebApp, *inputv1beta1.Input]

// XComposer is the concrete BaseComposer for this function.
type XComposer = composer.BaseComposer[*xwebapp.XWebApp, *inputv1beta1.Input]
