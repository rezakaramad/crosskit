package resources

import (
	"fmt"

	"github.com/rezakaramad/crosskit/modules/composer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/crossplane/function-sdk-go/resource"
)

const defaultReplicas = int32(2)

// Deployment composes the Deployment from the developer's pod template plus platform-owned fields.
type Deployment struct {
	XComposer
	ObservedResource *appsv1.Deployment
}

// NewDeployment creates a Deployment composer and deserializes any observed Deployment.
func NewDeployment(f XContext) (*Deployment, error) {
	resourceName := resource.Name(fmt.Sprintf("xwebapp-deployment-%s", f.XR.Name))
	observed, err := composer.ConvertObserved[appsv1.Deployment](f.Observed, resourceName)
	if err != nil {
		return nil, err
	}

	return &Deployment{
		XComposer: XComposer{
			FunctionContext: f,
			ResourceName:    resourceName,
			ConditionType:   "DeploymentReady",
		},
		ObservedResource: observed,
	}, nil
}

// ComposeDesiredResource builds the desired Deployment.
func (d *Deployment) ComposeDesiredResource() (*composer.DesiredResource, error) {
	return d.ComposeDesiredResourceFrom(d.createResource())
}

// IsReady returns true when the Deployment reports the Available condition.
func (d *Deployment) IsReady() bool {
	if d.ObservedResource == nil {
		return false
	}
	for _, c := range d.ObservedResource.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// createResource deep-copies the developer's template and overwrites the platform-owned fields;
// the API server validates everything else.
func (d *Deployment) createResource() *appsv1.Deployment {
	xr := d.FunctionContext.XR
	defaults := d.FunctionContext.Defaults

	tpl := xr.Spec.Template.DeepCopy()
	tpl.Labels = mergeLabels(tpl.Labels, selectorLabels(xr))
	tpl.Spec.ServiceAccountName = xr.GetName()
	tpl.Spec.Tolerations = defaults.Tolerations
	tpl.Spec.TopologySpreadConstraints = []corev1.TopologySpreadConstraint{{
		MaxSkew:           1,
		TopologyKey:       corev1.LabelTopologyZone,
		WhenUnsatisfiable: corev1.ScheduleAnyway,
		LabelSelector:     &metav1.LabelSelector{MatchLabels: selectorLabels(xr)},
	}}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        xr.GetName(),
			Labels:      mergeLabels(xr.GetLabels(), selectorLabels(xr)),
			Annotations: xr.GetAnnotations(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(ptr.Deref(xr.Spec.Replicas, defaultReplicas)),
			Strategy: *xr.Spec.Strategy.DeepCopy(),
			Selector: &metav1.LabelSelector{MatchLabels: selectorLabels(xr)},
			Template: *tpl,
		},
	}
}
