# comm

- [comm](#comm)
    - [comm?](#comm-1)
    - [옵션: 특정 열 제외](#옵션-특정-열-제외)
    - [예제](#예제)

## comm?

`comm` 명령어는 두 개의 정렬된 파일을 비교하여 공통된 라인과 각 파일에만 존재하는 라인을 출력하는 유틸리티입니다.

기본적인 사용법은 다음과 같습니다. 이 명령어는 세 개의 열로 결과를 출력합니다. 각 열은 탭 문자로 구분됩니다.

```bash
comm [옵션] 파일1 파일2
```

`comm` 명령어는 세 개의 열로 결과를 출력합니다:

1. 첫 번째 파일에만 있는 라인
2. 두 번째 파일에만 있는 라인
3. 두 파일 모두에 있는 라인

```bash
[첫번째 파일에만 있음] [두번째 파일에만 있음] [두 파일 모두에 있음]
    apple
                                        banana
                                        cherry
                        date
```

## 옵션: 특정 열 제외

- `-1`: 첫 번째 파일에만 있는 라인을 출력하지 않습니다.
- `-2`: 두 번째 파일에만 있는 라인을 출력하지 않습니다.
- `-3`: 두 파일 모두에 있는 라인을 출력하지 않습니다.

- `a` 파일에만 있는 것 제외

    ```bash
    ❯ comm -1 a b
        banana
        cherry
    date
    ```

- `b` 파일에만 있는 것 제외

    ```bash
    ❯ comm -2 a b
    apple
        banana
        cherry
    ```

- `a`, `b` 파일 모두에 있는 것 제외

    ```bash
    ❯ comm -3 a b
    apple
        date
    ```

## 예제

- `a` 파일:

    ```txt
    # pbpaste > a
    apple
    banana
    cherry
    ```

- `b` 파일:

    ```txt
    # pbpaste > b
    banana
    cherry
    date
    ```

```bash
❯ comm a b
apple
        banana
        cherry
    date
```

```bash
❯ comm b a
    apple
        banana
        cherry
date
```

이 결과는 다음과 같이 해석할 수 있습니다:
- `apple`은 첫 번째 파일에만 있습니다.
- `banana`와 `cherry`는 두 파일 모두에 있습니다.
- `date`는 두 번째 파일에만 있습니다.
