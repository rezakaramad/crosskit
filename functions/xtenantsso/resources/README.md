# EntraID Tenant resources

This package builds the Azure resources composed for each `XTenantEntra`.
Every resource is scoped to the tenant name and uses the provider configuration
from the function input defaults.

## Resources

| Resource | AzureAD kind | Purpose |
| --- | --- | --- |
| `ArgoCDGroup` | `Group` | Creates the tenant's ArgoCD ACL group. |
| `VaultGroup` | `Group` | Creates the tenant's Vault ACL group. |
| `ArgoCDAppRole` | `AppRole` | Creates the tenant-specific role on the ArgoCD application registration. The corresponding role assignment grants the tenant's ArgoCD group access to the Argo CD UI. |
| `VaultAppRole` | `AppRole` | Creates the tenant-specific role on the Vault application registration. The corresponding role assignment grants the tenant's Vault group access to Vault. |
| `ArgoCDSsoRoleAssignment` | `RoleAssignment` | Assigns the ArgoCD role to the tenant's ArgoCD ACL group on the ArgoCD enterprise application. |
| `VaultSsoRoleAssignment` | `RoleAssignment` | Assigns the Vault role to the tenant's Vault ACL group on the Vault enterprise application. |

The ArgoCD and Vault resources are independent pairs:

```text
Group -> RoleAssignment
AppRole ----^                 (the assignment uses the generated AppRole ID)
```

The role assignments also require the corresponding group's observed Azure
`ObjectID`. They are therefore omitted from the desired response until the
group has been reconciled and its ID is available; a later reconcile composes
the assignment.

## Generated names

Kubernetes resource names include the tenant name:

| Resource | Name pattern |
| --- | --- |
| ACL group | `acl-plt-{argocd\|vault}-tenant-{tenant}` |
| AppRole | `app-role-{argocd\|vault}-{tenant}` |
| SSO role assignment | `acl-plt-{argocd\|vault}-sso-role-{tenant}` |

The corresponding observed-resource lookup names are `group-{service}-{tenant}`
and `roleassignment-{service}-sso-role-{tenant}`. These names are internal
Crossplane composition names and are separate from the Kubernetes object names
listed above.
