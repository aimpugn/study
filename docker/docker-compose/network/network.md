# network

- [network](#network)
    - [컨테이너가 있는 네트워크 찾기](#컨테이너가-있는-네트워크-찾기)

## [컨테이너가 있는 네트워크 찾기](https://stackoverflow.com/a/43904733)

```shell
docker inspect "$NAME" -f "{{json .NetworkSettings.Networks }}"

docker network inspect '<tab complete to show avail. bridges>' | grep IPv4Address
```
