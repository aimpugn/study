# inline

## `inline` 속성의 개요

`inline` 속성은 Rust의 내장 속성 중 하나로, 컴파일러에게 함수 인라인화에 대한 힌트를 제공합니다. 이 속성은 성능 최적화를 위해 사용되며, 컴파일러의 인라인화 결정에 영향을 줍니다.

## `inline` 속성의 정의

[rustc_attr/src/builtin.rs](https://github.com/rust-lang/rust/blob/4fe1e2bd5bf5a6f1cb245f161a5e9d315766f103/compiler/rustc_attr/src/builtin.rs#L43-L49)에서 `InlineAttr` 열거형으로 정의됩니다:

```rust
#[derive(Copy, Clone, PartialEq, Encodable, Decodable, Debug, HashStable_Generic)]
pub enum InlineAttr {
    None,
    Hint,
    Always,
    Never,
}
```

이 정의는 `inline` 속성의 가능한 값들을 나타냅니다:
- `None`: 기본값, 인라인화에 대한 특별한 지시 없음
- `Hint`: 컴파일러에게 인라인화를 권장
- `Always`: 가능한 한 항상 인라인화
- `Never`: 인라인화하지 않음

## `inline` 속성의 처리 과정

`inline` 속성의 처리는 Rust 컴파일러의 여러 단계에 걸쳐 이루어집니다:

1. **구문 분석 (Parsing)**:
   [rustc_ast/src/attr/mod.rs](https://github.com/rust-lang/rust/blob/master/compiler/rustc_ast/src/attr/mod.rs)에서 `inline` 속성을 포함한 모든 속성을 파싱합니다. 이 단계에서는 속성의 구문이 올바른지 확인합니다.

2. **의미 분석 (Semantic Analysis)**:
   [rustc_attr/src/builtin.rs](https://github.com/rust-lang/rust/blob/master/compiler/rustc_attr/src/builtin.rs)에서 `inline` 속성의 의미를 해석합니다. 여기서 속성의 값(`None`, `Hint`, `Always`, `Never`)을 결정합니다.

3. **중간 표현 생성 (MIR Generation)**:
   [rustc_mir_build/src/thir/cx/expr.rs](https://github.com/rust-lang/rust/blob/master/compiler/rustc_mir_build/src/thir/cx/expr.rs)에서 MIR(Mid-level Intermediate Representation)을 생성할 때 `inline` 속성 정보를 포함시킵니다.

4. **코드 생성 (Code Generation)**:
   [rustc_codegen_ssa/src/mir/block.rs](https://github.com/rust-lang/rust/blob/master/compiler/rustc_codegen_ssa/src/mir/block.rs)에서 LLVM IR을 생성할 때 `inline` 속성 정보를 LLVM 메타데이터로 변환합니다.

5. **LLVM 최적화**:
   생성된 LLVM IR은 LLVM 최적화 파이프라인을 거치며, 이 과정에서 `inline` 속성 정보를 바탕으로 실제 인라인화 결정이 이루어집니다.

## `inline` 속성의 효과

`inline` 속성은 다음과 같이 작용합니다:

- `#[inline]` 또는 `#[inline(hint)]`: 컴파일러에게 인라인화를 권장하지만, 최종 결정은 컴파일러가 내립니다.
- `#[inline(always)]`: 컴파일러에게 가능한 한 항상 해당 함수를 인라인화하도록 지시합니다.
- `#[inline(never)]`: 컴파일러에게 해당 함수를 절대 인라인화하지 말도록 지시합니다.

예를 들어:

```rust
#[inline(always)]
fn always_inlined() {
    // 이 함수는 가능한 한 항상 호출 지점에 인라인화됩니다.
}

#[inline(never)]
fn never_inlined() {
    // 이 함수는 절대 인라인화되지 않습니다.
}
```

## 주의사항

- `inline` 속성은 힌트일 뿐이며, 컴파일러가 항상 이를 따르는 것은 아닙니다. 특히 `#[inline(always)]`도 특정 상황에서는 무시될 수 있습니다.
- 과도한 인라인화는 코드 크기 증가와 캐시 성능 저하를 일으킬 수 있으므로 신중하게 사용해야 합니다.
- 최신 컴파일러는 대부분의 경우 자동으로 적절한 인라인화 결정을 내리므로, `inline` 속성은 필요한 경우에만 사용하는 것이 좋습니다.

## 결론

`inline` 속성은 Rust 컴파일러의 여러 단계를 거쳐 처리되며, 최종적으로 LLVM 최적화 단계에서 실제 인라인화 결정에 영향을 미칩니다. 이는 성능 최적화를 위한 강력한 도구이지만, 그 효과는 복잡한 요인들에 의해 결정되므로 신중하게 사용해야 합니다.
