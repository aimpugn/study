# docker exec

## 컨테이너 접속하여 새로운 bash 셸 세션 시작

- 실행 중인 컨테이너 내부에서 추가적인 작업을 하고 싶을 때 사용

```bash
docker exec -it $CONTAINER_NAME /bin/bash
```

## docker exec 사용시 bashrc 적용하기

이 문제는 주로 다음과 같은 이유로 발생할 수 있습니다:

1. 환경 변수와 함수의 범위:
   - `source ~/.bashrc`를 실행해도 해당 세션에서 정의된 함수나 환경 변수가 `docker exec`의 새로운 shell 세션으로 전파되지 않을 수 있습니다.

2. 비-인터랙티브 셸:
   - `docker exec`로 실행된 명령은 비-인터랙티브 셸에서 실행되며, 이 경우 `.bashrc`가 로드되지 않을 수 있습니다.

3. `run` 명령어의 정의:
   - `run` 명령어가 함수나 alias로 정의되어 있다면, 이는 일반적으로 새로운 셸 세션에서 인식되지 않습니다.

해결 방법:

1. 전체 명령을 하나의 셸 세션에서 실행:

   ```bash
   docker exec container_name /bin/bash -c 'source ~/.bashrc && source /path/to/script && (cd /var/www/api && run test all)'
   ```

2. 환경 파일을 명시적으로 소스:

   ```bash
   docker exec container_name /bin/bash -c 'source ~/.bashrc && source /path/to/script && source /etc/profile && (cd /var/www/api && run test all)'
   ```

3. 로그인 셸 사용:

   ```bash
   docker exec -it container_name /bin/bash -lc 'source ~/.bashrc && source /path/to/script && (cd /var/www/api && run test all)'
   ```

4. `run` 함수를 직접 정의:

   ```bash
   docker exec container_name /bin/bash -c 'source ~/.bashrc && source /path/to/script && function run() { ... } && (cd /var/www/api && run test all)'
   ```

5. 전체 경로 사용:
   `run` 명령어가 실제로 실행 파일이라면, 전체 경로를 사용해보세요.

   ```bash
   docker exec container_name /bin/bash -c 'source ~/.bashrc && (cd /var/www/api && /full/path/to/run test all)'
   ```

6. 스크립트 파일 사용:
   모든 필요한 초기화와 명령을 포함하는 별도의 스크립트 파일을 만들고, 이를 실행하는 방법도 있습니다.

   ```bash
   docker exec container_name /bin/bash /path/to/your/script.sh
   ```

   여기서 `script.sh`는 다음과 같이 구성될 수 있습니다:

   ```bash
   #!/bin/bash
   source ~/.bashrc
   source /path/to/script
   cd /var/www/api
   run test all
   ```

이러한 방법들을 시도해보시고, 여전히 문제가 있다면 컨테이너 내부의 환경 설정, `run` 명령어의 정확한 정의, 그리고 관련 스크립트의 내용을 더 자세히 확인해볼 필요가 있습니다.
