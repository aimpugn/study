# Note

- [Note](#note)
    - [xv6와 riscv](#xv6와-riscv)
    - [riscv64-elf tools](#riscv64-elf-tools)
        - [riscv64-elf-gcc](#riscv64-elf-gcc)
        - [riscv64-elf-readelf](#riscv64-elf-readelf)
        - [riscv64-elf-objdump](#riscv64-elf-objdump)
        - [riscv64-elf-nm](#riscv64-elf-nm)
    - [asm](#asm)
    - [qemu](#qemu)
    - [The following two regions overlap (in the memory address space)](#the-following-two-regions-overlap-in-the-memory-address-space)
    - [relocation truncated to fit: R\_RISCV\_HI20](#relocation-truncated-to-fit-r_riscv_hi20)
    - [CFLAGS](#cflags)
    - [qemu 종료 방법](#qemu-종료-방법)
    - ['a'가 아니라 'aaa'로 출력되는 이유](#a가-아니라-aaa로-출력되는-이유)
    - [DTB(Device Tree Blob)와 `dtc`(Device Tree Compiler)](#dtbdevice-tree-blob와-dtcdevice-tree-compiler)

## xv6와 riscv

- xv6
    - `x`: x86(32비트) 아키텍처
    - `v6`: 1975년 발표되었던 [Unix Version 6](https://en.wikipedia.org/wiki/Version_6_Unix)

- risc-v: 오픈소스 CPU 명령어 집합 구조(ISA, Instruction Set Architecture)

    - `RISC`:
        - Reduced Instruction Set Computing
        - 명령어 세트를 최대한 단순화한 CPU 설계 철학으로, 명확한 명령어와 일관성, 적은 트랜지스터로 높은 성능과 에너지 효율성을 목표로 합니다.

            > 대부분의 CPU들이 방대한 명령어를 가지고 있음에도 실제로 자주 사용하는 명령어는 전체 명령어의 10%도 미치지 못함에 착안하여 명령어의 갯수를 줄이고, 그 대신 본질적인 CPU 성능 개선 방안들을 적용:
            > - CPU 인사이드 캐시
            > - 분기 예측 기능
            > - 수퍼스케일러
            > - 비순차 명령 실행
            > - 파이프라이닝
            > - 레지스터 개수 증가 등

        - 가령 기존 x86은 `CISC`(Complex Instruction Set Computing) 구조의 경우 명령어가 매우 다양하고 복잡하다고 합니다.
            - [CISC, RISC, CRISC(EPIC)](https://moltak.tistory.com/143)
            - [나무위키 RISC](https://namu.wiki/w/RISC)
            - [CICS vs RISC](https://jwprogramming.tistory.com/20)
            - [RISC, CISC, and EPIC difference?](https://www.reddit.com/r/cpudesign/comments/9j9zvd/risc_cisc_and_epic_difference)
    - `V`: 5번째 RISC 아키텍처(Version 5)

## riscv64-elf tools

### riscv64-elf-gcc

```sh
❯ riscv64-elf-gcc -v
Using built-in specs.
COLLECT_GCC=riscv64-elf-gcc
COLLECT_LTO_WRAPPER=/opt/homebrew/Cellar/riscv64-elf-gcc/15.1.0/libexec/gcc/riscv64-elf/15.1.0/lto-wrapper
Target: riscv64-elf
Configured with: ../configure --target=riscv64-elf --prefix=/opt/homebrew/Cellar/riscv64-elf-gcc/15.1.0 --infodir=/opt/homebrew/Cellar/riscv64-elf-gcc/15.1.0/share/info/riscv64-elf --disable-nls --without-isl --without-headers --with-as=/opt/homebrew/opt/riscv64-elf-binutils/bin/riscv64-elf-as --with-ld=/opt/homebrew/opt/riscv64-elf-binutils/bin/riscv64-elf-ld --with-system-zlib --enable-languages=c,c++
Thread model: single
Supported LTO compression algorithms: zlib zstd
gcc version 15.1.0 (GCC)
```

### riscv64-elf-readelf

```sh
❯ riscv64-elf-readelf -v
GNU readelf (GNU Binutils) 2.45
Copyright (C) 2025 Free Software Foundation, Inc.
This program is free software; you may redistribute it under the terms of
the GNU General Public License version 3 or (at your option) any later version.
This program has absolutely no warranty.
```

### riscv64-elf-objdump

```sh
❯ riscv64-elf-objdump -v
GNU objdump (GNU Binutils) 2.45
Copyright (C) 2025 Free Software Foundation, Inc.
This program is free software; you may redistribute it under the terms of
the GNU General Public License version 3 or (at your option) any later version.
This program has absolutely no warranty.
```

### riscv64-elf-nm

```sh
❯ riscv64-elf-nm --version
GNU nm (GNU Binutils) 2.45
Copyright (C) 2025 Free Software Foundation, Inc.
This program is free software; you may redistribute it under the terms of
the GNU General Public License version 3 or (at your option) any later version.
This program has absolutely no warranty.
```

## asm

```c
asm volatile("csrr %0, mstatus" : "=r"(x));
```

- `volatile`
    - 컴파일러가 이 `asm` 블록을 최적화로 제거하거나 재배치하지 않도록 보장합니다.
    - CSR 읽기 같은 하드웨어 동작은 부수효과가 있기 때문에 `volatile`을 붙입니다.
- `csrr %0, mstatus`에서 `%0` 은 GCC가 할당한 목적지 레지스터 자리입니다.
- `"=r"(x)` 제약조건이 `"=r"` 이므로, GCC는 이 값을 일반 레지스터에 넣어야 함을 알고, 자유로운 임시 레지스터를 하나 골라 `%0`에 매핑합니다.

```asm
csrr a5, mstatus   # mstatus CSR을 읽어서 a5 레지스터에 저장
                   # a5는 GCC가 %0에 할당한 실제 레지스터입니다.
sd   a5, -8(s0)    # a5 값을 지역변수 x의 메모리 위치에 저장
```

## qemu

qemu는 보드, CPU, 주변장치를 소프트웨어로 흉내내는 전체 시스템 에뮬레이터입니다.

선택한 머신의 메모리 맵과 장치를 소프트웨어로 구현합니다.
실행 시 DRAM에 바이너리들을 배치하고 CPU의 PC와 특권 모드를 초기화합니다.

RISC‑V `virt`에서 두 경로가 있습니다.
1. 펌웨어가 먼저 올라가서(M‑모드) 초깃값 설정과 위임을 한 뒤 S‑모드 커널을 호출하는 경로
2. 펌웨어를 생략하고 DRAM의 지정 주소에서 곧장 실행을 시작하는 경로

전자가 현실적인 보드(펌웨어 존재)에 가깝고, 후자는 교육용으로 부트 경로를 단순화합니다.

`-machine virt`로 실행하면 DRAM(기본 0x80000000 시작), UART, PLIC, CLINT 같은 가상 하드웨어가 만들어지고,
펌웨어(BIOS) 유무에 따라 부팅 흐름이 달라집니다.

- `-bios default`
    - qemu가 번들한 펌웨어(리눅스 배포판의 qemu는 보통 [OpenSBI fw_dynamic](https://github.com/riscv-software-src/opensbi/blob/master/docs/firmware/fw_dynamic.md))를 DRAM에 올립니다.
    - qemu가에 내장된 OpenSBI를 M‑모드 펌웨어로 올려서 S‑모드 OS를 부팅합니다.
    - 이 펌웨어는 0x80000000 부근을 차지하고 S‑모드 커널로 점프하면서 `a0=hartid, a1=DTB` 포인터를 넘겨줍니다.

- `-bios none`
    - 펌웨어를 올리지 않습니다.
    - qemu가 `-kernel`로 준 바이너리를 DRAM 시작(관례적으로 0x80000000)에 놓고 거기서 실행을 시작합니다.

따라서 만약 커널을 0x80000000에 링크한다면 `-bios default`와 충돌하고, `-bios none`이면 충돌하지 않습니다.
반대로 OpenSBI를 쓰고 싶다면 커널을 0x8020_0000 등으로 올리면 된다고 합니다.

## The following two regions overlap (in the memory address space)

```sh
❯ make qemu
qemu-system-riscv64 -machine virt -nographic -bios default -kernel kernel.elf
qemu-system-riscv64: Some ROM regions are overlapping
These ROM regions might have been loaded by direct user request or by default.
They could be BIOS/firmware images, a guest kernel, initrd or some other file loaded into guest memory.
Check whether you intended to load all this guest code, and whether it has been built to load to the correct addresses.

The following two regions overlap (in the memory address space):
  /opt/homebrew/Cellar/qemu/10.0.3/bin/../share/qemu/opensbi-riscv64-generic-fw_dynamic.bin (addresses 0x0000000080000000 - 0x0000000080042878)
  kernel.elf ELF program header segment 1 (addresses 0x0000000080000000 - 0x000000008000103e)
make: *** [qemu] Error 1
```

qemu를 `-bios default`로 띄워 OpenSBI가 DRAM 시작(0x80000000)에 올라갔고,
kernel.elf도 같은 주소(0x80000000)로 링크되어 겹쳤습니다.
즉, 두 영역이 정확히 동일 시작 주소라 충돌이 발생했습니다.

이를 해결하기 위해서는,
1. xv6처럼 -bios none으로 부팅하거나,
2. 또는 OpenSBI를 유지하되 커널 링크 주소를 0x80200000 등으로 옮깁니다.

## relocation truncated to fit: R_RISCV_HI20

```sh
❯ make qemu
riscv64-elf-gcc -march=rv64gc  -mabi=lp64  -Wall -O2  -ffreestanding  -nostdlib  -nostartfiles -c -o kernel/entry.o kernel/entry.S
riscv64-elf-gcc -march=rv64gc  -mabi=lp64  -Wall -O2  -ffreestanding  -nostdlib  -nostartfiles -c -o kernel/start.o kernel/start.c
riscv64-elf-gcc -march=rv64gc  -mabi=lp64  -Wall -O2  -ffreestanding  -nostdlib  -nostartfiles -c -o kernel/main.o kernel/main.c
riscv64-elf-ld -z max-page-size=4096 -T kernel/kernel.ld -o kernel/kernel.elf kernel/entry.o kernel/start.o kernel/main.o
kernel/start.o: in function `start':
start.c:(.text+0x2): relocation truncated to fit: R_RISCV_HI20 against symbol `edata' defined in .rodata section in kernel/kernel.elf
kernel/main.o: in function `main':
main.c:(.text.startup+0x0): relocation truncated to fit: R_RISCV_HI20 against `.LC0'
make: *** [kernel/kernel.elf] Error 1
```

레딧 답변을 보면 기본 [코드 모델](https://github.com/riscv-non-isa/riscv-elf-psabi-doc/blob/master/riscv-elf.adoc#code-models)이 [`medlow`라서 발생](https://www.reddit.com/r/RISCV/comments/1autmk1/r_riscv_hi20_error)합니다.

```sh
The default code model, and the only one that uses absolute addressing (which HI20 is for) is medlow, which requires your code to live in low memory (first ~2 GiB, and technically also the last ~2 GiB counts too), but you’re presumably linking at an address outside that range. You’ll need to either change your link address to be within that range or switch to the medany code model. See https://github.com/riscv-non-isa/riscv-elf-psabi-doc/blob/master/riscv-elf.adoc#code-models if you want the details.
```

링커가 HI20/LO12를 맞추는 과정에서 주소가 표현 범위를 넘으면
"relocation truncated to fit: R_RISCV_HI20"로 중단한다고 합니다.

기본 코드 모델이자 절대 주소 지정 방식(HI20이 사용하는 방식)을 사용하는 유일한 모델은 `medlow`이고, 이 모델을 사용하려면 코드가 낮은 메모리에 있어야 하는데, 그 범위를 벗어난 주소로 링크하기 때문에 발생합니다.

아래 명령은 `medlow` 코드 모델에서 어떻게 값을 로드하고, 값을 저장하고, 주소를 계산하는지 보여줍니다.

```asm
# Load value from a symbol
lui  a0, %hi(symbol)
lw   a0, %lo(symbol)(a0)

# Store value to a symbol
lui  a0, %hi(symbol)
sw   a1, %lo(symbol)(a0)

# Calculate address
lui  a0, %hi(symbol)
addi a0, a0, %lo(symbol)
```

qemu virt 보드에서 DRAM은 0x80000000에 매핑됩니다.
0x7FFF_FFFF(= 2 GiB – 1)보다 크기 때문에 `medlow` 범위를 벗어납니다.

> The ranges on RV64 are not `0x0` ~ `0x000000007FFFFFFF` and `0xFFFFFFFF80000000` ~ `0xFFFFFFFFFFFFFFFF` due to RISC-V’s sign-extension of immediates; the following code fragments show where the ranges come from:

```asm
# Largest postive number:
lui a0, 0x7ffff # a0 = 0x7ffff000
addi a0, 0x7ff # a0 = a0 + 2047 = 0x000000007FFFF7FF

# Smallest negative number:
lui a0, 0x80000 # a0 = 0xffffffff80000000
addi a0, a0, -0x800 # a0 = a0 + -2048 = 0xFFFFFFFF7FFFF800
```

```log
RV64 Address Space
- 0xFFFFFFFFFFFFFFFF (64비트 최대값)

+--------------------------------------------------------------+ 0xFFFFFFFFFFFFFFFF (max)
|                                                              |
| Last ~2 GiB range (negative addresses in sign-extended form) |
| medlow reachable: 0xFFFFFFFF7FFFF800 -> 0xFFFFFFFFFFFFFFFF   |
|                                                              |
+--------------------------------------------------------------+ 0xFFFFFFFF7FFFF800 (Smallest negative)
|                   (UNREACHABLE in medlow)                    |
|                         ...                                  |
|                                                              |
|                    QEMU virt DRAM: 0x0000000080000000        |
|                                                              |
|                         ...                                  |
|                                                              |
+--------------------------------------------------------------+ 0x000000007FFFF7FF (Largest postive)
| First ~2 GiB range (positive addresses)                      |
| medlow reachable: 0x0000000000000000 -> 0x000000007FFFF7FF   |
+--------------------------------------------------------------+ 0x0000000000000000
```

- 첫 2 GiB: 0x0000000000000000 ~ 0x000000007FFFF7FF => medlow 접근 가능
- 마지막 2 GiB: 0xFFFFFFFF7FFFF800 ~ 0xFFFFFFFFFFFFFFFF => medlow 접근 가능
- 중간: medlow 불가 (qemu virt의 DRAM 0x80000000가 바로 여기에 위치)

즉 `medlow`의 ±2 GiB 범위 밖에 있으니 `lui`/`addi`로는 바로 못만들고,
컴파일러가 HI20 relocation을 넣었다가 링커가 "truncated to fit" 에러를 냅니다.

이 문제를 해결하려면 `medany` 코드 모델로 변경하면 된다고 합니다.
`medany`는 PC-relative 주소 계산(`auipc` + `addi`)을 사용하여 주소를 만들기 때문에,
현재 PC로부터 ±2 GiB 범위 안의 어떤 위치든 접근 가능합니다.

> `AUIPC`는 현재 PC를 기준으로 상위 20비트를 더합니다.
> `ADDI`는 즉시값을 더한다.
> 둘을 이어서 64비트 주소를 PC 기준으로 만듭니다.
>
> 즉, PC-relative는 절대주소 대신 현재 명령의 주소(PC)를 기준으로 오프셋을 더해 목표 주소를 만듭니다.

## CFLAGS

```Makefile
CFLAGS = -march=rv64gc # 타겟 ISA: 64비트 RISC-V. 기본 ISA + 압축 명령어 + 원자적 연산 등 지원.
CFLAGS += -mabi=lp64 # ABI: 64비트 long, pointer
CFLAGS += -Wall -O2 # 경고 활성화, 최적화 레벨 2
CFLAGS += -mcmodel=medany # 코드 모델을 medany로 설정하여 PC-relative 주소 생성
CFLAGS += -ffreestanding # 표준 라이브러리 환경이 아니라 임베디드, 커널 환경이라는 표시
CFLAGS += -nostdlib # 표준 C 런타임(ctr0.o 등)과 라이브러리를 링크하지 않음
CFLAGS += -nostartfiles # crt0 등 시작파일 생략(드라이버 링크에 영향)
CFLAGS += -fno-builtin
```

- `-march=rv64gc`
    - CPU가 어떤 명령어 집합(ISA)을 이해하는지 지정하지 않으면, 컴파일러는 잘못된 명령어를 생성해 실행 시 illegal instruction 예외가 발생할 수 있습니다.
    - `rv64gc`
        - `rv64`: 64비트 주소 공간
        - `g`: 기본 확장 세트(IMAFD + Zicsr + Zifencei)
            - `IMAFD`: 정수(I), 곱셈/나눗셈(M), 원자적 연산(A), 단정도 부동소수점(F), 배정도 부동소수점(D)
            - Zicsr: CSR 접근
            - `Zifencei`: instruction fence.
        - `c`: compressed(압축) 명령어

        즉, 정수 연산, 부동소수점, 원자적 연산, CSR 접근, fence, 압축 인스트럭션까지 포함된 코드만 생성하라는 의미입니다.

- `-mabi=lp64`
    - ISA가 정해져도 함수 호출 규약(ABI)을 맞추지 않으면, 함수 인자와 반환값이 엉뚱한 레지스터나 스택에 실려 런타임 오류가 발생합니다.
    - `lp64`
        - long과 pointer가 64비트, int는 32비트. 리눅스 RISC-V에서 표준 ABI입니다.
        - 컴파일러는 이 규약에 따라 레지스터를 배치하고 스택 프레임을 만듭니다.

- `-Wall -O2`
    - 대부분의 경고를 활성화합니다.
    - 초기화되지 않은 변수, 사용되지 않는 값 등을 잡아냅니다.
    - `-O2`는 루프 언롤링, 상수 전파, 데드 코드 제거 등 주요 최적화를 적용합니다.

- `-mcmodel=medany`
    - RISC-V는 PC-relative 주소계산을 사용합니다.
    - 코드와 데이터가 2GB 이상 떨어진 위치에 있을 수 있으므로 코드 모델을 지정해야 합니다.
    - `medany`는 데이터/코드가 현재 PC에서 ±2GB 범위 어디든 있을 수 있다고 가정합니다.
    - 커널은 물리 메모리 상의 높은 주소에 로드되므로 이 설정이 필요합니다.
    - 기본 `medlow` 모델은 낮은 2GB 영역만 유효하므로, 고주소 영역 커널이 링크/실행되지 않습니다.

- `-ffreestanding`
    - `freestanding` 환경은 표준 C 환경(libc, `main`, `exit`)이 없음을 의미합니다. 따라서 커널은 스스로 초기화 코드를 제공해야 합니다.
    - 컴파일러는 표준 라이브러리 함수 존재를 가정하지 않고, 내장함수 최적화를 일부 비활성화합니다. 예컨대 `main` 함수가 없어도 에러를 내지 않습니다.

- `-nostdlib`
    - `freestanding`를 넘어서 아예 표준 라이브러리를 링크하지 않겠다는 옵션입니다.
    - `crt0.o`, `libc.a`, `libm.a` 등을 링크 단계에서 제외합니다. 커널은 자체 부트스트랩 코드를 제공해야 합니다.
    - 이 옵션이 없다면 리눅스 사용자 공간용 라이브러리가 커널에 끌려 들어와 링킹 충돌이 발생할 수 있습니다. (예: `_start` 심볼 중복)

- `-nostartfiles`
    - `crt0` 같은 시작 파일은 사용자 공간 프로그램 진입점(`_start`)을 초기화합니다.
    - 커널은 자체 어셈블리 진입 코드를 갖습니다.
    - GCC는 `crt0.o`, `crtbegin.o` 같은 시작 오브젝트 파일을 자동 포함하지 않습니다.
    - 이 옵션이 없다면 `_start` 심볼 중복 정의 충돌이 발생하여 링커가 어떤 진입점을 선택할지 알 수 없어 부트 실패합니다.

- `-fno-builtin`
    - 커널은 종종 표준 함수(`memcpy`, `strlen`)를 자체 구현합니다. 컴파일러가 이를 내부적으로 최적화해 다른 코드로 치환하면 위험합니다.
    - "표준 라이브러리 함수를 내장 함수로 간주하지 말라"고 지시합니다.
        - `memcpy` 호출이 그대로 남아 자체 구현 심볼에 연결됩니다.
        - 가령 `memcpy()`가 내부적으로 `__builtin_memcpy()`로 바뀌는 것을 방지합니다.

## [qemu 종료 방법](https://superuser.com/a/1211516)

```sh
Ctrl-A x
```

## 'a'가 아니라 'aaa'로 출력되는 이유

```c
void main(void) {
    uart_puts_sync('a');
}
```

분명 'a'가 한번만 출력되어야 하는데, 아래와 같이 세 번 'a'가 출력됩니다.

```sh
❯ make qemu
riscv64-elf-gcc -march=rv64gc  -mabi=lp64  -Wall  -mcmodel=medany  -ffreestanding  -nostartfiles  -c -o kernel/printf.o kernel/printf.c
riscv64-elf-ld -z max-page-size=4096  -T kernel/kernel.ld -o kernel/kernel.elf kernel/entry.o kernel/printf.o kernel/start.o kernel/main.o
qemu-system-riscv64 -machine virt  -bios none  -kernel kernel/kernel.elf  -m 128M -smp 3  -nographic
aaa
```

이는 `-smp 3`이라서 hart 3개가 같은 엔트리로 동시에 부팅했고,
각 hart가 `main()`을 실행해 `uart_puts_sync('a')`를 한 번씩 호출하기 때문입니다.

## DTB(Device Tree Blob)와 `dtc`(Device Tree Compiler)

RISC-V virt 보드는 전원을 넣자마자 어떤 레지스터가 어디에 있고 IRQ가 몇 번인지,
클럭 주파수가 얼마인지 CPU에게 일일이 말해 주지 않습니다.

대신 QEMU는 가상 SoC 전체를 한눈에 서술한 구조체(DTB, Device-Tree Blob)를 부팅 직전에 메모리에 만들어 둡니다.
커널이 DTB를 아직 파싱하기 전에 이 하드웨어 지도를 보고싶을 때 `dumpdtb=` 옵션을 사용합니다.

```Makefile
QEMU_DTB_OPTS = -machine virt,dumpdtb=virt.dtb
QEMU_DTB_OPTS += -bios none
QEMU_DTB_OPTS += -kernel $K/kernel.elf
QEMU_DTB_OPTS += -m 128M
QEMU_DTB_OPTS += -smp $(CPUS)
QEMU_DTB_OPTS += -nographic
```

DTB 파일은 바이너리인데 `dts`를 사용하면 사람이 읽을 수 있는 파일로 변환할 수 있습니다.

```sh
dtc -I dtb -O dts virt.dtb > virt.dts
```

<details>
<summary>vrit.dts</summary>

```arduino
/dts-v1/;

/ {
    #address-cells = <0x02>;
    #size-cells = <0x02>;
    compatible = "riscv-virtio";
    model = "riscv-virtio,qemu";

    poweroff {
        value = <0x5555>;
        offset = <0x00>;
        regmap = <0x08>;
        compatible = "syscon-poweroff";
    };

    reboot {
        value = <0x7777>;
        offset = <0x00>;
        regmap = <0x08>;
        compatible = "syscon-reboot";
    };

    platform-bus@4000000 {
        interrupt-parent = <0x07>;
        ranges = <0x00 0x00 0x4000000 0x2000000>;
        #address-cells = <0x01>;
        #size-cells = <0x01>;
        compatible = "qemu,platform", "simple-bus";
    };

    /*
     * memory@80000000의 부모는 루트이며 루트의 주소 셀 수와 크기 셀 수는 `<0x02>`입니다.
     * - #address-cells = <0x02>;
     * - #size-cells = <0x02>;
     *
     * 이는 주소가 반드시 2셀, 크기가 반드시 2셀을 사용해야 함을 의미합니다.
     * reg는 32비트 셀 4개 `[addr_hi, addr_lo, size_hi, size_lo]`가 됩니다.
     * 각 셀은 big-endian 32비트 단위 값입니다.
     */
    memory@80000000 {
        device_type = "memory";
        reg = <0x00 0x80000000 0x00 0x8000000>;
        /*     └─addr_hi │     └─size_hi │
         *       (32비트) │       (32비트) └─size_lo(8 * 16^6 = 2^27 = 134,217,728 의 hex값)
         *               │                 (32비트)
         *               └─addr_lo(8 * 16^7 = 2^31 = 2,147,483,648 의 hex값)
         *                 (32비트)
         *
         * 1 MiB = 1024 × 1024 = 2^20
         * 2^27 / 2^20 = 2^7 = 128 MiB
         */
    };

    cpus {
        #address-cells = <0x01>;
        #size-cells = <0x00>;
        timebase-frequency = <0x989680>;

        cpu@0 {
            phandle = <0x05>;
            device_type = "cpu";
            reg = <0x00>;
            status = "okay";
            compatible = "riscv";
            riscv,cbop-block-size = <0x40>;
            riscv,cboz-block-size = <0x40>;
            riscv,cbom-block-size = <0x40>;
            riscv,isa-extensions = "i", "m", "a", "f", "d", "c", "h", "zic64b", "zicbom", "zicbop", "zicboz", "ziccamoa", "ziccif", "zicclsm", "ziccrse", "zicntr", "zicsr", "zifencei", "zihintntl", "zihintpause", "zihpm", "zmmul", "za64rs", "zaamo", "zalrsc", "zawrs", "zfa", "zca", "zcd", "zba", "zbb", "zbc", "zbs", "shcounterenw", "shgatpa", "shtvala", "shvsatpa", "shvstvala", "shvstvecd", "ssccptr", "sscounterenw", "sstc", "sstvala", "sstvecd", "ssu64xl", "svadu", "svvptc";
            riscv,isa-base = "rv64i";
            riscv,isa = "rv64imafdch_zic64b_zicbom_zicbop_zicboz_ziccamoa_ziccif_zicclsm_ziccrse_zicntr_zicsr_zifencei_zihintntl_zihintpause_zihpm_zmmul_za64rs_zaamo_zalrsc_zawrs_zfa_zca_zcd_zba_zbb_zbc_zbs_shcounterenw_shgatpa_shtvala_shvsatpa_shvstvala_shvstvecd_ssccptr_sscounterenw_sstc_sstvala_sstvecd_ssu64xl_svadu_svvptc";
            mmu-type = "riscv,sv57";

            interrupt-controller {
                #interrupt-cells = <0x01>;
                interrupt-controller;
                compatible = "riscv,cpu-intc";
                phandle = <0x06>;
            };
        };

        cpu@1 {
            phandle = <0x03>;
            device_type = "cpu";
            reg = <0x01>;
            status = "okay";
            compatible = "riscv";
            riscv,cbop-block-size = <0x40>;
            riscv,cboz-block-size = <0x40>;
            riscv,cbom-block-size = <0x40>;
            riscv,isa-extensions = "i", "m", "a", "f", "d", "c", "h", "zic64b", "zicbom", "zicbop", "zicboz", "ziccamoa", "ziccif", "zicclsm", "ziccrse", "zicntr", "zicsr", "zifencei", "zihintntl", "zihintpause", "zihpm", "zmmul", "za64rs", "zaamo", "zalrsc", "zawrs", "zfa", "zca", "zcd", "zba", "zbb", "zbc", "zbs", "shcounterenw", "shgatpa", "shtvala", "shvsatpa", "shvstvala", "shvstvecd", "ssccptr", "sscounterenw", "sstc", "sstvala", "sstvecd", "ssu64xl", "svadu", "svvptc";
            riscv,isa-base = "rv64i";
            riscv,isa = "rv64imafdch_zic64b_zicbom_zicbop_zicboz_ziccamoa_ziccif_zicclsm_ziccrse_zicntr_zicsr_zifencei_zihintntl_zihintpause_zihpm_zmmul_za64rs_zaamo_zalrsc_zawrs_zfa_zca_zcd_zba_zbb_zbc_zbs_shcounterenw_shgatpa_shtvala_shvsatpa_shvstvala_shvstvecd_ssccptr_sscounterenw_sstc_sstvala_sstvecd_ssu64xl_svadu_svvptc";
            mmu-type = "riscv,sv57";

            interrupt-controller {
                #interrupt-cells = <0x01>;
                interrupt-controller;
                compatible = "riscv,cpu-intc";
                phandle = <0x04>;
            };
        };

        cpu@2 {
            phandle = <0x01>;
            device_type = "cpu";
            reg = <0x02>;
            status = "okay";
            compatible = "riscv";
            riscv,cbop-block-size = <0x40>;
            riscv,cboz-block-size = <0x40>;
            riscv,cbom-block-size = <0x40>;
            riscv,isa-extensions = "i", "m", "a", "f", "d", "c", "h", "zic64b", "zicbom", "zicbop", "zicboz", "ziccamoa", "ziccif", "zicclsm", "ziccrse", "zicntr", "zicsr", "zifencei", "zihintntl", "zihintpause", "zihpm", "zmmul", "za64rs", "zaamo", "zalrsc", "zawrs", "zfa", "zca", "zcd", "zba", "zbb", "zbc", "zbs", "shcounterenw", "shgatpa", "shtvala", "shvsatpa", "shvstvala", "shvstvecd", "ssccptr", "sscounterenw", "sstc", "sstvala", "sstvecd", "ssu64xl", "svadu", "svvptc";
            riscv,isa-base = "rv64i";
            riscv,isa = "rv64imafdch_zic64b_zicbom_zicbop_zicboz_ziccamoa_ziccif_zicclsm_ziccrse_zicntr_zicsr_zifencei_zihintntl_zihintpause_zihpm_zmmul_za64rs_zaamo_zalrsc_zawrs_zfa_zca_zcd_zba_zbb_zbc_zbs_shcounterenw_shgatpa_shtvala_shvsatpa_shvstvala_shvstvecd_ssccptr_sscounterenw_sstc_sstvala_sstvecd_ssu64xl_svadu_svvptc";
            mmu-type = "riscv,sv57";

            interrupt-controller {
                #interrupt-cells = <0x01>;
                interrupt-controller;
                compatible = "riscv,cpu-intc";
                phandle = <0x02>;
            };
        };

        cpu-map {

            cluster0 {

                core0 {
                    cpu = <0x05>;
                };

                core1 {
                    cpu = <0x03>;
                };

                core2 {
                    cpu = <0x01>;
                };
            };
        };
    };

    pmu {
        riscv,event-to-mhpmcounters = <0x01 0x01 0x7fff9 0x02 0x02 0x7fffc 0x10019 0x10019 0x7fff8 0x1001b 0x1001b 0x7fff8 0x10021 0x10021 0x7fff8>;
        compatible = "riscv,pmu";
    };

    fw-cfg@10100000 {
        dma-coherent;
        reg = <0x00 0x10100000 0x00 0x18>;
        compatible = "qemu,fw-cfg-mmio";
    };

    flash@20000000 {
        bank-width = <0x04>;
        reg = <0x00 0x20000000 0x00 0x2000000 0x00 0x22000000 0x00 0x2000000>;
        compatible = "cfi-flash";
    };

    aliases {
        serial0 = "/soc/serial@10000000";
    };

    /* 루트 노드와 고정 포인터 */
    chosen {
        /* 콘솔 장치 경로. soc 블록의 serial@10000000에 있다는 의미입니다. */
        stdout-path = "/soc/serial@10000000";
        rng-seed = <0xb6d3a50e 0xc5fc791b 0x6ba6be2c 0x4c8c65aa 0x73755381 0x97c947d6 0xaf2277e9 0x58c82dfa>;
    };

    soc {
        #address-cells = <0x02>;
        #size-cells = <0x02>;
        compatible = "simple-bus";
        ranges;

        rtc@101000 {
            interrupts = <0x0b>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x101000 0x00 0x1000>;
            compatible = "google,goldfish-rtc";
        };

        /*
         * UART 장치 노드를 해독. 베이스 주소가 0x10000000 임을 의미합니다.
         *
         * 부모는 soc라서 마찬가지로 주소 셀 수(<0x02>), 크기 셀 수(<0x02>)를 따릅니다
         */
        serial@10000000 {
            /* <10>. PLIC에서 UART가 쓰는 인터럽트 번호 */
            interrupts = <0x0a>;

            /*
             * PLIC 노드의 phandle 값입니다.
             * - `interrupt-parent = <0x07>;`: 인터럽트 라우팅은 phandle=7번(PIC=PLIC)으로 간다는 의미
             * - #interrupt-cells = <0x01>;`: PLIC 안에서 소스 번호 10번을 배정받았다는 의미
             *
             * 커널은 이걸 읽고 PLIC의 소스 10을 UART 드라이버에 연결하도록 세팅합니다.
             */
            interrupt-parent = <0x07>;

            /* UART가 사용할 클럭 주파수 */
            clock-frequency = "", "8@";

            /* MMIO 영역 주소와 크기 */
            reg = <0x00 0x10000000 0x00 0x100>;
            /*     └─addr_hi │     └─size_hi │
             *       (32비트) │       (32비트) └─size_lo(1 * 16^2 = 2^8 = 256 바이트)
             *               │                 (32비트)
             *               └─addr_lo
             *                 (32비트)
             *
             * UART 같은 메모리 매핑 장치는 내부 레지스터가 수십 개뿐입니다.
             * 예를 들어 16550A UART는 송신 버퍼, 수신 버퍼, 상태 레지스터, 제어 레지스터 등 합쳐야 수십 바이트 크기입니다.
             * 따라서 레지스터 뭉치 크기를 256바이트 정도로 잡습니다.
             * 이 범위 안에서만 접근해야 정상 동작하고,
             * 커널은 이 정보를 그대로 매핑해 장치 드라이버가 MMIO를 할 수 있게 합니다.
             */

            /* 어떤 드라이버가 이 장치를 다룰 수 있는지 알려줍니다. 표준 16550 드라이버가 붙습니다. */
            compatible = "ns16550a";
        };

        test@100000 {
            phandle = <0x08>;
            reg = <0x00 0x100000 0x00 0x1000>;
            compatible = "sifive,test1", "sifive,test0", "syscon";
        };

        virtio_mmio@10008000 {
            interrupts = <0x08>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10008000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10007000 {
            interrupts = <0x07>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10007000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10006000 {
            interrupts = <0x06>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10006000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10005000 {
            interrupts = <0x05>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10005000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10004000 {
            interrupts = <0x04>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10004000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10003000 {
            interrupts = <0x03>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10003000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10002000 {
            interrupts = <0x02>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10002000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        virtio_mmio@10001000 {
            interrupts = <0x01>;
            interrupt-parent = <0x07>;
            reg = <0x00 0x10001000 0x00 0x1000>;
            compatible = "virtio,mmio";
        };

        plic@c000000 {
            phandle = <0x07>;
            riscv,ndev = <0x5f>;
            reg = <0x00 0xc000000 0x00 0x600000>;
            interrupts-extended = <0x06 0x0b 0x06 0x09 0x04 0x0b 0x04 0x09 0x02 0x0b 0x02 0x09>;
            interrupt-controller;
            compatible = "sifive,plic-1.0.0", "riscv,plic0";
            #address-cells = <0x00>;
            #interrupt-cells = <0x01>;
        };

        clint@2000000 {
            interrupts-extended = <0x06 0x03 0x06 0x07 0x04 0x03 0x04 0x07 0x02 0x03 0x02 0x07>;
            reg = <0x00 0x2000000 0x00 0x10000>;
            compatible = "sifive,clint0", "riscv,clint0";
        };

        pci@30000000 {
            interrupt-map-mask = <0x1800 0x00 0x00 0x07>;
            interrupt-map = <0x00 0x00 0x00 0x01 0x07 0x20 0x00 0x00 0x00 0x02 0x07 0x21 0x00 0x00 0x00 0x03 0x07 0x22 0x00 0x00 0x00 0x04 0x07 0x23 0x800 0x00 0x00 0x01 0x07 0x21 0x800 0x00 0x00 0x02 0x07 0x22 0x800 0x00 0x00 0x03 0x07 0x23 0x800 0x00 0x00 0x04 0x07 0x20 0x1000 0x00 0x00 0x01 0x07 0x22 0x1000 0x00 0x00 0x02 0x07 0x23 0x1000 0x00 0x00 0x03 0x07 0x20 0x1000 0x00 0x00 0x04 0x07 0x21 0x1800 0x00 0x00 0x01 0x07 0x23 0x1800 0x00 0x00 0x02 0x07 0x20 0x1800 0x00 0x00 0x03 0x07 0x21 0x1800 0x00 0x00 0x04 0x07 0x22>;
            ranges = <0x1000000 0x00 0x00 0x00 0x3000000 0x00 0x10000 0x2000000 0x00 0x40000000 0x00 0x40000000 0x00 0x40000000 0x3000000 0x04 0x00 0x04 0x00 0x04 0x00>;
            reg = <0x00 0x30000000 0x00 0x10000000>;
            dma-coherent;
            bus-range = <0x00 0xff>;
            linux,pci-domain = <0x00>;
            device_type = "pci";
            compatible = "pci-host-ecam-generic";
            #size-cells = <0x02>;
            #interrupt-cells = <0x01>;
            #address-cells = <0x03>;
        };
    };
};
```

</details>

커널은 DTB에서 `serial@10000000`, `plic@c000000`처럼 그 안에 적힌 계층적 노드/프로퍼티를 파싱합니다.
그리고 그 파싱된 결과로 CPU, 메모리, 장치가 어느 주소에 있고 어떤 인터럽트를 쓰는가를 결정합니다.

문제는 장치마다 주소 폭(address width) 이 다르다는 겁니다.
어떤 버스는 32비트 주소만 쓰고, 어떤 버스는 64비트 주소까지 씁니다.
만약 통일된 규칙 없이 숫자 하나만 써버리면, 0x10000000이 "32비트 주소 0x10000000"인지, "상위 32비트=0x1, 하위 32비트=0x0000000"인지 구분할 방법이 없어집니다.

그래서 디바이스 트리는 각 부모 버스가 "내 자식들의 주소와 크기를 몇 개의 32비트 셀(cell)로 표현할 것인가"를 명시합니다.
- `#address-cells = <0x02>;`: 주소를 표현할 때 32비트 셀 2개(총 64비트 폭)를 사용
- `#size-cells = <0x02>;`: 크기(메모리 범위)도 2셀(64비트 폭)로 사용

이렇게 정하면 자식 노드의 `reg` 속성은 `[addr_hi, addr_lo, size_hi, size_lo]` 네 개의 셀로 고정됩니다.
