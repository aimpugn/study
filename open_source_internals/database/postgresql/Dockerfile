# PostgreSQL 18.6을 소스 줄 단위로 따라가기 위한 디버그 빌드입니다.
# 최적화를 끄고(-O0) 매크로까지 포함한 디버그 정보(-g3)를 남겨, gdb가 실제 소스 줄에서 멈추고
# 변수 값을 읽을 수 있게 합니다. 실행 속도가 운영 빌드와 다르므로 성능 측정에는 쓰지 않습니다.
FROM debian:bookworm-slim

ARG PG_VERSION=18.6
ARG PG_SHA256=555610c24d53e4316da5b7d3fc25c279d96856d5e0e23ee308c328c5fa881d9f

# 빌드 도구와 관측 도구를 한 번에 설치합니다. ICU는 이번 실습에 필요하지 않아 제외합니다.
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      build-essential pkg-config bison flex libreadline-dev zlib1g-dev \
      gdb strace procps ca-certificates curl bzip2 \
 && rm -rf /var/lib/apt/lists/*

# 공식 배포처의 tarball을 공개된 SHA-256과 대조한 뒤에만 풉니다.
WORKDIR /src
RUN curl -fsSLO "https://ftp.postgresql.org/pub/source/v${PG_VERSION}/postgresql-${PG_VERSION}.tar.bz2" \
 && echo "${PG_SHA256}  postgresql-${PG_VERSION}.tar.bz2" | sha256sum -c - \
 && tar -xjf "postgresql-${PG_VERSION}.tar.bz2" \
 && rm "postgresql-${PG_VERSION}.tar.bz2"

# 소스 트리를 이미지에 남겨 gdb가 줄을 보여 주고, 학습자가 같은 커밋의 소스를 읽게 합니다.
WORKDIR /src/postgresql-${PG_VERSION}
RUN ./configure --prefix=/usr/local/pgsql --enable-debug --without-icu \
      CFLAGS="-O0 -g3 -fno-omit-frame-pointer" \
 && make -j"$(nproc)" \
 && make install \
 && make -C contrib/pageinspect install

# 서버는 root로 실행되지 않으므로 일반 계정으로 데이터 디렉터리를 만듭니다.
RUN useradd --create-home postgres \
 && mkdir -p /var/lib/pgsql/data \
 && chown postgres /var/lib/pgsql/data
USER postgres
ENV PATH=/usr/local/pgsql/bin:$PATH \
    PGDATA=/var/lib/pgsql/data
RUN initdb --no-locale -E UTF8 -D "$PGDATA"

# TCP를 열지 않고 Unix socket만 사용합니다. 실습 명령은 docker exec로 컨테이너 안에서 실행합니다.
CMD ["postgres", "-D", "/var/lib/pgsql/data", "-c", "listen_addresses="]
