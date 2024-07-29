# Tar

## 압축하기

- [날짜](https://www.gnu.org/software/coreutils/manual/html_node/date-invocation.html#date-invocation)로 파일 만들기

```shell
echo "grpc.log.tar.gz.$(TZ='Asia/Seoul' date '+%Y%m%d_%H%M')"
```

[압축하기](https://stackoverflow.com/a/18498409/8562273)

```shell
tar -zcvf "grpc.log.tar.gz.$(TZ='Asia/Seoul' date '+%Y%m%d_%H%M')" grpc.log
```

```bash
/usr/bin/tar --posix -z -cf cache.tgz -P -C /home/runner/work/go/go --files-from manifest.txt
```
