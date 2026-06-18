# Windows · Nushell 환경 노트

Windows + Windows Terminal + nushell + oh-my-posh 환경을 운용하며 정리한 노트.
실제로 프롬프트 테마가 안 먹는 문제를 여러 번 겪으며 얻은 원리/문법/트러블슈팅을 묶었다.

## 문서

- [nushell-usage.md](nushell-usage.md) — **일반 nushell 활용법·문법** 레퍼런스(구조적 데이터 파이프라인, 변수/상수, `def`/클로저, 모듈, 외부 명령 연동, config 구조, 흔한 함정).
- [prompt-rendering.md](prompt-rendering.md) — 터미널·셸이 프롬프트를 그리는 원리, nushell 프롬프트 변수, oh-my-posh가 외부 렌더러로 동작하는 방식, **테마 결정 규칙**.
- [nushell-syntax.md](nushell-syntax.md) — 이 환경에서 쓰인 nushell 문법(`^` 외부호출, `path join`, 문자열 보간, 클로저, `def --wrapped`, `$env.X?`, `const`+`$nu.current-exe`, `source`, vendor/autoload)과 **Windows 경로 구분자(`/` vs `\`)** 처리.
- [troubleshooting-oh-my-posh.md](troubleshooting-oh-my-posh.md) — 실제로 겪은 고장 3건(WT→nu.exe 죽은 경로 / vendor autoload 최상위 `return` / `POSH_THEME` 무시)과 **레이어 격리 진단법**, 이동 내성 설계.

## 한 줄 요약

nushell 프롬프트는 `$env.PROMPT_COMMAND`(클로저)가 외부 `oh-my-posh.exe`를 호출해 만든다.
oh-my-posh는 테마를 **`--config`(또는 init이 채운 캐시)로만** 알고 `POSH_THEME` 환경변수만으로는 모른다(이 버전 기준).
그래서 프롬프트가 안 떠도 "부품(exe·테마파일·env)"은 멀쩡할 수 있고, **실제 외부 호출의 인자**를 봐야 한다.

## 핵심 교훈

1. **"값이 세팅됨" ≠ "그 값이 실제로 쓰임"** — env가 맞아도 외부 명령이 그걸 읽는지 따로 확인.
2. **레이어 격리** — nushell 로드 → config 로드 → env 세팅 → 파일 실재 → 실제 외부 호출 재현(A/B).
3. **상대화 가능한 것과 아닌 것** — config(테마·assets)는 `$nu.current-exe`로 상대화 가능, 그러나 **WT→nu.exe는 절대경로 부트스트랩**이라 이동 시 갱신 필요.
