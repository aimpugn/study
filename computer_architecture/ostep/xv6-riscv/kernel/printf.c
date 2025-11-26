
/*
 * 송신 경로 요약
 *
 *  1. main()에서 uart_init()이 실행:
 *  - IER = 0x00: 폴링만 할 것이므로 모든 UART 인터럽트를 끕니다.
 *  - LCR = 0x80:
 *
 *      LCR=DLAB=1로 두어 분주 레지스터(DLL/DLH)를 접근합니다.
 *      그리고 THR(=DLL) = 0x03, IER(=DLH) = 0x00: 38.4 kbps 분주값을 기록합니다.
 *
 *      ns16550a 같은 16550 계열 UART 칩은 한 번에 모든 레지스터를 위한 주소 공간을 마련하기 어려웠던
 *      초창기 ISA 호환 설계를 지금까지 이어 오고 있습니다.
 *
 *      그래서 라인 제어 레지스터(LCR, Line Control Register) 안에 DLAB(Divisor Latch Access Bit) 라는
 *      특별한 비트를 두고, 그 비트를 1로 세트했을 때만 속도(baud rate)를 정하는
 *      분주 레지스터(DLL/DLH, Divisor Latch Low/High)에 접근할 수 있게 만들어 놨습니다.
 *
 *      > 분주 레지스터(DLL/DLH, Divisor Latch Registers)는 주로
 *      > UART(Universal Asynchronous Receiver/Transmitter)와 같은 직렬 통신 인터페이스에서
 *      > 통신 속도(보레이트, baud rate)를 설정하기 위한 분주비(divisor) 값을 저장하는 레지스터입니다.
 *      >
 *      > 고정된 입력 클럭 주파수(예: 14.7456 MHz)를 원하는 통신 속도로 나누어 주기 위한
 *      > **나눗셈 값(divisor)**을 저장합니다. 이 분주비를 통해 최종적인 보레이트가 결정됩니다.
 *      >
 *      > 16비트 분주 값을 저장하기 위해 하위 8비트 레지스터인 DLL(Divisor Latch Low)과
 *      > 상위 8비트 레지스터인 DLH(Divisor Latch High) 두 개로 나누어져 있습니다.
 *
 *      16550 설계는 주소 공간이 좁아서, 각 레지스터에 고정된 주소 슬롯을 줄 수 없었습니다.
 *      대신 같은 주소에 상황에 따라 다른 레지스터가 겹쳐서 들어갑니다. 예를 들어,
 *
 *      - offset 0 (UART_BASE + 0)
 *          - DLAB = 0: THR(쓰기) / RBR(읽기)
 *          - DLAB = 1: DLL(분주 하위 8비트)
 *      - offset 1 (UART_BASE + 1)
 *          - DLAB = 0: IER(인터럽트 enable)
 *          - DLAB = 1: DLH(분주 상위 8비트)
 *
 *      이렇게 한 주소에 두 기능이 겹쳐 있으니, 지금은 분주 값을 만지는 중이라는 것을 하드웨어에게 알려 주려면
 *      LCR의 최상위 비트인 DLAB(0x80)을 1로 세트해야 합니다.
 *      그러면 같은 주소에서도 THR/IER 대신 DLL/DLH에 접근할 수 있습니다.
 *      필요한 값을 쓰고 나면 DLAB을 다 시 0으로 내려, 원래대로 THR(송신), IER(인터럽트) 쓰기가 가능하도록 돌려놓습니다.
 *
 *      UART 하드웨어는 내부적으로 일정한 기준 클럭(예: 1.8432 MHz)을 받습니다.
 *      이 클럭을 곧바로 직렬선으로 내보내면 속도가 너무 빠르니, 사용자(커널)가 셋업한 분주 값(divisor)으로 나눠서
 *      원하는 전송 속도(예: 115200 bps, 38400 bps)를 만듭니다:
 *
 *      baud rate = input_clock / (16 * divisor)
 *
 *      - input_clock: 보드가 UART에 공급하는 기준 클럭(ex. 1.8432 MHz).
 *      - divisor: DLL/DLH 두 바이트로 지정하는 16비트 값.
 *      - 16은 UART 내부에서 16배 초과 샘플링을 하기 때문에 포함된 상수입니다.
 *
 *      예를 들어, divisor에 3을 넣으면 1,843,200 / (16 * 3) = 38,400 bps가 됩니다.
 *
 *      - LCR = 0x03: DLAB=0으로 DLAB를 끄고 8N1(8bit, no parity, 1 stop)을 확정합니다.
 *      - FCR = 0x07: FIFO를 켜고(clear) 깨끗한 상태에서 시작합니다.
 *
 *  2. uart_puts_sync('a')는
 *      - LSR(offset 5)를 계속 읽어 THRE(0x20) 비트가 1일 때까지 확인합니다.(polling)
 *      - 1이 되면 THR(offset 0)에 0x61(‘a’)를 기록합니다.
 *
 *  3. 하드웨어는 'THR => TSR'(직렬선, Serial line)로 문자를 밀어 내고, QEMU가 표준 출력으로 연결해 터미널에 글자를 보여 줍니다.
 *      - "THR 내용이 TSR로 로드되고, TSR이 TXD 핀으로 비트를 shift-out" 합니다.
 *      - 직렬선(serial line)은 "TXD 핀을 따라 흘러가는 비트들의 경로"를 비유적으로 부르는 표현입니다.
 *      - TSR(Transmitter Shift Register)은 THR 바로 뒤에 연결된 하드웨어 내부 레지스터입니다.
 *      - THR에 값이 들어오고 TSR이 비어 있으면, UART는 THR 내용을 TSR로 복사한 뒤 THR을 비웁니다.
 *        이때 LSR.THRE(Transmit Holding Register Empty, 0x20) 값이 다시 1이 됩니다
 *      - TSR은 데이터 비트들을 한 비트씩 왼쪽으로 밀어 내보내면서 실제 파이프라인(직렬 출력 핀)으로 흘려보냅니다.
 *        그래서 "Shift Register"라는 이름을 갖고 있습니다.
 *      - TSR이 비트를 밀어 넣는 실제 출력선은 UART 칩의 'TXD(Transmit Data) 핀'입니다.
 *        이 핀을 따라 시간이 지나면서 전압이 "Start bit => 데이터 8비트 => Stop bit" 순서로 바뀌고,
 *        그것이 바로 직렬 통신선을 통해 흘러가는 신호입니다
 *      - QEMU나 실제 보드에서는 이 'TXD 핀'과 호스트 PC의 시리얼 포트(혹은 콘솔)가 서로 연결돼 있어서
 *        신호가 들어오는 즉시 터미널로 글자가 찍히게 됩니다.
 */
