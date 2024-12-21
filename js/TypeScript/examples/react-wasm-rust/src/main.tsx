import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
// Top-level await is not available in the configured target environment ("chrome87", "edge88", "es2020", "firefox78", "safari14" + 2 overrides)
// await init()
// 대신 Provider 컴포넌트로 초기화하고 감싸서 하위 컴포넌트들에서 사용할 수 있도록 만듭니다.
import { WasmProvider } from "./WasmProvider";

createRoot(document.getElementById("root")!).render(
  // `App` 및 `App`의 모든 자식 컴포넌트는 `WasmContext.Provider`가 제공하는 값을 사용할 수 있습니다.
  <StrictMode>
    <WasmProvider>
      <App />
    </WasmProvider>
  </StrictMode>
);
