# [vaultwarden](https://github.com/dani-garcia/vaultwarden)

- [vaultwarden](#vaultwarden)
    - [vaultwarden](#vaultwarden-1)
    - [실행하기](#실행하기)
    - [KDF Algorithms](#kdf-algorithms)
    - [Autofill](#autofill)
        - [Using URIs](#using-uris)
    - [bitwarden browser extension에 추가하기](#bitwarden-browser-extension에-추가하기)
    - [기타](#기타)

## vaultwarden

Rust로 작성되고 [업스트림 Bitwarden 클라이언트*](https://bitwarden.com/download/)와 호환되는 Bitwarden 서버 API의 대체 구현으로, 리소스를 많이 사용하는 공식 서비스를 실행하는 것이 이상적이지 않을 수 있는 셀프 호스팅 배포에 적합합니다.

## 실행하기

기본적인 실행 명령어는 다음과 같습니다.

```sh
docker pull vaultwarden/server:latest
docker run -d --name vaultwarden -v /vw-data/:/data/ --restart unless-stopped -p 80:80 vaultwarden/server:latest
```

로컬에서 임의 포트로 실행하기 위해 아래와 같이 포트를 수정해서 실행합니다.

```sh
#!/bin/sh

VAULTWARDEN_NAME="vaultwarden"
VAULTWARDEN_VOLUME="vaultwarden-data"

# VAULTWARDEN_VOLUME 볼륨이 존재하지 않는다면 명명된 볼륨(named volume) 생성합니다.
# 1. 컨테이너가 삭제되어도 데이터는 유지됩니다.
# 2. 호스트 파일 시스템에 최적화된 방식으로 데이터를 저장합니다.
# 3. 볼륨은 다른 컨테이너에 쉽게 재사용될 수 있습니다.
# 4. 볼륨 데이터를 쉽게 백업하고 복원할 수 있습니다.
if ! docker volume ls -q | grep -q "^${VAULTWARDEN_VOLUME}$"; then
    if docker volume create "${VAULTWARDEN_VOLUME}"; then
        echo "볼륨 '${VAULTWARDEN_VOLUME}'이(가) 성공적으로 생성되었습니다."
    else
        echo "볼륨 '${VAULTWARDEN_VOLUME}' 생성 중 오류가 발생했습니다."
        exit 1
    fi
fi

docker stop "$VAULTWARDEN_NAME"
docker rm "$VAULTWARDEN_NAME"
docker run -d --name "$VAULTWARDEN_NAME" \
  -v "$VAULTWARDEN_VOLUME":/data \
  -p 11111:80 \
  -e 'WEBSOCKET_ENABLED=true' \
  -e 'LOG_FILE=/data/vaultwarden.log' \
  --restart unless-stopped \
  "$VAULTWARDEN_NAME"/server:latest
```

> [OrbStack 경우 NFS를 사용](https://orbstack.dev/blog/fast-filesystem) 합니다.
> 따라서 [볼륨은 해당 NFS 안에 존재](https://docs.orbstack.dev/docker/file-sharing)하게 됩니다.

## [KDF Algorithms](https://bitwarden.com/help/kdf-algorithms/)

> `KDF`? Key Derivation Functions (KDFs)

## Autofill

### [Using URIs](https://bitwarden.com/help/uri-match-detection/)

## [bitwarden browser extension에 추가하기](https://source.adra.cloud/add-vaultwarden-account-in-bitwarden-web-browser-extension)

## 기타

- [VaultWarden: Your local password manager](https://medium.com/@disane1987/vaultwarden-your-local-password-manager-781da310c064)
