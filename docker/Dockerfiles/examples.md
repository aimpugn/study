# Examples

- [Examples](#examples)
    - [`ubuntu:22.04`, `mockery`, `go`](#ubuntu2204-mockery-go)

## `ubuntu:22.04`, `mockery`, `go`

```Dockerfile
FROM ubuntu:22.04

RUN apt update
RUN apt install -y git wget

RUN mkdir /downloads && cd /downloads
RUN wget https://github.com/vektra/mockery/releases/download/v2.38.0/mockery_2.38.0_Linux_x86_64.tar.gz -q -O mockery.tar.gz
RUN tar -xf mockery.tar.gz && mv mockery /usr/local/bin/mockery

RUN cd /downloads
RUN wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz -q -O go1.21.5.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
RUN echo 'PATH="$PATH:/usr/local/go/bin"' >> ~/.bashrc
RUN go version

WORKDIR /app

COPY . .

RUN git config \
    --global \
    url."https://aimpugn:${PAT}@github.com/".insteadOf "https://github.com/"


# 컨테이너가 시작될 때 실행되는 기본 명령을 지정
# `bash` 셸을 시작하여 컨테이너가 셸 세션으로 시작되도록 설정
# 사용자가 컨테이너 내에서 셸을 통해 직접 명령을 입력하고 싶을 때 유용하다
CMD ["bash"]
```
