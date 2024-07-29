# gh

- [gh](#gh)
    - [`gh`?](#gh-1)
    - [GitHub CLI 설치](#github-cli-설치)
    - [GitHub CLI 인증](#github-cli-인증)
    - [특정 커밋 해시와 관련된 PR 찾기](#특정-커밋-해시와-관련된-pr-찾기)
    - [특정 커밋 해시로 PR 찾기](#특정-커밋-해시로-pr-찾기)
    - [특정 커밋 해시가 포함된 PR 찾기](#특정-커밋-해시가-포함된-pr-찾기)

## `gh`?

GitHub CLI 도구

Git에서 특정 커밋 해시와 관련된 Pull Request(PR)를 찾는 등 다양한 GitHub 리소스와 상호작용할 수 있습니다.

## GitHub CLI 설치

먼저 GitHub CLI가 설치되어 있는지 확인하세요. 설치되지 않았다면 아래와 같이 설치할 수 있습니다.

- macOS:

  ```bash
  brew install gh
  ```

- Windows:
  GitHub CLI의 [설치 가이드](https://github.com/cli/cli#installation)를 참고하세요.

- Linux:

  ```bash
  sudo apt install gh
  ```

## GitHub CLI 인증

GitHub CLI를 사용하려면 먼저 GitHub 계정으로 인증해야 합니다.

```bash
gh auth login
```

- 브라우저 통해서 인증할 경우 로그

    ```bash
    ❯ gh auth login
    ? What account do you want to log into? GitHub.com
    ? What is your preferred protocol for Git operations on this host? HTTPS
    ? Authenticate Git with your GitHub credentials? Yes
    ? How would you like to authenticate GitHub CLI? Login with a web browser

    ! First copy your one-time code: 107D-A8B4
    Press Enter to open github.com in your browser...
    ✓ Authentication complete.
    - gh config set -h github.com git_protocol https
    ✓ Configured git protocol
    ✓ Logged in as aimpugn
    ```

## 특정 커밋 해시와 관련된 PR 찾기

이제 특정 커밋 해시와 관련된 PR을 찾는 명령어는 다음과 같습니다.

## 특정 커밋 해시로 PR 찾기

```bash
gh pr list --search <commit-hash>
```

## 특정 커밋 해시가 포함된 PR 찾기

GitHub API를 사용하여 특정 커밋 해시와 관련된 PR을 찾으려면 다음 명령어를 사용하세요.

```bash
# gh api -X GET "repos/{owner}/{repo}/commits/{commit-sha}/pulls" --jq '.[].html_url'
gh api -X GET "repos/portone-io/some-A-serivce/commits/1f8f13d9/pulls" --jq '.[].html_url'

gh api -X GET "repos/portone-io/some-B-serivce/commits/4eb3d429/pulls" --jq '.[].html_url'
```

- `{owner}`: 해당 리포지토리 소유자
- `{repo}` : 리포지토리 이름
- `{commit-sha}`: 커밋 해시

host:
port: 8101
