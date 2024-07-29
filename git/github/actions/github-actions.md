# GitHub Actions

- [GitHub Actions](#github-actions)
    - [GitHub Actions `yml`(`yaml`) 파일 작성 방법](#github-actions-ymlyaml-파일-작성-방법)
        - [Example](#example)
        - [Contexts](#contexts)
            - [`github`](#github)
            - [`env`](#env)
            - [`job`](#job)
            - [`steps`](#steps)
            - [`runner`](#runner)
    - [steps](#steps-1)
    - [PAT 대신 `github.token` 또는 `secrets.GITHUB_TOKEN` 등 사용](#pat-대신-githubtoken-또는-secretsgithub_token-등-사용)
        - [GitHub Actions에서 `PAT` 사용의 문제점](#github-actions에서-pat-사용의-문제점)
        - [GitHub Actions에서 `PAT`의 대안](#github-actions에서-pat의-대안)
            - [`github.token`](#githubtoken)
            - [`secrets.GITHUB_TOKEN`](#secretsgithub_token)
            - [`github.token`와 `secrets.GITHUB_TOKEN` 차이](#githubtoken와-secretsgithub_token-차이)
    - [deploy key](#deploy-key)

## GitHub Actions `yml`(`yaml`) 파일 작성 방법

### Example

```yaml
name: 현재 GitHub Action 이름

on:
  ## branch에서는 수동 트리거 버튼이 노출되지 않고, 검색해 보면 기본 브랜치에 합쳐진 후에 나오는 버그가 있다고 한다
  workflow_dispatch: ## 어떤 이벤트에 트리거 될 것인지
    inputs: ## 입력을 받게 된다
      who_am_i: ## 입력 input 타이틀
        description: Name of someone who trigger this workflow ## 입력 input 설명
        required: true ## 필수 여부
        type: text ## 타입
      what_to_do:
        description: Select one to do
        required: true
        type: choice
        options:
          - test
  push:
    branches: ## 브랜치 조건 부여 가능. 그 외에 `path` 및 파일 확장자(`**.js`) 조건도 가능
      - feature/dockerize

## `job`의 `services`을 제외하고 전역으로 쓸 수 있는 환경 변수
## shell script 내에서도 바로 사용 가능
env:
  TEST_MYSQL_VERSION: 5.5.59
  TEST_MYSQL_HOST: localhost
  TEST_MYSQL_HOST_PORT: 3306
  TEST_MYSQL_CONTAINER_PORT: 33060
  TEST_MYSQL_USER: testdb
  TEST_MYSQL_PASSWORD: testdb
  TEST_MYSQL_DATABASE: some_db
  TEST_MYSQL_ROOT_PASSWORD: root

jobs:
  test-some-service:
    runs-on: ubuntu-20.04

    ## 현재 `job` 컨테이너에서 제공하게 될 컨테이너 서비스.
    ## 지금 실행되는 `job`의 `step`으로 `status`, `services.mysql.id`, `services.mysql.ports`, `services.mysql.network` 전달 가능
    services:
      ## postgres, redis, mysql, mariadb 등 사용하는 문서는 많은데, 
      ## 레퍼런스 문서에는`redis`와 `postgres`만 있다
      ## https://docs.github.com/en/actions/using-containerized-services/about-service-containers
      mysql:
        image: mysql:5.5.59
        ## `services`에서는 `env` 컨텍스트의 환경 변수를 사용할 수 없다
        env:
          MYSQL_DATABASE: some_db
          MYSQL_HOST: 127.0.0.1
          MYSQL_USER: testdb
          MYSQL_PASSWORD: testdb
          MYSQL_ROOT_PASSWORD: root
        ports:
          ## - 3306:3306 호스트 포트:컨테이너 포트
          ## 이후 `job` 컨텍스트에서 `service.port.3306 = 33063`로 설정
          - 33060:3306
        ## docker run 옵션들(https://docs.docker.com/engine/reference/run/##healthcheck)
        ## 꺽쇠 괄호(`>`)로 사용할 경우 에러 발생: `Error: Exit code 125 returned from process`
        ## 파이프(`|`)로 사용할 경우 에러 발생: `Error: Exit code 1 returned from process`
        ## `>-`만 사용 가능
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=3

    steps:
      - name: Print job context
        ## `env` 전역 환경 변수 외에, `step`별 환경변수 설정 가능. 
        env:
          ## `workflow`(eg. some-service-test)와 `action`에서 표현식(`expression`) 사용 가능
          ## `toJSON`: https://docs.github.com/en/actions/learn-github-actions/expressions##tojson
          CURRENT_ENV: ${{ toJSON(job) }}
        run: |
          echo "${CURRENT_ENV}"

      - name: Print log what triggered this workflow
        run: | ## 파이프(`|`)는 여러 줄 의미
          if [[ '${{ github.event_name }}' == 'workflow_dispatch' ]]; then
            initial_message="${{ github.event.inputs.who_am_i }} start workflow for ${{ github.event.inputs.what_to_do }}"
          else
            initial_message="'${{ github.ref_type }}:${{ github.ref_name }}' triggered this workflow"
          fi
          echo "When '${{ github.event_name }}' event, $initial_message"

      ## https://github.com/shivammathur/setup-php
      - name: Set up php5.6
        uses: shivammathur/setup-php@v2
        ## `with` 사용 시에는 `run`을 쓸 수 없음에 주의
        with:
          php-version: 5.6

      ## https://github.com/actions/checkout
      - name: Checkout branch ## 현재 리파지토리로부터 체크아웃
        uses: actions/checkout@v3
        ## `with` 없어도 현재 push 된 브랜치로부터 체크아웃
        ## with:
        ##  ref: 'feature/dockerize'
        ##  fetch-depth: '1'

      - name: Check php version and if possible to connect to mysql container
        run: |
          docker ps -a | grep mysql
          php --version
          php ./resource/scripts/test_mysql_connection.php --host=$TEST_MYSQL_HOST \
            --port=$TEST_MYSQL_PORT \
            --username=$TEST_MYSQL_USER \
            --password=$TEST_MYSQL_PASSWORD
```

### Contexts

- `github action` 실행 중 사용할 수 있는 [`context`](https://docs.github.com/en/actions/learn-github-actions/contexts) 샘플 데이터

#### `github`

```json
{
  "token": "***",
  "job": "test-some-service",
  "ref": "refs/heads/feature/some",
  "sha": "<SHA>",
  "repository": "someorg/some-service",
  "repository_owner": "someorg",
  "repository_owner_id": "<SOME_ID>",
  "repositoryUrl": "git://github.com/someorg/some-service.git",
  "run_id": "<RUN_ID>",
  "run_number": "55",
  "retention_days": "7",
  "run_attempt": "1",
  "artifact_cache_size_limit": "10",
  "repository_id": "<REPOSITORY_ID>",
  "actor_id": "<ACTOR_ID>",
  "actor": "aimpugn",
  "workflow": "some-service-test",
  "head_ref": "",
  "base_ref": "",
  "event_name": "push",
  "event": {
    "after": "<SHA>",
    "base_ref": null,
    "before": "8c236f4e46b8b37e3123f8710f2e17b43c3c89a6",
    "commits": [
      {
        "author": {
          "email": "aimpugn@gmail.com",
          "name": "rody",
          "username": "aimpugn"
        },
        "committer": {
          "email": "aimpugn@gmail.com",
          "name": "rody",
          "username": "aimpugn"
        },
        "distinct": true,
        "id": "<SHA>",
        "message": "chore: context 정보 출력",
        "timestamp": "2022-05-25T17:45:44+09:00",
        "tree_id": "<TREE_ID>",
        "url": "https://github.com/someorg/some-service/commit/<SHA>"
      }
    ],
    "compare": "https://github.com/someorg/some-service/compare/<HASH1>...<HASH2>",
    "created": false,
    "deleted": false,
    "forced": false,
    "head_commit": {
      "author": {
        "email": "aimpugn@gmail.com",
        "name": "rody",
        "username": "aimpugn"
      },
      "committer": {
        "email": "aimpugn@gmail.com",
        "name": "rody",
        "username": "aimpugn"
      },
      "distinct": true,
      "id": "<SHA>",
      "message": "context 정보 출력",
      "timestamp": "2022-05-25T17:45:44+09:00",
      "tree_id": "<TREE_ID>",
      "url": "https://github.com/someorg/some-service/commit/<SHA>"
    },
    "organization": {
      "avatar_url": "https://avatars.githubusercontent.com/u/<SOME_ID>?v=4",
      "description": "어떤 서비스 리파지토리입니다.",
      "events_url": "https://api.github.com/orgs/someorg/events",
      "hooks_url": "https://api.github.com/orgs/someorg/hooks",
      "id": <SOME_ID>,
      "issues_url": "https://api.github.com/orgs/someorg/issues",
      "login": "someorg",
      "members_url": "https://api.github.com/orgs/someorg/members{/member}",
      "node_id": "<RANDOM_NODE_ID>",
      "public_members_url": "https://api.github.com/orgs/someorg/public_members{/member}",
      "repos_url": "https://api.github.com/orgs/someorg/repos",
      "url": "https://api.github.com/orgs/someorg"
    },
    "pusher": {
      "email": "aimpugn@gmail.com",
      "name": "aimpugn"
    },
    "ref": "refs/heads/feature/some",
    "repository": {
      "allow_forking": false,
      "archive_url": "https://api.github.com/repos/someorg/some-service/{archive_format}{/ref}",
      "archived": false,
      "assignees_url": "https://api.github.com/repos/someorg/some-service/assignees{/user}",
      "blobs_url": "https://api.github.com/repos/someorg/some-service/git/blobs{/sha}",
      "branches_url": "https://api.github.com/repos/someorg/some-service/branches{/branch}",
      "clone_url": "https://github.com/someorg/some-service.git",
      "collaborators_url": "https://api.github.com/repos/someorg/some-service/collaborators{/collaborator}",
      "comments_url": "https://api.github.com/repos/someorg/some-service/comments{/number}",
      "commits_url": "https://api.github.com/repos/someorg/some-service/commits{/sha}",
      "compare_url": "https://api.github.com/repos/someorg/some-service/compare/{base}...{head}",
      "contents_url": "https://api.github.com/repos/someorg/some-service/contents/{+path}",
      "contributors_url": "https://api.github.com/repos/someorg/some-service/contributors",
      "created_at": 1616146610,
      "default_branch": "main",
      "deployments_url": "https://api.github.com/repos/someorg/some-service/deployments",
      "description": null,
      "disabled": false,
      "downloads_url": "https://api.github.com/repos/someorg/some-service/downloads",
      "events_url": "https://api.github.com/repos/someorg/some-service/events",
      "fork": false,
      "forks": 0,
      "forks_count": 0,
      "forks_url": "https://api.github.com/repos/someorg/some-service/forks",
      "full_name": "someorg/some-service",
      "git_commits_url": "https://api.github.com/repos/someorg/some-service/git/commits{/sha}",
      "git_refs_url": "https://api.github.com/repos/someorg/some-service/git/refs{/sha}",
      "git_tags_url": "https://api.github.com/repos/someorg/some-service/git/tags{/sha}",
      "git_url": "git://github.com/someorg/some-service.git",
      "has_downloads": true,
      "has_issues": true,
      "has_pages": false,
      "has_projects": true,
      "has_wiki": true,
      "homepage": null,
      "hooks_url": "https://api.github.com/repos/someorg/some-service/hooks",
      "html_url": "https://github.com/someorg/some-service",
      "id": <REPOSITORY_ID>,
      "is_template": false,
      "issue_comment_url": "https://api.github.com/repos/someorg/some-service/issues/comments{/number}",
      "issue_events_url": "https://api.github.com/repos/someorg/some-service/issues/events{/number}",
      "issues_url": "https://api.github.com/repos/someorg/some-service/issues{/number}",
      "keys_url": "https://api.github.com/repos/someorg/some-service/keys{/key_id}",
      "labels_url": "https://api.github.com/repos/someorg/some-service/labels{/name}",
      "language": "PHP",
      "languages_url": "https://api.github.com/repos/someorg/some-service/languages",
      "license": null,
      "master_branch": "main",
      "merges_url": "https://api.github.com/repos/someorg/some-service/merges",
      "milestones_url": "https://api.github.com/repos/someorg/some-service/milestones{/number}",
      "mirror_url": null,
      "name": "some-service",
      "node_id": "MDEwOlJlcG9zaXRvcnkzNDkzNzAwMjY=",
      "notifications_url": "https://api.github.com/repos/someorg/some-service/notifications{?since,all,participating}",
      "open_issues": 6,
      "open_issues_count": 6,
      "organization": "someorg",
      "owner": {
        "avatar_url": "https://avatars.githubusercontent.com/u/<SOME_ID>?v=4",
        "email": "someorg@legacy.email.com",
        "events_url": "https://api.github.com/users/someorg/events{/privacy}",
        "followers_url": "https://api.github.com/users/someorg/followers",
        "following_url": "https://api.github.com/users/someorg/following{/other_user}",
        "gists_url": "https://api.github.com/users/someorg/gists{/gist_id}",
        "gravatar_id": "",
        "html_url": "https://github.com/someorg",
        "id": <SOME_ID>,
        "login": "someorg",
        "name": "someorg",
        "node_id": "<RANDOM_NODE_ID>",
        "organizations_url": "https://api.github.com/users/someorg/orgs",
        "received_events_url": "https://api.github.com/users/someorg/received_events",
        "repos_url": "https://api.github.com/users/someorg/repos",
        "site_admin": false,
        "starred_url": "https://api.github.com/users/someorg/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/someorg/subscriptions",
        "type": "Organization",
        "url": "https://api.github.com/users/someorg"
      },
      "private": true,
      "pulls_url": "https://api.github.com/repos/someorg/some-service/pulls{/number}",
      "pushed_at": 1653468352,
      "releases_url": "https://api.github.com/repos/someorg/some-service/releases{/id}",
      "size": 32610,
      "ssh_url": "git@github.com:someorg/some-service.git",
      "stargazers": 1,
      "stargazers_count": 1,
      "stargazers_url": "https://api.github.com/repos/someorg/some-service/stargazers",
      "statuses_url": "https://api.github.com/repos/someorg/some-service/statuses/{sha}",
      "subscribers_url": "https://api.github.com/repos/someorg/some-service/subscribers",
      "subscription_url": "https://api.github.com/repos/someorg/some-service/subscription",
      "svn_url": "https://github.com/someorg/some-service",
      "tags_url": "https://api.github.com/repos/someorg/some-service/tags",
      "teams_url": "https://api.github.com/repos/someorg/some-service/teams",
      "topics": [],
      "trees_url": "https://api.github.com/repos/someorg/some-service/git/trees{/sha}",
      "updated_at": "2022-02-07T04:00:00Z",
      "url": "https://github.com/someorg/some-service",
      "visibility": "private",
      "watchers": 1,
      "watchers_count": 1
    },
    "sender": {
      "avatar_url": "https://avatars.githubusercontent.com/u/<ACTOR_ID>?v=4",
      "events_url": "https://api.github.com/users/aimpugn/events{/privacy}",
      "followers_url": "https://api.github.com/users/aimpugn/followers",
      "following_url": "https://api.github.com/users/aimpugn/following{/other_user}",
      "gists_url": "https://api.github.com/users/aimpugn/gists{/gist_id}",
      "gravatar_id": "",
      "html_url": "https://github.com/aimpugn",
      "id": <ACTOR_ID>,
      "login": "aimpugn",
      "node_id": "MDQ6VXNlcjI4NTcwNDMy",
      "organizations_url": "https://api.github.com/users/aimpugn/orgs",
      "received_events_url": "https://api.github.com/users/aimpugn/received_events",
      "repos_url": "https://api.github.com/users/aimpugn/repos",
      "site_admin": false,
      "starred_url": "https://api.github.com/users/aimpugn/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/aimpugn/subscriptions",
      "type": "User",
      "url": "https://api.github.com/users/aimpugn"
    }
  },
  "server_url": "https://github.com",
  "api_url": "https://api.github.com",
  "graphql_url": "https://api.github.com/graphql",
  "ref_name": "feature/some",
  "ref_protected": false,
  "ref_type": "branch",
  "secret_source": "Actions",
  "workspace": "/home/runner/work/some-service/some-service",
  "action": "__run",
  "event_path": "/home/runner/work/_temp/_github_workflow/event.json",
  "action_repository": "",
  "action_ref": "",
  "path": "/home/runner/work/_temp/_runner_file_commands/add_path_11x1xxx1-44e7-4355-99b1-34eae1d1c9a4",
  "env": "/home/runner/work/_temp/_runner_file_commands/set_env_11x1xxx1-44e7-4355-99b1-34eae1d1c9a4",
  "step_summary": "/home/runner/work/_temp/_runner_file_commands/step_summary_11x1xxx1-44e7-4355-99b1-34eae1d1c9a4"
}
```

#### `env`

```json
{
    "MYSQL_ROOT_PASSWORD": "root",
    "MYSQL_TEST_DATABASE": "some_db",
    "MYSQL_TEST_USERNAME": "testdb",
    "MYSQL_TEST_PASSWORD": "testdb",
    "MYSQL_TEST_PORT": "33060"
}
```

#### `job`

```json
{
  "status": "success",
  "container": {
    "network": "github_network_01af74a2cc7e43b69d5519255a0b2e26"
  },
  "services": {
    "mysql": {
      "id": "882f31ef40e4105c2080f7fc5c4d753cdb19f739d3896e2acc549d711a4bb1ce",
      "ports": {
        "3306": "33060"
      },
      "network": "github_network_01af74a2cc7e43b69d5519255a0b2e26"
    }
  }
}
```

#### `steps`

```json
{
  "fcf266db5e204f419dcc2d7b825ce95b": {
    "outputs": {},
    "outcome": "success",
    "conclusion": "success"
  }
}
```

#### `runner`

```json
{
    "os": "Linux",
    "arch": "X64",
    "name": "GitHub Actions 2",
    "tool_cache": "/opt/hostedtoolcache",
    "temp": "/home/runner/work/_temp",
    "workspace": "/home/runner/work/some-service"
}
```

## steps

- GitHub Actions에서 `run` 명령어를 사용할 때, 각각의 `run` *스텝은 별도의 쉘 프로세스에서 실행*된다. 즉, 각 run 스텝이 독립적인 실행 환경을 갖는다.
- 그러나 일부 설정, 특히 `git config` 설정과 같은 것들은 Git의 전역 설정에 영향을 미치므로, GitHub Actions 워크플로우의 다른 스텝에서도 유지된다
    - 예를 들어, `git config --global` 명령어를 사용하면 이 설정은 워크플로우의 모든 후속 단계에서 유지된다
    - 그 이유는 이 설정이 *워크플로우가 실행되는 가상 환경의 전역 Git 설정에 적용되기 때문*입니다.
- 다만, 환경 변수나 `export` 명령어를 사용하여 설정된 변수들은 해당 `run` 스텝이 완료되면 초기화되어서 다음 `run` 스텝에서 유지되지 않는다. 이를 유지하고 싶다면 GitHub Actions의 `env` 또는 `secrets` 같은 기능을 사용해야 한다

## PAT 대신 `github.token` 또는 `secrets.GITHUB_TOKEN` 등 사용

- 참고 링크
    - [Automatic token authentication](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow)
    - [Using the `GITHUB_TOKEN` in a workflow](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#using-the-github_token-in-a-workflow)

### GitHub Actions에서 `PAT` 사용의 문제점

- 봇 계정이 가진 권한을 갖는, 모든 곳에 사용 가능한 `PAT`는 사용하지 않는 게 좋다. 왜? 관리하기 어렵다
    1. PAT 그 계정을 읽을 수 있는 사람이 퇴사할 때마다 계속 바꿔줘야 한다
    2. 어디서 사용하는지 모두 추적하기 어렵고, 따라서 지웠을 때 어떤 영향이 있을지 몰라서 로테이션을 못하게 된다
    3. 어디서 사용하고 있는지 기록한다 하더라도, 신뢰하기 어렵다
- 즉, `PAT` 본질적으로 관리하지 힘들거나 관리하지 못하므로, 사용을 자제하자
- 가령 AWS의 IAM 시크릿도 같은 문제, 트래킹 못하고, 로테이션 못하고

### GitHub Actions에서 `PAT`의 대안

- `PAT`는 다른 것들로 다 대체할 수 있다. `PAT`를 사용하는 대신 [자기 자신을 읽는 토큰](https://docs.github.com/en/actions/security-guides/automatic-token-authentication#permissions-for-the-github_token)이 있다.
    - `github.token`
    - `secrets.GITHUB_TOKEN`

#### `github.token`

```yaml
- name: Generate Mocks
  run: |
    git config --global url.https://${{ github.actor }}:${{ github.token }}@github.com/.insteadOf https://github.com/
    go env -w GOPRIVATE="github.com/some-qwerty-org.io/*"
```

- `github.token`은 GitHub Actions가 자동으로 생성하는 토큰
- 이 토큰은 각각의 워크플로우 실행에 대해 고유하며, 워크플로우가 완료되면 만료된다
- `github.token`은 github 컨텍스트의 일부로 사용되며, 이 컨텍스트는 Actions 워크플로우 내에서 실행 중인 작업과 관련된 데이터를 제공한다
- `github.token`는 기본적으로 실행 중인 워크플로우에 대한 리포지토리에 접근 권한을 갖는다
- [`permissions`](https://github.com/some-qwerty-org.io/actions/tree/main/openapi-to-ghp#usages) 추가해서 추가적인 권한을 명시적으로 추가할 수 있다
- 단 다른 리포에 대한 권한은 없다. 아무리 확장 시켜도 다른 리파지토리에는 확장 불가

#### `secrets.GITHUB_TOKEN`

- GitHub Actions 실행에 자동으로 생성되는 토큰
- 이 토큰은 `github.token`과 유사하게 각 실행에 대해 고유하게 생성되고, 워크플로우 완료 후 만료된다
- `secrets.GITHUB_TOKEN`은 GitHub Secrets를 통해 사용된다.
    - GitHub Secrets? 중요한 정보를 안전하게 저장하고, 워크플로우 내에서 이를 안전하게 사용할 수 있게 해주는 메커니즘
- `secrets.GITHUB_TOKEN`는 기본적으로 실행 중인 워크플로우에 대한 리포지토리에 접근 권한을 갖는다
- 단 다른 리포에 대한 권한은 없다. 아무리 확장 시켜도 다른 리파지토리에는 확장 불가

#### `github.token`와 `secrets.GITHUB_TOKEN` 차이

- 실질적인 기능적 차이는 없으며, 두 토큰 모두 같은 권한을 가지고 워크플로우 실행 시 자동으로 생성된다
- `github.token`은 github 컨텍스트의 일부로서 사용되고, `secrets.GITHUB_TOKEN`은 Secrets 메커니즘을 통해 사용된다는 차이가 있다.
    - 이러한 사용 방식의 차이는 주로 워크플로우 파일에서 **토큰을 참조하는 방법에 영향**을 미친다.
    - 예를 들어, `secrets.GITHUB_TOKEN`은 보안에 더 중점을 두는 설정에서 사용될 수 있다

## deploy key

- deploy key에 public key 등록하고, actions에 private key 등록
