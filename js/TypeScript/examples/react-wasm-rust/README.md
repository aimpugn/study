# TypeScript, Vite, React, WASM, 그리고 Rust

TypeScript, Vite, React, WASM, 그리고 Rust이 어떻게 동작하는지 간단히 테스트하는 예제입니다.

- Install pnpm, wasm-pack

    ```shell
    brew install pnpm wasm-pack
    ```

- vite with react-ts

    ```shell
    pnpm create vite@latest react-wasm-rust --template react-ts
    cd react-wasm-rust && pnpm install
    ```

    <details>
    <summary>`vite-plugin-wasm-pack`은 사용하지 않습니다. 만약 직접 `.wasm` 파일을 로드하려면 `vite-plugin-wasm` 추가합니다.</summary>

    ```sh
    # 최근 브라우저만 대상으로 하는 게 아니면 `vite-plugin-top-level-await`도 필요합니다.
    pnpm add vite-plugin-wasm
    ```

    > `vite-plugin-wasm-pack` 아닌 `vite-plugin-wasm`를 사용합니다:
    >
    > vite:4.5.5까지 동작하는 것을 확인했지만, vite 5 버전부터
    > 내부적으로 사용하는 express의 `request.url` 속성이 [`request.originalUrl` 속성](https://expressjs.com/en/api.html#req.originalUrl)으로 변경되었고
    > 이로 인해 [경로 이름 기반으로 `Content-Type: application/wasm`으로 응답](https://github.com/nshen/vite-plugin-wasm-pack/blob/5e626b9d387b9e9df87712479df2eb5110af02f7/src/index.ts#L161-L179)하도록 한 코드가
    > 정상적으로 동작하지 않습니다.
    >
    > 이에 대해 [수정 PR](https://github.com/nshen/vite-plugin-wasm-pack/pull/36)이 올라왔지만
    > 머지되지 않는 것으로 보아 앞으로도 유지보수가 안 될 것으로 보입니다.

    그리고 `vite.config.ts` 파일을 다음과 같이 수정합니다.

    ```ts
    import { defineConfig } from 'vite';
    import react from '@vitejs/plugin-react';
    import wasmPack from 'vite-plugin-wasm';

    // https://vite.dev/config/
    export default defineConfig({
        plugins: [
            react(),
            wasm(),
        ],
    })
    ```

    </details>

- Rust 코드 생성

    ```shell
    cargo init --lib
    ```

    그리고 `Cargo.toml`을 다음과 같이 설정합니다.

    ```toml
    [package]
    name = "backend"
    version = "0.1.0"
    edition = "2021"

    [dependencies]
    wasm-bindgen = "0.2.99"

    # The `console_error_panic_hook` crate provides better debugging of panics by
    # logging them with `console.error`. This is great for development, but requires
    # all the `std::fmt` and `std::panicking` infrastructure, so isn't great for
    # code size when deploying.
    console_error_panic_hook = { version = "0.1.7", optional = true }

    [lib]
    crate-type = ["cdylib", "rlib"]

    [profile.release]
    # Tell `rustc` to optimize for small code size.
    opt-level = "s"
    ```

    그리고 `wasm_bindgen` 어노테이션을 붙여서 JavaScript에서 사용할 코드를 작성합니다.

    ```rs
    use wasm_bindgen::prelude::*;

    // Rust에서 WebAssembly로 export할 함수 정의
    #[wasm_bindgen]
    pub fn add_one(input: i32) -> i32 {
        input + 1
    }
    ```

- `package.json`에 다음 명령어를 추가합니다.

    ```json
    {
        "scripts": {
            "wasm": "wasm-pack build ./backend --target web"
        }
    }
    ```

    그리고 해당 명령어를 실행하여 `wasm-bingen` 통해 글루 스크립트가 [backend/pkg](./backend/pkg/)에 생성되도록 만듭니다.

    ```sh
    pnpm wasm
    ```

- [`WasmProvider`](./src/WasmProvider.tsx) 통해 `WASM` 인스턴스를 초기화하고
    최상위 컴포넌트에 추가하여 `Context`를 제공하도록 합니다.

    ```ts
    createRoot(document.getElementById("root")!).render(
        <StrictMode>
            <WasmProvider>
                <App />
            </WasmProvider>
        </StrictMode>
    );
    ```

    [main.tsx](./src/main.tsx)를 위와 같이 수정하면, `App` 및 그 하위 컴포넌트에서
    `init`([`__wbg_init`](./backend/pkg/backend.js))된 `WebAssemblyInstantiatedSource` 인스턴스를 사용할 수 있게 됩니다.
