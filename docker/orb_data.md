# OrbStack data

## `data.img`

```sh
❯ ll ~/.orbstack/data/data.img
-rw-r--r--@ 1 rody  staff   8.0T  8 31 16:03 /Users/rody/.orbstack/data/data.img
```

`data.img` 파일은 [Sparse file](https://en.wikipedia.org/wiki/Sparse_file)입니다.
이는 실제로 사용하는 데이터만큼만 물리적 디스크 공간을 차지하는 특별한 유형의 파일입니다.

그래서 실제 파일 크기를 확인해 보면 다음과 같습니다:

```sh
❯ du -h ~/.orbstack/data/data.img
6.7G    /Users/rody/.orbstack/data/data.img
```

sparse file은 사용량이 다양한 대용량 파일을 빠르고 효율적이며 유연하게 저장할 수 있는 `APFS` 기능입니다.

파일 시스템은 파일의 전체 크기(예: 8.0T)를 기록하지만, 실제로 데이터가 있는 부분만 디스크에 저장합니다.
나머지 "빈" 공간은 논리적으로만 존재하며 물리적 공간을 차지하지 않습니다.

백업 시스템 (예: Backblaze)은 이러한 파일을 전체 크기로 인식할 수 있습니다.
이로 인해 실제보다 훨씬 큰 백업 크기가 표시될 수도 있습니다.

- [Why is there an 8 TB data file?](https://docs.orbstack.dev/faq#data-img)
