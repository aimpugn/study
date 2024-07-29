# tsh ssh

## tsh scp 통해 파일 다운로드 받기

### Teleport를 사용하여 SCP로 파일 다운로드하기

1. **Teleport 인증**: 먼저, `tsh`를 사용하여 Teleport 클러스터에 로그인합니다. 이 과정에서 사용자 인증 정보를 얻게 됩니다.

    ```bash
    tsh login --proxy=your-proxy.example.com --user=your-username
    ```

    여기서 `your-proxy.example.com`은 Teleport 프록시 서버의 주소이고, `your-username`은 사용자 이름입니다.

2. **SCP를 사용하여 파일 다운로드**: `tsh scp` 명령어를 사용하여 원격 시스템에서 파일을 다운로드할 수 있습니다.

    ```bash
    tsh scp [원격_호스트]:[원격_파일_경로] [로컬_저장_경로]
    ```

    예를 들어, `remote-host`라는 호스트의 `/home/user/file.txt`를 현재 디렉토리로 다운로드하려면 다음과 같이 입력합니다.

    ```bash
    tsh scp remote-host:/home/user/file.txt ./
    ```

    이 명령은 `remote-host`에서 `/home/user/file.txt` 파일을 현재 작업 중인 로컬 디렉토리로 복사합니다.

### 주의사항

- 위 예시에서 `[원격_호스트]`, `[원격_파일_경로]`, `[로컬_저장_경로]`는 실제 사용 환경에 맞게 변경해야 합니다.
- Teleport 서버 설정, 네트워크 구성, 사용자 권한 등에 따라 접근 방법이나 명령어 사용법이 다를 수 있습니다.
- Teleport 문서 및 `tsh` 명령어의 도움말(`tsh --help`)을 참고하여 추가 옵션을 확인하세요.

Teleport를 통한 SCP 사용은 Teleport의 보안 및 인증 메커니즘을 활용하여 파일을 안전하게 전송할 수 있는 방법을 제공합니다. Teleport 설정이나 네트워크 정책에 따라 명령어의 정확한 사용법이나 필요한 권한이 달라질 수 있으므로, 관련 문서를 참고하거나 시스템 관리자에게 문의하는 것이 좋습니다.
