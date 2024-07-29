# `form_urlencoded::parse` 함수

- [`form_urlencoded::parse` 함수](#form_urlencodedparse-함수)
    - [`form_urlencoded::parse` 분석](#form_urlencodedparse-분석)
    - [함수 호출 흐름 정리](#함수-호출-흐름-정리)
    - [구현 방식 및 배경 지식](#구현-방식-및-배경-지식)
        - [1. 관심사의 분리(Separation of Concerns)](#1-관심사의-분리separation-of-concerns)
        - [2. 타입 시스템 활용과 성능 최적화](#2-타입-시스템-활용과-성능-최적화)
        - [3. 사용자에게 명확한 선택 제공](#3-사용자에게-명확한-선택-제공)
    - [결론](#결론)

## `form_urlencoded::parse` 분석

```rs
/// Convert a byte string in the `application/x-www-form-urlencoded` syntax
/// into a iterator of (name, value) pairs.
///
/// Use `parse(input.as_bytes())` to parse a `&str` string.
///
/// The names and values are percent-decoded. For instance, `%23first=%25try%25` will be
/// converted to `[("#first", "%try%")]`.
#[inline]
pub fn parse(input: &[u8]) -> Parse<'_> {
    Parse { input }
}

#[derive(Copy, Clone)]
pub struct Parse<'a> {
    //          ^^^^ 생명주기 매개변수
    input: &'a [u8],
    // ^^^^^^^ `Parse` 인스턴스는 input이 유효한 동안에만 사용될 수 있음을 컴파일러에게 보장
}
```

`parse` 함수는 바이트 슬라이스(`&[u8]`) 입력을 받아, `Parse` 구조체의 인스턴스를 반환한다.
이 함수의 주석에서는 `&str` 타입의 문자열을 파싱하기 위해 `input.as_bytes()`를 사용하는 방법을 제시하고 있다.
반환되는 `Parse` 구조체는 입력 데이터에 대한 참조를 내부에 보관하며, 이 데이터는 퍼센트 인코딩이 해제된 이름과 값의 쌍을 순회하는 데 사용된다.

`Parse` 구조체는 생명주기 매개변수 `'a`를 가지며, 이는 *구조체가 보관하는 입력 데이터의 참조와 동일한 생명주기*를 갖는다.
이를 통해 `Parse` 인스턴스는 *입력 데이터가 유효한 동안에만 사용될 수 있음을 컴파일러에게 보장*한다.

구조체는 `#[derive(Copy, Clone)]` 어노테이션을 통해 복사와 클론이 가능하도록 설정되어 있다.
이는 `Parse` 구조체가 슬라이스의 참조만을 보관하고 있기 때문에, 깊은 복사가 아닌 참조의 복사로 구현이 가능하다.

```rs

    for (key, value) in pairs.into_owned() {
        //              ^^^^^ ^^^^^^^^^^ `ParseIntoOwned`
        //              ^`Parse`
        //
        ... Do something with key and value
    }
```

실제 파싱 로직과 `into_owned` 메서드는 코드에 포함되어 있지 않지만, 이 메서드는 퍼센트 인코딩된 이름과 값의 쌍을 디코딩하고, 소유권을 가진 데이터(여기서는 `String` 타입)로 변환하는 역할을 한다.
`Parse` 구조체의 이터레이터가 순회될 때, 각 요소(이름과 값의 쌍)는 원래 바이트 슬라이스의 일부에 대한 참조를 제공한다.
`into_owned`는 이 참조들을 실제 `String` 데이터로 변환하며, 이는 퍼센트 디코딩 과정을 포함할 수 있다.
예를 들어, `%23first=%25try%25`는 `("#first", "%try%")`로 변환됩니다.

`into_owned`와 같은 메서드는 이터레이터를 통해 얻은 데이터를 소유권을 가진 `String`으로 변환하는 역할을 한다.

```rs
impl<'a> Parse<'a> {
    /// Return a new iterator that yields pairs of `String` instead of pairs of `Cow<str>`.
    pub fn into_owned(self) -> ParseIntoOwned<'a> {
        ParseIntoOwned { inner: self }
    }
}

/// Like `Parse`, but yields pairs of `String` instead of pairs of `Cow<str>`.
pub struct ParseIntoOwned<'a> {
    inner: Parse<'a>,
}
```

`next` 메서드는 `Parse` 이터레이터의 `next` 메서드를 호출하여 이름과 값의 쌍을 받아온 후, `into_owned` 메서드를 사용하여 각 `Cow<'a, str>` 인스턴스를 `String` 타입으로 변환한다.
이 변환 과정은 퍼센트 디코딩된 데이터를 소유한 `String`으로 만들어, 더 이상 입력 데이터의 생명주기에 종속되지 않게 한다.

```rs
impl<'a> Iterator for ParseIntoOwned<'a> {
    type Item = (String, String);

    fn next(&mut self) -> Option<Self::Item> {
        self.inner
        //   ^^^^^ `Parse` 이터레이터의 next 호출하여 이름과 값의 쌍을 받아온다
            .next()
            .map(|(k, v)| (k.into_owned(), v.into_owned()))
            //             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
            //             디코딩된 데이터를 소유한 `String`으로 만들어,
            //             더 이상 입력 데이터의 생명주기에 종속되지 않게 한다.
    }
}
```

실제 파싱과 디코딩 로직은 `Parse` 구조체의 메서드로 구현되어 있다.
`ParseIntoOwned`를 반복하면 `ParseIntoOwned::next()`가 호출된다.
그러면 그 `ParseIntoOwned::next` 내부에서 `Parse::next()`가 호출되면서 입력 데이터의 생명주기로부터 벗어난 키-밸류 쌍을 반환한다.

```rs
impl<'a> Iterator for Parse<'a> {
    type Item = (Cow<'a, str>, Cow<'a, str>);

    fn next(&mut self) -> Option<Self::Item> {
        loop {
            if self.input.is_empty() {
                return None;
            }
            let mut split2 = self.input.splitn(2, |&b| b == b'&');
            let sequence = split2.next().unwrap();
            self.input = split2.next().unwrap_or(&[][..]);
            if sequence.is_empty() {
                continue;
            }
            let mut split2 = sequence.splitn(2, |&b| b == b'=');
            let name = split2.next().unwrap();
            let value = split2.next().unwrap_or(&[][..]);
            return Some((decode(name), decode(value)));
        }
    }
}
```

## 함수 호출 흐름 정리

1. **`form_urlencoded::parse` 호출**

    URL 인코딩된 문자열을 파싱하여, 이름과 값의 쌍을 순회할 수 있는 `Parse<'_>` 타입의 이터레이터를 반환한다.
    여기서 `'_`는 컴파일러가 적절한 생명주기를 추론할 수 있게 해주는 생명주기 엘리젼(`lifetime elision`) 규칙이다.

    > **생명주기 엘리젼(`lifetime elision`)?**
    >
    > 엘리젼(elision)이란 단어는 "생략"을 의미한다.
    > 생명주기 엘리젼이란  *함수나 메서드의 파라미터*와 *반환값* 사이의 생명주기 관계를 컴파일러가 자동으로 추론하는 기능을 말한다.
    > 명시적으로 생명주기 매개변수를 작성하지 않아도, 일반적인 패턴에 따라 생명주기를 "생략"하고 컴파일러가 내부적으로 추론할 수 있다는 의미다.
    >
    > **`'static` 생명주기?**
    >
    > `'static` 생명주기는 해당 데이터가 프로그램의 전 생애 주기 동안 유효하다는 것을 의미한다.
    > 만약 `Parse<'static>`으로 표시하는 경우, 이는 사용된 문자열 리터럴이나 컴파일 시점에 결정되는 상수 등이 전달되어, 반환된 `Parse` 인스턴스가 프로그램 전체 실행 기간 동안 유효한 `'static` 생명주기를 갖게 되었음을 의미한다.

2. **`pairs.into_owned()` 호출**

    이 메서드는 `Parse<'_>` 이터레이터를 `ParseIntoOwned<'_>` 타입으로 변환한다.
    `ParseIntoOwned`는 `Parse`와 유사하게 동작하지만, 이름과 값의 쌍을 `String` 타입으로 반환하므로, 퍼센트 디코딩 등의 변환을 거쳐 소유권을 가진 데이터를 필요로 할 때 유용하다.

3. **반복문을 통한 순회**

    `ParseIntoOwned` 인스턴스를 반복문에서 사용하면, 내부의 `Parse` 인스턴스(`inner`)의 `next` 메서드를 호출하여 이름과 값의 쌍을 `String` 타입으로 순회할 수 있다. 각 반복마다 `Cow<str>` 타입에서 `String` 타입으로의 변환(`into_owned`)이 이루어진다.

## 구현 방식 및 배경 지식

이러한 구현 방식은 Rust의 *타입 시스템*과 *소유권 모델*을 활용하여, *메모리 사용과 성능 최적화 사이의 균형*을 이루고자 하는 철학에 기반한다.

- **"레이지 평가"(Lazy Evaluation) 패턴**

    `Parse`의 `into_owned` 메서드가 `ParseIntoOwned` 인스턴스를 반환하는 구조는, Rust의 함수형 프로그래밍 패러다임과 타입 시스템을 활용하는 일종의 "레이지 평가"(Lazy Evaluation) 패턴을 따른다. 즉, 실제로 필요한 시점까지 데이터 변환을 미루고, 필요한 변환 작업을 최소화하여 실행 시점의 성능을 최적화한다.

    이러한 설계는 Rust가 메모리 안전성과 효율성을 중시하는 언어 철학과 일치한다.
    또한, 이는 Rust 개발자들이 성능과 안전성 사이에서 균형을 이루기 위해 고려하는 실무적인 접근 방식을 반영한다.

    - **불필요한 메모리 할당 감소**(메모리 안전성)

        레이지 평가를 사용하면, 실제로 필요하지 않은 데이터에 대한 처리를 피할 수 있다.
        이는 프로그램이 불필요한 메모리를 할당하고 해제하는 작업을 줄임으로써 메모리 사용의 효율성을 높이고, 메모리 누수의 가능성을 감소시킨다.

    - **안전한 자원 관리**(메모리 안전성)

        Rust에서 메모리 안전성은 핵심 원칙 중 하나다.
        데이터 처리가 필요한 순간까지 메모리 접근을 지연시키면, 불필요한 메모리 접근을 방지하고, 안전하게 자원을 관리할 수 있다.

    - **보다 안전한 타입 변환**(타입 안전성)

        `ParseIntoOwned`와 같은 구조에서 레이지 평가를 사용하면, 데이터의 타입 변환(예: `Cow<str>`에서 `String`으로)이 실제로 필요한 순간에만 수행된다. 개발자가 타입 변환의 시기를 명확하게 제어할 수 있게 하며, 변환의 필요성을 명확히 이해하는 데 도움을 준다. 이는 타입 안전성을 더욱 강화하는 동시에, 타입 변환 과정에서 발생할 수 있는 오류를 최소화한다.

    - **계산의 최적화**(효율성)

        레이지 평가는 필요한 데이터만을 계산하므로, 불필요한 계산을 줄이고 전체 프로그램의 효율성을 향상시킨다.
        특히 큰 데이터 세트나 복잡한 변환 작업에서 이점이 두드러진다.

    - **성능 개선**(효율성)

        메모리 할당과 해제, 그리고 데이터 처리에 대한 오버헤드를 줄임으로써, 레이지 평가는 프로그램의 실행 속도를 개선할 수 있다.
        실제로 필요한 작업만을 수행하기 때문에, 전반적인 성능이 최적화된다.

- **메모리 효율성**

    `Cow<str>` 타입을 사용함으로써, 가능한 경우 원본 데이터에 대한 참조를 유지하면서 메모리 할당을 최소화할 수 있다.
    이는 메모리 사용량을 줄이는 데 도움이 된다.

    > **`Cow` 타입과 메모리 효율성**
    >
    > `Cow`는 "Copy on Write"의 약자로, *데이터를 복사하기 전까지는 원본 데이터에 대한 참조를 유지*하는 타입이다.
    >
    > `Cow`는 두 가지 형태를 가질 수 있다:
    > - `Borrowed`: 데이터에 대한 참조를 의미
    > - `Owned`: 데이터의 소유권이 있는 경우를 의미
    >
    > `Cow` 타입을 사용하는 이유는 퍼포먼스와 메모리 효율성 때문이다.
    > 가능한 한 원본 데이터를 수정하지 않고 참조를 유지함으로써 *불필요한 데이터 복사를 피할 수 있다*.
    > 이는 특히 데이터의 크기가 크거나 복사 비용이 높은 경우 유용하다.
    >
    > `into_owned` 메서드를 호출하면 `Cow`는 `Owned` 상태가 되고, 데이터의 복사본을 만들고 소유권을 가지게 된다.
    >
    > 별도의 소유권은 다음과 같은 경우에 필요하다:
    > - 원본 데이터를 변경해야 하는 경우
    > - 원본 데이터의 생명주기가 끝난 후에도 데이터를 사용해야 하는 경우 등
    >
    > 결국 `into_owned`를 호출하는 순간 메모리 할당이 발생하게 되지만, 이는 필요한 경우, 즉 `into_owned`가 호출되는 경우에만 발생한다.
    > 따라서 `Cow`를 사용함으로써 불필요한 상황에서의 메모리 할당을 피할 수 있다.

- **유연성과 소유권 제공**

    특정 사용 사례에서는 변환된 데이터에 대한 소유권이 필요할 수 있다.
    예를 들어, 변환된 데이터를 다른 스레드로 전송하거나, 데이터의 생명주기를 호출 스택에 제한하지 않고 관리하고자 할 때 `String` 타입이 유용하다.
    `ParseIntoOwned`를 통해 이러한 요구 사항을 충족시킨다.

- **성능 최적화**

    사용자가 필요로 하는 경우에만 `String`으로의 변환을 수행함으로써, 불필요한 메모리 할당과 복사 작업을 방지한다.
    이는 성능 최적화의 한 예로, Rust의 "zero-cost abstractions" 철학을 따른다.

### 1. 관심사의 분리(Separation of Concerns)

- **유연성과 단일 책임 원칙**

    `Parse` 구조체는 파싱 로직에 초점을 맞추며, `ParseIntoOwned`는 데이터의 소유권 변환에 초점을 맞춘다.
    이는 각 구조체가 단일 책임 원칙을 따르며, 유연성을 제공한다.
    사용자는 파싱만 필요할 때 불필요한 소유권 변환 로직을 실행하지 않아도 되며, 필요에 따라 데이터 소유권 변환을 선택할 수 있다.

- **확장성과 유지보수성**

    미래에 파싱 로직이나 소유권 변환 로직에 변경이 필요할 경우, 관심사가 분리되어 있으면 해당 로직만 수정하면 된다.
    이는 코드의 유지보수성과 확장성을 향상시킨다.

### 2. 타입 시스템 활용과 성능 최적화

- **타입 안전성**

    `Parse`와 `ParseIntoOwned`는 서로 다른 타입의 이터레이터를 반환한다(`Cow<str>` 대 `String`).
    Rust의 타입 시스템을 통해 컴파일 타임에 이러한 차이를 명확히 함으로써, 타입 안전성과 명확한 API 사용을 보장한다.

- **성능 최적화**

`Parse`는 원본 데이터에 대한 참조를 유지함으로써 불필요한 메모리 할당을 피할 수 있다.
사용자가 명시적으로 `into_owned`를 호출하기 전까지는, 원본 데이터를 변형하지 않고 참조만 사용한다.
이는 메모리 사용량과 실행 시간에 있어 중요한 최적화가 될 수 있다.

### 3. 사용자에게 명확한 선택 제공

- **명시적인 API 설계**

    `ParseIntoOwned` 구조체를 제공함으로써, 개발자는 데이터를 파싱할 때와 파싱된 데이터의 소유권을 얻고자 할 때의 차이를 명확히 인식할 수 있다. 이는 API 사용 시의 의도를 명확히 하며, 코드를 읽고 이해하기 쉽게 만든다.

## 결론

이러한 구현은 Rust의 핵심 원칙 중 하나인 "성능, 안전성, 그리고 메모리 효율성 사이의 균형"을 잘 보여준다.
사용자는 필요에 따라 메모리 사용량을 최소화할 수 있는 `Cow<str>` 타입과, 필요한 경우 데이터의 소유권을 완전히 가질 수 있는 `String` 타입 사이를 선택할 수 있다. 이러한 설계는 Rust가 시스템 프로그래밍 언어로서 강력한 성능과 안전성을 제공하는 방법의 일례를 보여준다.
