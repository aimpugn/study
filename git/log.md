# git log

- [git log](#git-log)
    - [git log?](#git-log-1)
    - [특정 파일의 커밋 히스토리 보기](#특정-파일의-커밋-히스토리-보기)
    - [파일의 변경 사항만 보기](#파일의-변경-사항만-보기)
    - [특정 커밋부터의 히스토리 보기](#특정-커밋부터의-히스토리-보기)
    - [git log --graph --oneline --date-order](#git-log---graph---oneline---date-order)
    - [git log --oneline --decorate --graph --all](#git-log---oneline---decorate---graph---all)
        - [`| *`와`* |`의 차이?](#-와-의-차이)
    - [두 브랜치 사이의 커밋 보기](#두-브랜치-사이의-커밋-보기)

## git log?

- 커밋 히스토리 보기

## 특정 파일의 커밋 히스토리 보기

```bash
git log -- <PATH_TO_FILE>
```

## 파일의 변경 사항만 보기

```bash
git log -p -- <PATH_TO_FILE>
```

- `-p`, `--patch`: 파일의 커밋 히스토리와 함께 각 커밋에서의 변경 사항을 보여준다

## 특정 커밋부터의 히스토리 보기

```bash
git log <COMMIT_HASH>..HEAD -- <PATH_TO_FILE>
```

```bash
# git log feature/B..feature/A
# - feature/B에는 없고 feature/A에만 있는 커밋들을 보여줍니다.
# - 출력된 커밋이 없다면, feature/A가 feature/B의 모든 내용을 포함하고 있음을 의미합니다.
g log origin/release/240530..origin/release/240613

```

## git log --graph --oneline --date-order

```shell
git log --graph --oneline --date-order develop release/220831 feature/tosspayments-grpc-exception
```

[Introduction to GitHub Desktop: A GUI Enhancement to a CLI Approach](https://soshace.com/introduction-to-github-desktop-a-gui-enhancement-to-a-cli-approach/)

```shell
git log --all \
    --color \
    --graph \
    --pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset' \
    --abbrev-commit 
```

```shell
git log --since='2016-05-06' --pretty=format:"%h %an %cd" --graph --first-parent
```

## git log --oneline --decorate --graph --all

커밋의 상대적인 위치와 브랜치 간의 관계를 나타낸다

```shell
git log --oneline --decorate --graph --all
```

```log
* 6feea3867 fgsdfgdfsgdsfg              // `A` 브랜치와 `main` 브랜치가 합쳐진 병합 커밋
| * 39150730f sdfgrtegwergwerg
| * 17b3971e4 gergfd
| * a8fde270a dddddddddddddddddd        
|/                                      // `A` 브랜치 생성
* 1cdb2840f dgdsfgtgretgdfgdsfg         // `main` 브랜치에서 `a8fde270a` 다음에 이루어진 커밋
| * 3ab87505b asdfregwerg

...

| * | 3f0336c44 hrjtykr67uk896056f
| * | cff012165 y564y6hrtdhasff
| * | c972900b7 etfwergvsfdvhtyi
| * | 5983c6705 ytehjkjuylk,tui,gnmbn
| |/                                    // `B-fix` 브랜치 생성
| * bea1075ca gdfhdtfhokrpthm           // `B-additional` 브랜치와 B 브랜치가 합쳐진 병합 커밋
| | * ec66bcda2 gfnrdfgkhjperotj
| | * 46a5fef9f yuh5yj4576uty
| | * b85d0e9df hmghj,mghj,hgj,
| | * badde3ef4 fghjgfhjyukyuk
| | * e7d58cc8b o897o987o98l;iuhuk      // `B-additional` 브랜치의 커밋들
| |/                                    // `B-additional` 브랜치 생성
| * 02d4c2966 sdgfsregsertghsrethtr
|/                                      // `B` 브랜치 생성
```

- `*`: 특정 커밋을 나타낸다. 각 줄의 시작은 한 개의 커밋을 의미한다
- `|`: 브랜치 라인을 나타낸다. 하나의 `|`는 하나의 브랜치를 의미한다
- `| *`: 브랜치에서의 커밋을 나타낸다. 여기서 `|`는 브랜치의 경로이고, `*`는 해당 브랜치에서의 커밋이다
- `|/`: 브랜치가 병합되는 지점을 나타낸다

### `| *`와`* |`의 차이?

- 브랜치와 커밋 사이의 관계와 위치를 나타내는 기호들의 순서는 여러 요인에 따라 다를 수 있다
- 여기서 `| *`와`* |`의 차이는 크게 중요하지 않다
- 어떤 경우에는 메인 브랜치와 함께 다른 브랜치의 커밋을 나타낼 때 `* |` 형식이 사용될 수 있다
- `* |` 형식이 사용된 것은 `B-fix` 브랜치의 커밋들을 나타내기 위해서다
- 이는 로그 출력의 시각적 표현의 일부로, `main` 브랜치와 함께 다른 브랜치의 커밋을 나타내는 방식이다

## 두 브랜치 사이의 커밋 보기

두 브랜치 사이의 커밋 차이를 보고 싶을 때 사용할 수 있는 CLI 명령어는 `git log`와 함께 두 점(`..`) 또는 세 점(`...`)을 사용하는 방법입니다. 이 두 방법은 각각 다른 결과를 제공합니다.

1. **두 점(`..`) 사용하기**

    이 방법은 한 브랜치에는 있지만 다른 브랜치에는 없는 커밋들을 보여줍니다.

    예를 들어, `branch-X`가 `master`에 비해 가지고 있는 추가적인 커밋들을 보고 싶다면 다음과 같이 사용할 수 있습니다:

    ```bash
    # `master` 브랜치에는 없고 `branch-X` 브랜치에만 있는 커밋들 보기
    git log master..branch-X
    ```

    반대로 `branch-X`에는 없고 `master`에만 있는 커밋들을 보고 싶다면 인자를 뒤집어서 실행합니다:

    ```bash
    git log branch-X..master
    ```

2. **세 점(`...`) 사용하기**

    이 방법은 두 브랜치 간의 대칭적 차이를 보여줍니다.
    즉, 두 브랜치에서 공통적으로 발생하지 않은 커밋들을 모두 보여줍니다.

    예를 들어, `master`와 `branch-X` 간에 양쪽 모두에 없는 커밋들을 확인하고 싶다면 다음과 같이 사용할 수 있습니다:

    ```bash
    git log master...branch-X
    ```

    이 명령은 `master`와 `branch-X` 두 브랜치에서 각각 독립적으로 발생한 커밋들을 모두 보여줍니다.
    이 방법은 두 브랜치가 어떻게 다른지 전체적인 차이를 파악하고 싶을 때 유용합니다 [1].
