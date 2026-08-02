package resources

import (
	"strings"

	"github.com/google/uuid"
)

// pascal converts a string to PascalCase by splitting on '-', '_', or ' '
// and capitalizing the first letter of each resulting part.
func pascal(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
	}
	return strings.Join(parts, "")
}

// generateRoleUID returns a deterministic UUIDv5 (SHA-1, DNS namespace) for the
// given string, so the same input always produces the same role UID.
//
// Contract: each app-role/role-assignment pair derives its role UID from the
// same seed via this function and they MUST stay in sync — the assignment
// references the AppRole by UID, so any change to the derivation (input,
// namespace, algorithm) must be made in lockstep across both call sites or the
// assignment will point at a non-existent role. Distinct apps use distinct
// seeds so their per-tenant roles don't collide:
//   - ArgoCD:  seed = xr.GetName()            (app-reg-role-argocd.go, role-assignment-argocd-sso-role.go)
//   - Vault:   seed = "vault-" + xr.GetName() (app-reg-role-vault.go,   role-assignment-vault-sso-role.go)
func generateRoleUID(s string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(s)).String()
}
