# va_list

- [va\_list](#va_list)
    - [`va_list`?](#va_list-1)
        - [사용 목적](#사용-목적)
    - [`va_start` 매크로](#va_start-매크로)
    - [`va_end` 매크로](#va_end-매크로)
    - [예제](#예제)
        - [PHP `zend_parse_parameters` 함수](#php-zend_parse_parameters-함수)

## `va_list`?

- 가변 인자 리스트(variable argument list)
- 들은 표준 C 라이브러리의 일부로, 함수가 고정된 수의 인자 대신 다양한 수의 인자를 받아들일 수 있도록 해준다.

### 사용 목적

- 가변 인자 리스트는 함수가 다양한 수와 유형의 인자를 받아들일 수 있도록 돕는다.
- 이는 특히 로깅, 문자열 형식 지정, 에러 처리와 같은 상황에서 유용하다. 예를 들어, `printf` 함수는 다양한 유형과 수의 인자를 받아들이며, 이것이 가능한 이유는 내부적으로 가변 인자 리스트를 사용하기 때문.

## `va_start` 매크로

- `va_start` 매크로는 가변 인자 목록 `va_list`을 초기화.
- 이 매크로는 `va_list` 변수와 마지막 고정 인자를 매개변수로 받는다.
- 이 매크로를 호출한 후, `va_list` 변수는 첫 번째 가변 인자를 가리키게 된다.

## `va_end` 매크로

- `va_end` 매크로는 가변 인자 목록을 정리(clean up)하는 데 사용된다.
- 이것은 `va_start`로 시작된 각 가변 인자 목록에 대해 호출되어야 한다.
- `va_end`는 `va_list` 변수를 무효화하여 더 이상 해당 목록을 사용할 수 없게 만든다.

## 예제

### PHP `zend_parse_parameters` 함수

```c
END_API int zend_parse_parameters(int num_args TSRMLS_DC, const char *type_spec, ...) /* {{{ */
{
    va_list va;
    int retval;

    RETURN_IF_ZERO_ARGS(num_args, type_spec, 0);

    va_start(va, type_spec); // `va`라는 `va_list` 변수를 초기화하고, 
                             // `type_spec` 뒤에 오는 가변 인자들을 준비
    retval = zend_parse_va_args(num_args, type_spec, &va, 0 TSRMLS_CC);
    va_end(va); // 가변 인자 목록의 사용이 끝났음을 나타내며, 관련된 자원을 정리

    return retval;
}
```
