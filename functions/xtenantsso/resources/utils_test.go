package resources

import (
	"testing"
)

func TestPascal(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"pillow", "Pillow"},
		{"pillow-factory", "PillowFactory"},
		{"PILLOW", "Pillow"},
		{"pillow_factory_co", "PillowFactoryCo"},
		{"pillow factory", "PillowFactory"},
		{"pillow-factory_co inc", "PillowFactoryCoInc"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := pascal(tt.in); got != tt.want {
				t.Errorf("pascal(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestGenerateRoleUID_Deterministic locks in the contract from utils.go: the
// ArgoCD AppRole and the ArgoCD SSO RoleAssignment derive the AppRole UID from
// the same XR name via this helper, so the same input must always produce the
// same output.
func TestGenerateRoleUID_Deterministic(t *testing.T) {
	if a, b := generateRoleUID("pillow-factory"), generateRoleUID("pillow-factory"); a != b {
		t.Errorf("generateRoleUID not deterministic: %q vs %q", a, b)
	}
	if a, b := generateRoleUID("pillow-factory"), generateRoleUID("other-tenant"); a == b {
		t.Errorf("generateRoleUID collision for different inputs: %q", a)
	}
}
