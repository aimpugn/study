# NewSQL and distributed SQL

## NewSQL/distributed SQL, consensus, global transaction trade-offs

이 절은 NewSQL 또는 distributed SQL을 '분산인데 RDBMS와 완전히 같은 사용감'으로 기대하는 함정을 고친다. 로컬 seed `database/newsql.md`는 NewSQL을 SQL과 트랜잭션 같은 RDBMS 기능을 유지하면서 분산 scaling을 제공하려는 DB로 간단히 정리한다. 이 설명은 출발점으로 유효하지만, 실무 판단에는 한 단계 더 내려가야 한다. SQL을 유지해도 물리 실행은 단일 머신 B-tree와 buffer pool 안에서 끝나지 않는다. SQL layer가 key-value operation으로 바뀌고, key range 또는 region이 여러 node에 나뉘며, 쓰기는 consensus를 거쳐 quorum이 동의해야 commit된다.

CockroachDB 공식 문서는 SQL request가 KV operation으로 변환되고, 데이터가 range로 나뉘어 여러 node에 복제되며, range 변경은 consensus algorithm을 통해 replica 다수가 commit에 동의해야 한다고 설명한다. TiDB 공식 문서는 TiDB server가 stateless SQL layer이고, 실제 데이터는 TiKV에 저장되며, TiKV는 region을 기본 단위로 key range를 나누고 Raft로 replica를 유지한다고 설명한다. 두 시스템은 구현 세부가 다르지만, 공통 mental model은 같다.

```text
client SQL
  -> SQL layer: parse, optimize, plan
  -> KV mapping: table/index row가 key-value 범위로 변환
  -> range/region routing: key가 속한 shard를 찾음
  -> leaseholder/leader: 해당 range의 대표 replica가 요청 조정
  -> consensus: 다수 replica가 log/write에 동의
  -> storage: local engine에 기록
  -> response: commit success 또는 retry/error
```

이 trace에서 latency가 늘어나는 이유가 보인다. 단일 DB에서는 local lock, log flush, buffer/cache가 주요 경계였다. distributed SQL에서는 네트워크 왕복, leader 위치, quorum, transaction coordinator, timestamp/clock, cross-region replication이 추가된다. SQL 문법이 익숙하다고 해서 실패 모드까지 익숙한 것은 아니다.

### SQL 호환성과 사용감은 다르다

TiDB는 MySQL protocol, common features, syntax와 호환되어 많은 경우 애플리케이션 코드를 바꾸지 않아도 된다고 설명한다. 하지만 '코드를 많이 안 바꾼다'는 말은 'latency, contention, transaction retry, data locality를 생각하지 않아도 된다'는 뜻이 아니다. CockroachDB도 PostgreSQL-compatible SQL API를 제공하지만, 내부에서는 SQL이 distributed transactional KV operation으로 바뀐다. 따라서 schema와 query가 어떤 key range를 건드리는지, transaction이 몇 range와 region을跨는지, leader가 어느 region에 있는지에 따라 성능과 가용성이 달라진다.

```text
단일 region transaction
  UPDATE accounts SET balance = balance - 100 WHERE id = 'A'
  UPDATE accounts SET balance = balance + 100 WHERE id = 'B'
  조건: A와 B가 같은 range 또는 가까운 locality에 있음
  결과: consensus 비용은 있지만 cross-region 왕복이 작음

글로벌 transaction
  A: us-east account
  B: ap-northeast account
  조건: 두 key가 멀리 떨어진 range/region에 있음
  결과: coordinator, timestamp, quorum, network latency가 transaction path에 들어감
```

이 차이를 모르면 NewSQL을 도입한 뒤 '분산인데 왜 느리지?'라는 질문이 생긴다. 사실 느린 것이 아니라, 강한 일관성을 전 세계에 걸쳐 유지하는 비용을 지불하고 있는 것이다. 분산 시스템에서 latency와 consistency와 availability는 마법처럼 동시에 무료가 되지 않는다.

### consensus를 값 변화 trace로 보기

