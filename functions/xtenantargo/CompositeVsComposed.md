flowchart TD
    XR["Composite Resource (XR)<br/>XTenantArgo — ONE<br/>GetObservedCompositeResource"] --> C1["Composed: ApplicationSet"]
    XR --> C2["Composed: (other children…)"]
    C1 -.->|part of| OBS["GetObservedComposedResources<br/>MANY (map)"]
    C2 -.->|part of| OBS
