# Memory

- [Memory](#memory)
    - [메모리 사용 현황 보기](#메모리-사용-현황-보기)

## 메모리 사용 현황 보기

- `-s 0` 옵션을 주면 즉시 완료

```shell
top -l 1 -s 0 | grep PhysMem

# 좀 더 정리되어 보이게 하기
# PhysMem: 15G used (2868M wired
#          7017M compressor)
#          69M unused.
top -l 1 -s 0 | grep PhysMem | sed 's/, /\n         /g'
```
