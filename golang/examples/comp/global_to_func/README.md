# Global variable to function

## 1. 초기화 과정

1. **실행 시점**: globalVar는 프로그램 시작 시 한 번만 초기화되고, getGlobalVar는 호출될 때마다 초기화됩니다.
2. **메모리 할당**: globalVar는 정적 메모리에 할당되고, getGlobalVar는 매번 새로운 힙 메모리를 사용합니다.
3. **성능**: globalVar 접근은 단순 메모리 참조로 빠르지만, getGlobalVar는 함수 호출과 맵 생성 오버헤드가 있습니다.
4. **동시성**: globalVar는 공유 자원이 되어 동시성 이슈가 발생할 수 있지만, getGlobalVar는 매번 새로운 인스턴스를 생성하여 동시성 안전합니다.

### 전역 변수 (globalVar)

- `main.map.init.0` 함수에서 한 번만 초기화됩니다.
- 이 함수는 프로그램 시작 시 `main.init` 함수에 의해 호출됩니다.

```assembly
main_init_pc0:
        TEXT    main.init(SB), PKGINIT|ABIInternal, $8-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     main_init_pc17
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        FUNCDATA        $0, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        FUNCDATA        $1, gclocals·g2BeySu+wFnoycgXfElmcg==(SB)
        PCDATA  $1, $0
        CALL    main.map.init.0(SB)
        POPQ    BP
        RET
```

- `main_init_pc0`: 이는 프로그램 시작 시 자동으로 호출되는 초기화 함수입니다.
- 스택 프레임을 설정

    > [스택 프레임](../../../../computer_architecture/stack_frame.md)?
    >
    > 함수 호출 시 생성되는 메모리 영역으로, 함수의 지역 변수, 매개변수, 반환 주소 등을 저장하는 데 사용됩니다.

- `CALL main.map.init.0(SB)`: `globalVar`를 초기화하는 함수를 호출합니다.

아래는 `globalVar`를 초기화하는 함수 `main.map.init.0`의 어셈블리 코드입니다.

```assembly
main_map_init_0_pc0:
        TEXT    main.map.init.0(SB), ABIInternal, $56-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     main_map_init_0_pc417
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $48, SP
        FUNCDATA        $0, gclocals·ykHN0vawYuq1dUW4zEe2gA==(SB)
        FUNCDATA        $1, gclocals·sQxO+jiYy+d9ldxoWSePwQ==(SB)
        PCDATA  $1, $0
        CALL    runtime.makemap_small(SB)
        MOVQ    AX, main..autotmp_10+40(SP)
        LEAQ    type:main.Simple(SB), AX
        PCDATA  $1, $1
        CALL    runtime.newobject(SB)
        MOVQ    AX, main..autotmp_11+32(SP)
        MOVQ    $3, 8(AX)
        LEAQ    go:string."001"(SB), CX
        MOVQ    CX, (AX)
        MOVQ    $10, 24(AX)
        LEAQ    go:string."Some Name1"(SB), DX
        MOVQ    DX, 16(AX)
        MOVQ    main..autotmp_10+40(SP), BX
        MOVL    $3, DI
        LEAQ    type:map[string]*main.Simple(SB), AX
        PCDATA  $1, $2
        CALL    runtime.mapassign_faststr(SB)
        // ... (반복되는 코드 생략)
        MOVQ    DX, main.globalVar(SB)
        PCDATA  $0, $-1
        ADDQ    $48, SP
        POPQ    BP
        RET
```

1. `CALL runtime.makemap_small(SB)`: 작은 맵을 생성합니다.
2. `CALL runtime.newobject(SB)`: Simple 객체를 생성합니다.
3. `LEAQ` 및 `MOVQ` 명령어들: Simple 객체의 필드를 초기화합니다.
4. `CALL runtime.mapassign_faststr(SB)`: 맵에 키-값 쌍을 할당합니다.
5. `MOVQ DX, main.globalVar(SB)`: 생성된 맵을 globalVar에 할당합니다.

### 함수 (getGlobalVar)

```assembly
main_getGlobalVar_pc0:
        TEXT    main.getGlobalVar(SB), ABIInternal, $56-0
        CMPQ    SP, 16(R14)
        PCDATA  $0, $-2
        JLS     main_getGlobalVar_pc390
        PCDATA  $0, $-1
        PUSHQ   BP
        MOVQ    SP, BP
        SUBQ    $48, SP
        FUNCDATA        $0, gclocals·ykHN0vawYuq1dUW4zEe2gA==(SB)
        FUNCDATA        $1, gclocals·sQxO+jiYy+d9ldxoWSePwQ==(SB)
        PCDATA  $1, $0
        CALL    runtime.makemap_small(SB)
        MOVQ    AX, main..autotmp_11+40(SP)
        LEAQ    type:main.Simple(SB), AX
        PCDATA  $1, $1
        CALL    runtime.newobject(SB)
        MOVQ    AX, main..autotmp_12+32(SP)
        MOVQ    $3, 8(AX)
        LEAQ    go:string."001"(SB), CX
        MOVQ    CX, (AX)
        MOVQ    $10, 24(AX)
        LEAQ    go:string."Some Name1"(SB), DX
        MOVQ    DX, 16(AX)
        MOVQ    main..autotmp_11+40(SP), BX
        MOVL    $3, DI
        LEAQ    type:map[string]*main.Simple(SB), AX
        PCDATA  $1, $2
        CALL    runtime.mapassign_faststr(SB)
        // ... (반복되는 코드 생략)
        MOVQ    main..autotmp_11+40(SP), AX
        ADDQ    $48, SP
        POPQ    BP
        RET
```

