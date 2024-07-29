# macro

- [macro](#macro)
    - [매크로란?](#매크로란)
    - [예제](#예제)
        - [php-src: `convert_to_string`](#php-src-convert_to_string)
        - [php-src `Z_TYPE*` 매크로](#php-src-z_type-매크로)

## 매크로란?

- 매크로는 코드를 더 간결하고 효율적으로 만들어주는 기능
- 매크로는 C 프로그래밍에서 흔히 사용되는 기능으로, 코드를 더 간결하고 읽기 쉽게 만들기 위해 사용

## 예제

### php-src: `convert_to_string`

```c
// 주어진 변수가 문자열이 아닐 경우에만 그 변수를 문자열로 변환하는 작업을 수행하는 매크로
#define convert_to_string(op) if ((op)->type != IS_STRING) { _convert_to_string((op) ZEND_FILE_LINE_CC); }
```

- `#define convert_to_string(op)`
    - `convert_to_string`이라는 이름의 매크로를 정의
    - `op`는 매크로의 파라미터로, 변환하고자 하는 변수 의미
- `if ((op)->type != IS_STRING) { ... }`
    - 이 조건문은 `op`의 타입이 문자열(`IS_STRING`)이 아닐 경우에만 내부의 코드가 실행되도록 한다
    - 여기서 `->` 연산자는 `op`가 포인터라는 것을 가리키며, `type` 필드를 참조한다
- `_convert_to_string((op) ZEND_FILE_LINE_CC);`
    - `_convert_to_string` 함수는 실제로 변수를 문자열로 변환하는 역할
    - `ZEND_FILE_LINE_CC`는 PHP 내부에서 사용하는 매크로로, 현재 파일의 이름과 라인 번호를 매크로로 전달하는 데 사용되며, 디버깅과 오류 추적에 유용합니다.

### php-src `Z_TYPE*` 매크로

```c
#define Z_TYPE(zval)        (zval).type
#define Z_TYPE_P(zval_p)    Z_TYPE(*zval_p)
#define Z_TYPE_PP(zval_pp)    Z_TYPE(**zval_pp)
```

- `Z_TYPE(zval)`
    - `zval` 구조체의 `type` 필드에 접근한다.
    - `(zval).type`는 `zval` 구조체 인스턴스에 직접 접근하여 `type` 필드의 값을 가져온다.
- `Z_TYPE_P(zval_p)`
    - 포인터를 통해 `zval` 구조체의 `type` 필드에 접근한다
    - `*zval_p`는 포인터 `zval_p`가 가리키는 `zval` 구조체를 역참조하여 접근한다.
- `Z_TYPE_PP(zval_pp)`
    - 포인터의 포인터를 통해 `zval` 구조체의 `type` 필드에 접근한다.
    - `zval_pp`는 `zval_pp`가 가리키는 포인터를 역참조하여, 그 결과로 나온 또 다른 포인터를 다시 역참조하여 `zval` 구조체에 접근한다

```c
Z_TYPE_P(return_value) = IS_LONG;
```

- `return_value`: `zval` 구조체를 가리키는 포인터
- `Z_TYPE_P(return_value)`: `return_value`가 가리키는 `zval` 구조체의 `type` 필드에 접근
- `Z_TYPE_P(return_value) = IS_LONG;`: `return_value`가 가리키는 `zval` 구조체의 `type` 필드 값을 `IS_LONG`으로 설정g한다. 이는 해당 `zval` 변수가 정수형(long 타입) 데이터를 담고 있음을 나타낸다.

이러한 매크로 사용은 PHP의 내부 구현에서 변수의 타입과 값을 효율적으로 관리하고, 코드의 가독성을 높이기 위한 목적으로 사용됩니다. `zval` 구조체는 PHP 내부에서 변수의 타입, 값, 참조 카운트 등을 관리하는 데 핵심적인 역할을 합니다.
