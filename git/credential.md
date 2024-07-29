# git credential

## approve: 인증 정보 추가

```yaml
- name: Add credentials
  run: |
    git credential approve <<EOF
    protocol=https
    host=github.com
    username=your-username
    password=${{ secrets.GITHUB_TOKEN }}
    EOF
```

이 부분에서는 Git 인증 정보를 추가하는 작업이 수행됩니다:

- Heredoc(Here Document) 구문: `<<EOF ... EOF`는 여러 줄에 걸쳐 텍스트를 입력하는 Bash 스크립트의 heredoc 구문입니다. 이 구문을 사용하여 여러 줄의 인증 정보를 `git credential approve` 명령어로 전달합니다.

- Git Credential 추가: `git credential approve` 명령어는 Git에 사용할 인증 정보를 추가합니다. 여기서는 GitHub의 호스트(`github.com`), 사용자 이름(`your-username`), 그리고 GitHub에서 생성한 개인 접근 토큰(`GITHUB_TOKEN`)을 사용합니다.

- GitHub Personal Access Token (PAT): `${{ secrets.GITHUB_TOKEN }}`은 GitHub Actions에서 제공하는 특수 변수로, 이는 GitHub repository에 대한 접근 권한을 부여하는 토큰입니다. 이 토큰은 GitHub Actions에서만 사용 가능하며, 이를 통해 스크립트는 GitHub의 private repository에 안전하게 접근할 수 있습니다.

## credential.helper

Git이 인증을 필요로 하는 작업(예: 리포지토리에 푸시하기)을 수행할 때 사용자의 자격 증명을 어떻게 처리할지 구성합니다.

- 메모리 cache에 저장

    ```bash
    # https://git-scm.com/book/ko/v2/Git-%EB%8F%84%EA%B5%AC-Credential-%EC%A0%80%EC%9E%A5%EC%86%8C
    git config --global credential.helper cache
    ```

    이 명령어는 자격 증명을 메모리에 일시적으로 저장합니다.
    이렇게 하면 연속적인 Git 명령에서 자격 증명을 자주 재입력할 필요가 없어집니다.
    기본적으로 캐시는 15분 후에 만료되어 자격 증명이 사라집니다.

    `cache` 방법은 자격 증명을 일시적으로만 메모리에 저장하기 때문에 다소 더 안전합니다.

- 기본 파일(`~/.git-credentials`)에 저장

    ```bash
    # 파일에 저장. 기본 파일 경로: ~/.git-credentials
    git config --global credential.helper store

    # root@80f41657ce64:/var/www/api# cat ~/.git-credentials
    # https://<계정>:<토큰>@github.com
    ```

    이 명령어는 자격 증명을 디스크에 영구적으로 저장하도록 저장 방식을 변경합니다.
    기본 저장 파일은 `~/.git-credentials`입니다.
    이 명령어를 실행한 후 처음으로 Git 인증을 사용하면 자격 증명이 `.git-credentials` 파일에 평문 형식으로 저장됩니다.

    `store` 방법은 자격 증명을 무기한 디스크에 저장하여, 다른 사람이 파일 시스템에 접근할 수 있다면 위험할 수 있습니다.

- 임의 경로의 파일에 저장

    ```bash
    # 임의 경로의 파일에 저장
    git config --global credential.helper 'store --file ~/.my-credentials'
    ```

    이 변형된 `store` 명령어는 기본 `.git-credentials` 파일 대신 `~/.my-credentials`라는 사용자 지정 파일 경로에 자격 증명을 저장하도록 지정합니다.
    이전 `store` 명령어와 동일하게 작동하지만 지정된 파일에 자격 증명을 저장합니다.
