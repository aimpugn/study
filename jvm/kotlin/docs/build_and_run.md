# Build and run

- [Build and run](#build-and-run)
    - [Build and run](#build-and-run-1)
    - [컴파일 및 실행](#컴파일-및-실행)
    - [컴파일 시 주의 사항](#컴파일-시-주의-사항)

## Build and run

컴파일 시 같은 패키지에 속한 파일을 함께 컴파일하면 같은 패키지 내에서는 별도의 import 문 없이 클래스를 사용할 수 있습니다.
그러나 파일을 별도로 컴파일하거나 명시적으로 다른 파일을 import해야 하는 경우에는 `import` 문을 사용해야 합니다.

```bash
jvm/
└── kotlin/
    └── src/
        └── main/
            └── kotlin/
                └── visibility/
                    ├── ProtectedAndPublic.kt
                    └── Main.kt
```

```kotlin
// `jvm/kotlin/src/main/kotlin/visibility/ProtectedAndPublic.kt`
package visibility

open class ProtectedClass {
    protected val protectedProperty: String = "protected"
}

class PublicClass : ProtectedClass() {
    fun publicFunction() {
        println("Protected Property: ${this.protectedProperty}")
    }
}
```

```kotlin
// `jvm/kotlin/src/main/kotlin/visibility/Main.kt`
package visibility

import visibility.PublicClass

fun main() {
    val publicClass = PublicClass()
    publicClass.publicFunction()
}
```

## 컴파일 및 실행

- 컴파일

    ```bash
    cd jvm/kotlin/src/main/kotlin/visibility
    kotlinc ProtectedAndPublic.kt Main.kt -include-runtime -d output.jar
    ```

    - `-include-runtime`: 생성된 JAR 파일에 Kotlin 런타임 포함
    - `-d`: 출력 파일 이름 지정

- 실행

    `kotlin` 또는 `java`로 실행 가능합니다.

    `MainKt`는 `Main.kt` 파일에서 `main` 함수를 포함하는 클래스의 이름입니다.
    Kotlin 컴파일러는 파일 이름을 기준으로 클래스 이름을 만듭니다.
    `Main.kt` 파일을 컴파일한 경우, 기본적으로 `MainKt`라는 이름의 클래스를 생성합니다.

    ```bash
    kotlin -classpath output.jar visibility.MainKt
    ```

    ```bash
    java -jar output.jar
    ```

## 컴파일 시 주의 사항

- 패키지를 올바르게 선언했는지 확인합니다.
- 패키지 경로와 파일 경로가 일치하는지 확인합니다.
- 함께 컴파일하는 경우, 같은 패키지 내에서 `import` 문 없이 접근할 수 있습니다.
- 파일을 개별적으로 컴파일하거나 다른 경로에서 접근하는 경우, 명시적으로 `import` 문을 사용해야 합니다.
