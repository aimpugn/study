# mod_jk 페일오버 — 백엔드 WAS 한 대가 죽으면 요청은 어떻게 되나

## 짧게 답하면

mod_jk 로드밸런서의 페일오버는 **같은 요청 처리 안에서** 일어납니다 — 멤버 하나가 실패하면 그 요청을 다음 멤버로 재시도하므로, 두 대 중 한 대가 죽어도 클라이언트는 대개 정상 응답을 받습니다. 즉시 넘어가느냐 한참 매달리느냐는 **백엔드가 어떻게 죽었느냐**가 결정합니다. 프로세스만 죽어 포트가 닫히면(커널이 RST로 즉시 거부) 밀리초 수준에서 페일오버되고, 호스트째 사라져 패킷이 증발하면 mod_jk 기본 설정에는 자체 타임아웃이 전혀 없어서 OS의 TCP connect 타임아웃(수 초~약 2분)까지 그 요청이 매달립니다. 배포를 위한 kill은 전자에 해당합니다.

이 문서는 mod_jk 1.2.x 기준이며, 시간 수치 중 실측이 아닌 것은 계산 근거를 함께 적습니다.

## 무대 — workers.properties와 기본값

밸런서 하나(멤버 둘) + 상태 워커의 최소 구성입니다. 실제 구성이 밸런서 여러 개라도 각 밸런서는 서로 독립적으로 아래 동작을 반복합니다.

```properties
worker.list=apigw-lb,jk-status

worker.apigw-lb.type=lb                                  # 로드밸런서 워커
worker.apigw-lb.balance_workers=apigw-prod01,apigw-prod02
worker.apigw-lb.sticky_session=1                         # JSESSIONID의 jvmRoute로 고정 라우팅

worker.apigw-prod01.type=ajp13
worker.apigw-prod01.host=1.2.3.11
worker.apigw-prod01.port=12209
worker.apigw-prod01.lbfactor=1                           # 1:1 가중치 -> 평시 절반씩 분배

worker.apigw-prod02.type=ajp13
worker.apigw-prod02.host=1.2.3.12
worker.apigw-prod02.port=12209
worker.apigw-prod02.lbfactor=1

worker.jk-status.type=status                             # 런타임 상태 조회·조작 화면
```

이 구성에는 타임아웃·프로브 관련 속성이 하나도 없습니다. 그래서 아래 기본값들이 시나리오 전체를 지배합니다.

| 속성 | 기본값 | 뜻 | 어느 장면에 등장하나 |
| --- | --- | --- | --- |
| `socket_timeout` | 0 | 소켓 I/O 무제한 대기 | 시나리오 B의 매달림 |
| `socket_connect_timeout` | `socket_timeout*1000` = 0 | connect 단계 상한 없음 → OS에 위임 | 시나리오 B의 매달림 |
| `ping_mode` | 꺼짐 | CPing/CPong 생존 확인 안 함 (1.2.27+에서 설정 가능) | 처방 |
| `retries` / `retry_interval` | 2 / 100ms | 실패 시 재시도 횟수·간격 | 같은 요청 내 재시도 |
| `recover_time` | 60초 | ERROR 멤버를 다시 시도하기까지의 최소 시간 | 회복 프로브 |
| `worker.maintain` | 60초 | 상태 점검(글로벌 maintenance) 주기 | 회복 프로브 |
| `sticky_session_force` | false | 고정 대상이 ERROR면 에러 대신 다른 멤버로 | 세션의 운명 |
| `recovery_options` | 0 | 가능한 한 끝까지 다른 멤버로 재전송 | 처리 중이던 요청의 운명 |

## 페일오버가 작동하는 구조

```text
[httpd 자식 프로세스들]                [공유 메모리 (JkShmFile)]
  child#1 ── AJP keep-alive 풀 ──┐      멤버 상태: OK / ERR / REC ...
  child#2 ── AJP keep-alive 풀 ──┼──→   한 자식이 ERROR를 기록하면
  child#N ── AJP keep-alive 풀 ──┘      모든 자식이 즉시 같은 판단을 공유
        │
        ├──────→ apigw-prod01:12209 (AJP)
        └──────→ apigw-prod02:12209 (AJP)
```

각 httpd 자식은 백엔드별 keep-alive AJP 커넥션 풀을 따로 갖지만, 멤버의 ERROR/OK 상태는 공유 메모리로 전 프로세스가 공유합니다. 그래서 장애 감지는 자식 하나가 하면 충분하고, 죽은 keep-alive 커넥션의 정리는 각 자식이 다음 사용 시점에 따로 합니다.

