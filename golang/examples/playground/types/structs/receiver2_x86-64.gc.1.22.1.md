# receiver2 x86-64.gc.1.22.1

## 컴파일 시도한 코드

```go
package structs

import "fmt"

func Receiver2() {
    // 왜 동작하는가?
    // 메소드 호출 시 포인터 리시버가 필요한 경우, Go 컴파일러와 런타임은 자동으로 값을 포인터로 변환해줍니다.
    byPointerButValue := helloImplementedByPointer{}
    byPointerButValue.Greet()
    byPointerButValue.Greet2()
}

type Hello interface {
    Greet()
    Greet2()
}

type helloImplementedByValue struct{}

func (r helloImplementedByValue) Greet() {
    fmt.Println("helloByValue, Hello, World!")
}

func (r helloImplementedByValue) Greet2() {
    fmt.Println("helloByValue, Hello, Hello, World!")
}

type helloImplementedByPointer struct{}

func (r *helloImplementedByPointer) Greet() {
    fmt.Println("helloByPointer, Hello, World!")
}

func (r helloImplementedByPointer) Greet2() {
    fmt.Println("helloByPointer, Hello, Hello, World!")
}
```

## assembly

```assembly
command-line-arguments_Receiver2_pc0:
        TEXT    command-line-arguments.Receiver2(SB), ABIInternal, $80-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_Receiver2_pc150
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $72, SP
        FUNCDATA        $0, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        FUNCDATA        $1, gclocals·e2OAQw7RTI8D9/LnocWHCg==(SB)
        FUNCDATA        $2, command-line-arguments.Receiver2.stkobj(SB)
        XCHGL   AX, AX
        MOVUPS  X15, command-line-arguments..autotmp_16+56(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_16+56(SP)
        LEAQ    command-line-arguments..stmp_0(SB), R8
        MOVQ    R8, command-line-arguments..autotmp_16+64(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_16+56(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $0
        CALL    fmt.Fprintln(SB)
        XCHGL   AX, AX
        MOVUPS  X15, command-line-arguments..autotmp_18+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_18+40(SP)
        LEAQ    command-line-arguments..stmp_1(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_18+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_18+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        CALL    fmt.Fprintln(SB)
        ADDQ    $72, SP
        POPQ    BP
        RET
command-line-arguments_Receiver2_pc150:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        NOP
        JMP     command-line-arguments_Receiver2_pc0
command-line-arguments_helloImplementedByValue_Greet_pc0:
        TEXT    command-line-arguments.helloImplementedByValue.Greet(SB), ABIInternal, $64-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByValue_Greet_pc82
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        FUNCDATA        $0, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        FUNCDATA        $1, gclocals·EaPwxsZ75yY1hHMVZLmk6g==(SB)
        FUNCDATA        $2, command-line-arguments.helloImplementedByValue.Greet.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.helloImplementedByValue.Greet.arginfo1(SB)
        MOVUPS  X15, command-line-arguments..autotmp_9+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+40(SP)
        LEAQ    command-line-arguments..stmp_2(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_9+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $0
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        RET
command-line-arguments_helloImplementedByValue_Greet_pc82:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        JMP     command-line-arguments_helloImplementedByValue_Greet_pc0
command-line-arguments_helloImplementedByValue_Greet2_pc0:
        TEXT    command-line-arguments.helloImplementedByValue.Greet2(SB), ABIInternal, $64-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByValue_Greet2_pc82
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        FUNCDATA        $0, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        FUNCDATA        $1, gclocals·EaPwxsZ75yY1hHMVZLmk6g==(SB)
        FUNCDATA        $2, command-line-arguments.helloImplementedByValue.Greet2.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.helloImplementedByValue.Greet2.arginfo1(SB)
        MOVUPS  X15, command-line-arguments..autotmp_9+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+40(SP)
        LEAQ    command-line-arguments..stmp_3(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_9+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $0
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        RET
command-line-arguments_helloImplementedByValue_Greet2_pc82:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        JMP     command-line-arguments_helloImplementedByValue_Greet2_pc0
command-line-arguments_helloImplementedByPointer_Greet_pc0:
        TEXT    command-line-arguments.(*helloImplementedByPointer).Greet(SB), ABIInternal, $64-8
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByPointer_Greet_pc82
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        FUNCDATA        $0, gclocals·Plqv2ff52JtlYaDd2Rwxbg==(SB)
        FUNCDATA        $1, gclocals·EaPwxsZ75yY1hHMVZLmk6g==(SB)
        FUNCDATA        $2, command-line-arguments.(*helloImplementedByPointer).Greet.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.(*helloImplementedByPointer).Greet.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.(*helloImplementedByPointer).Greet.argliveinfo(SB)
        PCDATA  $3, $1
        MOVUPS  X15, command-line-arguments..autotmp_9+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+40(SP)
        LEAQ    command-line-arguments..stmp_4(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_9+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $0
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        RET
command-line-arguments_helloImplementedByPointer_Greet_pc82:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        JMP     command-line-arguments_helloImplementedByPointer_Greet_pc0
command-line-arguments_helloImplementedByPointer_Greet2_pc0:
        TEXT    command-line-arguments.helloImplementedByPointer.Greet2(SB), ABIInternal, $64-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByPointer_Greet2_pc82
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        FUNCDATA        $0, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        FUNCDATA        $1, gclocals·EaPwxsZ75yY1hHMVZLmk6g==(SB)
        FUNCDATA        $2, command-line-arguments.helloImplementedByPointer.Greet2.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.helloImplementedByPointer.Greet2.arginfo1(SB)
        MOVUPS  X15, command-line-arguments..autotmp_9+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+40(SP)
        LEAQ    command-line-arguments..stmp_5(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_9+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_9+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $0
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        RET
command-line-arguments_helloImplementedByPointer_Greet2_pc82:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        JMP     command-line-arguments_helloImplementedByPointer_Greet2_pc0
command-line-arguments_Hello_Greet_pc0:
        TEXT    command-line-arguments.Hello.Greet(SB), DUPOK|WRAPPER|ABIInternal, $16-16
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_Hello_Greet_pc48
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $8, SP
        MOVQ    32(R14), R12
        TESTQ   R12, R12
        JNE     command-line-arguments_Hello_Greet_pc75
command-line-arguments_Hello_Greet_pc23:
        NOP
        MOVQ    AX, command-line-arguments.~p0+24(FP)
        MOVQ    BX, command-line-arguments.~p0+32(FP)
        FUNCDATA        $0, gclocals·IuErl7MOXaHVn7EZYWzfFA==(SB)
        FUNCDATA        $1, gclocals·J5F+7Qw7O7ve2QcWC7DpeQ==(SB)
        FUNCDATA        $5, command-line-arguments.Hello.Greet.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.Hello.Greet.argliveinfo(SB)
        PCDATA  $3, $1
        MOVQ    24(AX), CX
        MOVQ    BX, AX
        PCDATA  $1, $1
        CALL    CX
        ADDQ    $8, SP
        POPQ    BP
        RET
command-line-arguments_Hello_Greet_pc48:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        MOVQ    AX, 8(SP)
        MOVQ    BX, 16(SP)
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        MOVQ    8(SP), AX
        MOVQ    16(SP), BX
        JMP     command-line-arguments_Hello_Greet_pc0
command-line-arguments_Hello_Greet_pc75:
        LEAQ    24(SP), R13
        CMPQ    (R12), R13
        JNE     command-line-arguments_Hello_Greet_pc23
        MOVQ    SP, (R12)
        JMP     command-line-arguments_Hello_Greet_pc23
command-line-arguments_Hello_Greet2_pc0:
        TEXT    command-line-arguments.Hello.Greet2(SB), DUPOK|WRAPPER|ABIInternal, $16-16
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_Hello_Greet2_pc48
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $8, SP
        MOVQ    32(R14), R12
        TESTQ   R12, R12
        JNE     command-line-arguments_Hello_Greet2_pc75
command-line-arguments_Hello_Greet2_pc23:
        NOP
        MOVQ    AX, command-line-arguments.~p0+24(FP)
        MOVQ    BX, command-line-arguments.~p0+32(FP)
        FUNCDATA        $0, gclocals·IuErl7MOXaHVn7EZYWzfFA==(SB)
        FUNCDATA        $1, gclocals·J5F+7Qw7O7ve2QcWC7DpeQ==(SB)
        FUNCDATA        $5, command-line-arguments.Hello.Greet2.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.Hello.Greet2.argliveinfo(SB)
        PCDATA  $3, $1
        MOVQ    32(AX), CX
        MOVQ    BX, AX
        PCDATA  $1, $1
        CALL    CX
        ADDQ    $8, SP
        POPQ    BP
        RET
command-line-arguments_Hello_Greet2_pc48:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        MOVQ    AX, 8(SP)
        MOVQ    BX, 16(SP)
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        MOVQ    8(SP), AX
        MOVQ    16(SP), BX
        JMP     command-line-arguments_Hello_Greet2_pc0
command-line-arguments_Hello_Greet2_pc75:
        LEAQ    24(SP), R13
        CMPQ    (R12), R13
        JNE     command-line-arguments_Hello_Greet2_pc23
        MOVQ    SP, (R12)
        JMP     command-line-arguments_Hello_Greet2_pc23
command-line-arguments_helloImplementedByValue_Greet_pc0_1:
        TEXT    command-line-arguments.(*helloImplementedByValue).Greet(SB), DUPOK|WRAPPER|ABIInternal, $64-8
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByValue_Greet_pc103_1
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        MOVQ    32(R14), R12
        TESTQ   R12, R12
        JNE     command-line-arguments_helloImplementedByValue_Greet_pc120_1
command-line-arguments_helloImplementedByValue_Greet_pc23_1:
        NOP
        FUNCDATA        $0, gclocals·wgcWObbY2HYnK2SU/U22lA==(SB)
        FUNCDATA        $1, gclocals·AzW08EQV0LVfnDEAZer1Nw==(SB)
        FUNCDATA        $2, command-line-arguments.(*helloImplementedByValue).Greet.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.(*helloImplementedByValue).Greet.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.(*helloImplementedByValue).Greet.argliveinfo(SB)
        PCDATA  $3, $1
        TESTQ   AX, AX
        JEQ     command-line-arguments_helloImplementedByValue_Greet_pc97_1
        NOP
        MOVUPS  X15, command-line-arguments..autotmp_10+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+40(SP)
        LEAQ    command-line-arguments..stmp_6(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_10+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $1
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        NOP
        RET
command-line-arguments_helloImplementedByValue_Greet_pc97_1:
        CALL    runtime.panicwrap(SB)
        XCHGL   AX, AX
command-line-arguments_helloImplementedByValue_Greet_pc103_1:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        MOVQ    AX, 8(SP)
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        MOVQ    8(SP), AX
        JMP     command-line-arguments_helloImplementedByValue_Greet_pc0_1
command-line-arguments_helloImplementedByValue_Greet_pc120_1:
        LEAQ    72(SP), R13
        NOP
        CMPQ    (R12), R13
        JNE     command-line-arguments_helloImplementedByValue_Greet_pc23_1
        MOVQ    SP, (R12)
        JMP     command-line-arguments_helloImplementedByValue_Greet_pc23_1
command-line-arguments_helloImplementedByValue_Greet2_pc0_1:
        TEXT    command-line-arguments.(*helloImplementedByValue).Greet2(SB), DUPOK|WRAPPER|ABIInternal, $64-8
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByValue_Greet2_pc103_1
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        MOVQ    32(R14), R12
        TESTQ   R12, R12
        JNE     command-line-arguments_helloImplementedByValue_Greet2_pc120_1
command-line-arguments_helloImplementedByValue_Greet2_pc23_1:
        NOP
        FUNCDATA        $0, gclocals·wgcWObbY2HYnK2SU/U22lA==(SB)
        FUNCDATA        $1, gclocals·AzW08EQV0LVfnDEAZer1Nw==(SB)
        FUNCDATA        $2, command-line-arguments.(*helloImplementedByValue).Greet2.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.(*helloImplementedByValue).Greet2.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.(*helloImplementedByValue).Greet2.argliveinfo(SB)
        PCDATA  $3, $1
        TESTQ   AX, AX
        JEQ     command-line-arguments_helloImplementedByValue_Greet2_pc97_1
        NOP
        MOVUPS  X15, command-line-arguments..autotmp_10+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+40(SP)
        LEAQ    command-line-arguments..stmp_7(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_10+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $1
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        NOP
        RET
command-line-arguments_helloImplementedByValue_Greet2_pc97_1:
        CALL    runtime.panicwrap(SB)
        XCHGL   AX, AX
command-line-arguments_helloImplementedByValue_Greet2_pc103_1:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        MOVQ    AX, 8(SP)
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        MOVQ    8(SP), AX
        JMP     command-line-arguments_helloImplementedByValue_Greet2_pc0_1
command-line-arguments_helloImplementedByValue_Greet2_pc120_1:
        LEAQ    72(SP), R13
        NOP
        CMPQ    (R12), R13
        JNE     command-line-arguments_helloImplementedByValue_Greet2_pc23_1
        MOVQ    SP, (R12)
        JMP     command-line-arguments_helloImplementedByValue_Greet2_pc23_1
command-line-arguments_helloImplementedByPointer_Greet2_pc0_1:
        TEXT    command-line-arguments.(*helloImplementedByPointer).Greet2(SB), DUPOK|WRAPPER|ABIInternal, $64-8
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     command-line-arguments_helloImplementedByPointer_Greet2_pc103_1
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $56, SP
        MOVQ    32(R14), R12
        TESTQ   R12, R12
        JNE     command-line-arguments_helloImplementedByPointer_Greet2_pc120_1
command-line-arguments_helloImplementedByPointer_Greet2_pc23_1:
        NOP
        FUNCDATA        $0, gclocals·wgcWObbY2HYnK2SU/U22lA==(SB)
        FUNCDATA        $1, gclocals·AzW08EQV0LVfnDEAZer1Nw==(SB)
        FUNCDATA        $2, command-line-arguments.(*helloImplementedByPointer).Greet2.stkobj(SB)
        FUNCDATA        $5, command-line-arguments.(*helloImplementedByPointer).Greet2.arginfo1(SB)
        FUNCDATA        $6, command-line-arguments.(*helloImplementedByPointer).Greet2.argliveinfo(SB)
        PCDATA  $3, $1
        TESTQ   AX, AX
        JEQ     command-line-arguments_helloImplementedByPointer_Greet2_pc97_1
        NOP
        MOVUPS  X15, command-line-arguments..autotmp_10+40(SP)
        LEAQ    type:string(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+40(SP)
        LEAQ    command-line-arguments..stmp_8(SB), DX
        MOVQ    DX, command-line-arguments..autotmp_10+48(SP)
        MOVQ    os.Stdout(SB), BX
        NOP
        LEAQ    go:itab.*os.File,io.Writer(SB), AX
        LEAQ    command-line-arguments..autotmp_10+40(SP), CX
        MOVL    $1, DI
        MOVQ    DI, SI
        PCDATA  $1, $1
        CALL    fmt.Fprintln(SB)
        ADDQ    $56, SP
        POPQ    BP
        NOP
        RET
command-line-arguments_helloImplementedByPointer_Greet2_pc97_1:
        CALL    runtime.panicwrap(SB)
        XCHGL   AX, AX
command-line-arguments_helloImplementedByPointer_Greet2_pc103_1:
        NOP
        PCDATA  $1, $-1
        PCDATA  $0, $-2
        MOVQ    AX, 8(SP)
        CALL    runtime.morestack_noctxt(SB)
        PCDATA  $0, $-1
        MOVQ    8(SP), AX
        JMP     command-line-arguments_helloImplementedByPointer_Greet2_pc0_1
command-line-arguments_helloImplementedByPointer_Greet2_pc120_1:
        LEAQ    72(SP), R13
        NOP
        CMPQ    (R12), R13
        JNE     command-line-arguments_helloImplementedByPointer_Greet2_pc23_1
        MOVQ    SP, (R12)
        JMP     command-line-arguments_helloImplementedByPointer_Greet2_pc23_1
```

