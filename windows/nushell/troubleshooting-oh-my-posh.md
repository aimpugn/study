# 트러블슈팅: oh-my-posh 테마가 안 먹을 때 (실제 사례 3건)

window-env-config 폴더를 여러 번 옮기며(`Workspace` → `C:\dev\Workspace` → `WorkspacePrivate`) 겪은 고장들.
증상은 매번 "테마가 안 뜸"으로 같았지만 **원인은 3가지로 달랐다.** 같은 증상에 원인이 여러 개라는 게 핵심 교훈.

## 진단 원칙 — 레이어 격리

"테마 안 뜸"을 한 덩어리로 보지 말고 사슬을 층별로 끊어 본다.

1. **nushell이 뜨나?** 안 뜨면 WT→nu.exe 문제 (사례 1).
2. **config.nu가 로드되나? 시작 에러 없나?** vendor autoload 에러 등 (사례 2).
3. **프롬프트 env가 세팅되나?** `$env.PROMPT_COMMAND`, `$env.POSH_THEME` 확인.
4. **그 env가 가리키는 파일/exe가 실재하나?** `path exists`.
5. **프롬프트 클로저가 부르는 외부 명령을 손으로 그대로 실행** → A/B 비교로 인자 문제를 잡는다 (사례 3).

> 5번이 결정적. "env는 맞는데 왜 안 되지?"에서 멈추지 말고, 그 env를 쓰는 **실제 외부 호출을 재현**하라.

유용한 명령:
```nu
view source $env.PROMPT_COMMAND     # 프롬프트가 무엇을 호출하는지
view source _omp_get_prompt         # 실제 oh-my-posh 호출 라인
$env.POSH_THEME                     # 테마 경로 확인
$env.POSH_THEME | path exists       # 실재 확인
```

---

## 사례 1 — WT가 죽은 nu.exe 경로를 가리킴

- **증상:** WT 열면 nushell이 제대로 안 뜸 / 테마 없음.
- **원인:** 폴더 이동 후 Windows Terminal `settings.json`의 nu 프로필 `commandline`이 **사라진 절대경로**(`C:/dev/Workspace/.../nu.exe`)를 가리킴. config.nu는 새 위치로 갱신됐는데 WT 설정만 옛 경로로 남음.
- **원리:** WT→nu.exe는 **상대화 불가능한 부트스트랩 절대경로**. WT가 바로 그 스크립트를 *시작*시키므로 "현재 스크립트 기준 상대경로"가 성립하지 않는다(시작 전엔 기준 스크립트가 없음).
- **고침:** `settings.json`의 nu 프로필 `commandline`을 현재 nu.exe 경로로 갱신.
- **재발 방지:** 이동 후 `setup` 재실행(install.nu의 WT 단계가 갱신), 또는 고정 junction(`%APPDATA%\nu-current`)으로 우회. config 쪽은 self-relative로 만들 수 있지만 **이 한 링크만은 절대경로**다.

## 사례 2 — vendor/autoload 의 최상위 `return`

- **증상:** nushell은 뜨는데 시작 직후
  `Error: Return used outside of custom command or closure` (`vendor/autoload/oh-my-posh.nu:11`).
- **원인:** 과거 `oh-my-posh init nu`로 생성된 **공식 init**이 `vendor/autoload`에 남아 있었음. (1) 최상위 `return`을 씀 — 최신 nushell이 금지, (2) 죽은 Workspace exe 경로 하드코딩. 게다가 config.nu의 커스텀 설정과 **중복**.
- **고침:** stale 파일을 autoload에서 제거(백업으로 이동).
- **교훈:**
  - oh-my-posh 공식 init은 최상위 `return` 때문에 현 nushell과 **근본적으로 안 맞는다.** 커스텀 `.oh-my-posh.nu`(if/else 구조)가 호환 버전. **둘을 섞지 말 것.**
  - 이 에러는 `nu -c`에선 안 난다(vendor autoload 미로드). **인터랙티브에서만** 재현됨.
  - "이전엔 됐는데 지금 깨짐" = 환경 변화(nushell 업그레이드로 `return` 금지 + 폴더 이동으로 경로 사망)가 stale 아티팩트를 터뜨린 것.

## 사례 3 — POSH_THEME 무시 → 기본 테마 (가장 헷갈림)

- **증상:** nushell·exe·테마파일·`PROMPT_COMMAND` 전부 정상인데 **peru가 아니라 기본 테마**가 뜸.
- **원인:** 커스텀 `.oh-my-posh.nu`가 `$env.POSH_THEME`만 세팅하고 `oh-my-posh print`에 `--config`를 안 넘김. **이 버전 oh-my-posh는 print에서 `POSH_THEME`를 안 읽음** → 내장 기본 테마.
- **진단(A/B):**
  ```nu
  ^…/oh-my-posh.exe print primary --shell nu                          # 기본 테마
  ^…/oh-my-posh.exe print primary --shell nu --config $env.POSH_THEME # peru
  ```
  → 차이는 오직 `--config` 유무. POSH_THEME 단독으론 안 먹는다는 게 드러남.
- **고침:** `_omp_get_prompt`의 print와 멀티라인 인디케이터 print에 `--config $env.POSH_THEME` 추가.
- **교훈:** **"값이 세팅됨" ≠ "그 값이 실제로 쓰임".** env가 맞아도 외부 명령이 그걸 읽는지 확인해야 한다. (테마 결정 규칙은 [prompt-rendering.md](prompt-rendering.md) §5)

---

## 이동 내성(move-resilience) 설계

폴더를 옮겨도 안 깨지게 하려면 두 가지를 분리해 생각한다.

- **config 쪽은 상대화 가능:**
  ```nu
  const TOOLS_DIR = ($nu.current-exe | path dirname | path dirname)
  ```
  → 테마·assets·exe 경로가 nu.exe 위치를 자동으로 따라감. `POSH_THEME`도 TOOLS_DIR 기준이라 함께 따라감.
- **WT→nu.exe만 절대경로:** 이동 시 한 곳 갱신이 필요. `setup` 자동화 또는 junction으로 무인화.

즉 "이동 내성"의 핵심 = **config 상대화 + 그 한 절대 링크 처리.**

## 경로 구분자 혼용 (부수 이슈)

`$env.POSH_THEME`가 `C:/…/tools\assets/…`처럼 섞이는 건 `포워드슬래시 문자열 + path join(네이티브 \)` 탓.
기능상 무해하나 보기 싫으면 `| str replace --all '\' '/'`로 통일(자세히는 [nushell-syntax.md](nushell-syntax.md) 마지막 절).

## 빠른 체크리스트

```nu
$nu.current-exe                       # WT가 띄운 nu.exe — 기대 위치인가?
$env.WINDOW_ENV_CONFIG_TOOLS_DIR      # config의 TOOLS_DIR — 현재 위치 맞나?
$env.POSH_THEME | path exists         # 테마 파일 실재?
($env.WINDOW_ENV_CONFIG_BIN_DIR | path join 'oh-my-posh.exe') | path exists  # exe 실재?
view source $env.PROMPT_COMMAND       # print 호출에 --config 있나?
ls ($nu.default-config-dir | path join vendor autoload)  # 충돌하는 공식 init 없나?
```
