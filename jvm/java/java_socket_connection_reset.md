# Java Socket Connection Reset

## 개요

Java에서 `java.net.SocketException: Connection reset`가 보이면, 먼저 Java API 이름보다 **TCP에서 누가 `RST`를 보냈는지**를 의심해야 합니다.

이 문서는 `Socket`, `ServerSocket`, `NIO`, `Netty` 같은 자바 소켓 전체를 설명하는 문서가 아닙니다. 여기서 다루는 질문은 더 좁습니다.

- Spring이나 Java 애플리케이션에서 `Connection reset`이 보일 때 실제로 무슨 일이 일어났는가
- 왜 서버 문제, 로드 밸런서 문제, TLS 문제, stale keep-alive 문제가 같은 예외로 보일 수 있는가
- 어디서부터 확인해야 direct cause와 root cause를 구분할 수 있는가
- 어떤 수정이 왜 통하는가

목차:

1. `Connection reset`에서 실제로 일어나는 일
2. Java와 Spring에서는 어떻게 보이는가
3. direct cause와 root cause를 나눠서 봐야 하는 이유
4. 가장 흔한 원인군
5. 실제 확인 순서
6. 수정이 통하는 이유
7. 자주 헷갈리는 오류 비교
8. replay와 검증 방법

## `Connection reset`에서 실제로 일어나는 일

`Connection reset`의 첫 의미는 단순합니다. **상대가 정상 종료(`FIN`) 대신 강제 종료(`RST`)를 보냈다**는 뜻입니다.

TCP 연결을 정상적으로 닫을 때는 보통 `FIN`이 오가고, 양쪽은 "남은 데이터는 여기까지다"라는 식으로 연결을 정리합니다. 반면 `RST`는 "이 연결은 더 이상 유효한 연결로 취급하지 않겠다"에 가깝습니다. 그래서 상대는 다음처럼 느낍니다.

```text
클라이언트                    서버 또는 중간 장비
    |  요청 전송 ---------------------> |
    |                                   |
    | <----------- TCP RST ------------ |
    |                                   |
    |  Java read/write 실패             |
    |  SocketException("Connection reset")
```

여기서 중요한 점은 두 가지입니다.

1. Java 예외는 **원인**이 아니라 **표면 증상**입니다.
2. `RST`를 보낸 주체는 항상 애플리케이션 서버라고 단정할 수 없습니다.

서버 프로세스일 수도 있고, 서버 앞단의 로드 밸런서나 방화벽일 수도 있고, 클라이언트 자신의 커널이 비정상 종료를 처리하면서 보냈을 수도 있습니다. 그래서 `Connection reset`를 해석할 때 제일 먼저 물어야 할 질문은 "`누가` reset을 보냈는가"입니다.

### `RST`와 `FIN`을 먼저 구분해야 하는 이유

이 둘을 구분하지 않으면 "연결이 끊겼다"는 말 하나로 너무 많은 상황을 섞게 됩니다.

- `FIN`
  - 정상 종료에 가깝습니다.
  - 상대는 EOF를 읽거나, 더 읽을 데이터가 없다는 식으로 관측하는 경우가 많습니다.
- `RST`
  - 강제 종료입니다.
  - 읽기/쓰기 도중 즉시 에러로 보이기 쉽습니다.
  - 애플리케이션 입장에서는 "이 연결은 정상적으로 끝난 것이 아니라 중간에 깨졌다"는 신호에 가깝습니다.

즉 `Connection reset`를 보면 "연결이 그냥 끝났다"가 아니라, "정상 종료가 아니라 abort 쪽 사건이 일어났다"로 이해해야 합니다.

## Java와 Spring에서는 어떻게 보이는가

Java는 TCP 레벨의 abort를 그대로 `RST`라는 이름으로 노출하지 않고, 보통 `SocketException` 계열 예외로 보여 줍니다. 그래서 애플리케이션 코드나 로그에서는 보통 아래처럼 보입니다.