## 시나리오 A — 프로세스만 죽는 경우 (배포 kill)

머신은 살아 있으므로 커널이 즉시 응답해 줍니다. 닫힌 포트로의 SYN은 RST(connection refused)로 거부되고, 풀에 있던 keep-alive 커넥션에는 FIN이 도착해 있습니다. 시간 순서로 따라갑니다.

### T0 — kill 순간 처리 중이던 요청들

유일하게 실제 피해가 생길 수 있는 구간입니다. lbfactor 1:1이면 그 순간 동시 처리량의 대략 절반이 해당합니다. mod_jk는 응답을 기다리던 read에서 즉시 EOF/RST를 받고, `recovery_options=0`에 따라 세 갈래로 갈립니다.

1. **응답이 한 바이트도 안 나갔고, 요청 본문이 없거나 첫 AJP 패킷(약 8KB) 안에 끝난 경우** — 다른 멤버로 통째 재전송되어 클라이언트는 정상 응답을 받습니다. mod_jk가 첫 본문 패킷을 재전송 버퍼에 보관하기 때문에 가능합니다(소스의 `op->post`/`reco_buf` 버퍼). 단 죽은 백엔드가 그 요청을 이미 일부 실행했다면(DB 반영 후 응답 직전 사망) **비멱등 요청은 이중 처리**됩니다.
2. **본문 추가 청크를 이미 스트리밍한 업로드** — 클라이언트로부터 다시 읽을 수 없으므로 복구 불가로 마킹되고, 클라이언트는 에러를 받습니다.
3. **응답 일부가 이미 클라이언트로 나간 경우** — 마찬가지로 복구 불가. 클라이언트는 잘린 응답을 받습니다.

### T0+ε — 죽은 멤버로 배정되는 첫 신규 요청

mod_jk는 풀의 커넥션을 재사용하기 전에 **요청을 쓰기 전 소켓 생존 검사**를 합니다(poll로 EOF 수신 여부 확인 — 소스의 `jk_is_socket_connected`). FIN을 받아 둔 죽은 커넥션은 여기서 걸러져 폐기되고, 새 connect 시도는 즉시 RST로 거부됩니다. 그 시점에 멤버가 ERROR로 마킹되고(공유 메모리로 전파), **같은 요청이 다른 멤버로 재시도되어 정상 응답**합니다. 이 요청의 체감 지연은 밀리초 수준입니다.

### ERROR 상태 동안 — 세션의 운명

신규 요청은 전부 살아 있는 멤버로만 갑니다. `sticky_session=1`이어도 `sticky_session_force`가 기본값(false)이므로, 죽은 멤버에 고정돼 있던 세션의 요청 역시 에러가 아니라 다른 멤버로 넘어갑니다 — 공식 문서 표현으로 "세션을 잃는 대신 페일오버". 톰캣 세션 복제가 없으면 사용자는 재로그인하게 되고, 새 세션은 살아남은 멤버의 jvmRoute로 발급되어 그쪽에 고정됩니다. force를 true로 바꾸면 페일오버 대신 클라이언트가 에러를 받으므로, 세션 보존이 복제로 보장되지 않는 한 기본값이 맞습니다.

### T0+60초~ — 회복 프로브

`recover_time`(60초)이 지나면 글로벌 maintenance가 ERROR 멤버를 회복 시도 상태(REC)로 바꾸고, **다음 실제 사용자 요청 하나가 프로브로 사용**됩니다. 백그라운드 헬스체크가 아닙니다. 시나리오 A에서는 포트가 닫혀 있어 프로브도 즉시 RST를 받으므로 비용이 사실상 없고, WAS가 재기동을 마치면 다음 프로브가 성공해 멤버가 OK로 복귀합니다.

## 시나리오 B — 호스트째 사라지는 경우

전원 차단, 커널 패닉, 네트워크 단절처럼 **RST조차 돌아오지 않는** 죽음입니다. 페일오버 논리는 시나리오 A와 같지만, "실패를 인지하기까지"가 달라집니다.

