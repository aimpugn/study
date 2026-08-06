# 리눅스 메모리 계측 — 커널이 무엇을 세고, 도구가 어떻게 보여 주는가

`free` 의 `available` 은 어디서 오는가. `vmstat` 의 `si`/`so` 는 무엇의 차분인가. "메모리 사용률 90%" 라는 한 문장은 어느 필드로 계산된 것인가.
이 문서는 커널이 메모리를 관리하면서 무엇을 세고, 그 카운터가 어느 파일로 노출되며, `free`·`vmstat`·`sar`·모니터링 에이전트가 그것을 어떻게 가공해 화면에 올리는지를 한 줄기로 정리한다.

**이 문서는 판정 기준이나 임계치 권고를 담지 않는다.** "이 숫자는 무엇인가" 와 "어떻게 확인하는가" 까지만 닫는다. "몇 %면 문제인가" 는 워크로드·용량·서비스 성격에 따라 달라지므로 다루지 않는다. 다만 **어떤 숫자를 골라야 그 판단이 가능해지는가** 는 5장에서 근거와 함께 정리한다.

짝이 되는 실행물이 같은 디렉터리에 있다.

- [memory_metrics_verify.sh](./memory_metrics_verify.sh) — 이 문서의 주장을 실행으로 확인하는 읽기 전용 스크립트 (7장)

이웃 문서와의 경계는 이렇다.

- 스왑의 개념·동작·운영 영향 → [swap.md](./swap.md)
- `vmstat` 전체 컬럼 사전과 I/O 판정 → [storage_io_vmware.md 6.1](../commands/metrics/storage_io_vmware.md) (이 문서는 메모리 컬럼만 다루고 나머지는 그쪽을 가리킨다)
- macOS `vm_stat` (이름만 비슷한 다른 도구) → [vm_stat.md](../commands/metrics/vm_stat.md)

