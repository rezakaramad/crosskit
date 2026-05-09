# xtenant-validate Architecture

This function validates an observed `XTenant`, checks DNS availability through the configured DNS provider, and gates the tenant on approval before the next pipeline step proceeds.

## Flow Diagram

```mermaid
flowchart LR
    A[main.go\nCLI.Run] --> B[function.Serve]
    A --> K[Kubernetes client]
    B --> C[fn.go\nRunFunction]

    C --> D[Load observed XR\nand function input]
    D --> E[buildDNSClient]
    E --> E1[PowerDNS client]
    E --> E2[Cloud DNS client]
    E --> E3[readSecretKey\nfor PowerDNS]

    C --> F[validate.go\nValidate]
    F --> G[BuildFQDN]
    F --> H[dns.go\nDNSClient.CheckDNSAvailable]
    H --> E1
    H --> E2

    C --> I[approve.go\nIsApproved]
    I --> J[status.go\nPendingApproval]
    I --> L[status.go\nProvisioning]
    C --> M[fn.go\nresponse helpers]
```

## File Roles

- `main.go`: bootstraps the function process, creates the Kubernetes client, and starts the gRPC server.
- `fn.go`: orchestration entry point. It reads the XR and function input, resolves the DNS client, invokes validation, sets XR status, and applies the approval gate.
- `validate.go`: pure validation logic. It builds one FQDN per workload cluster and asks the `DNSClient` whether each name is available.
- `dns.go`: the provider-neutral contract. `Validate` depends on this interface instead of provider-specific clients.
- `pdns_client.go`: PowerDNS implementation of `DNSClient`. It derives a zone from the FQDN, queries the PowerDNS zone endpoint, and inspects `rrsets`.
- `gcp_dns_client.go`: Cloud DNS implementation of `DNSClient`. It discovers a matching managed zone in the configured GCP project and scans record sets for an exact FQDN match.
- `approve.go`: encapsulates the tenant approval check.
- `status.go`: writes `status.phase` back onto the XR.
- `input/v1beta1/input.go`: defines the function input schema used by the Composition pipeline step.

## Who Calls Whom

1. `main.go` creates `Function` and hands it to `function.Serve`.
2. Crossplane calls `Function.RunFunction` in `fn.go`.
3. `RunFunction` parses the observed XR with `fromObservedXR` and parses pipeline input with `request.GetInput`.
4. `RunFunction` calls `buildDNSClient` to choose PowerDNS or Cloud DNS from `input.DNS.provider`.
5. `buildDNSClient` may call `readSecretKey` to fetch the PowerDNS API key from a Kubernetes `Secret` on every reconcile.
6. `RunFunction` calls `Validate`, passing the `DNSClient`, base domain, and workload clusters.
7. `Validate` calls `BuildFQDN`, then `DNSClient.CheckDNSAvailable` for each cluster-specific hostname.
8. The selected provider implementation performs the external DNS lookup and returns `DNSAvailabilityResult`.
9. Back in `RunFunction`, validation failures update XR conditions and phase; successful validation moves to `IsApproved`.
10. If not approved, the function sets `PendingApproval`; if approved, it sets `Provisioning` and returns control to the next composition step.

## Key Design Boundaries

- `fn.go` owns orchestration and Crossplane request/response handling.
- `validate.go` owns the policy decision: is the tenant DNS-safe to provision?
- `dns.go` separates provider-agnostic validation from provider-specific API code.
- Provider implementations own only remote lookup behavior; they do not mutate XR state.
