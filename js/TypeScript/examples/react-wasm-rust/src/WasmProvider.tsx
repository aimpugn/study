import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  PropsWithChildren,
} from "react";
// 브라우저나 JavaScript 런타임에서는 WebAssembly 파일을 직접 ES 모듈로 다루는 것을 지원하지 않습니다.
// Rust의 `wasm-bindgen`은 JavaScript 글루 코드(.js 파일)를 생성하여 이를 중간에서 연결해주는 역할을 합니다.
import init, { InitOutput } from "../backend/pkg/backend.js";

export enum WasmInitStatus {
  READY = 1,
  LOADING,
  ERROR,
}

// WASM 상태와 WASM 모듈 인스턴스를 관리할 타입입니다.
// `createContext` 통해서 `React.Context<WasmContextType>` 타입의 인스턴스가 생성됩니다.
interface WasmContextType {
  wasm: InitOutput | null; // WASM 인스턴스
  status: WasmInitStatus; // 초기화 상태
}

// Context 생성합니다.
// `createContext`를 호출하면 React는 자동으로 두 가지 컴포넌트를 생성합니다:
// - `WasmContext.Provider`: 데이터를 공급하는 컴포넌트
// - `WasmContext.Consumer`: 데이터를 읽는 컴포넌트 (대부분 현재는 useContext 훅으로 대체)
export const WasmContext = createContext<WasmContextType>({
  wasm: null,
  status: WasmInitStatus.LOADING,
});

// `WasmContext.Provider` 컴포넌트를 감싸는 컴포넌트입니다.
// `WasmProvider`는 `WasmContext.Provider`를 사용하여
// `Context` 값을 트리 구조 내의 모든 하위 컴포넌트들에게 전파합니다.
export const WasmProvider: React.FC<PropsWithChildren> = ({ children }) => {
  const wasmRef = useRef<InitOutput | null>(null); // WASM 인스턴스를 저장할 ref
  const [status, setStatus] = useState<WasmInitStatus>(WasmInitStatus.LOADING);

  useEffect(() => {
    // WASM 초기화
    const initializeWasm = async () => {
      try {
        wasmRef.current = await init();
        setStatus(WasmInitStatus.READY); // 초기화 성공 시 상태 업데이트
        console.log("WASM initialized successfully.");
      } catch (error) {
        setStatus(WasmInitStatus.ERROR); // 에러 발생 시 상태 업데이트
        console.error("Failed to initialize WASM:", error);
      }
    };

    initializeWasm();
  }, []);

  return (
    // `Provider`는 React 트리 내에서 자식 컴포넌트들에게 데이터를 전파하는 역할을 합니다.
    // 여기서 `value` 속성은 `Context`를 통해 공급될 데이터를 정의합니다.
    // 트리 구조의 모든 자식 컴포넌트는 이 `value`에 전달된 값을 사용할 수 있습니다.
    <WasmContext.Provider value={{ wasm: wasmRef.current, status }}>
      {children}
    </WasmContext.Provider>
  );
};

// `WasmContext`에 저장된 값을 쉽게 가져오기 위한 Custom Hook입니다.
// `useWasm`을 호출하는 컴포넌트는 React의 `Context` 시스템을 통해 `WasmContext.Provider.value` 값을 읽어옵니다.
// `WasmContext`에서 제공되는 값, 즉 `WasmProvider.value` 속성으로 전달된 데이터를 읽는 역할을 합니다.
export const useWasm = () => {
  // `WasmContext`는 애플리케이션 내에서 유일하지만, 이를 사용하는 `Provider`는 여러 개 있을 수 있습니다.
  // `useContext`를 호출하면:
  // 1. React는 현재 컴포넌트 트리를 탐색하여 가장 가까운 `WasmContext.Provider`를 찾습니다.
  // 2. 해당 `Provider.value` 값을 반환합니다.
  // 3. 만약 트리 상단에 `WasmContext.Provider`가 없으면, `createContext` 호출 시 제공된 기본값이 사용됩니다.
  const context = useContext(WasmContext);
  if (!context) {
    throw new Error("useWasm must be used within a WasmProvider");
  }
  return context;
};
