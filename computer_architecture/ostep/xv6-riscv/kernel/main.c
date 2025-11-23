#include "defs.h"

/**
 * ```
 * void main(void) {
 *     uart_init();
 *     uart_puts_sync('a');
 * }
 *
 * 위 코드는 UART를 초기화하고, 그걸 통해 문자열을 찍어 보는 것입니다.
 * 이 호출 체인은 결국 "인스트럭션의 연속"입니다.
 *
 * 컴파일이 끝나면 위의 `main()` 함수도 도 RISC-V 인스트럭션들의 묶음으로 바뀝니다.
 * 이 인스트럭션들은 `auipc` / `lui` / `addi` / `jal` / `sw` 같은 기계어이고,
 * CPU는 그저 PC를 따라가며 이 기계어들을 실행할 뿐입니다.
 *
 * 함수 호출은 결국 "`jal`로 PC를 점프하고, 리턴 주소를 레지스터에 저장하는 패턴"일 뿐입니다.
 */
void main(void) {
    // UART를 먼저 초기화합니다.
    // 부팅 직후 ns16550a의 라인 제어/피포/인터럽트 설정은 구현체마다 다릅니다.
    // 초기화 없이 LSR(송신 버퍼 비었는지)을 폴링하면 0이 계속 읽혀 무한 대기할 수 있으므로
    // 8N1, FIFO 사용, 인터럽트 비활성(폴링)으로 명시적으로 맞춰 둡니다.
    uart_init();

    // 이제 한 글자를 내보냅니다. QEMU `-nographic` 환경에서 표준 출력으로 바로 보입니다.
    // 만약 문자가 보이지 않으면, 레지스터 간격(reg-shift) 또는 권한 모드, 베이스 주소를 의심합니다.
    uart_puts_sync('a');
    uart_puts_sync('b');
    uart_puts_sync('\n');
    // // puts_sync("\nhello\n");

    // // 디버깅을 위해 눈에 확 띄는 문자열을 여러 번 출력합니다.
    // const char *msg = "HELLO from xv6-riscv kernel\n";
    // for (int i = 0; i < 3; i++) {
    //     for (const char *p = msg; *p; p++) {
    //         uart_puts_sync(*p);
    //     }
    // }

    // 여기서 끝나면 start()가 WFI 루프로 들어간다.
}
