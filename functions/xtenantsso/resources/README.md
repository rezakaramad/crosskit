# Entra ID tenant SSO resources

This package builds the Azure resources composed for each `XTenantSSO`.
The function input lists the apps (`apps`); every resource below is composed once per app,
scoped to the tenant name, and uses the provider configuration from the input.

## Resources

| Composer | AzureAD kind | Purpose |
| --- | --- | --- |
| `Group` | `Group` | Creates the tenant's ACL group for the app. |
| `AppRole` | `AppRole` | Creates the tenant-specific role on the app's registration. |
| `RoleAssignment` | `RoleAssignment` | Assigns the role to the tenant's group on the app's enterprise application. |

For each app:

```text
Group -> RoleAssignment
AppRole ----^                 (the assignment uses the AppRole's ID)
```

The role assignment also needs the group's observed Azure `ObjectID`. It is therefore omitted from
the desired response until the group has been reconciled and its ID is available; a later reconcile
composes it.

## Generated names

Kubernetes names include the app name and the tenant name:

| Resource | Name pattern |
| --- | --- |
| ACL group | `acl-plt-{app}-tenant-{tenant}` |
| AppRole | `app-role-{app}-{tenant}` |
| Role assignment | `acl-plt-{app}-sso-role-{tenant}` |

The internal Crossplane composition names are separate from the Kubernetes names:
`group-{app}-{tenant}`, `app-role-{app}-{tenant}` and `roleassignment-{app}-sso-role-{tenant}`.
Condition types use the app's display name, for example `ArgoCDGroupReady`.

## Role IDs

The AppRole's ID is a UUIDv5 of `{app}-{tenant}`, so it is stable. The role assignment derives the
same value. Role IDs are immutable in Entra, so changing the derivation makes the provider refuse the update.

Removing an app from the input removes its three resources, and with them the group and role in Entra.
