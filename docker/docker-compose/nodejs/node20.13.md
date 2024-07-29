# node:20.13

```yaml
version: '3.8'

services:
  app:
    image: node:20.13.1-alpine3.20
    # Docker Compose를 사용하여 볼륨을 마운트한 상태라면, 
    # 일반적으로 Dockerfile에서 `COPY` 명령을 사용할 필요가 없습니다.
    # 
    # 대신, Docker Compose 파일에서 볼륨을 정의하고, 애플리케이션 코드를 컨테이너에 직접 마운트하여
    # 호스트 시스템의 디렉토리를 컨테이너 내의 디렉토리로 연결하여 개발 환경에서 변경 사항을 즉시 반영할 수 있습니다.
    volumes:
      - ./:/usr/src/app
      # `/usr/src/app/node_modules`를 명시적으로 마운트한 이유는, 
      # 호스트 시스템의 `node_modules`를 컨테이너와 분리하여 관리하기 위함입니다.
      # 이렇게 하면 호스트 시스템의 `node_modules`와 충돌 없이 컨테이너 내에서 의존성을 설치할 수 있습니다.
      - /usr/src/app/node_modules
    working_dir: /usr/src/app
    command: tail -f /dev/null
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
    depends_on:
      - db

  db:
    image: postgres:latest
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mydb

volumes:
  pgdata:
```
