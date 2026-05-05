# Database And Storage

> 원문 보존형 이동본입니다. 이 파일의 source chunk 본문은 원본 `intervie*.md`에서 그대로 복사되었고, 기술적 보강과 딥 리라이트는 다음 단계에서 수행합니다.

## Source Chunks

<!-- source-chunk: sha256=b771bbde329427fd1fef84e21d961a0a1b502d12b1c9ea3c4e5040bc04d6d575 topic=database-and-storage sources=interview_questions.md:1735-2084, interviews.md:1735-2084 -->

> Source: `interview_questions.md:1735-2084`
> Duplicate source aliases: `interview_questions.md:1735-2084, interviews.md:1735-2084`

## DB와 레플리케이션과 고가용성

데이터베이스 레플리케이션은 단순히 마스터의 모든 변경사항을 그대로 복제하는 것이 아니라, 다음과 같은 보호 메커니즘을 포함합니다:
1. 지연된 복제(Delayed Replication)
2. 백업 시스템
3. 변경 로그(Change Log) 관리
4. Point-in-Time Recovery (PITR)

구체적으로 데이터 손실 방지 메커니즘은 다음과 같습니다.

1. 지연된 복제

    ```mermaid
    sequenceDiagram
        participant Master
        participant Buffer as Replication Buffer
        participant Slave

        Master->>Buffer: 1. 변경사항 기록
        Note over Buffer: 2. 설정된 시간만큼 대기
        Buffer->>Slave: 3. 지연 후 복제
    ```

    작동 방식:
    - 슬레이브는 마스터의 변경사항을 즉시 적용하지 않고 일정 시간(예: 1시간) 지연
    - 실수로 데이터가 삭제되었을 때 지연 시간 내에 복구 가능

    ```sql
    -- MySQL에서 복제 지연 설정 예시
    CHANGE MASTER TO MASTER_DELAY = 3600; -- 1시간 지연
    ```

2. 바이너리 로그(Binary Log)

    ```mermaid
    graph LR
        A[트랜잭션 시작] --> B[바이너리 로그 기록]
        B --> C[데이터 변경]
        C --> D[트랜잭션 커밋]
        D --> E[레플리케이션 전송]
    ```

    - 목적: 모든 데이터 변경사항을 시간 순서대로 기록
    - 활용: 특정 시점으로 복구 가능
    - 보관: 일정 기간(예: 7일) 동안 보관

    ```sql
    -- MySQL 바이너리 로그 설정
    SET GLOBAL binlog_retention_period = 604800; -- 7일간 보관
    ```

    단, 디스크 용량 관리에 대한 고려가 필요합니다.
    바이너리 로그는 용량을 많이 차지할 수 있으므로, 백업과 삭제 주기를 정하고 디스크 용량이 부족해지지 않도록 모니터링하는 것이 중요합니다.

    ```sql
    -- 바이너리 로그 정리 (용량 관리)
    PURGE BINARY LOGS BEFORE 'YYYY-MM-DD HH:MM:SS';
    ```

3. Point-in-Time Recovery (PITR)

    전체 백업과 바이너리 로그를 조합하여 특정 시점까지 데이터를 복구하는 방식입니다.

    ```mermaid
    graph TD
        A[전체 백업] --> B[바이너리 로그]
        B --> C[특정 시점 선택]
        C --> D[복구 실행]
    ```

    작동 방식:
    1. 전체 백업본 사용
    2. 바이너리 로그로 특정 시점까지 복구
    3. 원하는 시점 이후의 변경사항은 무시

    ```sql
    -- MySQL에서 특정 시점으로 복구 예시
    RESTORE FROM BACKUP;
    APPLY BINARY LOGS UNTIL '2024-01-20 15:30:00';
    ```

    PITR을 실제로 사용할 때 데이터 무결성을 확보하는 것도 중요합니다.
    예를 들어, 복구 시 중단된 트랜잭션이나 불완전한 데이터 업데이트가 없도록 트랜잭션 일관성을 유지해야 합니다.

    또한, 복구하는 과정에서 데이터가 많을 경우 복구 시간(RTO)이 길어질 수 있으므로, 복구 속도를 고려하여 백업 주기를 설정하는 것이 좋습니다.

실제 상황에서의 데이터 손실 대응은 다음과 같이 이뤄질 수 있습니다:

1. 시나리오 1: 실수로 데이터 삭제

    지연된 복제를 통해 실수로 데이터가 삭제된 경우 복구할 수 있습니다.

    ```mermaid
    sequenceDiagram
        participant Admin
        participant Master
        participant Slave

        Admin->>Master: DELETE 실수 발생
        Note over Master,Slave: 복제 지연 시간 내
        Admin->>Master: 복제 중지 명령
        Admin->>Slave: 데이터 복구
        Admin->>Master: Slave에서 데이터 복사
    ```

    - 슬레이브를 마스터로 승격하는 방식 (Failover 방식):
        실수로 삭제된 데이터가 슬레이브에 남아 있다면, 슬레이브를 새로운 마스터로 승격하여 복구할 수 있습니다.

        1. 데이터 삭제 확인

            마스터에서 실수로 데이터를 삭제했음을 확인합니다.
            삭제된 시점과 데이터를 기록해 두어야 나중에 복구 시 필요한 정보를 확보할 수 있습니다.

        2. 복제 중지 명령 실행 (슬레이브에서 삭제 반영 방지)
            지연된 복제 상태에서 데이터가 실수로 삭제된 경우 복제를 중지하지 않으면, 지연된 슬레이브에도 삭제가 반영될 수 있습니다.

            ```sql
            -- 복제 중지 명령
            STOP SLAVE;
            ```

        3. 슬레이브 상태 확인

            슬레이브가 어느 시점까지 마스터의 변경 사항을 반영했는지 확인합니다.
            이를 통해 슬레이브에 복구할 수 있는 데이터가 있는지 확인해야 합니다.

            ```sql
            -- 슬레이브 상태 확인
            SHOW SLAVE STATUS;
            ```

            이를 통해 슬레이브가 마스터의 변경 사항을 얼마나 반영했는지 확인하고, 복제 지연이 올바르게 동작했는지를 점검해야 합니다.

        4. 슬레이브를 새로운 마스터로 승격: 슬레이브에서 데이터를 직접 복구하는 대신, 슬레이브를 새로운 마스터로 승격시킵니다.

            ```sql
            RESET MASTER;
            ```

        5. 애플리케이션 연결 전환:
            애플리케이션이 새로운 마스터로 정상적으로 작동하도록 설정을 업데이트합니다.
            로드밸런서를 사용하거나 직접 DB 연결 설정을 변경하여 새로운 마스터로 연결되도록 합니다.

        6. 기존 마스터를 슬레이브로 전환: 기존 마스터 DB가 복구된 후에는 기존 마스터를 새로운 마스터의 슬레이브로 전환할 수 있습니다.

            ```sql
            CHANGE MASTER TO MASTER_HOST='new_master_ip', MASTER_LOG_FILE='new_master_log_file', MASTER_LOG_POS=new_master_log_position;
            START SLAVE;
            ```

    - 복제 시스템을 사용한 복구:

        1. 데이터 삭제 확인

            마스터에서 실수로 데이터를 삭제했음을 확인합니다.
            삭제된 시점과 데이터를 기록해 두어야 나중에 복구 시 필요한 정보를 확보할 수 있습니다.

        2. 복제 중지 명령 실행 (슬레이브에서 삭제 반영 방지)
            지연된 복제 상태에서 데이터가 실수로 삭제된 경우 복제를 중지하지 않으면, 지연된 슬레이브에도 삭제가 반영될 수 있습니다.

            ```sql
            -- 복제 중지 명령
            STOP SLAVE;
            ```

        3. 슬레이브 상태 확인

            슬레이브가 어느 시점까지 마스터의 변경 사항을 반영했는지 확인합니다.
            이를 통해 슬레이브에 복구할 수 있는 데이터가 있는지 확인해야 합니다.

            ```sql
            -- 슬레이브 상태 확인
            SHOW SLAVE STATUS;
            ```

            이를 통해 슬레이브가 마스터의 변경 사항을 얼마나 반영했는지 확인하고, 복제 지연이 올바르게 동작했는지를 점검해야 합니다.

        4. 슬레이브의 데이터 덤프: 슬레이브에서 데이터를 mysqldump와 같은 도구를 사용하여 전체 데이터를 덤프합니다.

            ```sql
            mysqldump -u root -p --all-databases > slave_backup.sql
            ```

        5. 마스터에서 데이터 복구: 덤프된 데이터를 마스터에 복구합니다. 이 방식은 모든 데이터를 일관되게 복구할 수 있어 쿼리를 수동으로 추출하는 방식보다 안전합니다.

            ```sql
            mysql -u root -p < slave_backup.sql
            ```

        6. 복제 재설정: 복구가 완료되면, 복제를 다시 설정하고 슬레이브로부터 최신 상태의 데이터를 마스터로 복제할 수 있도록 합니다.

