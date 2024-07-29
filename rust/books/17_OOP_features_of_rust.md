# 17. OOP features of Rust

## [17.3 Implementing and Object Oriented Design Pattern](https://doc.rust-lang.org/book/ch17-03-oo-design-patterns.html)

### Examples

#### `impl`

```rs
/// The return type of `parse()`.
#[derive(Copy, Clone)]
pub struct Parse<'a> {
    input: &'a [u8],
}

impl<'a> Parse<'a> {
    ... 생략 ...
}
```

- **`impl`**

    Rust에서 특정 타입에 대한 메서드나 연관 함수(`associated functions`)를 구현할 때 사용된다.
    여기서는 `Parse` 구조체에 대한 구현을 시작함을 의미한다.

- **`<'a>`**

    이것은 생명주기 매개변수(`lifetime parameter`)다.
    Rust의 생명주기 매개변수는 *참조가 유효한 범위를 컴파일러에게 알려주는 역할*을 한다.
    여기서 `'a`는 `Parse` 구조체가 참조하는 데이터(`input: &'a [u8]`)의 생명주기를 나타내며,
    이 구조체와 관련된 모든 참조는 같은 생명주기 `'a`를 공유한다.

- **`Parse<'a>`**

    `Parse` 구조체에 생명주기 매개변수 `'a`를 명시함으로써, 이 구조체의 인스턴스가 특정 데이터의 참조를 보유하고 있음을 나타낸다.
    해당 데이터는 `'a` 생명주기 동안 유효하다.

```rs
impl<'a> Parse<'a> {
    /// Return a new iterator that yields pairs of `String` instead of pairs of `Cow<str>`.
    pub fn into_owned(self) -> ParseIntoOwned<'a> {
        //                                   ^^^^ 반환되는 `ParseIntoOwned` 인스턴스 내부적으로 참조하는 데이터가
        //                                        `'a` 생명주기 동안 유효함을 보장
        ParseIntoOwned { inner: self }
    }
}
```

이 메서드는 기본적으로 `Parse` 인스턴스를 받아, 이를 `ParseIntoOwned` 인스턴스로 변환하는 "소유권 이전"(`ownership transfer`) 패턴을 구현한다. 이 변환을 통해, 원래 `Parse` 이터레이터가 생성하는 `Cow<str>` 타입의 이름과 값의 쌍을 `String` 타입으로 소유하는 새로운 이터레이터를 얻게 된다. 이 과정은 퍼센트 디코딩이나 기타 변환 작업 후에 소유권을 가진 데이터를 필요로 할 때 유용하게 사용될 수 있다.

- **`pub`**

    "public"의 약자로, 이 메서드가 현재 모듈의 외부에서도 접근 가능함을 의미한다.

- **`fn`**

    "function"의 약자로, 함수를 정의할 때 사용된다.

- **`into_owned(self)`**

    `into_owned`라는 이름의 메서드를 정의한다.
    `self` 키워드는 이 메서드가 `Parse` 구조체의 인스턴스에 대해 호출될 수 있음을 나타낸다.
    이 인스턴스는 메서드 호출 후에는 더 이상 사용할 수 없다.(`self`를 사용하면 메서드 내에서 인스턴스의 소유권을 가져간다).

- **`-> ParseIntoOwned<'a>`**

    이 메서드가 `ParseIntoOwned<'a>` 타입의 값을 반환함을 나타낸다.
    여기서도 생명주기 매개변수 `'a`가 사용되고 있다.
    이는 반환되는 `ParseIntoOwned` 인스턴스 내부적으로 참조하는 데이터가 `'a` 생명주기 동안 유효함을 보장한다.

- **`ParseIntoOwned { inner: self }`**

    `ParseIntoOwned` 구조체의 인스턴스를 생성하며, 해당 구조체의 `inner` 필드에 `self` (즉, 현재 `Parse` 인스턴스)를 할당한다.
    이 과정에서 `Parse` 인스턴스의 소유권이 `ParseIntoOwned` 인스턴스로 이전된다.
