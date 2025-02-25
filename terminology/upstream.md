# upstream

- [upstream](#upstream)
    - [`upstream`이란?](#upstream이란)
    - [사용례](#사용례)
        - [Git에서의 사용](#git에서의-사용)
        - [오픈 소스 프로젝트](#오픈-소스-프로젝트)
        - [네트워킹 및 데이터 흐름](#네트워킹-및-데이터-흐름)
        - [웹 서버 및 프록시 서버 (예: Nginx)](#웹-서버-및-프록시-서버-예-nginx)

## `upstream`이란?

- `upstream`이라는 용어는 원래 하천에서 유래한 것으로, **강의 상류**를 의미
- 개발에서는 이를 비유적으로 사용하여, 데이터 흐름, 정보, 코드 등이 **'출처' 또는 '원점'으로부터 흘러오는 방향**을 가리키는 데 사용된다

## 사용례

### Git에서의 사용

- Git에서 `upstream`은 주로 로컬 브랜치와 연관된 **원격 브랜치**(remote)를 의미한다
- `upstream` 브랜치 설정은 로컬 브랜치가 추적하는 원격 저장소의 브랜치를 지정하는 것을 의미한다. 이 원격 저장소의 브랜치는 변경사항을 푸시(push)하거나 풀(pull)할 때 기준점이 되는 원격 브랜치가 된다

```shell
# `feature-branch`라는 로컬 브랜치가 `origin` 원격 저장소의 `feature-branch` 브랜치를 추적하도록 설정
git push --set-upstream origin feature-branch
```

### 오픈 소스 프로젝트

- 오픈 소스 소프트웨어 프로젝트에서의 기여 과정에서도 사용된다
- `upstream`은 프로젝트의 주 저장소, 즉 메인 리포지토리를 가리킨다. 즉, 개인 개발자가 작업한 코드를 이 주 저장소에 기여하는 방향을 의미한다
- 가령, 오픈 소스 프로젝트에 기여할 때, 개인 포크(fork)에서 메인 리포지토리로 풀 리퀘스트(pull request)를 보내는 과정이 `upstream`으로 기여하는 것이라고 한다

### 네트워킹 및 데이터 흐름

- 네트워킹, 특히 데이터 전송 및 통신 프로토콜에서 사용된다. "upstream"과 "downstream"은 데이터의 흐름 방향을 나타낸다.
- 데이터가 사용자나 클라이언트로부터 중앙 서버나 메인 네트워크 노드로 향하는 방향을 `upstream`이라고 한다. 반대 방향은 `downstream`이라고 한다
    - `Upstream`:
        - 사용자의 컴퓨터나 기기에서 ISP의 데이터 센터로 향하는 데이터 전송을 의미한다
        - 예를 들어, 이메일 보내기, 파일 업로드, 웹사이트에 데이터 전송 등이 여기에 해당한다
    - `Downstream`
        - ISP의 데이터 센터에서 사용자의 컴퓨터나 기기로 향하는 데이터 전송을 의미한다
        - 예를 들어, 웹 페이지 로딩, 영상 스트리밍, 파일 다운로드 등이 여기에 해당한다

### 웹 서버 및 프록시 서버 (예: Nginx)

- **웹 서버 구성** 및 **리버스 프록시 서버 설정**에서 사용된다
- `upstream`은 **프록시 서버 뒤에 위치하는 하나 이상의 실제 웹 서버**를 의미한다.
    - 프록시 서버 뒤에 위치하는 실제 서버? 클라이언트의 요청을 받아 처리하는 서버로, 프록시 서버는 이 `upstream` 서버로 요청을 전달한다
    - 그리고 그 결과를 클라이언트에게 다시 보낸다
- 이를 통해 로드 밸런싱, 고가용성, 캐싱 등의 기능을 구현할 수 있다
    - 가령, Nginx에서 `upstream` 모듈을 사용하여 백엔드 서버의 그룹을 구성하고, 로드 밸런싱을 설정한다

```nginx
# Nginx 설정 예시
http {
    # `myapp1`이라는 이름의 `upstream`을 정의
    # 이 `upstream`에는 세 개의 서버가 포함된다
    upstream myapp1 {
        server srv1.example.com;
        server srv2.example.com;
        server srv3.example.com;
    }

    server {
        # 모든 HTTP 요청은 이 `upstream` 그룹의 서버 중 하나로 전달된다
        # Nginx는 위의 세 개의 서버들 사이에서 로드 밸런싱을 수행한다
        location / {
            proxy_pass http://myapp1;
        }
    }
}
```
