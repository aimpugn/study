# AWS Swap Memory

- [AWS Swap Memory](#aws-swap-memory)
    - [Swap memory](#swap-memory)
    - [설정하기](#설정하기)
        - [`dd` 명령어로 스왑 파일 생성하기](#dd-명령어로-스왑-파일-생성하기)
        - [`fallocate`로 스왑 파일 생성하기](#fallocate로-스왑-파일-생성하기)
    - [swap 용량?](#swap-용량)
    - [스왑 파일 사용](#스왑-파일-사용)
    - [기타](#기타)

## Swap memory

스왑 메모리(또는 스왑 공간)는 시스템의 물리적 RAM(랜덤 액세스 메모리)의 확장으로 사용되는 컴퓨터 저장소(일반적으로 HDD 또는 SSD)의 일부입니다.
시스템은 자주 사용하지 않는 메모리 페이지를 스왑 공간으로 이동시켜 RAM을 확보합니다.

Linux 시스템에서 스왑은 두 가지 방식으로 구현할 수 있습니다:
- 스왑 파티션: 스왑 전용으로 사용되는 하드 드라이브의 전용 파티션입니다.
- 스왑 파일: 스왑 공간 역할을 하는 파일 시스템 내의 특수 파일입니다.

## 설정하기

### `dd` 명령어로 스왑 파일 생성하기

1. `dd` 명령어 실행하여 루트 파일 시스템에 스왑 파일을 생성합니다.

    > Note: It's a best practice to create swap space only on ephemeral storage instance store volumes.
    >
    > 여기서 [ephemeral storage instance store volumes](https://docs.aws.amazon.com/ko_kr/AWSEC2/latest/UserGuide/InstanceStorage.html)은 EBS가 아닌 별도의 인스턴스 스토어라는 블록 스토리지입니다.

    ```sh
    $ sudo dd if=/dev/zero of=/swapfile bs=128M count=32
    32+0 records in
    32+0 records out
    4294967296 bytes (4.3 GB, 4.0 GiB) copied, 27.7181 s, 155 MB/s
    ```

    파일 시스템의 특성에 관계없이 연속된 블록을 할당합니다.
    파일 시스템 독립적으로 작동할 수 있으며, 연속된 블록을 할당하므로 파편화가 적습니다.

    - `if`: input file
    - `of`: output file
    - `bs`: block size
    - `count`: number of block

    스왑 파일의 크기는 $bs \times count = 128M \times 32 = 4096M$

    블록 크기는 인스턴스에서 사용 가능한 메모리보다 반드시 작은 값을 지정해야 합니다.
    그렇지 않으면 memory exhausted 에러를 받게 됩니다.

    > `bs=128M count=32` vs `bs=1M count=4096`?
    >
    > - bs=128M count=32
    >
    >     각 작업마다 128MB를 전송하며, 총 32번의 작업으로 4GB의 데이터를 생성합니다.
    >
    >     - 각 작업마다 128MB의 메모리 버퍼를 사용합니다.
    >     - 스왑 파일 생성시 32번의 I/O 작업으로 더 빠를 수 있습니다.
    >     - 스왑 파일을 생성할 때 큰 블록 크기(`bs=128M`)를 사용하면 파일 시스템이 더 큰 연속된 공간을 할당하려고 시도하므로 파일의 단편화를 줄일 수 있습니다.
    >     - 큰 블록 크기로 인해 디스크 캐시 효율성이 떨어질 수 있습니다.
    >     - 큰 단위로 작업하므로 오류 발생 시 더 많은 데이터를 재작성해야 할 수 있습니다.
    >     - 각 작업이 더 오래 걸려 시스템의 반응성에 영향을 줄 수 있습니다.
    >
    > - bs=1M count=4096
    >
    >     각 작업마다 1MB를 전송하며, 총 4096번의 작업으로 동일한 4GB의 데이터를 생성합니다.
    >
    >     - 각 작업마다 1MB의 메모리 버퍼를 사용합니다.
    >     - 스왑 파일 생성시 4096번의 I/O 작업으로 더 느릴 수 있습니다.
    >     - 작은 블록 크기(`bs=1M`)를 사용하면 파일 시스템이 작은 블록을 여러 번 할당하므로, 파일이 디스크 상에서 더 많이 단편화될 수 있습니다.
    >     - 작은 블록 크기로 인해 디스크 캐시를 더 효율적으로 사용할 수 있습니다.
    >     - 작은 단위로 작업하므로 중간에 오류 발생 시 복구가 더 쉽습니다.
    >     - 더 자주 프로세스 스케줄러에 제어권을 반환합니다(더 자주 컨텍스트 스위칭 가능).
    >
    > 이때 주의할 점은, 스왑 파일 생성 시 `dd` 명령어의 `bs`와 `count` 값은 파일 생성 시간과 메모리 사용량에 영향을 줄 수 있지만, **실제 스왑 사용 시의 성능에는 큰 차이가 없습니다**.
    > 스왑 파일은 생성된 후에 운영체제의 메모리 관리 시스템에 의해 페이지 단위(보통 4KB)로 관리되기 때문에, 스왑 파일의 내부 구조는 추상화됩니다.
    > 스왑 파일이 어떻게 생성되었는지와 관계없이 운영체제는 스왑 파일을 연속된 주소 공간으로 간주하고 필요한 위치에 페이지를 저장합니다.

2. 스왑 파일에 대한 읽고 쓰기 권한을 추가합니다.

    ```sh
    $ sudo chmod 600 /swapfile
    ```

    - `400`: Allow read by owner
    - `200`: Allow write by owner

3. 리눅스 스왑 영역을 초기화합니다.

    ```sh
    $ sudo mkswap /swapfile
    mkswap: /swapfile: warning: wiping old swap signature.
    Setting up swapspace version 1, size = 4 GiB (4294963200 bytes)
    no label, UUID=9346ce0f-abf4-40ca-b2bb-759aa0ffa3e0
    ```

4. 스왑 파일을 스왑 공간에 추가하고 스왑 공간을 즉시 사용할 수 있도록 합니다.

    ```sh
    $ sudo swapon /swapfile
    ```

5. 절차가 성공했는지 확인합니다.

    ```sh
    $ sudo swapon -s
    ```

    ```sh
    ❯ free -h
                   total        used        free      shared  buff/cache   available
    Mem:           3.8Gi       540Mi       228Mi       928Ki       3.4Gi       3.3Gi
    Swap:          4.0Gi          0B       4.0Gi
    ```

6. 스왑 파일을 부트시에 시작하려면, `/etc/fstab` 파일을 수정합니다.

    ```sh
    $ sudo vi /etc/fstab

    # Add the following new line at the end of the file:
    # /swapfile swap swap defaults 0 0
    ```

    파일을 저장하고 종료합니다.

### `fallocate`로 스왑 파일 생성하기

`fallocate`는 실제로 파일에 데이터를 쓰지 않고도 디스크 공간을 할당하므로, 시간과 시스템 자원을 절약할 수 있습니다.
다만, 일부 파일 시스템에서는 `fallocate`로 생성된 스왑 파일을 지원하지 않을 수 있으므로, `dd` 명령어를 사용하는 것이 더 호환성이 높을 수 있습니다.

1. 스왑 파티션 또는 스왑 파일을 생성합니다.

    ```sh
    sudo fallocate -l 2G /swapfile
    ```

    실제로 데이터를 쓰지 않고 파일 시스템 수준에서 공간만 예약합니다.
    하지만 실제 데이터를 쓰지 않기 때문에 파일 시스템 특성에 따라 연속성이 보장되지 않을 수 있습니다.

    > `fallocate`
    >
    > 파일에 할당된 디스크 공간을 조작하여 할당을 해제하거나 사전 할당하는 데 사용됩니다.
    > `fallocate`(2) 시스템 호출을 지원하는 파일시스템의 경우, 블록을 할당하고 초기화되지 않은 것으로 표시하여 데이터 블록에 대한 IO가 필요 없는 사전 할당을 빠르게 수행합니다.
    > 이는 파일을 0으로 채워서 생성하는 것보다 훨씬 빠릅니다.

2. 스왑 파일의 권한을 수정합니다.

    ```sh
    sudo chmod 600 /swapfile
    ```

3. 스왑 영역을 초기화합니다.

    ```sh
    sudo mkswap /swapfile
    ```

4. 스왑 메모리를 활성화 합니다.

    ```sh
    sudo swapon /swapfile
    ```

## [swap 용량?](https://help.ubuntu.com/community/SwapFaq#How_much_swap_do_I_need.3F)

물리적 메모리(RAM)가 1GB 미만인 경우 스왑 공간은 기본적으로 RAM 용량과 동일한 최소값을 유지하는 것이 좋습니다.
또한 수확 체감(diminishing returns) 때문에 스왑 공간은 시스템에서 사용할 수 있는 하드 디스크 공간의 양에 따라 최대 RAM 용량의 두 배를 사용하는 것이 좋습니다.

> 수확 체감?
>
> 어떤 생산요소의 투입을 고정시키고 다른 생산요소의 투입을 증가시킬 경우 산출량이 점진적으로 증가하다가 투입량이 일정수준을 넘게 되면 산출량의 중가율이 점차적으로 감소하게 되는 현상을 의미합니다.

최신 시스템(1GB 이상)의 경우:
- "hibernation 사용하는 경우" 스왑 공간은 최소한 물리적 메모리(RAM) 크기와 같아야 하며,
    > Hibernation(최대 절전 모드, 휴면 모드)
    >
    > 컴퓨터의 현재 상태를 디스크에 저장하고 전원을 완전히 끄는 기능입니다.
    > 다시 시작할 때 이전 상태로 빠르게 복원됩니다.

- 그렇지 않으면 최소 `round(sqrt(RAM))`에서 최대 2배의 RAM 용량을 사용해야 합니다.
실제 사용량보다 많은 스왑 공간을 확보하는 것의 유일한 단점은 스왑을 위해 확보해야 하는 디스크 공간입니다.

"수확 체감"은 RAM 2배 크기의 swap이 필요하다면, RAM을 추가하는 게 좋다는 의미입니다.
HDD 액세스는 RAM 액세스보드 10^3(1000)배 느리기 때문에, 이는 1초 소요되던 것이 15분 이상 소요될 수 있습니다.
그리고 빠른 SSD에서도 여전히 1분 이상 소요됩니다.

| RAM    | No hibernation | With Hibernation | Maximum |
| :----- | -------------- | ---------------- | ------- |
| 256MB  | 256MB          | 512MB            | 512MB   |
| 512MB  | 512MB          | 1024MB           | 1024MB  |
| 1024MB | 1024MB         | 2048MB           | 2048MB  |

| RAM   | No hibernation | With Hibernation | Maximum |
| :---- | -------------- | ---------------- | ------- |
| 1GB   | 1GB            | 2GB              | 2GB     |
| 2GB   | 1GB            | 3GB              | 4GB     |
| 3GB   | 2GB            | 5GB              | 6GB     |
| 4GB   | 2GB            | 6GB              | 8GB     |
| 5GB   | 2GB            | 7GB              | 10GB    |
| 6GB   | 2GB            | 8GB              | 12GB    |
| 8GB   | 3GB            | 11GB             | 16GB    |
| 12GB  | 3GB            | 15GB             | 24GB    |
| 16GB  | 4GB            | 20GB             | 32GB    |
| 24GB  | 5GB            | 29GB             | 48GB    |
| 32GB  | 6GB            | 38GB             | 64GB    |
| 64GB  | 8GB            | 72GB             | 128GB   |
| 128GB | 11GB           | 139GB            | 256GB   |
| 256GB | 16GB           | 272GB            | 512GB   |
| 512GB | 23GB           | 535GB            | 1TB     |
| 1TB   | 32GB           | 1056GB           | 2TB     |
| 2TB   | 46GB           | 2094GB           | 4TB     |
| 4TB   | 64GB           | 4160GB           | 8TB     |
| 8TB   | 91GB           | 8283GB           | 16TB    |

## 스왑 파일 사용

- 페이지 아웃 (RAM에서 스왑으로):
    - 메모리 부족 시 커널은 덜 사용되는 메모리 페이지를 스왑 공간으로 이동시키며 이때 I/O 발생합니다.
    - 이는 생성 시의 블록 크기와 무관하게 일반적으로 페이지 단위(보통 4KB)로 발생

- 페이지 인 (스왑에서 RAM으로):
    - 스왑된 페이지가 필요할 때, 커널이 스왑 파일에서 RAM으로 데이터를 읽어올 때 I/O 발생합니다.
    - 마찬가지로 페이지 단위로 발생

- 스왑 공간 검색:
    - 시스템이 특정 데이터가 스왑에 있는지 확인할 때 발생하는 읽기 I/O

실제 스왑 사용 시의 I/O는 운영 체제의 메모리 관리 정책과 페이지 크기에 따라 결정됩니다.
스왑 사용 중 I/O 성능은 스왑 파일이 위치한 스토리지의 특성(예: SSD vs HDD)에 더 큰 영향을 받습니다.

## 기타

- [How do I allocate memory to work as a swap file in an Amazon EC2 instance?](https://repost.aws/knowledge-center/ec2-memory-swap-file)
    - [All about Linux swap space](https://www.linux.com/news/all-about-linux-swap-space/)
- [[Linux] 간단하게 복사 붙여넣기로 Ubuntu EC2에 Swap 메모리 설정하기](https://engineerinsight.tistory.com/276)
