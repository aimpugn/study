# print zval

- [print zval](#print-zval)
    - [PHP의 zval을 출력하는 함수](#php의-zval을-출력하는-함수)

## PHP의 zval을 출력하는 함수

```c
#include <stdio.h>
void print_zval(zval *val, const char *spec) {
    // zval의 타입에 따라 다른 출력
    switch (Z_TYPE_P(val)) {
        case IS_NULL:
            printf("Spec: %s, \t\t\tType: NULL\n", spec);
            break;
        case IS_LONG:
            printf("Spec: %s, \t\t\tType: Long, Value: %ld\n", spec, Z_LVAL_P(val));
            break;
        case IS_DOUBLE:
            printf("Spec: %s, \t\t\tType: Double, Value: %f\\n", spec, Z_DVAL_P(val)); // %f를 사용하여 double 값을 출력
            break;
        case IS_BOOL:
            printf("Spec: %s, \t\t\tType: Bool, Value: %s\\n", spec, Z_BVAL_P(val) ? "true" : "false"); // 불리언 값을 문자열로 출력
            break;
        case IS_ARRAY:
            printf("Spec: %s, \t\t\tType: Array\n", spec);
            // 배열의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
            break;
        case IS_OBJECT:
            printf("Spec: %s, \t\t\tType: Object\n", spec);
            // 객체의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
            break;
        case IS_STRING:
            printf("Spec: %s, \t\t\tType: String, Value: %s\n", spec, Z_STRVAL_P(val));
            break;
        case IS_RESOURCE:
            printf("Spec: %s, \t\t\tType: Resource, Resource ID: %ld\n", spec, Z_RESVAL_P(val));
            break;
        case IS_CONSTANT:
            // 상수 타입의 경우, 상수의 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
            printf("Spec: %s, \t\t\tType: Constant\n", spec);
            break;
        case IS_CONSTANT_AST:
            // AST 노드의 경우, 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
            printf("Spec: %s, \t\t\tType: Constant AST\n", spec);
            break;
        case IS_CALLABLE:
            // 콜백의 경우, 내용을 출력하는 것은 복잡하므로, 여기서는 타입만 출력
            printf("Spec: %s, \t\t\tType: Callable\n", spec);
            break;
        default:
            printf("Spec: %s, \t\t\tUnknown type\n", spec);
            break;
    }
}
```
