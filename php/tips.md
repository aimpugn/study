# Tips

- [Tips](#tips)
    - [grpc 모듈 버전 업그레이드 절차](#grpc-모듈-버전-업그레이드-절차)

## grpc 모듈 버전 업그레이드 절차

1. LB에서 제거
2. php5.6-cli 버전체크

    ```shell
    php -r "echo phpversion('grpc') . PHP_EOL;"
    ```

3. php5.6-fpm 버전체크

    ```shell
    sudo tee /var/www/service/app/webroot/index2.php <<< "<?php phpinfo();" && curl -s localhost/index2.php | grep -i grpc && sudo rm /var/www/service/app/webroot/index2.php
    ```

4. php5.6-fpm 종료

    ```shell
    sudo systemctl stop php5.6-fpm && systemctl status php5.6-fpm
    ```

5. 버전 업그레이드

    ```shell
    sudo apt-get install autoconf zlib1g-dev php5.6-dev php-pear
    sudo pecl install grpc-1.32.0
    ```

6. php-fpm 리로드

    ```shell
    sudo systemctl start php5.6-fpm && systemctl status php5.6-fpm
    ```

7. 버전 확인

    ```shell
    php -r "echo phpversion('grpc') . PHP_EOL;"
    sudo tee /var/www/service/app/webroot/index2.php <<< "<?php phpinfo();" && curl -s localhost/index2.php | grep -i grpc && sudo rm /var/www/service/app/webroot/index2.php
    ```

8. 정상 기능 확인 `curl http://localhost/users/pg/imp67011510 | sha1sum`
9. LB에 붙이기 (편집됨)