#include <stdint.h>
#include "defs.h"

// 주소 배치:
// - QEMU RISC-V virt 보드에서 UART는 물리 주소 0x1000_0000에 메모리 매핑되어 있습니다.
// - 레지스터 간격(reg-shift)은 1바이트입니다. 따라서 실제 MMIO 주소는 `BASE + 오프셋`으로 계산합니다.
#define UART_BASE 0x10000000UL
#define REG(off)  ((volatile uint8_t *)(UART_BASE + (off)))

enum {
    // Transmitter Holding Register (write)
    // - offset 0
    // - 소프트웨어가 메모리 매핑된 주소에 쓰는 레지스터입니다.
    // - THR은 그 자체로 하드웨어가 바로 비트를 내보내는 장치는 아니고, "다음에 내보낼 문자"를 잠시 보관하는 버퍼입니다.
    // - 여기에 1바이트를 쓰면 UART가 직렬선으로 내보냅니다.
    THR = 0,
    // Interrupt Enable Register( write)
    // - offset 1
    // - 폴링으로만 동작하므로 0으로 두어 인터럽트를 끕니다.
    IER = 1,
    // FIFO Control Register (write)
    // - offset 2
    // - FIFO를 켜고(clear) 깨끗한 상태에서 시작하게 합니다.
    FCR = 2,
    // Line Control Register (write)
    // - offset 3
    // - DLAB 비트를 이용해 분주(baud divisor)를 설정하고, 8N1 형식을 확정합니다.
    // - 8N1: 8 data bits, no parity, 1 stop bit로 프레이밍을 정해 가장 단순한 기본 형식을 보장합니다.
    LCR = 3,
    // Line Status Register (read)
    // - offset 5
    // - THR(Transmitter Holding Register)이 비었는지(THRE bit) 확인하기 위해 폴링합니다.
    LSR = 5,
};

