#!/bin/bash

set -eo pipefail

docker-compose down -v
docker-compose up -d
