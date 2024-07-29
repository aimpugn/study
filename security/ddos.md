# DDoS

- [DDoS](#ddos)
    - [DDoS](#ddos-1)
    - [DDoS 공격의 주요 특징](#ddos-공격의-주요-특징)
    - [DDoS 공격 유형](#ddos-공격-유형)
    - [DDoS 공격의 주요 목적](#ddos-공격의-주요-목적)
    - [DDoS 방어 전략](#ddos-방어-전략)
        - [구체적인 방어 방법](#구체적인-방어-방법)

## DDoS

DDoS(Distributed Denial of Service) 공격은
- 다수의 시스템이
- 특정 서버, 서비스, 네트워크에 과도한 트래픽을 유도하여
- 정상적인 사용자들이 서비스를 이용하지 못하게 하는 공격입니다.

**Denial of Service (DoS)**가 단일 시스템에서 대량의 요청을 보내서 특정 서비스나 서버의 자원을 고갈시켜 서비스를 마비시키는 공격이라면,
**Distributed Denial of Service (DDoS)**는 여러 대의 공격 시스템(보통 봇넷으로 구성됨)이 분산된 형태로 대규모 트래픽을 발생시켜 공격 대상의 시스템을 마비시키는 공격입니다.

이 공격은 분산된 여러 공격 원에서 동시다발적으로 이루어지며, 피해 대상의 자원을 고갈시켜 서비스의 가용성을 저하시킵니다.

## DDoS 공격의 주요 특징

- **분산된 공격원**: DDoS 공격은 수천에서 수십만 대의 공격 원(주로 감염된 컴퓨터, IoT 장치 등)으로부터 트래픽이 발생하므로, 특정 IP를 차단해도 전체 공격을 막기 어렵습니다.
- **대규모 트래픽**: 짧은 시간 내에 대규모 트래픽을 발생시켜 네트워크 대역폭, CPU, 메모리 등의 자원을 고갈시킵니다.
- **다양한 공격 벡터**: DDoS 공격은 네트워크 레벨, 전송 레벨, 애플리케이션 레벨에서 다양한 방식으로 이루어질 수 있습니다.

## DDoS 공격 유형

1. **볼륨 기반 공격 (Volumetric Attacks)**

    네트워크 대역폭을 고갈시키는 것이 목적입니다.
    공격자는 대규모의 트래픽을 만들어 네트워크의 용량을 초과하게 하여 서비스에 접근할 수 없게 만듭니다.

    - **예시**:
        - **UDP Flood**: UDP 프로토콜을 사용하여 대상 서버에 대량의 패킷을 전송합니다.
        - **ICMP Flood (Ping Flood)**: ICMP Echo Request 메시지를 대량으로 전송하여 대상 서버의 대역폭을 고갈시킵니다.
    - **대응 방법**:
        - IP 필터링
        - 레이트 리미팅
        - 트래픽 분석 및 차단

2. **프로토콜 공격 (Protocol Attacks)**

    네트워크 및 전송 계층의 프로토콜을 악용하여 서버 자원을 고갈시킵니다.
    주요 표적은 서버의 연결 상태와 방화벽, 로드밸런서 등의 중간 장치들입니다.

    - **예시**:
        - **SYN Flood**: TCP 3-way handshake를 악용하여, 많은 연결 요청을 보내면서 서버의 자원을 고갈시킵니다.
        - **Ping of Death**: 비정상적으로 큰 Ping 패킷을 보내 대상 시스템을 충돌시키거나 정지시킵니다.
        - **Smurf Attack**: 스푸핑된 IP 주소를 사용하여 다수의 네트워크 장비가 한 서버로 응답하도록 유도해 트래픽 폭탄을 일으킵니다.
    - **대응 방법**:
        - TCP 상태 모니터링
        - 방화벽 설정 강화
        - 공격 패턴 식별 및 차단

3. **애플리케이션 레이어 공격 (Application Layer Attacks)**

    애플리케이션 레벨에서의 취약점을 이용하여 공격합니다.
    서버가 요청을 처리하는 데 많은 자원을 사용하도록 하여, 서버를 과부하 상태로 만듭니다.

    - **예시**:
        - **HTTP Flood**: HTTP GET/POST 요청을 대량으로 보내 웹 서버를 마비시킵니다.
        - **Slowloris**: HTTP 요청을 비정상적으로 느리게 전송하여 서버의 연결 수를 고갈시킵니다.
    - **대응 방법**:
        - 웹 애플리케이션 방화벽(WAF)
        - 캐시 설정
        - 요청 검증 및 필터링

## DDoS 공격의 주요 목적

- **서비스 중단**: 대상 서비스의 가용성을 떨어뜨려 사용자들이 정상적으로 서비스를 이용하지 못하도록 합니다.
- **자원 고갈**: 서버의 CPU, 메모리, 네트워크 대역폭 등의 자원을 소진시켜 정상적인 운영을 방해합니다.
- **보안 체계 시험**: 조직의 보안 체계를 시험하여 취약점을 발견하거나, 더 큰 공격을 준비하는 데 사용됩니다.
- **금전적 손실 유발**: 서비스 중단으로 인한 금전적 손실을 유발하거나, 금전적인 이득을 목적으로 몸값을 요구하는 경우도 있습니다.

## DDoS 방어 전략

- **네트워크 인프라 보호**
    - **트래픽 필터링**: 공격성 트래픽을 필터링하여 네트워크로 유입되지 않도록 방지합니다.
    - **IP 블랙리스트**: 악의적인 IP 주소를 블랙리스트에 추가하여 차단합니다.
    - **애니캐스트 네트워크**: 트래픽을 여러 서버로 분산시켜 특정 서버에 집중되는 공격을 완화합니다.

- **애플리케이션 레이어 보호**
    - **WAF (Web Application Firewall)**: 애플리케이션 레벨에서 발생하는 공격을 탐지하고 차단합니다.
    - **Rate Limiting**: 특정 IP에서 발생하는 과도한 요청을 제한하여 서버 자원을 보호합니다.

- **리소스 확장**
    - **오토스케일링**: 클라우드 환경에서는 트래픽이 급증할 경우 자동으로 리소스를 확장하여 공격에 대응할 수 있습니다.
    - **CDN 사용**: 콘텐츠를 CDN으로 분산하여 특정 서버에 집중되는 트래픽을 분산시킵니다.

- **공격 탐지 및 모니터링**
    - **실시간 모니터링**: 트래픽 패턴을 실시간으로 모니터링하여 비정상적인 트래픽을 빠르게 탐지합니다.
    - **공격 대응 시스템**: DDoS 공격 탐지 시스템을 구축하여 공격이 발생할 때 즉각적으로 대응할 수 있습니다.

### 구체적인 방어 방법 예시

- **Istio Ingress Gateway**:

    Istio의 Ingress Gateway는 클러스터 외부에서 들어오는 트래픽을 관리하고 라우팅하는 역할을 합니다.
    DDoS 방어를 위해 이 Gateway에서 트래픽을 제한하는 방법을 고려합니다.

    이 경우 Istio Ingress Gateway에 부하가 클 것으로 예상됩니다.

    - **Rate Limiting**

        DDoS 공격을 완화하기 위해 Istio Ingress Gateway에 **Rate Limiting**을 설정하는 것을 고민합니다.
        이를 통해 일정 시간 내에 들어오는 요청 수를 제한하여 과도한 트래픽을 막을 수 있습니다.

        Istio Ingress Gateway에 설정된 Rate Limiting은 *클러스터 외부에서 내부로 들어오는 모든 트래픽에 대해 적용*됩니다.
        이 설정은 클러스터로 유입되는 트래픽을 제한하여, 외부에서 발생하는 DDoS 공격이나 과도한 트래픽으로 인한 부하를 완화하는 데 사용됩니다.

        클러스터의 "입구"에서 전체적으로 Rate Limiting을 적용하는 것이며, 이 설정은 모든 서비스에 동일하게 적용됩니다.
        이는 Istio Ingress Gateway가 트래픽의 첫 번째 진입점이기 때문에, 클러스터 전체에 영향을 미치는 전방위적인 방어를 제공합니다.

        ```yaml
        apiVersion: networking.istio.io/v1alpha3
        kind: Gateway
        metadata:
            name: my-ingress-gateway
        spec:
            selector:
            istio: ingressgateway
            servers:
            - port:
                number: 80
                name: http
                protocol: HTTP
                hosts:
                - "*"
        ---
        apiVersion: networking.istio.io/v1alpha3
        kind: VirtualService
        metadata:
            name: my-rate-limiting
        spec:
            hosts:
            - "*"
            gateways:
            - my-ingress-gateway
            http:
            - route:
                - destination:
                    host: my-service
                retries:
                attempts: 3
                perTryTimeout: 2s
                fault:
                delay:
                    percentage:
                    value: 100.0
                    fixedDelay: 5s
        ```

    - **Istio Virtual Service Rate Limiting**:

        Virtual Service 내에서 특정 서비스, 특정 경로, 호스트, HTTP 메서드 등에 맞춤형으로 Rate Limiting을 적용할 수 있습니다.
        이 경우, 클러스터 내의 특정 서비스나 API 경로에만 제한이 가해지며, 전체 트래픽이 아닌 특정 서비스에 집중됩니다.
        특정 호스트나 경로에 대해 맞춤형 Rate Limiting을 설정할 수 있습니다.

        ```yaml
        apiVersion: networking.istio.io/v1alpha3
        kind: VirtualService
        metadata:
            name: my-service
        spec:
            hosts:
            - my-service.default.svc.cluster.local
            http:
            - match:
                - uri:
                    prefix: "/api/v1/resource"
                route:
                - destination:
                    host: my-service
                retries:
                attempts: 3
                perTryTimeout: 2s
                fault:
                delay:
                    percentage:
                    value: 100.0
                    fixedDelay: 5s

        ```

    - **Circuit Breaking**

        **Circuit Breaker**를 설정하여 시스템의 과부하를 방지할 수 있습니다.
        트래픽이 특정 한도를 넘을 경우 요청을 차단하거나 지연시키는 기능을 활용할 수 있습니다.

- **Web Application Firewall (WAF) 설정**
    - 모니터링
        - **onlyKorea: Block -> Count**

            한국 이외의 트래픽을 차단하던 규칙을 일시적으로 `Count` 모드로 변경하여 트래픽을 모니터링하고,
            이후에 다시 `Block` 모드로 전환합니다.

        - **Sample Enabling**

            WAF에서 샘플링 기능을 활성화하여 트래픽 패턴을 분석하고, 이를 기반으로 규칙을 최적화합니다.

    - **DDoS 방어를 위한 규칙 설정**: WAF에서 DDoS 공격 방어를 위한 규칙을 설정합니다.

        - **Web ACL (Access Control List)**

            WAF의 Web ACL을 사용하여 특정 트래픽 패턴에 대한 규칙을 설정합니다.

            예를 들어, 특정 호스트명을 검사하여 DDoS 공격을 방어하는 규칙을 적용할 수 있습니다.

        - **Hostname 검사**

            특정 호스트명에 대한 트래픽을 모니터링하고, 필요시 이를 차단하는 규칙을 설정합니다.

        - **Single Header 검사**

            특정 HTTP 헤더(예: Host)를 검사하여 의심스러운 트래픽을 필터링합니다.

        - **Blacklist (Blackhole 규칙)**

            특정 IP 또는 트래픽 패턴에 대해 모든 요청을 차단하는 Blackhole 규칙을 설정합니다.

        - **특정 요청 차단**:

            `GET /some/path` 경로로 들어오는 모든 요청을 차단하며, 한국 이외의 지역에서 들어오는 트래픽을 모두 차단합니다.
            이후 공격 패턴에 따라 규칙을 조정합니다.

    - **WAF 원상복구**

        DDoS 공격이 진정된 후 WAF 설정을 원상복구하고, 필요시 DDoS 대응 전의 설정으로 되돌립니다.

- Istio 시스템 내 모니터링

    Istio 시스템 내에서 실행 중인 모든 Pod 상태를 확인하여 트래픽이 제대로 라우팅되고 있는지 모니터링합니다.

    ```sh
    kubens istio-system
    kubecolor get pods
    ```