```text
java.net.SocketException: Connection reset
```

Spring에서는 이것이 한 번 더 감싸져 보이는 경우가 많습니다.

```text
org.springframework.web.client.ResourceAccessException:
I/O error on POST request for "...":
Connection reset; nested exception is java.net.SocketException: Connection reset
```

같은 사건을 계층별로 다시 쓰면 보통 아래 흐름에 가깝습니다.

1. 서버, 로드 밸런서, 방화벽, 또는 클라이언트 반대편 커널이 TCP `RST`를 보냅니다.
2. 로컬 커널은 그 연결에 대한 다음 `read` 또는 `write`를 실패로 처리합니다.
3. 운영체제 수준에서는 `ECONNRESET` 같은 네트워크 오류로 관측될 수 있습니다.
4. JDK는 이것을 `SocketException` 계열로 올립니다.
5. Spring client는 다시 이를 `ResourceAccessException` 같은 상위 예외로 감쌀 수 있습니다.

이때 중요한 것은 **예외를 읽는 위치가 아니라, reset이 발생한 시점**입니다. 같은 `Connection reset`라도 아래처럼 의미가 달라집니다.

- 연결 재사용 직후 첫 요청에서 터진다
  - stale keep-alive를 먼저 의심할 만합니다.
- TLS 핸드셰이크 도중 터진다
  - TLS 정책 불일치나 중간 장비 개입 가능성이 더 큽니다.
- 서버 처리 시간이 길어질 때만 터진다
  - 서버 앞단 timeout이나 서버 측 abort 가능성이 커집니다.
- 서버 로그에는 `Connection reset by peer`가 찍힌다
  - 방향을 바꿔 읽어야 합니다. 이 경우는 서버가 "클라이언트가 끊었다"를 본 것일 수 있습니다.

즉 같은 예외명이라도, **언제 끊겼는지**와 **누가 그 방향의 reset을 봤는지**를 같이 봐야 합니다.

## direct cause와 root cause를 나눠서 봐야 하는 이유

`Connection reset` troubleshooting이 자주 꼬이는 이유는, direct cause와 root cause를 한 줄로 합쳐 버리기 때문입니다.

direct cause는 더 가깝고 관측 가능한 사건입니다.

- 어떤 TCP peer가 `RST`를 보냈다
- 커넥션 풀에서 죽은 연결을 재사용했다
- TLS 핸드셰이크가 끝나기 전에 연결이 abort됐다

root cause는 왜 그런 direct cause가 생겼는지에 대한 더 깊은 설명입니다.

- 로드 밸런서 idle timeout이 서버나 클라이언트보다 짧았다
- 서버가 과부하나 재시작으로 기존 소켓을 갑자기 정리했다
- TLS 버전, cipher suite, SNI, ALPN 정책이 맞지 않았다
- 방화벽이 세션을 정리하거나 트래픽을 차단했다
- 클라이언트가 읽지 않은 데이터를 남긴 채 소켓을 닫았다

이 구분이 중요한 이유는, direct cause를 모르면 **어디를 봐야 할지**를 못 정하고, root cause를 모르면 **왜 그 수정이 통하는지**를 설명하지 못하기 때문입니다.

예를 들어 "재시도하면 되더라"는 direct cause도 root cause도 아닙니다. 그것은 완화책입니다. 왜 재시도가 먹히는지까지 내려가야 문서가 닫힙니다.

## 가장 흔한 원인군

여기서는 Java/Spring 애플리케이션에서 특히 자주 만나는 원인군만 추립니다.

### 1. stale keep-alive, 즉 죽은 연결 재사용

이 경우가 실무에서 가장 흔합니다.

HTTP client는 성능 때문에 TCP 연결을 풀에 넣어 두고 재사용합니다. 그런데 서버나 로드 밸런서가 유휴 연결을 먼저 정리했는데, 클라이언트는 그 사실을 아직 모를 수 있습니다. 그러면 풀에서 꺼낸 연결은 자바 객체 관점에서는 "재사용 가능한 연결"처럼 보이지만, 네트워크 관점에서는 이미 죽은 연결입니다.