- [0. 결론부터 — 이 문서가 닫는 판단 다섯 가지](#0-결론부터--이-문서가-닫는-판단-다섯-가지)
- [1. 커널이 메모리를 다루는 방식](#1-커널이-메모리를-다루는-방식)
- [2. 커널이 내보내는 카운터 — 어디에 무엇이 있는가](#2-커널이-내보내는-카운터--어디에-무엇이-있는가)
- [3. MemAvailable — 커널이 계산해 주는 유일한 "여유" 추정치](#3-memavailable--커널이-계산해-주는-유일한-여유-추정치)
- [4. 도구는 그 숫자를 어떻게 가공하는가](#4-도구는-그-숫자를-어떻게-가공하는가)
- [5. "메모리 사용률" — 같은 서버, 같은 순간, 다른 숫자](#5-메모리-사용률--같은-서버-같은-순간-다른-숫자)
- [6. 흔한 오독과 그 구조적 원인](#6-흔한-오독과-그-구조적-원인)
- [7. 검증 절차](#7-검증-절차)
- [8. 참고 자료](#8-참고-자료)

검증 환경 — 이 문서에 실린 실측값은 모두 아래 환경에서 [memory_metrics_verify.sh](./memory_metrics_verify.sh) 로 얻었다.

| 항목 | 값 |
| --- | --- |
| 커널 | `6.18.33.2-microsoft-standard-WSL2` (x86_64, PAGE_SIZE 4096) |
| 배포판 | Ubuntu on WSL2 |
| procps-ng | 4.0.4 |
| MemTotal / SwapTotal | 8028284 kB / 2097152 kB |

WSL2 는 호스트와 메모리를 동적으로 주고받는 특수 환경이라 **절대 수치의 대표성은 없다.** 이 문서가 실측으로 닫는 것은 값이 아니라 **유도 관계** — 어떤 필드에서 어떤 필드가 나오는가 — 이며, 그 관계는 일반 리눅스에서도 같다. 독자의 환경에서 다시 확인하는 절차가 7장이다.

---

## 0. 결론부터 — 이 문서가 닫는 판단 다섯 가지

이 문서를 읽은 뒤 아래 다섯 가지를 자기 말로 설명할 수 있어야 한다.

1. **`MemFree` 가 적은 것은 정상이다.** 리눅스는 남는 물리 메모리를 파일 캐시로 채운다. 비어 있는 메모리는 성능상 낭비이므로, 잘 도는 서버일수록 `MemFree` 는 바닥에 붙어 있다. 부족의 지표는 `MemFree` 가 아니다.

2. **커널이 "여유" 로 인정하는 값은 `MemAvailable` 하나뿐이다.** 이것은 단순 합이 아니라 커널 함수 `si_mem_available()` 이 계산한 **추정치**이고, 회수 불가능한 캐시(tmpfs 등)와 커널 예약분을 제외한다. 3장에서 계산식을 재현해 오차 0.0017% 로 검증한다.

3. **`free` 의 `used` 컬럼은 버전에 따라 정의가 다르다.** procps-ng 4.x 는 `total - available` 로 계산한다 (man page 명시 + 실측 확인). 구버전은 `total - free - buffers - cache` 였다. `used` 는 파생값이므로 계약으로 삼지 말고 원본 필드를 직접 읽어야 한다.

4. **`/proc/meminfo` 는 게이지(현재값), `/proc/vmstat` 은 대부분 누적 카운터다.** 이 구분을 놓치면 `vmstat` 첫 줄을 현재 상태로 읽거나, 누적값을 순간값으로 착각하는 오독이 생긴다.

5. **"메모리 사용률" 이라는 단일 숫자는 존재하지 않는다.** 같은 스냅샷에 계산식만 바꿔도 결과가 수십 %p 갈린다. 어떤 식을 썼는지 모르는 사용률 숫자는 해석할 수 없다.

같은 순간의 실측이 5번을 그대로 보여 준다.

```text
$ ./memory_metrics_verify.sh compare-usage
계산식                                          결과  무엇을 재는가
(MemTotal - MemFree) / MemTotal                 22.2%  커널이 손대지 않은 페이지 외 전부. 캐시 포함
(MemTotal - MemAvailable) / MemTotal             7.1%  새 할당에 내줄 수 없는 몫. 커널 추정
(MemTotal - MemFree - buffcache) / Total         5.2%  익명 페이지 + 회수 불가 커널 메모리
buffcache / MemTotal                            17.0%  캐시가 차지한 몫 (참고)
```

22.2% 와 7.1% 는 서로 다른 서버의 값이 아니다. **같은 서버, 같은 `/proc/meminfo` 스냅샷 하나에서 나온 값이다.** 검증 환경은 여유가 많아 간격이 15%p 에 그쳤고, 캐시를 많이 쓰는 서버에서는 훨씬 크게 벌어진다 — 5.1 의 예에서는 약 29%p 다.

---

## 1. 커널이 메모리를 다루는 방식

계측 이야기를 하기 전에, 커널이 무엇을 하고 있는지부터 잡는다. 뒤에 나오는 모든 카운터는 이 동작의 부산물이다.

### 1.1 페이지 — 계측의 최소 단위

커널은 물리 메모리를 **페이지**(x86_64 기본 4 KiB) 단위로 쪼개 관리한다. 모든 메모리 카운터는 결국 "어떤 상태의 페이지가 몇 장인가" 이고, `/proc/meminfo` 가 kB 로 보여 주는 것은 그 장수에 페이지 크기를 곱한 값이다.

```bash
getconf PAGE_SIZE     # 4096
```

`/proc/vmstat` 은 대부분 **페이지 장수 그대로**, `/proc/meminfo` 는 **kB 로 환산해서** 낸다. 두 파일을 섞어 계산할 때 이 단위 차이가 첫 번째 함정이다.

### 1.2 두 종류의 페이지 — 이 구분이 나머지 전부를 결정한다

메모리가 부족할 때 커널이 페이지를 회수하는 방법은 그 페이지에 **원본이 있느냐**로 갈린다.

| 종류 | 정체 | 내용의 원본 | 회수 방법 | 대표 필드 |
| --- | --- | --- | --- | --- |
| **file-backed** | 파일 내용을 담은 페이지 (페이지 캐시) | 디스크의 파일 | clean 이면 **그냥 버린다**. dirty 면 디스크에 먼저 쓴다 | `Cached`, `Active(file)`, `Inactive(file)` |
| **anonymous** | 힙, 스택, `MAP_ANONYMOUS` | **없다** | **스왑에 써야만** 내보낼 수 있다. 스왑이 없으면 회수 불가 | `AnonPages`, `Active(anon)`, `Inactive(anon)` |

이 표 한 장에서 세 가지가 따라 나온다.

- **clean file-backed 페이지의 회수는 거의 공짜다.** 링크를 끊고 페이지를 넘기면 끝이며 디스크 I/O 가 없다. 그래서 파일 캐시는 "쓰고 있지만 필요하면 즉시 내놓는" 메모리다.
- **dirty 페이지는 공짜가 아니다.** 디스크 쓰기가 선행되므로 회수에 지연이 붙는다. 갑작스러운 대량 할당이 stall 로 이어지는 경로 중 하나가 이것이다.
- **anonymous 페이지는 스왑 없이는 회수할 수 없다.** 스왑을 끈 시스템에서 메모리가 마르면 커널이 할 수 있는 일은 파일 캐시를 끝까지 버리는 것뿐이고, 그다음은 OOM Killer 다.

**가장 헷갈리는 예외가 tmpfs/shmem 이다.** `/dev/shm`, `tmpfs` 마운트, System V 공유 메모리는 `Cached` 에 집계되지만 디스크에 원본 파일이 없다. **회계상으로는 캐시, 성질상으로는 익명** 이라서 스왑으로만 나갈 수 있다. `free` 의 `shared` 컬럼이 이 몫이고, 3장에서 볼 `MemAvailable` 계산이 이것을 명시적으로 제외하는 이유다.

### 1.3 LRU 리스트 — 무엇부터 버릴지 정하는 자료구조

커널은 회수 후보를 네 개의 LRU(Least Recently Used) 리스트로 관리한다. 1.2 의 두 종류 × active/inactive 조합이다.

```text
                 active                    inactive
              (최근에 쓰임)            (회수 후보 대기열)
 anon    Active(anon)     <--승격--   Inactive(anon)    --회수--> 스왑
 file    Active(file)     <--승격--   Inactive(file)    --회수--> 그냥 버림(clean)

 Unevictable : mlock 된 페이지 등, 회수 대상에서 아예 제외
```

회수는 inactive 리스트의 꼬리부터 일어난다. 참조된 페이지는 active 로 올라가고, 오래 안 쓰인 페이지는 inactive 로 내려온다. `/proc/meminfo` 의 `Active`/`Inactive` 는 이 리스트의 현재 길이다.

> 커널 6.1 부터 MGLRU(Multi-Gen LRU)라는 대체 구현이 선택적으로 들어갔다. 내부 알고리즘은 다르지만 `/proc/meminfo` 로 나오는 필드 이름과 의미는 유지된다. 활성 여부는 `/sys/kernel/mm/lru_gen/enabled` 로 확인한다 (검증 환경에서는 해당 파일이 없어 비활성). **독자 환경에서의 활성 여부는 확인 필요.**

### 1.4 워터마크 — 언제 회수를 시작하는가

커널은 zone 별로 세 개의 수위를 두고, 빈 페이지가 그 아래로 내려가면 회수를 발동한다.

| 수위 | 넘어서면 벌어지는 일 |
| --- | --- |
| `high` | 회수 목표선. kswapd 는 여기까지 확보하면 다시 잠든다 |
| `low` | **kswapd 기동.** 백그라운드 회수라 할당하는 프로세스는 멈추지 않는다 |
| `min` | **direct reclaim.** 할당하려던 프로세스가 직접 회수를 수행한다 — **그 프로세스가 그동안 멈춘다** |

이 값들은 `/proc/zoneinfo` 에 **페이지 단위**로 있다. 검증 환경 실측:

```text
Node 0, zone DMA32  | min 5538  low 6922  high 8306   protection: (0, 3979, 3979, 3979)
Node 0, zone Normal | min 5725  low 7156  high 8587   protection: (0, 0, 0, 0)
```

`low` 합계 14110 pages = 56440 kB. 8 GiB 시스템에서 약 55 MB 다.

관련 sysctl 은 `vm.min_free_kbytes`(검증 환경 45056)와 `vm.watermark_scale_factor`(기본 10 = 0.1%)다.
**이 값을 공식으로 추정하지 말고 `/proc/zoneinfo` 를 직접 읽어야 한다.** 커널 버전, NUMA 노드 수, hugepage 설정, 배포판 기본값에 따라 달라지고, `min_free_kbytes` 를 명시적으로 올려 둔 서버도 흔하다.

`min` 아래로 내려가 direct reclaim 이 발생한 횟수는 `/proc/vmstat` 의 `allocstall_*` 로 누적된다. 이 카운터가 증가한다는 것은 **프로세스가 실제로 메모리 때문에 멈췄다** 는 직접 증거다.

### 1.5 dirty 페이지와 writeback

`write()` 는 페이지를 더럽히고(dirty) 즉시 반환한다. 실제 디스크 쓰기는 writeback 스레드가 나중에 몰아서 한다.

- `vm.dirty_background_ratio` (기본 10%) 를 넘으면 백그라운드로 내려쓰기 시작
- `vm.dirty_ratio` (기본 20%) 를 넘으면 **쓰는 프로세스 자체를 멈춰 세운다**(throttling)

현재 dirty 량은 `/proc/meminfo` 의 `Dirty`, 진행 중인 쓰기는 `Writeback` 이다. 이 메커니즘이 I/O 지표에 미치는 영향은 [storage_io_vmware.md 3.4](../commands/metrics/storage_io_vmware.md) 에 정리돼 있다.

### 1.6 스왑 — 무엇을 얼마나 내보낼지

`vm.swappiness` (검증 환경 기본값 60) 는 회수 시 **익명 페이지와 파일 캐시 중 어느 쪽을 선호할지**의 가중치다. 높을수록 익명 페이지(= 프로세스 힙)를 스왑으로 내보내는 쪽을 선호한다. 상한은 커널 버전에 따라 100 또는 200 이므로 **자기 환경에서 확인해야 한다** (`Documentation/admin-guide/sysctl/vm.rst`).

`swappiness=0` 도 스왑 완전 금지가 아니다. 커널 3.5 이후 0 은 "압박이 심각하지 않은 한 익명 회수를 피한다" 는 의미이며, 정말 마르면 스왑한다.

계측에서 반드시 구분해야 할 두 가지가 있다.

- **스왑 사용량(`SwapTotal - SwapFree`, `vmstat` 의 `swpd`)** — 게이지. 과거에 나간 페이지가 아직 돌아오지 않은 누적 상태다. **값이 있다는 것만으로는 현재 압박의 증거가 아니다.**
- **스왑 활동(`pswpin`/`pswpout`, `vmstat` 의 `si`/`so`)** — 카운터의 차분. **지금** 페이지가 오가고 있다는 뜻이다.

이 둘의 비대칭도 중요하다. `so > 0` 은 지금 내보내는 중이므로 현재 압박이지만, `si` 만 있고 `so`=0 이면 과거 스왑 잔재를 늦게 되읽는 중일 수 있다 — 리눅스는 메모리가 남아돌아도 스왑 페이지를 선제적으로 복귀시키지 않기 때문이다. 사례 분석은 [storage_io_vmware.md 10장](../commands/metrics/storage_io_vmware.md) 에 있다.

### 1.7 오버커밋 — 약속한 양과 쓰는 양

리눅스는 기본적으로 실제 물리 메모리보다 많은 양을 할당해 준다. 프로세스가 `malloc` 한 영역을 전부 건드리지는 않기 때문이다.

| sysctl | 값 | 의미 |
| --- | --- | --- |
| `vm.overcommit_memory` | 0 (기본) | 휴리스틱. 터무니없는 요청만 거절하고 `CommitLimit` 은 강제하지 않는다 |
| | 1 | 항상 허용 |
| | 2 | 엄격. `Committed_AS` 가 `CommitLimit` 을 넘으면 할당 실패 |
| `vm.overcommit_ratio` | 50 (기본) | 모드 2 에서 쓰는 비율 |

```text
CommitLimit = SwapTotal + MemTotal * overcommit_ratio / 100
```

검증 환경 실측 — `2097152 + 8028284 * 50/100 = 6111294`, 커널 값 `6111292` (정수 나눗셈 차이 2 kB).

`Committed_AS` 는 **약속한 총량**이지 사용량이 아니다. 물리 메모리를 초과해도 모드 0 에서는 아무 일도 일어나지 않는다. 사용량 지표로 오해하기 쉬운 필드다.

---

## 2. 커널이 내보내는 카운터 — 어디에 무엇이 있는가

### 2.1 게이지와 카운터 — 모든 오독의 출발점

| 성격 | 의미 | 읽는 법 | 예 |
| --- | --- | --- | --- |
| **게이지(gauge)** | 지금 이 순간의 상태값 | 한 번 읽으면 끝 | `MemFree`, `Dirty`, `swpd` |
| **카운터(counter)** | 부팅 이후 단조 증가하는 누적값 | **두 시점을 읽어 차분해야** 의미가 생긴다 | `pswpin`, `pgfault`, `oom_kill` |

`vmstat 1` 의 첫 줄이 부팅 이후 평균인 이유가 여기 있다. 카운터에서 파생된 컬럼(`si`/`so`/`bi`/`bo`/`in`/`cs`/CPU%)은 직전 시점이 없으니 부팅 시점을 기준으로 나눌 수밖에 없다. 반면 게이지 컬럼(`r`/`b`/`swpd`/`free`/`buff`/`cache`)은 첫 줄도 현재값이다.

**두 성격이 한 줄에 섞여 나온다는 것이 `vmstat` 출력을 읽을 때 가장 먼저 알아야 할 사실이다.**

### 2.2 인터페이스 지도

| 경로 | 성격 | 담는 것 | 대표 소비자 |
| --- | --- | --- | --- |
| `/proc/meminfo` | 게이지 | 시스템 전체 메모리 현황 (50여 필드) | `free`, `top`, `vmstat` 의 메모리 컬럼 |
| `/proc/vmstat` | 대부분 카운터, `nr_*` 는 게이지 | 페이징·스왑·회수 **이벤트** | `vmstat` 의 `si`/`so`/`bi`/`bo`, `sar -B`/`-W` |
| `/proc/zoneinfo` | 게이지 | zone 별 워터마크, free, LRU 길이 | 진단 (도구가 잘 안 읽는다) |
| `/proc/stat` | 카운터 | CPU 시간, `procs_running`/`procs_blocked` | `vmstat` 의 `r`/`b`/CPU |
| `/proc/pressure/memory` | 카운터 + 이동평균 | 메모리 때문에 **지연된 시간** | PSI 기반 모니터 |
| `/proc/<pid>/status`, `smaps_rollup` | 게이지 | 프로세스별 RSS/PSS/Swap | `ps`, `top`, `pmap` |
| `/sys/fs/cgroup/**/memory.*` | 게이지 + 카운터 | cgroup 한도와 사용량 | 컨테이너 런타임, k8s |

### 2.3 `/proc/meminfo` 필드 사전

주요 필드만 추린다. 전체 정의는 커널 `Documentation/filesystems/proc.rst` 에 있다.

| 필드 | 의미 | 주의점 |
| --- | --- | --- |
| `MemTotal` | 커널이 관리하는 물리 메모리 총량 | 펌웨어 예약분·커널 이미지가 빠져 있어 장착 용량보다 작다 |
| `MemFree` | 완전히 비어 있는 페이지 | **적은 것이 정상.** 부족 지표가 아니다 |
| `MemAvailable` | 스왑 없이 새 할당에 내줄 수 있는 양의 **추정치** | 단순 합이 아니다. 3장 |
| `Buffers` | 블록 디바이스 메타데이터 캐시 | 요즘은 대개 작다 |
| `Cached` | 페이지 캐시 | **`Shmem` 을 포함한다.** 전부 회수 가능하지 않다 |
| `SwapCached` | 스왑에 있으면서 메모리에도 남아 있는 페이지 | 스왑을 안 썼으면 0 |
| `Active`/`Inactive` | LRU 리스트 길이 | `(anon)` + `(file)` 의 합 |
| `Unevictable`/`Mlocked` | 회수 대상에서 제외된 페이지 | `mlock`, ramdisk 등 |
| `Dirty` | 아직 디스크에 안 쓴 수정된 페이지 | 회수에 쓰기 지연이 붙는 몫 |
| `Writeback` | 지금 디스크로 쓰는 중인 페이지 | |
| `AnonPages` | 파일 원본이 없는 사용자 페이지 | 힙·스택 |
| `Mapped` | 프로세스 주소 공간에 매핑된 파일 페이지 | `Cached` 의 부분집합 |
| `Shmem` | tmpfs/공유 메모리 | **`Cached` 안에 있지만 회수 불가.** `free` 의 `shared` |
| `Slab` | 커널 자료구조 | `SReclaimable + SUnreclaim` |
| `SReclaimable` | 회수 가능한 slab (dentry·inode 캐시) | `free` 의 `buff/cache` 에 포함된다 |
| `SUnreclaim` | 회수 불가 slab | |
| `KReclaimable` | 회수 가능한 커널 메모리 전체 | `SReclaimable` ⊆ `KReclaimable` |
| `PageTables` | 페이지 테이블이 먹는 메모리 | 프로세스가 많으면 무시 못 할 크기 |
| `CommitLimit`/`Committed_AS` | 오버커밋 한도 / 약속한 총량 | 1.7 |

검산으로 확인되는 관계다. 아래 네 줄은 각각 하나의 `/proc/meminfo` 스냅샷 안에서 성립한다 (스냅샷이 서로 다르므로 줄 사이의 값은 비교하지 말 것).

```text
Slab (66144)      = SReclaimable (22284) + SUnreclaim (43860)              정확히 일치
Active (216632)   = Active(anon) 3028    + Active(file) 213604             정확히 일치
Inactive(1233960) = Inactive(anon)109304 + Inactive(file) 1124656          정확히 일치
KReclaimable (22284) >= SReclaimable (22284)                               포함 관계 (이 순간엔 같음)
```

`KReclaimable = SReclaimable + NR_KERNEL_MISC_RECLAIMABLE` 이므로 후자가 0 이면 두 값이 같아진다. 검증 환경이 그 경우였다. **`>` 가 아니라 `>=` 관계라는 점**만 기억하면 된다.

`free` 컬럼과의 대응(`buff/cache = Buffers + Cached + SReclaimable`, `shared = Shmem`)은 4.1 에서 차이 0 으로 확인한다.

**`MemTotal` 은 나머지 필드의 단순 합과 일치하지 않는다.** 커널 자체가 쓰는 메모리가 여러 항목에 흩어져 있고 일부는 어느 항목에도 안 잡히기 때문이다. `/proc/meminfo` 는 분할 회계표가 아니라 **관점별 집계표**이며, 항목끼리 겹친다(`Shmem` 은 `Cached` 안에 있고, `Mapped` 도 `Cached` 안에 있다). 합계를 맞추려 들면 안 된다.

### 2.4 `/proc/vmstat` 주요 카운터

| 카운터 | 단위 | 의미 |
| --- | --- | --- |
| `pgpgin` / `pgpgout` | **kB** | 블록 디바이스에서 읽고 쓴 양 (`vmstat` 의 `bi`/`bo` 원천) |
| `pswpin` / `pswpout` | **페이지** | 스왑 in/out (`vmstat` 의 `si`/`so` 원천) |
| `pgfault` | 회 | 페이지 폴트 전체 |
| `pgmajfault` | 회 | **major fault** — 디스크를 실제로 읽어야 했던 폴트 |
| `pgscan_kswapd` / `pgsteal_kswapd` | 페이지 | 백그라운드 회수가 훑은 양 / 실제로 뺏은 양 |
| `pgscan_direct` / `pgsteal_direct` | 페이지 | **direct reclaim** 이 훑은 양 / 뺏은 양 |
| `allocstall_*` | 회 | direct reclaim 발생 횟수 — **프로세스가 멈춘 횟수** |
| `oom_kill` | 회 | OOM Killer 발동 횟수 |
| `nr_free_pages`, `nr_dirty`, `nr_writeback` | 페이지 | `nr_` 로 시작하면 게이지 |

**이름은 둘 다 `pg`(page)로 시작하지만 단위가 다르다.** `pgpgin`/`pgpgout` 은 kB, `pswpin`/`pswpout` 은 페이지 장수다. 두 카운터를 같은 코드에서 다룰 때 이 비대칭이 사고 지점이다.

`vmstat -s` 가 원시값을 그대로 통과시키면서 붙이는 단위 라벨로 확인된다.

```text
$ grep -E '^(pgpgin|pgpgout|pswpin|pswpout) ' /proc/vmstat
pgpgin 659079
pgpgout 8752
pswpin 0
pswpout 0

$ vmstat -s | grep -iE 'paged|swapped'
       659079 K paged in        <-- 같은 값에 "K" 를 붙였다 = kB 로 해석
         8752 K paged out
            0 pages swapped in  <-- 이쪽은 "pages"
            0 pages swapped out
```

그래서 `vmstat` 이 화면에 낼 때 `bi`/`bo` 는 변환 없이 KiB/s 가 되고, `si`/`so` 는 페이지 크기를 곱해 KiB/s 로 바꾼다. **이것은 procps-ng 4.0.4 의 해석을 실측으로 확인한 것이며 커널 소스로 직접 닫은 것은 아니다.** 다른 커널·procps 조합에서 쓰려면 위 대조를 다시 해 보는 편이 안전하다.

`pgscan` 과 `pgsteal` 의 비율도 유용하다. 많이 훑었는데 적게 뺏었다면(scan ≫ steal) 회수할 만한 페이지가 별로 없다는 뜻이고, 이는 압박이 심하다는 신호다.

### 2.5 PSI — 양이 아니라 결과를 재는 지표

커널 4.20 부터 `/proc/pressure/memory` 가 생겼다 (`CONFIG_PSI` 필요).

```text
$ cat /proc/pressure/memory
some avg10=0.00 avg60=0.00 avg300=0.00 total=0
full avg10=0.00 avg60=0.00 avg300=0.00 total=0
```

- `some` — **하나 이상의** 태스크가 메모리 때문에 지연된 시간의 비율
- `full` — **모든** non-idle 태스크가 동시에 멈춘 시간의 비율
- `avg10`/`avg60`/`avg300` — 백분율 이동평균 (10초/60초/300초)
- `total` — 누적 지연 시간 (마이크로초)

PSI 의 성격이 앞의 모든 지표와 다르다. **메모리가 얼마나 남았는지가 아니라, 메모리 때문에 실제로 얼마나 느려졌는지를 직접 잰다.** 5장에서 볼 "어떤 계산식을 쓸 것인가" 논쟁 자체를 우회하는 지표이며, 사용률과 달리 워크로드가 달라져도 의미가 유지된다.

`full` 이 0 보다 크다면 그 시간 동안 시스템 전체가 메모리 회수를 기다리며 멈춰 있었다는 뜻이다. 해석에 계산식이 끼어들 여지가 없다.

### 2.6 cgroup — 컨테이너에서 달라지는 것

**컨테이너 안에서 `/proc/meminfo` 를 읽으면 호스트 전체의 값이 나온다.** `lxcfs` 같은 것으로 덮어쓰지 않는 한 그렇다. 컨테이너 메모리 모니터링 오독의 최대 원인이 이것이며, 컨테이너에 1 GiB 한도가 걸려 있어도 `/proc/meminfo` 는 호스트의 128 GiB 를 보고한다.

cgroup 의 값을 봐야 한다.

| cgroup v2 | v1 대응 | 의미 |
| --- | --- | --- |
| `memory.max` | `memory.limit_in_bytes` | 하드 한도. 넘으면 OOM |
| `memory.high` | (없음) | 소프트 한도. 넘으면 throttle |
| `memory.current` | `memory.usage_in_bytes` | 현재 사용량 — **페이지 캐시를 포함한다** |
| `memory.stat` | `memory.stat` | `anon`/`file`/`kernel`/`slab` 등 세부 내역 |
| `memory.events` | | `low`/`high`/`max`/`oom`/`oom_kill` 발생 횟수 |
| `memory.pressure` | (없음) | cgroup 단위 PSI |

**`memory.current` 에도 캐시가 포함된다.** 즉 5장에서 다룰 함정이 컨테이너 계층에서 그대로 반복된다. 실제 압박은 `memory.stat` 의 `anon` 과 `memory.events` 의 증가, 그리고 `memory.pressure` 로 본다.

검증 환경의 `memory.stat` 발췌:

```text
anon 111022080
file 1373175808
kernel 25460736
pagetables 5304320
shmem 3608576
file_mapped 162471936
file_dirty 3948544
```

### 2.7 프로세스 단위 — RSS 의 함정

| 지표 | 의미 | 함정 |
| --- | --- | --- |
| `VSZ` | 가상 주소 공간 크기 | 실제로 쓰는 물리 메모리와 무관하다. 거의 쓸모없다 |
| `RSS` | 물리 메모리에 올라와 있는 양 | **공유 페이지를 각 프로세스가 중복 계산한다** |
| `PSS` | 공유 페이지를 공유자 수로 나눈 몫 | 프로세스별 합계가 의미를 갖는 유일한 지표 |
| `USS` | 그 프로세스만 쓰는 몫 | 죽였을 때 회수되는 양 |

**프로세스들의 RSS 합계는 시스템 메모리 사용량과 일치하지 않는다.** 공유 라이브러리, `fork` 후 CoW 페이지, 공유 메모리가 중복으로 세어지기 때문이다. 프로세스별로 나눠 보려면 PSS 를 써야 한다.

```bash
grep -E '^(Rss|Pss|Private|Shared|Swap)' /proc/<pid>/smaps_rollup
grep -E '^Vm(RSS|Swap)|^Rss(Anon|File|Shmem)' /proc/<pid>/status
```

`RssAnon`/`RssFile`/`RssShmem` 분해가 특히 유용하다. 1.2 의 구분이 프로세스 단위로 그대로 내려온 것이라, `RssAnon` 이 크면 스왑으로만 회수 가능한 몫이 크다는 뜻이다.

---

## 3. MemAvailable — 커널이 계산해 주는 유일한 "여유" 추정치

### 3.1 왜 MemFree 로는 안 되는가

리눅스는 남는 물리 메모리를 파일 캐시로 채운다. 비워 두면 아무 이득이 없고, 캐시로 쓰면 디스크 읽기를 줄이기 때문이다. 그 결과 **정상적으로 잘 도는 서버의 `MemFree` 는 항상 바닥에 가깝다.**

그런데 그 캐시의 상당 부분은 필요하면 즉시 내줄 수 있다. 그래서 "지금 새 프로세스에 얼마나 줄 수 있는가" 라는 질문의 답은 `MemFree` 가 아니라 `MemFree + 회수 가능한 캐시` 여야 한다.

문제는 **캐시 전부가 회수 가능한 것이 아니라는 점**이다(1.2 의 tmpfs, dirty 페이지, 커널이 붙들고 있는 slab). 사용자 공간에서 정확히 계산하기 어렵다. 그래서 커널 3.14 에서 이 추정을 커널이 직접 해 주는 필드가 추가됐다.

`free(1)` man page 는 `available` 을 "스왑 없이 새 애플리케이션을 시작하는 데 쓸 수 있는 메모리의 추정치" 로 정의하면서, `cache` 나 `free` 필드와 달리 **페이지 캐시를 고려하고 회수 가능 slab 이 전부 회수되지는 않는다는 점까지 반영한다** 고 덧붙인다. 도입 시점도 그 자리에 적혀 있다 — "available on kernels 3.14" (procps-ng 4.0.4 기준).

### 3.2 계산식

커널 `mm/page_alloc.c` 의 `si_mem_available()` 이다.

```text
available  = MemFree      - totalreserve_pages
           + pagecache    - min(pagecache   / 2, wmark_low)
           + reclaimable  - min(reclaimable / 2, wmark_low)

  pagecache          = NR_ACTIVE_FILE + NR_INACTIVE_FILE
  reclaimable        = NR_SLAB_RECLAIMABLE + NR_KERNEL_MISC_RECLAIMABLE   (~= KReclaimable)
  wmark_low          = Σ_zone ( low 워터마크 )
  totalreserve_pages = Σ_zone ( max(lowmem_reserve[]) + high 워터마크 )
```

여기서 읽어야 할 설계 결정이 셋이다.

1. **`pagecache` 는 파일 LRU 만 센다.** `NR_ACTIVE_FILE + NR_INACTIVE_FILE` 이다. tmpfs/shmem 페이지는 익명 LRU 에 있으므로 **자동으로 빠진다.** 1.2 에서 말한 "회계상 캐시, 성질상 익명" 이 여기서 정확히 처리된다.

2. **캐시를 통째로 여유로 치지 않는다.** `min(pagecache/2, wmark_low)` 만큼을 남긴다. 커널 주석의 표현대로 "캐시를 전부 비우면 시스템이 스왑하거나 스래싱하기 시작" 하기 때문이다. 캐시가 충분히 크면(`pagecache/2 > wmark_low`) 실제 차감액은 `wmark_low` 로 고정된다.

3. **커널 예약분을 뺀다.** `totalreserve_pages` 는 zone 별 high 워터마크와 lowmem_reserve 의 합으로, 커널이 절대 사용자에게 내주지 않는 몫이다.

### 3.3 재현 검증

위 계산식을 사용자 공간에서 재현해 커널 값과 대조했다.

```text
$ ./memory_metrics_verify.sh check-avail
  PAGE_SIZE          = 4096 B
  wmark_low 합계     = 14110 pages = 56440
  totalreserve 합계  = 20863 pages = 83452
  MemFree            = 6245924
  pagecache          = Active(file) 116980 + Inactive(file) 1221796 = 1338776
  KReclaimable       = 23092
  (MemFree - totalreserve)          = 6162472
  + (pagecache   - min(half,wlow))  = 1282336   [차감 56440]
  + (reclaimable - min(half,wlow))  = 11546     [차감 11546]
  = 재현값                          = 7456354
    커널 MemAvailable               = 7456484
    차이                            = -130  (-0.0017%)
```

**오차 130 kB (0.0017%).** 남은 차이는 `/proc/meminfo` 와 `/proc/zoneinfo` 를 연속으로 읽는 사이의 드리프트다. 계산식은 위와 같다고 봐도 된다.

`reclaimable` 항의 차감액이 `wmark_low`(56440)가 아니라 `11546` 인 것에 주목할 것. `KReclaimable` 이 23092 로 작아서 `min()` 의 앞쪽 항(절반 = 11546)이 선택됐다. 커널 주석이 말한 "적어도 절반, 또는 low 워터마크만큼" 이 그대로 동작한 것이다.

### 3.4 무엇에 둔감하도록 설계됐는가

`MemAvailable` 의 가장 중요한 성질은 **파일 캐시의 증감에 불변** 이라는 것이다. 계산식으로 확인할 수 있다.

clean 파일 캐시가 300 MB 늘어나고 그만큼 `MemFree` 가 줄었다고 하자.

```text
MemFree   항:  -300 MB
pagecache 항:  +300 MB - Δ[min(pagecache/2, wmark_low)]
                        └─ pagecache/2 > wmark_low 인 동안 min() 은 wmark_low 로 고정 → Δ = 0
                                          합계:  0
```

**불변식 성립 조건은 `pagecache / 2 > wmark_low` 다.** 캐시가 `wmark_low` 의 두 배보다 크면 성립하고, 캐시가 아주 작은 시스템에서는 완전히 불변은 아니다. 검증 환경 기준으로 캐시가 약 110 MB 를 넘으면 조건이 충족된다.

같은 상황에서 다른 지표는 이렇게 움직인다.

| 지표 | 캐시 300 MB 증가 시 |
| --- | --- |
| `MemFree` | **-300 MB** |
| `(MemTotal - MemFree) / MemTotal` | 검증 환경(7840 MiB) 기준 **+3.8 %p** |
| `MemAvailable` | **변화 없음** |
| `(MemTotal - MemAvailable) / MemTotal` | **변화 없음** |

즉 `MemFree` 기반 지표는 **파일을 읽는 행위 자체를 메모리 소비로 계측한다.** `MemAvailable` 기반 지표는 그것을 구조적으로 배제한다. 이 차이는 정적인 값의 높낮이보다 **시계열의 노이즈**에서 훨씬 크게 드러나며, 추세 분석이나 예측 알고리즘을 얹으면 그대로 증폭된다.

실제로 관찰하려면 파일 I/O 가 있는 시간대에 다음을 돌린다.

```bash
./memory_metrics_verify.sh watch-usage 12 5
```

---

## 4. 도구는 그 숫자를 어떻게 가공하는가

### 4.1 free

컬럼별 출처를 실측으로 대조한 결과다 (procps-ng 4.0.4, 차이 전부 0).

| 컬럼 | 유도식 |
| --- | --- |
| `total` | `MemTotal` |
| `free` | `MemFree` |
| `available` | `MemAvailable` |
| `shared` | `Shmem` |
| `buff/cache` | `Buffers + Cached + SReclaimable` |
| `used` | **버전에 따라 다르다 — 아래** |

`free -w` 를 쓰면 `buff/cache` 가 `buffers` 와 `cache` 로 분리된다. man page 가 각각 "Buffers in /proc/meminfo", "Cached and SReclaimable in /proc/meminfo" 로 명시한다 — 즉 `cache` 쪽에 회수 가능 slab 이 섞여 있다.

**`used` 의 정의가 procps-ng 버전에 따라 바뀌었다.** 4.x 의 man page 는 이렇게 적는다.

> "used — Used or unavailable memory (calculated as total - available)"
> — `man 1 free`, procps-ng 4.0.4

실측으로도 확정된다.

```text
$ ./memory_metrics_verify.sh check-free
  used 후보 A  total - available                        = 571588
               free.used 와 차이                        = 0
  used 후보 B  total - free - buffers - cached - srecl  = 416916
               free.used 와 차이                        = -154672
```

후보 B 는 구버전(procps-ng 3.3.x 계열)의 정의이고 지금은 15만 kB 어긋난다. **`used` 는 파생값이므로 계약으로 삼으면 안 된다.** 스크립트나 모니터링에서 이 컬럼을 파싱하고 있다면, 배포판을 올릴 때 의미가 바뀐다. `/proc/meminfo` 의 원본 필드를 직접 읽는 편이 안전하다.

자기 환경의 정의를 확인하는 방법은 둘이다.

```bash
man free | grep -A2 '^ *used'
./memory_metrics_verify.sh check-free
```

### 4.2 vmstat

전체 컬럼 사전은 [storage_io_vmware.md 6.1](../commands/metrics/storage_io_vmware.md) 에 있다. 여기서는 메모리 관련 부분만 짚는다.

| 컬럼 | 출처 | 성격 |
| --- | --- | --- |
| `swpd` | `SwapTotal - SwapFree` | 게이지 |
| `free` | `MemFree` | 게이지 |
| `buff` | `Buffers` | 게이지 |
| `cache` | `Cached` | 게이지 |
| `si` / `so` | `/proc/vmstat` 의 `pswpin`/`pswpout` 차분 | **카운터 파생** (기본 KiB/s, `--unit` 영향) |

메모리·스왑 컬럼은 `vmstat --unit` 옵션의 영향을 받고 `bi`/`bo` 는 받지 않는다. man page 가 `bi` 를 "Kibibyte received from a block device (KiB/s)" 로 못박은 반면 `si`/`so` 는 "Amount of memory swapped in from disk (/s)" 로만 적고 단위를 `--unit` 에 넘긴 것이 그 차이다. 스크립트로 파싱한다면 `-S k` 를 명시해 단위를 고정하는 편이 안전하다.

**`vmstat` 의 메모리 컬럼에는 `MemAvailable` 이 없다.** `free`, `buff`, `cache` 만 있으므로, 이 출력만 보고 메모리 여유를 판단할 수 없다. `vmstat -a` 를 쓰면 `buff`/`cache` 대신 `active`/`inactive` 가 나오는데 이것도 여유 판단에는 쓸 수 없다. 여유는 `free -w` 나 `/proc/meminfo` 로 따로 봐야 한다.

한 가지 주의할 표기가 있다. **`vmstat -s` 는 `Cached` 를 "swap cache" 라는 라벨로 출력한다.**

```text
$ vmstat -s | grep -iE 'buffer|cache|swap'
        51632 K buffer memory
       993624 K swap cache        <-- 실제로는 Cached. SwapCached 가 아니다
      2097152 K total swap
            0 K used swap
            0 pages swapped in
```

검증 환경은 스왑을 한 번도 쓰지 않아 `SwapCached` 가 0 인데 "swap cache" 는 993624 K 를 보고한다. procps-ng 4.0.4 에서 관찰된 라벨 오류이며, 값 자체는 `Cached` 다. **`vmstat -s` 의 라벨을 신뢰하지 말고 `/proc/meminfo` 를 원본으로 삼아야 한다.**

### 4.3 sar (sysstat)

`vmstat` 이 지금을 보여 준다면 `sar` 는 **과거로 거슬러 갈 수 있다.** `sysstat` 이 주기적으로 `/var/log/sa/saNN` 에 기록해 두기 때문이다.

```bash
sar -r 1 5                    # 메모리 — kbavail, %memused, kbcached 등
sar -S 1 5                    # 스왑 사용량
sar -W 1 5                    # 스왑 in/out 비율
sar -B 1 5                    # 페이징 — pgpgin/s, majflt/s, pgscan/s, pgsteal/s
sar -r -f /var/log/sa/sa15    # 이번 달 15일의 기록
```

`sar -r` 출력에는 `kbmemfree`, `kbavail`, `kbmemused`, `%memused`, `kbbuffers`, `kbcached` 가 함께 나온다. `kbavail` 이 `MemAvailable` 이므로 **5장의 차이를 한 출력 안에서 직접 대조할 수 있다.**

> `%memused` 가 캐시를 사용 중으로 세는지는 **sysstat 버전에 따라 다르므로 확인이 필요하다.** 검증 환경에 sysstat 이 없어 이 문서에서 닫지 못했다. 확인은 같은 출력 안에서 산수로 하면 된다 — `%memused` 가 `(kbmemtotal - kbmemfree) / kbmemtotal` 에 가까우면 캐시 포함, `(kbmemtotal - kbavail) / kbmemtotal` 에 가까우면 캐시 제외다. `sar -r 1 1` 한 번이면 갈린다.

### 4.4 top / htop

`top` 의 메모리 요약 줄도 `/proc/meminfo` 에서 온다. `free`/`used`/`buff/cache`/`avail Mem` 을 보여 주며, `avail Mem` 이 `MemAvailable` 이다.

프로세스별 `%MEM` 은 **RSS 기준**이므로 2.7 의 중복 계산 문제를 그대로 갖는다. `%MEM` 합계가 100% 를 넘는 것은 버그가 아니라 공유 페이지가 여러 번 세어진 결과다.

### 4.5 언어 런타임과 모니터링 에이전트

에이전트가 시스템 메모리 사용률을 보고할 때 **어느 필드를 읽는지는 구현마다 다르다.** 그리고 대부분 문서화하지 않는다.

특히 `MemAvailable` 은 2014년(커널 3.14)에 추가된 필드다. 그 이전에 작성된 수집 코드는 구조적으로 `MemFree` 밖에 쓸 수 없었고, **그 코드가 그대로 남아 있는 경우가 흔하다.** 모니터링 에이전트의 시스템 메트릭 수집부는 한 번 만들면 잘 건드리지 않는 영역이기도 하다.

> 개별 런타임·라이브러리가 어느 필드를 읽는지는 **버전마다 다르며 이 문서에서 확정하지 않았다.** 아래 검증 경로로 직접 확인해야 한다.

확인 절차는 필드에 의존하지 않는 방식이라 어떤 에이전트에도 적용된다.

1. 에이전트가 보고하는 사용률 값을 기록한다.
2. **같은 시각**에 대상 호스트에서 `./memory_metrics_verify.sh compare-usage` 를 실행한다.
3. 출력된 네 값 중 어느 것과 일치하는지 본다.

- **PASS**: 한 값과 1~2%p 이내로 일치 → 그 계산식을 쓰고 있다.
- **FAIL**: 어느 값과도 안 맞음 → 컨테이너 안에서 호스트 `/proc/meminfo` 를 읽고 있거나(2.6), cgroup 값을 쓰거나, 자체 보정을 하고 있다. 셋 다 별도 확인이 필요하다.

---

## 5. "메모리 사용률" — 같은 서버, 같은 순간, 다른 숫자

### 5.1 후보 계산식

"메모리 사용률" 이라는 이름으로 유통되는 식은 최소 넷이다.

| 계산식 | 무엇을 재는가 | 성질 |
| --- | --- | --- |
| `(MemTotal - MemFree) / MemTotal` | 커널이 손대지 않은 페이지를 뺀 전부. **캐시 포함** | 운영 중인 서버에서 항상 높다. 캐시 변동에 민감 |
| `(MemTotal - MemAvailable) / MemTotal` | 새 할당에 내줄 수 없는 몫 | 커널 추정. 캐시 변동에 불변(3.4) |
| `(MemTotal - MemFree - buff/cache) / MemTotal` | 익명 페이지 + 회수 불가 커널 메모리 | 회수 가능성을 과대평가한다 (tmpfs·dirty 를 여유로 셈) |
| `buff/cache / MemTotal` | 캐시가 차지한 몫 | 사용률이 아니라 참고값 |

캐시를 많이 쓰는 서버의 예로 16 GiB 시스템을 놓고 계산해 보면 차이가 뚜렷해진다.

```text
MemTotal 16384  MemFree 1200  Buffers 50  Cached 5400 (Shmem 500 포함)
SReclaimable 150  MemAvailable 5900          (단위 MiB)

(16384 - 1200) / 16384            = 92.7%    <- 캐시 포함
(16384 - 5900) / 16384            = 64.0%    <- 커널이 보는 압박
(16384 - 1200 - 5600) / 16384     = 58.5%    <- 익명 + 회수 불가
```

**세 숫자 모두 틀리지 않았다.** 서로 다른 질문에 답하고 있을 뿐이다. 문제는 이 중 어느 것인지 밝히지 않은 채 "메모리 사용률 92%" 라고 전달될 때 생긴다.

### 5.2 캐시 변동에 대한 민감도

위 예에서 파일을 읽어 clean 캐시가 500 MiB 늘고 그만큼 `MemFree` 가 줄었다고 하자.

| 계산식 | 변화 전 | 변화 후 | 변동 |
| --- | --- | --- | --- |
| `(Total - Free) / Total` | 92.7% | 95.7% | **+3.0 %p** |
| `(Total - Available) / Total` | 64.0% | 64.0% | 0.0 %p |
| `(Total - Free - buff/cache) / Total` | 58.5% | 58.5% | 0.0 %p |

**시스템에는 아무 문제도 생기지 않았고 오히려 캐시 적중률이 올라갔는데, 첫 번째 식만 3 %p 움직였다.**

이 민감도가 실무에서 문제가 되는 지점은 절대값이 아니라 **시계열**이다. 파일 I/O 가 오르내릴 때마다 첫 번째 식은 계속 출렁이고, 그 위에 추세 분석·이동평균·예측 같은 것을 얹으면 캐시 변동이 그대로 "메모리가 증가하는 추세" 로 해석된다. 반면 두 번째 식은 그 입력을 애초에 받지 않는다.

### 5.3 그래서 무엇을 봐야 하는가

계산식 선택 자체를 피하는 길이 둘 있다.

1. **`MemAvailable` 을 직접 본다.** 사용률로 환산하지 않고 절대량(kB)과 총량 대비 비율을 함께 기록한다. 커널이 이미 어려운 판단을 대신 해 둔 값이다.
2. **PSI 를 본다(2.5).** "얼마나 남았는가" 가 아니라 "메모리 때문에 실제로 얼마나 멈췄는가" 를 재므로 계산식 논쟁이 성립하지 않는다.

보조 신호로 함께 볼 것들이다.

- `/proc/vmstat` 의 `allocstall_*` 증가 — direct reclaim 이 실제로 발생했다
- `pswpout` 증가 — 지금 스왑으로 내보내는 중이다 (`pswpin` 단독은 과거 잔재일 수 있다, 1.6)
- `pgscan_*` ≫ `pgsteal_*` — 훑는데 뺏을 게 없다
- `oom_kill` 증가 — 이미 죽였다

이 카운터들은 모두 **"압박이 있었다" 는 사후 사실**이지 추정이 아니다. 사용률처럼 계산식에 따라 달라지지 않는다.

---

## 6. 흔한 오독과 그 구조적 원인

| 오독 | 왜 틀렸는가 | 대신 볼 것 |
| --- | --- | --- |
| `free` 가 적으니 메모리가 부족하다 | 리눅스는 남는 메모리를 캐시로 채운다. `MemFree` 가 낮은 것이 정상 상태다 | `MemAvailable` |
| `buff/cache` 가 크니 메모리를 낭비하고 있다 | 캐시는 대부분 즉시 회수 가능하며, 없으면 읽기가 전부 디스크로 떨어진다 | 캐시는 성능 자산이다 |
| 스왑이 잡혀 있으니 메모리가 부족하다 | `swpd` 는 게이지다. 과거에 나간 페이지가 안 돌아온 상태일 뿐 | `pswpout` 의 증가율 |
| `si` 가 보이니 지금 압박이다 | `si` 단독(so=0)은 과거 스왑 잔재를 되읽는 중일 수 있다 | `so` 와 `MemAvailable` 을 함께 |
| `vmstat` 첫 줄이 현재 상태다 | 카운터 파생 컬럼은 부팅 이후 평균이다 | 둘째 줄부터 |
| 컨테이너 안 `/proc/meminfo` 가 컨테이너 메모리다 | 호스트 전체 값이 나온다 | `/sys/fs/cgroup/memory.*` |
| 프로세스 RSS 를 합치면 시스템 사용량이다 | 공유 페이지가 중복 계산된다 | PSS, 또는 시스템 레벨 지표 |
| `Committed_AS` 가 물리 메모리를 넘었으니 위험하다 | 약속한 양이지 쓰는 양이 아니다. 모드 0 에서는 한도가 강제되지도 않는다 | 실제 사용량 지표 |
| `free` 의 `used` 를 파싱해 쓰면 된다 | 버전에 따라 정의가 바뀌는 파생값이다 | `/proc/meminfo` 원본 필드 |
| `vmstat -s` 의 "swap cache" 가 SwapCached 다 | procps-ng 4.0.4 에서 `Cached` 를 그 라벨로 출력한다 | `/proc/meminfo` 의 `SwapCached` |
| JVM 힙이 정상이니 시스템 메모리도 정상이다 | 힙 **사용률**과 OS 가 보는 **commit 크기**는 다른 값이다. GC 가 객체를 정리해도 페이지는 커널로 안 돌아간다 | 프로세스 RSS 와 `MemAvailable` 을 따로 |

마지막 항목을 조금 더 풀면 이렇다. JVM 같은 런타임은 힙을 한 번 확보하면(commit) 그 페이지를 OS 관점에서 계속 점유한다. 힙 내부 사용률이 20% 든 90% 든 OS 는 구분하지 못한다. 반대로 **힙이 커지는 순간에는 페이지 캐시와 직접 경쟁한다** — 새 페이지를 처음 건드릴 때 페이지 폴트가 나고, 빈 페이지가 모자라면 커널이 회수(1.4)를 돌려 캐시를 축출하기 때문이다. 즉 "런타임 내부 사용률" 과 "시스템 여유 메모리" 는 무관하지만, "런타임 commit 크기" 와 "시스템 여유 메모리" 는 직결된다.

---

## 7. 검증 절차

[memory_metrics_verify.sh](./memory_metrics_verify.sh) 는 읽기 전용이며 시스템 상태를 바꾸지 않는다. `sudo` 도 필요 없다.

```bash
./memory_metrics_verify.sh all              # check-* 전부
./memory_metrics_verify.sh check-free       # free 컬럼 <- /proc/meminfo
./memory_metrics_verify.sh check-avail      # si_mem_available() 재현
./memory_metrics_verify.sh check-commit     # CommitLimit 검산
./memory_metrics_verify.sh compare-usage    # 계산식 네 개 동시 비교
./memory_metrics_verify.sh watch-usage 12 5 # 5초 간격 12회 변동 관찰
./memory_metrics_verify.sh snapshot         # 진단용 원시 덤프
```

판정 기준.

| 검사 | PASS | FAIL 이면 |
| --- | --- | --- |
| `check-free` | `total`/`free`/`available`/`shared`/`buff/cache` 차이가 수 MB 이내 (스냅샷 드리프트) | procps 가 다른 소스를 읽고 있다. 그 환경에서는 4.1 표를 신뢰하면 안 된다 |
| `check-free` (used) | 후보 A 또는 B 중 하나가 0 에 가까움 | 그 버전의 `used` 정의를 별도로 조사해야 한다 |
| `check-avail` | 재현값과 커널 `MemAvailable` 의 오차 1% 이내 | 커널 버전이 다른 계산식을 쓰고 있다. `mm/page_alloc.c` 확인 필요 |
| `check-commit` | 계산식과 커널 `CommitLimit` 이 정수 나눗셈 오차 범위 내 일치 | `overcommit_ratio` 대신 `overcommit_kbytes` 가 설정돼 있을 수 있다 |
| `compare-usage` | 네 값이 모두 출력됨 | — (비교용이며 판정 항목 아님) |
| `watch-usage` | `MemFree` 가 움직이는 동안 avail 기준이 거의 고정 | 3.4 의 불변식 조건(`pagecache/2 > wmark_low`)을 확인 |

주의할 점 셋.

- **스냅샷 드리프트는 정상이다.** `/proc/meminfo` 와 `/proc/zoneinfo` 를 연속으로 읽는 사이 수 ms 가 지나므로 오차가 남는다. 수 MB 는 정상, 수백 MB 는 다른 식이라는 신호다.
- **컨테이너 안에서는 `check-avail` 이 SKIP 될 수 있다.** `/proc/zoneinfo` 가 없거나 호스트 값이면 결과가 무의미하다.
- **WSL2·가상 환경의 절대값은 대표성이 없다.** 유도 관계 검증에는 쓸 수 있지만 용량 판단에는 쓸 수 없다.

---

## 8. 참고 자료

커널 문서 (`Documentation/` 아래, 또는 <https://docs.kernel.org>)

- `filesystems/proc.rst` — `/proc/meminfo` 전 필드 정의
- `admin-guide/sysctl/vm.rst` — `min_free_kbytes`, `swappiness`, `dirty_*`, `overcommit_*`
- `admin-guide/mm/concepts.rst` — 메모리 관리 개념
- `accounting/psi.rst` — PSI 정의와 해석
- `admin-guide/cgroup-v2.rst` — cgroup v2 메모리 컨트롤러

커널 소스

- `mm/page_alloc.c` 의 `si_mem_available()` — 3.2 의 계산식 원본
- `fs/proc/meminfo.c` — `/proc/meminfo` 를 만들어 내는 코드
- `MemAvailable` 은 커널 3.14 에서 도입 (`free(1)` man page 에 명시)

man page

- `proc(5)` — `/proc/meminfo`, `/proc/vmstat` 필드 설명
- `free(1)` — 컬럼 정의. **버전마다 다르므로 자기 환경의 것을 읽어야 한다**
- `vmstat(8)`, `sar(1)`, `top(1)`

저장소 내부

- [swap.md](./swap.md) — 스왑의 개념과 운영 영향
- [storage_io_vmware.md](../commands/metrics/storage_io_vmware.md) — `vmstat` 전체 컬럼 사전(6.1), 페이지 캐시와 I/O(3.4), 스왑 경보 사례(10장)
- [vm_stat.md](../commands/metrics/vm_stat.md) — macOS `vm_stat` (이름만 비슷한 다른 도구)