2. 시나리오 2: 마스터 DB 장애

    ```mermaid
    sequenceDiagram
        participant Master
        participant Slave
        participant Monitor

        Monitor->>Master: 장애 감지
        Monitor->>Slave: Promote to Master
        Note over Slave: 새로운 마스터로 승격
        Note over Master: 복구 후 Slave로 전환
    ```

    1. 슬레이브를 새로운 마스터로 승격 (Failover)

        슬레이브를 새로운 마스터로 승격시키기 위해, 필요한 설정을 변경합니다.
        이 과정에서 GTID(Globally Unique Transaction ID)가 설정되어 있으면, 데이터 불일치 문제를 최소화할 수 있습니다.

        ```sql
        -- 복제 중지 명령
        STOP SLAVE;
        -- 슬레이브의 새로운 로그 파일을 생성하고, 현재 슬레이브가 새로운 마스터로 동작하도록 설정합니다.
        RESET MASTER;
        ```

    2. 슬레이브의 데이터 일관성 확인

        슬레이브를 마스터로 승격한 후, 데이터가 일관성 있는지 확인합니다.
        이를 위해 SHOW SLAVE STATUS를 통해 확인하고, 장애 시점 전까지 모든 트랜잭션이 제대로 반영되었는지 확인합니다.

        ```sql
        -- 슬레이브 상태 및 데이터 일관성 확인
        SHOW SLAVE STATUS;
        ```

        ```log
        # 슬레이브가 마스터로부터 얼마나 뒤처져 있는지를 초 단위로 나타냅니다.
        # 0에 가까울수록 슬레이브가 마스터와 거의 일치함을 의미합니다.
        Seconds_Behind_Master: 0

        # 두 필드 모두 Yes여야 슬레이브가 정상적으로 복제 작업을 수행하고 있다는 것을 의미합니다.
        # 만약 하나라도 No인 경우, 복제가 중단되었거나 장애가 발생한 것입니다.
        Slave_IO_Running: Yes
        Slave_SQL_Running: Yes

        # 슬레이브의 복제가 정상적으로 진행되고 있다면, Master_Log_File와 Relay_Master_Log_File이 동일하거나 근접해야 합니다.
        # 두 파일 간의 차이가 크다면 슬레이브가 복제 작업을 늦게 처리하고 있다는 의미일 수 있습니다.
        # 마스터에서 현재 슬레이브가 읽고 있는 마스터의 바이너리 로그 파일입니다.
        Master_Log_File: mysql-bin.000123
        # 슬레이브에서 현재 처리 중인 마스터의 바이너리 로그 파일입니다.
        Relay_Master_Log_File: mysql-bin.000123

        # 아래 두 값이 가까울수록 슬레이브가 마스터의 바이너리 로그를 빠르게 처리하고 있음을 나타냅니다.
        # 두 값이 많이 차이가 난다면 슬레이브의 SQL 실행이 느리거나 병목이 발생했을 가능성이 있습니다.
        # 슬레이브가 마스터에서 바이너리 로그를 읽어온 로그 위치입니다.
        Read_Master_Log_Pos: 345678
        # 슬레이브가 받은 바이너리 로그 중에서 현재 SQL로 실행된 로그 위치입니다.
        Exec_Master_Log_Pos: 345678

        # Last_Errno가 0이고 Last_Error가 비어 있어야 슬레이브 복제가 정상적으로 진행되었음을 의미합니다.
        # 오류가 발생한 경우 해당 번호와 메시지를 참고하여 문제를 해결해야 합니다.
        Last_Errno: 0
        Last_Error:

        # 슬레이브가 현재 읽고 있는 릴레이 로그 파일 및 로그 위치를 나타냅니다.
        # 릴레이 로그는 마스터에서 받은 바이너리 로그를 슬레이브에서 재생하는 용도로 사용됩니다.
        Relay_Log_File: relay-bin.000456
        Relay_Log_Pos: 789012
        ```

    3. 애플리케이션 연결 전환

        슬레이브가 새로운 마스터로 승격되었으므로, 애플리케이션을 새로운 마스터로 연결합니다.
        애플리케이션의 DB 연결 설정을 변경하거나, 로드밸런서를 사용하는 경우 로드밸런서를 통해 새로운 마스터로 연결 경로를 변경합니다.

    4. 장애 복구 후 기존 마스터를 슬레이브로 전환: 마스터 장애 복구가 완료되면, 기존 마스터를 슬레이브로 재구성합니다.

        ```sql
        -- 기존 마스터를 슬레이브로 설정
        CHANGE MASTER TO MASTER_HOST='새로운 마스터 주소', MASTER_USER='replication_user', MASTER_PASSWORD='password', MASTER_LOG_FILE='마스터 로그 파일', MASTER_LOG_POS=마스터 로그 위치;
        START SLAVE;
        ```

    5. 모니터링 및 복제 상태 확인

        모든 설정이 완료된 후, 복제 상태와 데이터 일관성을 다시 한 번 확인하고 시스템을 정상적으로 모니터링합니다.

        ```sql
        -- 복제 상태 확인
        SHOW SLAVE STATUS;
        ```

고가용성 보장을 위한 모범 사례는 다음과 같습니다.

1. 다중 레플리케이션

    ```plaintext
    Master  -> Slave1 (즉시 복제)
            -> Slave2 (1시간 지연)
            -> Slave3 (24시간 지연)
    ```

2. 정기적인 백업
    - 전체 백업(Full Backup)
    - 증분 백업(Incremental Backup)
    - 차등 백업(Differential Backup)