그래서 보통 이런 패턴이 나옵니다.

1. 요청 A 후 연결이 pool에 남아 있습니다.
2. 서버 또는 LB가 idle timeout으로 그 연결을 정리합니다.
3. 클라이언트는 그 사실을 모른 채 같은 연결을 다시 꺼냅니다.
4. 요청 B를 보내는 순간 `RST`가 돌아옵니다.

이 패턴의 특징은 **오랫동안 놀고 있던 뒤 첫 요청이 실패하고, 직후 재시도는 성공할 수 있다**는 점입니다. 첫 요청이 stale connection을 밟아 죽고, 두 번째 요청은 새 연결을 맺기 때문입니다.

### 2. TLS 핸드셰이크 또는 정책 불일치

클라이언트와 서버가 TCP 연결까지는 만들었지만, TLS 버전이나 암호군, SNI, ALPN 같은 정책이 맞지 않으면 핸드셰이크 단계에서 연결이 abort될 수 있습니다.

이 경우의 특징은 다음과 같습니다.

- HTTP 요청 바디를 보내기도 전에 실패할 수 있습니다.
- Java에서는 `SocketException` 앞뒤로 SSL 관련 스택이 보일 수 있습니다.
- `curl`과 Java의 결과가 다를 수 있습니다.
  - 같은 서버라도 TLS 기본 설정이 다르기 때문입니다.

즉 "Java에서만 깨진다"는 사실은 자바 라이브러리 버그를 뜻하기보다, **클라이언트별 TLS 정책 차이**를 먼저 의심하게 만드는 신호일 수 있습니다.

### 3. 서버 재시작, 크래시, 과부하, 처리 시간 초과

서버 프로세스가 내려가거나, 처리 중 timeout이 나거나, 앞단 프록시가 더 못 기다리겠다고 판단하면 열린 연결을 강제로 정리할 수 있습니다.

이 경우는 stale keep-alive와 달리 **특정 시간대에 집중적으로** 발생하거나, 서버 로그에 OOM, restart, upstream timeout 같은 다른 신호가 함께 남는 경우가 많습니다.

즉 이 경우의 root cause는 "reset 그 자체"가 아니라, **왜 서버 측에서 연결을 유지하지 못했는가**입니다.

### 4. 로드 밸런서, 프록시, 방화벽 같은 중간 장비

애플리케이션 개발자는 종종 "클라이언트 <-> 서버"만 생각하지만, 실제 운영 경로에는 그 사이에 여러 장비가 끼어 있습니다.

그래서 direct cause가 "서버가 reset을 보냈다"가 아닐 수도 있습니다. 예를 들어:

- 로드 밸런서 idle timeout이 더 짧다
- 방화벽이 세션을 정리했다
- 보안 장비가 특정 요청 패턴을 차단했다

이 경우 애플리케이션 로그만 보면 서버가 원인처럼 보일 수 있지만, 패킷 캡처를 해 보면 reset의 실제 발신 IP가 서버가 아니라 LB인 경우가 있습니다.

### 5. 클라이언트 측 비정상 종료 또는 잘못된 종료 절차

덜 흔하지만 방향을 바꿔 보면 이것도 중요합니다.

클라이언트 프로세스가 죽거나, 읽지 않은 데이터를 남긴 채 비정상적으로 소켓을 닫으면 **서버 쪽에서** `Connection reset by peer` 비슷한 증상을 볼 수 있습니다.

즉 한쪽에서 본 `Connection reset`는 상대방이 문제라는 뜻이 아니라, **상대 또는 그 경로 어딘가가 이 연결을 abort했다**는 뜻입니다.

## 실제 확인 순서

`Connection reset`를 보면 보통 모두가 설정값부터 뒤지기 쉽습니다. 하지만 가장 강한 순서는 "무슨 계층에서, 어느 순간, 누가 끊었는가"를 좁혀 가는 순서입니다.

