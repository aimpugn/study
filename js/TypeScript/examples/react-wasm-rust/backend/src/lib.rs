use wasm_bindgen::prelude::*;

// Rust에서 WebAssembly로 export할 함수 정의
#[wasm_bindgen]
pub fn add_one(input: i32) -> i32 {
    input + 1
}
