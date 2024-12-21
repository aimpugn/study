import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
// `wasm-pack build`로 생성된 글루 스크립틀르 사용하면 `fetch`로 WASM 파일을 동적으로 로드하기 때문에
// `vite-plugin-wasm`가 없어도 동작하지만, WASM 모듈을 직접 가져오는 경우, import 문으로 간편하게 가져와 사용할 수 있도록 지원합니다.
// 그리고 `wasm-pack`으로 생성된 모듈을 지원한다고 합니다.(근데 예제가 없어서 어떻게 지원한다는 건지...)
import wasm from "vite-plugin-wasm";

// https://vite.dev/config/
export default defineConfig({
  plugins: [wasm(), react()],
  //   plugins: [react()],
  optimizeDeps: {
    // `optimizeDeps.exclude`는 [사전 번들링(pre-bundling)](https://ko.vite.dev/guide/dep-pre-bundling)에서 제외할 디펜던시를 명시하는 옵션입니다.
    // 이러한 번들링 과정을 거치지 않고, `wasm-bindgen`이 생성한 글루 코드를 그대로 사용하도록 합니다.
    exclude: ["backend"],
  },
  build: {
    // 모든 에셋에 대해 Base64 인코딩을 비활성화합니다.
    // 만약 활성화될 경우, 크기가 작은 `.wasm` 파일 자체가 base64로 인코딩됩니다.
    // ex:
    // data:application/wasm;base64,AGFzbQEAAAABCQJgAABgAX8BfwInAQN3YmcfX193YmluZGdlbl9pbml0X2V4dGVybnJlZl90YWJsZQAAAwIBAQQFAW8AgAEFAwEAEQc9BAZtZW1vcnkCAAdhZGRfb25lAAETX193YmluZGdlbl9leHBvcnRfMAEAEF9fd2JpbmRnZW5fc3RhcnQAAAoJAQcAIABBAWoLC/UGAQBBiIDAAAvrBgEAAAABAAAAY2Fubm90IGFjY2VzcyBhIFRocmVhZCBMb2NhbCBTdG9yYWdlIHZhbHVlIGR1cmluZyBvciBhZnRlciBkZXN0cnVjdGlvbi9Vc2Vycy9yb2R5Ly5ydXN0dXAvdG9vbGNoYWlucy9zdGFibGUtYWFyY2g2NC1hcHBsZS1kYXJ3aW4vbGliL3J1c3RsaWIvc3JjL3J1c3QvbGlicmFyeS9zdGQvc3JjL3RocmVhZC9sb2NhbC5ycwAAAFYAEABvAAAABAEAABoAAAACAAAABQAAAAwAAAAEAAAABgAAAAcAAAAIAAAAL3J1c3QvZGVwcy9kbG1hbGxvYy0wLjIuNi9zcmMvZGxtYWxsb2MucnNhc3NlcnRpb24gZmFpbGVkOiBwc2l6ZSA+PSBzaXplICsgbWluX292ZXJoZWFkAPQAEAApAAAAqAQAAAkAAABhc3NlcnRpb24gZmFpbGVkOiBwc2l6ZSA8PSBzaXplICsgbWF4X292ZXJoZWFkAAD0ABAAKQAAAK4EAAANAAAAQWNjZXNzRXJyb3JtZW1vcnkgYWxsb2NhdGlvbiBvZiAgYnl0ZXMgZmFpbGVkAAAApwEQABUAAAC8ARAADQAAAHN0ZC9zcmMvYWxsb2MucnPcARAAEAAAAGMBAAAJAAAABQAAAAwAAAAEAAAACQAAAAAAAAAIAAAABAAAAAoAAAAAAAAACAAAAAQAAAALAAAADAAAAA0AAAAOAAAADwAAABAAAAAEAAAAEAAAABEAAAASAAAAEwAAAGNhcGFjaXR5IG92ZXJmbG93AAAAVAIQABEAAABhbGxvYy9zcmMvcmF3X3ZlYy5yc3ACEAAUAAAAGAAAAAUAAAA6IAAAAQAAAAAAAACUAhAAAgAAAH0gfTAwMDEwMjAzMDQwNTA2MDcwODA5MTAxMTEyMTMxNDE1MTYxNzE4MTkyMDIxMjIyMzI0MjUyNjI3MjgyOTMwMzEzMjMzMzQzNTM2MzczODM5NDA0MTQyNDM0NDQ1NDY0NzQ4NDk1MDUxNTI1MzU0NTU1NjU3NTg1OTYwNjE2MjYzNjQ2NTY2Njc2ODY5NzA3MTcyNzM3NDc1NzY3Nzc4Nzk4MDgxODI4Mzg0ODU4Njg3ODg4OTkwOTE5MjkzOTQ5NTk2OTc5ODk5AG8JcHJvZHVjZXJzAghsYW5ndWFnZQEEUnVzdAAMcHJvY2Vzc2VkLWJ5AwVydXN0Yx0xLjgzLjAgKDkwYjM1YTYyMyAyMDI0LTExLTI2KQZ3YWxydXMGMC4yMy4zDHdhc20tYmluZGdlbgYwLjIuOTkASQ90YXJnZXRfZmVhdHVyZXMEKw9tdXRhYmxlLWdsb2JhbHMrCHNpZ24tZXh0Kw9yZWZlcmVuY2UtdHlwZXMrCm11bHRpdmFsdWU=
    assetsInlineLimit: 0,
  },
});