### 1. 먼저 끊기는 시점을 분류합니다

다음 네 가지 중 어디에 가까운지 먼저 고릅니다.

1. 연결 직후 곧바로 실패한다
2. TLS 핸드셰이크 중 실패한다
3. 유휴 시간이 길었던 뒤 첫 요청에서만 실패한다
4. 응답을 받는 도중 또는 서버가 오래 처리할 때 실패한다

이 분류만 해도 후보 원인이 많이 줄어듭니다.

- `3`에 가깝다
  - stale keep-alive를 먼저 의심합니다.
- `2`에 가깝다
  - TLS 정책 불일치를 먼저 의심합니다.
- `4`에 가깝다
  - 서버 timeout, upstream timeout, 재시작, 과부하를 먼저 의심합니다.

### 2. 같은 요청을 다른 클라이언트로 비교합니다

`curl -v`, 브라우저, 혹은 다른 Java client로 같은 엔드포인트를 호출해 봅니다.

이 비교가 중요한 이유는 "서버가 항상 틀렸다" 또는 "Java가 항상 틀렸다"를 가르는 것이 아니라, **정책 차이와 재사용 차이**를 가르는 데 도움이 되기 때문입니다.

- `curl`은 성공하고 Java만 실패한다
  - Java 쪽 TLS 설정, connection pool, keep-alive 재사용을 먼저 봅니다.
- 모두 실패한다
  - 서버나 중간 장비 가능성이 더 커집니다.

### 3. 서버 로그와 LB 로그를 시간축으로 맞춰 봅니다

이때는 단순히 "에러가 있었다"가 아니라, **같은 시각에 무슨 일이 있었는지**를 봐야 합니다.

- 애플리케이션 예외
- 웹 서버 / LB timeout
- 서버 restart
- OOM
- upstream connection closed

같은 시각에 이런 사건이 겹치면 root cause 후보가 많이 좁혀집니다.

### 4. stale keep-alive 여부를 확인합니다

이 축은 별도로 확인할 가치가 큽니다. 특히 RestTemplate + Apache HttpClient, 혹은 커넥션 풀을 쓰는 환경이라면 더 그렇습니다.

다음 패턴이면 의심 강도가 높습니다.

- 평소엔 괜찮다가 한동안 idle 후 첫 요청만 실패
- 바로 재시도하면 성공
- 특정 배포 환경에서만 간헐적으로 발생
- LB idle timeout과 client idle eviction이 따로 놀고 있음

이 경우는 단순 timeout 증가보다, **왜 죽은 연결을 재사용하게 되었는지**를 먼저 봐야 합니다.

### 5. 패킷 캡처로 `누가` reset을 보냈는지 확인합니다

이 단계가 가장 강한 증거입니다.

`tcpdump`나 Wireshark로 보면 다음 질문에 답할 수 있습니다.

- reset 발신 IP가 서버인가, LB인가, 다른 장비인가
- reset이 TLS handshake 전에 왔는가, 후에 왔는가
- reset 전에 FIN이 있었는가, 아니면 정말 abrupt abort였는가

애플리케이션 로그만 보면 "서버가 끊었다"고 읽기 쉬운 상황도, 패킷 캡처를 보면 사실은 중간 장비가 끊은 경우가 있습니다.

## 수정이 통하는 이유

좋은 troubleshooting 문서는 "이 설정을 넣어라"로 끝나지 않습니다. 왜 그 수정이 통하는지를 같이 알아야 다음 문제에도 전이됩니다.

### 클라이언트의 idle 정리를 서버/LB보다 더 짧게 두는 이유

stale keep-alive의 핵심은 **상대가 먼저 연결을 죽였는데 나는 아직 살아 있다고 믿는 상태**입니다. 그러니 해결 방향은 단순합니다.

- 서버나 LB가 죽이기 전에
- 클라이언트가 먼저 그 유휴 연결을 버리게 만들면 됩니다

