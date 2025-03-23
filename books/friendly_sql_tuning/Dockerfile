FROM mysql:latest


COPY my.cnf /etc/mysql/conf.d/my.cnf

# 초기 실행할 SQL 파일을 복사합니다.
# `.sh`, `.sql`, `.sql.gz` 파일들이 자동으로 실행됩니다.
# - https://hub.docker.com/_/mysql#:~:text=Initializing%20a%20fresh%20instance
COPY init.sql /docker-entrypoint-initdb.d/

# 환경 변수 설정
ENV MYSQL_ROOT_PASSWORD=root
ENV MYSQL_DATABASE=friendly
ENV MYSQL_USER=rody
ENV MYSQL_PASSWORD=rody
ENV LANG=C.UTF-8

# MySQL 볼륨 위치를 명시합니다.
# 단, 데이터를 저장할 것이라고 도커에게 힌트를 주는 것이지,
# 호스트의 디렉토리와 매핑해주지는 않습니다.
VOLUME ["/var/lib/mysql", "/var/log/mysql"]