// 비트 정의
enum {
    LCR_BAUD_LATCH = 0x80, // DLAB bit: 1이면 DLL/DLH에 접근
    LCR_8N1        = 0x03, // 8 data bits, no parity, 1 stop bit
    FCR_FIFO_EN    = 0x01, // FIFO enable
    FCR_FIFO_CLR   = 0x06, // FIFO reset (송·수신 큐 비우기)
    LSR_THRE       = 0x20, // THR(Transmit Holding Register) Empty 상태()
};

/**
 * 메모리 매핑 I/O(MMIO) 8비트 레지스터 접근을 위한 공통 프리미티브.
 *
 * 현재는 `uint8_t`를 대상으로 해서 사실상 바이트 단위 MMIO이지만,
 * 32비트, 64비트 레지스터에 접근해야 할 경우 `mmio_read32`, `mmio_write32` 같은 함수를 정의합니다.
 */
static inline uint8_t mmio_read8(uintptr_t base, int reg_offset) {
    // base는 장치의 MMIO 베이스 주소(UART_BASE 등)입니다.
    // reg_offset은 장치 내 레지스터 오프셋입니다(THR=0, LSR=5 같은 값).
    // base+off는 CPU 관점에서 그저 "물리 주소"입니다.
    volatile uint8_t *addr = (volatile uint8_t *)(base + reg_offset);
    return *addr; // 이 load는 실제 하드웨어 레지스터 read로 컴파일됩니다.
}

/**
 * 메모리 매핑 I/O(MMIO) 8비트 레지스터 접근을 위한 공통 프리미티브.
 *
 * 현재는 `uint8_t`를 대상으로 해서 사실상 바이트 단위 MMIO이지만,
 * 32비트, 64비트 레지스터에 접근해야 할 경우 `mmio_read32`, `mmio_write32` 같은 함수를 정의합니다.
 *
 * 레지스터에 들어 있는 값들을 더해서 어떤 물리 주소를 만들고,
 * 그 주소에 store 인스트럭션 한 번을 실행합니다.
 * 즉, c 값을 들고 있는 레지스터와, `UART_BASE+0(=THR)` 주소를 들고 있는 레지스터를 조합해서,
 * 그 주소에 sb(store byte) 인스트럭션으로 바이트 하나를 씁니다.
 */
static inline void mmio_write8(uintptr_t base, int reg_offset, uint8_t val)  {
    // base + reg_offset 주소에 1바이트를 쓰는 동작입니다.
    // CPU는 단지 해당 주소에 store byte 명령을 실행할 뿐이고,
    // 메모리 맵에 따라 DRAM이든 UART 레지스터든 해당 하드웨어가 이를 소비합니다.
    volatile uint8_t *addr = (volatile uint8_t *)(base + reg_offset);
    *addr = val;  // 이 store가 제거되거나 합쳐지면 안 되기 때문에 volatile이 중요합니다.
}

/**
 * ns16550 compatible UART(Universal Asynchronous Receiver/Transmitter) 장치를 초기화합니다.
 *
 * 메모리 매핑된 I/O를 통해 UART0 장치 레지스터(UART_BASE ~ UART_BASE+..)에 적절한 값을 써 넣어서
 * 원하는 모드(폴링 + 8N1 + FIFO 사용)로 하드웨어를 구성합니다.
 *
 * - 모든 UART 인터럽트를 비활성화하여, 폴링 기반 TX만 사용합니다.
 * - DLAB(Divisor Latch Access Bit)을 1로 설정하여 분주 레지스터(DLL/DLH)에 접근한 뒤,
 *   원하는 baud rate(예: 38400bps)에 맞춰 분주기를 설정합니다.
 * - DLAB을 0으로 되돌리고, 데이터 프레임 형식을 8N1(8 data bits, no parity, 1 stop bit)로 고정합니다.
 * - FIFO를 enable하고, 남아 있을지 모르는 이전 데이터들을 flush 합니다.
 */
