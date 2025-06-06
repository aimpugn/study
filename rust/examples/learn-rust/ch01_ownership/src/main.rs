fn main() {
    // 스택과 힙의 구별이 소유권 규칙의 출발점입니다.
    // - 스칼라 타입(i32, bool 등) 경우:
    //
    //      스칼라는 결국 언제나 같은 크기, 같은 메모리 배치라는 불변 특성을 지닌 값을 가리킵니다.
    //      따라서 "복사"는 바로 "값 복제"를 의미합니다.
    //      - 정수형(i8~i128, isize, usize)
    //      - 부동소수점(f32, f64)
    //      - boolean
    //      - char(유니코드)
    //      "값 크기가 고정"이고, "복사 비용이 예측 가능"하므로, 값 자체를 스택에 저장합니다.
    //      그래서 이동(move) 대신 단순 복사(copy)를 허용해도 성능 손실이 거의 없고,
    //      별도의 힙 해제(drop) 로직이 필요 없습니다.
    //
    // - 문자열, 벡터처럼 크기가 달라지는 경우
    //
    //      실제 데이터가 가변 크리가 스택에 담을 수가 없습니다. 따라서,
    //      - 힙에 실제 데이터를 저장합니다.
    //      - 스택에는 '포인터 + 길이 + 용량' 등의 메타데이터를 저장합니다.
    //
    //      소유권이 이동할 때는 메타데이터만 옮겨 가고, 힙의 실제 데이터는 그대로 남습니다.
    //      그리고 원 소유자는 더이상 그 데이터를 만질 수 없게 막아야 메모리 안전이 보장됩니다.

    // 힙에 다섯 글자짜리 메모리 블록을 할당합니다.
    // 그리고 그 주소, 길이, 용량 세 필드를 스택 프레임에 올려 놓습니다.
    // 이 '세 필드를 가지고 있는 변수'가 곧 '소유권을 가진 주체'입니다.
    // 스코프가 종료될 때 러스트는 "당신이 마지막 소유자였다"라는 사실을 근거로
    // `drop()`을 호출하여 힙 블록을 정확히 한 번만 해제합니다.
    let s1 = String::from("hello");
    let s2 = s1; // 메타데이터를 s2로 '이동'시킵니다.
    let s3 = String::from("world");

    // 이제 s1은 더 이상 유효하지 않게 되므로 다음 코드에서 컴파일 에러가 발생합니다.
    // println!("{}", s1); // 컴파일 오류!
    //                ^^ value borrowed here after move

    // Copy 트레이트를 구현한 스칼라형
    let n1: i32 = 42;
    let n2 = n1; // 복사(copy)로 간주되며 둘 다 유효합니다.
    println!("n1 = {n1}, n2 = {n2}");
    // Output:
    //  n1 = 42, n2 = 42
    let n3 = n1;
    println!("n3 = {n3}");

    // Copy와 Clone의 차이점은 다음과 같습니다.
    // - Copy: 이동처럼 사용해도 자동으로 복사
    // - Clone: 복사를 호출해야 함
    // 그래서 앞서 봤듯이 Copy 구현이 된 스칼라형은 자동으로, 암묵적으로 복사가 이뤄지며,
    // 원래 변수와 새로 복사가 이뤄진 변수 모두 유효합니다.
    let n1: u32 = 5; // 스택에 4바이트
    let n2 = n1; // 이동이 아니라 "복사"가 이뤄집니다.
    println!("n1 = {n1}, n2 = {n2}"); // 둘 다 살아있습니다.

    let heap1 = String::from("clone demo");
    // let heap2 = heap1; // 이 경우 메타데이터의 이동이 이뤄집니다.
    let heap2 = heap1.clone(); // 이 경우 깊은 복사(Clone)가 이뤄집니다.
    println!("heap1: {heap1}, heap2: {heap2}");
    let heap3 = heap1; // heap1의 lifetime은 여기까지 연장됩니다.
    println!("heap3: {heap3}");

    // heap3에서 이동이 이뤄졌기 때문에 이제 컴파일 에러가 발생합니다.
    // println!("heap1 again: {heap1}");
    // ```
    // let heap1 = String::from("clone demo");
    //     ----- move occurs because `heap1` has type `String`, which does not implement the `Copy` trait
    // let heap3 = heap1; // heap1의 lifetime은 여기까지 연장됩니다.
    //             ----- value moved here
    // println!("heap3: {heap3}");
    // println!("heap1 again: {heap1}");
    //                        ^^^^^^^ value borrowed here after move
    // ```

    // `String` 타입을 인자로 받는 함수의 경우:
    // 호출 시점에 러스트는 세 필드(주소, 길이, 용량)를 통째로 복사해 새 스택 프레임으로 옮겨 두고,
    // 원래 변수 이름을 "더 이상 접근 불가"로 표시합니다.
    // 이 시점까지는 `s2`가 소유권을 가진 상태입니다.
    let len_of_s3 = get_length(s3);
    //                                -- value moved here
    println!("len_of_s3 = {len_of_s3}");
    // println!("s3 after get_length: {s3}");
    //                                ^^^^ value borrowed here after move

    // 그런데 위와 같은 경우 함수 호출이 연쇄될 때 문제가 됩니다.
    // `get_length`가 내부에서 다시 `get_length2(msg)`를 호출한다면,
    // 소유권은 또다시 이동하고 해제 책임자도 한 번 더 바뀝니다.
    // 결과적으로 가장 안쪽 함수가 끝날 때까지 힙 블록을 해제할 수 없고,
    // 호출자 쪽에서는 변수 이름을 도중에 사용할 방법이 없습니다.
    // 매번 반납했다 다시 받아오는 식으로 짜면 코드가 장황해지고,
    // 스택 언와인딩(unwinding) 과정에서 `drop()`이 언제 실행될지 추론하기도 복잡해집니다.
    // 이에 대한 해결책으로 제시된 것이 대여(borrowing)입니다.

    // `&String` 하나를 인자로 받으면, 힙 주소와 길이만 복사해서 전달합니다.
    // 이로 인해서 소유권의 귀속 정보는 바꾸지 않습니다.
    // 따라서 함수에서는 읽기(또는 `&mut`이면 쓰기)만 하고,
    // 블록을 해제할 책임은 여전히 원래 변수, 즉 호출자에게 남아 있습니다.
    // 이렇게 하면 호출 체인이 아무리 길어져도 "메모리가 해제되는 시점"은 항상
    // "가장 바깥쪽 원 소유자의 스코프 종료 순간"이 됩니다.
    let len = calculate_length(&s2);
    println!("len = {len}");

    // 정리하자면,
    // - 소유권:
    //     - 이동할 때마다 해제 책임자를 바꿉니다.
    //     - 값을 한 번만 써야 하고 호출자에게 남길 필요가 없을 때는 '이동'을 사용합니다.
    // - 대여:
    //     - 해제 책임자를 바꾸지 않고 잠시 권한만 나눠줍니다.
    //     - 호출자도 써야 하거나 여러 곳이 관찰해야 할 때는 '대여'를 사용합니다.
    //
    // 핵심은 "메모리를 해제할 권한자가 프로그램 흐름 안에서 단 하나만 존재"하도록 소유권 타입을 설계했습니다.
    // 가령 C에서는 포인터 복사가 저렴하고 자유롭다 보니 권한자를 추적하기 어렵습니다.
    // 하지만 러스트는 '복사할 수 있는 값'과 '이동해야 하는 값'을 타입 체계에서 엄격히 구분합니다.
    // 그 결과 댕글링 포인터와 이중 해제를 막기 위한 런타임 레퍼런스 카운팅조차 필요 없고,
    // 컴파일 타임 규칙만으로 메모리 주소값을 정교하게 조작해서 시스템의 의도치 않은 위치를 읽거나 덮어쓰는 행위 등을 막을 수 있습니다.

    let mut s3 = String::from("original");
    // 한 시점에 단 하나의 가변 참조만 존재할 수 있습니다.
    // 하지만 다음은 매번 새로운 가변 참조를 만들기 때문에 가능합니다.
    change(&mut s3, 1); // 가변 참조 A 생성 -> 함수에 잠시 넘김 -> 끝나면 drop
    change(&mut s3, 2); // 가변 참조 B 생성 -> 함수에 잠시 넘김 -> 끝나면 drop
    println!("s3: {s3}"); // 원래 변수 s3를 다시 사용

    // rust는 스코프에 들어온 참조가 언제까지 사용되는가를 정확히 추적하며,
    // 이를 life time이라 부릅니다.
    // 컴파일 타임에 어떤 참조가 아직 사용된다고 판단되면,
    // 그 동안은 어떤 다른 가변 참조도 허용되지 않습니다.
    // let r1 = &mut s3;
    // error[E0499]: cannot borrow `s3` as mutable more than once at a time
    // let r2 = &mut s3; // `r1`이 살아있기 때문에 불가능합니다.
    // println!("r1: {r1}, r2: {r2}");
    // change(r1);

    // 하지만 `r1`이 사용되고 그 lifetime이 끝나는 게 명확하다면, 다시 불변 참조가 가능해집니다.
    {
        let r1 = &mut s3;
        change(r1, 3);
        println!("r1: {r1}"); // 이 라인이 끝나는 순간, `r1`의 마지막 사용으로 간주됩니다.

        // 그리고 스코프가 닫히기 전에, `r1`은 이미 소멸된 것으로 추론됩니다.
    } // 스코프 종료로 `r1`은 더이상 유효하지 않습니다.

    // 또는 바로 함수를 호출하면 일시적인 참조로 간주되어 lifetime이 좁게 추론됩니다.
    change(&mut s3, 4);

    let r2 = &mut s3;
    change(r2, 5);
    println!("r2: {r2}");

    // 불변 참조는 여러번 할 수 있습니다.
    let r3 = &s3; // 불변 참조 1
    let r4 = &s3; // 불변 참조 2

    // 불변 참조가 살아 있는 동안에는 가변 참조가 불가능합니다.
    // - 하나 이상의 불변 참조(`&T`)가 존재할 경우 가변 참조(`&mut T`)를 만들 수 없습니다.
    // - 가변 참조(`&mut T`)가 이미 존재하면, 불변 참조(`&T`)도 만들 수 없습니다.
    //
    // 동일한 값을 읽고 있는 참조들(`r3`, `r4`)이 존재하는 상태에서 `s3`의 값을 수정하면,
    // 읽고 있는 쪽에서 읽은 값이 정확하다는 보장을 할 수 없기 때문입니다.
    // ```log
    // error[E0502]: cannot borrow `s3` as mutable because it is also borrowed as immutable
    // 99  |     let r3 = &mut s3;
    //     |              ^^^^^^^ mutable borrow occurs here
    // ```
    // 이미 불변 참조 두 개가 존재하므로, 동시에 가변 참조를
    // 만들면 '잠재적 데이터 레이스'로 간주합니다.
    println!("r3: {r3}, r4: {r4}"); // `r3`, `r4`의 lifetime은 여기서 종료됩니다.

    // 단, `r3`, `r4`를 다시 사용하면,
    // 두 참조의 생존 범위는 두 번째 println! 호출이 끝나는 시점까지로 확정되고,
    // 그 직후 더이상 사용되지 않는다고 판단됩니다.
    println!("r3: {r3}, r4: {r4}");

    // 그래서 이 시점에서 `r3`, `r4`의 lifetime은 확실히 종료됩니다.

    let r5 = &mut s3;
    println!("r5: {r5}");

    // 만약 `r3`, `r4`의 lifetime이 여기까지 확장되어 여전히 살아있다고 판단하게 된다면,
    // 위의 `&mut s3`에서 다음과 같은 컴파일 에러가 발생합니다.
    //
    // ```log
    // cannot borrow `s3` as mutable because it is also borrowed as immutable mutable borrow occurs here
    // ```
    // println!("r3: {r3}, r4: {r4}");

    let text = String::from("ownership");
    let first = first_word(&text); // &str은 text의 '부분 빌림'입니다.
    println!("first word: {first}");

    let _d1 = ShowDrop("_d1");
    {
        let _d2 = ShowDrop("_d2");
        println!("inner scope about to end");
    } // `d2::drop`이 호출됩니다.

    let _d3 = ShowDrop("_d3");
    // `_d3.drop()`처럼 직접 호출하는 것은 금지되어 있으며, 호출하려고 하면 컴파일 에러가 발생합니다.
    // ```
    // _d3.drop();
    //     ^^^^ explicit destructor calls not allowed
    // ```
    //
    // 이를 허용하면, c에서 처럼 이중으로 free하여 문제가 발생할 수 있기 때문입니다.
    // - [What does "double free" mean?](https://stackoverflow.com/a/21057524)
    // - [double-free 취약점](https://showx123.tistory.com/59)
    //
    // 이러한 위험을 차단하기 위해 `std::mem::drop`, `ptr::drop_in_place` 등을
    // 사용하여 drop하는 것만 허용됩니다.
    // - [Running Code on Cleanup with the Drop Trait](https://doc.rust-lang.org/book/ch15-03-drop.html?utm_source=chatgpt.com#:~:text=Rust%20doesn%E2%80%99t%20let%20you%20call%20the%20Drop%20trait%E2%80%99s%20drop%20method%20manually%3B%20instead%2C%20you%20have%20to%20call%20the%20std%3A%3Amem%3A%3Adrop%20function%20provided%20by%20the%20standard%20library%20if%20you%20want%20to%20force%20a%20value%20to%20be%20dropped%20before%20the%20end%20of%20its%20scope.)
    std::mem::drop(_d3);
    println!("main about to end");
    // Output:
    // inner scope about to end
    //  >> drop(): '_d2' 해제
    //  >> drop(): '_d3' 해제
    // main about to end
    //  >> drop(): '_d1' 해제
} // 여기서 main 스코프가 종료됩니다.
  // `s2`, `s3`, `text`에 대해 `drop()`가 호출되어 힙 메모리를 자동 해제합니다.