3. 모니터링 및 자동화

    ```python
    def monitor_replication_lag():
        if get_replication_lag() > threshold:
            alert_admin()
            if is_critical():
                pause_replication()
    ```

4. 복구 계획 및 테스트
    - 정기적인 복구 테스트 수행
    - 복구 시간 목표(RTO) 및 복구 시점 목표(RPO) 설정

```sql
-- 레플리케이션 상태 확인
SHOW SLAVE STATUS;

-- 복제 중지
STOP SLAVE;

-- 특정 시점으로 복구
-- 1. 바이너리 로그 위치 확인
SHOW BINARY LOGS;

-- 2. 특정 시점까지 복구
CHANGE MASTER TO MASTER_LOG_FILE='mysql-bin.000123',
                 MASTER_LOG_POS=456789;

-- 3. 복제 재시작
START SLAVE;
```

권장 사항:

1. 레플리케이션 설정
    - 최소 하나의 지연된 복제본 유지
    - 바이너리 로그 충분한 기간 보관

2. 모니터링
    - 복제 지연 모니터링
    - 데이터 불일치 감지

3. 백업 전략
    - 정기적인 전체 백업
    - 실시간 바이너리 로그 백업

4. 문서화
    - 복구 절차 문서화
    - 긴급 연락망 유지

<!-- /source-chunk -->

<!-- source-chunk: sha256=fea2e1bafed0b5b9afea28c93665a0d2590debac93641cd0ce8ffa3c4c17a784 topic=database-and-storage sources=interview_questions.md:2924-3024, interviews.md:2924-3024 -->

> Source: `interview_questions.md:2924-3024`
> Duplicate source aliases: `interview_questions.md:2924-3024, interviews.md:2924-3024`

## B Tree & B+ Tree

B Tree와 B+ Tree는 모두 균형 잡힌 트리 구조로, 대용량 데이터의 효율적인 검색, 삽입, 삭제를 위해 설계된 자료구조입니다.
B Tree와 B+ Tree는 모두 다중 경로 검색 트리이지만, 데이터 저장 방식, 리프 노드의 구조, 검색 효율성 등에서 차이가 있습니다.

- B Tree는 *모든 노드에서 데이터를 찾을 수 있어 특정 검색에서 더 효율적*일 수 있습니다.
- B+ Tree는 *순차 접근과 범위 검색에서 뛰어난 성능*을 보이며, *더 낮은 트리 높이로 인해 대량의 데이터를 다룰 때 유리*할 수 있습니다.

실제 적용 시에는 사용 사례, 데이터의 특성, 접근 패턴 등을 고려하여 적절한 구조를 선택해야 합니다.
데이터베이스 시스템에서는 주로 B+ Tree가 선호되는 경향이 있지만, 특정 상황에서는 B Tree가 더 적합할 수 있습니다.

1. 데이터 저장 위치
    - B Tree
        - 모든 노드(내부 노드와 리프 노드)에 키와 데이터를 저장합니다.
        - 각 노드는 키와 해당 키에 연관된 데이터, 그리고 자식 노드에 대한 포인터를 포함합니다.

    - B+ Tree
        - 내부 노드에는 키만 저장하고, 실제 데이터는 리프 노드에만 저장합니다.
        - 내부 노드는 키와 자식 노드에 대한 포인터만을 포함합니다.

2. 리프 노드 구조
    - B Tree
        - 리프 노드와 내부 노드의 구조가 동일합니다.
        - 리프 노드들 사이에 특별한 연결이 없습니다.

    - B+ Tree
        - 리프 노드들이 연결 리스트로 연결되어 있습니다.
        - 이 연결을 통해 순차적 접근이 가능합니다.

3. 검색 효율성
    - B Tree
        - 검색 시 루트에서 리프까지 내려가지 않아도 중간에 데이터를 찾을 수 있습니다.
        - 최악의 경우 루트에서 리프까지 탐색해야 합니다.

    - B+ Tree
        - 항상 리프 노드까지 내려가야 데이터를 찾을 수 있습니다.
        - 그러나 내부 노드에 키만 저장하므로 더 많은 키를 저장할 수 있어, 트리의 높이가 낮아질 수 있습니다.

4. 범위 검색 (Range Query) 성능
    - B Tree
        - 범위 검색 시 여러 노드를 오가며 검색해야 할 수 있습니다.

    - B+ Tree
        - 리프 노드가 연결되어 있어 범위 검색이 매우 효율적입니다.
        - 시작 지점을 찾은 후 연결된 리프 노드를 따라 순차적으로 접근할 수 있습니다.

5. 트리 높이와 균형
    - B Tree
        - 데이터가 내부 노드에도 저장되므로 B+ Tree에 비해 트리의 높이가 높을 수 있습니다.

    - B+ Tree
        - 내부 노드에 키만 저장하므로 더 많은 키를 저장할 수 있어 트리의 높이가 낮아질 수 있습니다.
        - 이는 검색 시 더 적은 I/O 연산을 필요로 할 수 있음을 의미합니다.

6. 노드 크기와 분기 요소 (Fanout)
    - B Tree
        - 각 노드가 키와 데이터를 모두 저장하므로 노드의 크기가 더 클 수 있습니다.
        - 이로 인해 각 노드의 자식 수(분기 요소)가 B+ Tree에 비해 적을 수 있습니다.

    - B+ Tree
        - 내부 노드가 키만 저장하므로 더 많은 키를 포함할 수 있습니다.
        - 이는 더 높은 분기 요소를 가능하게 하여, 트리의 높이를 낮출 수 있습니다.

7. 구현 복잡성
    - B Tree
        - 모든 노드가 동일한 구조를 가지므로 구현이 상대적으로 간단할 수 있습니다.

    - B+ Tree
        - 내부 노드와 리프 노드의 구조가 다르고, 리프 노드 간의 연결을 관리해야 하므로 구현이 더 복잡할 수 있습니다.

다음은 B Tree와 B+ Tree의 노드 구조를 Java로 간단히 표현한 예시입니다:

```java
// B Tree 노드는 키, 데이터, 자식 노드 참조를 모두 포함합니다.
public class BTreeNode {
    private int[] keys;  // 키 배열
    private Object[] data;  // 데이터 배열
    private BTreeNode[] children;  // 자식 노드 배열

    // 생성자, getter, setter 등 메서드...
}

// B+ Tree는 내부 노드와 리프 노드가 다른 구조를 가집니다.
// B+ Tree 내부 노드는 키와 자식 노드 참조만을 포함합니다.
public class BPlusTreeInternalNode {
    private int[] keys;  // 키 배열
    private BPlusTreeNode[] children;  // 자식 노드 배열

    // 생성자, getter, setter 등 메서드...
}

// B+ Tree 리프 노드는 키, 데이터, 그리고 다음 리프 노드에 대한 참조를 포함합니다.
public class BPlusTreeLeafNode {
    private int[] keys;  // 키 배열
    private Object[] data;  // 데이터 배열
    private BPlusTreeLeafNode next;  // 다음 리프 노드에 대한 참조

    // 생성자, getter, setter 등 메서드...
}
```

<!-- /source-chunk -->

