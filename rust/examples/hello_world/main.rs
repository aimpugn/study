// main 함수는 Rust 프로그램에서 항상 가장 먼저 실행되는 코드
fn main() {
    // 탭 대신 4 스페이스 사용
    println!("Hello, world!"); // 매크로 호출. 함수 호출이면 `!` 없이 호출.
}

// `gcc` 또는 `clang`처럼 `rustc` 사용해서 러스트 프로그램 실행 전에 컴파일 해야 한다.
// Rust는 ahead-of-time 컴파일된 언어.
// 따라서 ruby, js, python 등과 달리 실행 파일을 공유하면 Rust 설치 없이 실행 가능.
// $ rustc main.rs
