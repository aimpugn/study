#include <stdarg.h> // 가변 인자 함수(e.g. printf("%d %s", 10, "hi");)를 만들기 위한 매크로를 제공합니다.
#include <stdint.h> // uint8_t, uint32_t, uintptr_t 같은 고정 폭 정수 타입을 제공합니다.
#include "defs.h"

// UART_BASE는 RISC-V QEMU virt 머신에서 NS16550A UART 장치가 0x10000000부터 메모리에 매핑되어 있음을 의미합니다.
// 즉, 이 위치부터 UART 레지스터들이 배치되어 있다는 약속입니다.
// 이 값은 임의로 정하는 게 아니라 플랫폼(하드웨어 보드 혹은 에뮬레이터)의 설계에서 이미 정해져 있는 값입니다.
#define UART_BASE 0x10000000UL

// THR(Transmitter Holding Register)는 송신 버퍼 1바이트짜리 레지스터입니다.
// 보통 I/O 포트 기반 접근 시 offset 0에 해당합니다.
// QEMU virt에서도 UART_BASE + 0 위치에 있습니다.
// 따라서 UART_BASE + THR에 쓰면 UART 송신 레지스터에 값이 들어갑니다.
#define THR 0

// LSR(Line Status Register)은 UART의 상태를 알려주는 레지스터입니다.
// PC의 전통적인 16550A 사양에서는 offset이 5(즉, UART_BASE + 5)입니다.
// 단, QEMU virt에서는 레지스터 간격(stride)이 1이 아닐 수 있습니다.
// 가령 offset이 5라면 실제 주소는 'UART_BASE + 5 * stride'입니다.
#define LSR 5

// LSR(Line Status Register) 안에는 여러 비트가 있는데,
// 0x20(=16*2 = 32 = 0010 0000)는 THR Empty (Transmit Holding Register Empty) 상태를 나타냅니다.
// 이 값이 1이면 송신 버퍼가 비어 있으니 새로운 문자를 써도 된다는 뜻입니다.
#define LSR_THRE 0x20

#define ASCII_BACKSPACE 0x08

/**
 * inb의 첫 번째 인자는 하드웨어 I/O 레지스터의 메모리 주소입니다.
 */
static inline uint8_t inb(uintptr_t a) {
    return *(volatile uint8_t*)(UART_BASE + a);
}

/**
 * outb의 첫 번째 인자는 하드웨어 I/O 레지스터의 메모리 주소입니다.
 */
static inline void outb(uintptr_t a, uint8_t v) {
    *(volatile uint8_t*)(UART_BASE + a) = v;
}

/**
 * 하나의 문자를 커널 콘솔에 내보내는 함수
 *
 * 여기서 콘솔은 QEMU가 virt 보드의 NS16550A UART를 호스트 표준출력에 붙여서 보여줍니다.
 */
void uart_puts_sync(unsigned char c) {
    // LSR(Line Status Register)의 LSR_THRE 비트(0x20)는 “송신 버퍼가 비었는지”를 나타냅니다.
    while((inb(LSR) & LSR_THRE) == 0) {
        // UART에는 송신 버퍼(THR, Transmit Holding Register)가 있으며,
        // 한 번에 한 바이트만 담을 수 있습니다.
        // 여기서는 그 송신 버퍼가 빌 때까지 대기합니다.
    }

    outb(THR, (uint8_t) c); // 한 바이트 송신
}

/**
 * 외부에서 링크 가능한 심볼을 만들기 위해 static이 아닌 함수로 선언합니다.
 *
 * 실제로 문자열 s(가령 "hello %d")은 리터럴이므로 메모리의 읽기 전용 영역에 올라갑니다.
 * 만약 함수 안에서 수정하면 UB(Undefined Behavior)입니다.
 */
void puts_sync(const char *s) {
    for(; *s; s++) {
        if(*s == ASCII_BACKSPACE) {
            // if the user typed backspace, overwrite with a space.
            uart_puts_sync('\b'); uart_puts_sync(' '); uart_puts_sync('\b');
        } else {
            uart_puts_sync(*s);
        }
    }
}
