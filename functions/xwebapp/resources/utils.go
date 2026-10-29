package resources

import (
	"fmt"
	"maps"

	"github.com/rezakaramad/crosskit/types/xwebapp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	nameLabel = "app.kubernetes.io/name"

	// httpPortName is the container port the HTTPRoute sends traffic to.
	httpPortName = "http"
)

// selectorLabels are the immutable labels that tie the Deployment and Service to the pods.
func selectorLabels(xr *xwebapp.XWebApp) map[string]string {
	return map[string]string{nameLabel: xr.GetName()}
}

// mergeLabels returns base overlaid with platform; the platform value wins on a duplicate key.
func mergeLabels(base, platform map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(platform))
	maps.Copy(out, base)
	maps.Copy(out, platform)
	return out
}

// Validate checks the rules the XRD schema cannot express; a returned error must stop the composition.
func Validate(xr *xwebapp.XWebApp) error {
	pod := xr.Spec.Template.Spec
	if len(pod.Containers) == 0 {
		return fmt.Errorf("spec.template.spec.containers must have at least one container")
	}

	seen := map[string]string{}
	check := func(path string, containers []corev1.Container) error {
		for i, c := range containers {
			for j, p := range c.Ports {
				where := fmt.Sprintf("%s[%d].ports[%d]", path, i, j)
				if p.Name == "" {
					return fmt.Errorf("%s: port %d must be named (e.g. %q), names are used by NetworkPolicies", where, p.ContainerPort, httpPortName)
				}
				if errs := validation.IsValidPortName(p.Name); len(errs) > 0 {
					return fmt.Errorf("%s: invalid port name %q: %s", where, p.Name, errs[0])
				}
				if prev, dup := seen[p.Name]; dup {
					return fmt.Errorf("%s: port name %q is already used at %s", where, p.Name, prev)
				}
				seen[p.Name] = where
			}
		}
		return nil
	}
	if err := check("spec.template.spec.initContainers", pod.InitContainers); err != nil {
		return err
	}
	if err := check("spec.template.spec.containers", pod.Containers); err != nil {
		return err
	}

	if hostname(xr) != "" && httpPort(xr) == nil {
		return fmt.Errorf("platformCapabilities.hostname requires a container port named %q", httpPortName)
	}
	return nil
}

func hostname(xr *xwebapp.XWebApp) string {
	if xr.Spec.PlatformCapabilities == nil {
		return ""
	}
	return xr.Spec.PlatformCapabilities.Hostname
}

// httpPort returns the container port named "http", or nil if there is none.
func httpPort(xr *xwebapp.XWebApp) *corev1.ContainerPort {
	for _, c := range xr.Spec.Template.Spec.Containers {
		for _, p := range c.Ports {
			if p.Name == httpPortName {
				return &p
			}
		}
	}
	return nil
}
