// 이 파일은 GNU as를 C 전처리기로 돌려 빌드합니다.
//   riscv64-unknown-elf-gcc -x assembler-with-cpp -c entry.S
//
// 부트 경로에 따라 두 가지를 지원합니다.
// - -bios none: M-mode에서 바로 시작합니다. mhartid CSR로 hart ID를 읽습니다.
// - OpenSBI(-bios default): -mode로 진입하며 a0=hartid, a1=DTB 포인터 관례를 따릅니다.
//
// 어셈블러에 _entry 심볼을 전역으로 노출하라고 지시합니다.
// 이를 통해 링커가 이 심볼을 프로그램의 시작 주소로 사용할 수 있게 됩니다.
// RISC-V에서 부트 로더(OpenSBI 등)는 커널 진입점 주소를 알고 있어야 하는데, 이 심볼이 바로 그 진입점이 됩니다.
.global _entry
// _entry 라벨을 선언합니다. 실제로 CPU가 처음 실행을 시작하는 어드레스가 됩니다.
_entry:
    // 현재 실행 중인 CPU가 몇 번째 코어인지를 레지스터 tp에 보관합니다.
    // mv는 addi의 별칭으로 레지스터 값을 그대로 복사합니다.
    // - OpenSBI가 M‑모드에서 S‑모드 커널을 부팅하는 경우 a0=hartid, a1=DTB 포인터를 넘겨주는 관례가 있다고 합니다.
    // - tp(Thread Pointer)는 xv6에서 hart ID(CPU 코어 번호)를 저장하는 용도로 사용합니다.
    // - 부트 로더(OpenSBI) 등 펌웨어가 넘겨줄 때, a0에는 현재 실행 중인 하트(hart)의 ID를 넘겨줍니다.
    // - 만약 스레드‑로컬 저장소를 쓸 계획이라면, tp에 hart ID를 계속 두는 설계는 깨집니다.
    //   초기에만 사용하고 이후에는 별도의 per‑CPU 구조체에 hart ID를 보관해야 합니다.
    mv      tp, a0 // tp=a0

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

// - 1:은 라벨입니다.
// - wfi(Wait For Interrupt)는 CPU를 대기 상태로 두고,
//   저전력 상태에서 인터럽트가 오면 깨어납니다.
//   부팅 시 특정 하트(hart)에서만 커널 초기화를 진행하고,
//   나머지 하트는 이렇게 대기시킨다고 합니다.
1: wfi
    // wfi -> j 1b로 무한 루프를 돌게 합니다.
    // - j는 무조건 분기입니다. 1
    // - 1b
    //   - b는 backward를 의미합니다.
    //   - 같은 라벨 번호 1 중 이전에 선언된 곳으로 점프하라는 뜻입니다.
    j 1b

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
