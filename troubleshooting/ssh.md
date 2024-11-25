# ssh

- [ssh](#ssh)
    - [Control socket connect(/Users/rody/.ssh/sock/dev-masterDB.sock): Connection refused](#control-socket-connectusersrodysshsockdev-masterdbsock-connection-refused)
        - [문제](#문제)
        - [원인](#원인)
            - [`overlayfs`를 파일시스템으로 사용할 경우 유닉스 소켓과 잘 맞지 않아서 메모리를 대신 사용하도록 한 경우](#overlayfs를-파일시스템으로-사용할-경우-유닉스-소켓과-잘-맞지-않아서-메모리를-대신-사용하도록-한-경우)
            - [ssh: sharing control sockets over nfs](#ssh-sharing-control-sockets-over-nfs)
            - [Clearing ssh control sockets](#clearing-ssh-control-sockets)
        - [해결](#해결)
    - [no mutual signature algorithm](#no-mutual-signature-algorithm)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)

## Control socket connect(/Users/rody/.ssh/sock/dev-masterDB.sock): Connection refused

### 문제

ssh 연결 시 소켓 통해서 컨트롤할 수 있도록 했는데, exit이 먹히지 않음

```conf
Host *
  ServerAliveInterval 120
  TCPKeepAlive no
  PubkeyAcceptedKeyTypes +ssh-rsa

Host dev-workstation
  Hostname 3.35.156.78
  User ubuntu
  IdentityFile ~/.ssh/some-dev-pemfile.pem

Host dev-masterDB
  HostName 10.0.12.39
  User ubuntu
  IdentityFile ~/.ssh/some-dev-pemfile.pem
  ControlMaster auto
  ControlPath ~/.ssh/sock/dev-masterDB.sock
  ControlPersist yes
  ProxyJump dev-workstation
  LocalForward 3307 127.0.0.1:3306
  PubkeyAcceptedKeyTypes +ssh-rsa
```

### 원인

#### [`overlayfs`를 파일시스템으로 사용할 경우 유닉스 소켓과 잘 맞지 않아서 메모리를 대신 사용하도록 한 경우](https://stackoverflow.com/a/36479771)

- mac os에서 안 되는 거라서 상관없을 거 같다
- 다만 해결 시 다음과 같은 방법을 사용했다고 함

```shell
ControlPath /var/shm/control:%h:%p:%r
```

#### [ssh: sharing control sockets over nfs](https://superuser.com/questions/352263/ssh-sharing-control-sockets-over-nfs)

- 소켓 파일들은 디렉토리 엔트리 외에, 실제로 파일 시스템에 존재하지 않음. 같은 머신의 VFS 내에서 포인터로만 사용된다. 여러 머신에서 VFS를 공유할 수 없으므로, 소켓 파일도 공유할 수 없다.
- 하지만 여러 머신에서 사용하고 있는 것도 아니므로, 이 케이스도 아닌 거 같다

#### [Clearing ssh control sockets](https://shallowsky.com/blog/linux/ssh-tips.html)

ssh 연결이 열린 상태에서 네트워크 장애가 있었고, 복구 된 후 ssh 연결을 할 때마다 `Control socket connect(/home/username/ssh-username@example.com:port): Connection refused` 라는 메시지가 나오며, 어쨌든 연결은 된다고 함.

### 해결

- 해당 소켓을 사용하는 `ssh dev-masterDB` 연결을 다시 한 후에 종료 시키니 정상적으로 종료 됨

```shell
ssh -S ~/.ssh/sock/dev-masterDB.sock -O exit dev-masterDB
```

- `/home/username/ssh-username@example.com:port` 같은 문제가 되는 파일을 삭제

## no mutual signature algorithm

### 문제

```log
sign_and_send_pubkey: no mutual signature supported
ubuntu@10.0.12.39: Permission denied (publickey).
```

### 원인

[해당 SSH key 생성했던 알고리즘이 deprecated](https://confluence.atlassian.com/bitbucketserverkb/ssh-rsa-key-rejected-with-message-no-mutual-signature-algorithm-1026057701.html) 됐기 때문

### 해결

`PubkeyAcceptedKeyTypes` 통해서 알고리즘을 추가

```config
Host dev-masterDB
  HostName 10.0.12.39
  User ubuntu
  IdentityFile ~/.ssh/some-dev-pemfile.pem
  ControlPath ~/.ssh/sock/dev-masterDB.sock
  ProxyJump dev-workstation
  LocalForward 3307 127.0.0.1:3306
  PubkeyAcceptedKeyTypes +ssh-rsa <- 추가
```
