# linux

## debug3: obfuscate_keystroke_timing: stopping: chaff time expired (134 chaff packets sent)

### 문제

EC2 인스턴스에 ssh로 접근 후 아래와 같은 메시지가 터미널에 계속 출력되는 현상이 발생했습니다.

```log
debug3: obfuscate_keystroke_timing: stopping: chaff time expired (134 chaff packets sent)
```

### 원인

이 메시지는 OpenSSH 클라이언트의 디버그 출력으로, 키스트로크 타이밍 난독화(obfuscation) 기능과 관련이 있습니다.

```log
debug3: obfuscate_keystroke_timing: stopping: chaff time expired (134 chaff packets sent)
```

- `debug3`: 가장 상세한 디버그 레벨을 나타냅니다.
- `obfuscate_keystroke_timing`: 키스트로크 타이밍 난독화 기능을 가리킵니다.
- `stopping`: 기능이 중지되고 있음을 나타냅니다.
- `chaff time expired`: 난독화 시간이 만료되었음을 의미합니다.
- `(134 chaff packets sent)`: 134개의 더미 패킷이 전송되었음을 나타냅니다.

`debug3`는 매우 상세한 로깅 레벨을 나타내며, 일반적으로 문제 해결 목적으로 사용됩니다.
이는 현재 SSH 클라이언트가 디버그 모드로 실행되고 있을 가능성이 높습니다.

### 해결

이 메시지를 보지 않으려면:

1. SSH 설정 확인:

    ```bash
    cat ~/.ssh/config
    ```

    `LogLevel` 설정이 있다면 `QUIET` 또는 `ERROR`로 변경하세요.

2. SSH 명령어 옵션 수정:

    ```bash
    ssh -o "LogLevel=QUIET" user@hostname
    ```

3. 시스템 전역 설정 확인 (관리자 권한 필요):

    ```bash
    sudo nano /etc/ssh/ssh_config
    ```

    `LogLevel` 설정을 찾아 수정하세요.
