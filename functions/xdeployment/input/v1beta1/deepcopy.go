package v1beta1

import "k8s.io/apimachinery/pkg/runtime"

func (in *Input) DeepCopyObject() runtime.Object {
	out := *in
	return &out
}
