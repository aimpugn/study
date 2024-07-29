# du, disk usage

- [du, disk usage](#du-disk-usage)
    - [특정 디렉토리의 용량 확인](#특정-디렉토리의-용량-확인)
    - [특정 디렉토리의 N-레벨 하위디렉토리까지 확인](#특정-디렉토리의-n-레벨-하위디렉토리까지-확인)
    - [하위 디렉토리들의 용량 확인](#하위-디렉토리들의-용량-확인)
    - [가장 용량이 큰 N개의 디렉토리](#가장-용량이-큰-n개의-디렉토리)
    - [파일 크기로 정렬](#파일-크기로-정렬)
    - [du와 sort를 이용한 파일 크기 정렬 명령어](#du와-sort를-이용한-파일-크기-정렬-명령어)
        - [구성 요소 분석](#구성-요소-분석)
        - [작동 원리](#작동-원리)
        - [실행 예시](#실행-예시)
        - [변형 및 추가 옵션](#변형-및-추가-옵션)
        - [주의사항](#주의사항)
        - [다른 운영 체제에서의 사용](#다른-운영-체제에서의-사용)
        - [성능 최적화](#성능-최적화)
    - [참고](#참고)

## 특정 디렉토리의 용량 확인

```shell
# `s`: Display only the total size of the specified directory, do not display file size totals for subdirectories.
# `h`: Print sizes in a human-readable format
sudo du -sh /var
```

## 특정 디렉토리의 N-레벨 하위디렉토리까지 확인

```shell
# `c`: produce a grand total
# sudo du -shc --max-depth=N /var/*
sudo du -shc --max-depth=1 /var/*

#77G      /var/lib
#24K      /var/db
#4.0K    /var/empty
#4.0K    /var/local
#4.0K    /var/opt
#196K    /var/spool
#4.0K    /var/games
#3.3G    /var/log
#5.0G    /var/cache
#28K    /var/tmp
#85G    /var
#85G    total
```

## 하위 디렉토리들의 용량 확인

```shell
du -h -d 1 .

8.6M    ./interface
 64K    ./notification-service
232K    ./util
 52K    ./coretelemetry
1.3M    ./corepay
128K    ./url-zipper
 32K    ./.github
 24K    ./lambda
724K    ./some-qwerty-api-service
 60K    ./fixtures
 96M    ./.git
240K    ./corewebhook
1.3M    ./portdoc
118M    ./.idea
```

## 가장 용량이 큰 N개의 디렉토리

```shell
# sudo du -h /var/ | sort -rh | head -N
sudo du -h /var/ | sort -rh | head -5
```

## [파일 크기로 정렬](https://serverfault.com/a/156648)

```bash
du -hs * | sort -h

brew install coreutils
du -hs * | gsort -h
```

파일 크기 순서로 파일 경로를 보여주는 CLI 명령어에 대해 상세히 설명하겠습니다.

## du와 sort를 이용한 파일 크기 정렬 명령어

Unix 및 Linux 시스템에서 파일 크기 순서로 파일 경로를 보여주는 가장 일반적인 명령어 조합은 다음과 같습니다:

```bash
du -sh * | sort -rh
```

이 명령어의 구성 요소를 자세히 분석해보겠습니다.

### 구성 요소 분석

1. `du` (Disk Usage):
   - 파일 및 디렉토리의 디스크 사용량을 표시하는 명령어입니다.

2. `du` 옵션:
   - `-s`: summarize, 각 인자에 대해 총 사용량만 표시합니다.
   - `-h`: human-readable, 사람이 읽기 쉬운 형식(K, M, G 등)으로 크기를 표시합니다.
   - `*`: 현재 디렉토리의 모든 파일과 디렉토리를 대상으로 합니다.

3. `|`: 파이프(pipe) 연산자
   - 앞 명령어의 출력을 뒤 명령어의 입력으로 전달합니다.

4. `sort`: 텍스트 정렬 명령어

5. `sort` 옵션:
   - `-r`: reverse, 역순(내림차순) 정렬을 수행합니다.
   - `-h`: human-numeric-sort, 사람이 읽기 쉬운 숫자 형식(K, M, G 등)을 인식하여 정렬합니다.

### 작동 원리

1. `du -sh *`가 현재 디렉토리의 모든 항목의 디스크 사용량을 계산합니다.
2. 이 결과가 `sort` 명령어로 파이프됩니다.
3. `sort -rh`가 인간 친화적 숫자 형식을 인식하여 크기별로 내림차순 정렬합니다.

### 실행 예시

```bash
$ du -sh * | sort -rh
1.5G    largedir
234M    bigfile.zip
15M     documents
2.3M    script.py
500K    notes.txt
4.0K    empty_dir
```

### 변형 및 추가 옵션

1. 전체 경로 표시:

   ```bash
   du -sh * | sort -rh | sed 's/^/$(pwd)\//'
   ```

2. 특정 깊이까지만 검색:

   ```bash
   du -h --max-depth=2 | sort -rh
   ```

3. 특정 파일 형식만 검색:

   ```bash
   find . -name "*.txt" -type f -exec du -sh {} + | sort -rh
   ```

4. 상위 N개 결과만 표시:

   ```bash
   du -sh * | sort -rh | head -n 5
   ```

### 주의사항

1. 대용량 디렉토리에서는 실행 시간이 길어질 수 있습니다.
2. 권한 문제로 일부 파일/디렉토리의 크기를 읽지 못할 수 있습니다.
3. 심볼릭 링크는 실제 파일 크기가 아닌 링크 자체의 크기만 표시될 수 있습니다.

### 다른 운영 체제에서의 사용

1. macOS:
   - 기본적으로 동일한 명령어가 작동하지만, `gnu-coreutils`를 설치하여 더 다양한 옵션을 사용할 수 있습니다.

2. Windows (PowerShell):

   ```powershell
   Get-ChildItem | Sort-Object Length -Descending | Select-Object Name, Length
   ```

### 성능 최적화

대규모 디렉토리에서의 성능 향상을 위해:

1. `du`의 `--apparent-size` 옵션 사용: 실제 디스크 사용량이 아닌 파일 크기만 계산합니다.
2. `find`와 결합하여 병렬 처리:

   ```bash
   find . -type f -print0 | xargs -0 -P$(nproc) du -sh | sort -rh
   ```

이 명령어와 그 변형들을 이용하면 파일 시스템의 구조와 용량을 효과적으로 분석할 수 있습니다. 시스템 관리, 디스크 정리, 용량 계획 등 다양한 작업에 유용하게 활용될 수 있습니다.

## 참고

- [How to Get the Size of a Directory in Linux](https://linuxize.com/post/how-get-size-of-file-directory-linux/)
