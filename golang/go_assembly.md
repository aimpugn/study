# Go compiled assembly

- [Go compiled assembly](#go-compiled-assembly)
    - [어셈블리 명령어 정리](#어셈블리-명령어-정리)
        - [**ADDQ**](#addq)
        - [**CALL**](#call)
        - [**CMPQ**](#cmpq)
        - [**FUNCDATA**](#funcdata)
        - [**JEQ**](#jeq)
        - [**JLS**](#jls)
        - [**JMP**](#jmp)
        - [**JNE**](#jne)
        - [**LEAQ**](#leaq)
        - [**MOVL**](#movl)
        - [**MOVQ**](#movq)
        - [**MOVUPS**](#movups)
        - [**NOP**](#nop)
        - [**PCDATA**](#pcdata)
        - [**POPQ**](#popq)
        - [**PUSHQ**](#pushq)
        - [**RET**](#ret)
        - [**SUBQ**](#subq)
        - [**TESTQ**](#testq)
        - [**TEXT**](#text)
        - [**XCHGL**](#xchgl)
    - [명령어 접미사의 의미](#명령어-접미사의-의미)
    - [어셈블리 코드의 목적](#어셈블리-코드의-목적)

## 어셈블리 명령어 정리

### **ADDQ**

"Add Quadword" - 두 개의 64비트 정수를 더하고 결과를 저장합니다.

### **CALL**

프로시저나 함수를 호출합니다. 호출된 함수의 주소로 점프하고, 반환 주소를 스택에 푸시합니다.

### **CMPQ**

"Compare Quadwords" - 두 64비트 정수를 비교합니다. 이 연산은 플래그 레지스터를 설정하지만 데이터를 저장하지는 않습니다.

### **FUNCDATA**

함수 데이터를 지정합니다. 일반적으로 컴파일러나 런타임에서 사용되는 메타데이터를 제공합니다.

### **JEQ**

"Jump if Equal" - 두 값이 같을 경우 지정된 위치로 점프합니다.

### **JLS**

"Jump if Less or Same" - 첫 번째 피연산자가 두 번째 피연산자보다 작거나 같을 경우 지정된 위치로 점프합니다.

### **JMP**

무조건적인 점프를 수행합니다. 지정된 주소나 레이블로 점프합니다.

### **JNE**

"Jump if Not Equal" - 두 값이 같지 않을 경우 지정된 위치로 점프합니다.

### **LEAQ**

"Load Effective Address Quadword" - `LEA`는 주소 계산을 목적으로 하며, 주소 연산의 결과를 레지스터에 로드합니다.

### **MOVL**

"Move Long" - 32비트 정수를 한 위치에서 다른 위치로 이동합니다.

### **MOVQ**

"Move Quadword" - 64비트 정수를 한 위치에서 다른 위치로 이동합니다.

### **MOVUPS**

"Move Unaligned Packed Single-Precision Floating-Point Values" - 정렬되지 않은 단정밀도 부동 소수점 값을 메모리에서 레지스터로, 또는 레지스터에서 메모리로 이동합니다.

### **NOP**

"No Operation" - 아무 작업도 수행하지 않는 명령어입니다. 프로그램의 특정 부분을 채우거나 지연 시간을 조절하는 데 사용됩니다.

### **PCDATA**

Pseudo-Code Data - 컴파일러나 디버거가 사용하는 메타데이터를 제공합니다. 주로 프로파일링이나 최적화에 사용되는 정보를 포함합니다.

### **POPQ**

"Pop Quadword" - 스택의 최상위 값을 64비트 레지스터나 메모리 위치로 팝합니다.

### **PUSHQ**

"Push Quadword" - 64비트(`Q`) 레지스터나 메모리 위치의 값을 스택에 푸시합니다.

### **RET**

"Return" - 함수에서 반환합니다. 스택에서 반환 주소를 팝하여 실행을 계속합니다.

### **SUBQ**

"Subtract Quadword" - 두 64비트 정수를 빼고 결과를 저장합니다.

### **TESTQ**

"Test Quadwords" - 두 64비트 정수를 AND 연산하고 결과에 따라 플래그를 설정합니다.

### **TEXT**

텍스트 세그먼트, 즉 코드 세그먼트를 정의하는 데 사용되는 지시어입니다. 이는 실행 가능 코드의 시작을 나타냅니다.

### **XCHGL**

"Exchange Long" - 두 32비트 레지스터의 내용을 교환합니다.

## 명령어 접미사의 의미

각 명령어의 접미사는 해당 데이터의 크기를 나타냅니다.

가령 접미사 `Q`는 64비트 길이의 데이터를 다루는 명령어임을 나타냅니다.
x86_64 아키텍처에서는 다음과 같은 접미사가 일반적으로 사용됩니다:
- **B** (Byte, 8 bits)
- **W** (Word, 16 bits)
- **D** (Doubleword, 32 bits)
- **Q** (Quadword, 64 bits)

이 접미사들은 명령어가 처리할 데이터의 크기를 명시적으로 표시하여, 명령어의 정확한 동작을 지정합니다.

## 어셈블리 코드의 목적

어셈블리 코드는 하드웨어 수준에서 프로그램의 동작을 제어하며, 메모리 관리, 입출력 제어, 복잡한 연산 실행 등의 작업을 가능하게 합니다. 어셈블리 언어를 이해하고 사용할 수 있다면, 시스템의 성능을 최적화하거나 특정 하드웨어 기능을 직접 제어하는 등 더 깊은 수준에서 프로그래밍이 가능합니다.