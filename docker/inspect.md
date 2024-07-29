# inspect

- [inspect](#inspect)
    - [상태 체크](#상태-체크)

## 상태 체크

```shell
# Waiting for all services to be ready
/usr/bin/docker inspect --format="{{if .Config.Healthcheck}}{{print .State.Health.Status}}{{end}}"

```
