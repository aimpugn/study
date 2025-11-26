extern void main(void);
/**
 * `__bss_start__`와 `__bss_end__`는 kernel/kernel.ld에서 `PROVIDE(__bss_start__ = .);`와
 * `PROVIDE(__bss_end__ = .);`로 정의해 둔 심볼입니다.
 *
 * 링크 스크립트(kernel/kernel.ld)가 '.sdata까지 쌓고 난 뒤의 주소를 bss_start'라고 지정하고,
 * '.sbss와 .bss를 모두 쌓고 난 뒤의 주소를 bss_end'라고 지정했기 때문에,
 * 이 범위는 "초기값이 0이거나, 초기화를 하지 않은 전역/정적 변수들이 들어간 영역 전체"가 됩니다.
 */
extern char __bss_start__[], __bss_end__[];

#define MAXCPU 10


//
// sp(i) = S + (C * (i + 1)) 라고 할 때, `sp(i) % 16 == 0`는 항상 참이 됩니다.
//
// 한 코어(CPU, hart)마다 4KB짜리 스택을 잘라 쓰기 위해 큰 버퍼(4KB × MAXCPU)를 만들고,
// 그 버퍼의 시작 주소를 16바이트 경계에 맞춰 함수 호출 규약을 지킵니다.
// 코어(=hart)마다 스택이 따로 필요한데, 부팅 초기에 malloc 같은 것도 없으므로
// 전역 배열을 크게 확보해서 잘라서 사용합니다.
/**
 * 이 전역 배열은 초기값이 없으므로 .bss 섹션에 들어가게 되고,
 * __bss_start__ ~ __bss_end__ 범위 안에 stack0 전체(모든 hart의 스택)가 포함됩니다.
 *
 * 가령 bss 범위가 다음과 같으면 stack0는 그 사이 어딘가에 배치됩니다.
 * - __bss_start__ = 0x80000180
 * - __bss_end__   = 0x8000b000
 *
 * entry.S 파일에서 스택 포인터(sp, Stack Pointer)를 다음과 같이 설정하고 있습니다.
 *
 * la      sp, stack0
 * li      t1, 4096     // 4096
 * addi    t2, t0, 1    // hartid + 1
 * mul     t1, t1, t2   // 4096 * (hartid + 1)
 * add     sp, sp, t1   // &stack0 + (4096 * (hartid + 1))
 *
 * 즉, 현재 스택 포인터(sp)는 BSS 영역 안, `stack0` 배열 안 어딘가를 가리킵니다.
 *
 *  [BSS: __bss_start__ ... __bss_end__]
 *  ├─ (다른 전역들 …)
 *  ├─ stack0[0 .. 4096 * MAXCPU - 1]  ← 우리가 스택으로 쓰고 있는 배열
 *  │    ├─ hart0 스택 영역 ……… (sp가 여기쯤)
 *  │    └─ hart1, hart2 …
 *  └─ 그 밖의 BSS 전역 …
 *
 */
__attribute__((aligned(16))) char stack0[(256 * 16) * MAXCPU];

// entry.S jumps here in machine mode on stack0.
void start(void) {
    /**
     * entry.S에서 이미 BSS 제로화가 끝났다고 가정합니다.
     */
    main();

    // 끝까지 가도 반환하지 않고 wfi 루프에 들어갑니다.
    // 즉, start()가 _entry로 돌아가지 않습니다.
    for(;;) __asm__ volatile("wfi");
}

/**
 * 참고:
 *
 * C 언어는 초기화하지 않은 전역/정적 변수는 프로그램 시작 시 반드시 0으로 초기화되어야 한다는 규칙을 강하게 요구합니다.
 *
 * 일반적으로 유저 공간 프로그램에서는 OS 로더와 C 런타임 라이브러리(glibc 같은)가 이 일을 대신 해 줍니다.
 * 가령, 로더는 실행 파일(.exe, a.out)을 읽어 메모리에 올리고, C 코드가 시작되기 전에 필요한
 * 모든 환경(BSS 영역을 0으로 채우는 것 포함)을 준비해 줍니다.
 *
 * `-bios none` 경우 펌웨어가 로드되지 않으므로 펌웨어나 런타임이 BSS 초기화를 대신해주지 않습니다.
 * 커널을 도와줄 상위의 프로그램(상위 계층)이 없기 때문에 커널이 스스로 이 규칙을 만족시켜야 합니다.
 * BSS 영역(전역/정적 영역), 즉 `__bss_start__=0x80000180`부터 `__bss_end__=0x8000b000`까지의
 * 범위를 루프로 0으로 채웁니다.(BSS 제로화)
 *
 * 이때 `__bss_start__` 및 `__bss_end__`는 링커 스크립트가 노출하는 섹션 심볼입니다.
 *
 * `__bss_start__`부터 `__bss_end__`까지 한 칸씩 이동(*p++)하면서 0으로 초기화합니다.
 *
 * ```c
 * for (char *p = __bss_start__; p < __bss_end__; p++) {
 *     *p = 0;
 * }
 * // 또는
 * char *p = __bss_start__;
 * while (p < __bss_end__)
 *     *p++ = 0;
 * ```
 *
 * 만약 이런 과정을 거치지 않는다면, UART 상태 변수든, 스핀락이든, 전역 배열이든 모두 쓰레기 값에서 시작하게 됩니다.
 * 어떤 환경에서는 우연히 잘 돌아가다가도 다른 환경에서 부팅이 아예 안 되거나, 스핀락이 영원히 풀리지 않는 등의
 * 문제가 발생할 수 있습니다.
 */