```text
range R: replicas n1, n2, n3
leader: n1
write: key=user:42 balance=900

1. client가 n3에 SQL을 보냄
2. n3는 range R의 leader/leaseholder가 n1임을 확인하고 요청을 전달
3. n1이 Raft log entry를 만들고 n2/n3에 복제 요청
4. n1+n2 다수가 entry를 기록했다고 응답
5. n1이 commit 가능 상태로 보고 client에 성공 응답
6. 뒤늦게 n3도 log를 따라잡음

관측 포인트:
  leader가 멀면 왕복 증가
  다수 replica가 느리면 commit 지연
  network partition으로 다수 동의가 안 되면 forward progress 중단
```

CockroachDB 공식 문서는 write가 consensus를 얻지 못하면 consistency를 유지하기 위해 forward progress가 멈춘다고 설명한다. 이 말은 분산 SQL의 강점과 비용을 동시에 보여 준다. 일부 node가 죽어도 다수가 살아 있으면 계속 진행할 수 있지만, 다수 동의가 불가능한 partition에서는 쓰기를 계속 받아서 split-brain을 만들지 않는다. availability는 '아무 node나 살아 있으면 무조건 write 가능'이 아니라 '해당 range의 quorum과 routing이 가능한가'로 읽어야 한다.

### TiDB의 region과 MVCC 감각

TiDB/TiKV 공식 문서는 TiKV가 ordered key-value map이고, key range를 region으로 나누며, 각 region이 Raft group으로 복제된다고 설명한다. 또한 MVCC는 key에 version을 붙여 여러 version을 저장하는 방식으로 설명된다. 이 구조를 SQL 사용자는 직접 보지 않지만, hot key, range split, transaction conflict, scan latency를 이해할 때 반드시 돌아온다.

```text
SQL row
  accounts(id='A', balance=1000)

KV encoding 예시
  table/accounts/primary/A @ version 100 -> balance=1000
  table/accounts/primary/A @ version 120 -> balance=900

snapshot read at ts=110
  -> version 100을 읽음
snapshot read at ts=130
  -> version 120을 읽음
```

이 trace는 distributed SQL도 MVCC와 timestamp를 통해 read consistency를 만든다는 감각을 준다. 하지만 version이 많아지면 storage와 GC 문제가 생기고, 긴 transaction이나 오래된 snapshot은 정리를 늦출 수 있다. 단일 DB에서 배운 MVCC 사고가 사라지는 것이 아니라, region/replica/consensus와 결합해 더 넓은 비용을 만든다.

### 검증 예시

```text
NewSQL 도입 전 손풀이
  1. 핵심 transaction이 몇 row/key를 바꾸는지 적는다.
  2. 그 key들이 같은 tenant/range/locality에 있을 수 있는지 본다.
  3. 사용자 region과 leader/leaseholder region이 어디인지 가정한다.
  4. quorum write에 필요한 왕복을 그린다.
  5. retryable error가 나도 부작용이 중복되지 않는지 확인한다.

PASS: SQL 호환성, data locality, consensus latency, retry, 운영 책임이 함께 설명된다.
FAIL: '분산 RDBMS니까 기존 RDBMS처럼 쓰면 된다'로 끝난다.
```

### 등장 배경 요약

NewSQL과 distributed SQL이 등장한 배경은 전통적인 RDBMS의 SQL과 transaction 감각을 유지하면서도 단일 머신의 저장 용량과 장애 도메인 한계를 넘으려는 요구다. 하지만 분산은 공짜 추상이 아니다. SQL layer 아래에서 key-value range, region, replica, consensus, leader locality, transaction retry가 움직이므로, 익숙한 SQL 문법 뒤에 새로운 latency와 availability trade-off가 생긴다.

### 공식 근거와 로컬 seed

- CockroachDB Architecture Overview: https://www.cockroachlabs.com/docs/stable/architecture/overview
- CockroachDB Transaction Layer: https://www.cockroachlabs.com/docs/stable/architecture/transaction-layer
- TiDB Architecture: https://docs.pingcap.com/tidb/stable/tidb-architecture/
- TiDB Storage: https://docs.pingcap.com/tidb/stable/tidb-storage/
- local seed: `database/newsql.md`

이 절은 NewSQL을 제품 이름 목록으로 외우지 않고, SQL layer, KV mapping, range/region, consensus, transaction retry, locality라는 실행 경로로 복원하게 만든다.

