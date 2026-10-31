package main

import (
	"context"
	"fmt"

	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xwebapp/input/v1beta1"
	"github.com/rezakaramad/crosskit/functions/xwebapp/resources"
	"github.com/rezakaramad/crosskit/modules/composer"
	"github.com/rezakaramad/crosskit/types/xwebapp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function is the gRPC server that Crossplane calls to render an XWebApp.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// init registers the composed resource types; a failure is a programming error, so it panics.
func init() {
	utilruntime.Must(appsv1.AddToScheme(composed.Scheme))
	utilruntime.Must(corev1.AddToScheme(composed.Scheme))
	utilruntime.Must(gwapiv1.Install(composed.Scheme))
}

// buildComposers constructs the composer for every child resource of an XWebApp.
func buildComposers(fnContext resources.XContext) ([]composer.ComposableResource, error) {
	serviceAccount, err := resources.NewServiceAccount(fnContext)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create ServiceAccount")
	}
	deployment, err := resources.NewDeployment(fnContext)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Deployment")
	}
	service, err := resources.NewService(fnContext)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Service")
	}
	httpRoute, err := resources.NewHTTPRoute(fnContext)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create HTTPRoute")
	}

	// One composer per child; for an XR named "foo" the resource names are
	// xwebapp-{serviceaccount,deployment,service,httproute}-foo.
	return []composer.ComposableResource{serviceAccount, deployment, service, httpRoute}, nil
}

// RunFunction composes the child resources of the observed XWebApp.
func (f *Function) RunFunction(_ context.Context, req *fnv1.RunFunctionRequest) (*fnv1.RunFunctionResponse, error) {
	var xw xwebapp.XWebApp

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

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(xr.Resource.UnstructuredContent(), &xw); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot convert composite resource to %s", xr.Resource.GetKind()))
		return rsp, nil
	}

	log := f.log.WithValues(
		"xr-version", xr.Resource.GetAPIVersion(),
		"xr-kind", xr.Resource.GetKind(),
		"xr-name", xr.Resource.GetName(),
	)

	if err := resources.Validate(&xw); err != nil {
		response.ConditionFalse(rsp, "FunctionSuccess", "InvalidSpec").
			WithMessage(err.Error()).
			TargetCompositeAndClaim()
		response.Fatal(rsp, errors.Wrap(err, "invalid XWebApp"))
		return rsp, nil
	}

	composers, err := buildComposers(resources.XContext{
		Observed:         observed,
		FunctionResponse: rsp,
		XR:               &xw,
		Defaults:         input,
		Log:              log,
	})
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot initialize resources"))
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
			response.ConditionTrue(rsp, r.GetConditionType(), "Available").TargetComposite()
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
