#!/bin/bash

# run php-fpm7.4
php-fpm7.4

# 컨테이너는 메인 프로세스가 종료 시 종료된다.
# ㄴ https://www.tutorialworks.com/why-containers-stop/
# nginx 서버를 foreground로 돌리지 않으면 컨테이너 안의 서버가 실행이 안 된 상태이므로 컨테이너는 exit
# ㄴ http://nginx.org/en/docs/ngx_core_module.html#daemon
# ㄴ https://github.com/nginxinc/docker-nginx
# 일반적으로 프로덕션에서는 `daemon on;` 지시어를 사용해서 background에서 실행되도록 하지만,
# 컨테이너에서는 nginx만 사용할 경우 명령어 실행이 끝나면 컨테이너도 종료되며 detach 모드와 연결될 수 없으므로,
# `daemon off;` 지시어 통해서 foreground에서 실행되도록 한다
# ㄴ https://stackoverflow.com/a/34821579
# ㄴ https://stackoverflow.com/a/28099946
nginx -g 'daemon off;'