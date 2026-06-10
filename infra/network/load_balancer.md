# 로드밸런서의 정체 — 'L4 장비'는 어떻게 HTTPS를 끊고 평문 HTTP를 보내는가

이 문서의 출발 질문은 세 개가 연쇄된 것입니다.

1. **Citrix 'L4' 로드밸런서가 HTTPS 요청을 받으면, 세션키 같은 걸로 복호화해서 뒤의 물리 서버로 전달하는가? 전달은 HTTP인가, TCP 소켓인가?**
2. **"TCP로 보낸다"는 게 구체적으로 뭔가? 연결된 소켓이 실제로 존재하는가? 소켓 프로그래밍처럼 동작한다는 뜻인가? L4 장비가 어떻게 복호화된 평문 HTTP를 만들어 보내는가?**
3. **그 장비(ADC라는 것)는 대체 뭐고, 내부적으로 어떻게 동작하며, 가능한 여러 방식 중 왜 하필 그렇게 하고, 왜 그런 장비를 쓰는가?**

세 질문이 가리키는 진짜 질문은 하나입니다. **"이 박스의 정체가 뭔가 — 패킷을 만지는 네트워크 장비인가, 소켓을 가진 프로그램인가?"** 'L4 장비'라는 호칭은 스위치·라우터의 친척처럼 들리는데, 실제 하는 일(TLS 복호화, HTTP 헤더 조작)은 애플리케이션 프로그램의 일이라서 둘이 충돌하는 것처럼 보입니다. 이 충돌을 푸는 것이 이 문서의 목표입니다.

## 목차

