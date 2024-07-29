# style

- [style](#style)
    - [public, private, internal 등](#public-private-internal-등)

## public, private, internal 등

대문자로 시작할 경우 public

internal로 사용하지만, 그럼에도 대문자로 사용하는 경우 있다

```go
// https://github.com/istio/istio/blob/7e6b0590cbbfc79034594553d2222ef14d3e7dee/cni/pkg/constants/constants.go#L54-L66

// Internal constants
const (
    DefaultKubeconfigMode = 0o600

    CNIAddEventPath = "/cmdadd"
    UDSLogPath      = "/log"

    // K8s liveness and readiness endpoints
    LivenessEndpoint  = "/healthz"
    ReadinessEndpoint = "/readyz"
    ReadinessPort     = "8000"
    NetNsPath         = "/var/run/netns"
)
```
