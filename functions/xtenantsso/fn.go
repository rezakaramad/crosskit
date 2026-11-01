package main

import (
	"context"
	"fmt"

	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
	"github.com/rezakaramad/crosskit/functions/xtenantsso/resources"
	"github.com/rezakaramad/crosskit/modules/composer"
	xtenantssov1beta1 "github.com/rezakaramad/crosskit/types/xtenantsso/v1beta1"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

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

// init registers the AzureAD types with the composed resource scheme so they can be
// marshalled during reconciliation; a failure is a programming error, so it panics.
// See [Schema Registration](../SchemeRegistration.md).
func init() {
	utilruntime.Must(appv1beta1.AddToScheme(composed.Scheme))
	utilruntime.Must(applicationsv1beta1.AddToScheme(composed.Scheme))
	utilruntime.Must(groupsv1beta1.AddToScheme(composed.Scheme))
}

// buildComposers constructs the Group, AppRole and RoleAssignment composers for every app in the input.
func buildComposers(fnContext resources.XContext) ([]composer.ComposableResource, error) {
	var composers []composer.ComposableResource

	for _, app := range fnContext.Defaults.Apps {
		group, err := resources.NewGroup(fnContext, app)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create Group for %s", app.Name)
		}

		role, err := resources.NewAppRole(fnContext, app)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create AppRole for %s", app.Name)
		}

		assignment, err := resources.NewRoleAssignment(fnContext, app)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create RoleAssignment for %s", app.Name)
		}

		composers = append(composers, group, role, assignment)
	}

	return composers, nil
}

// RunFunction is the entry point for the composition function.
// Crossplane passes everything the function needs to run in a RunFunctionRequest struct.
// The function tells Crossplane what resources it should compose by returning a RunFunctionResponse struct.
func (f *Function) RunFunction(
	_ context.Context,
	req *fnv1.RunFunctionRequest,
) (*fnv1.RunFunctionResponse, error) {
	var xd xtenantssov1beta1.XTenantSSO

	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	// Create an initially empty function response, initialized with the request's metadata and a default TTL.
	rsp := response.To(req, response.DefaultTTL)

	// Initialize an empty Input to fill in.
	input := &inputv1beta1.Input{}
	// Parse the function input.
	if err := request.GetInput(req, input); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	// Get the observed composed resources (the child resources) from the request.
	observed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed resources from %T", req))
		return rsp, nil
	}

	// Get the desired composed resources (the child resources) from the request.
	// desired = what earlier functions in the pipeline have already requested.
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

	// Convert the observed composite resource (the XR itself) to the strongly typed XTenantSSO struct.
	// xr - Right before this, we have retrieved the Crossplane's unstructured representation of
	// the composite resource.
	// xd - Now we convert it to the strongly typed Go representation.
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