<!-- source-chunk: sha256=b5b2c64cf37ef19189b8b6b6b77f0114610335009fa46c9426704bdc02ab0fa9 topic=database-and-storage sources=interview_questions.md:7760-7805, interviews.md:7708-7753 -->

> Source: `interview_questions.md:7760-7805`
> Duplicate source aliases: `interview_questions.md:7760-7805, interviews.md:7708-7753`

## 데이터베이스 락(Lock)과 격리 수준(Isolation Level)

데이터베이스 락과 격리 수준은 *동시성 제어*와 *데이터 일관성을 유지*하는 데 중요한 역할을 합니다.
- 락(Lock): 데이터베이스 리소스에 대한 동시 접근을 제어하는 메커니즘
- 격리 수준(Isolation Level): 동시에 실행되는 트랜잭션들 사이의 상호작용 정도를 정의
- 트랜잭션(Transaction): 데이터베이스의 상태를 변화시키는 하나의 논리적 작업 단위
- 동시성 제어(Concurrency Control): 동시에 실행되는 트랜잭션들이 데이터베이스의 일관성을 해치지 않도록 보장하는 기법
- 데드락(Deadlock): 두 개 이상의 트랜잭션이 서로가 점유하고 있는 자원을 요구하며 무한정 기다리는 상황

낙관적 잠금과 비관적 잠금
- 낙관적 잠금:
    "낙관적" 잠금은 CAS(compare-and-swap)라는 Couchbase의 값에 달려 있습니다.
    모든 문서에는 불투명한 값인 CAS 값이 있습니다.
    해당 문서가 변경될 때마다 새로운 CAS 값을 얻습니다.
    문서를 업데이트하려고 할 때 작업의 일부로 CAS 값을 전달합니다.
    CAS 값이 일치하면 Couchbase는 작업을 허용합니다.
    일치하지 않으면 작업이 허용되지 않고 대신 오류를 반환합니다.

    예를 들어 두 개의 프로세스가 있다고 가정해 보겠습니다: A와 B.
    A와 B는 모두 문서를 가져오기 위해 Couchbase에 "get" 요청을 합니다.
    Couchbase는 CAS 값과 함께 문서를 반환합니다.
    그런 다음 A와 B는 모두 이전에 받은 CAS 값을 전달하면서 Couchbase에 "set" 요청을 보냅니다.
    처리는 경쟁을 통해 둘 중 하나에서 먼저 발생합니다.
    B의 차례가 되면 문서의 CAS 값이 이미 변경되었으므로 작업이 실패한다고 가정해 보겠습니다.

    이 프로세스가 끝나면 B는 실패하고 플레이어의 검은 레벨 2에 머물러 있습니다.
    B가 성공하도록 하려면 문서를 다시 가져와서 최신 CAS 값을 가져온 다음 다시 시도하는 것이 한 가지 해결책입니다.

    물론 이 솔루션도 다시 실패하고 또 다시 실패할 수 있습니다.
    하지만 이것이 바로 "낙관적"이라고 불리는 이유입니다.
    이 방법은 문서가 심한 경합 상태에 있지 않을 것이고 *결국에는 성공할 것이라고 가정*합니다.
    서버에 의한 실제 잠금은 필요하지 않으며 값만 확인하면 됩니다.

- 비관적 잠금

    비관적 잠금을 사용하여 실제로 잠금을 설정할 수 있습니다.
    이 기능은 여러 문서를 변경하기 위해 문서 그래프를 잠그고 싶을 때 유용할 수 있습니다.

    카우치베이스에는 "GetAndLock"이라는 원자 연산이 있습니다.
    이 연산은 문서와 CAS 값을 반환합니다.
    이 시점에서 문서는 "잠긴" 것으로 간주됩니다.
    다른 프로세스에서 더 이상 잠글 수 없으며 CAS 값만 문서의 잠금을 해제할 수 있습니다.

    또한 GetAndLock을 사용할 때는 시간 제한 기간을 설정해야 합니다.
    타임아웃 기간이 지나면 Couchbase는 자동으로 잠금을 해제합니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=2e97940a6ed53cea83d7bf96dfe959c9151ae38e7a968ed7758f10129c81015c topic=database-and-storage sources=interview_questions.md:7806-7811, interviews.md:7754-7759 -->

> Source: `interview_questions.md:7806-7811`
> Duplicate source aliases: `interview_questions.md:7806-7811, interviews.md:7754-7759`

## MVCC와 스냅숏 격리

데이터베이스는 객체마다 커밋된 버전 여러 개를 유지할 수 있어야 한다.
진행중인 여러 트랜잭션에서 서로 다른 시점의 데이터베이스 상태를 봐야 할 수도 있기 때문이다.
데이터베이스가 객체의 여러 버전을 함께 유지하므로 이 기법은 다중 버전 동시성 제어(multi-version concurrency control, MVCC)라고 한다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=680eb04c6ec09929ee4ccdef10556c4587f66e75907eb0e07782f8a41489cb81 topic=database-and-storage sources=interview_questions.md:7980-8056, interviews.md:7928-8004 -->

> Source: `interview_questions.md:7980-8056`
> Duplicate source aliases: `interview_questions.md:7980-8056, interviews.md:7928-8004`

## 2단계커밋과 2단계 잠금

2단계 커밋(Two-Phase Commit, 2PC)은 분산 시스템에서 트랜잭션의 원자성(Atomicity)을 보장하기 위한 프로토콜입니다.
주로 데이터베이스 시스템이나 분산 트랜잭션 관리에서 사용됩니다.
여러 시스템이나 노드가 하나의 트랜잭션에 참여할 때, 트랜잭션의 모든 참여자가 성공적으로 트랜잭션을 수행했는지 확인하고, 모든 참여자가 동의했을 때만 최종적으로 커밋하는 방식입니다.

1. 준비(Prepare) 단계
    - 코디네이터(또는 트랜잭션 관리자)는 각 참여 노드(서버나 데이터베이스)에 트랜잭션을 준비할 수 있는지 물어봅니다.
    - 각 노드는 트랜잭션을 수행할 준비가 되면, "준비 완료(Prepared)" 상태가 됩니다.
    - 모든 노드가 준비 상태라는 응답을 보내면, 코디네이터는 트랜잭션을 커밋할 준비가 된 것으로 간주합니다.

2. 커밋(Commit) 단계
    - 코디네이터는 모든 참여 노드가 준비되었다는 응답을 받으면, 각 노드에 트랜잭션을 커밋하라고 지시합니다.
    - 각 노드는 트랜잭션을 실제로 커밋하고, 성공 여부를 코디네이터에게 보고합니다.
    - 만약 어떤 노드라도 준비되지 않았다고 응답하거나, 중간에 실패하면 코디네이터는 트랜잭션을 롤백(Rollback) 하라고 지시합니다.

2단계 커밋의 특징:
- 원자성 보장: 모든 참여 노드가 트랜잭션에 성공하거나, 실패하면 트랜잭션 전체가 롤백됩니다. 중간에 일부 노드만 성공하고 다른 노드가 실패하는 경우는 없습니다.
- 동기화 비용: 2PC는 각 노드와 코디네이터 간의 통신을 여러 번 요구하기 때문에, 성능 상의 오버헤드가 발생할 수 있습니다.
- 문제점: 코디네이터가 장애가 나거나 노드 중 하나가 응답하지 않으면 트랜잭션이 오랫동안 대기 상태에 머물 수 있는 블록킹 문제가 발생할 수 있습니다.

