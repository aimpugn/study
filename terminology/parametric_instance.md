# Parametric Instance

- [Parametric Instance](#parametric-instance)
    - [기본 개념](#기본-개념)
    - [예시](#예시)
    - [장점](#장점)
    - [결론](#결론)
    - [참고 자료](#참고-자료)

## 기본 개념

"Parametric Instance"는 소프트웨어 공학, 특히 객체 지향 프로그래밍과 디자인 패턴에서 중요한 개념입니다.
이는 특정 매개변수(Parameter)나 설정(Configuration)에 따라 다양한 형태로 인스턴스화(객체 생성)될 수 있는 클래스나 객체를 의미합니다.
이를 통해 소프트웨어의 유연성과 재사용성을 높일 수 있습니다.

- 객체를 생성할 때 매개변수를 사용하여 객체의 속성이나 동작을 지정할 수 있는 인스턴스.
- 동일한 클래스 정의로부터 다양한 매개변수를 사용하여 서로 다른 동작이나 상태를 가진 여러 객체를 생성할 수 있습니다.

## 예시

1. **제네릭 프로그래밍**:
   - 제네릭 프로그래밍에서는 클래스나 함수를 정의할 때 매개변수를 사용하여 데이터 타입을 지정할 수 있습니다.
   - 예를 들어, C++의 템플릿 클래스나 Java의 제네릭 클래스는 파라메트릭 인스턴스의 대표적인 예입니다.

   - ```cpp
     template <typename T>
     class MyClass {
         T value;
     public:
         MyClass(T val) : value(val) {}
         T getValue() { return value; }
     };
     ```

2. **디자인 패턴**:
   - **팩토리 메서드 패턴**이나 **추상 팩토리 패턴**은 파라메트릭 인스턴스를 사용하는 디자인 패턴의 예입니다.
   - 클라이언트 코드가 구체적인 클래스의 인스턴스를 직접 생성하지 않고, 매개변수를 통해 생성할 객체의 타입을 지정할 수 있습니다.

   - ```java
     public abstract class ShapeFactory {
         abstract Shape createShape(String type);
     }
     
     public class ConcreteShapeFactory extends ShapeFactory {
         Shape createShape(String type) {
             if (type.equals("circle")) return new Circle();
             else if (type.equals("square")) return new Square();
             else return null;
         }
     }
     ```

## 장점

1. **유연성**:
   - 파라메트릭 인스턴스를 통해 동일한 클래스 정의로부터 다양한 객체를 생성할 수 있으므로, 소프트웨어의 유연성을 크게 향상시킵니다.

2. **재사용성**:
   - 클래스 정의를 재사용하면서도 서로 다른 설정을 가진 객체를 생성할 수 있으므로, 코드의 재사용성을 높일 수 있습니다.

3. **유지보수 용이**:
   - 매개변수화된 클래스나 객체는 변경 사항을 더 쉽게 관리할 수 있으며, 유지보수가 용이합니다.

## 결론

Parametric Instance는 소프트웨어 공학에서 중요한 개념으로, 매개변수를 통해 다양한 인스턴스를 생성할 수 있게 하여 유연성과 재사용성을 높입니다. 이는 제네릭 프로그래밍이나 디자인 패턴에서 자주 사용되며, 소프트웨어의 유지보수와 확장성을 크게 향상시킬 수 있습니다.

## 참고 자료

- [Java Generics Tutorial](https://docs.oracle.com/javase/tutorial/java/generics/index.html)
- [C++ Templates](https://www.cplusplus.com/doc/oldtutorial/templates/)
- [Design Patterns: Elements of Reusable Object-Oriented Software](https://en.wikipedia.org/wiki/Design_Patterns)
