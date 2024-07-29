# nginx

- [nginx](#nginx)
    - [apache to nginx(feat. PHP)](#apache-to-nginxfeat-php)
    - [accesslog format](#accesslog-format)

## apache to nginx(feat. PHP)

1. chrony

    ```shell
    sudo apt install chrony
    ```

    ```conf
    # /etc/chrony/chrony.conf
    server 169.254.169.123 prefer iburst minpoll 4 maxpoll 4
    ```

2. Timezone KST설정

    ```shell
    sudo timedatectl set-timezone Asia/Seoul
    ```

3. locale 설정

    ```shell
    sudo locale-gen ko_KR.UTF-8
    # ssh 재접속해야 반영되므로 source /etc/default/locale 하면 바로 반영됨
    sudo update-locale LANG=ko_KR.UTF-8 LC_MESSAGES=POSIX
    ```

4. EUC-KR locale 설치

    ```shell
    sudo locale-gen ko_KR.EUC-KR
    # sudo apt-get install language-pack-ko -y ?
    ```

5. PHP 5.6

    ```shell
    sudo add-apt-repository ppa:ondrej/php
    sudo apt-get install -y \
        php5.6 \
        php5.6-bcmath \
        php5.6-curl \
        php5.6-dev \
        php5.6-fpm \
        php5.6-gd \
        php5.6-intl
        php5.6-json \
        php5.6-mbstring \
        php5.6-mcrypt \
        php5.6-mysql \
        php5.6-opcache \
        php5.6-soap \
        php5.6-xml \
        php5.6-zip \
    ```

6. Nginx 설치

    ```shell
    sudo apt install nginx nginx-extras -y
    ```

7. Nginx 설정

    ```conf
    worker_rlimit_nofile 65536

    events {
        worker_connections 4096;
        # multi_accept on
    }

    server_tokens off;

    log_format proxied_combined '$http_x_forwarded_for - $remote_user [$time_iso8601 $msec] #$connection_requests "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_time $upstream_response_time $pipe';

    /etc/nginx/site-enables/default

    root /var/www/api/app/webroot;
    index index.php;

    server_name service.some_name.kr;

    access_log /var/log/nginx/api_access.log proxied_combined;
    error_log /var/log/nginx/api_error.log;
    ```

8. PHP-FPM

    ```conf
    # /etc/php/5.6/fpm/pool.d/www.conf
    listen = /run/php/php5.6-fpm.sock
    # FPM 프로세스가 open할 수 있는 file descriptor 개수를 기본값인 1024에서 65536으로 증가
    rlimit_files = 65536

    pm = static
    pm.max_children = 400

    # FPM status 모니터링을 위해 아래의 설정을 추가 (실제 application route 와 겹치지 않을 고유한 URL을 등록)
    pm.status_path = /nginx-fpm-status

    # 원하는 경우, ping / pong 설정을 하면 health check 용으로 간단히 사용할 수 있음
    # ping.path
    # ping.reponse

    # logging & rotate
    prefix = /etc/php/5.6/fpm/pool.d
    access.log = $pool.access.log
    access.format = "%R - %u %t \"%m %r%Q%q\" %s %f %{mili}d %{kilo}M %C%%"
    slowlog = $pool.log.slow
    request_slowlog_timeout = 3

    # ETC
    request_terminate_timeout = 60s
    php_admin_value[error_log] = /var/log/fpm-php.www.log
    php_admin_flag[log_errors] = on
    ```

9. log lotate

    ```conf
    /var/log/php5.6-fpm.log
    /etc/php/5.6/fpm/pool.d/www.access.log
    /etc/php/5.6/fpm/pool.d/www.log.slow
    {
            rotate 12
            weekly
            missingok
            notifempty
            compress
            delaycompress
            postrotate
                    if [ -x /usr/lib/php/php5.6-fpm-reopenlogs ]; then
                            /usr/lib/php/php5.6-fpm-reopenlogs;
                    fi
            endscript
    }
    ```

## accesslog format

```conf
$http_x_forwarded_for - $remote_user [$time_iso8601 $msec] #$connection_requests "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" $request_time $upstream_response_time $pipe "$request_body" "$http_x_amzn_trace_id"
```

- `$http_x_forwarded_for`
    - HTTP 헤더 X-Forwarded-For의 값
    - 클라이언트의 원래 IP 주소를 식별하는 데 사용
    - 로드 밸런서나 프록시 서버를 통해 요청이 전달될 때 이 헤더는 원래 클라이언트의 IP 주소를 유지
- `-`
    - 이 하이픈은 로그 항목의 구분자 역할
    - 로그를 파싱할 때 이 구분자를 사용하여 로그 항목을 분리할 수 있다
    - 또한, 일부 변수 값이 없거나 빈 경우에 이 하이픈이 사용되어 값의 부재를 나타내기도 합니다.
- `$remote_user`
    - 클라이언트에 의해 제공된 사용자 이름. 이 사용자 이름은 HTTP 기본 인증을 통해 제공될 수 있다
        - HTTP 기본 인증? 사용자 이름과 비밀번호를 요구하는 간단한 인증 방법
    - 클라이언트가 이러한 인증 정보를 제공하면, 해당 정보는 Nginx 로그의 `$remote_user` 변수에 저장된다.
- `$time_iso8601`
    - 요청이 발생한 시점의 시간을 ISO 8601 표준 형식으로 나타낸다. (예: 2023-10-17T11:35:10+09:00)
    - 로그에 기록된 각 요청의 타임스탬프로 사용되어, 언제 어떤 요청이 발생했는지를 추적할 수 있도록 돕는다
- `$msec`
    - 요청이 발생한 시점의 시간을 밀리초(1/1000초) 단위로 나타낸다.
    - 이 변수는 더 높은 시간 해상도가 필요할 때 사용되며, 특히 빠른 응답 시간을 가진 시스템에서 미세한 성능 측정에 유용하다
- `#$connection_requests`
    - 현재 연결에서 처리된 요청의 수
    - Nginx는 클라이언트와의 연결을 유지하며, 이 변수를 사용하여 한 연결에서 처리된 요청의 수를 추적할 수 있다
- `"$request"`
    - 클라이언트의 요청 라인을 포함
    - 이는 HTTP 메서드 (예: GET, POST), 요청 URI, 그리고 HTTP 프로토콜 버전을 포함
- `$status`
    - HTTP 응답 상태 코드
    - 요청의 성공, 리다이렉션, 클라이언트 오류, 서버 오류 등 요청의 결과를 나타낸다
- `$body_bytes_sent`
    - 응답 본문의 바이트 크기
    - 네트워크 트래픽의 크기와 응답 시간을 분석하는 데 유용할 수 있음
- `"$http_referer"`
    - `Referer` 헤더는 클라이언트가 현재 페이지에 어떻게 도달했는지를 나타내는 URL을 포함한다
    - 이 정보는 웹 사이트의 트래픽 출처를 분석하고, 외부 링크의 효과를 평가하는 데 유용할 수 있다
- `"$http_user_agent"`
    - `User-Agent` 헤더는 클라이언트의 브라우저와 운영 체제에 대한 정보를 제공한다
    - 이 정보는 브라우저 호환성 문제를 진단하고, 웹 사이트의 사용자 인터페이스를 최적화하는 데 사용될 수 있다
- `$request_time`
    - 클라이언트의 요청을 받아 처리하고, 응답을 반환하는 데 걸린 전체 시간을 초 단위로 나타낸다
    - 이 변수는 요청 처리의 성능을 평가하거나, 특정 요청에 대한 서버의 반응 시간을 모니터링하는데 사용된다
- `$upstream_response_time`
    - 이 변수는 Nginx가 업스트림 서버로부터 응답을 받는 데 걸린 시간(초 단위)
    - 업스트림 서버는 Nginx 뒤에 있는 실제 웹 서버나 애플리케이션 서버를 의미하며, 이 변수는 업스트림 서버의 성능을 모니터링하는 데 사용
- `$pipe`
    - 이 변수는 요청이 `keep-alive` 연결을 통해 전송되었는지(., 즉 연결이 종료되지 않음) 또는 파이프 연결을 통해 전송되었는지(p, 즉 파이프 연결)를 나타낸다.
    - 이 정보는 네트워크 트래픽과 연결 관리를 분석하는 데 도움이 된다
- `"$request_body"`:
    - 클라이언트의 요청 본문
    - 이 변수는 POST나 PUT과 같은 HTTP 메서드에서 사용되며, 클라이언트가 서버에 전송하는 데이터를 포함한다
- [`$http_x_amzn_trace_id`](https://docs.aws.amazon.com/ko_kr/elasticloadbalancing/latest/application/load-balancer-request-tracing.html)

    ```text
    Field=version-time-id
    ex: X-Amzn-Trace-Id: Root=1-63441c4a-abcdef012345678912345678
    ```

    - `Field`: 지원되는 값은 `Root` 및 `Self`
    - `version`: 버전 번호입
    - `time`: epoch 시간(초)
    - `id`: 트레이스 식별자
