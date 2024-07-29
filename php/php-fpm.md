# fpm

- [fpm](#fpm)
    - [FPM과 프로세스 라이프 사이클](#fpm과-프로세스-라이프-사이클)
        - [PHP FPM과 스크립트 컴파일](#php-fpm과-스크립트-컴파일)
        - [Static 변수와 메서드의 관리](#static-변수와-메서드의-관리)
        - [Static 변수의 스택 관리](#static-변수의-스택-관리)
        - [프로세스 및 스레드 실행과 파괴](#프로세스-및-스레드-실행과-파괴)
    - [PHP와 PHP-FPM](#php와-php-fpm)
    - [FPM conf](#fpm-conf)
        - [nginx](#nginx)
            - [/etc/nginx/sites-enabled/default](#etcnginxsites-enableddefault)
        - [PHP](#php)
            - [fpm](#fpm-1)
                - [/etc/php/5.6/fpm/pool.d/www.conf](#etcphp56fpmpooldwwwconf)
    - [기타](#기타)
        - [소켓 직접 생성](#소켓-직접-생성)
    - [php-fpm-status 조회](#php-fpm-status-조회)

## FPM과 프로세스 라이프 사이클

- PHP-FPM (FastCGI Process Manager) 환경에서는 각 요청이 별도의 프로세스 또는 스레드에서 처리된다
- 스크립트의 실행은 요청마다 시작되고 종료되므로, `static` 변수는 **각 요청의 생명 주기 동안만 존재**한다.
- 가령, FPM이 100번의 요청마다 리로드되도록 설정되어 있는 경우
    - 각 요청은 독립적으로 처리된다
    - `static $client`로 선언돼서 한번 초기화되는 Guzzle 클라이언트는 해당 요청이 끝나면 사라진다.
    - 따라서 요청 간에 클라이언트가 지속되지는 않는다

### PHP FPM과 스크립트 컴파일

- PHP FPM은 FastCGI 프로세스 매니저로, PHP 스크립트를 처리하는 독립된 프로세스 풀을 관리한다
- PHP 스크립트가 실행될 때, FPM은 스크립트를 컴파일하여 opcode(기계가 실행할 수 있는 중간 코드)로 변환하고, 이 opcode는 FPM의 프로세스 내에서 실행된다.
- 스크립트가 변경되지 않는 경우, opcode는 OPCache를 통해 저장되어 재사용될 수 있다

### Static 변수와 메서드의 관리

- PHP에서 static 변수나 메서드는 클래스 수준에서 관리된다
- 이는 개별 객체 인스턴스가 아닌 클래스 자체와 연결되어 있어, 클래스의 모든 인스턴스가 해당 데이터를 공유한다.
- PHP FPM 환경에서, 각 요청은 별도의 프로세스 또는 스레드에서 처리되므로, static 변수는 해당 요청이 처리되는 동안에만 유지된다
- 요청이 완료되면, 해당 프로세스 또는 스레드는 종료되고, static 변수도 함께 초기화된다

### Static 변수의 스택 관리

- PHP에서 static 변수는 스택에 저장되지 않고 힙 메모리에 저장되며, 이는 static 변수가 요청 간에 유지되지 않는 이유 중 하나다
- 요청 처리 동안에만 존재하며, 요청이 끝나면 함께 정리된다

### 프로세스 및 스레드 실행과 파괴

- PHP FPM은 프로세스 기반이다. 각 요청은 별도의 프로세스에서 처리되며, 요청 처리가 완료되면 이 프로세스는 종료된다
- 스레드 대신 프로세스를 사용함으로써, 요청 간의 메모리 격리가 보장되고, 한 요청에서 발생한 문제가 다른 요청에 영향을 미치지 않는다
- 프로세스가 종료될 때, PHP는 할당된 모든 리소스와 메모리를 해제하고 정리합니다.

## PHP와 PHP-FPM

1. PHP와 PHP-FPM의 차이:

    PHP (PHP: Hypertext Preprocessor):
    - PHP는 서버 사이드 스크립팅 언어입니다.
    - CLI(Command Line Interface) 모드로 실행될 때 사용됩니다.
    - 웹 서버와 직접 통신하지 않습니다.

    PHP-FPM (PHP FastCGI Process Manager):
    - PHP 스크립트를 실행하기 위한 FastCGI 구현체입니다.
    - 웹 서버(예: Nginx)와 PHP 간의 인터페이스 역할을 합니다.
    - 여러 PHP 프로세스를 관리하여 성능을 향상시킵니다.
    - 웹 환경에서 PHP 실행에 최적화되어 있습니다.

2. PHP-FPM의 역할:

    1. 프로세스 관리:
        - PHP 작업자 프로세스를 생성, 관리, 종료합니다.
        - PHP 요청을 처리할 worker 프로세스 풀을 생성하고 관리합니다.
        - 요청에 따라 프로세스를 동적으로 확장하거나 축소할 수 있습니다.

    2. 성능 최적화:
        - 요청에 따라 프로세스를 동적으로 조절합니다.
        - 프로세스 재사용을 통해 새로운 PHP 인터프리터 시작 오버헤드를 줄입니다.
        - 동시 요청 처리 능력을 향상시킵니다.

    3. 웹 서버 연동:
        - 웹 서버로부터 PHP 실행 요청을 받아 처리합니다.
        - FastCGI 프로토콜을 통해 웹 서버와 효율적으로 통신합니다.
        - PHP 스크립트 실행 결과를 웹 서버에 전달합니다.

    4. 리소스 관리:
        - 메모리 사용량, 요청 제한, 요청 처리 시간 등을 모니터링하고 제어합니다.
        - 과도한 리소스 사용을 방지하여 시스템 안정성을 유지합니다.

    5. 로깅:
        - PHP 스크립트 실행과 관련된 로그를 생성합니다.

3. macOS에서 PHP-FPM 상태 확인 및 제어:

    macOS에서는 `systemctl` 대신 `brew services` 명령을 사용할 수 있습니다:

    - 서비스 등록 상태 확인:

        ```bash
        brew services list | grep php

        # php@8.2          started rody /Users/rody/Library/LaunchAgents/homebrew.mxcl.php@8.2.plist
        ```

        - `brew services list`는 Homebrew로 설치된 서비스들의 목록과 그 상태를 보여줍니다.
        - PHP-FPM의 서비스 등록 상태를 확인할 수 있지만, 실제 실행 중인 프로세스를 직접 보여주지는 않습니다.

    - 실행 중인 프로세스 확인:

        ```bash
        ps aux | grep php-fpm
        ```

        - 시스템에서 실제로 실행 중인 PHP-FPM 프로세스를 직접 보여줍니다.
        - 현재 활성화된 PHP-FPM 프로세스의 상세 정보(PID, CPU/메모리 사용량 등)를 제공합니다.
    - PHP-FPM 마스터 프로세스 PID 확인:

        ```bash
        pgrep -f "php-fpm: master process"
        ```

    - PHP-FPM 설정 파일 위치 및 문법 검사:

        ```bash
        php-fpm -t
        ```

    - 시작:

        ```bash
        brew services start php@8.2
        ```

    - 중지:

        ```bash
        brew services stop php@8.2
        ```

    - 재시작:

        ```bash
        brew services restart php@8.2
        ```

    - 수동으로 시작/중지하려면:

        - 시작:

            ```bash
            php-fpm --daemonize
            ```

        - 중지 (PID를 찾아 종료):

            ```bash
            pkill -f php-fpm
            ```

        - 또는 PID 파일을 사용하여:

            ```bash
            kill $(cat /usr/local/var/run/php-fpm.pid)
            ```

    - 프로세스 확인:

        ```bash
        ps aux | grep php-fpm
        ```

    - PHP-FPM 상태 페이지 확인 (설정되어 있는 경우):

        ```bash
        curl http://localhost/php-fpm-status
        ```

    - 로그 확인 (기본 위치):

        ```bash
        tail -f /usr/local/var/log/php-fpm.log
        ```

- PHP-FPM은 마스터 프로세스와 여러 워커 프로세스로 구성됩니다. 귀하의 `ps` 출력에서 볼 수 있듯이, 여러 "pool www" 프로세스가 실행 중입니다.
- 마스터 프로세스는 워커 프로세스들을 관리하고, 각 워커 프로세스는 실제 PHP 요청을 처리합니다.
- PHP-FPM의 설정에 따라 워커 프로세스의 수와 동작 방식이 결정됩니다.
- 정확한 경로는 PHP 버전과 설치 방법에 따라 다를 수 있습니다.
Homebrew로 설치한 경우, 설정 파일은 보통 `/opt/homebrew/etc/php/8.2/php-fpm.d/` 디렉토리에 있습니다.

## FPM conf

### nginx

#### /etc/nginx/sites-enabled/default

- `fastcgi_pass` 사용하여 어떤 소켓과 통신할 것인지 설정

```nginx
location ~ \.php$ {
    fastcgi_buffer_size       64k;
    fastcgi_buffers           4 128k;
    fastcgi_busy_buffers_size 128k;
    fastcgi_split_path_info ^(.+\.php)(/.+)$;

    # NOTE: You should have "cgi.fix_pathinfo = 0;" in php.ini
    fastcgi_pass unix:/run/php/php5.6-fpm.sock;
    fastcgi_index index.php;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
}
```

```shell
$ sudo nginx -t
nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
nginx: configuration file /etc/nginx/nginx.conf test is successful
$ sudo nginx -s reload
```

### PHP

#### fpm

##### /etc/php/5.6/fpm/pool.d/www.conf

- 위의 nginx에서 설정한 `fastcgi_pass`의 소켓과 `listen`을 맞춰준다

```ini
[www]
prefix = /etc/php/5.6/fpm/pool.d
user = www-data
group = www-data

listen = /run/php/php5.6-fpm.sock
```

```shell
service php5.6-fpm reload
service php5.6-fpm status
```

## 기타

### [소켓 직접 생성](https://stackoverflow.com/a/43762862)

```shell
python -c "import socket as s; sock = s.socket(s.AF_UNIX); sock.bind('/run/php/php-fpm.sock')"
```

## php-fpm-status 조회

```shell
sudo -u www-data \
  SCRIPT_NAME=/php-fpm-status \
  SCRIPT_FILENAME=/php-fpm-status \
  QUERY_STRING='json&full' \
  REQUEST_METHOD=GET \
  cgi-fcgi -bind -connect /run/php/php5.6-fpm.sock | tail -n1 |
  jq '.processes[] | select(.state != "Idle")'
```

```bash
sudo -u www-data \
    SCRIPT_NAME=/php-fpm-status \
    SCRIPT_FILENAME=/php-fpm-status \
    QUERY_STRING='json&full' \
    REQUEST_METHOD=GET \
    cgi-fcgi \
    -bind \
    -connect /run/php/php5.6-fpm.sock | tail -n1 | jq '.processes[] | {pid, state}'
```

```bash
# `https://` 없이 테스트
curl -v -X POST sub_b\.domain.kr/partners/receipts/imp_756741676240 \
    --resolve sub_b\.domain.kr:80:10.0.5.238 \
    --header "Content-Type: application/json" \
    -d '{"apiKey":"apiKeyValue","apiSecret":"apiSecretValue"}'
```
