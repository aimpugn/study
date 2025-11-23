# xv6-riscv

- [xv6-riscv](#xv6-riscv)
    - [전체적인 흐름](#전체적인-흐름)
    - [When macOS](#when-macos)
        - [Xcode CLI Tools](#xcode-cli-tools)
        - [RISC-V 크로스 컴파일러, GDB, QEMU 설치](#risc-v-크로스-컴파일러-gdb-qemu-설치)

## 전체적인 흐름

- `Makefile`의 QEMU 명령은 QEMU virt 보드를 띄우고, [kernel.elf](./kernel/kernel.elf)를 DRAM 0x8000_0000에 올려 `ENTRY` 주소에서 실행을 시작합니다.
- [kernel.ld](./kernel/kernel.ld)는 ELF 안에서 `.text`, `.data`, `.bss` 등을 이 주소에 맞게 배치하고, BSS 범위에 이름을 붙여 C 코드에서 참조할 수 있게 합니다.
- [entry.S](./kernel/entry.S)는 C 코드가 돌기 전에 RISC-V psABI가 요구하는 초기 상태를 직접 만들어 주는 코드입니다. 각 hart의 스택을 준비한 뒤 [start.c](./kernel/start.c)의 `start()`를 호출하고, `start()`는 BSS를 0으로 채운 뒤 [main.c](./kernel/main.c)의 `main()`을 호출합니다.
- `main()`은 처음으로 현실 세계와 상호작용하는 코드로서 `uart_init()`으로 UART를 초기화하고, `uart_puts_sync()`를 반복 호출하여 문자를 `THR` 레지스터에 써 넣습니다.
- [printf.c](./kernel/printf.c) 안의 `mmio_write`와 `mmio_read`는 단지 메모리 주소에 대한 store/load 인스트럭션 묶음이고, QEMU는 그 주소들을 DRAM이나 UART 모델의 레지스터로 매핑해 줍니다.
- CPU 입장에서는 언제나 인스트럭션을 fetch하고 execute할 뿐이고, 그 인스트럭션들이 어떤 장치와 연결되느냐를 결정하는 것은 링크 스크립트와 QEMU의 메모리 맵입니다.

## When macOS

### Xcode CLI Tools

```sh
xcode-select --install
```

### RISC-V 크로스 컴파일러, GDB, QEMU 설치

```sh
brew install riscv64-elf-gcc riscv64-elf-gdb qemu \
    riscv64-elf-ld riscv64-elf-as riscv64-elf-objcopy
```

- ELF(Executable and Linkable Format)?

    [실행 파일, 오브젝트 파일, 공유 라이브러리와 코어 덤프를 위한 표준 파일 포맷](https://junsoolee.gitbook.io/linux-insides-ko/summary/theory/linux-theory-2)
