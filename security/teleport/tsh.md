# tsh

## tsh?

Teleport는 현대적인 SSH 서버로, 사용자가 SSH나 SCP를 통해 원격 시스템에 안전하게 접근할 수 있도록 설계된 도구입니다.
`tsh`는 Teleport의 클라이언트 도구로, 사용자 인증, 세션 로깅, 클러스터 관리 등을 담당합니다.

## 목록 출력

```bash
tsh ls -d
```

## remote 서버 접속

tsh를 사용하여 원격 서버에 SSH로 접속할 수 있습니다.

```bash
tsh ssh [user]@[remote-server]
```

```bash
tsh ssh ubuntu@prod-adm-server-cron
```
