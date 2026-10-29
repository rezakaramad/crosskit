package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/modules/composer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/crossplane/function-sdk-go/resource"
)

// Service composes a ClusterIP Service exposing every named container port.
type Service struct {
	XComposer
	ObservedResource *corev1.Service
}

// NewService creates a Service composer and deserializes any observed Service.
func NewService(f XContext) (*Service, error) {
	resourceName := resource.Name(fmt.Sprintf("xwebapp-service-%s", f.XR.Name))
	observed, err := composer.ConvertObserved[corev1.Service](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &Service{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ServiceReady",
		},
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired Service, or nil when no container declares a port.
func (s *Service) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return s.ComposeDesiredResourceFrom(s.createResource())
}

// IsReady returns true once the Service has a ClusterIP.
func (s *Service) IsReady() bool {
	return s.ObservedResource != nil && s.ObservedResource.Spec.ClusterIP != ""
}

func (s *Service) createResource() *corev1.Service {
	xr := s.FunctionContext.XR

	var ports []corev1.ServicePort
	for _, c := range xr.Spec.Template.Spec.Containers {
		for _, p := range c.Ports {
			ports = append(ports, corev1.ServicePort{
				Name:       p.Name,
				Protocol:   p.Protocol,
				Port:       p.ContainerPort,
				TargetPort: intstr.FromString(p.Name),
			})
		}
	}
	if len(ports) == 0 {
		s.FunctionContext.Log.Debug("no container ports, skipping service", "name", s.ResourceName)
		return nil
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        xr.GetName(),
			Labels:      mergeLabels(xr.GetLabels(), selectorLabels(xr)),
			Annotations: xr.GetAnnotations(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: selectorLabels(xr),
			Ports:    ports,
		},
	}
}
