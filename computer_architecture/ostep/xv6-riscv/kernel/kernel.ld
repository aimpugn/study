/* kernel.ld: DRAM 시작(virt 보드) 0x80000000에서 실행합니다. */

/* 출력 대상 아키텍처를 RISC-V로 지정합니다. */
OUTPUT_ARCH(riscv)

/*
 * 링커가 _entry를 ELF 엔트리로 잡아두면 QEMU는 커널을 메모리에 적재하고 그 주소에서 실행을 시작합니다.
 * - 프로그램 진입점을 entry.S에서 정의한 라벨 _entry 심볼로 지정합니다.
 * - 부트 시 CPU가 처음 실행할 위치입니다.
 */
ENTRY(_entry)

/*
 * - 링크 주소 시작점을 0x80000000으로 지정합니다.
 * - QEMU RISC-V virt 머신의 DRAM 시작 주소입니다.
 */
BASE = 0x80000000;

SECTIONS {
    . = BASE;

    /*
     * - 코드 영역으로 모든 .text* 섹션을 모아서 배치합니다.
     * - ALIGN(4K):
     *    - 현재 위치(.)를 4KB 경계(4096바이트 배수)로 맞추라는 뜻입니다.
     *      예를 들어, 현재 위치가 0x80001234라고 할 때, ALIGN(4K)를 만나면 0x80002000으로 앞으로 당겨서
     *      그 지점부터 섹션 내용을 채웁니다. 중간의 빈 공간은 0으로 채워집니다.
     *    - RISC-V ‘virt’ 머신에서 페이지 크기는 4KB입니다.
     *    - 코드(.text), 읽기 전용 데이터(.rodata), 쓰기 가능한 데이터(.data)와 같이 성격이 다른 섹션들을
     *      페이지 경계에 맞추면 나중에 페이지 권한(PTE R/W/X)을 깔끔하게 설정할 수 있습니다.
     */
    .text : ALIGN(4K) {
        /*
         * - `*(패턴)`은 모든 오브젝트 파일에서 해당 섹션 이름과 매치되는 부분들을 전부 모아 넣으라는 의미입니다.
         * - `.text.*`는 `.text.startup`, `.text.main` 등 세부 섹션들을 모두 포함합니다.
         */
        *(.text*)
    }

    /*
     * - 상수 문자열 등 읽기 전용(read only) 데이터
     */
    .rodata : ALIGN(4K) {
        *(.rodata*)
    }

    /*
     * - 초기화된 전역, 정적 변수
     */
    .data : ALIGN(4K) {
        *(.data*)
    }

    /*
     * - 초기화되지 않은 전역, 정적 변수.
     * - COMMON 심볼을 포함합니다.
     */
    .bss : ALIGN(4K) {
        /* .bss 시작(.data의 끝) */
        edata = .; /* start.c의 extern char edata[]; */
        *(.bss* COMMON)
        end = .; /* start.c의 extern char end[]; */
        /* 커널 이미지 끝 */
    }
}