1. **함수 프롤로그와 에필로그**

    각 함수의 시작 부분에는 스택 프레임을 설정하는 명령어(`SUBQ`로 스택 할당, `PUSHQ`와 `POPQ`로 베이스 포인터 조작)가 있고, 함수 종료 시에는 이를 해제하는 명령어(`ADDQ`로 스택 해제, `RET`으로 반환)가 있습니다.

    이러한 부분은 함수 호출의 안정성을 보장합니다.

2. **레지스터 사용**

    Go 언어는 함수 호출 시 매개변수를 전달하고 결과를 반환하는 데 레지스터를 사용합니다.

    예를 들어, `CALL fmt.Fprintln(SB)`에서는 `DX`, `CX`, `BX`, `AX` 레지스터를 사용하여 필요한 인자를 전달합니다.

3. **자동 리시버 변환**

    메소드 호출 시 리시버 타입이 값 타입에서 포인터 타입으로 필요한 경우, 컴파일러는 `LEAQ` 명령어를 사용하여 리시버의 주소를 계산하고 이를 적절한 레지스터(`CX` 등)에 할당합니다.

    예를 들어, `helloImplementedByPointer.Greet` 메소드의 호출에서 이러한 변환이 발생할 수 있습니다.

코드 조각에서 직접적인 자동 변환의 예를 찾는 것은 어셈블리 코드만으로는 명확하지 않습니다.
그러나 일반적으로 Go에서 이러한 변환은 내부적으로 발생하며, 특정 레지스터에 객체의 주소를 로딩(`LEAQ`)하는 명령어를 통해 확인할 수 있습니다. 이 과정에서 `LEAQ` 명령어는 특정 객체의 주소를 취하고, 이 주소를 메소드 호출 시 사용합니다.

어셈블리 코드에서 `helloImplementedByPointer.Greet`와 `helloImplementedByPointer.Greet2`의 구현을 보면, 값 리시버와 포인터 리시버를 처리하는 방식의 차이를 알 수 있습니다.

포인터 리시버의 경우, 메소드를 호출하기 전에 객체의 주소를 확보해야 하며, 값 리시버는 해당 객체의 복사본을 직접 사용합니다.

어셈블리 코드만으로는 모든 디테일을 완벽하게 설명하기 어렵습니다만, 주어진 코드는 Go의 컴파일 과정에서 리시버의 자동 변환과 메소드 호출의 처리 방식을 일부 설명해 줍니다.

Go 컴파일러는 효율적인 실행을 위해 레지스터를 활용하며, 필요에 따라 메모리 주소 계산을 자동으로 수행하여 메소드 호출의 안정성과 일관성을 보장합니다.