- [한 줄 답](#한-줄-답)
- [용어와 정체 — L4 스위치, 로드밸런서, ADC](#용어와-정체--l4-스위치-로드밸런서-adc)
- [근본 구현 방식 두 가지 — 패킷 모드와 프록시 모드](#근본-구현-방식-두-가지--패킷-모드와-프록시-모드)
- [TLS가 올라가면 — 패스스루, 종단, 재암호화](#tls가-올라가면--패스스루-종단-재암호화)
- [요청 한 건의 일생 — 바이트 단위 추적](#요청-한-건의-일생--바이트-단위-추적)
- [동작하는 미니 구현](#동작하는-미니-구현)
- [실제 장비는 무엇이 다른가](#실제-장비는-무엇이-다른가)
- [왜 이렇게 하는가 — 설계 결정들의 이유](#왜-이렇게-하는가--설계-결정들의-이유)
- [백엔드 개발자에게 체감되는 결과들](#백엔드-개발자에게-체감되는-결과들)
- [처음 세 질문에 대한 압축 답](#처음-세-질문에-대한-압축-답)
- [용어집](#용어집)
- [근거와 더 읽을거리](#근거와-더-읽을거리)

## 한 줄 답

**현대 로드밸런서 장비의 기본 동작 모드는 "최적화된 컴퓨터 위에서 도는 프록시 프로그램"이며, 'L4'는 동작 계층의 설명이 아니라 1990년대 제품 카테고리에서 굳어진 호칭입니다.** HTTPS를 복호화해 평문 HTTP로 보내는 구성(SSL 오프로드)에서 그 장비는 nginx 리버스 프록시와 의미적으로 동일하게 동작합니다. 클라이언트 쪽 TCP 연결을 종단하고(소켓 ①), TLS 핸드셰이크의 서버 측 당사자가 되어 세션키를 직접 협상하고, 복호화된 HTTP를 파싱·변형한 뒤, 백엔드로 **자기가 직접 맺은 별개의 TCP 연결(소켓 ②)** 위에 평문 HTTP 바이트를 써 넣습니다. "연결되어 있는 소켓이 있는가?"의 답은 "그렇다, 두 개가 있다"이고, "소켓 프로그래밍처럼 하는가?"의 답은 "의미상 정확히 그렇다(전용 장비는 성능을 위해 OS 소켓 API 대신 자체 TCP 스택을 쓸 뿐)"입니다.

복호화 여부와 전달 형태는 박스가 정하는 게 아니라 설정(가상서버 타입)이 정합니다. 진짜 4계층으로만 동작시키면(패스스루) 복호화 없이 암호문 바이트를 그대로 중계하고, TLS 종단으로 설정하면 위 동작을 합니다.

## 용어와 정체 — L4 스위치, 로드밸런서, ADC

### ADC라는 이름

**ADC는 Application Delivery Controller의 약자**로, 로드밸런서가 기능을 확장하면서 업계(특히 Gartner의 제품 카테고리 분류)가 붙인 이름입니다. 순수 부하분산을 넘어 SSL 종단, HTTP 압축·캐싱, 콘텐츠 기반 라우팅, WAF, GSLB(DNS 기반 글로벌 분산)까지 묶은 박스를 가리킵니다. 대표 제품이 Citrix NetScaler, F5 BIG-IP, A10 Thunder, Radware Alteon입니다.

NetScaler의 연혁이 이 시장의 역사를 압축합니다. 1990년대 말 설립된 NetScaler를 2005년 Citrix가 인수했고, 2018년 제품명을 'Citrix ADC'로 바꿨다가, Citrix가 Cloud Software Group으로 합병된 뒤 2023년경 다시 'NetScaler' 브랜드로 돌아왔습니다. 그래서 문서·현장에서 NetScaler, Citrix ADC, Citrix NetScaler가 모두 같은 물건을 가리킵니다.

### 왜 한국에서는 'L4'라고 부르는가

1990년대 중후반, 웹 서버 부하분산 장비가 처음 상용화될 때의 제품 카테고리 이름이 **'Layer 4 스위치'** 였습니다. Cisco LocalDirector, Alteon(이후 Nortel을 거쳐 Radware로), Foundry ServerIron 같은 제품들이 그 세대입니다. 당시 L2/L3 스위치와의 차별점이 "TCP/UDP 포트(4계층 정보)까지 보고 스위칭한다"였기 때문에 붙은 이름입니다. 이후 장비들이 HTTP를 파싱하는 7계층 기능을 흡수하며 'L4-L7 스위치' → 'ADC'로 카테고리명이 바뀌었지만, 한국 엔터프라이즈 현장에서는 **장비 통칭으로 'L4'가 그대로 굳었습니다.** "L4 한 대 넣자"는 말은 동작 계층의 진술이 아니라 "로드밸런서 장비 한 대 넣자"는 뜻입니다.

이 어휘의 관성이 정확히 출발 질문의 혼란을 만듭니다. 'L4 장비'가 TLS를 복호화한다는 말은 형용모순처럼 들리지만, 실제로는 "ADC라고 불러야 할 박스를 L4라고 부르고 있고, 그 박스가 L7 프록시로 설정되어 있다"는 상황일 뿐입니다.

### 박스를 열면 무엇이 있는가

물리적으로 ADC는 **서버 컴퓨터입니다.** x86 CPU, RAM, NIC가 들어 있고, 모델에 따라 TLS 핸드셰이크의 비대칭 암호 연산을 전담하는 가속 칩이 추가됩니다. 그 위에 전용 OS가 돕니다. NetScaler는 FreeBSD 기반이고, F5 BIG-IP는 Linux 기반 TMOS입니다. 그리고 그 OS 위에서 **트래픽을 처리하는 프로그램**(NetScaler의 Packet Engine, F5의 TMM)이 돌아갑니다.

즉 이 박스는 라우터의 친척이라기보다 **nginx가 설치된 서버의 친척**입니다. 다른 점은 범용 OS의 커널 네트워크 스택과 소켓 API를 거치지 않고 자체 구현으로 대체해 성능을 끌어올렸다는 것, 그리고 관리 UI·이중화·헬스체크 같은 운영 기능이 제품으로 묶여 있다는 것입니다. 이 차이는 [실제 장비는 무엇이 다른가](#실제-장비는-무엇이-다른가)에서 다룹니다.

## 근본 구현 방식 두 가지 — 패킷 모드와 프록시 모드

로드밸런서 구현은 깊은 곳에서 두 가족으로 갈립니다. 이 구분이 모든 후속 질문("소켓이 있는가", "복호화가 가능한가", "클라이언트 IP가 보존되는가")의 답을 결정합니다.

### 패킷 모드 — 진짜 4계층, 소켓이 없는 세계

패킷 모드 장비는 **클라이언트가 보낸 패킷 그 자체의 헤더를 고쳐서 서버로 흘려보냅니다.** 라우터 + NAT의 확장이라고 보면 정확합니다. 장비 안에는 소켓이 없습니다. TCP 연결을 종단하지 않으므로 시퀀스 번호를 관리하지 않고, 수신 버퍼도 없고, 재전송도 하지 않습니다. 있는 것은 "이 흐름(5-tuple)은 어느 서버로 보내기로 했다"를 기억하는 흐름 테이블(conntrack)뿐입니다.

```text
# 패킷 모드 의사코드 — 장비 안에 소켓이 없다. 흐름 테이블만 있다.
flow_table = {}   # (client_ip, client_port, vip, vport) -> server

def on_packet_from_client(pkt):
    key = (pkt.src_ip, pkt.src_port, pkt.dst_ip, pkt.dst_port)
    server = flow_table.get(key)
    if server is None:
        if not pkt.tcp.SYN:          # 모르는 흐름의 중간 패킷이면 버린다
            return drop(pkt)
        server = pick_server(key)    # 해시 / 라운드로빈 / 최소연결
        flow_table[key] = server
    pkt.dst_ip  = server.ip          # DNAT: 목적지 IP만 서버로 바꾼다
    pkt.dst_mac = arp(server.ip)
    fix_checksums(pkt)
    transmit(pkt)                    # 시퀀스 번호도 페이로드도 클라이언트 것 그대로

def on_packet_from_server(pkt):
    pkt.src_ip = VIP                 # un-NAT: 출발지를 VIP로 되돌린다
    fix_checksums(pkt)
    transmit(pkt)
```

TCP 핸드셰이크는 **클라이언트와 백엔드 서버 사이에 end-to-end로 하나만** 성립합니다. 장비는 그 대화의 당사자가 아니라 우체부입니다. 세부 변형이 몇 가지 있습니다.

- **NAT 모드** — 위 의사코드. 왕복 트래픽이 모두 장비를 경유합니다. (LVS-NAT)
- **DR/DSR 모드(Direct Server Return)** — 목적지 **MAC 주소만** 서버 것으로 바꿔 같은 L2 세그먼트의 서버로 던집니다. 서버는 루프백 인터페이스에 VIP를 들고 있어서(ARP 응답은 억제) 그 패킷을 자기 것으로 받아들이고, **응답은 장비를 거치지 않고 클라이언트로 직행**합니다. 응답이 요청보다 훨씬 큰 웹 트래픽에서 장비 대역폭을 극적으로 아낍니다. (LVS-DR)
- **터널 모드** — 원본 패킷을 IPIP로 감싸 다른 네트워크의 서버로 보냅니다. (LVS-TUN)

이 가족의 현대적 구성원이 리눅스 커널의 LVS/IPVS, Google의 Maglev(ECMP + 일관 해싱으로 상태 최소화), Meta의 katran(XDP/eBPF로 NIC 드라이버 레벨에서 패킷 처리), AWS NLB입니다.

패킷 모드의 본질적 한계가 중요합니다. **스트림을 재조립하지 않으므로 페이로드를 "읽을" 수 없고, TLS 핸드셰이크의 당사자가 아니므로 세션키가 존재하지 않아 복호화가 원리적으로 불가능합니다.** HTTP 경로 기반 라우팅, 쿠키 persistence, 헤더 삽입 모두 불가능합니다. 대신 패킷당 처리 비용이 극히 작아 초당 수백만 패킷을 다루고, 클라이언트 IP가 자연히 보존됩니다.

### 프록시 모드 — 연결을 끝내고, 다시 만든다

프록시 모드 장비는 **TCP의 종단(endpoint)입니다.** 클라이언트의 3-way 핸드셰이크를 자기가 완성하고(연결 ①), 백엔드로 자기가 새 핸드셰이크를 시작합니다(연결 ②). 그리고 두 연결 사이에서 바이트를 복사합니다. "전달"의 실체는 패킷 포워딩이 아니라 **연결 ①의 수신 버퍼에서 연결 ②의 송신 버퍼로의 메모리 복사**입니다.

```text
# 프록시 모드 의사코드 — accept / connect / copy. 소켓 프로그래밍 그 자체.
def on_new_client(listen_socket):
    c = accept(listen_socket)        # 연결 ①: 클라이언트 ↔ LB
    s = connect(pick_server())       # 연결 ②: LB ↔ 서버 (LB가 클라이언트 역할)
    splice_bidirectional(c, s)       # 양방향 바이트 복사
```

두 연결은 **완전히 독립**입니다. 각자 자기 시퀀스 번호 공간, 자기 윈도 크기, 자기 MSS, 자기 혼잡 제어, 자기 재전송을 가집니다. 클라이언트가 보낸 패킷은 백엔드에 단 하나도 도달하지 않습니다. 살아남는 것은 애플리케이션 바이트뿐입니다. 백엔드가 보는 연결의 출발지 IP도 클라이언트가 아니라 장비의 주소(NetScaler 용어로 SNIP)입니다.

NetScaler, F5 BIG-IP의 표준 가상서버, HAProxy, nginx, Envoy가 모두 이 가족입니다. 특히 **NetScaler는 "풀 프록시 아키텍처"라서, 복호화를 전혀 하지 않는 TCP 모드나 SSL_BRIDGE 모드로 설정해도 여전히 연결 두 개를 종단하고 바이트를 복사합니다.** 그때는 단지 복사하는 바이트가 (자기가 풀 수 없는) 암호문일 뿐입니다.

한 박스가 두 모드를 다 제공하기도 합니다. F5에서 Standard 프로파일 가상서버는 풀 프록시, FastL4 프로파일은 패킷 가속 경로입니다. NetScaler도 MAC 기반 포워딩(MBF) 같은 패킷성 동작을 옵션으로 지원합니다. **모드를 정하는 것은 박스가 아니라 설정**이라는 사실이 여기서도 확인됩니다.

### 비교

| | 패킷 모드 (진짜 L4) | 프록시 모드 |
|---|---|---|
| 장비 내부 상태 | 흐름 테이블(NAT 항목) | 연결마다 TCB 2개 + 버퍼 |
| TCP 연결 | 클라이언트↔서버 1개 (end-to-end) | 클라이언트↔LB, LB↔서버 2개 |
| 소켓 존재 | 없음 | 있음 (양쪽에) |
| 페이로드 접근 | 불가 (첫 패킷 엿보기 정도) | 가능 (재조립된 스트림) |
| TLS 복호화 | 원리적 불가 | 설정 시 가능 (TLS 종단) |
| 클라이언트 IP | 보존됨 | 기본적으로 LB IP로 대체 |
| 응답 경로 | DSR이면 LB 우회 가능 | 반드시 LB 경유 |
| 성능 특성 | 패킷당 비용 극소 | 연결당 메모리·CPU 소요 |
| 대표 구현 | LVS, Maglev, katran, AWS NLB | NetScaler, F5 Standard, HAProxy, nginx, AWS ALB |

## TLS가 올라가면 — 패스스루, 종단, 재암호화

### '복호화 권한'은 어디서 오는가 — 세션키의 진실

출발 질문의 "세션키 등을 통해 복호화해서"라는 표현에는 흔한 오해가 숨어 있습니다. **TLS 세션키는 어디서 받아오는 것이 아니라, 핸드셰이크의 두 당사자가 그 자리에서 함께 만들어내는 것입니다.**

TLS 핸드셰이크를 키 관점에서만 요약하면 이렇습니다. 클라이언트와 서버가 각자 임시(ephemeral) 키 쌍을 만들어 공개값을 교환하고(ECDHE), 양쪽 모두 같은 공유 비밀을 계산해 거기서 대칭 세션키를 유도합니다. 서버 인증서의 **개인키가 하는 일은 이 교환 내용에 서명해서 "나는 인증서의 주인"임을 증명하는 것이지, 트래픽을 복호화하는 것이 아닙니다.** 임시 키는 세션이 끝나면 버려지므로, 나중에 개인키가 유출돼도 과거 트래픽을 풀 수 없습니다(전방 비밀성, PFS). TLS 1.3은 아예 이 방식만 남겼습니다.

따라서 중간 장비가 트래픽을 복호화할 수 있는 경로는 사실상 하나뿐입니다. **자기가 핸드셰이크의 서버 측 당사자가 되는 것.** 인증서와 개인키를 장비에 설치한다는 것은 "이 장비가 클라이언트에 대해 TLS 서버 노릇을 한다"는 선언이고, 그 결과 세션키는 클라이언트와 장비 사이에서 협상되어 장비가 자연히 보유하게 됩니다. 옆에서 패킷을 복사해 듣고 있다가 키를 "얻어서" 푸는 모델은 PFS 때문에 성립하지 않습니다.

### 세 가지 처리 방식

| 방식 | NetScaler 설정 | 복호화 | 백엔드가 받는 것 | 인증서·개인키 위치 |
|---|---|---|---|---|
| 패스스루 | `SSL_BRIDGE` vserver | 안 함 | 암호문 TLS 스트림 그대로 (TCP 중계) | 백엔드 서버 |
| 종단(오프로드) | `SSL` vserver + `HTTP` service | 함 | 평문 HTTP | LB |
| 종단 + 재암호화 | `SSL` vserver + `SSL` service | 함 | 새 TLS 세션으로 재암호화된 HTTPS | LB와 백엔드 양쪽 |

- **패스스루**에서 LB는 암호문을 운반만 합니다. 그래도 할 수 있는 일이 아예 없지는 않습니다. 핸드셰이크 초반은 평문이므로, ClientHello에 실리는 **SNI(요청 호스트명)를 엿보고 호스트별로 다른 서버 풀로 보내는 라우팅**이 가능하고(nginx stream 모듈의 `ssl_preread`, HAProxy의 `req_ssl_sni`가 같은 기법), SSL 세션 ID 기반 persistence도 가능합니다. 다만 TLS 1.3에서 세션 ID는 호환용 잔재가 되어 이 persistence는 사실상 수명을 다했고, SNI 평문 노출도 ECH(Encrypted Client Hello)가 보급되면 막힙니다(2026년 초 기준 ECH는 아직 보편화 전).
- **오프로드**에서 백엔드와의 구간은 평문입니다. "내부망은 신뢰 구간"이라는 전제가 깔린 구성이고, 그 전제를 받아들일 수 없으면(제로 트러스트, 금융권 내부 규정, 개인정보 구간 암호화 요건) 세 번째 방식으로 갑니다.
- **재암호화(end-to-end TLS)** 에서도 TLS 세션은 클라이언트–LB, LB–서버 **두 개로 분리**됩니다. 하나의 세션이 관통하는 게 아닙니다. LB는 평문을 보는 지점이며, 그래서 L7 기능(라우팅·WAF·로깅)을 유지한 채 전 구간 암호화를 달성합니다. 비용은 암호 연산 2배입니다.

참고로 NetScaler의 `SSL_TCP` 타입은 "TLS는 종단하되 그 안의 내용을 HTTP로 파싱하지 않는" 모드입니다. TLS 위에 HTTP가 아닌 프로토콜(자체 전문 프로토콜 등)이 올라갈 때 씁니다.

용어 함정 하나. **Citrix의 `SSL_BRIDGE`는 복호화 없는 패스스루**지만, **F5 문서의 "SSL bridging"은 정반대로 종단 후 재암호화**를 뜻합니다. 벤더 문서를 교차해 읽을 때 단어가 아니라 동작으로 확인해야 합니다.

## 요청 한 건의 일생 — 바이트 단위 추적

SSL 오프로드 구성에서 `GET /api` 한 건이 흘러가는 전 과정입니다.

```text
클라이언트                  NetScaler (VIP:443)                물리 서버 (10.0.0.11:8080)
    │                            │                                  │
    ├── TCP SYN ───────────────►│                                  │
    │◄── SYN/ACK ────────────────┤  연결 ① 성립. LB가 TCB 할당       │
    ├── ACK ────────────────────►│  (시퀀스번호·윈도·버퍼)            │
    │                            │                                  │
    ├── ClientHello ────────────►│                                  │
    │◄── ServerHello+인증서 ──────┤  LB가 개인키로 서명 = TLS 서버      │
    ├── (키 교환 완료) ──────────►│  양쪽이 세션키 유도                │
    │                            │                                  │
    ├── TLS레코드[GET /api ...] ─►│  TCP 재조립 → TLS 복호화           │
    │                            │  → 평문 HTTP 파싱                 │
    │                            │  → 정책: 서버 선택, 헤더 삽입       │
    │                            │                                  │
    │                            ├── TCP SYN (src=SNIP) ──────────►│
    │                            │◄── SYN/ACK ──────────────────────┤  연결 ② 성립
    │                            ├── ACK ──────────────────────────►│  (또는 풀의 기존 연결 재사용)
    │                            │                                  │
    │                            ├── "GET /api HTTP/1.1\r\n         │
    │                            │    Host: ...\r\n                 │
    │                            │    X-Forwarded-For: <클라IP>     │
    │                            │    ..." ────────────────────────►│  평문 HTTP
    │                            │◄── "HTTP/1.1 200 OK ..." ────────┤
    │◄── TLS레코드[200 OK ...] ───┤  재암호화 후 연결 ①로 송신          │
```

각 단계에서 일어나는 일을 풀면 이렇습니다.

1. **연결 ① 성립** — 핸드셰이크의 상대는 물리 서버가 아니라 LB 자신입니다. LB 내부에 이 연결의 제어 블록(TCB: 시퀀스 번호, 윈도, 수신·송신 버퍼, 타이머)이 생깁니다. OS 커널이 소켓마다 들고 있는 것과 같은 자료구조를, 전용 장비는 자체 스택의 연결 테이블에 만듭니다.
2. **TLS 종단** — LB가 인증서를 제시하고 개인키로 서명합니다. 세션키가 클라이언트와 LB 사이에 만들어집니다. 이후 도착하는 TLS 레코드를 LB가 복호화하면 메모리에 평문 HTTP 바이트가 생깁니다.
3. **L7 처리** — 평문이므로 비로소 가능한 일들: 요청 라인·헤더 파싱, 경로/호스트 기반 서버군 선택, 쿠키 기반 persistence, `X-Forwarded-For`·`X-Forwarded-Proto` 삽입, WAF 검사, 압축, 액세스 로깅.
4. **연결 ② 확보** — 선택된 서버로 LB가 직접 connect 합니다(출발지 IP는 SNIP). 실제 장비는 매 요청 새로 맺지 않고 **서버 쪽 keep-alive 연결 풀을 유지하면서 여러 클라이언트의 요청을 소수의 백엔드 연결에 다중화**합니다(NetScaler의 connection multiplexing). 요청 경계를 파싱할 수 있는 L7 프록시라서 가능한 일입니다.
5. **전송** — 변형된 요청 바이트가 연결 ②의 송신 버퍼에 쓰이고, LB의 TCP 스택이 그것을 새 세그먼트로 잘라 보냅니다. **백엔드로 나가는 패킷은 클라이언트 패킷을 고친 것이 아니라 처음부터 새로 만들어진 것입니다.**
6. **백엔드 관점** — 평범한 HTTP 클라이언트가 접속해 평문 요청을 보낸 것과 구별할 수 없습니다. accept 하고 read 하면 `GET /api HTTP/1.1`이 나옵니다. curl이 보낸 것과 다른 점은 출발지가 SNIP라는 것, 그리고 `X-Forwarded-For` 같은 흔적 헤더뿐입니다.
7. **응답** — 역방향으로 같은 일이 일어납니다. 평문으로 받아, 세션키로 암호화해, 연결 ①로 내보냅니다.

## 동작하는 미니 구현

세 모드를 코드로 직접 만들어 보면 "장비의 마법"이 사라집니다. 아래 두 프로그램은 실제로 실행됩니다.

### 패스스루 프록시 — SSL_BRIDGE의 의미론 (실행 가능)

내용을 전혀 모른 채 바이트만 복사합니다. 평문 HTTP든 TLS 암호문이든 똑같이 동작합니다 — **바로 그 무지(無知)가 패스스루의 정의입니다.**

```python
#!/usr/bin/env python3
"""최소 TCP 프록시 = NetScaler SSL_BRIDGE(패스스루)의 의미론.
페이로드가 무엇인지 모른 채 양방향으로 바이트만 복사한다."""
import socket
import threading
import itertools

LISTEN = ("0.0.0.0", 8081)                      # VIP 역할 (데모용 비특권 포트)
BACKENDS = [("127.0.0.1", 8080)]                # 백엔드 목록
rr = itertools.cycle(BACKENDS)                  # 라운드로빈

def relay(src: socket.socket, dst: socket.socket) -> None:
    """src에서 읽어 dst로 쓴다. EOF(FIN)는 half-close로 전파한다."""
    try:
        while True:
            data = src.recv(65536)
            if not data:                        # 상대가 보낼 것을 다 보냄
                break
            dst.sendall(data)
    except OSError:
        pass
    finally:
        try:
            dst.shutdown(socket.SHUT_WR)        # FIN 전파
        except OSError:
            pass

def handle(client: socket.socket) -> None:
    target = next(rr)                           # ← 로드밸런싱 결정 지점
    try:
        backend = socket.create_connection(target)   # 연결 ②: LB가 직접 connect
    except OSError:
        client.close()
        return
    t = threading.Thread(target=relay, args=(backend, client), daemon=True)
    t.start()
    relay(client, backend)
    t.join()
    client.close()
    backend.close()

def main() -> None:
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.bind(LISTEN)
    srv.listen(128)
    while True:
        client, _ = srv.accept()                # 연결 ①의 핸드셰이크는 OS가 완료
        threading.Thread(target=handle, args=(client,), daemon=True).start()

if __name__ == "__main__":
    main()
```

실험: `python -m http.server 8080`을 띄우고 위 프록시를 실행한 뒤 `curl http://127.0.0.1:8081/`을 치면 8080의 응답이 그대로 옵니다. BACKENDS를 TLS 서버로 바꿔도 코드 수정 없이 동작합니다. 프록시는 자기가 나르는 것이 TLS인지조차 모릅니다.

### TLS 종단 + 평문 HTTP 포워딩 — SSL 오프로드의 의미론 (실행 가능)

이 80줄이 "NetScaler가 HTTPS를 받아 복호화하고 평문 HTTP로 백엔드에 보낸다"의 의미적 전부입니다.

```python
#!/usr/bin/env python3
"""TLS 종단 리버스 프록시 = NetScaler 'SSL vserver + HTTP service'(SSL 오프로드)의 의미론."""
import socket
import ssl
import threading
import itertools

LISTEN = ("0.0.0.0", 8443)
BACKENDS = [("127.0.0.1", 8080), ("127.0.0.1", 8090)]
rr = itertools.cycle(BACKENDS)

# 장비에 인증서를 '설치'하는 행위에 해당 — 이로써 이 프로그램이 TLS 서버가 된다
ctx = ssl.SSLContext(ssl.PROTOCOL_TLS_SERVER)
ctx.load_cert_chain(certfile="lb.crt", keyfile="lb.key")

def pump(src, dst) -> None:
    """남은 바이트를 한쪽에서 다른쪽으로 복사. dst가 TLS 소켓이면 쓰는 순간 암호화된다."""
    try:
        while True:
            data = src.recv(65536)
            if not data:
                break
            dst.sendall(data)
    except OSError:
        pass
    finally:
        try:
            dst.shutdown(socket.SHUT_WR)
        except OSError:
            pass

def handle(tls_client: ssl.SSLSocket, client_addr) -> None:
    try:
        # recv가 돌려주는 것은 이미 복호화된 평문이다. 복호화는 TLS 계층(ssl 모듈)이 수행.
        buf = b""
        while b"\r\n\r\n" not in buf:
            chunk = tls_client.recv(8192)
            if not chunk:
                return
            buf += chunk
        head, _, body_start = buf.partition(b"\r\n\r\n")
        lines = head.decode("latin-1").split("\r\n")
        request_line, headers = lines[0], lines[1:]

        # ── L7 프록시만 할 수 있는 일: 평문 헤더를 읽고 고친다 ──
        headers = [h for h in headers
                   if not h.lower().startswith(("connection:", "x-forwarded-for:"))]
        headers.append(f"X-Forwarded-For: {client_addr[0]}")
        headers.append("X-Forwarded-Proto: https")
        headers.append("Connection: close")     # 데모 단순화: 요청 1건 = 연결 1개

        backend = socket.create_connection(next(rr))      # 연결 ② (평문 TCP)
        new_head = "\r\n".join([request_line, *headers]).encode("latin-1")
        backend.sendall(new_head + b"\r\n\r\n" + body_start)   # 평문 HTTP로 송신

        # 남은 요청 바디(있다면)와 응답을 양방향 중계.
        # backend→tls_client 방향은 sendall 시점에 자동으로 TLS 암호화되어 나간다.
        t = threading.Thread(target=pump, args=(tls_client, backend), daemon=True)
        t.start()
        pump(backend, tls_client)
        t.join()
        backend.close()
    except (OSError, ssl.SSLError):
        pass
    finally:
        tls_client.close()

def main() -> None:
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.bind(LISTEN)
    srv.listen(128)
    while True:
        raw, addr = srv.accept()                # 연결 ①: TCP 성립
        try:
            # TLS 핸드셰이크 수행 지점 — 여기서 세션키가 '협상으로 생성'된다
            tls_client = ctx.wrap_socket(raw, server_side=True)
        except (ssl.SSLError, OSError):
            raw.close()
            continue
        threading.Thread(target=handle, args=(tls_client, addr), daemon=True).start()

if __name__ == "__main__":
    main()
```

실험 절차와 관찰 포인트:

```sh
# 1) 자가서명 인증서 = '장비에 인증서 설치'
openssl req -x509 -newkey rsa:2048 -nodes -keyout lb.key -out lb.crt \
        -days 365 -subj "/CN=localhost"

# 2) 백엔드 두 대 (각각 다른 터미널에서)
python -m http.server 8080
python -m http.server 8090

# 3) 프록시 실행 후
curl -vk https://127.0.0.1:8443/
```

- `curl -v` 출력의 핸드셰이크 부분에서 **CN=localhost인 우리 인증서**가 보입니다. 클라이언트의 TLS 상대가 백엔드가 아니라 프록시라는 직접 증거입니다.
- `python -m http.server`의 액세스 로그에는 **127.0.0.1발 평문 GET**이 찍힙니다. 백엔드는 TLS의 존재를 모릅니다. 클라이언트 IP도 모릅니다(헤더로만 전달됨).
- 여러 번 호출하면 8080과 8090에 번갈아 찍힙니다. 로드밸런싱 결정이 코드의 `next(rr)` 한 줄임을 확인할 수 있습니다.

### 두 코드의 차이가 곧 두 모드의 차이

패스스루 코드에는 `ssl`도 HTTP 파싱도 없습니다. 종단 코드에는 인증서 로드, `wrap_socket`(핸드셰이크), 헤더 파싱·변형이 있습니다. **NetScaler에서 vserver 타입을 `SSL_BRIDGE`에서 `SSL`로 바꾸는 것은, 개념적으로 첫 번째 프로그램을 두 번째 프로그램으로 갈아 끼우는 것입니다.** 박스는 그대로인데 돌아가는 로직이 달라집니다.

## 실제 장비는 무엇이 다른가

위 데모와 실제 NetScaler의 차이는 의미론이 아니라 **공학적 규모**입니다.

- **스레드-퍼-커넥션이 아니라 이벤트 루프** — 데모는 연결마다 스레드를 만들지만, 수십만 동시 연결에서는 불가능한 구조입니다. 실제 장비는 nginx처럼 소수의 코어 전용 프로세스가 이벤트 루프(어느 연결에 읽을 데이터가 생겼는지 통지받아 처리)를 돌립니다.

```text
# 이벤트 루프 골격
while true:
    events = wait_for_io()               # epoll/kqueue 상당
    for (conn, ev) in events:
        data = read(conn)
        if conn.tls:  data = tls_decrypt(conn, data)
        out = apply_policy(conn, data)   # 파싱·라우팅·헤더 변형
        peer = conn.peer
        if peer.tls:  out = tls_encrypt(peer, out)
        write(peer, out)
```

- **커널 우회와 자체 TCP/IP 스택** — NetScaler는 FreeBSD 위에서 돌지만 데이터 경로는 커널 소켓을 쓰지 않습니다. Packet Engine이라는 프로세스가 NIC에서 패킷을 직접 가져와 자체 구현 TCP/IP 스택으로 처리합니다(인터럽트·시스템콜·컨텍스트 스위치 제거, run-to-completion). 멀티코어 모델(nCore)에서는 코어마다 Packet Engine이 하나씩 돌고 NIC의 RSS가 흐름을 코어에 분배합니다. F5의 TMM도 같은 사상입니다. DPDK 기반 소프트웨어 LB들이 같은 기법을 범용 서버에서 재현합니다. **"OS 소켓 API를 쓰지 않는다"는 것이지 "소켓 의미론이 없다"는 것이 아닙니다.** 연결 테이블의 항목 하나가 곧 소켓 하나에 해당합니다.
- **암호 가속** — TLS 핸드셰이크의 비대칭 연산(서명)은 CPU에 비쌉니다. 하드웨어 어플라이언스는 이를 전담하는 가속 칩에 오프로드하고, 대칭(벌크) 암복호는 CPU의 AES-NI로 처리합니다. "SSL 오프로드 장비"라는 상품성 자체가, 웹서버 CPU에서 이 비용을 떼어내던 2000년대의 필요에서 나왔습니다.
- **운영 기능의 제품화** — 헬스체크(주기적으로 백엔드에 프로브를 보내 풀에서 제외/복귀), persistence 테이블, 연결 드레이닝, Active-Standby 이중화(연결 테이블 동기화로 장애 시 세션 유지), SYN flood 흡수(핸드셰이크가 완성되기 전에는 백엔드에 아무것도 전달하지 않고, SYN 쿠키로 상태 없이 버팀) 등이 박스에 묶여 있습니다.

## 왜 이렇게 하는가 — 설계 결정들의 이유

### 왜 LB에서 TLS를 끊는가 (오프로드의 동기)

1. **인증서 수명주기를 한 곳에서** — 서버 N대에 인증서·개인키를 배포·갱신·폐기하는 대신 장비 한 쌍에서 관리합니다. 갱신 누락으로 인한 장애 표면적이 줄고, 개인키 보관 지점이 줄며, HSM 연동도 한 곳에서 끝납니다.
2. **L7 기능은 평문을 요구** — 경로·호스트 기반 라우팅, 쿠키 persistence, WAF, 헤더 조작, 압축, 캐싱, 상세 액세스 로깅 전부 평문에서만 가능합니다. 오프로드는 이 기능들의 전제 조건입니다.
3. **암호 연산의 집중과 가속** — 핸드셰이크 비용을 앱 서버에서 떼어 가속 하드웨어가 있는 장비로 모읍니다. AES-NI 보급 이후 벌크 암호화는 싸졌지만, 핸드셰이크 폭주(트래픽 스파이크, 커넥션 스톰) 흡수와 TLS 정책(허용 버전·암호스위트)의 단일 관리 지점이라는 가치는 남아 있습니다.
4. **반대급부** — LB 뒤 구간이 평문이 됩니다. 이를 수용할 수 없는 보안 요건(금융권 내부 규정, 제로 트러스트)에서는 재암호화 구성으로 가고, 암호 연산 2배를 비용으로 치릅니다.

### 왜 풀 프록시를 기본 구조로 택했는가

패킷 모드가 더 싸고 빠른데도 NetScaler·F5가 풀 프록시를 기본으로 삼은 이유입니다.

1. **양쪽 TCP를 따로 최적화** — 클라이언트 구간은 손실 많고 RTT 큰 WAN, 서버 구간은 깨끗한 LAN입니다. 연결이 분리되어 있으면 윈도·혼잡제어·MSS·keep-alive를 구간별로 다르게 튜닝할 수 있습니다.
2. **느린 클라이언트로부터 서버 보호** — 모바일 클라이언트가 응답을 1분에 걸쳐 받아가도, 서버는 LB 버퍼에 응답을 밀어 넣고 즉시 다음 일을 합니다. 서버의 워커 스레드 점유 시간이 LB의 버퍼링으로 단축됩니다(nginx를 앱 서버 앞에 두는 고전적 이유와 동일).
3. **연결 다중화** — 클라이언트 1만 명의 연결을 백엔드 연결 수십 개로 모아 줍니다. 백엔드의 accept 부하, 연결당 메모리, TIME_WAIT 적체가 줄어듭니다. 요청 경계를 아는 L7 프록시만 할 수 있습니다.
4. **공격 흡수** — 3-way 핸드셰이크가 완성되기 전까지 백엔드는 아무것도 보지 않습니다. SYN flood가 LB에서 멈춥니다.
5. **연결 수립 시점의 지능** — 그 순간 헬스한 서버에만 새 연결을 보내고, connect 실패 시 다른 서버로 재시도하는 것이 자연스럽습니다. 패킷 모드는 흐름이 일단 박히면 끝까지 그 서버입니다.

### 왜 패킷 모드도 여전히 쓰이는가

프록시는 연결마다 TCB 2개와 버퍼를 듭니다. 초당 수십만 신규 연결, 수천만 동시 연결 규모(통신사·클라우드 엣지·DDoS 방어 전면)에서는 이 비용이 감당이 안 됩니다. 그래서 **최전면에 상태 최소화 패킷 모드(Maglev, katran, NLB), 그 뒤에 기능 담당 L7 프록시(Envoy, ALB, nginx)** 를 두는 2단 구성이 하이퍼스케일러의 표준 패턴이 됐습니다. DSR로 응답 대역폭을 우회시킬 수 있다는 점, 클라이언트 IP가 그대로 보존된다는 점도 패킷 모드를 선택하는 이유입니다.

### 왜 전용 장비를 쓰(었)는가 — 그리고 지금

2000년대에는 범용 서버 + 범용 OS로 라인레이트 트래픽 처리와 TLS 핸드셰이크 폭주를 감당할 수 없었습니다. 커널 우회 스택과 암호 가속 칩을 묶은 어플라이언스가 그 격차를 메웠고, 헬스체크·이중화·persistence가 제품으로 통합되어 "네트워크팀이 운영하는 박스"라는 운영 모델이 자리잡았습니다.

2026년 기준으로 그 성능 격차는 대부분 사라졌습니다. CPU의 AES-NI, DPDK/eBPF, 멀티코어 이벤트 루프 덕에 소프트웨어 LB(nginx, HAProxy, Envoy)와 클라우드 관리형 LB가 신규 구축의 기본값이 됐습니다. 그럼에도 엔터프라이즈(특히 금융권)에 어플라이언스가 남아 있는 이유는 성능보다 **운영 체계** 쪽입니다. 검증된 Active-Standby 이중화와 세션 동기화, 벤더 기술지원과 보안 인증, 네트워크팀–개발팀의 책임 경계(장비는 네트워크팀 관할), 그리고 기존 구성의 관성입니다.

클라우드와의 대응 관계를 알아 두면 개념 이전이 쉽습니다. AWS ALB ≈ L7 프록시(SSL vserver + HTTP service와 동일 의미론), AWS NLB ≈ 패킷 모드 L4, nginx `proxy_pass` ≈ SSL 오프로드 그 자체입니다.

## 백엔드 개발자에게 체감되는 결과들

LB 뒤에서 애플리케이션을 운영할 때, 위 구조가 만들어내는 실무 현상들입니다.

- **클라이언트 IP가 SNIP로 보인다** — 풀 프록시의 직접 귀결. 실제 IP는 `X-Forwarded-For`로 받아야 하며, Tomcat이라면 RemoteIpValve, Spring이라면 ForwardedHeaderFilter로 처리합니다. 이때 **신뢰할 프록시 대역을 지정하지 않으면 클라이언트가 XFF를 위조해 IP를 속일 수 있습니다.** 패스스루 구성이라면 반대로 XFF 자체가 없습니다(평문을 못 만지므로). 그 경우 클라이언트 IP가 필요하면 Proxy Protocol 같은 별도 메커니즘이 필요합니다.
- **`request.isSecure()`가 false** — 서버는 평문 HTTP를 받았으므로 자신을 http라고 인식합니다. 애플리케이션이 "http면 https로 리다이렉트" 로직을 가지고 있으면 **무한 리다이렉트 루프**가 됩니다. `X-Forwarded-Proto`를 해석하도록 설정해야 합니다.
- **idle timeout 정합** — LB는 백엔드와의 keep-alive 연결을 풀로 유지합니다. **백엔드의 keep-alive 타임아웃은 LB의 서버측 idle 타임아웃보다 길어야 합니다.** 반대면 백엔드가 연결을 닫는 순간 LB가 그 연결로 요청을 밀어 넣는 레이스가 생겨 간헐적 connection reset이 발생합니다. LB 환경의 단골 장애 패턴입니다.
- **"연결 = 사용자" 가정 금지** — connection multiplexing이 켜져 있으면 한 백엔드 연결 위로 여러 사용자의 요청이 섞여 옵니다. 연결 단위로 인증 상태나 컨텍스트를 들고 있으면 안 됩니다. 반대로, 금융권 전문(고정길이 TCP) 인터페이스처럼 연결 자체에 의미가 있는 프로토콜은 HTTP가 아니므로 TCP/SSL_TCP 모드로 다뤄지고, 이때는 다중화 없이 연결 의미가 보존됩니다.
- **헬스체크 트래픽** — 액세스 로그에 SNIP발 프로브 요청이 주기적으로 찍힙니다. 로그 노이즈 필터링과, 헬스체크 엔드포인트가 가벼워야 한다는(매번 DB를 두드리지 않는) 설계가 필요합니다.
- **TLS 문제는 LB에서 본다** — 암호스위트 불일치, 인증서 만료, TLS 버전 거부 같은 문제는 백엔드 로그에 아무 흔적이 없습니다. 핸드셰이크가 LB에서 끝나기 때문입니다. 반대로 패스스루 구성이면 전부 백엔드 문제입니다.

## 처음 세 질문에 대한 압축 답

**Q1. 복호화해서 전달하는가? HTTP인가 TCP인가?**
설정에 달려 있습니다. 패스스루(`SSL_BRIDGE`)면 복호화 없이 암호문 바이트를 TCP로 중계하고(TLS는 클라이언트–서버 간 end-to-end), 오프로드(`SSL` vserver + `HTTP` service)면 LB가 TLS를 종단해 복호화한 뒤 평문 HTTP로, 재암호화 구성이면 새 TLS 세션으로 보냅니다. "HTTP냐 TCP냐"는 층위가 겹치는 질문이고 — HTTP도 TCP 위로 갑니다 — 실제 구분은 "바이트를 이해하고 재구성해 보내는가(L7), 모른 채 복사하는가(L4 패스스루)"입니다.

**Q2. 소켓이 실재하는가? 어떻게 평문 HTTP가 나가는가?**
실재합니다. 클라이언트 쪽 하나, 백엔드 쪽 하나. LB는 한쪽에서 accept 하고 반대쪽으로 connect 하는 프로그램이며, 복호화된 요청 바이트를 백엔드 쪽 연결에 write 하는 것이 "평문 HTTP로 보낸다"의 전부입니다. 백엔드에 도달하는 패킷은 클라이언트 패킷이 아니라 LB가 자기 연결 위에서 새로 만든 패킷입니다. 전용 장비는 OS 소켓 API 대신 자체 TCP 스택을 쓰지만 의미는 동일합니다.

**Q3. 그 장비는 뭐고, 왜 그렇게 하고, 왜 쓰는가?**
ADC(Application Delivery Controller) — 로드밸런싱에 TLS 종단·L7 라우팅·WAF 등을 묶은 어플라이언스이고, 'L4'는 90년대 제품 카테고리에서 굳은 호칭입니다. 정체는 커널 우회 스택과 암호 가속을 갖춘 컴퓨터 + 프록시 프로그램입니다. TLS를 LB에서 끊는 이유는 인증서 단일 관리와 L7 기능의 전제 조건 확보, 풀 프록시를 기본으로 하는 이유는 구간별 TCP 최적화·서버 보호·연결 다중화·공격 흡수, 전용 장비를 쓰는 이유는 (과거) 성능 격차와 (현재) 운영 체계·이중화·조직 경계입니다.

## 용어집

- **VIP (Virtual IP)** — 클라이언트가 접속하는 서비스 대표 IP. LB가 소유합니다.
- **SNIP (Subnet IP)** — NetScaler가 백엔드와 통신할 때 출발지로 쓰는 IP. 백엔드 입장에서 "클라이언트 IP"로 보이는 주소.
- **vserver / service** — NetScaler 설정 단위. vserver는 클라이언트를 받는 쪽(VIP:port + 타입), service는 백엔드 서버 하나를 가리키는 쪽. vserver 타입(`TCP`/`SSL_BRIDGE`/`SSL_TCP`/`SSL`/`HTTP`)이 동작 모드를 결정합니다.
- **persistence (스티키 세션)** — 같은 클라이언트를 같은 백엔드로 계속 보내는 기능. 소스 IP·쿠키·SSL 세션 ID 기반 등이 있고, 무엇이 가능한지는 동작 모드(평문 접근 가능 여부)가 결정합니다.
- **DSR (Direct Server Return)** — 패킷 모드에서 응답이 LB를 거치지 않고 서버에서 클라이언트로 직행하는 구성.
- **TCB (TCP Control Block)** — 연결 하나의 상태(시퀀스 번호, 윈도, 버퍼, 타이머)를 담는 자료구조. "소켓이 있다"의 실체.
- **SNI (Server Name Indication)** — TLS ClientHello에 평문으로 실리는 목적지 호스트명. 패스스루 장비가 복호화 없이 호스트 기반 라우팅을 할 수 있는 근거.
- **PFS (Perfect Forward Secrecy)** — 임시 키 교환(ECDHE) 덕에 개인키가 유출돼도 과거 트래픽을 복호화할 수 없는 성질. "중간자가 키를 얻어 복호화"하는 모델이 불가능한 이유.
- **USIP** — NetScaler에서 서버측 연결의 출발지를 SNIP 대신 클라이언트 IP로 쓰는 모드. 백엔드의 응답이 LB로 돌아오도록 라우팅 설계가 필요해져 잘 쓰지 않습니다.
- **connection multiplexing** — 다수 클라이언트 연결의 요청들을 소수의 백엔드 keep-alive 연결로 모아 보내는 기능.

## 근거와 더 읽을거리

- NetScaler 공식 문서 — SSL offloading, vserver 타입, connection multiplexing: docs.netscaler.com
- F5 — "Full Proxy Architecture" 백서 및 FastL4 프로파일 문서: f5.com
- LVS(IPVS) HOWTO — NAT/DR/TUN 모드: linuxvirtualserver.org
- Google Maglev 논문 — "Maglev: A Fast and Reliable Software Network Load Balancer" (NSDI 2016)
- Meta katran — github.com/facebookincubator/katran (XDP/eBPF 기반 L4)
- TLS 1.3 — RFC 8446 (키 교환·PFS·세션 재개)
- 관련 내부 문서: [소켓 프로그래밍](../../interviews/socket-programming.md), [네트워크 스택과 I/O 멀티플렉싱](../../interviews/os-kernel-distributed-systems-deep-dive/01d_network_stack_and_io_multiplexing.md), [nginx](../../web/nginx/nginx.md), [DDoS](../../security/ddos.md)
