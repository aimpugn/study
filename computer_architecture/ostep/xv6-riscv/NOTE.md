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
    - [GDB로 디버깅](#gdb로-디버깅)

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
The default code model, and the only one that uses absolute addressing (which HI20 is for) is medlow, which requires your code to live in low memory (first ~2 GiB, and technically also the last ~2 GiB counts too), but you're presumably linking at an address outside that range. You'll need to either change your link address to be within that range or switch to the medany code model. See https://github.com/riscv-non-isa/riscv-elf-psabi-doc/blob/master/riscv-elf.adoc#code-models if you want the details.
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

> The ranges on RV64 are not `0x0` ~ `0x000000007FFFFFFF` and `0xFFFFFFFF80000000` ~ `0xFFFFFFFFFFFFFFFF` due to RISC-V's sign-extension of immediates; the following code fragments show where the ranges come from:

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

## GDB로 디버깅

- `make qemu CPUS=1 QEMU_OPTS+=" -S -s"` 실행하여 완전히 멈춘 상태로 둡니다

    ```sh
    ❯ make qemu CPUS=1 QEMU_OPTS+=" -S -s"
    riscv64-elf-gcc -march=rv64gc  -mabi=lp64  -Wall  -mcmodel=medany  -ffreestanding  -nostartfiles  -c -o kernel/printf.o kernel/printf.c
    riscv64-elf-ld -z max-page-size=4096  -T kernel/kernel.ld -o kernel/kernel.elf kernel/entry.o kernel/printf.o kernel/start.o kernel/main.o
    qemu-system-riscv64 -S -s
    ```

    - `-S`: CPU를 리셋 직후, 첫 명령을 실행하기 전 상태에 멈춰 둡니다.
    - `-s`: `-gdb tcp::1234`와 동일합니다. QEMU가 1234 포트에서 GDB 접속을 대기합니다.
    - `CPUS=1`: 관찰하기 쉽도록 하나의 hart만 켭니다.

- 별도의 터미널에서 `riscv64-elf-gdb kernel/kernel.elf` 실행

    ```sh
    ❯ riscv64-elf-gdb kernel/kernel.elf
    GNU gdb (GDB) 16.3
    Copyright (C) 2024 Free Software Foundation, Inc.
    License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
    This is free software: you are free to change and redistribute it.
    There is NO WARRANTY, to the extent permitted by law.
    Type "show copying" and "show warranty" for details.
    This GDB was configured as "--host=aarch64-apple-darwin24.4.0 --target=riscv64-elf".
    Type "show configuration" for configuration details.
    For bug reporting instructions, please see:
    <https://www.gnu.org/software/gdb/bugs/>.
    Find the GDB manual and other documentation resources online at:
        <http://www.gnu.org/software/gdb/documentation/>.

    For help, type "help".
    Type "apropos word" to search for commands related to "word"...
    Reading symbols from kernel/kernel.elf...
    (No debugging symbols found in kernel/kernel.elf)
    (gdb) set pagination off
    (gdb) target remote :1234
    Remote debugging using :1234
    0x00000000800004f0 in ?? ()
    ```

```sh
(gdb) monitor xp /1bx 0x10000005
```

- `xp`는 "physical address를 x (hex) 형태로 출력"하는 QEMU 모니터 명령입니다.
- `/1b`는 1바이트를 16진수로 보겠다는 뜻입니다.
- 결과가 0x20 또는 0x60처럼 나오면 THR Empty 비트가 1이라는 의미입니다.

  make qemu CPUS=1 QEMU_OPTS+=" -S -s -monitor telnet:127.0.0.1:4444,server,nowait"

```sh
❯ make qemu-gdb CPUS=1
qemu-system-riscv64 -machine virt  -bios none  -kernel kernel/kernel.elf  -m 128M -smp 1  -nographic  -S -s -monitor telnet:127.0.0.1:4444,server,nowait
```

- 기존 `-S -s`로 CPU를 정지시키고 gdbstub를 열어 둔 상태 그대로 유지하면서,
- 추가로 QEMU 모니터를 127.0.0.1:4444 텔넷 포트에 붙여 놓습니다. 다른 포트를 사용할 수도 있습니다.

    `telnet 127.0.0.1 4444` 또는 `nc 127.0.0.1 4444` 등으로 접속하면
    모니터 프롬프트(`(qemu)`)를 바로 받을 수 있습니다.

- UNIX 소켓을 사용할 수도 있습니다: `-monitor unix:/tmp/qemu-monitor,server,nowait`

1. `nc 127.0.0.1 4444`: QEMU 모니터(HMP, 텔넷으로 붙는 창)

    QEMU 자체(에뮬레이터)가 대상입니다. 게스트 OS가 아니라 '에뮬레이터 내부 상태와 장치'를 다룹니다.

    QEMU가 에뮬레이터 자신에게 요청하여 장치의 물리 메모리(MMIO)를 읽어 주기 때문에 MMIO가 잘 보입니다.
    `xp /1bx 0x10000005`로 UART LSR 같은 장치 레지스터를 바로 볼 수 있습니다.

    ```sh
    (gdb) xp /1bx 0x10000005
    ```

    그래서 장치 상태 관찰에 사용합니다.
    `xp`(물리 메모리 읽기), `info mtree`(장치 맵), `help` 등의 명령어가 가능합니다.

2. `riscv64-elf-gdb`: QEMU gdbstub(원격 GDB)

    GDB는 게스트 CPU/프로세스가 대상으로, QEMU `gdbstub`에 붙어 "게스트 CPU/코드"를 조종합니다.
    한 줄씩 실행하고 함수 경계에서 멈추고, 레지스터 전달값을 확인하는 데 최적입니다.
    명령 단위 스텝, 브레이크포인트, 함수 호출 등 '코드 디버깅이 목적'입니다.

    반대로, `gdbstub`은 보통 MMIO를 직접 읽는 `x` 패킷을 제한하므로 GDB에서 `x/...`로 MMIO를 읽으려 하면 실패합니다.

    `run`/`step`/`break`/`call`(함수 호출), 게스트 RAM 읽기/쓰기(`x`, `set`) 등이 가능합니다.
    - GDB의 `x`는 "게스트 메모리"를 읽는 명령이고, `gdbstub`은 보통 RAM만 허용합니다.
    - 그래서 장치 MMIO 는 gdbstub에서 거부해 “Cannot access memory”가 뜹니다.

    심볼을 써서 함수/변수에 접근할 수 있습니다.

`xp`는 물리 주소를 덤프하는 명령이고, `/1bx`는 1바이트를 16진수로 보겠다는 의미입니다.

```sh
# LSR
(qemu) xp /1bx 0x10000005
xp /1bx 0x10000005
0000000010000005: 0x60
```

`LSR`이 0x10000005에서 0x60이 읽힌다는 것은 레지스터가 1바이트 간격(reg‑shift=0)이라는 의미입니다.
그리고 `THR Empty`(0x20) + `Transmitter Empty`(0x40)로 "지금 당장 써도 된다"는 뜻입니다.
이제 실제로 `THR`(0x10000000)에 `"a"`(0x61)가 기록되는 것을 확인할 수 있습니다.

만약 0x10000014는 4바이트 간격(reg‑shift=2)을 가정할 경우 "Cannot access memory"라고 나와야 정상입니다.

```sh
(qemu) xp /1bx 0x10000014
xp /1bx 0x10000014
0000000010000014: Cannot access memory
```

"a"가 출력되는 과정을 디버깅해보기 위해 다음과 같이 HMP와 gdbstub에 붙어봅니다.

HMP에 붙는 과정은 다음과 같습니다.

```sh
telnetl 127.0.0.1 4444
```

그리고 `info mtree`로 'serial@10000000'이 '0x10000000..0x100000ff'로 매핑되어 있는지 확인합니다.

```sh
address-space: I/O
  0000000000000000-000000000000ffff (prio 0, i/o): io

address-space: cpu-memory-0
address-space: memory
  # memory 아래가 게스트 물리 주소 공간의 장치/메모리 배치입니다. 괄호의 태그는 유형입니다.
  # i/o: 메모리 맵드 I/O(MMIO)
  0000000000000000-ffffffffffffffff (prio 0, i/o): system
    # rom/romd: 읽기 전용(펌웨어/플래시 등)
    0000000000001000-000000000000ffff (prio 0, rom): riscv_virt_board.mrom
    0000000000100000-0000000000100fff (prio 0, i/o): riscv.sifive.test
    0000000000101000-0000000000101023 (prio 0, i/o): goldfish_rtc
    0000000002000000-0000000002003fff (prio 0, i/o): riscv.aclint.swi
    0000000002004000-000000000200bfff (prio 0, i/o): riscv.aclint.mtimer
    0000000003000000-000000000300ffff (prio 0, i/o): gpex_ioport_window
      0000000003000000-000000000300ffff (prio 0, i/o): gpex_ioport
    0000000004000000-0000000005ffffff (prio 0, i/o): platform bus
    000000000c000000-000000000c5fffff (prio 0, i/o): riscv.sifive.plic
    # 0x10000000–0x10000007
    # - 길이가 정확히 8바이트이므로, ns16550a 레지스터가 1바이트 간격(reg‑shift=0)으로 8개(오프셋 0..7) 배치됐음을 뜻합니다.
    # - 따라서 주소는 "BASE + 오프셋" 그대로입니다.
    # - THR=0x10000000, IER=0x10000001, FCR/IIR=0x10000002, LCR=0x10000003, LSR=0x10000005
    0000000010000000-0000000010000007 (prio 0, i/o): serial
    0000000010001000-00000000100011ff (prio 0, i/o): virtio-mmio
    0000000010002000-00000000100021ff (prio 0, i/o): virtio-mmio
    0000000010003000-00000000100031ff (prio 0, i/o): virtio-mmio
    0000000010004000-00000000100041ff (prio 0, i/o): virtio-mmio
    0000000010005000-00000000100051ff (prio 0, i/o): virtio-mmio
    0000000010006000-00000000100061ff (prio 0, i/o): virtio-mmio
    0000000010007000-00000000100071ff (prio 0, i/o): virtio-mmio
    0000000010008000-00000000100081ff (prio 0, i/o): virtio-mmio
    0000000010100000-0000000010100007 (prio 0, i/o): fwcfg.data
    0000000010100008-0000000010100009 (prio 0, i/o): fwcfg.ctl
    0000000010100010-0000000010100017 (prio 0, i/o): fwcfg.dma
    0000000020000000-0000000021ffffff (prio 0, romd): virt.flash0
    0000000022000000-0000000023ffffff (prio 0, romd): virt.flash1
    0000000030000000-000000003fffffff (prio 0, i/o): alias pcie-ecam @pcie-mmcfg-mmio 0000000000000000-000000000fffffff
    0000000040000000-000000007fffffff (prio 0, i/o): alias pcie-mmio @gpex_mmio_window 0000000040000000-000000007fffffff
    # ram: 실제 RAM
    # - 0x80000000–0x87ffffff (128 MiB)
    # - 커널 링크 주소(0x8000_0000)와 일치합니다.
    0000000080000000-0000000087ffffff (prio 0, ram): riscv_virt_board.ram
    0000000400000000-00000007ffffffff (prio 0, i/o): alias pcie-mmio-high @gpex_mmio_window 0000000400000000-00000007ffffffff

address-space: gpex-root
  0000000000000000-ffffffffffffffff (prio 0, i/o): bus master container

memory-region: pcie-mmcfg-mmio
  0000000000000000-000000000fffffff (prio 0, i/o): pcie-mmcfg-mmio

memory-region: gpex_mmio_window
  0000000000000000-ffffffffffffffff (prio 0, i/o): gpex_mmio_window
    0000000000000000-ffffffffffffffff (prio 0, i/o): gpex_mmio

memory-region: system
  0000000000000000-ffffffffffffffff (prio 0, i/o): system
    0000000000001000-000000000000ffff (prio 0, rom): riscv_virt_board.mrom
    0000000000100000-0000000000100fff (prio 0, i/o): riscv.sifive.test
    0000000000101000-0000000000101023 (prio 0, i/o): goldfish_rtc
    0000000002000000-0000000002003fff (prio 0, i/o): riscv.aclint.swi
    0000000002004000-000000000200bfff (prio 0, i/o): riscv.aclint.mtimer
    0000000003000000-000000000300ffff (prio 0, i/o): gpex_ioport_window
      0000000003000000-000000000300ffff (prio 0, i/o): gpex_ioport
    0000000004000000-0000000005ffffff (prio 0, i/o): platform bus
    000000000c000000-000000000c5fffff (prio 0, i/o): riscv.sifive.plic
    0000000010000000-0000000010000007 (prio 0, i/o): serial
    0000000010001000-00000000100011ff (prio 0, i/o): virtio-mmio
    0000000010002000-00000000100021ff (prio 0, i/o): virtio-mmio
    0000000010003000-00000000100031ff (prio 0, i/o): virtio-mmio
    0000000010004000-00000000100041ff (prio 0, i/o): virtio-mmio
    0000000010005000-00000000100051ff (prio 0, i/o): virtio-mmio
    0000000010006000-00000000100061ff (prio 0, i/o): virtio-mmio
    0000000010007000-00000000100071ff (prio 0, i/o): virtio-mmio
    0000000010008000-00000000100081ff (prio 0, i/o): virtio-mmio
    0000000010100000-0000000010100007 (prio 0, i/o): fwcfg.data
    0000000010100008-0000000010100009 (prio 0, i/o): fwcfg.ctl
    0000000010100010-0000000010100017 (prio 0, i/o): fwcfg.dma
    0000000020000000-0000000021ffffff (prio 0, romd): virt.flash0
    0000000022000000-0000000023ffffff (prio 0, romd): virt.flash1
    0000000030000000-000000003fffffff (prio 0, i/o): alias pcie-ecam @pcie-mmcfg-mmio 0000000000000000-000000000fffffff
    0000000040000000-000000007fffffff (prio 0, i/o): alias pcie-mmio @gpex_mmio_window 0000000040000000-000000007fffffff
    0000000080000000-0000000087ffffff (prio 0, ram): riscv_virt_board.ram
    0000000400000000-00000007ffffffff (prio 0, i/o): alias pcie-mmio-high @gpex_mmio_window 0000000400000000-00000007ffffffff
```

이를 통해 virt 보드 메모리 맵이 정상적으로 잡혔음을 나타냅니다.
또한 serial 범위가 0x10000000–0x10000007로 짧게 잡힌 덕분에 reg‑shift가 0임을 확정할 수 있습니다.

- `xp /1bx 0x10000005`: LSR이 '0x20' 또는 '0x60'이면 송신 가능 상태입니다.
- `xp /1bx 0x10000003`: LCR 최종값이 '0x03'(`8N1`)인지 반드시 확인합니다.
    만약 '0x83'이라면 `DLAB=1`이 켜진 상태라 `THR`에 쓰는 바이트가 `DLL`로 샙니다.
    이는 문자 미출력의 대표 원인입니다.

    ```sh
    (qemu) xp /1bx 0x10000003
    0000000010000003: 0x00
    ```

    이는 `LCR`(Line Control Register)이 리셋 직후의 기본값 그대로라는 뜻입니다.

    ns16550A의 `LCR` 비트는 다음과 같이 정의됩니다:
    - 하위 2비트는 데이터 비트 수(00=5비트, 01=6비트, 10=7비트, 11=8비트),
    - 그다음은 스톱 비트(0=1스톱),
    - 패리티 사용/종류,
    - 브레이크,
    - 최상위 비트는 DLAB(Divisor Latch Access Bit)입니다.

    따라서 `LCR=0x00`은 "5비트 데이터, 1스톱, 패리티 없음, 브레이크 꺼짐, DLAB=0" 상태를 의미합니다

    여기서 `0x00`인 이유는 아직 커널 초기화가 실행되지 않았기 때문입니다.
    `-S -s`로 QEMU를 띄우면 CPU는 리셋 직후(MROM 0x1000 부근)에서 멈춰있고,
    커널의 '_entry => start => main => uart_init' 흐름이 실행되기 전입니다.

    ns16550A의 `LSR`은 리셋 직후에도 `THRE`/`TEMT`(0x60)가 1로 읽히는반면,
    `LCR`은 '0x00'이 기본값이므로 `LSR=0x60`, `LCR=0x00` 조합은 아직 `LCR`을 세팅하지 않았다는 의미이며,
    `uart_init`이 돌지 않았다는 의미입니다.

gdbstub에 붙는 과정은 다음과 같습니다.

```sh
❯ e
GNU gdb (GDB) 16.3
Copyright (C) 2024 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.
Type "show copying" and "show warranty" for details.
This GDB was configured as "--host=aarch64-apple-darwin24.4.0 --target=riscv64-elf".
Type "show configuration" for configuration details.
For bug reporting instructions, please see:
<https://www.gnu.org/software/gdb/bugs/>.
Find the GDB manual and other documentation resources online at:
    <http://www.gnu.org/software/gdb/documentation/>.

For help, type "help".
Type "apropos word" to search for commands related to "word"...
Reading symbols from kernel/kernel.elf...
(No debugging symbols found in kernel/kernel.elf)
(gdb) break uart_init
Breakpoint 1 at 0x800000a4
(gdb) break uart_puts_sync
Breakpoint 2 at 0x800000e6
(gdb) continue
The program is not being run.
```

"The program is not being run."은 커널이 멈춰 있어서가 아니라, GDB가 "아직 원격 대상(QEMU gdbstub)에 붙어 있지 않다"는 뜻입니다.
같은 세션에서 브레이크포인트는 미리(오프라인으로) 걸 수 있지만, 연결하지 않은 상태에서 `continue`를 입력했기 때문입니다.

따라서 먼저 "붙는다 -> 멈춘 지점 확인 -> 실행을 진행한다" 순서를 확실히 밟아야 합니다:
- QEMU를 `-S -s`로 띄우면 CPU는 "리셋 직후 정지 상태"입니다. 이 상태는 프로그램이 종료된 게 아니라, 아직 한 줄도 실행하지 않은 초기 정지 상태입니다.
- HMP(텔넷 모니터)는 QEMU 자체와 대화합니다. 그래서 `xp`로 MMIO를 읽을 수 있지만, 코드 흐름(`continue`/`step`/`break`)은 GDB가 맡아야 합니다.
- GDB는 QEMU의 `gdbstub`와 "TCP로 붙어야만" 원격 CPU를 제어할 수 있습니다. 붙지 않은 상태에서 `continue`를 입력하면 "The program is not being run."라고 나옵니다.

GDB가 QEMU에 TCP로 붙기 위해서 `target remote :1234`를 입력합니다.

```sh
(gdb) target remote :1234
Remote debugging using :1234
0x0000000000001000 in ?? ()
```

그리고 타겟에 대한 정보를 출력해 봅니다.

```sh
(gdb) info target
Symbols from "/Users/rody/VscodeProjects/study/computer_architecture/ostep/xv6-riscv/kernel/kernel.elf".
Remote target using gdb-specific protocol:
    `/Users/rody/VscodeProjects/study/computer_architecture/ostep/xv6-riscv/kernel/kernel.elf', file type elf64-littleriscv.
    Entry point: 0x80000000
    0x0000000080000000 - 0x0000000080000174 is .text
    0x0000000080001000 - 0x000000008000b000 is .bss
    While running this, GDB does not access memory from...
Local exec file:
    `/Users/rody/VscodeProjects/study/computer_architecture/ostep/xv6-riscv/kernel/kernel.elf', file type elf64-littleriscv.
    Entry point: 0x80000000
    0x0000000080000000 - 0x0000000080000174 is .text
    0x0000000080001000 - 0x000000008000b000 is .bss
```

`x/i $pc`를 입력하면  대개 0x0000000000001000 부근(MROM)의 명령 한 줄이 보입니다.
이 시점은 아직 커널 코드로 진입 전입니다.

```sh
(gdb) x/i $pc
=> 0x1000:    auipc    t0,0x0
```

현재 걸린 중단점을 확인할 수도 있습니다.

```sh
(gdb) info break
Num     Type           Disp Enb Address            What
1       breakpoint     keep y   0x00000000800000a4 <uart_init+12>
2       breakpoint     keep y   0x00000000800000e6 <uart_puts_sync+10>
```

`continue`를 입력하여 진행할 수 있습니다.

```sh
(gdb) continue
Continuing.
```

만약 아무 반응이 없으면 Ctrl‑C로 중단해서 PC를 읽고(아래), 브레이크가 올바르게 풀렸는지 재확인합니다.
아래의 경우 Ctrl‑C로 중단시키는 순간의 PC가 `start()` 안에 있었다는 의미입니다.
이 커널의 [start.c](./kernel/start.c)를 보면 `main()`을 호출한 뒤 `for(;;) wfi;`로 "무한 대기(인터럽트 대기) 루프"에 들어갑니다. `-bios none` 환경에서 따로 인터럽트를 발생시키지 않으면 CPU는 그 wfi 루프에서 영원히 멈춰 보입니다.

그래서 Ctrl‑C로 끊으면 지금 실행 중인 위치가 `start()`로 잡힙니다.

```sh
^C
Program received signal SIGINT, Interrupt.
0x000000008000012e in start ()
```

근데 위에서 Continuing인 동안 0x10000005 값을 확인해보면 0x61로 변경이 되어 있습니다.

```sh
(qemu) xp /1bx 0x10000003
0000000010000003: 0x00
(qemu) xp /1bx 0x10000005
0000000010000005: 0x61
```

LSR 비트 구성(16550A):
- bit0=0x01: Data Ready (수신 버퍼 RBR에 바이트가 도착)
- bit5=0x20: THRE (송신 홀딩 레지스터 비움)
- bit6=0x40: TEMT (송신 시프트 레지스터까지 비움)

0x61 = 0b 0110 0001 = TEMT(1) + THRE(1) + DR(1)

이는 수신 경로에 문자가 들어왔다는 뜻입니다.
숫자 0x61이 우연히 ASCII ‘a’와 같지만, 이 값은 "상태 레지스터의 비트 패턴"이지 ‘a’를 의미하는 게 아닙니다.

또한 "송신이 실행되었다"는 증거는 아닙니다.
`LSR`의 `THRE`/`TEMT`이 1인 건 "언제든 THR에 써도 된다"는 의미이고, `DR=1`은 입력 바이트가 준비됐다는 의미입니다.

```sh
# HMP에서 확인
(qemu) xp /1bx 0x10000000
0000000010000000: 0x78
```

수신 경로가 살아 있어 `DR=1`로 바뀐 것이고, 송신 초기화(`uart_init`)가 아직 `LCR`을 0x03으로 바꾸지 못한 상태입니다.
즉, a가 출력되었다는 의미가 아닙니다.
'a' 출력은 "송신 경로에서 `THR`(0x10000000)에 0x61을 쓰는 순간"이 있어야 합니다.

중단점을 지우고 다시 시도해 봅니다.
정확한 지점에 중단점을 설정하기 위해 심볼 주소를 확인합니다.

`-g` 없이 최적화/정렬 변화가 있으면 GDB가 "b main"을 애매하게 해석할 수 있다고 합니다.
그래서 `nm`으로 확인한 정확한 진입 주소에 걸어주는 게 확실하다고 합니다.

```sh
❯ riscv64-elf-nm -n kernel/kernel.elf | rg ' (main|uart_init|uart_puts_sync)$'
0000000080000098 T uart_init
00000000800000dc T uart_puts_sync
0000000080000156 T main
```

또는 `gdbstub`에서 `info address`를 사용합니다.

```sh
(gdb) info address main
Symbol "main" is at 0x80000156 in a file compiled without debugging.
(gdb) info address uart_init
Symbol "uart_init" is at 0x80000098 in a file compiled without debugging.
(gdb) info address uart_puts_sync
Symbol "uart_puts_sync" is at 0x800000dc in a file compiled without debugging.
```

그리고 해당 주소에 대해 임시 중단점을 설정합니다.

```sh
(gdb) tbreak *0x80000156
Temporary breakpoint 1 at 0x80000156
(gdb) tbreak *0x80000098
Temporary breakpoint 2 at 0x80000098
(gdb) tbreak *0x800000dc
Temporary breakpoint 3 at 0x800000dc
```

그리고 HMP에서 `system_reset`를 입력하고, 다시 `gdbstub`에서 `continue`를 입력합니다.

```sh
(gdb) continue
Continuing.


^C
Program received signal SIGINT, Interrupt.
0x0000000080000148 in start ()
(gdb) x/i $pc
=> 0x80000148 <start+52>:    bltu    a4,a5,0x8000012a <start+22>
```

`0x80000148 <start+52>`는 `start` 함수 시작 주소에서 52바이트 떨어진 기계어 위치를 의미합니다.
즉, C 소스의 52번째 줄이 아닙니다. 라인 정보(`-g`)가 없어서 GDB가 바이트 오프셋으로만 표기합니다.

그리고 'continue -> Ctrl^C -> `x/i $pc`' 를 반복하면 계속 결과는 바뀝니다.

```sh
Program received signal SIGINT, Interrupt.
0x0000000080000136 in start ()
(gdb) x/i $pc
=> 0x80000136 <start+34>:    addi    a5,a5,1
(gdb) continue
Continuing.

^C
Program received signal SIGINT, Interrupt.
0x000000008000013c in start ()
(gdb) x/i $pc
=> 0x8000013c <start+40>:    ld    a4,-24(s0)
(gdb) continue
Continuing.

^C
Program received signal SIGINT, Interrupt.
0x0000000080000144 in start ()
(gdb) x/i $pc
=> 0x80000144 <start+48>:    addi    a5,a5,-320
(gdb) continue
Continuing.

^C
Program received signal SIGINT, Interrupt.
0x0000000080000148 in start ()
(gdb) x/i $pc
=> 0x80000148 <start+52>:    bltu    a4,a5,0x8000012a <start+22>
(gdb) continue
Continuing.

^C
Program received signal SIGINT, Interrupt.
0x0000000080000140 in start ()
(gdb) x/i $pc
=> 0x80000140 <start+44>:    auipc    a5,0xb
```

HMP에서 `system_reset`을 치면 QEMU는 즉시 리셋 및 재실행합니다.
이때 브레이크포인트를 메모리에 심어 둔 `ebreak`가 리셋 과정에서 사라질수 있습니다.
일반적으로 GDB는 `continue` 시점마다 브레이크를 재삽입하지만,
리셋 직후 이미 CPU가 달려서(`main` 호출을 지나) `start`의 `wfi` 루프까지 가버리면 못 잡습니다.

`tbreak *0x80000156`가 올바르게 작동하게 하려면 "리셋 직후 즉시 정지" 상태를 확보하고,
그 상태에서 브레이크포인트를 재확인한 다음 진행해야 합니다.

아직 BSS 제로화 루프를 도는 중이었고(`0x8000012a`로 분기하는 `bltu`), `main()` 진입 직전에서 멈춘 것으로 보입니다.
`wfi`는 `main()`이 끝난 뒤에야 실행되므로 아직 `wfi`가 아닙니다.
HMP에서 `LCR=0x00`, `LSR=0x60`인 것도 초기화 전, 리셋 기본 상태를 의미합니다.

`LSR=0x61`로 보였던 것은 수신 바이트가 들어왔다는 뜻이지만, 순서상 'a' 송신과는 무관합니다.

```sh
(gdb) x/i $pc
▌ => 0x8000012e <start+26>:    sb    zero,0(a5)
```

'start+26'의 `sb zero, 0(a5)`는 메모리 a5에 0을 1바이트 저장하라는 명령입니다.
즉 BSS 제로화 루프의 본문입니다.