그래서 `evictExpiredConnections`, `evictIdleConnections`, `validateAfterInactivity` 같은 설정이 통합니다. 이 설정들의 목적은 성능 튜닝이 아니라, **죽은 연결을 재사용하지 않게 만드는 것**입니다.

예를 들어 Apache HttpClient 계열에서는 보통 아래처럼 커넥션 매니저와 idle eviction을 같이 둡니다.

```java
var cm = PoolingHttpClientConnectionManagerBuilder.create()
    .setValidateAfterInactivity(TimeValue.ofSeconds(2))
    .build();

var client = HttpClients.custom()
    .setConnectionManager(cm)
    .evictExpiredConnections()
    .evictIdleConnections(TimeValue.ofSeconds(30))
    .build();
```

핵심은 숫자 자체가 아니라 관계입니다. 서버나 로드 밸런서가 유휴 연결을 60초에 정리한다면, 클라이언트는 그보다 더 이른 시점에 해당 연결을 버리는 편이 안전합니다.

### TLS 설정을 맞추는 이유

TLS 불일치 문제는 결국 "서버가 이 조건으로는 대화를 계속하지 않겠다"는 사건입니다. 그러니 수정 방향은 timeout 증가가 아니라 **정책 정합**입니다.

- TLS 버전
- cipher suite
- SNI
- ALPN
- 인증서 체인과 신뢰 저장소

즉 handshake 단계 reset에는 pool eviction보다 TLS 증거를 먼저 보는 편이 맞습니다.

### timeout을 늘리는 것이 맞는 경우와 아닌 경우

서버가 실제로 정상적인 긴 작업을 하고 있는데 중간 timeout이 너무 짧아서 reset이 난다면 timeout 조정이 맞습니다.

반대로 이미 죽은 keep-alive를 재사용하는 문제가 핵심이라면 timeout을 늘려도 근본 해결이 아닙니다. 죽은 연결은 오래 기다린다고 살아나지 않기 때문입니다.

### retry가 통할 때와 조심해야 할 때

retry는 완화책이지 만능 해결책이 아닙니다.

- stale keep-alive라면 첫 요청이 죽고 두 번째는 새 연결로 성공할 수 있으므로 retry가 잘 먹힐 수 있습니다.
- 하지만 POST처럼 부작용이 있는 요청은 무조건 blind retry 하면 안 됩니다.

즉 retry를 넣더라도 다음이 같이 필요합니다.

- 이 요청이 idempotent한가
- 서버가 이미 일부 처리했을 가능성은 없는가
- request ID나 중복 방지 키가 있는가

## 자주 헷갈리는 오류 비교

### `Connection reset` vs `timeout`

- `Connection reset`
  - 누군가 연결을 **강제로 abort**했습니다.
  - 보통 `RST` 사건이 중심입니다.
- `timeout`
  - 정해진 시간 안에 기대한 응답이 오지 않았습니다.
  - 적극적인 abort가 없을 수도 있습니다.

즉 reset은 "누가 끊었다"에 가깝고, timeout은 "기다렸는데 안 왔다"에 가깝습니다.

### `Connection reset` vs `Connection refused`

- `Connection refused`
  - 연결 수립 단계에서 상대 포트가 열려 있지 않다는 뜻에 가깝습니다.
  - 아예 accept할 대상이 없을 때 주로 나옵니다.
- `Connection reset`
  - 연결은 수립됐거나, 적어도 더 진행된 뒤에 깨졌을 가능성이 큽니다.

즉 refused는 "문 앞에서 거절", reset은 "대화 도중 강제 종료"에 더 가깝습니다.

### `Connection reset` vs `FIN` 이후 EOF

- EOF
  - 정상 종료 경로일 가능성이 큽니다.
- reset
  - 정상 종료보다 abort 쪽입니다.

이 둘을 같은 "연결 종료"로 뭉개면 direct cause가 흐려집니다.

