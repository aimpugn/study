#!/bin/bash

set -e

NAME=ubuntu-env

# to stop and delete running container
echo "Remove container: $NAME"
docker rm -f "$NAME"

# build
echo "Build container: $NAME"
docker build -t "$NAME" .

# ex: docker run --name asm-linux-env -it asm-linux-env
RUN_OPTIONS=(
    --volume "$(pwd)/tmp:/home/$NAME" # mount volume
    --rm
    -it
    "$NAME" # tag name of image to run
)
DOCKER_RUN_CMD="docker run ${RUN_OPTIONS[*]}"
echo "Execute '$DOCKER_RUN_CMD'"
sh -c "$DOCKER_RUN_CMD"


