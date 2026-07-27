package argocd

import (
	"maps"

	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSet) DeepCopyInto(out *ApplicationSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a new ApplicationSet by copying the receiver.
func (in *ApplicationSet) DeepCopy() *ApplicationSet {
	if in == nil {
		return nil
	}
	out := new(ApplicationSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object by copying the receiver.
func (in *ApplicationSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetList) DeepCopyInto(out *ApplicationSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ApplicationSet, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

// DeepCopy creates a new ApplicationSetList by copying the receiver.
func (in *ApplicationSetList) DeepCopy() *ApplicationSetList {
	if in == nil {
		return nil
	}
	out := new(ApplicationSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object by copying the receiver.
func (in *ApplicationSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetSpec) DeepCopyInto(out *ApplicationSetSpec) {
	*out = *in
	if in.GoTemplateOptions != nil {
		out.GoTemplateOptions = make([]string, len(in.GoTemplateOptions))
		copy(out.GoTemplateOptions, in.GoTemplateOptions)
	}
	if in.Generators != nil {
		out.Generators = make([]ApplicationSetGenerator, len(in.Generators))
		for i := range in.Generators {
			in.Generators[i].DeepCopyInto(&out.Generators[i])
		}
	}
	in.Template.DeepCopyInto(&out.Template)
	if in.IgnoreApplicationDifferences != nil {
		out.IgnoreApplicationDifferences = make([]ApplicationSetResourceIgnoreDifferences, len(in.IgnoreApplicationDifferences))
		for i := range in.IgnoreApplicationDifferences {
			in.IgnoreApplicationDifferences[i].DeepCopyInto(&out.IgnoreApplicationDifferences[i])
		}
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetGenerator) DeepCopyInto(out *ApplicationSetGenerator) {
	*out = *in
	if in.Matrix != nil {
		out.Matrix = new(MatrixGenerator)
		in.Matrix.DeepCopyInto(out.Matrix)
	}
	if in.List != nil {
		out.List = new(ListGenerator)
		in.List.DeepCopyInto(out.List)
	}
	if in.Git != nil {
		out.Git = new(GitGenerator)
		in.Git.DeepCopyInto(out.Git)
	}
	if in.Clusters != nil {
		out.Clusters = new(ClusterGenerator)
		in.Clusters.DeepCopyInto(out.Clusters)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *MatrixGenerator) DeepCopyInto(out *MatrixGenerator) {
	*out = *in
	if in.Generators != nil {
		out.Generators = make([]ApplicationSetGenerator, len(in.Generators))
		for i := range in.Generators {
			in.Generators[i].DeepCopyInto(&out.Generators[i])
		}
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ListGenerator) DeepCopyInto(out *ListGenerator) {
	*out = *in
	if in.Elements != nil {
		out.Elements = make([]map[string]string, len(in.Elements))
		for i, m := range in.Elements {
			if m != nil {
				out.Elements[i] = maps.Clone(m)
			}
		}
	}
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopyInto copies the receiver into out.
func (in *GitGenerator) DeepCopyInto(out *GitGenerator) {
	*out = *in
	if in.Directories != nil {
		out.Directories = make([]GitDirectoryGeneratorItem, len(in.Directories))
		copy(out.Directories, in.Directories)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ClusterGenerator) DeepCopyInto(out *ClusterGenerator) {
	*out = *in
	in.Selector.DeepCopyInto(&out.Selector)
	in.Template.DeepCopyInto(&out.Template)
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetTemplate) DeepCopyInto(out *ApplicationSetTemplate) {
	*out = *in
	in.ApplicationSetTemplateMeta.DeepCopyInto(&out.ApplicationSetTemplateMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetTemplateMeta) DeepCopyInto(out *ApplicationSetTemplateMeta) {
	*out = *in
	if in.Labels != nil {
		out.Labels = maps.Clone(in.Labels)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSpec) DeepCopyInto(out *ApplicationSpec) {
	*out = *in
	if in.Source != nil {
		out.Source = new(ApplicationSource)
		in.Source.DeepCopyInto(out.Source)
	}
	out.Destination = in.Destination
	if in.SyncPolicy != nil {
		out.SyncPolicy = new(SyncPolicy)
		in.SyncPolicy.DeepCopyInto(out.SyncPolicy)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSource) DeepCopyInto(out *ApplicationSource) {
	*out = *in
	if in.Helm != nil {
		out.Helm = new(HelmSource)
		in.Helm.DeepCopyInto(out.Helm)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *HelmSource) DeepCopyInto(out *HelmSource) {
	*out = *in
	if in.ValueFiles != nil {
		out.ValueFiles = make([]string, len(in.ValueFiles))
		copy(out.ValueFiles, in.ValueFiles)
	}
	if in.Parameters != nil {
		out.Parameters = make([]HelmParameter, len(in.Parameters))
		copy(out.Parameters, in.Parameters)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *SyncPolicy) DeepCopyInto(out *SyncPolicy) {
	*out = *in
	if in.Automated != nil {
		out.Automated = new(SyncPolicyAutomated)
		in.Automated.DeepCopyInto(out.Automated)
	}
	if in.SyncOptions != nil {
		out.SyncOptions = make([]string, len(in.SyncOptions))
		copy(out.SyncOptions, in.SyncOptions)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *SyncPolicyAutomated) DeepCopyInto(out *SyncPolicyAutomated) {
	*out = *in
	if in.Enabled != nil {
		enabled := *in.Enabled
		out.Enabled = &enabled
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetResourceIgnoreDifferences) DeepCopyInto(out *ApplicationSetResourceIgnoreDifferences) {
	*out = *in
	if in.JSONPointers != nil {
		out.JSONPointers = make([]string, len(in.JSONPointers))
		copy(out.JSONPointers, in.JSONPointers)
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetStatus) DeepCopyInto(out *ApplicationSetStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]ApplicationSetCondition, len(in.Conditions))
		copy(out.Conditions, in.Conditions)
	}
	if in.ApplicationStatus != nil {
		out.ApplicationStatus = make([]ApplicationSetApplicationStatus, len(in.ApplicationStatus))
		for i := range in.ApplicationStatus {
			in.ApplicationStatus[i].DeepCopyInto(&out.ApplicationStatus[i])
		}
	}
}

// DeepCopyInto copies the receiver into out.
func (in *ApplicationSetApplicationStatus) DeepCopyInto(out *ApplicationSetApplicationStatus) {
	*out = *in
	if in.TargetRevisions != nil {
		out.TargetRevisions = make([]string, len(in.TargetRevisions))
		copy(out.TargetRevisions, in.TargetRevisions)
	}
}
