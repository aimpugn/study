// 이 파일은 GNU as를 C 전처리기로 돌려 빌드합니다.
//   riscv64-unknown-elf-gcc -x assembler-with-cpp -c entry.S
//
// 부트 경로에 따라 두 가지를 지원합니다.
// - -bios none: M-mode에서 바로 시작합니다. mhartid CSR로 hart ID를 읽습니다.
// - OpenSBI(-bios default): -mode로 진입하며 a0=hartid, a1=DTB 포인터 관례를 따릅니다.
//
// NOTE:
// 커널 입장에서는 여러 개의 hart(코어)가 동시에 이 코드를 실행할 수 있습니다.
// 따라서 'hart 0'만 커널 부팅을 진행하고, 나머지는 대기시키는 구조가 일반적입니다.
// 메모리 초기화, 페이지 테이블 구성, 장치 초기화는 한 번만 해야 하기 때문입니다.
// 만약 여러 hart가 동시에 실행한다면 동시에 같은 초기화 코드를 실행하여 메모리 할당 충돌, 콘솔 중복 초기화 등
// 다양한 문제가 발생할 수 있습니다.
//
//
// 어셈블러에 _entry 심볼을 전역으로 노출하라고 지시합니다.
// 이를 통해 링커가 이 심볼을 프로그램의 시작 주소로 사용할 수 있게 됩니다.
// RISC-V에서 부트 로더(OpenSBI 등)는 커널 진입점 주소를 알고 있어야 하는데, 이 심볼이 바로 그 진입점이 됩니다.
.global _entry
// _entry 라벨을 선언합니다. 실제로 CPU가 처음 실행을 시작하는 어드레스가 됩니다.
_entry:
    // 모든 하트가 여기까지 실행 가능하지만, 보통 hart 0만 start()를 호출하여 시스템 초기화를 담당하게 한다고 합니다.
    // 이를 위해 hart id를 검사합니다.
    bnez    a0, park_secondary   // a0 != 0 이면 secondary 라벨로 점프

boot_primary:
    // 현재 실행 중인 hart(코어, CPU)가 몇 번째 코어인지를 레지스터 tp에 보관합니다.
    // mv는 addi의 별칭으로 레지스터 값을 그대로 복사합니다.
    // - OpenSBI가 M‑모드에서 S‑모드 커널을 부팅하는 경우 a0=hartid, a1=DTB 포인터를 넘겨주는 관례가 있다고 합니다.
    // - tp(Thread Pointer)는 xv6에서 hart ID(CPU 코어 번호)를 저장하는 용도로 사용합니다.
    // - 부트 로더(OpenSBI) 등 펌웨어가 넘겨줄 때, a0에는 현재 실행 중인 하트(hart)의 ID를 넘겨줍니다.
    // - 만약 스레드‑로컬 저장소를 쓸 계획이라면, tp에 hart ID를 계속 두는 설계는 깨집니다.
    //   초기에만 사용하고 이후에는 별도의 per‑CPU 구조체에 hart ID를 보관해야 합니다.
    mv      tp, a0 // tp=a0

    // 모든 C 코드 진입에는 스택이 필요합니다.
    // - la(load address)는 심볼의 실제 주소를 레지스터에 적재합니다.
    //   bootstack_top의 주소를 스택 포인터 sp에 로드합니다.
    // - CPU가 커널 코드를 실행하려면 스택이 필요하기 때문에,
    //   커널이 사용할 임시 부트 스택 공간(bootstack ~ bootstack_top)을 준비합니다.
    // - sp(스택 포인터)를 설정하면 이후 C 코드 호출 시 로컬 변수, 함수 인자, 리턴 주소 저장 등에 스택을 쓸 수 있습니다.
    la      sp, bootstack_top // sp=bootstack_top

    // - 현재 위치에서 start라는 C 함수로 점프하면서 ra(리턴 주소) 레지스터에 복귀 주소를 저장합니다.
    //   즉, 어셈블리 초기화가 끝났으니 이제 C 코드의 진입점(start)으로 넘어갑니다.
    // - start 안에서 페이지 테이블 설정, 메모리 초기화, 콘솔 초기화 등 나머지 부팅 절차가 진행됩니다.
    call    start

.Lboot_primary:
    // - wfi(Wait For Interrupt)는 CPU를 대기 상태로 전환합니다.
    //   저전력 상태에서 인터럽트 또는 구현 정의 이벤트로 깨어납니다.
    //   부팅 시 특정 하트(hart)에서만 커널 초기화를 진행하고,
    //   나머지 하트는 이렇게 대기시킨다고 합니다.
    wfi
    // wfi -> j 1b로 무한 루프를 돌게 합니다.
    // - j(= jal x0, label)는 무조건 분기입니다.
    // - .Lboot_primary로 점프하라는 뜻입니다.
    j   .Lboot_primary

// 보조 hart가 처음 도달하는 주차 지점입니다.
// 스택을 건드리지 않아도 안전합니다.
// 보조 hart는 커널이 IPI 등으로 깨울 때까지 대기합니다.
park_secondary:
.Lpark_secondary:
    wfi
    j   .Lpark_secondary // 무한 대기

// 다음에 나오는 심볼(bootstack)의 주소를 16바이트 경계로 맞춥니다(alignment).
// 스택과 일부 데이터 구조는 정렬이 맞아야 성능과 안정성이 보장됩니다.
.balign 16

// 전역 심볼 bootstack을 선언하고, 여기서부터 실제 스택 메모리를 할당합니다.
.global bootstack
bootstack:
    // 현재 위치에서 4096바이트(4KB)를 예약합니다.
    // 스택은 위에서 아래로 자라기 때문에,
    // - bootstack_top이 스택의 시작점이 됩니다.
    // - bootstack이 스택의 바닥이 됩니다.
    .space 4096

// bootstack에서 4KB 뒤의 주소를 가리키는 심볼입니다
// 실제 sp(스택 포인터)는 이 bootstack_top을 시작점으로 사용합니다.
.global bootstack_top
bootstack_top:
