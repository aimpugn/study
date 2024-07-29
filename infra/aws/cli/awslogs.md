# [awslogs](https://github.com/jorgebastida/awslogs)

- [awslogs](#awslogs)
    - [install](#install)
    - [usage](#usage)
        - [get](#get)

## install

```shell
brew install awslogs
```

## usage

### get

```shell
awslogs get some-A-serivce/application_log \
    -G \
    -S \
    --color never \
    --filter-pattern '[client_ip, server_ip, login_name, remote_user, request_time, request ="*payments/status*"]' \
    -s "$(toutc "2023-05-23 00:00:00 +0900")" \
    -e "$(toutc "2023-08-02 23:59:59 +0900")" > ./application_log.log


awslogs get some-A-serivce/application_log \
    -G \
    -S \
    --color never \
    -s "$(toutc "2023-05-23 00:00:00 +0900")" \
    -e "$(toutc "2023-08-02 23:59:59 +0900")" > ./application_log.log

    Payment Prepare Request

awslogs get some-A-serivce/application_log \
    -G \
    -S \
    --color never \
    --filter-pattern 'Payment Prepare Request' \
    -s "$(toutc "2023-05-23 00:00:00 +0900")" \
    -e "$(toutc "2023-08-02 23:59:59 +0900")" > ./application_log.log
```

- `-G`, `--no-group`
    - Do not display group nam
- `-S`, `--no-stream`
    - Do not display stream name
- `--color`
    - When to color output. WHEN can be `auto` (default if omitted), `never`, or `always`.
    - With `--color=auto`, output is colored only when standard output is connected to a terminal.
- `-f`, `--filter-pattern`
    - A valid CloudWatch Logs filter pattern to use for filtering the response.
    - If not provided, all the events are matched.
- `-s`, `--start`
    - Start time (default 5m)
- `-e`, `--end`
    - End time
