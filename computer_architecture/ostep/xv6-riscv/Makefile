# `make` 명령어는 "타깃 <- 의존물" 그래프와 타임스탬프를 보고 필요한 것만 빌드합니다.
# 가령 `make qemu` 경우 qemu 타깃의 의존물 $K/kernel.elf를 만들기 위해 오브젝트들을 빌드하고,
# 링크 스크립트로 ELF를 만든 뒤 QEMU를 실행한다.

K=kernel

# make 호출 시 CPUS 미지정이면 기본값 3
ifndef CPUS
CPUS := 3
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

# 시스템 에뮬레이터 옵션
QEMU_OPTS = -machine virt # 가상 보드: DRAM=0x80000000, UART/PLIC/CLINT 등 제공
QEMU_OPTS += -bios none # 펌웨어 없이 `-kernel`로 준 바이너리를 DRAM 시작에 로드하여 실행
QEMU_OPTS += -kernel $K/kernel.elf # 앞서 생성한 ELF를 커널로 사용
QEMU_OPTS += -m 128M
QEMU_OPTS += -smp $(CPUS) # vCPU 개수. 현재 기본값은 3이고, make CPUS=1로 재정의 가능
QEMU_OPTS += -nographic # 그래픽 끄고 stdio로 콘솔 사용(Ctrl-a x로 종료)

qemu: $K/kernel.elf
	$(QEMU) $(QEMU_OPTS)

# - dumpdtb: QEMU가 가상 보드를 만들 때 내부에서 동적으로 생성하는 DTB(Device Tree Blob)를 파일로 출력할 것을 지시
# - 가상 보드: DRAM=0x80000000, UART/PLIC/CLINT 등 제공
QEMU_DTB_OPTS = -machine virt,dumpdtb=virt.dtb
QEMU_DTB_OPTS += -bios none # 펌웨어 없이 `-kernel`로 준 바이너리를 DRAM 시작에 로드하여 실행
QEMU_DTB_OPTS += -kernel $K/kernel.elf # 앞서 생성한 ELF를 커널로 사용
QEMU_DTB_OPTS += -m 128M
QEMU_DTB_OPTS += -smp $(CPUS) # vCPU 개수. 현재 기본값은 3이고, make CPUS=1로 재정의 가능
QEMU_DTB_OPTS += -nographic # 그래픽 끄고 stdio로 콘솔 사용(Ctrl-a x로 종료)

qemu-dtb: $K/kernel.elf
	$(QEMU) $(QEMU_DTB_OPTS)

clean:
	rm -f $K/kernel.elf \
		$K/*.o
