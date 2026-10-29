package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/modules/composer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// HTTPRoute composes the Gateway API HTTPRoute that exposes the "http" port on the requested hostname.
type HTTPRoute struct {
	XComposer
	ObservedResource *gwapiv1.HTTPRoute
}

// NewHTTPRoute creates an HTTPRoute composer and deserializes any observed HTTPRoute.
func NewHTTPRoute(f XContext) (*HTTPRoute, error) {
	resourceName := resource.Name(fmt.Sprintf("xwebapp-httproute-%s", f.XR.Name))
	observed, err := composer.ConvertObserved[gwapiv1.HTTPRoute](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &HTTPRoute{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "HTTPRouteReady",
		},
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired HTTPRoute, or nil when no hostname is requested.
func (h *HTTPRoute) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return h.ComposeDesiredResourceFrom(h.createResource())
}

// IsReady returns true once a parent Gateway has accepted the route.
func (h *HTTPRoute) IsReady() bool {
	if h.ObservedResource == nil {
		return false
	}
	for _, parent := range h.ObservedResource.Status.Parents {
		for _, c := range parent.Conditions {
			if c.Type == string(gwapiv1.RouteConditionAccepted) && c.Status == metav1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func (h *HTTPRoute) createResource() *gwapiv1.HTTPRoute {
	xr := h.FunctionContext.XR
	gateway := h.FunctionContext.Defaults.Gateway

	host := hostname(xr)
	port := httpPort(xr)
	if host == "" || port == nil {
		h.FunctionContext.Log.Debug("no hostname requested, skipping httproute", "name", h.ResourceName)
		return nil
	}

	gatewayName, gatewayNamespace := gateway.Name, gateway.Namespace
	if gatewayName == "" {
		gatewayName = xr.GetNamespace()
	}
	if gatewayNamespace == "" {
		gatewayNamespace = xr.GetNamespace()
	}

	return &gwapiv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:        xr.GetName(),
			Labels:      mergeLabels(xr.GetLabels(), selectorLabels(xr)),
			Annotations: xr.GetAnnotations(),
		},
		Spec: gwapiv1.HTTPRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{{
					Name:      gwapiv1.ObjectName(gatewayName),
					Namespace: new(gwapiv1.Namespace(gatewayNamespace)),
				}},
			},
			Hostnames: []gwapiv1.Hostname{gwapiv1.Hostname(host)},
			Rules: []gwapiv1.HTTPRouteRule{{
				BackendRefs: []gwapiv1.HTTPBackendRef{{
					BackendRef: gwapiv1.BackendRef{
						BackendObjectReference: gwapiv1.BackendObjectReference{
							Name: gwapiv1.ObjectName(xr.GetName()),
							Port: new(gwapiv1.PortNumber(port.ContainerPort)),
						},
					},
				}},
			}},
		},
	}
}
