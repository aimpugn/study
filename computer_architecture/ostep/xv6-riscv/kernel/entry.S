/*========================================================================================*
 * 이 파일은 GNU as를 C 전처리기로 돌려 빌드합니다.
 *   riscv64-unknown-elf-gcc -x assembler-with-cpp -c entry.S
 *
 * 이 커널은 외부 로더나 C 런타임이 없다는 것을 전제합니다.
 * 따라서 CPU가 첫 명령을 어디서 시작하는지, 첫 스택을 어디로 둘지, psABI가 요구하는 레지스터 상태(gp, sp 16B 정렬 등)를
 * 직접 수립해야 합니다. 이 단계가 틀리면 이후 C 코드의 전제들이 모두 무너집니다.
 *
 * 부트 경로에 따라 두 가지를 지원합니다.
 * - -bios none (bare metal, M-mode):
 *     CPU가 M-mode에서 커널을 직접 실행합니다.
 *     이 경우 `mhartid` CSR로 hart ID를 직접 읽을 수 있습니다.
 *
 * - OpenSBI(-bios default):
 *     S-mode로 진입하며, OpenSBI는 커널에 a0=hartid, a1=DTB(FDT 포인터)를 넘깁니다.
 *     이 경우 hart ID를 CSR로 읽지 않고 a0를 신뢰해야 하며, M-mode 전용 CSR(csrr mhartid 등)을
 *     사용하는 명령은 S-mode에서 illegal instruction trap을 발생시킵니다.
 *
 * psABI(processor-specific Application Binary Interface)를 준수합니다.
 * 현재 프로젝트는 RISC-V이므로 RISC-V ABI 문서를 참고합니다.
 * - https://github.com/riscv-non-isa/riscv-elf-psabi-doc
 * - https://riscv-non-isa.github.io/riscv-elf-psabi-doc/
 *
 * > NOTE:
 * >
 * > 커널 입장에서는 여러 개의 hart(코어, CPU)가 동시에 이 코드를 실행할 수 있습니다.
 * > 여기서 hart(Hardware thread, 하드웨어 스레드)는 커널(특히 RISC-V 아키텍처 환경)에서
 * > 하드웨어 실행 컨텍스트(Hardware execution context)를 의미합니다.
 * >
 * > 소프트웨어 실행 관점에서, HART는 독립적으로 명령을 가져와 실행하는 자율적인 리소스로 작동합니다.
 * > 즉, 운영 체제나 애플리케이션이 보기에는 하나의 독립된 CPU 코어 또는 하이퍼 스레드처럼 보이는 추상화된 처리 단위입니다.
 * > 실제 마이크로아키텍처 구현 전략(예: 실제 물리적 하드웨어 스레드)과 소프트웨어가 인식하는 실행 컨텍스트 간의 혼동을 없애기 위해 도입된 용어입니다.
 * > 예를 들어, 하나의 물리적 코어에 여러 개의 하이퍼 스레드가 있을 경우,
 * > 각각의 하이퍼 스레드는 별도의 HART로 간주될 수 있습니다.
 * >
 * > 따라서 'hart 0'만 커널 부팅을 진행하고, 나머지는 대기시키는 구조가 일반적입니다.
 * > 메모리 초기화, 페이지 테이블 구성, 장치 초기화는 한 번만 해야 하기 때문입니다.
 * > 만약 여러 hart가 동시에 실행한다면 동시에 같은 초기화 코드를 실행하여 메모리 할당 충돌, 콘솔 중복 초기화 등
 * > 다양한 문제가 발생할 수 있습니다.
 * > 나머지 하트는 wfi 등으로 주차합니다. SMP를 깨울 때만 초기화 코드를 1회 보장합니다.
 *
 * ELF 파일에는 '이 바이너리를 실행할 때 첫 PC를 어디에 둘 것인지'를 나타내는 `e_entry` 필드가 있습니다.
 * 링커는 `ENTRY(_entry)` 지시문을 보고 `_entry` 심볼의 주소를 계산해서 그 값으로 `e_entry`를 채웁니다.
 * QEMU의 `-kernel` 모드는 ELF를 직접 파싱해서 `e_entry` 값을 RISC-V CPU의 PC 레지스터에 그대로 넣고
 * 실행을 시작합니다.
 *
 * 커널이 스스로 부팅하려면 "CPU가 처음 실행해야 할 가상주소"를 누군가는 알려 주어야 합니다.
 * ELF 형식은 그 주소를 헤더의 e_entry 필드에 숫자로 기록해 둡니다.
 *
 * 링커가 이 필드를 채우는 방법은 다음 두 가지 방법이 있습니다:
 * 1) 링커 스크립트(./kernel/kernel.ld)에서 ENTRY(_entry) 지시문을 쓰면,
 *    _linker는 ‘_entry’ 심볼의 실제 주소를 계산해 e_entry 에 복사합니다.
 *
 *    고정된 링커 스크립트를 사용할 경우 ENTRY 하나면 충분합니다.
 *
 * 2) ld 실행 시 `-e _entry` 옵션을 주면, 스크립트에 무엇이 있든
 *    해당 옵션이 우선권을 가져서 동일한 작업을 수행합니다.
 *
 *    여러 버전의 스크립트가 존재하는 빌드 시스템 등의 경우 `-e` 옵션 사용에 이점이 있습니다.
 *
 * e_entry 는 그저 의도를 적어 놓은 숫자일 뿐이고, 실제 그 주소로 점프할지는 실행 환경의 몫입니다.
 *
 * - QEMU -kernel 모드는 ELF 파일을 직접 파싱하고 e_entry 에 적힌 가상주소를 그대로 CPU PC 레지스터에
 *   넣습니다. 따라서 ENTRY(_entry) 지시문이 곧바로 하드웨어 실행 위치가 됩니다.
 *
 * - OpenSBI는 커널이 ELF냐 RAW 이미지냐에 따라 진입점을 다르게 정합니다.
 *     - ELF payload: ELF 헤더의 e_entry 값을 그대로 읽어 jump.
 *     - RAW 이미지: 빌드 시 지정한 고정 주소(`FW_TEXT_START` 값, 기본 `0x8020_0000`)로 jump
 *
 *     둘 다 OpenSBI 내부의 고정 점프 루틴이 수행하므로,
 *     커널이 RAW 이미지로 빌드된 경우에는 ELF e_entry 자체가 존재하지 않거나 무시됩니다.
 *
 * - 실제 SoC ROM 코드 또한 칩 설계자가 하드코딩한 물리주소로만 점프할 수 있으므로,
 *   ELF e_entry를 무시할 수 있습니다.
 *
 * 따라서 e_entry라는 것은 읽히는 환경에서만 유효합니다.
 * QEMU -kernel 실행처럼 ELF를 그대로 로드하는 경우에는 반드시 설정해야 하고,
 * OpenSBI 전용 이미지라면 없어도 부팅은 가능합니다.
 *
 * 링커 스크립트(./kernel/kernel.ld)에서 ENTRY(_entry)로 지정하고 있으므로,
 * 링커가 이 심볼을 프로그램의 시작 주소로 사용할 수 있게 됩니다.
 *========================================================================================*/

