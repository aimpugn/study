# Zend Engine

- [Zend Engine](#zend-engine)
    - [PHP 실행 단계](#php-실행-단계)
    - [Zend Engine](#zend-engine-1)
        - [Zend Engine의 작동 원리](#zend-engine의-작동-원리)
        - [Zend Engine의 확장성과 모듈성](#zend-engine의-확장성과-모듈성)
    - [Zend Engine이 PHP 실행하는 방법](#zend-engine이-php-실행하는-방법)
    - [바이트코드와 기계어의 차이](#바이트코드와-기계어의-차이)
    - [Zend Engine Virtual Machine](#zend-engine-virtual-machine)
        - [Zend Engine의 메모리 관리](#zend-engine의-메모리-관리)
        - [Zend Engine의 `가비지 컬렉션` 코드 예시](#zend-engine의-가비지-컬렉션-코드-예시)
    - [PHP 스크립트 실행과 프로세스 관리](#php-스크립트-실행과-프로세스-관리)
        - [CGI 방식](#cgi-방식)
        - [FastCGI 방식](#fastcgi-방식)
        - [PHP-FPM 설정 예시](#php-fpm-설정-예시)
        - [Zend Engine의 메모리 관리 비교](#zend-engine의-메모리-관리-비교)
            - [CGI 방식](#cgi-방식-1)
            - [FastCGI 방식](#fastcgi-방식-1)
        - [요약](#요약)
    - [다른 인터프리터 언어 비교: Python과 CPython](#다른-인터프리터-언어-비교-python과-cpython)
    - [기타](#기타)

## PHP 실행 단계

1. 소스 코드 작성

    개발자가 클래스를 정의하고, 프로그램의 로직을 작성합니다.

    ```php
    class MyClass { ... }
    ```

2. 컴파일 단계 by Zend Engine

    PHP는 보통 인터프리터 방식으로 실행되지만, 캐시된 바이트코드 또는 Just-In-Time (JIT) 컴파일러를 사용할 수 있습니다.
    이 단계에서 코드는 바이트코드로 변환됩니다.

    가령 `::class` 연산자는 이 단계에서 문자열로 변환됩니다.
    `Does\Not\Exist::class;`는 컴파일 타임에 `"Does\Not\Exist"`로 변환됩니다.

3. 오토로딩 준비

    PHP의 자동 로더가 준비되는 단계입니다. 이는 `spl_autoload_register`을 통해 설정됩니다.
    PHP는 클래스를 필요로 할 때 클래스를 자동으로 로드할 수 있도록 오토로더를 설정합니다.

4. 런타임 실행 (Runtime Execution)

    PHP 스크립트가 실행되는 단계입니다.

    이 단계에서 실제로 클래스가 사용될 때 PHP는 오토로딩을 통해 클래스 정의를 메모리에 로드합니다.
    - 클래스 정의 로드: 클래스 파일이 디스크에서 읽혀져 메모리에 로드됩니다.
    - 생성자 호출: 클래스 인스턴스화 시 생성자가 호출됩니다.
    - 메서드 실행: 정적 메서드나 객체 메서드가 호출됩니다.
    - 속성 접근: 정적 속성이나 객체 속성에 접근합니다.

## Zend Engine

Zend Engine의 주요 구성 요소

- 파서 (Parser)

    PHP 소스 코드를 구문 분석하여 추상 구문 트리(Abstract Syntax Tree, AST)를 생성합니다.

    - 구문 분석 도구: Zend Engine의 파서는 `Bison` 같은 도구를 사용하여 PHP 소스 코드를 분석합니다.
    - 토큰화 (Tokenization): PHP 소스 코드를 개별 토큰으로 분해합니다.
    - 구문 트리 생성: 토큰을 기반으로 구문 규칙에 따라 AST를 생성합니다.

- 컴파일러 (Compiler)

    파서에서 생성된 AST를 바이트코드(Bytecode)로 변환합니다.

    - 중간 표현 (Intermediate Representation): AST를 바이트코드로 변환하는 중간 단계를 거칩니다.
    - 최적화 (Optimization): 최적화 기법을 적용하여 실행 성능을 향상시킵니다.
    - 바이트코드 생성: 최종적으로 Zend Engine의 가상 머신이 실행할 수 있는 바이트코드를 생성합니다.

- 가상 머신 (Virtual Machine, VM)

    컴파일러가 생성한 바이트코드를 실행합니다.

    - 바이트코드 해석기 (Bytecode Interpreter): 바이트코드를 해석하고 실행합니다.
    - 실행 스택 (Execution Stack): 함수 호출과 같은 실행 컨텍스트를 관리합니다.
    - 명령어 집합 (Instruction Set): 바이트코드를 해석하기 위한 명령어 집합을 포함합니다.

- 메모리 관리자 (Memory Manager)

    Zend Engine이 사용하는 메모리를 관리합니다.

    - 메모리 할당 (Memory Allocation): 필요한 메모리를 할당합니다.
    - 가비지 컬렉션 (Garbage Collection): 더 이상 사용되지 않는 메모리를 해제합니다.
    - 참조 카운팅 (Reference Counting): 객체와 변수의 참조를 추적하여 메모리 누수를 방지합니다.

- 실행기 (Executor)

    바이트코드를 실제로 실행하며, 실행 중 발생하는 각종 연산을 수행합니다.

    - 연산 수행 (Operation Execution): 바이트코드 명령을 순차적으로 실행합니다.
    - 함수 호출 관리 (Function Call Management): 함수 호출 및 반환을 관리합니다.
    - 예외 처리 (Exception Handling): 실행 중 발생하는 예외를 처리합니다.

- 확장 매니저 (Extension Manager)

    PHP 확장을 로드하고 관리합니다.

    - 확장 로딩 (Extension Loading): PHP 확장 모듈을 로드합니다.
    - 확장 초기화 (Extension Initialization): 확장의 초기화를 수행합니다.
    - 확장 함수와 클래스 등록 (Extension Function and Class Registration): 확장에서 제공하는 함수와 클래스를 등록합니다.

### Zend Engine의 작동 원리

1. 초기화 (Initialization)

    Zend Engine은 PHP가 시작될 때 초기화됩니다.
    이 과정에서 필요한 메모리와 실행 환경이 설정됩니다.

2. 파싱 (Parsing)

    PHP 소스 코드는 파서에 의해 구문 분석되어 AST가 생성됩니다.

3. 컴파일 (Compiling)

    컴파일러가 AST를 바이트코드로 변환합니다.
    이 과정에서 최적화가 수행됩니다.

4. 실행 (Execution)

    가상 머신이 바이트코드를 해석하고 실행합니다.
    실행기와 메모리 관리자가 이 과정에서 중요한 역할을 합니다.

5. 종료 (Shutdown)

    PHP 스크립트가 종료되면 Zend Engine은 메모리를 해제하고 필요한 정리 작업을 수행합니다.

### Zend Engine의 확장성과 모듈성

Zend Engine은 모듈화된 구조로 설계되어 있어, 다양한 확장을 통해 기능을 확장할 수 있습니다.
이는 PHP의 유연성과 강력한 기능을 제공하는 중요한 요소입니다.

- 확장 모듈 (Extension Modules)

    PHP는 다양한 확장을 통해 기능을 추가할 수 있습니다.

    예를 들어, 데이터베이스 연결, 이미지 처리, 암호화 등 다양한 기능이 확장 모듈을 통해 제공됩니다.

- API (Application Programming Interface)

    Zend Engine은 확장 모듈 개발을 위한 다양한 API를 제공합니다.
    이를 통해 개발자는 자신만의 확장을 쉽게 만들 수 있습니다.

## Zend Engine이 PHP 실행하는 방법

PHP는 인터프리터 언어로, 코드를 직접 기계어로 컴파일하지 않습니다.
대신, PHP 코드는 *바이트코드로 변환되어 가상 머신(예: Zend Engine)에 의해 실행*됩니다.
이 과정에서 바이트코드는 최종적으로 기계어로 변환되어 CPU에서 직접 실행됩니다.

작성된 PHP 코드는 아래 과정을 통해 실행됩니다.

1. PHP 인터프리터는 PHP 코드를 파싱하여 AST를 생성하고, 이를 바이트코드로 변환합니다.

    - `Zend/zend_language_scanner.l`: PHP 코드를 토큰화
    - `Zend/zend_language_parser.y`: 토큰화된 코드를 문법 규칙에 따라 파싱해서 AST로 변환

2. 생성된 AST는 `Zend/zend_compile.c` 파일의 함수들을 통해 바이트코드로 변환되며, `zend_op_array` 구조체에 저장됩니다.
3. 바이트코드는 `Zend/zend_vm_execute.h` 파일의 `zend_execute()` 함수에 의해 실행되며, Zend Engine에 의해 기계어로 변환되어 CPU에서 실행됩니다.

```php
<?php
$name = "John";
echo "Hello, " . $name . "!";
```

PHP 인터프리터는 위의 PHP 코드를 파싱하여 추상 구문 트리(AST)를 생성합니다.

1. 토큰화

    PHP 코드는 먼저 `Zend/zend_language_scanner.l` 파일에 정의된 *Flex 규칙*에 따라 토큰화됩니다.
    예제 코드의 토큰화 결과는 다음과 같습니다:

    ```php
    $name = "John";
    ```

    - `T_VARIABLE`
    - `=`
    - `T_CONSTANT_ENCAPSED_STRING`
    - `;`

    ```php
    echo "Hello, " . $name . "!";
    ```

    - `T_ECHO`
    - `T_CONSTANT_ENCAPSED_STRING`
    - `.`
    - `T_VARIABLE`
    - `.`
    - `T_CONSTANT_ENCAPSED_STRING`
    - `;`

2. 파싱

    토큰화된 코드는 `Zend/zend_language_parser.y` 파일의 문법 규칙에 따라 파싱됩니다.
    파싱 과정에서 AST가 생성됩니다.

    예제 코드의 AST는 다음과 같은 구조를 가집니다:

    ```bash
    ZEND_AST_STMT_LIST
    ├── ZEND_AST_ASSIGN
    │   ├── ZEND_AST_VAR
    │   │   └── "name"
    │   └── ZEND_AST_ZVAL
    │       └── "John"
    └── ZEND_AST_ECHO
        └── ZEND_AST_BINARY_OP
            ├── ZEND_AST_BINARY_OP
            │   ├── ZEND_AST_ZVAL
            │   │   └── "Hello, "
            │   └── ZEND_AST_VAR
            │       └── "name"
            └── ZEND_AST_ZVAL
                └── "!"
    ```

3. AST에서 바이트코드 생성 위해 컴파일러 호출

    생성된 AST는 `Zend/zend_compile.c` 파일의 함수들을 통해 바이트코드로 변환됩니다.

    컴파일러는 `zend_compile_file` 함수를 호출하여 파일을 컴파일합니다.
    이 함수는 AST를 바이트코드로 변환하는 역할을 합니다.

    ```c
    zend_op_array *zend_compile_file(zend_file_handle *file_handle, int type) {
        // 파일을 파싱하여 AST 생성
        zend_ast *ast = zend_parse(file_handle, type);
        
        // AST를 바이트코드로 변환
        zend_op_array *op_array = zend_compile_ast(ast);
        
        return op_array;
    }
    ```

4. 바이트코드 생성

    컴파일러는 `zend_compile_assign()` 함수와 `zend_compile_echo()` 함수를 사용하여 AST 노드를 바이트코드로 변환합니다.

    `zend_compile_assign()` 함수는 `$name = "John";` 할당문 노드를 바이트코드로 변환합니다:

    - `zend_emit_op()` 함수를 사용하여 `ZEND_ASSIGN` opcode를 생성합니다.
    - 변수 `$name`과 값 `"John"`에 대한 정보를 바이트코드에 포함합니다.

    `zend_compile_echo()` 함수는 `echo "Hello, " . $name . "!";` 문 노드를 바이트코드로 변환합니다:

    - `zend_compile_expr()` 함수를 사용하여 문자열 연결 노드를 컴파일합니다.
    - `zend_emit_op()` 함수를 사용하여 `ZEND_CONCAT` opcode를 생성하여 문자열을 연결합니다.
    - `zend_emit_op()` 함수를 사용하여 `ZEND_ECHO` opcode를 생성하여 결과 문자열을 출력합니다.

    생성된 바이트코드는 다음과 같은 구조를 가집니다:

    ```bash
    line  #* E I O op                           fetch          ext  return  operands
    ---------------------------------------------------------------------------------
       2  0E >     ASSIGN                                                   !0, 'John'
       3  1        CONCAT                                                   !1, 'Hello, ', !0
       3  2        CONCAT                                                   !2, !1, '!'
       3  3        ECHO                                                     !2
       3  4      > RETURN                                            1
    ```

    헤더의 의미:

    - line: 해당 바이트코드 명령어가 원본 PHP 코드의 몇 번째 줄에서 생성되었는지를 나타냅니다. 이는 디버깅과 오류 추적에 유용합니다.
        > ex: 2, 3 - 원본 PHP 코드의 줄 번호.
    - #: 바이트코드 명령어의 인덱스를 나타냅니다. 바이트코드 명령어는 순차적으로 실행되며, 이 인덱스는 명령어의 순서를 나타냅니다.
        > 0, 1, 2, 3, 4 - 바이트코드 명령어의 인덱스.
    - \*: 바이트코드 명령어의 플래그를 나타냅니다. 예를 들어, `E`는 이 명령어가 예외를 발생시킬 수 있음을 나타냅니다.
        > E - 예외를 발생시킬 수 있는 명령어.
    - E: 명령어가 예외를 발생시킬 수 있는지 여부를 나타냅니다. `E`가 표시되면 해당 명령어는 예외를 발생시킬 수 있습니다.
        > \> - 예외를 발생시킬 수 있는 명령어.
    - I: 명령어가 인터럽트를 발생시킬 수 있는지 여부를 나타냅니다. `I`가 표시되면 해당 명령어는 인터럽트를 발생시킬 수 있습니다.
        > (비어 있음) - 인터럽트를 발생시키지 않음.
    - O: 명령어가 옵코드 핸들러를 호출할 수 있는지 여부를 나타냅니다. `O`가 표시되면 해당 명령어는 옵코드 핸들러를 호출할 수 있습니다.
        > (비어 있음) - 옵코드 핸들러를 호출하지 않음.
    - op: 바이트코드 명령어의 이름을 나타냅니다. 예를 들어, `ASSIGN`, `ECHO`, `CONCAT` 등이 있습니다.
        > `ASSIGN`, `CONCAT`, `ECHO`, `RETURN`은 바이트코드 명령어 이름입니다.

        - `ASSIGN` 명령어: 변수 `0`에 `'John'` 값을 할당합니다. fetch 열이 비어 있지만, 이는 operands 열에서 직접 값을 가져옵니다.
        - 첫 번째 `CONCAT` 명령어: `'Hello, '`와 `0` (`$name`)을 연결하고 결과를 임시 변수 `1`에 저장합니다.
        - 두 번째 `CONCAT` 명령어: `1`과 `'!'`를 연결하고 결과를 임시 변수 `2`에 저장합니다.
        - `ECHO` 명령어: 임시 변수 `2`의 값을 출력합니다.

    - fetch: 명령어가 데이터를 가져오는 방식을 나타냅니다. 예를 들어, `FETCH_CONSTANT`, `FETCH_LOCAL` 등이 있습니다.
        > (비어 있음) - 데이터 가져오기 방식.
    - ext: 명령어의 확장 정보를 나타냅니다. 이는 특정 명령어에 대한 추가 정보를 제공할 수 있습니다.
        > (비어 있음) - 확장 정보.
    - return: 명령어의 반환 값을 나타냅니다. 예를 들어, `1`은 명령어가 성공적으로 실행되었음을 나타낼 수 있습니다.
        > 1 - 반환 값.
    - operands: 명령어의 피연산자를 나타냅니다. 예를 들어, 변수, 상수, 임시 값 등이 포함될 수 있습니다.
        > `!0, 'John'`, `!1, 'Hello, ', !0`, `!2, !1, '!'`, `!2` - 명령어의 피연산자.

        `0`, `1`, `2`와 같은 표현은 임시 변수를 나타냅니다.
        PHP 바이트코드에서 임시 변수는 실행 중에 생성되는 값을 저장하는 데 사용됩니다.

        예를 들어, `CONCAT` 명령어는 두 문자열을 연결하고 결과를 임시 변수에 저장합니다.
        이 임시 변수는 나중에 다른 명령어에서 사용될 수 있습니다.

5. Zend Engine의 바이트코드 실행

    Zend Engine은 바이트코드를 실행합니다. 이 과정에서 바이트코드는 기계어로 변환되어 CPU에서 실행됩니다.

    Zend Engine은 바이트코드를 실행하기 위해 `zend_execute` 함수를 사용합니다.
    이 함수는 [`Zend/zend_vm_execute.h` 파일](https://github.com/php/php-src/blob/master/Zend/zend_vm_execute.h)에 정의되어 있습니다.

    ```c
    ZEND_API void zend_execute(zend_op_array *op_array, zval *return_value)
    {
        zend_execute_data *execute_data;

        if (EG(exception) != NULL) {
            return;
        }

        execute_data = zend_vm_stack_push_call_frame(ZEND_CALL_TOP_CODE, (zend_function*)op_array, 0, zend_get_called_scope(EG(current_execute_data)));
        zend_init_execute_data(execute_data, op_array, return_value);
        ZEND_OBSERVER_FCALL_BEGIN(execute_data);
        zend_execute_ex(execute_data);
        zend_vm_stack_free_call_frame(execute_data);
    }
    ```

    1. 현재 실행 중인 코드의 `this` 객체나 호출된 스코프를 가져옵니다.
    2. `zend_vm_stack_push_call_frame` 함수를 사용하여 새로운 실행 프레임을 생성합니다.
    3. 심볼 테이블을 설정합니다.
    4. `zend_init_execute_data` 함수를 사용하여 실행 데이터를 초기화합니다.
    5. `ZEND_OBSERVER_FCALL_BEGIN` 매크로를 사용하여 함수 호출 시작을 알립니다.
    6. `zend_execute_ex` 함수를 호출하여 실제 바이트코드를 실행합니다.
    7. `zend_vm_stack_free_call_frame` 함수를 사용하여 실행 프레임을 해제합니다.

6. 바이트코드 실행

    실제 바이트코드 실행은 `zend_execute_ex` 함수에서 이루어집니다.
    이 함수는 `Zend/zend_vm_execute.h` 파일에 정의되어 있습니다.

    ```c
    ZEND_API void zend_execute_ex(zend_execute_data *execute_data)
    {
        const zend_op *opline;

        LOAD_OPLINE();
        while (1) {
            ZEND_VM_LOOP_INTERRUPT_CHECK();
            ZEND_VM_LOOP_HOOK(execute_data, opline);
            ZEND_VM_DISPATCH(opline, EX(opline)->handler);
        }
    }
    ```

    이 함수는 `zend_execute` 함수에서 호출되며, 실행 데이터를 인자로 받아 바이트코드를 실행합니다.
    각 opcode에 대한 처리는 `ZEND_VM_DISPATCH` 매크로를 통해 이루어집니다.

    예를 들어, `ZEND_ECHO` opcode는 `ZEND_ECHO_SPEC` 매크로에 의해 구현되며, 문자열을 출력하는 역할을 합니다.

    ```c
    #define ZEND_ECHO_SPEC() \
        do { \
            zval *z = EX_CONSTANT(opline->op1); \
            zend_print_zval(z, 0); \
        } while (0)
    ```

    실 행 결과로 `"Hello, John!"`이 출력됩니다.

## 바이트코드와 기계어의 차이

- 바이트코드

    바이트코드는 고수준 언어의 명령어를 저수준의 명령어로 변환한 것입니다.
    변환된 바이트코드는 가상 머신(예: Zend Engine)에서 실행됩니다.

    예: PHP 바이트코드

    ```bytecode
    0  ECHO "Hello, World!"
    1  RETURN 1
    ```

    - `ECHO "Hello, World!"`: 문자열을 출력하는 명령어입니다.
    - `RETURN 1`: 스크립트를 종료하는 명령어입니다.

- 기계어:

    기계어는 CPU가 직접 실행할 수 있는 명령어입니다.
    이런 기계어는 특정 CPU 아키텍처에 종속적입니다.

    예: x86 아키텍처 기계어 명령어

    ```assembly
    B8 04 00 00 00    ; mov eax, 4
    BB 01 00 00 00    ; mov ebx, 1
    B9 00 00 00 00    ; mov ecx, msg
    BA 0E 00 00 00    ; mov edx, 14
    CD 80             ; int 0x80
    ```

    - `mov eax, 4`: 시스템 호출 번호 4 (sys_write)를 `eax` 레지스터에 저장합니다.
    - `mov ebx, 1`: 파일 디스크립터 1 (stdout)을 `ebx` 레지스터에 저장합니다.
    - `mov ecx, msg`: 출력할 문자열의 주소를 `ecx` 레지스터에 저장합니다.
    - `mov edx, 14`: 출력할 문자열의 길이를 `edx` 레지스터에 저장합니다.
    - `int 0x80`: 시스템 호출을 실행합니다.

PHP와 Python의 실행 과정과 Virtual Machine(VM)의 역할에 대해 설명드리겠습니다. 두 언어의 VM은 각각의 생태계에서 중요하지만, 운영 방식과 최적화 방법은 다릅니다.

## Zend Engine Virtual Machine

Zend Engine은 PHP 스크립트를 실행하기 위한 핵심 컴포넌트로, *PHP 코드를 바이트코드로 컴파일하고 이를 실행*합니다.
하지만 이는 JVM(Java Virtual Machine)과는 몇 가지 중요한 차이점이 있습니다:

1. JVM과의 차이점:

    JVM은 장기 실행 프로세스로, 애플리케이션이 실행되는 동안 메모리를 지속적으로 관리하고 최적화합니다.
    반면 Zend Engine은 일반적으로 PHP 스크립트가 실행될 때마다 새로 인스턴스화되고 스크립트 실행이 끝나면 종료됩니다.

    JVM은 Eden, Survivor, Tenured, PermGen 등 여러 메모리 영역을 관리하고 `가비지 컬렉션`(GC)을 통해 메모리를 최적화합니다.
    Zend Engine은 단기 실행의 특성상 이와 같은 복잡한 메모리 관리 기법을 사용하지 않습니다.
    PHP 스크립트의 실행 중에 필요한 메모리를 할당하고, 스크립트 종료 시 이를 해제하는 방식입니다.

    Zend Engine의 VM은 "가상 머신"이라는 이름을 사용하지만, 이는 JVM처럼 독립적인 프로세스가 아닌 실행 환경의 의미입니다.
    Zend VM은 바이트코드를 실행할 때 필요한 가상 CPU와 스택을 관리하는 역할을 합니다.

2. Zend Engine의 동작 방식:

    PHP 스크립트가 실행되면 Zend Engine이 해당 스크립트를 바이트코드로 컴파일합니다.

    바이트코드는 Zend VM에 의해 실행됩니다.

    스크립트 실행이 완료되면 Zend Engine은 종료됩니다.

PHP의 Zend Engine이 메모리를 관리하는 방식은 주로 `참조 카운팅`(reference counting)과 주기적인 `가비지 컬렉션`(garbage collection)을 통해 이루어집니다. 이 두 가지 방식은 PHP 스크립트가 실행되는 동안 메모리를 효율적으로 관리하기 위해 사용됩니다. 아래에 Zend Engine의 메모리 관리 방식을 단계별로 자세히 설명하겠습니다.

### Zend Engine의 메모리 관리

PHP 스크립트가 실행되면 Zend Engine은 필요한 변수를 생성하고 이들을 메모리에 할당합니다.

PHP 스크립트에서 변수가 선언되거나 객체가 생성될 때마다 메모리가 할당됩니다.
Zend Engine은 `zend_alloc.c` 파일에 정의된 메모리 할당 함수를 사용하여 메모리를 관리합니다.

예를 들어, `emalloc` 함수는 메모리를 할당하고, `efree` 함수는 메모리를 해제합니다.

PHP는 기본적인 메모리 관리 기법으로 `참조 카운팅`(Reference Counting)을 사용합니다.
참조 카운팅은 객체나 변수가 참조되는 횟수를 카운트하여 관리하는 방식입니다. 이는 입니다.

- 참조 카운트 증가

    변수나 객체가 할당될 때 참조 카운트는 1로 초기화됩니다.
    다른 변수가 이 변수를 참조하면 참조 카운트가 증가합니다.

- 참조 카운트 감소

    변수를 더 이상 참조하지 않게 되면 참조 카운트가 감소합니다.

    예를 들어, 변수가 함수 스코프를 벗어나거나 `unset` 함수를 통해 해제되면 참조 카운트가 감소합니다.

- 메모리 해제

    참조 카운트가 0이 되면 메모리가 해제됩니다.
    이는 Zend Engine의 `가비지 컬렉션` 루틴에 의해 처리됩니다.

하지만 `참조 카운팅`만으로는 순환 참조(circular reference) 문제를 해결할 수 없습니다.
주기적인 `가비지 컬렉션`은 이러한 문제를 해결하기 위해 사용됩니다.

- 순환 참조 탐지:

    Zend Engine의 가비지 컬렉터는 순환 참조를 탐지하고 이를 해제합니다.
    이는 `zend_gc.c` 파일에 정의된 `가비지 컬렉션` 루틴에 의해 이루어집니다.

- 가비지 컬렉션 주기:

     주기적으로 `가비지 컬렉션`이 실행되어 순환 참조로 인해 해제되지 않은 메모리를 회수합니다.

```php
// 문자열 `"Hello, World!"`가 메모리에 할당됩니다.
// 참조 카운트는 1입니다.
$a = "Hello, World!";

// `$b`가 `$a`를 참조하므로, 문자열의 참조 카운트는 2가 됩니다.
$b = $a;

// `$a`의 참조가 해제되므로, 문자열의 참조 카운트는 1로 감소합니다.
unset($a);

// `$x`와 `$y`는 서로를 참조하는 순환 참조를 만듭니다.
// 이 경우 참조 카운트로는 메모리를 해제할 수 없습니다.
// 주기적인 `가비지 컬렉션`이 실행되어 `$x`와 `$y`의 순환 참조를 탐지하고 이를 해제합니다.
$x = new stdClass();
$y = new stdClass();
$x->y = $y;
$y->x = $x;
```

### Zend Engine의 `가비지 컬렉션` 코드 예시

- 참조 카운팅: `zend_types.h` 파일의 `zval` 구조체에 `refcount` 필드가 있어 참조 카운트를 관리합니다.

    ```c
    typedef struct _zval_struct {
        zend_value value;       /* value */
        union {
            struct {
                zend_uchar type;    /* active type */
                zend_uchar type_flags;
                zend_uchar const_flags;
                zend_uchar reserved; /* call info for EX(This) */
            } v;
            uint32_t type_info;
        } u1;
        union {
            uint32_t next;       /* hash collision chain */
            uint32_t cache_slot; /* cache slot (for RECV_INIT) */
            uint32_t lineno;     /* line number (for ast nodes) */
            uint32_t num_args;   /* arguments number for EX(This) */
            uint32_t fe_pos;     /* foreach position */
            uint32_t fe_iter_idx;/* foreach iterator index */
        } u2;
        uint32_t refcount;        /* reference count */
    } zval;
    ```

- 가비지 컬렉션: `zend_gc.c` 파일에서 주기적인 `가비지 컬렉션`이 이루어집니다.

    ```c
    void gc_collect_cycles(void) {
        // 순환 참조 탐지 및 해제
    }
    ```

## PHP 스크립트 실행과 프로세스 관리

- CGI 방식

    - 각 요청마다 새로운 PHP 프로세스가 생성되고 종료됩니다.
    - 프로세스 종료 시 운영체제가 모든 메모리를 해제하므로 메모리 누수가 발생할 여지가 적습니다.
    - 요청마다 프로세스를 생성하고 종료하는 오버헤드로 인해 성능 저하가 발생할 수 있습니다.

- FastCGI 방식

    - PHP 프로세스가 여러 요청을 처리하기 위해 지속적으로 실행됩니다.
    - PHP-FPM이 프로세스를 관리하고 주기적으로 재시작하여 메모리 누수를 방지합니다.
    - 프로세스를 재사용하므로 성능이 향상됩니다.

### CGI 방식

CGI(Common Gateway Interface) 방식에서는 각 HTTP 요청마다 새로운 PHP 인터프리터 프로세스가 생성되고, 요청 처리가 끝나면 해당 프로세스가 종료됩니다.

- 참조 카운팅과 가비지 컬렉션: 스크립트 실행 중에 참조 카운팅과 가비지 컬렉션을 통해 불필요한 메모리를 회수합니다.
- 프로세스 종료 시 메모리 해제: 스크립트 실행이 끝난 후 프로세스가 종료되면서 운영체제는 모든 메모리를 자동으로 해제합니다.

```plaintext
            +---------------------+
            |      Web Server     |
            +---------------------+
                      |
        +-------------+-------------+
        |                           |
+---------------+           +---------------+
|  PHP Process  |           |  PHP Process  |
|      (1)      |           |      (2)      |
+---------------+           +---------------+
        |                           |
        v                           v
+---------------------+   +---------------------+
| Execute Script      |   | Execute Script      |
+---------------------+   +---------------------+
        |                           |
        v                           v
+---------------------+   +---------------------+
| Terminate Process   |   | Terminate Process   |
+---------------------+   +---------------------+
        |                           |
        v                           v
+---------------------+   +---------------------+
| Release Memory      |   | Release Memory      |
+---------------------+   +---------------------+
```

1. 웹 서버: 클라이언트 요청이 들어오면 웹 서버가 새로운 PHP 프로세스를 생성합니다.
2. PHP 프로세스: 각 요청마다 새로운 PHP 프로세스가 생성되어 스크립트를 실행합니다.
3. 스크립트 실행: PHP 프로세스가 스크립트를 실행하고, 결과를 웹 서버에 반환합니다.
4. 프로세스 종료: 스크립트 실행이 완료되면 PHP 프로세스가 종료됩니다.
5. 메모리 해제: 프로세스 종료 시 운영체제가 해당 프로세스와 연관된 모든 메모리를 해제합니다.

### FastCGI 방식

FastCGI 방식에서는 PHP 인터프리터 프로세스가 여러 요청을 처리하기 위해 지속적으로 실행됩니다.

- 참조 카운팅과 가비지 컬렉션: 스크립트 실행 중에 참조 카운팅과 가비지 컬렉션을 통해 불필요한 메모리를 회수합니다.
- 프로세스 종료 시 메모리 해제: FastCGI 방식에서는 프로세스가 지속적으로 실행되므로, PHP-FPM이 주기적으로 프로세스를 재시작하여 메모리를 해제합니다. 이는 메모리 누수를 방지하기 위해 필요합니다.

PHP-FPM(FastCGI Process Manager)이 이를 관리합니다:

- 지속적 실행: PHP 프로세스는 여러 요청을 처리하기 위해 지속적으로 실행됩니다.
- 메모리 모니터링: 각 PHP 프로세스는 메모리 사용량을 모니터링하여 메모리 누수를 감지합니다.
- 요청 수 모니터링: PHP-FPM은 각 프로세스가 처리한 요청 수를 모니터링합니다.
- 프로세스 재시작: 설정된 요청 수나 메모리 사용 한계에 도달하면 PHP-FPM은 해당 프로세스를 재시작하여 메모리를 해제하고 새로운 프로세스를 시작합니다.

```plaintext
                +---------------------+
                |      Web Server     |
                +---------------------+
                          |
        +-----------------+-----------------------+
        |                                         |
+---------------------+                +---------------------+
|     PHP-FPM         |                |    PHP-FPM          |
+---------------------+                +---------------------+
        |                                         |
+---------------------+                +---------------------+
|  PHP Process Pool   |                |  PHP Process Pool   |
+---------------------+                +---------------------+
        |                                         |
        |                             +-----------+------------+
        |                             |                        |
+---------------+            +---------------+        +---------------+
| PHP Process 1 |            | PHP Process 2 |        | PHP Process 3 |
+---------------+            +---------------+        +---------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Execute Script      |     | Execute Script      | | Execute Script      |
+---------------------+     +---------------------+ +---------------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Return Response     |     | Return Response     | | Return Response     |
+---------------------+     +---------------------+ +---------------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Monitor Memory Usage|     | Monitor Memory Usage| | Monitor Memory Usage|
+---------------------+     +---------------------+ +---------------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Check Request Count |     | Check Request Count | | Check Request Count |
+---------------------+     +---------------------+ +---------------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Serve Next Request  |     | Serve Next Request  | | Serve Next Request  |
+---------------------+     +---------------------+ +---------------------+
        |                             |                        |
        v                             v                        v
+---------------------+     +---------------------+ +---------------------+
| Restart if Necessary|     | Restart if Necessary| | Restart if Necessary|
+---------------------+     +---------------------+ +---------------------+
```

1. 웹 서버: 클라이언트로부터 HTTP 요청을 받습니다.
2. PHP-FPM: 웹 서버는 요청을 PHP-FPM에 전달합니다.
3. PHP 프로세스 풀: PHP-FPM은 프로세스 풀을 관리합니다. 프로세스 풀에는 여러 개의 PHP 프로세스가 존재합니다.
4. 프로세스 선택: PHP-FPM은 요청을 처리하기 위해 프로세스 풀에서 하나의 프로세스를 선택합니다.
5. 스크립트 실행: 선택된 PHP 프로세스가 스크립트를 실행합니다.
6. 응답 반환: 스크립트 실행이 완료되면, PHP 프로세스는 결과를 PHP-FPM에 반환하고, PHP-FPM은 이를 웹 서버를 통해 클라이언트에게 전달합니다.
7. 메모리 사용량 모니터링: 각 PHP 프로세스는 스크립트 실행 후 메모리 사용량을 모니터링합니다.
8. 요청 수 확인: PHP-FPM은 각 PHP 프로세스가 처리한 요청 수를 모니터링합니다.
9. 다음 요청 처리 준비: 메모리 사용량과 요청 수가 한계에 도달하지 않은 경우, PHP 프로세스는 다음 요청을 처리할 준비를 합니다.
10. 필요 시 프로세스 재시작: 메모리 사용량이 한계에 도달하거나, 처리한 요청 수가 설정된 값을 초과하면, PHP-FPM은 해당 PHP 프로세스를 재시작하여 메모리 누수를 방지합니다.

### PHP-FPM 설정 예시

```ini
; PHP-FPM 설정 파일 (php-fpm.conf)
[www]
pm = dynamic
pm.max_children = 50
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 35
pm.max_requests = 500 # 각 프로세스가 500개의 요청을 처리한 후 재시작
```

### Zend Engine의 메모리 관리 비교

#### CGI 방식

- 스크립트 실행 중 참조 카운팅과 가비지 컬렉션을 통해 메모리를 관리합니다.
- 각 요청마다 프로세스가 종료되므로, 운영체제가 자동으로 메모리를 해제합니다.

#### FastCGI 방식

- 스크립트 실행 중 참조 카운팅과 가비지 컬렉션을 통해 메모리를 관리합니다.
- PHP-FPM이 주기적으로 프로세스를 재시작하여 메모리를 해제하고 메모리 누수를 방지합니다.

### 요약

PHP의 CGI와 FastCGI 방식은 각각의 요청을 처리하는 방식에서 차이가 있습니다. CGI 방식은 각 요청마다 새로운 프로세스를 생성하고 종료하며, FastCGI 방식은 지속적인 프로세스를 사용하여 성능을 향상시킵니다. 두 방식 모두 Zend Engine의 참조 카운팅과 가비지 컬렉션을 사용하여 메모리를 관리하지만, FastCGI 방식에서는 PHP-FPM이 주기적으로 프로세스를 재시작하여 메모리 누수를 방지합니다.

## 다른 인터프리터 언어 비교: Python과 CPython

Python의 기본 구현인 CPython도 VM을 사용합니다. CPython의 VM은 Python 바이트코드를 실행하는 역할을 합니다.

1. Python 코드 실행 과정:
    - 파싱 및 AST 생성:
        - Python 소스 코드는 파서(Parser)에 의해 토큰으로 분해되고, 이 토큰들은 AST(Abstract Syntax Tree)로 변환됩니다.
        - 이 과정은 `tokenizer.c`와 `parser.c` 파일에서 이루어집니다.
    - AST를 바이트코드로 컴파일:
        - AST는 컴파일러에 의해 바이트코드로 변환됩니다. 이는 `compile.c` 파일에서 처리됩니다.
    - 바이트코드 실행:
        - 바이트코드는 Python VM(PVM, Python Virtual Machine)에서 실행됩니다. PVM은 바이트코드를 해석하고 실행하는 역할을 합니다.
        - PVM은 스택 기반의 가상 머신으로, 바이트코드 명령어를 실행하기 위해 스택을 사용합니다.

2. CPython VM의 특징:
    - 상주 프로세스: Python 인터프리터는 상주 프로세스가 아니며, 스크립트가 실행될 때마다 새로 시작되고 종료됩니다.
    - 메모리 관리: CPython은 가비지 컬렉터(GC)를 통해 메모리를 관리합니다. Python의 GC는 `참조 카운팅`과 사이클 검출 알고리즘을 사용하여 메모리를 회수합니다.
    - 실행 방식: Python의 바이트코드는 `.pyc` 파일에 저장될 수 있으며, 이는 이후에 재사용될 수 있습니다. 바이트코드는 Python 인터프리터가 실행할 때마다 해석되어 실행됩니다.

PHP의 Zend Engine은 주로 웹 서버 환경에서 동작하며, 각 요청마다 새로운 PHP 스크립트를 실행하고 종료하는 방식으로 메모리를 관리합니다.
이 과정에서 Zend Engine이 사용하는 메모리 관리 기법은 `참조 카운팅`과 `가비지 컬렉션`입니다.

- PHP와 Zend Engine:
    - PHP 코드는 Zend Engine에 의해 바이트코드로 컴파일되고, 해당 바이트코드는 Zend VM에서 실행됩니다.
    - Zend Engine은 스크립트 실행 시마다 인스턴스화되고, 스크립트 종료 시 종료됩니다.
    - 복잡한 메모리 관리 기법 없이 단기 실행을 위해 최적화되어 있습니다.

- Python과 CPython:
    - Python 코드는 CPython에 의해 바이트코드로 컴파일되고, 해당 바이트코드는 PVM에서 실행됩니다.
    - CPython 인터프리터는 스크립트 실행 시마다 새로 시작되고 종료됩니다.
    - `참조 카운팅`과 사이클 검출을 이용한 `가비지 컬렉션`을 통해 메모리를 관리합니다.

각 언어의 VM은 그 언어의 특성과 사용 패턴에 맞게 설계되어 있으며, JVM처럼 상주하는 프로세스와는 다른 방식으로 동작합니다.

## 기타

- 파서 정의: [zend_language_parser.y](https://github.com/php/php-src/blob/master/Zend/zend_language_parser.y)
- AST 구조체 정의: [zend_ast.h](https://github.com/php/php-src/blob/master/Zend/zend_ast.h)
- 컴파일러 코드: [zend_compile.c](https://github.com/php/php-src/blob/master/Zend/zend_compile.c)
- 바이트코드 구조체 정의: [zend_compile.h](https://github.com/php/php-src/blob/master/Zend/zend_compile.h)
- 바이트코드 실행 코드: [zend_execute.c](https://github.com/php/php-src/blob/master/Zend/zend_execute.c)
