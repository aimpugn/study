# IP

## IP 버전 체크

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    ipv4Addr := net.ParseIP("192.0.2.1")
    ipv6Addr := net.ParseIP("2001:db8::68")

    // IPv4 체크
    if ipv4Addr.To4() != nil {
        fmt.Println(ipv4Addr, "is an IPv4 address")
    } else {
        fmt.Println(ipv4Addr, "is not an IPv4 address")
    }

    // IPv6 체크
    if ipv6Addr.To16() != nil && ipv6Addr.To4() == nil {
        fmt.Println(ipv6Addr, "is an IPv6 address")
    } else {
        fmt.Println(ipv6Addr, "is not an IPv6 address")
    }
}
```
