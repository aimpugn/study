# docker status

- [docker status](#docker-status)
    - [Docker 컨테이너 상태 관리와 확인](#docker-컨테이너-상태-관리와-확인)
    - [Docker 컨테이너 다시 시작하기](#docker-컨테이너-다시-시작하기)
    - [일시 정지된 컨테이너를 재개하기](#일시-정지된-컨테이너를-재개하기)

## Docker 컨테이너 상태 관리와 확인

`docker ps` 명령어를 사용하여 현재 실행 중인 컨테이너를 확인할 수 있습니다:

```bash
docker ps
```

멈춰있는(stopped) 컨테이너를 포함하여 모든 컨테이너를 확인하려면 다음 명령어를 사용합니다:

```bash
docker ps -a
```

여기서, STATUS 열을 통해 컨테이너의 상태를 확인할 수 있습니다. 상태는 다음과 같이 나타날 수 있습니다:
- `Up` : 컨테이너가 실행 중
- `Exited` : 컨테이너가 종료됨
- `Paused` : 컨테이너가 일시 정지됨

## Docker 컨테이너 다시 시작하기

멈춰있거나 종료된 컨테이너를 다시 시작하는 명령어는 다음과 같습니다:

```bash
docker start <컨테이너 이름 또는 ID>
```

예를 들어, `mycontainer`라는 이름의 컨테이너를 다시 시작하려면 다음과 같이 입력합니다:

```bash
docker start mycontainer
```

## 일시 정지된 컨테이너를 재개하기

만약 컨테이너가 일시 정지 상태(`Paused`)라면 다음 명령어를 사용하여 재개할 수 있습니다:

```bash
docker unpause <컨테이너 이름 또는 ID>
```
