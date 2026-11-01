# Argo tenant resources

This package builds the Argo CD resources composed for each `XTenant`.

| Composer | Kind | Internal name | Purpose |
| --- | --- | --- | --- |
| `ArgoCDApplication` (sso) | `Application` | `xtenant-application-sso-{tenant}` | Deploys the `tenant-sso` chart to the management cluster. |
| `ArgoCDApplication` (management) | `Application` | `xtenant-application-management-{tenant}` | Deploys the `tenant-management` chart to the management cluster. |
| `ArgoCDApplicationSet` | `ApplicationSet` | `xtenant-applicationset-workload-{tenant}` | Deploys the `tenant-workload` chart to the workload clusters. |

All of them:

- Are created in the Argo namespace from the function input.
- Pass the tenant name and `XTenant.Spec.ShortName` to Helm as `tenant.name` and `tenant.tenantShortName`.
- Apply the configured repository revision, Helm value files, and Helm parameters.

The Applications use `{tenant}-sso` and `{tenant}-management` as their Kubernetes names, carry the
cascade finalizer, and sync automatically with self-heal (no prune), because readiness requires `Synced`.

The `ApplicationSet` is named `{tenant}-workload` and has a single `clusters` generator that creates
`{tenant}-workload-{env}` for every cluster matching the configured cluster-type and environment labels.

## Reconciliation

Each composer looks up its previously observed child by internal name and composes the desired
resource on every reconcile. The internal Crossplane name and the Kubernetes object name are separate values.

An `Application` is ready when Argo CD reports `health=Healthy` and `sync=Synced`. The `ApplicationSet` is
ready when its `ResourcesUpToDate` condition is `True`.
