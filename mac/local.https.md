# local https

- [local https](#local-https)
    - [설치](#설치)
    - [적용](#적용)

## 설치

```shell
brew install mkcert

# firefox를 사용하는 경우
brew install nss
```

## 적용

```shell
❯ mkcert -install
# Sudo password:
# The local CA is now installed in the system trust store! ⚡️
# The local CA is now installed in the Firefox trust store (requires browser restart)! 🦊
```

```shell
mkdir ./traefik/certs

mkcert -cert-file ./traefik/certs/some_name.test.crt \
  -key-file ./traefik/certs/some_name.test.key \
  sub_a.localhost \
  sub_b.localhost \
  sub_c.localhost

# Created a new certificate valid for the following names 📜
#  - "sub_a.localhost"
#  - "sub_b.localhost"
#  - "sub_c.localhost"
#
# The certificate is at "./traefik/certs/some_name.test.crt" and the key at "./traefik/certs/some_name.test.key" ✅
#
# It will expire on 24 November 2025 🗓
```
