package main

import (
	"context"
	"fmt"

	"github.com/rezakaramad/crosskit/functions/xtenantargo/argocd"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantargo/input/v1beta1"
	"github.com/rezakaramad/crosskit/functions/xtenantargo/resources"
	"github.com/rezakaramad/crosskit/modules/composer"
	"github.com/rezakaramad/crosskit/types/xtenantargo"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function is the gRPC server that Crossplane calls to render tenant resources.
type Function struct {
	// Generated from Crossplane's protobuf service definition.
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// NewFunction creates a Function and registers the ArgoCD types with the
// composed resource scheme so they can be marshalled during reconciliation.
// See [Schema Registration](../SchemeRegistration.md).
func NewFunction(log logging.Logger) (*Function, error) {
	if err := argocd.AddToScheme(composed.Scheme); err != nil {
		return nil, errors.Wrap(err, "cannot register ArgoCD types with the scheme")
	}
	return &Function{log: log}, nil
}

// buildComposers constructs all resource composers for the XTenantArgo
// composition, returning the full set of composers to run during reconciliation.
func buildComposers(fnContext resources.XContext) ([]composer.ComposableResource, error) {
	argocdApplicationSet, err := resources.NewArgoCDApplicationSet(fnContext)
	if err != nil {
		return nil, err
	}
	return []composer.ComposableResource{
		argocdApplicationSet,
	}, nil
}

// RunFunction is the entry point for the composition function.
// Crossplane passes everything the function needs to run in a RunFunctionRequest struct.
// The function tells Crossplane what resources it should compose by returning a RunFunctionResponse struct.
func (f *Function) RunFunction(
	_ context.Context,
	req *fnv1.RunFunctionRequest,
) (*fnv1.RunFunctionResponse, error) {
	var xd xtenantargo.XTenantArgo

	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	// Initialize the function response with a default TTL.
	rsp := response.To(req, response.DefaultTTL)

	// Initialize an empty Input to fill in.
	input := &inputv1beta1.Input{}
	// Parse the function input.
	if err := request.GetInput(req, input); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	// Get the observed composed resources from the request.
	observed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed resources from %T", req))
		return rsp, nil
	}

	// Get the desired composed resources (the child resources) from the request.
	// desired = what the functions before it in the pipeline have already asked for.
	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired resources from %T", req))
		return rsp, nil
	}

	// Get the observed composite resource (the parent resource - the XR itself) from the request.
	xr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	// Convert the observed composite resource to the strongly typed XTenantArgo struct.
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		xr.Resource.UnstructuredContent(),
		&xd,
	); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot convert composite resource to %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-version", xr.Resource.GetAPIVersion(),
		"xr-kind", xr.Resource.GetKind(),
		"xr-name", xr.Resource.GetName(),
	)

	// Build the function context that will be passed to the buildComposers function.
	fnContext := resources.XContext{
		Observed:         observed,
		FunctionResponse: rsp,
		XR:               &xd,
		Defaults:         input,
		Log:              log,
	}

	// Build the composers that will generate the desired child resources.
	composers, err := buildComposers(fnContext)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot initialize resources"))
		return rsp, nil
	}

	// Iterate over the composers and generate the desired child resources.
	// If the desired resource is nil, skip it.
	// If the desired resource is not ready, mark the condition as unavailable.
	for _, r := range composers {
		desiredResource, err := r.ComposeDesiredResource()
		if err != nil {
			response.ConditionFalse(rsp, r.GetConditionType(), "CompositionError").
				WithMessage(err.Error()).
				TargetComposite()
			return rsp, nil
		}

		if desiredResource == nil {
			continue
		}

		if r.IsReady() {
			desiredResource.Resource.Ready = resource.ReadyTrue
			response.ConditionTrue(rsp, r.GetConditionType(), "Available").
				TargetComposite()
		} else {
			response.ConditionFalse(rsp, r.GetConditionType(), "Unavailable").
				WithMessage(fmt.Sprintf("%s is not yet available", r.GetConditionType())).
				TargetComposite()
		}

		log.Info("Added desired resource", "name", desiredResource.Name)
		desired[desiredResource.Name] = desiredResource.Resource
	}

	// Set the desired composed resources in the function response.
	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	return rsp, nil
}
