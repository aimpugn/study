FROM ubuntu:20.04

###################################################################
############################ VARIABLES ############################
###################################################################

ARG PHP_VERSION=7.4
ARG NGINX_RECV_KEYS=ABF5BD827BD9BF62
ARG NGINX_RELEASE=focal
ARG SERVICE_DIRECTORY=/var/www/example.com
ARG PHP_SOCKET_DIRECTORY=/var/run/php

###############################################################
############################ NGINX ############################
###############################################################
# WARNING: apt does not have a stable CLI interface. Use with caution in scripts.
# apt 명령어 경우 사용자와의 interaction에 중점을 두었기 때문에 script 내에서 사용하기 부적절(https://technote.kr/372)
RUN apt-get update && apt-get -y upgrade
RUN apt-get -y install  \
    ca-certificates \
    curl  \
    git  \
    zip  \
    wget \
    unzip  \
    vim  \
    locales  \
    zsh  \
    ntp \
    language-pack-ko  \
    dnsutils \
    software-properties-common  \
    bash-completion

# https://www.nginx.com/resources/wiki/start/topics/tutorials/install/#official-debian-ubuntu-packages
# 여러 줄 cat: https://stackoverflow.com/a/68349357
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys ${NGINX_RECV_KEYS}
RUN echo  \
    "deb https://nginx.org/packages/ubuntu/ ${NGINX_RELEASE} nginx\n" \
    "deb-src https://nginx.org/packages/ubuntu/ ${NGINX_RELEASE} nginx\n" > /etc/apt/sources.list.d/nginx.list

RUN apt-get update && apt-get -y upgrade

# nginx-full? https://askubuntu.com/a/556382
RUN apt-get -y install nginx

# systemd는 컨테이너 시작 시 실행되지 않는다: https://stackoverflow.com/a/62489079
# 컨테이너에서 systemd 사용 불가: https://stackoverflow.com/a/51985145
# RUN systemctl start nginx

# Nginx 구성
ADD nginx/conf.d/ /etc/nginx/conf.d/
ADD nginx/sites-enabled/ /etc/nginx/sites-enabled/
COPY nginx/nginx.conf /etc/nginx/nginx.conf
# fast cgi에서 사용하는 파라미터
COPY nginx/fastcgi_params /etc/nginx/fastcgi_params

RUN mkdir -p ${SERVICE_DIRECTORY}

#############################################################
############################ PHP ############################
#############################################################
ARG PHP_VERSION=7.4
RUN apt-get install -y  \
    php${PHP_VERSION}  \
    php${PHP_VERSION}-bcmath  \
    # https://www.tutorialspoint.com/php/php_bzip2_functions.htm
    php${PHP_VERSION}-bz2  \
    php${PHP_VERSION}-common  \
    php${PHP_VERSION}-cli  \
    php${PHP_VERSION}-curl  \
    # Files for PHP7.4 module development
    php${PHP_VERSION}-dev  \
    php${PHP_VERSION}-fpm  \
    php${PHP_VERSION}-gd  \
    php${PHP_VERSION}-intl  \
    php${PHP_VERSION}-json  \
    php${PHP_VERSION}-mbstring  \
    php${PHP_VERSION}-mysql  \
    php${PHP_VERSION}-opcache \
    # PostgreSQL module for PHP
    php${PHP_VERSION}-pgsql \
    # server-side, HTML-embedded scripting language (PHPDBG binary)
    # ㄴ https://www.php.net/manual/en/intro.phpdbg.php  \
    php${PHP_VERSION}-phpdbg \
    php${PHP_VERSION}-pspell \
    php${PHP_VERSION}-readline \
    php${PHP_VERSION}-soap \
    php${PHP_VERSION}-sqlite3 \
    php${PHP_VERSION}-xdebug \
    php${PHP_VERSION}-xml  \
    # https://www.php.net/manual/en/intro.xmlrpc.php
    php${PHP_VERSION}-xmlrpc  \
    # https://www.php.net/manual/en/intro.xsl.php
    php${PHP_VERSION}-xsl \
    php${PHP_VERSION}-zip

# Socket directory for php-fpm, www.conf
RUN mkdir -p ${PHP_SOCKET_DIRECTORY}

# PHP CLI
COPY cli/php.ini /etc/php/${PHP_VERSION}/cli/php.ini

# PHP FPM
COPY fpm/php.ini /etc/php/${PHP_VERSION}/fpm/php.ini
COPY fpm/php-fpm.conf /etc/php/${PHP_VERSION}/fpm/php-fpm.conf
COPY fpm/pool.d/www.conf /etc/php/${PHP_VERSION}/fpm/pool.d/www.conf

#############################################################
############################ ETC ############################
#############################################################
COPY ./.bashrc /root/.bashrc
COPY ./.inputrc /root/.inputrc


####################################################################
############################ ENTRYPOINT ############################
####################################################################
# https://docs.docker.com/engine/reference/builder/#understand-how-cmd-and-entrypoint-interact
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]


#############################################################
############################ CMD ############################
#############################################################