connect 시도가 OS의 TCP 타임아웃까지 매달립니다. 같은 L2 세그먼트라면 ARP 응답 실패로 약 3초 만에 EHOSTUNREACH가 나지만(기본 ARP 재시도 3회 × 1초), 중간 장비가 패킷을 조용히 버리는 경우 리눅스 기본 `tcp_syn_retries=6` 기준 **약 2분**을 다 기다립니다 — 현행 커널 문서 기준 최종 타임아웃 약 131초(선형 재시도 `tcp_syn_linear_timeouts=4` 반영), 순수 지수 백오프를 쓰던 구형 커널 계산으로는 약 127초(1+2+4+8+16+32+64). 한편 중간 장비가 DROP이 아니라 REJECT(ICMP 회신)로 막으면 connect가 즉시 실패해 시나리오 A에 가깝게 동작합니다 — 긴 매달림은 조용히 버리는 블랙홀일 때의 이야기입니다. 매달린 요청은 그동안 멈춰 있다가, 실패 확정 후에야 ERROR 마킹과 페일오버가 일어납니다 — 클라이언트나 앞단 장비가 먼저 끊지 않았다면 결국 응답은 받습니다.

더 아픈 것은 그 다음입니다. 회복 프로브가 실제 요청을 제물로 쓰므로, 호스트가 복구될 때까지 **약 1분에 한 건씩 OS 타임아웃만큼 매달리는 요청이 주기적으로 발생**합니다.

| 축 | A. 프로세스 kill | B. 호스트 다운/블랙홀 |
| --- | --- | --- |
| 커널의 응답 | RST 즉시 | 없음 (또는 지연된 unreachable) |
| 첫 요청의 운명 | ms 수준 지연 후 페일오버 | 3초~약 2분 매달린 후 페일오버 |
| 회복 프로브 비용 | 0에 가까움 (즉시 거부) | 60초마다 한 건씩 긴 매달림 |
| 운영 체감 | 거의 무중단 | 주기적인 슬로우 요청 민원 |
| 처방 | 드레인 절차로 충분 | `socket_connect_timeout`·`ping_mode` 필수 |

## 시나리오 C — graceful shutdown이 끼어들 때

배포 스크립트는 보통 "graceful 시도 → 안 죽으면 kill"입니다. 그런데 graceful 구간의 동작이 WAS마다 달라서, 일부는 즉사보다 사용자 체감이 나쁩니다.

| WAS | 종료 중 신규 요청 | 함정 | 대응 |
| --- | --- | --- | --- |
| Spring Boot 2.3+ (`server.shutdown=graceful`) | HTTP 주 커넥터는 리슨 소켓까지 닫혀 즉시 거부. **AJP 추가 커넥터는 accept만 멈추고 소켓이 살아남음** | AJP로 붙는 mod_jk의 connect가 **성공해 버리고** 응답 없이 매달림 — drain 타임아웃(기본 30초)까지 | AJP `Connector`에 `bindOnInit=false` 명시(아래 메커니즘 참고), 보조로 `socket_connect_timeout`·`ping_mode` |
| WildFly / EAP 7 (suspend) | 새 요청에 503을 **정상 응답으로** 반환 | mod_jk는 HTTP 상태로는 페일오버하지 않으므로 503이 사용자에게 그대로 전달 | `fail_on_status=503` 또는 드레인 후 종료 |
| 구형 JBoss (JBossWeb 계열) | 커넥터 리스너를 먼저 닫음 → 즉시 RST | 거의 없음 — 시나리오 A와 동일하게 흘러감 | — |

Spring Boot 행의 HTTP/AJP 비대칭은 Boot·톰캣 소스에서 확인한 동작입니다. Boot의 GracefulShutdown은 모든 커넥터에 `pause()`와 `closeServerSocketGraceful()`을 호출하는데, 톰캣의 `pause()`는 리슨 소켓을 닫지 않고 `closeServerSocketGraceful()`은 **start 시점에 바인딩된(BOUND_ON_START) 소켓만** 닫습니다. Boot는 주 커넥터에 `bindOnInit=false`를 자동 설정해 이 조건을 만들어 주지만, `addAdditionalTomcatConnectors()`로 추가한 AJP 커넥터에는 커넥터 커스터마이즈가 적용되지 않습니다 — 톰캣 기본값 `bindOnInit=true`로 init 시점에 바인딩(BOUND_ON_INIT)되어 graceful 중에도 리슨 소켓이 열려 있고, 커널은 backlog 한도까지 3-way handshake를 완성해 둡니다. mod_jk의 connect는 성공하지만 응답은 오지 않다가, drain이 끝나 소켓이 완전히 닫힐 때에야 RST를 받고 페일오버합니다. AJP `Connector`를 만들 때 `bindOnInit=false`를 함께 주면 HTTP 주 커넥터와 똑같이 즉시 거부(fail-fast)로 바뀝니다.