예를 들어, 두 개의 데이터베이스 시스템이 동일한 트랜잭션에 참여하는 경우, 하나의 데이터베이스에서 작업이 완료되었지만 다른 데이터베이스에서 작업이 실패할 수 있습니다.
이때 2단계 커밋 프로토콜은 두 데이터베이스가 모두 성공했는지 확인한 후, 성공하면 트랜잭션을 커밋하고, 하나라도 실패하면 롤백하여 데이터의 일관성을 보장합니다.

2단계 잠금(Two-Phase Locking, 2PL)은 데이터베이스 트랜잭션에서 일관성을 보장하기 위해 사용되는 잠금(lock) 프로토콜입니다.
2PL은 직렬 가능성(Serializability), 즉 트랜잭션이 동시에 수행될 때 발생할 수 있는 경쟁 조건(Race Condition)이나 데이터 불일치 문제를 방지하는 데 사용됩니다.

1. 잠금 획득 단계 (Growing Phase)
    - 트랜잭션은 필요한 모든 자원에 대해 잠금을 요청합니다.
    - 트랜잭션이 필요한 자원에 대한 모든 잠금을 획득할 때까지 잠금만 요청할 수 있으며, 이 동안 잠금을 해제할 수 없습니다.

2. 잠금 해제 단계 (Shrinking Phase)
    - 트랜잭션이 더 이상 자원에 대한 잠금을 획득하지 않고, 잠금을 해제하는 단계입니다.
    - 일단 잠금을 해제하기 시작하면, 그 이후에는 더 이상 새로운 자원에 대한 잠금을 요청할 수 없습니다.

여러 트랜잭션들이 동시에 어떤 자원에 대해 잠금을 요청하는 경우, 잠금 관리자(lock manager)는 누가 먼저 요청했는지를 기준으로 우선순위를 결정합니다.
일반적으로 요청 시점에 따라 우선순위가 정해지며, 먼저 요청한 트랜잭션이 우선권을 갖습니다.

2단계 잠금의 특징:
- 직렬 가능성 보장: 2PL은 트랜잭션 간의 직렬 가능성을 보장합니다. 즉, 동시에 여러 트랜잭션이 실행되더라도 그 결과는 순차적으로 실행한 것과 동일하게 보장됩니다.
- 교착 상태 가능성: 2PL은 교착 상태(Deadlock)를 발생시킬 가능성이 있습니다. 트랜잭션이 자원을 획득하기 위해 서로 대기하는 상황이 발생할 수 있기 때문입니다.
- 엄격 2단계 잠금 (Strict 2PL): 모든 자원의 잠금을 트랜잭션이 완료될 때까지 유지하다가 트랜잭션이 끝나면 한 번에 모든 잠금을 해제하는 방식으로, 교착 상태를 방지할 수 있지만 더 많은 잠금을 오랫동안 유지하게 됩니다.

예를 들어, 트랜잭션 A가 먼저 자원 X에 대해 잠금을 요청하고, 자원 Y에 대한 잠금을 요청하는 동안, 트랜잭션 B가 자원 Y에 대해 잠금을 요청할 수 있습니다.
2단계 잠금 프로토콜에서는 트랜잭션이 필요한 자원에 대한 잠금을 요청할 때, 먼저 요청한 트랜잭션이 우선적으로 처리됩니다.
따라서 2단계 잠금 프로토콜을 통해 트랜잭션 A가 자원 Y에 대한 잠금을 성공적으로 획득하고 작업을 완료한 후 자원을 해제하면, 그제서야 트랜잭션 B는 자원 Y에 접근할 수 있습니다.
이를 통해 트랜잭션 간 충돌을 피할 수 있습니다.

만약 트랜잭션 A와 B가 서로 다른 자원에 대해 잠금을 요청하고, 동시에 다른 트랜잭션이 이미 소유하고 있는 자원에 대해 잠금을 요청하는 상황이 발생하면 *교착 상태(Deadlock)*가 발생할 수 있습니다.
이를 방지하기 위해 여러 가지 기법이 사용됩니다:
- 타임아웃(Timeouts):

    특정 시간 동안 잠금을 획득하지 못하면 트랜잭션이 실패하고 롤백되는 방식으로, 교착 상태를 피할 수 있습니다.

- 교착 상태 탐지(Deadlock Detection):

    시스템이 교착 상태를 감지하고, 특정 트랜잭션을 중단시켜 교착 상태를 해결합니다.

- 락 순서 규칙(Lock Ordering):

    자원에 대한 잠금을 요청할 때 항상 일정한 순서를 따르게 함으로써 교착 상태를 방지하는 방법입니다.

    예를 들어, 자원 X를 먼저 잠근 트랜잭션이 자원 Y에 대한 잠금을 요청할 때, 자원 Y를 먼저 잠근 트랜잭션이 자원 X에 대한 잠금을 요청하지 않도록 규칙을 설정하는 방식입니다.

2단계 커밋과 2단계 잠금의 차이점:
- 2단계 커밋:

    분산 시스템에서 트랜잭션의 원자성(Atomicity)을 보장하기 위한 프로토콜입니다.
    트랜잭션이 여러 노드에서 동시에 성공했는지 확인하고 최종적으로 커밋 또는 롤백을 결정합니다.

- 2단계 잠금:

    데이터베이스에서 트랜잭션 간의 일관성(Consistency)과 직렬 가능성을 보장하는 프로토콜입니다.
    자원에 대한 잠금을 두 단계로 나누어 트랜잭션이 완료될 때까지 필요한 자원을 독점적으로 사용할 수 있게 합니다.

이 두 가지 개념은 분산 트랜잭션이나 데이터베이스의 일관성을 보장하는 중요한 메커니즘이지만, 적용되는 대상과 해결하고자 하는 문제의 관점에서 다릅니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=544da3e54cefe9923c6c09d44d83b774d5874f064a659d2381363b9d6c11760b topic=database-and-storage sources=interview_questions2.md:1120-1408, interviews2.md:1120-1408 -->

> Source: `interview_questions2.md:1120-1408`
> Duplicate source aliases: `interview_questions2.md:1120-1408, interviews2.md:1120-1408`

## 수억건의 데이터가 존재하는 테이블 처리 방법

### 문제 상황

- MySQL 5.5버전으로 3억건의 데이터가 저장된 테이블이 존재합니다.
- 하나의 테이블이 너무 커져서 몇 백 기가에 달하는 데이터가 존재합니다.
- 결제 내역을 저장하는 테이블이어서 서비스 중간은 불가능한 상태입니다.
- 또한 DB 버전 업그레이드도 현재로서는 어려운 상황입니다.
- 별도의 reader 서버나 replica 서버 등도 있지만, 결국 하나의 큰 테이블을 어떻게 해야 하는 상황입니다.
- incremental integer id가 있어서 이 아이디가 조회에 유의미하게 사용되고 있습니다. 따라서 데이터가 쌓이는 순서도 중요합니다.
- 조회도 실시간으로 이뤄지고 있습니다. 따라서 파티셔닝을 해도 조회는 일관되게 제공해야 합니다.

