#!/bin/bash

NAME="friendly-mysql"

docker rm -f "$NAME"

docker build -t "$NAME" .

docker run -d \
  --name "$NAME" \
  -v "$(pwd)/tmp/data:/var/lib/mysql" \
  -v "$(pwd)/tmp/logs:/var/log/mysql" \
  -v "$(pwd)/tmp/run:/var/run/mysqld" \
  -p 3306:3306 \
  "$NAME"


MYSQLD_SOCK="/var/run/mysqld/mysqld.sock"
MYSQLD_SOCK_CHECK="while [ ! -S $MYSQLD_SOCK ]; do sleep 1; done; echo 'MySQL ready'"

docker exec "$NAME" /bin/bash -c "$MYSQLD_SOCK_CHECK"
