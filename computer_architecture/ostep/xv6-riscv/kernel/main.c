#include "defs.h"

void main(void) {

    uart_puts_sync('a');
    // puts_sync("hello\n");

    // 여기서 끝나면 start()가 WFI 루프로 들어간다.
}
