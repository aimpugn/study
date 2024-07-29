# container registry

- [container registry](#container-registry)
    - [Working with the Container registry](#working-with-the-container-registry)
    - [login](#login)
    - [docker build from repo](#docker-build-from-repo)
    - [push to ghcr.io](#push-to-ghcrio)

## [Working with the Container registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)

> The Container registry stores container images within your organization or personal account, and allows you to associate an image with a repository.

- `ghcr`: **G**it**H**ub **C**ontainer **R**egistry

## login

```shell
echo $CR_PAT | docker login ghcr.io -u $(git config user.name) --password-stdin
# Login Succeeded
```

## docker build from repo

```shell
docker build https://github.com/docker/rootfs.git#container:docker
```

```shell
docker build https://github.com/some-org/core-r1-docker.git#main:apache2 -t core_apache2
```

## push to ghcr.io

- [Pushing container images](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#pushing-container-images)

```shell
docker push ghcr.io/NAMESPACE/IMAGE_NAME:latest
```

```shell
/usr/bin/docker buildx build \
    --build-arg XDEBUG_ENABLE=true \
    --iidfile /tmp/docker-build-push-q7g18E/iidfile \
    --platform linux/amd64,linux/arm64 \
    --provenance false \
    --secret id=GIT_AUTH_TOKEN,src=/tmp/docker-build-push-q7g18E/tmp-2492-4mPPCwFgsyix \
    --tag ghcr\.io/some-orgcore_apache2:latest \
    --metadata-file /tmp/docker-build-push-q7g18E/metadata-file \
    --no-cache \
    --push https://github.com/some-org/core-r1-docker.git#2f821173fdaf4a774d2dc5b02522b06a58bdc18b:apache2
```

- [docker buildx build](https://docs.docker.com/engine/reference/commandline/buildx_build/): build using buildkit
- `--iidfile`: Write the image ID to the file
- `--platform`: Set platform if server is multi-platform capable
- `--secret`:
    - [Build Images with Secrets Locally](https://render.com/docs/docker-secrets#building-images-with-secrets-locally)
    - [Read more about docker secret commands](https://docs.docker.com/engine/swarm/secrets/#read-more-about-docker-secret-commands)
    - [Create a service with secrets (--secret)](https://docs.docker.com/engine/reference/commandline/service_create/#secret)
    - [Add or remove secrets (--secret-add, --secret-rm)](https://docs.docker.com/engine/reference/commandline/service_update/#secret-add)
- `--metadata-file`:  The metadata will be written as a JSON object to the specified file.
- `--push`: Shorthand for `--output=type=registry`. Will automatically push the build result to registry.
- `--output`:
    - `docker build`에서는 컨테이너 이미지 생성하고 이를 `docker images`로 내보내면(export) 모든 빌드가 끝난다
    - `type`
        - `image`: The `image` exporter writes the build result as an image or a manifest list. When using `docker` driver the image will appear in docker images. Optionally, image can be automatically pushed to a registry by specifying attributes.
        - `registry`: `type=image,push=true`의 단축
