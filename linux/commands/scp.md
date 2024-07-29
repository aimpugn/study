# scp

## [scp](https://linux.die.net/man/1/scp)?

> OpenSSH secure file copy

copies files between hosts on a network.

uses the `SFTP` protocol over a `ssh`(1) connection for data transfer, and uses the same authentication and provides the same
security as a login session.

## 다운로드

```bash
scp -v account@remote:/path/to/remote/file ./
```

## 업로드

```bash
# -v: verbose
# -r: recursive for directory
scp -v -r ./DIRECTORY account@remote:/path/to/remote

# 만약 ~/.ssh/config에 미리 remote 접속에 대한 설정이 되어 있다면, 그대로 사용 가능
scp -v -r ./DIRECTORY <my-remote>:/path/to/remote
```
