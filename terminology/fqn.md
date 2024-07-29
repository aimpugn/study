# FQN

- [FQN](#fqn)
    - [FQN?](#fqn-1)
    - [언어별 예제](#언어별-예제)
        - [1. PHP](#1-php)
        - [2. Go](#2-go)
        - [3. Java](#3-java)
        - [4. Kotlin](#4-kotlin)

## FQN?

Fully Qualified Name (FQN)와 Fully Qualified Class Name (FQCN)은 프로그래밍에서 객체, 클래스, 함수 또는 변수를 문맥에 관계없이 명확하게 식별할 수 있는 전체 경로를 의미합니다.
특히 계층 구조 내에서 FQN은 계층 순서상의 모든 이름과 주어진 요소의 이름을 포함하여 완전하다고 여겨집니다.
이는 네임스페이스나 패키지 구조를 포함하여 클래스, 함수, 변수 등의 충돌을 방지하고 명확하게 식별할 수 있도록 도와줍니다.
이렇게 함으로써 동일한 이름의 객체나 클래스가 여러 네임스페이스나 패키지에서 중복될 수 있는 상황을 방지할 수 있습니다.

- PHP: `namespace`와 `use` 키워드를 사용하여 네임스페이스와 클래스를 관리합니다.
- Go: `import` 키워드를 사용하여 패키지를 가져오고, 패키지 이름을 통해 함수나 타입에 접근합니다.
- Java: `package`와 `import` 키워드를 사용하여 패키지와 클래스를 관리합니다.
- Kotlin: `package`와 `import` 키워드를 사용하여 패키지와 클래스를 관리합니다.

각 언어는 네임스페이스 또는 패키지를 통해 FQN과 FQCN을 관리하며, 이를 통해 코드의 모듈화와 충돌 방지를 도모합니다.

## 언어별 예제

### 1. PHP

- Fully Qualified Class Name (FQCN): 네임스페이스와 클래스 이름을 포함한 전체 경로.

    ```php
    namespace App\Controllers;

    // FQCN: \App\Controllers\UserController
    class UserController {
        // ...
    }

    // FQN: \App\Controllers\convert
    function convert() {
        // 함수 내용
    }

    ```

- 네임스페이스 관리: `namespace` 키워드를 사용하여 네임스페이스를 정의하고, `use` 키워드를 사용하여 네임스페이스를 가져옵니다.

    ```php
    use App\Controllers\UserController;

    $controller = new UserController();
    ```

### 2. Go

- Fully Qualified Name (FQN): 패키지 이름과 함수 또는 타입 이름을 포함한 전체 경로.

    ```go
    package main

    import (
        "fmt"
        "mypackage/subpackage"
    )

    // FQN: subpackage.MyFunction
    func main() {
        subpackage.MyFunction()
    }

    ```

- 패키지 관리: `import` 키워드를 사용하여 패키지를 가져오고, 패키지 이름을 통해 접근합니다.

    ```go
    import "mypackage/subpackage"

    subpackage.MyFunction()
    ```

### 3. Java

- Fully Qualified Class Name (FQCN): 패키지 이름과 클래스 이름을 포함한 전체 경로.

    ```java

    // FQCN: com.myapp.utilities.Converter
    public class Converter {
        // 클래스 내용
    }
    
    // FQN: com.myapp.utilities.Converter.convert
    public static void convert() {
        // 메서드 내용
    }
    ```

- 패키지 관리: `package` 키워드를 사용하여 패키지를 정의하고, `import` 키워드를 사용하여 패키지를 가져옵니다.

    ```java
    import com.myapp.utilities.Converter;

    public class Main {
        public static void main(String[] args) {
            Converter converter = new Converter();
            Converter.convert();
        }
    }
    ```

### 4. Kotlin

- Fully Qualified Class Name (FQCN): 패키지 이름과 클래스 이름을 포함한 전체 경로.

    ```kotlin
    package com.example.controllers

    // FQCN: com.example.controllers.UserController
    class UserController {
        // ...
    }

    // FQN: com.example.controllers.convert
    fun convert() {
        // 함수 내용
    }
    ```

- 패키지 관리: `package` 키워드를 사용하여 패키지를 정의하고, `import` 키워드를 사용하여 패키지를 가져옵니다.

    ```kotlin
    import com.example.controllers.UserController
    import com.example.controllers.convert

    fun main() {
        val controller = UserController()
        convert()
    }
    ```