1. 함수가 호출될 때마다 이 코드가 실행됩니다.
2. 함수가 호출될 때마다 새로운 맵을 생성합니다.

    `main.getGlobalVar` 함수 내에서 매번 초기화가 이루어집니다:

    ```assembly
    main_getGlobalVar_pc0:
        CALL    runtime.makemap_small(SB)
    ```

3. 이후 과정은 globalVar 초기화와 유사하지만, *함수 호출마다 반복*됩니다.
4. 마지막에 생성된 맵을 반환합니다 (`MOVQ main..autotmp_11+40(SP), AX`).

## 2. 메모리 할당

### 전역 변수

- `runtime.makemap_small`을 프로그램 시작 시 한 번만 호출합니다.

```assembly
main_map_init_0_pc0:
    CALL    runtime.makemap_small(SB)
    MOVQ    AX, main..autotmp_10+40(SP)
```

### 함수

- `getGlobalVar` 함수가 호출될 때마다 `runtime.makemap_small`을 호출합니다.

```assembly
main_getGlobalVar_pc0:
    CALL    runtime.makemap_small(SB)
    MOVQ    AX, main..autotmp_11+40(SP)
```

## 3. 맵 요소 할당

### 전역 변수

- 초기화 시 한 번만 `runtime.mapassign_faststr`을 호출하여 요소를 할당합니다.

### 함수

- 함수가 호출될 때마다 `runtime.mapassign_faststr`을 호출하여 요소를 새로 할당합니다.

두 경우 모두 다음과 같은 패턴을 보입니다:

```assembly
CALL    runtime.newobject(SB)
...
CALL    runtime.mapassign_faststr(SB)
```

## 4. 메인 함수에서의 사용

### 전역 변수

- `main.globalVar(SB)`를 직접 참조합니다.

```assembly
MOVQ    main.globalVar(SB), DX
```

### 함수

- `main.getGlobalVar` 함수를 호출하고 그 결과를 사용합니다.

```assembly
CALL    runtime.makemap_small(SB)
MOVQ    AX, main.~r0+56(SP)
```

## 5. 성능 및 리소스 사용 분석

### 전역 변수

1. **메모리 사용**: 프로그램 실행 중 지속적으로 메모리를 점유합니다.
2. **초기화 비용**: 프로그램 시작 시 한 번만 초기화 비용이 발생합니다.
3. **접근 속도**: 직접 메모리 주소를 참조하므로 매우 빠릅니다.
4. **동시성**: 여러 고루틴에서 동시에 접근할 경우 동기화 문제가 발생할 수 있습니다.

### 함수

1. **메모리 사용**: 함수 호출 시에만 일시적으로 메모리를 사용합니다.
2. **초기화 비용**: 매 호출마다 초기화 비용이 발생합니다.
3. **접근 속도**: 함수 호출 오버헤드와 맵 생성 비용으로 인해 상대적으로 느립니다.
4. **동시성**: 각 호출마다 새로운 인스턴스를 생성하므로 동시성 문제가 없습니다.

## 6. 가비지 컬렉션 영향

### 전역 변수

- 프로그램 종료 시까지 가비지 컬렉션 대상이 되지 않습니다.

### 함수

- 함수 호출이 끝나면 생성된 맵이 가비지 컬렉션 대상이 됩니다.
- 이는 `main_main` 함수에서 여러 번의 `runtime.makemap_small` 호출로 확인할 수 있습니다.

## 7. 코드 최적화 가능성

### 전역 변수

- 컴파일러가 전역 변수의 사용 패턴을 분석하여 최적화할 가능성이 있습니다.
- 예를 들어, 반복적인 접근을 레지스터에 캐싱하는 등의 최적화가 가능합니다.

### 함수

- 매번 새로운 인스턴스를 생성하므로 컴파일러의 최적화 여지가 상대적으로 적습니다.
- 단, 함수 인라이닝 등의 최적화는 가능할 수 있습니다.

## 8. 스택 사용

### 전역 변수

- 스택 사용이 최소화됩니다. 주로 포인터 참조만 이루어집니다.

### 함수

- 함수 호출마다 새로운 스택 프레임이 생성됩니다.
- `main_getGlobalVar_pc0`에서 볼 수 있듯이, 48바이트의 추가 스택 공간을 사용합니다.

```assembly
SUBQ    $48, SP
```

## 9. 코드 크기 및 복잡성

### 전역 변수

- 초기화 코드(`main.map.init.0`)가 별도로 존재하여 전체 바이너리 크기가 증가할 수 있습니다.
- 사용 시 코드가 간단합니다 (직접 참조).

### 함수

- 매번 맵을 생성하고 초기화하는 코드가 함수 내에 포함되어 있어, 함수 크기가 상대적으로 큽니다.
- 사용 시 함수 호출 코드가 필요합니다.

## 결론

전역 변수 방식:
- 초기 로딩 시간이 조금 더 걸리지만, 런타임 성능이 우수하고 메모리 사용이 일정합니다.
- 정적 분석과 최적화에 유리합니다.

함수 방식:
- 초기 로딩은 빠르지만 호출 시마다 오버헤드가 있고 메모리 사용이 동적입니다.
- 동시성과 메모리 관리에 장점이 있습니다.
