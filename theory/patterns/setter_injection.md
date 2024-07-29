# Setter Injection

## 명칭과 설명

```cpp
class Container {
private:
    int A;
    int B;
    int C;

public:
    void setA(int value) { A = value; }
    void setB(int value) { B = value; }
    void setC(int value) { C = value; }

    int getA() const { return A; }
    int getAWithSomeProcessed() const { return A + 10; }
    // 다른 getter 메서드들
};

void someLogic(Container& instance) {
    instance.setA(1);
    instance.setB(2);
    instance.setC(3);
}

int main() {
    Container instance;
    someLogic(instance);
    std::cout << "A: " << instance.getA() << std::endl;
    std::cout << "A with some processing: " << instance.getAWithSomeProcessed() << std::endl;
    return 0;
}
```

이는 특정 명칭을 가진 패턴이라기보다는 일반적인 프로그래밍 관행에 더 가깝습니다.
이런 방식을 설명할 때 몇 가지 관련 개념들이 있습니다:

1. **Setter Injection**:

    여러 값을 객체의 세터 메서드를 통해 설정하는 방식입니다.
    이는 의존성 주입(Dependency Injection) 패턴의 일종으로, 객체의 속성을 외부에서 설정할 수 있게 합니다.

2. **Initializer Method**:

    객체의 여러 속성을 한 번에 초기화하는 메서드를 통해 설정하는 방식입니다.
    이 메서드는 객체의 상태를 설정하는 데 사용됩니다.

3. **Facade Pattern**:
   - 복잡한 서브시스템에 대한 단순화된 인터페이스를 제공하는 패턴입니다. `someLogic` 함수는 여러 설정 로직을 캡슐화하여 단일 메서드로 제공하는 역할을 할 수 있습니다.

## 구체적인 예시

이 개념들을 C++ 코드 예시로 구체화해 보겠습니다.

## 요약

- **Setter Injection**: `someLogic` 함수는 `Container` 객체의 세터 메서드를 사용하여 객체의 속성을 설정합니다.
- **Initializer Method**: `someLogic` 함수는 객체의 여러 속성을 한 번에 초기화하는 역할을 합니다.
- **Facade Pattern**: `someLogic` 함수는 복잡한 초기화 로직을 단순화된 인터페이스로 제공합니다.

위와 같은 방식을 특정 명칭으로 부르기는 어렵지만, 객체의 속성을 외부에서 설정하고 초기화하는 일반적인 패턴을 따르고 있습니다. 이는 여러 디자인 패턴의 요소를 혼합하여 사용한 예시라고 볼 수 있습니다.

## C언어에서 사용

C와 같은 언어에서 복잡한 로직의 다양한 값을 채워주는 방식을 사용하는 이유는 여러 가지가 있습니다.
이 방식은 주로 메모리 관리, 함수의 제한 사항, 코드의 재사용성 및 유지보수성을 고려한 결과입니다.
이러한 방식은 특히 복잡한 초기화 로직이나 다양한 타입의 인자를 처리해야 할 때 매우 유용합니다.

1. **메모리 관리와 성능**:

    C 언어에서는 함수가 여러 개의 값을 반환하기 어렵기 때문에, 구조체 포인터를 함수 인자로 전달하여 값을 설정하는 방식이 효율적입니다.
    이 방식은 다음과 같은 이유로 성능을 향상시킬 수 있습니다.
    - 불필요한 메모리 복사를 방지
    - 직접적인 메모리 접근

    예를 들어, 큰 데이터를 복사하는 대신 포인터를 통해 직접 값을 설정하면 성능 향상과 메모리 사용량 감소가 가능합니다.

2. **함수의 제한 사항**:

    C 언어에서는 함수가 한 번에 하나의 값만 반환할 수 있습니다.
    복잡한 데이터를 반환해야 하는 경우, 구조체 포인터를 사용하여 함수 외부에서 값을 설정하는 방식이 자주 사용됩니다.
    이 방법은 함수가 다수의 출력을 필요로 할 때 유용합니다.

3. **코드의 재사용성**:

    동일한 구조체에 대해 여러 함수를 작성하여 다양한 설정을 할 수 있습니다.
    이는 코드의 재사용성을 높이고, 중복 코드를 줄이는 데 도움이 됩니다.
    다양한 초기화 로직을 캡슐화하여 별도의 함수로 분리하면, 동일한 구조체를 여러 함수에서 일관되게 설정할 수 있습니다.

4. **유지보수성과 가독성**:

    코드의 가독성과 유지보수성을 높이기 위해, 복잡한 로직을 함수로 분리하여 구조체 포인터를 통해 값을 설정하는 방식이 유용합니다.
    이는 코드의 모듈화를 촉진하여 특정 기능을 변경할 때 전체 코드를 수정할 필요 없이 해당 함수만 수정하면 됩니다.

```c
#include <stdio.h>

typedef struct {
    int a;
    int b;
    int c;
} Container;

void initializeContainer(Container *container) {
    container->a = 1;
    container->b = 2;
    container->c = 3;
}

int main() {
    Container container;
    initializeContainer(&container);
    printf("a: %d, b: %d, c: %d\n", container.a, container.b, container.c);
    return 0;
}
```

### php-src에서 가변 인자 목록(`va_list`)

C 언어에서 가변 인자를 받아서 처리하는 방식은 특히 많은 개수의 인자를 함수로 전달해야 할 때 유용합니다.
이는 특히 라이브러리나 프레임워크에서 자주 사용됩니다.

- 가변 인자 리스트 (`va_list`)?

    > [C 표준 라이브러리 `stdarg.h`](https://en.cppreference.com/w/c/variadic)

    `va_list`는 C 표준 라이브러리에서 가변 인자 함수를 처리하기 위해 사용되는 타입입니다.
    `va_start`, `va_arg`, `va_end`와 같은 매크로를 사용하여 가변 인자 리스트를 초기화하고, 순차적으로 인자를 추출할 수 있습니다.

- `zend_parse_va_args`

    이 함수는 `va_list` 포인터를 사용하여 가변 인자를 받아들입니다.
    `type_spec` 문자열을 통해 인자의 타입을 지정하고, 이를 순회하며 각 인자를 처리합니다.

    ```c
    // https://github.com/php/php-src/blob/c97885b3ccd1446a8938422cadcf30ee1e3f1d19/Zend/zend_API.c#L1137
    static zend_result zend_parse_va_args(uint32_t num_args, const char *type_spec, va_list *va, int flags) {
        const char *spec_walk;
        char c;
        uint32_t i;
        uint32_t min_num_args = 0;
        uint32_t max_num_args = 0;
        uint32_t post_varargs = 0;
        zval *arg;
        bool have_varargs = 0;
        bool have_optional_args = 0;
        zval **varargs = NULL;
        uint32_t *n_varargs = NULL;

        for (spec_walk = type_spec; *spec_walk; spec_walk++) {
            c = *spec_walk;
            // 스펙을 분석하여 인자들을 처리하는 코드...
        }
        // 생략...
    }
    ```

- `zend_parse_arg`

    이 함수는 특정 인자를 `va_list`로부터 추출하여, 해당 인자의 타입을 확인하고 처리합니다.

    ```c
    // https://github.com/php/php-src/blob/c97885b3ccd1446a8938422cadcf30ee1e3f1d19/Zend/zend_API.c#L1084
    static zend_result zend_parse_arg(uint32_t arg_num, zval *arg, va_list *va, const char **spec, int flags) {
        const char *expected_type = NULL;
        char *error = NULL;
        // 특정 인자를 파싱하여 적절한 타입으로 변환하는 코드...
    }
    ```
