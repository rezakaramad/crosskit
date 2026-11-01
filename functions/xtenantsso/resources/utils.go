package resources

import (
	"strings"

	"github.com/google/uuid"
	inputv1beta1 "github.com/rezakaramad/crosskit/functions/xtenantsso/input/v1beta1"
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
func generateRoleUID(s string) string {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(s)).String()
}

// roleUID is the UID of an app's role for a tenant. The AppRole sets it and the
// RoleAssignment refers to it, so both must derive it here. Role IDs are immutable
// in Entra, so changing the seed would make the provider refuse the update.
func roleUID(app inputv1beta1.AppConfig, tenant string) string {
	return generateRoleUID(app.Name + "-" + tenant)
}