## replay와 검증 방법

이 문서의 핵심 주장을 실제로 확인하려면 아래 순서를 따라가면 됩니다.

### 1. 요청 시점과 실패 시점을 고정합니다

- 같은 엔드포인트를 연속 두 번 호출합니다.
- 중간에 LB idle timeout보다 길게 쉬어 봅니다.
- 첫 요청만 실패하고 두 번째는 성공하는지 봅니다.

PASS 신호:

- idle 뒤 첫 요청만 reset
- 즉시 재시도는 성공

이 경우 stale keep-alive 가능성이 높습니다.

### 2. `curl -v`로 같은 엔드포인트를 비교합니다

```bash
curl -v https://example.com
```

PASS 신호:

- Java만 실패하고 `curl`은 안정적으로 성공

이 경우 TLS 정책 차이나 Java client 재사용 정책을 먼저 좁혀 볼 가치가 있습니다.

### 3. TLS 핸드셰이크를 의심하면 Java SSL 디버그를 켭니다

```bash
-Djavax.net.debug=ssl:handshake
```

PASS 신호:

- handshake 단계에서 바로 끊기는 지점이 보인다

이 경우는 keep-alive보다 TLS 축을 먼저 봐야 합니다.

### 4. 패킷 캡처로 reset 발신자를 확인합니다

```bash
sudo tcpdump -nn -i any 'tcp port 443 and (tcp[tcpflags] & tcp-rst != 0)'
```

PASS 신호:

- 어느 IP가 reset을 보냈는지 보인다

이 단계까지 가면 "서버가 끊었는가, LB가 끊었는가"를 감으로 말할 필요가 줄어듭니다.

## 실무용 정리

`Connection reset`를 보면 다음 순서로 생각하는 편이 가장 안전합니다.

1. Java 예외명보다 먼저 TCP `RST` 사건으로 해석합니다.
2. 실패 시점이 handshake인지, idle reuse 직후인지, 응답 도중인지 분류합니다.
3. direct cause와 root cause를 분리합니다.
4. stale keep-alive, TLS 정책 불일치, 서버 abort, middlebox 개입을 주요 후보로 둡니다.
5. 패킷 캡처와 시간축 로그로 `누가` reset을 보냈는지 확인합니다.
6. retry는 부작용과 idempotency를 확인한 뒤 제한적으로 씁니다.

이 순서를 몸에 익히면, 다음에 `Connection reset`이 다시 나와도 "네트워크가 이상하다" 수준에서 머무르지 않고 어디서부터 의심해야 하는지 바로 잡을 수 있습니다.

## 관련 문서

- [Java HTTP](/Users/rody/VscodeProjects/study/jvm/java/java_http.md)
- [Linux errno](/Users/rody/VscodeProjects/study/linux/errors/errno.md)
- [cURL Errors](/Users/rody/VscodeProjects/study/linux/commands/curl/curl_errors.md)

## 참고 자료

- [RFC 9293: Transmission Control Protocol (TCP)](https://www.rfc-editor.org/rfc/rfc9293)
- [SocketException (Oracle Javadoc)](https://docs.oracle.com/javase/8/docs/api/java/net/SocketException.html)
- [recv(2) - Linux manual page](https://man7.org/linux/man-pages/man2/recv.2.html)
- [Apache HttpClient connection pooling](https://hc.apache.org/httpcomponents-client-5.6.x/connection-pooling.html)
- [HttpClientBuilder evictIdleConnections](https://hc.apache.org/components/httpcomponents-client-5.3.x/5.3.1/httpclient5/apidocs/org/apache/hc/client5/http/impl/classic/HttpClientBuilder.html)
- [PoolingHttpClientConnectionManagerBuilder.setValidateAfterInactivity](https://hc.apache.org/components/httpcomponents-client-5.1.x/5.1.4/httpclient5/apidocs/org/apache/hc/client5/http/impl/io/PoolingHttpClientConnectionManagerBuilder.html)