void uart_init(void) {
    // 1) 모든 UART 인터럽트 비활성:
    // 폴링만 사용할 것이므로 IER=0으로 초기화합니다.(모든 인터럽트 off)
    mmio_write8(UART_BASE, IER, 0x00);

    // 2) DLAB을 켜고 분주(baud divisor)를 설정합니다.
    // 기본 입력 클럭(1.8432MHz 가정)에서 0x0003은 38400bps에 해당합니다.
    mmio_write8(UART_BASE, LCR, LCR_BAUD_LATCH); // DLAB=1일 때 divisor latch 접근 모드

    // THR(=DLL, Divisor Latch Low), IER(=DLH, Divisor Latch High)에
    // 분주(divisor) 값을 써서 baud rate를 정합니다.
    //
    // DLL (LSB) — DLAB=1일 때 THR 주소가 DLL로 재해석됩니다.
    // (이 시점에서는 DLL에 3을 씀)
    mmio_write8(UART_BASE, THR, 0x03);

    // DLH (MSB) — DLAB=1일 때 IER 주소가 DLH로 재해석됩니다.
    // (DLH에 0을 씀 → divisor = 0x0003)
    mmio_write8(UART_BASE, IER, 0x00);

    // 3) DLAB을 끄고 8 데이터 비트, 패리티 없음, 1 스톱 비트(8N1)를 설정합니다.
    // (DLAB=0, 8N1 설정)
    mmio_write8(UART_BASE, LCR, LCR_8N1);

    // 4) FIFO를 활성화하고 즉시 비워 깨끗한 상태로 만듭니다.
    // FCR=0x07 (FIFO enable + clear)
    mmio_write8(UART_BASE, FCR, FCR_FIFO_EN | FCR_FIFO_CLR);
}

/**
 * 문자를 하나 동기(synchronous) 방식으로 전송합니다.
 *
 * 이 함수는 LSR(Line Status Register)의 THRE(Transmitter Holding Register Empty) 비트를
 * 폴링하면서, THR(Transmitter Holding Register)가 비어 있는 순간까지 busy-wait 합니다.
 *
 * 1. mmio_read8(UART_BASE, LSR)로 LSR을 읽습니다.
 * 2. LSR_THRE 비트가 0이면(THR에 이전 문자가 남아 있으면), 루프를 반복하며 기다립니다.
 * 3. LSR_THRE 비트가 1이 되는 순간, THR이 비었다는 뜻이므로 mmio_write8(UART_BASE, THR, c)를 호출합니다.
 * 4. 이 호출이 리턴되는 시점에는 "이 문자 c가 최소한 UART 내부 버퍼(THR/TSR)로 진입했다"는 것만 보장하며,
 *    실제 직렬선으로 물리적으로 전송이 모두 끝났는지는 보장하지 않습니다.
 *
 * 이 함수는 하드웨어가 THR을 비울 때까지 CPU를 바쁘게 돌리므로, 저수준 초기화나
 * 간단한 디버깅 출력(커널 패닉 메시지 등)에 적합합니다. 대량의 출력이나 높은 성능이 필요한
 * 경우에는 인터럽트 기반 또는 링 버퍼 기반 비동기 전송 루틴이 필요합니다.
 */
void uart_puts_sync(unsigned char c) {
    // 폴링 기반 busy-wait
    for (;;) {
        uint8_t lsr = mmio_read8(UART_BASE, LSR); // LSR 읽어서 하드웨어 상태를 확인합니다.
        if (lsr & LSR_THRE) {                     // THRE 비트가 1이면 THR이 비어 있는 상태입니다.
            break;                                // 이제 새 문자를 써도 안전합니다.
        }
        // THR이 비어있지 않으면 아무 것도 하지 않고 다시 LSR을 읽습니다.
    }

    mmio_write8(UART_BASE, THR, c); // 실제 문자 쓰기를 하는 부분으로, UART 내부 버퍼로 전달합니다.
}
