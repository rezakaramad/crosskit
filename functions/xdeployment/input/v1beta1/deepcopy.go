package v1beta1

import "k8s.io/apimachinery/pkg/runtime"

// DeepCopyObject returns a deep copy of the Input as a runtime.Object.
func (in *Input) DeepCopyObject() runtime.Object {
	out := *in
	return &out
}