이런 상황에서 현재 테이블을 어떻게 해야 할지 한번 논해보세요.

### 핵심 제약

1. MySQL 5.5의 한계:
    - 최신 MySQL 버전에서 제공하는 `InnoDB`의 native partitioning 및 온라인 작업 기능을 사용할 수 없습니다.
2. 실시간 데이터 일관성:
    - 조회와 삽입 작업 모두 서비스 중단 없이 이루어져야 하며, 모든 데이터가 일관된 인터페이스로 접근 가능해야 합니다.
3. 데이터 크기와 성능 문제:
    - 테이블 크기가 크기 때문에 스캔 속도, 인덱스 유지 비용, 쿼리 응답 시간이 점점 느려질 가능성이 높습니다.

### 테이블 파티셔닝 (수동 구현)

MySQL 5.5에서 제공되는 기본적인 파티셔닝 기능은 제한적입니다.
대신, 수동으로 테이블을 분리하고 애플리케이션 레벨에서 이를 관리하는 방식을 사용할 수 있습니다.

#### 수동 파티셔닝 설계

- 기존 테이블(`payments`)을 기간 기반 또는 범위 기반으로 나누어 별도의 테이블로 분리합니다.
- ID 범위 기준으로 파티션 분리:
    - ID는 증가형 정수이므로, 특정 범위를 기준으로 물리적 테이블을 분리할 수 있습니다.
    - 예를 들어:

        ```sql
        CREATE TABLE payments_1_to_100M LIKE payments;
        CREATE TABLE payments_100M_to_200M LIKE payments;
        CREATE TABLE payments_200M_to_300M LIKE payments;
        ```

    - 기존 데이터를 새로운 테이블로 옮깁니다:

        ```sql
        INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
        DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
        ```

- 데이터 삽입:
    - 데이터는 새로운 `payments` 테이블에 삽입하며, 일정 기준에 도달하면 다음 파티션 테이블로 데이터를 옮깁니다.

#### 조회 시 파티션 통합

조회 작업은 뷰(View)를 사용하여 통합 인터페이스를 제공합니다:

```sql
CREATE VIEW all_payments AS
SELECT * FROM payments_1_to_100M
UNION ALL
SELECT * FROM payments_100M_to_200M
UNION ALL
SELECT * FROM payments_200M_to_300M;
```

- 실시간 쿼리에서 `all_payments`를 사용하면 기존 단일 테이블과 동일한 인터페이스를 제공합니다.
- 쿼리는 MySQL의 최적화 엔진에 따라 각 파티션 테이블에 병렬로 접근합니다.

하지만 뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다.

뷰를 사용하는 경우, 뷰 자체에는 인덱스를 생성할 수 없기 때문에 기존 인덱스의 기능을 유지하기 위한 별도의 접근 방식이 필요합니다. 각 파티션(분리된 테이블)에 존재하는 인덱스를 활용하면서, 통합 조회 성능을 유지할 수 있는 몇 가지 전략을 아래와 같이 제안합니다.

1. 개별 파티션 테이블에 동일한 인덱스 유지

    뷰가 여러 파티션 테이블을 병합(`UNION ALL`)하므로, 각 파티션 테이블에 필요한 인덱스를 동일하게 설정해야 합니다.

    파티션 테이블에 인덱스 생성:
    - 기존 단일 테이블에 존재하던 주요 인덱스를 분리된 테이블에도 동일하게 생성합니다.
    - 예를 들어, 단일 테이블의 인덱스들을 각 파티션 테이블에도 동일하게 생성합니다.:

        ```sql
        -- 기존 단일 테이블의 인덱스
        ALTER TABLE payments ADD INDEX idx_created (created);
        ALTER TABLE payments ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments ADD INDEX idx_name_created (name, created);

        -- 파티션된 테이블의 인덱스
        ALTER TABLE payments_1_to_100M ADD INDEX idx_created (created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_created (created);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_uuid (uuid);
        ALTER TABLE payments_1_to_100M ADD INDEX idx_name_created (name, created);
        ALTER TABLE payments_100M_to_200M ADD INDEX idx_name_created (name, created);
        ```

        이렇게 하면 MySQL은 뷰를 통한 조회 시 각 파티션 테이블에 존재하는 인덱스를 활용하여 조건에 맞는 데이터를 효율적으로 검색할 수 있습니다.

