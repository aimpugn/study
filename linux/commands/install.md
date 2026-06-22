# install

## `install`이란?

`install`은 파일을 목적지로 **복사하면서 권한(mode)·소유자(owner)·그룹(group)까지 한 번에 지정**하는 coreutils 명령어다. 한 번의 호출이 `cp` + `chmod` + `chown`(필요하면 `mkdir`까지)을 합친 일을 한다. 이름 그대로 빌드·배포 스크립트의 "설치 단계"(`make install`)에서, 또는 서비스 계정용 디렉토리·설정 파일을 올바른 소유·권한으로 깔 때 쓴다.

## 왜 cp가 아니라 install인가

- `cp`로 옮긴 뒤 `chmod`·`chown`을 따로 거는 대신 **복사 시점에 속성을 한 번에** 지정한다(중간에 잘못된 권한으로 노출되는 창을 줄인다).
- `install -d`로 **권한·소유자를 지정하며 디렉토리를 생성**할 수 있다(`mkdir` + `chmod` + `chown`).
- 기본 권한이 **0755**이며, 타임스탬프를 보존하지 않고 현재 시각으로 둔다(`cp`와 다른 점). 보존하려면 `-p`.

## 호출 형식

```sh
install [옵션]... 원본 대상파일          # 파일 하나 → 지정 경로로
install [옵션]... 원본... 디렉토리        # 여러 파일 → 디렉토리 안으로
install -d [옵션]... 디렉토리...          # 디렉토리 생성(내용 복사 아님)
```

## 주요 옵션

- `-m, --mode=MODE` : 권한 설정(기본 `0755`). 예: `-m 640`.
- `-o, --owner=OWNER` : 소유자 설정(보통 루트 권한 필요).
- `-g, --group=GROUP` : 그룹 설정.
- `-d, --directory` : 인자를 디렉토리로 보고 생성(`mkdir -p`처럼 중간 경로 포함). 권한·소유자 동시 지정 가능.
- `-D` : 대상 파일의 **상위 디렉토리들을 필요 시 생성**한 뒤 복사. 예: `install -D -m 644 src /etc/app/conf/app.conf`.
- `-t, --target-directory=DIR` : 대상 디렉토리를 명시.
- `-p, --preserve-timestamps` : 원본 타임스탬프 유지.
- `-s, --strip` : 실행 파일의 심볼 테이블 제거(바이너리 설치 시 용량↓).
- `-b` : 기존 파일이 있으면 백업.
- `-C, --compare` : 내용이 같으면 복사하지 않음(불필요한 갱신 방지).
- `-v, --verbose` : 처리 내역 출력.

## 사용 예시 (Nexus 같은 서비스 계정 구성에서)

Nexus는 보통 전용 비로그인 서비스 계정(`nexus`)으로 돌린다. 그 계정 소유로 작업·데이터 디렉토리와 설정 파일을 한 번에 깔 때 `install`이 편하다.

```sh
# 데이터/설정 디렉토리를 nexus 소유·권한으로 생성  (= mkdir + chown + chmod)
install -d -o nexus -g nexus -m 755 /opt/sonatype-work/nexus3/etc

# 설정 파일을 소유자·권한까지 지정하며 배치  (= cp + chown + chmod)
install -o nexus -g nexus -m 640 nexus.properties /opt/sonatype-work/nexus3/etc/nexus.properties

# 상위 경로가 없을 수도 있으면 -D 로 한 번에
install -D -o nexus -g nexus -m 640 nexus.properties /opt/sonatype-work/nexus3/etc/nexus.properties
```

같은 일을 `cp`로 하면 `mkdir -p` → `cp` → `chown` → `chmod` 네 단계가 된다. `install`은 이를 한 줄로 묶고, 권한이 잘못 잡힌 채 파일이 잠깐 노출되는 구간도 없앤다.

## 주의사항

- `-o`, `-g`로 소유자/그룹을 바꾸려면 보통 **루트 권한**이 필요하다.
- `install`은 **디렉토리를 재귀 복사하지 않는다**(`cp -r`과 다름). 디렉토리 트리 복사 용도가 아니라 "파일 배치 + 속성 지정" 용도다.
- 기본적으로 타임스탬프를 보존하지 않는다 → 빌드 산출물 비교가 필요하면 `-p` 또는 `-C`.
- `-D`는 GNU(리눅스) 옵션으로, BSD/macOS의 `install`과 동작이 다르다(`linux/commands` 기준으로 정리).
