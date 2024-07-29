# GHCR

- [GHCR](#ghcr)
    - [GHCR?](#ghcr-1)
    - [로그인](#로그인)
    - [build \& push](#build--push)

## GHCR?

- GitHub Container Registry

## 로그인

```bash
$ docker login ghcr.io

Username (Rody): aimpugn
Password:
```

- Username: github user name
- Password: Personal Access Token

## build & push

```bash
/usr/bin/docker buildx build \
    --build-arg XDEBUG_ENABLE=true \
    --build-arg XDEBUG_ENALBE=true \
    --iidfile /tmp/docker-build-push-SUdVri/iidfile \
    --platform linux/amd64,linux/arm64 \
    --secret id=GIT_AUTH_TOKEN,src=/tmp/docker-build-push-SUdVri/tmp-2620-yrtaXsT6EMER \
    --tag ghcr.io/private-org/prefix_service-name:latest \
    --metadata-file /tmp/docker-build-push-SUdVri/metadata-file \
    --no-cache \
    --push 
```

```bash
#19 exporting to image
#19 exporting manifest sha256:fd473f873531eedb63fcbfc7f2063173fe7c37906f86fd6f29c0e9823746f078 done
#19 exporting config sha256:33e7b336bdb3c94f6682cb7e2db04b71ac801c6889a9dec9747b3be2100733ac done
#19 exporting attestation manifest sha256:befbb50642d43f9b7032026d37b58b47b612b92688c8478f5a025fd834d6c1b3 done
#19 exporting manifest sha256:6dfb76ac8dc3735554d751ec768fb239a92e3d93cc3602482dd4e097f68d34a7 done
#19 exporting config sha256:2049757826ecff0cadcbb910bfada5bb0d6a4a67ba7b8a66423079a2185e67cb done
#19 exporting attestation manifest sha256:30314bebf6b990e7f443aea31c4199d8c598dbe26396297c11f80095bf0ec542 done
#19 exporting manifest list sha256:287a09647d576f877425b12901b5ff03ab1f1d624adf6dd8df48df90c5b6ff2a done
#19 pushing layers
#19 pushing layers 6.7s done
#19 pushing manifest for ghcr.io/private-org/prefix_service-name:latest@sha256:287a09647d576f877425b12901b5ff03ab1f1d624adf6dd8df48df90c5b6ff2a
#19 pushing manifest for ghcr.io/private-org/prefix_service-name:latest@sha256:287a09647d576f877425b12901b5ff03ab1f1d624adf6dd8df48df90c5b6ff2a 1.8s done
```
