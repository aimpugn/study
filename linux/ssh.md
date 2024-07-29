# ssh

## config

- 보통 `~/.ssh/config`에 위치한다

## config 구성 방법

```config
Host ${이 설정의 별명}
  HostName ${실제로 SSH 접속할 대상 서버의 IP 주소 또는 도메인}
  User ${SSH 접속 시 사용할 사용자 이름}
  IdentityFile ${접속에 사용할 키 파일의 경로}
  LocalForward ${로컬의 포트} ${대상 서버}:${대상 서버의 포트}
```

- `HostName`: SSH 대상이 되는 서버. 이 경우 포워딩 용으로 사용되므로 다음과 같이 불릴 수 있다
    - 바운스 서버(Bounce Server)
    - 중계 서버(Relay Server)
    - 게이트웨이 서버(Gateway Server)
    - 점프 서버(Jump Server)
- 이러한 구성은 특히 SSH 터널링, VPN, 또는 리버스 프록시 등에서 자주 볼 수 있다
- 여기서 중요한 것은, 첫 번째 "hop"을 거치는 "중계 서버"가 어떤 역할을 하는지, 그리고 최종 "대상 서버"가 어떤 서비스를 제공하는지이다

E177HZ9TH90FZI 생성