재기동 쪽에도 비대칭이 있습니다. WildFly는 AJP 리스너가 애플리케이션 배포 완료 전에 열릴 수 있어 복귀 직후 잠깐 404/500이 정상 응답으로 중계될 수 있습니다(mod_jk 입장에선 멀쩡한 응답이라 페일오버하지 않습니다). Spring Boot는 컨텍스트가 준비된 뒤 포트를 열고, standalone Tomcat도 커넥터를 마지막에 시작하므로 이 문제가 없습니다.

`fail_on_status=503`을 줄 때는 애플리케이션이 업무 응답으로 503을 쓰는 경로가 없는지 먼저 확인합니다 — 있다면 정상 요청이 페일오버·재전송되는 부작용이 생깁니다.

## 처방 — 설정과 절차

### 설정: 매달림 제거와 레이스 축소

```properties
worker.apigw-prod01.ping_mode=A                  # 연결 직후(C)·요청 직전(P)·유휴(I) CPing/CPong 전부 — 1.2.27+
worker.apigw-prod01.ping_timeout=10000           # CPong 대기 상한 (기본 10초, 문서 권장 5~15초)
worker.apigw-prod01.socket_connect_timeout=3000  # connect 단계 상한 — 시나리오 B의 127초 매달림을 3초로 (문서 권장 1~5초)
# prod02 및 다른 밸런서 멤버에도 동일하게
```

httpd 쪽에는 `JkWatchdogInterval 60`을 줍니다. 상태 점검(maintenance)이 사용자 요청에 편승하지 않고 백그라운드 스레드에서 돌게 됩니다. `reply_timeout`은 응답이 오래 걸리는 정상 요청(배치성 조회 등)까지 실패 처리할 수 있으므로, 쓴다면 `max_reply_timeouts`와 함께 신중하게 잡습니다.

`recovery_options`는 트레이드오프입니다. 기본값 0은 가용성 우선(가능한 한 재전송 — 이중 처리 위험 감수), `3`(1+2)은 백엔드가 요청을 받은 뒤에는 재전송하지 않는 멱등성 우선(그 케이스의 클라이언트는 에러를 받음)입니다. 결제처럼 이중 처리가 치명적인 서비스만 올리는 것이 합리적입니다.

### 절차: 드레인 배포

시나리오 A의 피해(처리 중이던 요청, 세션 유실)는 전부 "트래픽이 살아 있는 멤버를 죽여서" 생깁니다. 죽이기 전에 빼면 사라집니다.

1. jk-status 화면에서 대상 멤버의 activation을 **Disabled**로 — 신규 세션은 안 받고, 기존 sticky 세션의 요청만 계속 처리합니다(드레인).
2. jk-status의 busy 카운트가 빠진 것을 확인한 뒤 WAS를 종료하고 배포합니다.
3. 애플리케이션 레벨 헬스체크(단순 포트 오픈이 아니라 실제 URL 응답)를 확인한 뒤 **Active**로 되돌립니다.

jk-status에서 바꾼 activation은 공유 메모리에만 저장되므로 **Apache 재시작 시 workers.properties의 값으로 초기화**됩니다. 영구 제외가 필요하면 파일의 `activation` 속성으로 박아야 합니다.

## 전제 조건 — sticky가 동작하려면

`sticky_session=1`은 JSESSIONID 끝의 jvmRoute 접미사를 보고 라우팅하므로, **각 톰캣/WAS의 jvmRoute가 워커 이름과 정확히 일치**해야 합니다(`apigw-prod01` 등). Spring Boot 내장 톰캣은 AJP 커넥터가 기본 비활성이고 jvmRoute를 설정 파일로 받지 않으므로, 둘 다 코드(커스터마이저)로 직접 구성돼 있어야 합니다. jvmRoute가 비어 있으면 sticky는 조용히 동작하지 않고 매 요청 재분배됩니다.

## 직접 확인해 보기

핵심 주장 두 가지를 환경에서 검증하는 절차입니다. 본 문서의 시간 수치(3초, 약 2분)는 커널 문서·기본값에서 끌어온 값이므로, 실측으로 닫는 것까지가 검증입니다.

1. **시나리오 A — 즉시 페일오버.** 한 터미널에서 `while true; do curl -s -o /dev/null -w "%{http_code} %{time_total}\n" http://HOST/app/; sleep 0.2; done`을 돌리고 prod01의 WAS 프로세스를 kill 합니다.
   - PASS: 모든 응답 200 유지, 최대 1건 정도만 수십 ms 지연. jk-status에서 prod01이 ERR로 전환.
   - FAIL: 5xx 발생(페일오버 미동작) 또는 수십 초 매달림(시나리오 B로 행동 — 환경 재확인 필요).
