# xv6-riscv

- [xv6-riscv](#xv6-riscv)
    - [When macOS](#when-macos)
        - [Xcode CLI Tools](#xcode-cli-tools)
        - [RISC-V 크로스 컴파일러, GDB, QEMU 설치](#risc-v-크로스-컴파일러-gdb-qemu-설치)

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
