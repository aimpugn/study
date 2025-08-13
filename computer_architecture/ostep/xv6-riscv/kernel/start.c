extern void main(void);
extern char edata[], end[]; // kernel.ld가 제공

void start(void) {
    // -bios none 경우 펌웨어가 로드되지 않으므로 펌웨어나 런타임이 BSS 초기화를 대신해주지 않습니다.
    // 따라서 코드가 C로 진입하는 순간부터 링커 스크립트가 노출하는 섹션 심볼을 이용해
    // 전역, 정적 0 초기화(BSS 초기화)를 수행합니다.
    // 전역/정적 0 보장
    for (char *p = edata; p < end; p++) *p = 0;

    main();

    // main 끝나면 대기
    for(;;) __asm__ volatile("wfi");
}
