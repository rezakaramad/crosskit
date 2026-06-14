package main

import (
	"context"
	"os"

	inputv1beta1 "github.com/rezakaramad/crossplane-toolkit/functions/xdeployment/input/v1beta1"
	render "github.com/rezakaramad/crossplane-toolkit/functions/xdeployment/internal"
	"github.com/rezakaramad/crossplane-toolkit/modules/nextinsight"
	xdeployment "github.com/rezakaramad/crossplane-toolkit/types/xdeployment"
	"k8s.io/apimachinery/pkg/runtime"

	xperrors "github.com/crossplane/crossplane-runtime/v2/pkg/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1 "github.com/crossplane/function-sdk-go/proto/v1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

// Function is the gRPC server that Crossplane calls to render deployment resources.
type Function struct {
	fnv1.UnimplementedFunctionRunnerServiceServer

	log         logging.Logger
	nextInsight nextinsight.Client
}

// newFunction builds the Function with all external clients initialised.
func newFunction(log logging.Logger) *Function {
	return &Function{
		log:         log,
		nextInsight: newNextInsightClient(),
	}
}

// newNextInsightClient returns a configured Next-Insight client when
// NEXTINSIGHT_BASE_URL is set, or nil to skip metadata enrichment.
func newNextInsightClient() nextinsight.Client {
	baseURL := os.Getenv("NEXTINSIGHT_BASE_URL")
	if baseURL == "" {
		return nil
	}
	return nextinsight.New(baseURL, os.Getenv("NEXTINSIGHT_TOKEN"))
}

func (f *Function) RunFunction(
	ctx context.Context,
	req *fnv1.RunFunctionRequest,
) (*fnv1.RunFunctionResponse, error) {
	log := f.log.WithValues("tag", req.GetMeta().GetTag())
	log.Info("Running function-xdeployment")

	rsp := response.To(req, response.DefaultTTL)

	// ---------------------------------------------------------------------
	// 1. Load XR
	// ---------------------------------------------------------------------
	observedXR, err := request.GetObservedCompositeResource(req)
	if err != nil {
		return fatal(rsp, err, "cannot get observed composite resource")
	}
	if observedXR == nil || observedXR.Resource == nil || len(observedXR.Resource.UnstructuredContent()) == 0 {
		return fatal(rsp, nil, "missing observed composite resource")
	}

	// ---------------------------------------------------------------------
	// 2. Parse XR into XDeployment
	// ---------------------------------------------------------------------
	var xd xdeployment.XDeployment
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
		observedXR.Resource.UnstructuredContent(), &xd,
	); err != nil {
		return fatal(rsp, err, "cannot convert XR to XDeployment")
	}

	// ---------------------------------------------------------------------
	// 3. Parse function input
	// ---------------------------------------------------------------------
	var input inputv1beta1.Input
	if err := request.GetInput(req, &input); err != nil {
		return fatal(rsp, err, "cannot parse function input")
	}

	// ---------------------------------------------------------------------
	// 4. Build Deployment
	// ---------------------------------------------------------------------
	deployment := render.BuildDeployment(
		xd.GetName(),
		xd.Spec.Namespace,
		xd.Spec.Image,
		xd.Spec.Replicas,
	)

	// ---------------------------------------------------------------------
	// 5. Enrich with Next-Insight application metadata labels (optional)
	// ---------------------------------------------------------------------
	appLabels, err := render.FetchApplicationLabels(ctx, f.nextInsight, xd.Spec.AppID, input.NextInsight.LabelPrefix)
	if err != nil {
		// Non-fatal: log and continue — metadata enrichment must not block provisioning.
		log.Info("Skipping Next-Insight label enrichment", "error", err)
		appLabels = map[string]string{}
	}

	render.ApplyLabels(appLabels, deployment)

	// ---------------------------------------------------------------------
	// 6. Set desired resources
	// ---------------------------------------------------------------------
	desired, err := request.GetDesiredComposedResources(req)
	if err != nil {
		return fatal(rsp, err, "cannot get desired composed resources")
	}

	desired[resource.Name(xd.GetName()+"-deployment")] = &resource.DesiredComposed{Resource: deployment}

	if err := response.SetDesiredComposedResources(rsp, desired); err != nil {
		return fatal(rsp, err, "cannot set desired composed resources")
	}

	return rsp, nil
}

// fatal marks the function response as fatally failed due to an internal error.
// It sets a FunctionSuccess=False condition with reason InternalError on both
// the composite resource and claim, then records the error as fatal to stop
// the composition pipeline.
func fatal(rsp *fnv1.RunFunctionResponse, err error, msg string) (*fnv1.RunFunctionResponse, error) {
	if err != nil {
		err = xperrors.Wrap(err, msg)
	} else {
		err = xperrors.New(msg)
	}
	response.ConditionFalse(rsp, "FunctionSuccess", "InternalError").
		WithMessage(err.Error()).
		TargetCompositeAndClaim()
	response.Fatal(rsp, err)
	return rsp, nil
}
