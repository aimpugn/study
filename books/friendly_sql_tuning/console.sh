#!/bin/sh

NAME="friendly-mysql"

docker exec -it "$NAME" \
    mysql  --database friendly -urody -prody
