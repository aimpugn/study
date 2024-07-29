# pointer

## function pointer

```c
struct _zend_object_handlers {
    // 생략 
    // 특정 객체를 다른 타입으로 캐스팅하는 함수를 가리키는 포인터
    zend_object_cast_t                        cast_object;
    // 생략 
};

// 이 함수 포인터는 PHP 객체가 다른 타입으로 캐스팅될 때 호출된다.
typedef int (*zend_object_cast_t)(zval *readobj, zval *retval, int type TSRMLS_DC);
```

- `zend_object_cast_t`
    - 특정 객체를 다른 타입으로 캐스팅하는 함수를 가리키는 포인터
    - 함수 포인터는 변수가 함수를 가리킬 수 있게 해주며, 이를 통해 다양한 함수를 동적으로 할당하고 호출할 수 있다.
- `typedef`: 새로운 타입을 정의하는 C 언어의 키워드
- `int (*zend_object_cast_t)`: `zend_object_cast_t`는 함수 포인터 타입으로, 반환 값이 int인 함수를 가리킨다
- `zval *readobj, zval *retval, int type TSRMLS_DC`: 함수의 매개변수 목록. 이 함수는 `zval` 구조체 포인터, 정수형 타입 인자 등을 받는다.

실제 `cast_object` 함수의 구현은 해당 객체 타입의 정의에 따라 다르며, 객체를 캐스팅하는 데 필요한 로직을 구현한다.