// `_entry`는 보통 어셈블리 파일(`entry.S`)에 있고, 하는 일은 크게 두 가지입니다.
// 1. 스택 포인터를 올바른 위치로 맞추는 것
// 2. C 함수 start()를 호출하는 것
.global _entry // 링커가 참조할 진입 심볼을 전역으로 공개합니다.

// _entry 라벨을 선언합니다. 실제로 CPU가 처음 실행을 시작하는 어드레스가 됩니다.
_entry:
    /*======================================================================================*
     * .data는 초기값이 있는 전역/정적의 저장소이고, .bss는 초기값이 0인 전역/정적의 저장소입니다.
     *
     * - .data: 초기값이 있는 전역/정적의 저장소
     * - .bss: 초기값이 0인 전역/정적의 저장소
     *
     * .sdata와 .sbss는 그 중에서도 작은 크기의 전역만 따로 모은 영역입니다.
     *
     * 컴파일러는 "전역 크기가 임계값 이하일 때" 이 작은 전역을 `.sdata`/`.sbss`에 배치할 수 있고,
     * 이 임계값은 -msmall-data-limit=N 옵션으로 조정합니다.
     * - https://llvm.org/docs/RISCVUsage.html#global-pointer-gp-relaxation-and-the-small-data-limit
     *
     * RISC-V의 load/store(I형) 즉시값은 부호 있는 12비트입니다.
     * 작은 전역을 .sdata/.sbss로 모으면, 그 묶음이 gp에서 −2048..+2047 바이트 범위(12비트에서 부호 1비트 빼고 2^11) 안에 들도록 배치할 수 있습니다.
     *
     * 그리고 이 묶음 안의 전역은 모두 `lw/sw …, off(gp)` 한 한 명령으로 읽고 쓸 수 있습니다.
     * 예를 들어, 전역 foo에 대해 `lw a0, off(foo)(gp)` 처럼 gp-relative 명령으로 접근할 수 있습니다.
     *
     * `__global_pointer$`는 링커가 배치하는 특수 심볼로 .sdata, .sbss 묶음의 가운데쯤을 가리키는 주소값입니다.
     * .sdata 시작 + 0x800 지점에 배치하여 .sdata, .sbss 전역 대부분이 gp의 ±2048 바이트 범위에 들게 합니다.
     *
     * 그러면 전역 접근을 "gp + 오프셋" 한 번의 load/store(GP relaxation)로 끝낼 수 있어서
     * PC-relative 주소 형성 또는 절대 주소 형성 방식보다 명령 수가 줄고 코드 크기가 작아집니다.
     * - 절대 주소 형성(LUI+ADDI): 상위 20비트를 즉시값으로 적재하고 하위 12비트를 0으로 둡니다. 이후 ADDI로 12비트 상수를 더해 절대 주소를 형성합니다.
     * - PC-relative 주소 형성(AUIPC+ADDI): (현재 PC의 상위 20비트에 즉시값을 더한 값을 만들고, 이후 ADDI 또는 LD 등으로 PC-relative 주소를 형성합니다.
     *
     * 즉, 주소 형성의 경우에는 기준을 만들고(+20비트) 미세 조정(+12비트)을 합쳐 최소 2단계가 필요합니다.
     *
     * 전역 포인터(GP)는 __global_pointer$ 심볼의 값((Value of __global_pointer$ symbol)입니다.
     * RISCV ABI는 프로그램 시작 코드가 __global_pointer$ 심볼의 값을 gp(x3) 레지스터에 로드한다고 가정합니다.
     * 시작 코드가 이 값을 gp(x3)에 먼저 적재해야, 링크러가 만들어 둔 gp-relative 접근이 전부 유효해집니다.
     *
     * > Global Pointer: It is assumed that program startup code will load the value of the __global_pointer$ symbol into register gp (aka x3).
     * > - [8.4.2. Calculation Symbols](https://riscv-non-isa.github.io/riscv-elf-psabi-doc/#_calculation_symbols)
     *======================================================================================*/
    .option push
    /*
     * 이 블록 안은 링크 시간 동안 릴랙세이션 간섭을 금지합니다.
     *
     * `la gp, __global_pointer$`는 링크 전에는 "pc-relative AUIPC + ADDI로 gp에 주소를 넣는 패턴"으로
     * 남고, 릴랙스 과정에서 더 좋은 표현이 있다면 바뀔 수 있습니다.
     * 그런데 `gp`는 psABI 수준에서 이 심볼 값 그대로를 넣으라고 강제하는 레지스터라, 이 블록을 건드리지 않도록 합니다.
     *
     * 이 옵션으로 보호된 상태에서 `la gp, __global_pointer$`를 실행하면, 링커와 툴체인 릴랙스가 있어도
     * 최종 바이너리에서 `gp`는 `__global_pointer$`를 정확히 가리키게 되고, 이후의 gp-relative load/store는
     * 모두 의도한 주소를 바라보게 됩니다.
     */
    .option norelax
    la      gp, __global_pointer$ // la(load address)는 심볼의 실제 주소를 레지스터에 적재합니다.
    .option pop

    /*
     * bare metal(M-mode) 경우 hartid를 읽습니다.
     * hart는 RISC-V에서 하드웨어 실행 컨텍스트를 나타내는 단위입니다.
     * 물리 코어에 여러 하이퍼스레드가 있다면, 각 하이퍼스레드가 각각 하나의 hart가 됩니다
     *
     * mhartid CSR은 현재 hart의 ID를 돌려주고, QEMU virt에서는 간단히 0,1,2, 등의 번호를 줍니다.
     *
     * 참고로 mhartid CSR은 M-mode 전용이므로, S-mode에서 실행하면 illegal instruction trap입니다.
     */
    csrr    t0, mhartid

    /*========================================<스택 배치>========================================*
     * `stack0` 배열을 hart 수만큼 4KiB 조각으로 나누고,
     * 각 hart가 자기 조각의 가장 위 주소를 sp로 잡습니다.
     *
     * C로 진입하면 콜 프레임, 지역 변수, 리턴 주소 저장이 필요합니다.
     * 규약상 `sp(스택 포인터) % 16 == 0` 입니다.
     *
     * 전역 배열 stack0를 하트당 4KiB 조각으로 가정하고, 각 하트의 '조각 꼭대기'에 sp를 둡니다.
     * 4096은 16의 배수이므로 시작 주소가 16바이트 정렬이면 sp도 자동 정렬이 유지됩니다.
     * 이를 위해 start.c에 stack0를 `__attribute__((aligned(16)))`으로 선언합니다.
     *
     * 하트별로 스택을 준비하지 않는다면 하트 간 스택 경합 또는 미정렬로 인해 하트별 콜러/콜리가 섞여서
     * 충돌이 발생합니다.
     *
     * 아래로 증가하는 스택(stack0) 메모리는 아래와 같이 배치됩니다.
     * [hart0 4KiB][hart1 4KiB][hart2 4KiB]...[hartN 4KiB]
     *            ^ '각 조각의 끝'을 sp로 잡고 아래로 사용합니다.
     *
     * Note:
     * 단, hart 1 이상이 동시에 진입하면 스택 프레임, 중첩 call, 인터럽트 stack push가 충돌할 수 있습니다.
     * 그래서 xv6 원본처럼 `stack0 + (hartid + 1) * 4096` 계산으로 hart별 스택 최상단을 잡아야 안전합니다.
     * sp = &stack0 + 4096 * (hartid + 1)
     *======================================================================================*/
    la      sp, stack0          // stack0 심볼의 주소를 sp에 로드합니다. 스택 풀의 베이스 주소가 됩니다.
                                // 이떄 stack0는 C 코드 어디선가 정의되어 있어야 합니다.
                                // e.g. __attribute__((aligned(16))) char stack0[4096 * NCPU];
    li      t1, 4096            // 하트당 스택 크기를 4096(1024 * 4)으로 잡습니다.
    addi    t2, t0, 1           // t2 <= hartid + 1 (스택 상단으로 이동하기 위함)
    mul     t1, t1, t2          // t1 <= 4096 * (hartid + 1)
    add     sp, sp, t1          // sp <= &stack0 + 4096 * (hartid + 1)
    // 이 과정을 거치면 메모리에서 stack0 배열은 다음과 같이 배치되고,
    // 각 hart는 자기 조각의 맨 위 주소를 sp로 갖게 됩니다:
    //
    // [hart0용 4KiB][hart1용 4KiB][hart2용 4KiB] ...

    /*======================================================================================*
     * hart ID를 tp(Thread Pointer)에 보관합니다.
     *
     * hart ID는 RISC-V의 하드웨어 스레드 식별자입니다.
     * tp(Thread Pointer)는 RISC-V ABI에서 각 스레드별로 값이 달라지는 전역인 TLS(Thread-Local Storage)
     * 기준 레지스터로 예약된 이름입니다.
     * 이는 컴파일러가 TLS 변수를 접근할 때 tp를 기준으로 오프셋으로 접근하는 특별한 코드열을 생성한다는 것을 의미합니다.
     * - https://riscv-non-isa.github.io/riscv-elf-psabi-doc/#_thread_local_storage
     *
     * hart ID를 일관된 방식으로 참조하지 않으면 hart(코어, CPU)별 자료구조 인덱싱이 틀어지게 되고,
     * 스케줄러/락 등 hart 로컬 상태 관리가 깨지게 됩니다.
     *
     * xv6-riscv 경우에는 `tp = hartid`로 고정합니다.
     * `tp = hartid`로 고정하면:
     * 1) 어떤 커널 함수에서도 한 명령으로 tp를 읽어 현재 hart가 누구인지를 알 수 있습니다.
     * 2) 그리고 hart(코어, CPU)별 인덱싱을 분기 없이 수행할 수 있습니다.
     *
     * 예를들어, xv6-riscv의 r_tp C 함수는 다음과 같습니다:
     * ```c
     * // read and write tp, the thread pointer, which xv6 uses to hold
     * // this core's hartid (core number), the index into cpus[].
     * static inline uint64
     * r_tp()
     * {
     *     uint64 x;
     *     asm volatile("mv %0, tp" : "=r" (x) );
     *     return x;
     * }
     * ```
     *
     * 하지만 이를 위해서는 두 가지 조건이 필요합니다:
     * 1) 커널을 빌드할 때 TLS 코드를 생성하지 않습니다. 가령 `__thread` 같은 TLS를 사용하지 않습니다.
     * 2) 어셈블리나 인라인 어셈블리가 tp를 임시 레지스터로 쓰지 않습니다.
     *
     * 그래서 만약 스레드 로컬 저장소(TLS)를 사용한다면 tp에 hart ID를 계속 두면 안 됩니다.
     * 대신 tp를 hart(코어, CPU)별 베이스 포인터로 변환하여 정적 오프셋으로 접근하도록 하고,
     * 해당 hart(코어, CPU)별 구조체 안에 hart ID를 보관해야 합니다.
     *======================================================================================*/
    mv      tp, t0 // mv는 addi의 별칭으로 레지스터 값을 그대로 복사합니다.

    /*======================================================================================*
     * 멀티 hart인 경우를 처리합니다.
     * hart 0만 부트로 진입하고, 나머지는 주차시킵니다.
     *
     * 즉, 메모리 초기화, 페이지 테이블 구성, 장치 초기화 등을 hart 0가 "오직 한 번"만 처리하도록 합니다.
     * - hartid == 0이면 boot_primary 루프로 보내 start()를 호출하게 합니다.
     * - hartid != 0이면 park_secondary 루프로 보내 WFI로 대기하게 합니다.
     *======================================================================================*/
    bnez    t0, park_secondary   // t0 != 0 이면 secondary 라벨로 점프