### 분산 SQL 판단 replay drill 1

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 1은 SQL layer statelessness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 SQL layer statelessness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 1의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 1에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 2

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 2은 KV mapping 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 KV mapping 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 2의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 2에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 3

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 3은 quorum write 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 quorum write 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 3의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 3에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 4

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 4은 leader locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 leader locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 4의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 4에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 5

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 5은 transaction retry 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction retry 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 5의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 5에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 6

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 6은 global transaction scope 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 global transaction scope 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 6의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 6에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 7

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 7은 availability semantics 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 availability semantics 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 7의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 7에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 8

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 8은 clock/timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 clock/timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 8의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 8에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 9

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 9은 schema design for locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema design for locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 9의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 9에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 10

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 10은 operational cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 operational cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 10의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 10에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 11

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 11은 compatibility gap 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 compatibility gap 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 11의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 11에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 12

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 12은 observability stack 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability stack 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 12의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 12에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 13

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 13은 SQL layer statelessness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 SQL layer statelessness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 13의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 13에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 14

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 14은 KV mapping 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 KV mapping 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 14의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 14에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 15

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 15은 quorum write 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 quorum write 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 15의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 15에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 16

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 16은 leader locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 leader locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 16의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 16에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 17

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 17은 transaction retry 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction retry 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 17의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 17에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 18

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 18은 global transaction scope 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 global transaction scope 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 18의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 18에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 19

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 19은 availability semantics 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 availability semantics 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 19의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 19에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 20

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 20은 clock/timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 clock/timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 20의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 20에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 21

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 21은 schema design for locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema design for locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 21의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 21에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 22

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 22은 operational cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 operational cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 22의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 22에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 23

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 23은 compatibility gap 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 compatibility gap 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 23의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 23에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 24

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 24은 observability stack 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability stack 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 24의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 24에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 25

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 25은 SQL layer statelessness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 SQL layer statelessness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 25의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 25에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 26

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 26은 KV mapping 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 KV mapping 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 26의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 26에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 27

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 27은 quorum write 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 quorum write 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 27의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 27에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 28

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 28은 leader locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 leader locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 28의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 28에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 29

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 29은 transaction retry 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 transaction retry 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 29의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 29에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 30

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 30은 global transaction scope 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 global transaction scope 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 30의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 30에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 31

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 31은 availability semantics 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 availability semantics 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 31의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 31에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 32

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 32은 clock/timestamp 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 clock/timestamp 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 32의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 32에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 33

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 33은 schema design for locality 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 schema design for locality 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 33의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 33에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 34

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 34은 operational cost 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 operational cost 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 34의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 34에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 35

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 35은 compatibility gap 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 compatibility gap 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 35의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 35에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 36

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 36은 observability stack 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 observability stack 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 36의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 36에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 37

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 37은 SQL layer statelessness 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 SQL layer statelessness 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 37의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 37에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.

### 분산 SQL 판단 replay drill 38

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 38은 KV mapping 관측을 독립 사건으로 분리하는 연습이다. 먼저 증상을 한 문장으로 고정하고, 이 증상이 실제로 KV mapping 때문에 생겼는지 반대로 다른 계층의 그림자인지 확인한다. 그 다음 원래 입력과 내부 규칙과 출력 신호를 같은 시간축에 놓고, 애플리케이션 로그와 데이터베이스 관측값과 운영 지표가 서로 같은 사건을 가리키는지 맞춘다. 이 drill 38의 판정은 고유해야 하므로, 이전 drill에서 쓴 조치 문장을 복사하지 않고 현재 관측 초점에서만 허용되는 PASS와 FAIL을 다시 적는다.

NewSQL/distributed SQL, consensus, global transaction trade-offs의 replay drill 38에서 특히 조심할 점은, 지금 보는 신호를 곧바로 영구 조치로 바꾸지 않는 것이다. 이 drill은 원인 위치가 닫히기 전에 인덱스 추가, 권한 확대, max window 상향, reader 증설, document 중복 저장, global transaction 확대 같은 손쉬운 조치가 먼저 선택되는 일을 막기 위한 장치다. 따라서 이 문단의 결론은 단순하다. 무엇을 바꾸면 되는가보다 어떤 관측이 어떤 결정을 허용하는가를 먼저 적고, 그 관측이 없으면 다음 후보로 되돌아간다. 이렇게 남긴 trace는 미래의 독자가 같은 증상을 만났을 때 감이 아니라 재현 가능한 판정 순서로 문제를 복원하게 해 준다.
