package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/modules/composer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/function-sdk-go/resource"
)

// ServiceAccount composes the ServiceAccount the pods run as.
type ServiceAccount struct {
	XComposer
	ObservedResource *corev1.ServiceAccount
}

// NewServiceAccount creates a ServiceAccount composer and deserializes any observed ServiceAccount.
func NewServiceAccount(f XContext) (*ServiceAccount, error) {
	resourceName := resource.Name(fmt.Sprintf("xwebapp-serviceaccount-%s", f.XR.Name))
	observed, err := composer.ConvertObserved[corev1.ServiceAccount](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &ServiceAccount{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "ServiceAccountReady",
		},
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired ServiceAccount.
func (s *ServiceAccount) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return s.ComposeDesiredResourceFrom(s.createResource())
}

// IsReady returns true once the API server has created the ServiceAccount.
func (s *ServiceAccount) IsReady() bool {
	return s.ObservedResource != nil
}

func (s *ServiceAccount) createResource() *corev1.ServiceAccount {
	xr := s.FunctionContext.XR

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:        xr.GetName(),
			Labels:      mergeLabels(xr.GetLabels(), selectorLabels(xr)),
			Annotations: xr.GetAnnotations(),
		},
	}
}
