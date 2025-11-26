# `make` 명령어는 "타깃 <- 의존물" 그래프와 타임스탬프를 보고 필요한 것만 빌드합니다.
# 가령 `make qemu` 경우 qemu 타깃의 의존물 $K/kernel.elf를 만들기 위해 오브젝트들을 빌드하고,
# 링크 스크립트로 ELF를 만든 뒤 QEMU를 실행한다.

K=kernel

# make 호출 시 CPUS 미지정이면 기본값 3
# make qemu는 -smp 3으로, hart 0, 1, 2 세 개를 동시에 부팅합니다
# RISC‑V에서 hart는 "하드웨어 스레드 = 코어"라고 볼 수 있습니다.
# QEMU는 이 세 hart 모두에 대해 동일한 _entry 주소에서 동시에 실행을 시작시킵니다.
ifndef CPUS
CPUS := 3
endif

ifndef MEMORY
MEMORY := 32M
endif

TOOL_PREFIX = riscv64-elf-
# C 컴파일러, 링커, 어셈블러, 오브젝트 변환 툴
GCC = $(TOOL_PREFIX)gcc
LD = $(TOOL_PREFIX)ld

CFLAGS = -march=rv64gc # 타겟 ISA: 64비트 RISC-V. 기본 ISA + 압축 명령어 + 원자적 연산 등 지원.
CFLAGS += -mabi=lp64 # ABI: 64비트 long, pointer
CFLAGS += -Wall # 경고 활성화
CFLAGS += -mcmodel=medany # 코드 모델을 medany로 설정하여 PC-relative 주소 생성
CFLAGS += -ffreestanding # 표준 라이브러리 환경이 아니라 임베디드, 커널 환경이라는 표시
# CFLAGS += -nostdlib # 표준 C 런타임(ctr0.o 등)과 라이브러리를 링크하지 않음
CFLAGS += -nostartfiles # crt0 등 시작파일 생략(드라이버 링크에 영향)

# 빌드 경로 포함한 타깃 ELF를 구성하는 오브젝트 목록.
# 최종적으로 kernel.elf 파일을 만들게 합니다.
# 그러기 위해서 모든 오브젝트를 kernel.ld 스크립트로 링크합니다.
OBJS = $K/entry.o \
	$K/printf.o \
	$K/start.o \
	$K/main.o

# 링크 옵션
LDFLAGS = -z max-page-size=4096 # 세그먼트 페이지 정렬 상한을 4KiB로(virt 페이지 크기와 일치)
LDFLAGS += -g # DWARF 디버그 정보를 넣기

# 최종 ELF. 의존물(오브젝트들과 링크 스크립트)이 갱신되면 다시 링크합니다.
$K/kernel.elf: $(OBJS) $K/kernel.ld
	$(LD) $(LDFLAGS) -T $K/kernel.ld -o $K/kernel.elf $(OBJS)

# C 소스(%.c) -> 오브젝트(패턴 규칙, %.o)
# 자동변수:
# - `$@`: 타깃(.o)
# - `$<`: 첫 의존물(.c)
# - `$^`: 모든 의존물
%.o: %.c
	$(GCC) $(CFLAGS) -c -o $@ $<

# 어셈블리 -> 오브젝트. .S는 C 전처리기 적용(S 의미)을 의미하고, GCC 드라이버로 컴파일합니다.
$K/entry.o: $K/entry.S
	$(GCC) $(CFLAGS) -c -o $@ $<

# QEMU 실행 관련
# 시스템 에뮬레이터 바이너리
QEMU = qemu-system-riscv64