2. **시나리오 B — OS 타임아웃 매달림.** prod01 호스트에서 `iptables -A INPUT -p tcp --dport 12209 -j DROP`으로 블랙홀을 재현하고 같은 curl 루프를 관찰합니다.
   - 예상: 첫 요청 1건이 환경의 connect 타임아웃만큼 매달린 뒤 200, 이후 정상, 약 60초마다 매달리는 요청 1건씩 재발.
   - `socket_connect_timeout=3000` 적용 후 같은 실험에서 매달림이 3초 이내로 줄면 처방 검증 완료.
3. **상태 전환 관찰.** jk-status(`worker.jk-status.type=status`를 JkMount한 URL)에서 멤버 상태가 OK → ERR → REC → OK로 도는 것을 실험 중 함께 봅니다.
4. **시나리오 C — graceful 중 AJP 소켓 생존.** Spring Boot 백엔드에 graceful 종료를 걸어 둔 상태에서 해당 호스트의 `ss -ltn | grep 12209`를 봅니다. LISTEN이 남아 있으면 backlog 함정이 유효한 구성(추가 커넥터 기본값), 사라졌으면 fail-fast 구성(`bindOnInit=false` 적용됨)입니다.

## 출처

- [Tomcat Connectors — workers.properties Reference](https://tomcat.apache.org/connectors-doc/reference/workers.html) — 기본값 표 전체(`socket_connect_timeout=socket_timeout*1000`, `recover_time=60`, `retries=2`, `sticky_session_force=false`, `recovery_options=0`), force=false 시 "세션을 잃는 대신 페일오버" 문구
- [Tomcat Connectors — Timeouts How-To](https://tomcat.apache.org/connectors-doc/common_howto/timeouts.html) — "모든 타임아웃 기본 비활성", connect 실패 감지가 TCP 재전송 때문에 수 분 걸릴 수 있다는 경고, `socket_connect_timeout` 1~5초·CPing 권장값
- [Tomcat Connectors — Status Worker Reference](https://tomcat.apache.org/connectors-doc/reference/status.html) — 멤버 상태·activation 런타임 변경(비영구)
- [jk_ajp_common.c (tomcat-connectors 소스)](https://github.com/apache/tomcat-connectors/blob/main/native/common/jk_ajp_common.c) — 전송 전 소켓 생존 검사(`jk_is_socket_connected`), 첫 본문 패킷 재전송 버퍼(`op->post`/`reco_buf`), 복구 불가 전환 조건
- [Spring Boot — Graceful Shutdown](https://docs.spring.io/spring-boot/reference/web/graceful-shutdown.html) — "network layer에서 수신 중단" / drain 기본 30초는 [LifecycleProperties.java](https://github.com/spring-projects/spring-boot/blob/2.7.x/spring-boot-project/spring-boot-autoconfigure/src/main/java/org/springframework/boot/autoconfigure/context/LifecycleProperties.java)의 `Duration.ofSeconds(30)`
- [GracefulShutdown.java (Spring Boot 소스)](https://github.com/spring-projects/spring-boot/blob/2.7.x/spring-boot-project/spring-boot/src/main/java/org/springframework/boot/web/embedded/tomcat/GracefulShutdown.java) — 모든 커넥터에 `pause()` + `closeServerSocketGraceful()` 호출
- [TomcatServletWebServerFactory.java (Spring Boot 소스)](https://github.com/spring-projects/spring-boot/blob/2.7.x/spring-boot-project/spring-boot/src/main/java/org/springframework/boot/web/embedded/tomcat/TomcatServletWebServerFactory.java) — 주 커넥터에만 `bindOnInit=false`, 추가 커넥터는 커스터마이즈 미적용
- [AbstractEndpoint.java (Tomcat 9 소스)](https://github.com/apache/tomcat/blob/9.0.x/java/org/apache/tomcat/util/net/AbstractEndpoint.java) — `pause()`는 소켓을 닫지 않음, `closeServerSocketGraceful()`은 BOUND_ON_START만 닫음
- [WildFly Admin Guide — Suspend, Resume and Graceful Shutdown](https://docs.wildfly.org/26/Admin_Guide.html#graceful-shutdown) — suspend 중 신규 요청 503
- [Linux ip-sysctl — tcp_syn_retries](https://docs.kernel.org/networking/ip-sysctl.html) — 기본 6회. 현행 문서 기준 최종 타임아웃 약 131초(`tcp_syn_linear_timeouts` 반영), 구형 지수 백오프 계산으로는 약 127초