# t0 == 0으로 분기 불발되면 곧장 boot_primary 라벨로 점프
boot_primary:
    /*======================================================================================*
     * 현재 권한 모드(bare metal로 테스트중이므로 M-mode)에서 start.c의 start 함수를 호출합니다.
     *
     * 현재 위치에서 start라는 C 함수로 점프하면서 ra(리턴 주소) 레지스터에 복귀 주소를 저장합니다.
     *
     * start 함수 안에서는 BSS 제로화(링커 스크립트의 edata..end 범위를 0으로 채우기),
     * 메모리 초기화, 페이지 테이블 설정, 콘솔 초기화 등을 C 코드에서 수행합니다.
     *
     * 만약 이런 초기화를 제대로 수행하지 않으면 전역 락/포인터/카운터들이 쓰레기값에서 시작하게 되어
     * 코드를 신뢰할 수 없게 됩니다.
     *======================================================================================*/
    call    start
.Lboot_primary:
    fence   rw, rw
    // - wfi(Wait For Interrupt)는 CPU를 대기 상태로 전환합니다.
    //   저전력 상태에서 인터럽트 또는 구현 정의 이벤트로 깨어납니다.
    //   부팅 시 특정 하트(hart)에서만 커널 초기화를 진행하고,
    //   나머지 하트는 이렇게 대기시킨다고 합니다.
    wfi
    // wfi -> j 1b로 무한 루프를 돌게 합니다.
    // - j(= jal x0, label)는 무조건 분기입니다.
    // - .Lboot_primary로 점프하라는 뜻입니다.
    j       .Lboot_primary

/*======================================================================================*
 * hart 0 아닌 다른 hart들을 위한 주차 루프입니다.
 * hart 0와의 경쟁하거나 중복 초기화하는 것 등을 방지합니다.
 * 나중에 SMP를 깨울 때만 IPI(Inter Processor Interrupt)/플래그로 빠져나옵니다.
 *======================================================================================*/
park_secondary:
.Lpark_secondary:
    wfi
    j   .Lpark_secondary // 무한 대기