# <시스템 에뮬레이터 옵션 정리>
#
# 1. `-machine virt`:
# "virt라는 이름의 가상 보드"를 선택합니다.
# 이 보드는 하나의 RISC-V SoC를 흉내 내는데, 그 안에는 여러 개의 hart(코어), DRAM, UART,
# PLIC, CLINT 등이 자리 잡고 있습니다.
#
# 이 보드의 하드웨어 구성은 virt.dts 같은 디바이스 트리에서 대략 확인할 수 있고,
# 그 안에는 "DRAM은 0x8000_0000에서 시작해서 얼마까지", "UART0는 0x1000_0000에 있다" 같은 정보가 들어 있습니다.
#
# kernel.ld가 `BASE`를 0x80000000으로 두고, printf.c가 `UART_BASE`를 0x10000000으로 두는 것은
# 바로 이 QEMU virt 보드의 메모리 맵을 따라가기 때문입니다.
QEMU_OPTS = -machine virt
# 2. `-bios none`:
# 부트롬이나 펌웨어를 쓰지 말고, `-kernel`로 전달되는 ELF 파일을 바로 DRAM에 올리고
# 그 ENTRY에서 실행을 시작하라는 의미입니다.
#
# 일반적인 실제 기계에서는 전원이 켜지면 먼저 마이크로코드, 부트 ROM, 펌웨어(예: U-Boot, OpenSBI)가 실행되고,
# 그 펌웨어가 다시 OS 커널을 로딩합니다.
# 이 옵션은 그런 모든 단계를 삭제하고 QEMU가 바로 커널 역할만 남겨 둡니다.
#
# QEMU는 kernel/kernel.elf 파일을 열고, ELF 헤더 안의 프로그램 헤더(PT_LOAD)를 보면서
# '이 세그먼트를 DRAM의 어느 주소에 올려야 하는지'를 계산합니다.
# 바로 이때 kernel.ld에서 정한 주소가 사용됩니다.
QEMU_OPTS += -bios none
QEMU_OPTS += -kernel $K/kernel.elf # 앞서 생성한 ELF를 커널로 사용
QEMU_OPTS += -m $(MEMORY)
QEMU_OPTS += -smp $(CPUS) # vCPU 개수. 현재 기본값은 3이고, make CPUS=1로 재정의 가능
# `-nographic` 옵션은 두 가지를 동시에 합니다.
# - 그래픽(디스플레이) 끔.
# - 첫 번째 시리얼 포트(serial0)를 stdio에 연결하고, 모니터를 Ctrl-a 키 조합으로 multiplex.
#   그래서 Ctrl-a x로 종료할 수 있습니다.
QEMU_OPTS += -nographic

qemu: $K/kernel.elf
	$(QEMU) $(QEMU_OPTS)

# - dumpdtb: QEMU가 가상 보드를 만들 때 내부에서 동적으로 생성하는 DTB(Device Tree Blob)를 파일로 출력할 것을 지시
# - 가상 보드: DRAM=0x80000000, UART/PLIC/CLINT 등 제공
QEMU_DTB_OPTS = -machine virt,dumpdtb=virt.dtb
QEMU_DTB_OPTS += -bios none # 펌웨어 없이 `-kernel`로 준 바이너리를 DRAM 시작에 로드하여 실행
QEMU_DTB_OPTS += -kernel $K/kernel.elf # 앞서 생성한 ELF를 커널로 사용
QEMU_DTB_OPTS += -m $(MEMORY)
QEMU_DTB_OPTS += -smp $(CPUS) # vCPU 개수. 현재 기본값은 3이고, make CPUS=1로 재정의 가능
QEMU_DTB_OPTS += -nographic # 그래픽 끄고 stdio로 콘솔 사용(Ctrl-a x로 종료)

qemu-dtb: $K/kernel.elf
	$(QEMU) $(QEMU_DTB_OPTS)

QEMU_GDB_OPTS = -machine virt # 가상 보드: DRAM=0x80000000, UART/PLIC/CLINT 등 제공
QEMU_GDB_OPTS += -bios none # 펌웨어 없이 `-kernel`로 준 바이너리를 DRAM 시작에 로드하여 실행
QEMU_GDB_OPTS += -kernel $K/kernel.elf # 앞서 생성한 ELF를 커널로 사용
QEMU_GDB_OPTS += -m $(MEMORY)
QEMU_GDB_OPTS += -smp $(CPUS) # vCPU 개수. 현재 기본값은 3이고, make CPUS=1로 재정의 가능
#QEMU_GDB_OPTS += -nographic # 그래픽 끄고 stdio로 콘솔 사용(Ctrl-a x로 종료)
# - `-S`: QEMU가 CPU를 reset 상태에서 정지시킵니다. 전원이 들어갔지만 아직 한 줄도 실행하지 않은 상태에 해당합니다.
# - `-s`: `-gdb tcp::1234`와 동일합니다. TCP 포트 1234에서 GDB 접속을 대기하도록
#         QEMU 안에 간단한 gdbstub을 띄웁니다.
QEMU_GDB_OPTS += -S -s
# 줄 단위 스텝/지역 변수 등 풍부한 정보를 보려면 `-g`로 DWARF 디버그 정보를 넣습니다.
# QEMU_GDB_OPTS += -g -O0
# QEMU_GDB_OPTS += -fno-omit-frame-pointer
# 모니터를 텔넷 포트로 분리해 키보드 매핑 문제가 없습니다.
# telnet 127.0.0.1 4444 또는 nc 127.0.0.1 4444 같은 프로그램으로 접속하면,
# 모니터 프롬프트((qemu))를 바로 받을 수 있습니다.
QEMU_GDB_OPTS += -monitor telnet:127.0.0.1:4444,server,nowait

qemu-gdb: $K/kernel.elf
	$(QEMU) $(QEMU_GDB_OPTS)

clean:
	rm -f $K/kernel.elf \
		$K/*.o
