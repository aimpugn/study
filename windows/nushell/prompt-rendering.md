# 셸·터미널 프롬프트 렌더링 원리 (nushell + oh-my-posh)

## 1. 터미널 vs 셸 vs 프롬프트 엔진

세 계층을 구분해야 문제를 어디서 찾을지 안다.

- **터미널 에뮬레이터** (Windows Terminal): 화면에 글자를 그리고 키 입력을 셸에 전달하는 "표시 장치". 프롬프트 내용은 모른다 — 셸이 출력한 ANSI 텍스트를 그대로 렌더할 뿐.
- **셸** (nushell / PowerShell / bash): 명령을 읽기 전에 **프롬프트 문자열을 stdout에 출력**한다. 이 문자열을 무엇으로 채울지가 "프롬프트 설정".
- **프롬프트 엔진** (oh-my-posh): 셸에 독립적인 외부 프로그램. 테마(JSON) + 런타임 컨텍스트(exit code, 실행시간, 터미널 폭, cwd, git 상태…)를 받아 **완성된 ANSI 프롬프트 문자열을 출력**한다.

> 핵심: oh-my-posh는 셸 "안에서 도는" 게 아니라, 셸이 **매 프롬프트마다 호출하는 서브프로세스**다. 셸은 그 출력(stdout)을 받아 프롬프트로 쓴다.

## 2. 셸별 프롬프트 훅

| 셸 | 프롬프트 지정 방식 |
|----|------------------|
| bash | `PS1` (문자열 + 이스케이프) |
| zsh | `PROMPT` / `precmd` |
| PowerShell | `function prompt { ... }` |
| nushell | `$env.PROMPT_COMMAND` 등 (아래) |

oh-my-posh가 "cross-shell"인 이유: 각 셸의 훅에 맞는 init 스크립트를 생성해주고, 실제 렌더는 공통 바이너리가 담당하기 때문.

## 3. nushell 프롬프트 변수

nushell은 프롬프트를 그릴 때마다 다음 env 값을 평가한다.

| 변수 | 역할 |
|------|------|
| `$env.PROMPT_COMMAND` | 왼쪽 프롬프트. **문자열 또는 클로저**. 클로저면 매번 실행해 반환값 사용 |
| `$env.PROMPT_COMMAND_RIGHT` | 오른쪽 정렬 프롬프트 |
| `$env.PROMPT_INDICATOR` | 프롬프트 뒤 입력 표시 (예: `❯ `) |
| `$env.PROMPT_MULTILINE_INDICATOR` | 여러 줄 입력 시 이어지는 줄 표시 |

→ nushell에서 "테마"란 이 변수들에 **렌더 결과를 만드는 코드/문자열을 꽂는 것**이다.

## 4. nushell + oh-my-posh 통합 사슬

```
WT가 nu.exe 실행
  └ nu.exe가 config.nu 로드
      └ config.nu가 .oh-my-posh.nu 를 source
          └ $env.PROMPT_COMMAND = {|| ^oh-my-posh.exe print primary --config <theme> ... }
프롬프트 그릴 때마다:
  nushell → 클로저 실행 → oh-my-posh.exe 호출 → ANSI 문자열 → 화면
```

설정 파일을 만드는 두 방식:

- **공식 init**: `oh-my-posh init nu --config <theme>` → 위 클로저들을 세팅하는 nu 스크립트를 **생성**. source 하거나 `vendor/autoload`에 둠. 경로를 **정적으로 박는다**.
- **수작업 등가물** (이 repo의 `assets/configs/.oh-my-posh.nu`): 같은 일을 손으로 작성. 경로를 환경변수(`$env.WINDOW_ENV_CONFIG_TOOLS_DIR` 등) 기반으로 **동적**으로 잡아 이동에 강하고, nushell 호환 문법(if/else)을 쓴다.

### vendor/autoload

nushell은 시작 시 `$nu.vendor-autoload-dirs`(기본: `<config>/vendor/autoload`) 안의 `*.nu`를 **자동 source**한다.

- **인터랙티브 시작에서만** 로드된다. `nu -c "..."`에서는 로드되지 않는다 → 진단할 때 이 차이가 중요(인터랙티브에서만 나는 에러).

## 5. oh-my-posh 테마 결정 규칙 (가장 큰 함정)

`oh-my-posh print`가 "어느 테마인지" 아는 경로(우선순위):

1. **`--config <파일>`** (명시) — 항상 동작
2. **세션 캐시** — `oh-my-posh init … --config X`가 미리 채움 (`--save-cache` + `POSH_SESSION_ID`)
3. **내장 기본 테마**

> ⚠️ **이 버전(29.10.0)의 `print`는 `POSH_THEME` 환경변수를 읽지 않는다.**
> `POSH_THEME`만 세팅하고 `--config`도 캐시도 없으면 → 3번(기본 테마)으로 떨어진다.
>
> 검증(A/B): `print … --config <peru>` → peru / `print …`(POSH_THEME만) → 기본 테마 == 아무것도 안 줬을 때.

따라서 수작업 `.oh-my-posh.nu`에서는 print 호출에 반드시 `--config $env.POSH_THEME`를 넘겨야 한다.

```nu
^$_omp_executable print $type
    --config $env.POSH_THEME     # 이게 없으면 기본 테마가 뜬다
    --save-cache
    --shell=nu
    ...
```
