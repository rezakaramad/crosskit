package main

import (
	"context"
	"fmt"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantentra/input/v1beta1"
	"github.com/rezakaramad/crosskit/functions/xtenantentra/resources"
	"github.com/rezakaramad/crosskit/modules/composer"
	"github.com/rezakaramad/crosskit/types/xtenantentra"
	appv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/app/v1beta1"
	applicationsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/applications/v1beta1"
	groupsv1beta1 "github.com/upbound/provider-azuread/v2/apis/namespaced/groups/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Function is the gRPC server that Crossplane calls to render tenant resources.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer
	log logging.Logger
}

func init() {
	must := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	must(appv1beta1.AddToScheme(composed.Scheme))
	must(applicationsv1beta1.AddToScheme(composed.Scheme))
	must(groupsv1beta1.AddToScheme(composed.Scheme))
}

// InternalErrorResponse marks the function response as fatally failed due to an internal error.
func InternalErrorResponse(rsp *fnv1.RunFunctionResponse, err error) {
	response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
		TargetCompositeAndClaim()

	response.Fatal(rsp, err)
}

// initResources constructs all resource composers for the XExample
// composition, returning the full set of resources to compose during reconciliation.
func initResources(fnContext resources.XContext) ([]composer.ComposableResource, error) {
	argocdGroup, err := resources.NewArgoCDGroup(fnContext)
	if err != nil {
		return nil, err
	}
	argocdAppRegRole, err := resources.NewArgoCDAppRole(fnContext)
	if err != nil {
		return nil, err
	}
	argocdSsoRoleAssignment, err := resources.NewArgoCDSsoRoleAssignment(fnContext)
	if err != nil {
		return nil, err
	}
	return []composer.ComposableResource{
		argocdGroup,
		argocdAppRegRole,
		argocdSsoRoleAssignment,
	}, nil
}

// RunFunction is the entry point for the composition function.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	var xd xtenantentra.XTenantEntra

	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	input := &inputv1beta1.Input{}
	if err := request.GetInput(req, input); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	observed, err := request.GetObservedComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed resources from %T", req))
		return rsp, nil
	}

	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired resources from %T", req))
		return rsp, nil
	}

	xr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(xr.Resource.UnstructuredContent(), &xd); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot convert composite resource to %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-version", xr.Resource.GetAPIVersion(),
		"xr-kind", xr.Resource.GetKind(),
		"xr-name", xr.Resource.GetName(),
	)

	fnContext := resources.XContext{
		Observed:         observed,
		FunctionResponse: rsp,
		XR:               &xd,
		Defaults:         input,
		Log:              log,
	}

	composers, err := initResources(fnContext)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot initialize resources"))
		return rsp, nil
	}

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

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources in %T", rsp))
		return rsp, nil
	}

	return rsp, nil
}