2. 쿼리 최적화: 조건절로 파티션 테이블 선택

    MySQL의 쿼리 최적화 엔진은 `UNION ALL`이 포함된 뷰를 사용할 때, `WHERE` 조건을 기반으로 쿼리를 최적화하여 불필요한 테이블에 접근하지 않도록 설계되어 있습니다.

    - 특정 조건이 있는 쿼리

        뷰에서 특정 조건(예: `id`, `created`, `uuid`)이 있는 경우, MySQL은 해당 조건이 충족되는 테이블만 스캔합니다.
        예를 들어, 다음과 같은 쿼리를 실행한다고 가정합니다:

        ```sql
        SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

        MySQL은 `payments_1_to_100M`과 `payments_100M_to_200M`의 `created` 인덱스를 활용하여 쿼리를 최적화할 수 있습니다.

    - MySQL 실행 계획 확인

        뷰가 제대로 작동하는지 확인하려면 `EXPLAIN`을 사용하여 실행 계획을 점검합니다:

        ```sql
        EXPLAIN SELECT * FROM all_payments WHERE created >= '2023-01-01' AND created < '2023-02-01';
        ```

        결과에서 MySQL이 조건에 맞는 테이블만 스캔하고, 해당 테이블의 인덱스를 활용하는지 확인할 수 있습니다.

3. 필요 시 애플리케이션 레벨에서 파티션 관리

    뷰가 모든 테이블을 병합하는 방식은 간단하지만, 데이터 규모가 매우 클 경우 비효율적일 수 있습니다.
    애플리케이션 로직에서 조건에 따라 적절한 파티션 테이블에 직접 접근하도록 설계하면 성능을 극대화할 수 있습니다.

    - 애플리케이션 레벨에서 파티션 테이블을 직접 선택하면 불필요한 테이블 스캔을 완전히 제거할 수 있습니다.
    - 조건이 명확한 경우 뷰를 사용하지 않아도 되므로 성능이 크게 개선됩니다.

    예를 들어, ID나 날짜를 기준으로 각 파티션 테이블을 선택하는 애플리케이션 레벨 로직을 추가할 수 있습니다:

    ```python
    def get_partition_table(created_date):
        if created_date < '2023-01-01':
            return 'payments_1_to_100M'
        elif created_date < '2024-01-01':
            return 'payments_100M_to_200M'
        else:
            return 'payments_200M_to_300M'

    query = f"SELECT * FROM {get_partition_table(requested_date)} WHERE created = '{requested_date}'"
    ```

4. 인덱스 중복 제거 및 통합 테이블 조회 성능 최적화

    뷰와 파티션 테이블 접근을 병행하는 경우, 다음과 같은 최적화 전략을 추가로 사용할 수 있습니다.

    - 중복 인덱스 최소화
        - 뷰를 사용하면 각 테이블에 동일한 인덱스가 있어야 하므로, 필요한 최소 인덱스만 유지합니다.
        - 조회 패턴 분석을 통해 자주 사용되지 않는 인덱스를 제거합니다.

    - 데이터 분할 기준 검토
        - ID 기반 분할 외에, 데이터 접근 패턴을 분석하여 `created` 날짜 또는 `uuid`와 같은 다른 기준으로 분할을 고려할 수 있습니다.
        - 예: 날짜 기반 테이블 분리:

            ```sql
            CREATE TABLE payments_2023 LIKE payments;
            CREATE TABLE payments_2024 LIKE payments;
            ```

5. 대규모 뷰 최적화가 어려운 경우

    뷰를 통한 통합 조회가 특정 쿼리에서 비효율적이라면, 머티리얼라이즈드 뷰를 사용하거나 주기적으로 통합 테이블을 생성하는 방식을 도입할 수 있습니다.

    > 머티리얼라이즈드 뷰(Materialized View)는 데이터베이스에서 뷰(View)의 일종으로, 질의 결과를 디스크에 저장하여 실제 데이터처럼 사용할 수 있는 데이터베이스 객체입니다.
    >
    > ```sql
    > CREATE MATERIALIZED VIEW mv_sales_summary AS
    > SELECT region, SUM(sales) AS total_sales
    > FROM sales
    > GROUP BY region;
    >
    > -- 머티리얼라이즈드 뷰를 최신 상태로 유지하려면 새로 고침을 수행해야 합니다:
    > -- Oracle 예제: Complete Refresh
    > BEGIN
    > DBMS_MVIEW.REFRESH('MV_SALES_SUMMARY', 'COMPLETE');
    > END;
    >
    > -- PostgreSQL 예제: Refresh Materialized View
    > REFRESH MATERIALIZED VIEW mv_sales_summary;
    > ```
    >
    > 일반적인 뷰(View)는 질의 실행 시점에 데이터를 실시간으로 조회하는 반면, 머티리얼라이즈드 뷰는 질의 결과를 물리적으로 저장하여 나중에 필요할 때 빠르게 조회할 수 있습니다.

    - 통합 테이블 생성
        - 일정 주기로 각 파티션 테이블의 데이터를 하나의 통합 테이블로 병합합니다.
        - 통합 테이블은 읽기 전용으로 사용하며, 정기적으로 업데이트합니다.

            ```sql
            CREATE TABLE consolidated_payments LIKE payments;
            INSERT INTO consolidated_payments
            SELECT * FROM payments_1_to_100M
            UNION ALL
            SELECT * FROM payments_100M_to_200M
            UNION ALL
            SELECT * FROM payments_200M_to_300M;
            ```

    - 주기적인 업데이트
        - Cron 스케줄러나 배치 작업을 통해 통합 테이블을 갱신합니다.
        - 장점: 복잡한 조회 쿼리를 단순화할 수 있음.
        - 단점: 통합 데이터가 실시간으로 업데이트되지 않을 수 있음.

### 데이터 아카이빙

과거 데이터를 아카이브 테이블로 이동하여 현재 테이블 크기를 줄입니다.
이 방법은 오래된 결제 데이터를 조회할 필요가 적은 경우에 적합합니다.

#### 아카이브 테이블 설계

- 과거 데이터를 별도의 테이블로 이동:

    ```sql
    CREATE TABLE payments_archive LIKE payments;
    INSERT INTO payments_archive SELECT * FROM payments WHERE created_at < '2022-01-01';
    DELETE FROM payments WHERE created_at < '2022-01-01';
    ```

- 정기적인 스케줄링:
    - 데이터를 특정 기간 동안 유지하고, 그 이후에 아카이빙 작업을 수행합니다.
    - 스케줄링 도구(Cron + Shell Script 또는 MySQL Event)를 사용하여 배치 작업으로 구현.

#### 아카이브 데이터 통합 조회

- 아카이브 테이블을 포함하여 조회할 수 있도록 뷰를 제공합니다:

    ```sql
    CREATE VIEW all_payments AS
    SELECT * FROM payments
    UNION ALL
    SELECT * FROM payments_archive;
    ```

- 최적화: 과거 데이터에 접근이 적다면, 아카이브 테이블을 읽기 전용 서버(레플리카)로 옮겨 읽기 부하를 분산시킵니다.

### 인덱스 최적화 및 테이블 정리

대규모 테이블에서는 인덱스 최적화가 성능에 큰 영향을 미칩니다.

#### 3.1 중복 인덱스 제거

- `SHOW INDEX FROM payments`로 중복 인덱스를 확인합니다.
- 불필요한 인덱스를 삭제하여 인덱스 유지 비용을 줄입니다:

    ```sql
    DROP INDEX idx_column_name ON payments;
    ```

### 3.2 복합 인덱스 추가

- 조회 패턴을 분석하여 복합 인덱스를 추가합니다.
- 예: `id`와 `created_at`을 함께 사용하는 쿼리 최적화를 위해 복합 인덱스를 생성:

    ```sql
    ALTER TABLE payments ADD INDEX idx_id_created_at (id, created_at);
    ```

### 점진적 테이블 분리

대규모 테이블을 분리하기 위해 서비스 중단 없이 데이터 마이그레이션을 수행합니다.

#### 배치로 데이터 옮기기

- 데이터를 분리하는 동안 트랜잭션을 사용하여 무결성을 유지:

    ```sql
    INSERT INTO payments_1_to_100M SELECT * FROM payments WHERE id BETWEEN 1 AND 100000000;
    DELETE FROM payments WHERE id BETWEEN 1 AND 100000000;
    ```

- 한 번에 모든 데이터를 옮기지 말고, 작은 청크 단위로 배치 처리합니다:

    ```sql
    SET @start = 1, @batch_size = 100000;
    WHILE (@start <= 100000000) DO
        INSERT INTO payments_1_to_100M
        SELECT * FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        DELETE FROM payments
        WHERE id BETWEEN @start AND @start + @batch_size - 1;
        SET @start = @start + @batch_size;
    END WHILE;
    ```

#### 점진적 데이터 분리 전략

- 새로운 데이터를 추가할 때부터 파티션 테이블로 직접 저장.
- 예: ID 생성 규칙에 따라 `payments_200M_to_300M` 테이블로 삽입.

<!-- /source-chunk -->

<!-- source-chunk: sha256=ff783a17d8347f66fc6f37b3b83e6a083d5ab87fe00eb81c404c52d3fdb9d978 topic=database-and-storage sources=interview_questions2.md:1409-1498, interviews2.md:1409-1498 -->

> Source: `interview_questions2.md:1409-1498`
> Duplicate source aliases: `interview_questions2.md:1409-1498, interviews2.md:1409-1498`

## 3억건의 데이터의 B+Tree 깊이 계산

B+ Tree에서 데이터의 깊이를 계산하려면 트리의 구조와 노드에 저장할 수 있는 데이터의 양을 기반으로 깊이를 추정할 수 있습니다.
B+ Tree는 모든 리프 노드가 동일한 깊이에 있고, 내부 노드가 많은 키를 저장할 수 있으므로 깊이가 비교적 얕은 것이 특징입니다.

### B+ Tree의 주요 특징

- 차수(Order, `m`): 각 노드는 최대 `m-1`개의 키를 저장할 수 있고, 최대 `m`개의 자식 노드를 가질 수 있습니다.
- 데이터 저장 위치: 모든 데이터는 리프 노드에 저장되며, 내부 노드는 검색을 위한 키만 저장합니다.
- 균형 트리: 트리는 항상 균형 상태를 유지하며, 모든 리프 노드는 동일한 깊이를 가집니다.

### B+ Tree의 최대 용량

- 한 레벨의 노드 수가 증가할수록 저장 가능한 키의 수는 기하급수적으로 증가합니다.
- 루트 노드는 최대 `m-1`개의 키를 저장합니다.
- 각 레벨의 노드는 최대 `m^h`개의 키를 관리합니다(`h`는 깊이).

### 인덱싱 예시

MySQL에서 일반적으로 사용되는 InnoDB의 B+ Tree는 16KB 페이지 크기를 사용하며, 페이지마다 약 200~400개의 키를 저장할 수 있습니다(각 키의 크기에 따라 다름).

### 깊이 계산 방법

1. 노드당 키 수 추정: 데이터가 3억 건이고, 각 노드(페이지)가 최대 300개의 키를 저장할 수 있다고 가정합니다.

2. 트리 깊이 공식

    - 리프 노드의 개수를 `L`이라 하고, 데이터 개수를 `N`, 노드의 키 수를 `m`이라 할 때:
        $$
        L = \lceil \frac{N}{m} \rceil
        $$
        $$
        h = \lceil \log_{m} L \rceil
        $$
    여기서:
    - $N = 300,000,000$ (데이터 수).
    - $m = 300$ (노드당 최대 키 수).

3. 구체적인 계산

    - 리프 노드의 수 $L$:
        $$
        L = \lceil \frac{300,000,000}{300} \rceil = 1,000,000 \, \text{(리프 노드 개수)}.
        $$

    - 트리 깊이 계산

        내부 노드의 자식 수(차수)를 $m = 300$으로 가정했을 때:
        $$
        h = \lceil \log_{300} 1,000,000 \rceil
        $$

        이를 계산하면:
        1. $\log_{300} 1,000,000 = \frac{\log_{10} 1,000,000}{\log_{10} 300}$
        2. $\log_{10} 1,000,000 = 6$ (10의 6제곱).
        3. $\log_{10} 300 \approx 2.477$.
        4. $\frac{6}{2.477} \approx 2.42$.

        즉, $h \approx 3$.
        트리의 깊이는 약 3~4 레벨로 유지됩니다.

        이러한 얕은 깊이 덕분에 데이터 접근 시 최대 3~4번의 I/O만으로도 원하는 데이터를 찾을 수 있습니다.

### B+ Tree 깊이와 데이터 접근 시간

- 접근 시간
    - B+ Tree는 깊이가 얕기 때문에 검색 성능이 매우 우수합니다.
    - 3억 건의 데이터에서도 3~4번의 디스크 페이지 읽기만으로 원하는 데이터를 찾을 수 있습니다.

- 다양한 차수와 키 수 변화

    차수가 클수록(페이지당 저장 가능한 키 수가 많을수록), 트리의 깊이는 더 얕아집니다:
    - $m = 400$일 경우:
        - 리프 노드 수 $L = 750,000$.
        - 깊이 $h = \lceil \log_{400} 750,000 \rceil \approx 3$.

### 추가 고려 사항

- InnoDB의 실제 구현
    - InnoDB는 리프 노드에 데이터 레코드를 직접 저장하고, 내부 노드는 검색 키만 저장합니다.
    - 내부 노드는 더 많은 키를 저장할 수 있으므로, 실제 트리의 깊이는 이론적으로 계산된 값보다 조금 더 얕을 수 있습니다.

- 다양한 데이터 분포
    - 데이터가 균등 분포되지 않고 특정 범위에 집중되어 있으면, 일부 리프 노드가 비대해질 수 있습니다.
    - 그러나 B+ Tree는 자동으로 균형을 유지하므로 깊이는 크게 변하지 않습니다.

- 데이터 추가 시 깊이 변화
    - 데이터가 지속적으로 추가되면 리프 노드와 내부 노드가 분할되며 트리의 깊이가 증가할 수 있습니다.
    - 그러나 실질적으로 데이터 수가 수십 배로 증가하지 않는 이상 깊이는 일정하게 유지됩니다.

<!-- /source-chunk -->

<!-- source-chunk: sha256=6460b6776fe2899a33d8a67993374a814f47c68e83af6056f64776f8c47dacdd topic=database-and-storage sources=interview_questions3.md:165-188 -->

> Source: `interview_questions3.md:165-188`

## 3. MySQL 관련 질문

### 질문 7. MySQL에서 인덱스는 어떤 구조이며, 쿼리 성능을 어떻게 높일 수 있나요?

#### **답변**

1. MySQL(InnoDB 엔진 기준)에서 대부분의 인덱스는 **B+Tree** 구조로 되어 있습니다. B+Tree는 트리 높이를 낮게 유지하도록 설계되어, 많은 데이터를 저장하면서도 빠른 검색을 가능케 합니다.
2. InnoDB에서 **PRIMARY KEY**(기본키) 인덱스는 “클러스터링 인덱스”로서, 실제 테이블 레코드 자체가 B+Tree leaf에 저장됩니다. 반면, 보조 인덱스(secondary index)는 leaf 노드에 PK 값을 별도로 저장하고, 그 PK를 통해 테이블 데이터를 찾아가는 방식입니다.
3. 쿼리가 느릴 때는 **쿼리 실행 계획**(`EXPLAIN`)을 통해 인덱스 사용 여부, 풀 스캔 여부, 조인 방식 등을 점검해야 합니다. 필요한 열(WHERE절, JOIN에 사용되는 열)에 적절한 인덱스를 추가하면 성능이 개선될 수 있습니다.
4. 하지만 지나친 인덱스 생성은 INSERT/UPDATE 시 성능을 떨어뜨릴 수 있으므로, 빈도와 활용도를 고려해 인덱스를 설계해야 합니다.

---

### 질문 8. MySQL 트랜잭션 격리 수준과 InnoDB의 MVCC 방식은 어떻게 동작하나요?

#### **답변**

1. 트랜잭션 격리 수준은 한 트랜잭션이 다른 트랜잭션의 중간 변경 내용을 얼마나 볼 수 있는지를 정의합니다. MySQL의 기본값은 **REPEATABLE READ**이며, 다른 선택지로 READ COMMITTED, SERIALIZABLE 등이 있습니다.
2. REPEATABLE READ는 같은 트랜잭션 내에서 같은 SELECT 쿼리를 여러 번 실행해도 결과가 동일하게 보이도록 보장합니다.
3. InnoDB는 **MVCC(Multi-Version Concurrency Control)**를 통해 이 격리 수준을 구현합니다. 구체적으로, 각 행에는 트랜잭션 ID 정보가 기록되며, Undo Log에 이전 버전이 보관됩니다. SELECT 시점의 트랜잭션 ID보다 작은 버전 중에서 최신 것을 읽어 **일관된 스냅샷**을 제공합니다.
4. 이렇게 MVCC를 사용하면 대부분의 읽기 작업이 락 없이도 동작 가능해지고, 동시에 쓰기가 일어날 때도 Undo Log를 참고해 충돌을 최소화합니다.

---

<!-- /source-chunk -->
