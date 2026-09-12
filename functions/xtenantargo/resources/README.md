# Argo tenant resources

This package builds the Argo CD resources composed for each `XTenantArgo`.
It currently contains one child-resource composer: `ArgoCDApplicationSet`.

## Resource

| Resource | Kind | Purpose |
| --- | --- | --- |
| `ArgoCDApplicationSet` | `ApplicationSet` | Deploys the tenant's platform applications to the management cluster and selected workload clusters. |

The generated `ApplicationSet`:

- Uses the tenant name as its Kubernetes object name.
- Is created in the ApplicationSet namespace from the function input defaults.
- Creates a management-cluster application from the configured management repository and path.
- Creates workload-cluster applications for clusters matching the configured cluster-type and environment labels.
- Passes the tenant name and `XTenantArgo.Spec.ShortName` to Helm as `tenant.name` and `tenant.tenantShortName`.
- Applies the configured repository revision, Helm value files, and Helm parameters to both application sources.

## Reconciliation

The composer looks up the previously observed child using the internal name
`applicationset-argocd-{tenant}`. It composes the desired `ApplicationSet` on
every reconcile. The resource is considered ready when its observed
`ResourcesUpToDate` condition is `True`.

The Kubernetes object name is simply `{tenant}`. The internal Crossplane
composition name and the Kubernetes object name are separate values.
