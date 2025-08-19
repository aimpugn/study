// 이 파일은 함수의 선언(Declaration)이 있는 곳입니다.
// C언어에서 프로젝트가 여러 .c 파일로 나뉘게 되면, 각 파일은 독립적으로 컴파일됩니다.
// 그래서 main.c를 컴파일할 때 컴파일러는 uart_puts_sync라는 함수가 실제로 어떻게 구현되었는지 전혀 모르고,
// 그저 어딘가에 존재한다고 믿고 컴파일을 진행할 뿐입니다.
// defs.h와 같은 헤더 파일은 어딘가에 구현되어 있을 함수들의 명세서 또는 목차와 같습니다.

// << printf.c >>
// 문자열의 주소(const char)를 인자로 받습니다.
void puts_sync(const char* s);

// 'a' 등은 문자(char)이고, 이런 단일 문자는 보통 정수(int)로 취급합니다.
void uart_puts_sync(unsigned char c);