struct ShowDrop(&'static str);

impl std::fmt::Display for ShowDrop {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl Drop for ShowDrop {
    fn drop(&mut self) {
        println!(" >> drop(): '{}' 해제", self.0);
    }
}

/// `String`을 인자로 받아서 그 길이를 리턴하는 함수입니다.
fn get_length(s: String) -> usize {
    print_length(s.clone());
    s.len()
}

/// `String`을 인자로 받아서 그 길이를 출력하는 함수입니다.
fn print_length(s: String) {
    println!("length of '{}' is {}", s, s.len());
}

/// 불변 참조를 받아 길이를 계산하는 함수입니다.
/// 인자로 `&String`을 받으므로 호출자는 소유권을 잃지 않습니다.
fn calculate_length(s: &String) -> usize {
    s.len() // 읽기 전용으로 힙 데이터에 접근
} // 여기서 `s`는 스코프를 벗어나지만 빌렸던 `&String`이므로 해제(drop)가 없습니다.

use std::fmt::Write; // write! 매크로 사용위해 추가

/// 가변 참조를 받아 내용을 바꾸는 함수입니다.
/// 파라미터 타입이 `&mut String`이므로 호출 시점에 해당 변수는 유일 참조여야 합니다.
fn change(s: &mut String, suffix: i32) {
    s.push_str(".");
    // s.push_str(suffix.to_string().as_str());
    //
    // `suffix.to_string()` 호출은 새로운 String 오브젝트를 생성하고,
    // `&str`로 변환하는 데에만 사용되는데, 이는 불필요한 오버헤드를 발생시킵니다.
    // 대신, 접미사(suffix)를 문자열로 직접 추가하기 위해 `format!` 매크로를 사용합니다.
    // 이를 통해 중간 할당을 피하고 성능 향상이 가능합니다.
    // s.push_str(&format!("{}", suffix));
    // s.push_str(&suffix.to_string());

    // 또는 `write!` 매크로를 사용하면 임시 버퍼를 만들지 않고
    // 바로 기존 `String`에 문자열을 집어넣을 수 있습니다.
    // 하지만 항상 그런 것은 아니고, 경우에 따라 allocator가 호출될 수 있습니다.
    // - https://stackoverflow.com/a/73188591
    write!(s, "{}", suffix).unwrap();
}

/// 입력 문자열에서 첫 공백 전까지의 슬라이스 `&str`을 리턴합니다.
/// '반환값의 수명'이 입력 파라미터와 동일하다는 사실을 컴파일러가
/// 자동 추론하기 때문에, 명시적인 lifetime 파라미터가 필요 없습니다.
fn first_word(s: &String) -> &str {
    for (i, &b) in s.as_bytes().iter().enumerate() {
        if b == b' ' {
            return &s[..i]; // 부분 빌림(slice)
        }
    }
    &s[..]
}
