# lsof

- [lsof](#lsof)
    - [lsof?](#lsof-1)
    - [포트 열려 있는지 확인하기](#포트-열려-있는지-확인하기)
        - [포트 열려 있는지 확인하기 by `ss`](#포트-열려-있는지-확인하기-by-ss)

## [lsof](https://phoenixnap.com/kb/lsof-command)?

> List of open files

`lsof`는 실행 중인 프로세스가 열고 있는 파일 목록을 보여준다.
*네트워크 소켓도 파일*로 취급되기 때문에, 이를 통해 특정 포트를 사용하는 프로세스 정보를 얻을 수 있다.

## 포트 열려 있는지 확인하기

```bash
# 3307포트 열려있는지 확인
lsof -i:3307

# 443 포트를 리슨하고 있는 프로세스 찾기
sudo lsof -i :443
```

- `-i` 옵션은 인터넷 관련 파일(네트워크 소켓)만을 대상으로 확인
- `:443`는 443번 포트에 대한 정보만을 필터링해서 보여달라는 의미

결과에는 프로세스 이름, 프로세스 ID(PID), 사용자, 프로토콜 타입(TCP/UDP), 상태(LISTEN 등) 등의 정보가 포함될 수 있다.

```bash
$ sudo lsof -i :80
COMMAND   PID     USER   FD   TYPE  DEVICE SIZE/OFF NODE NAME
nginx   10639     root   10u  IPv4 1075233      0t0  TCP *:http (LISTEN)
nginx   10639     root   11u  IPv6 1075234      0t0  TCP *:http (LISTEN)
```

### 포트 열려 있는지 확인하기 by `ss`

또한, `netstat` 명령어 또는 `ss` 명령어를 사용하여 같은 정보를 얻을 수도 있다.

```sh
sudo ss -ltnp | grep ':443'
```

- `-l` 옵션은 리스닝 상태의 소켓만 보여준다.
- `-t` 옵션은 TCP 소켓만을 대상으로 한다.
- `-n` 옵션은 서비스 이름 대신 포트 번호를 표시한다.
- `-p` 옵션은 소켓을 소유하고 있는 프로세스의 정보를 표시한다 (root 권한 필요).
