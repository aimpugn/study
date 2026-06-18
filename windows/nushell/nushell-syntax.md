# nushell 문법 노트 (이 환경에서 쓰인 것 + Windows 경로)

oh-my-posh 연동 스크립트(`.oh-my-posh.nu`, `config.nu`)에 등장한 문법 정리.

## 외부 명령 호출 `^`

```nu
^$_omp_executable print primary
```

`^`(caret)는 뒤 문자열을 **외부 실행파일로 실행**하라는 표시. `$_omp_executable`는 exe 경로 문자열인데, `^`가 없으면 nushell이 데이터(문자열)로 취급한다. 변수에 담긴 경로를 명령으로 돌릴 때 필수.

## 플래그 전달 두 형태

```nu
--config $env.POSH_THEME              # 두 토큰: 플래그 + 값
$"--status=($env.LAST_EXIT_CODE)"     # 한 토큰: 문자열 보간으로 --status=0
```

- `--config $env.POSH_THEME` — `$env.POSH_THEME` 값(문자열)이 그대로 다음 인자가 됨.
- `$"...($x)..."` — **문자열 보간**. `$"..."` 안의 `(...)`는 subexpression으로 평가돼 삽입.
- 외부 명령엔 `--k v`(공백)와 `--k=v`(등호) 둘 다 유효.

## spread `...$args`

```nu
def --wrapped f [...args] { ^cmd ...$args }
```

리스트를 개별 인자로 펼친다. rest 인자를 외부 명령에 그대로 포워딩할 때.

## `def --wrapped`

```nu
def --wrapped _omp_get_prompt [type: string, ...args: string] { ... }
```

`--wrapped`는 정의에 없는 임의 플래그까지 받아 `...args`로 모은다. oh-my-posh 같은 외부 명령 래퍼에 필요.

## 클로저 `{|| ... }`

```nu
$env.PROMPT_COMMAND = {|| _omp_get_prompt primary }
```

인자 없는 클로저 리터럴. nushell이 매 프롬프트마다 호출하고 **마지막 식의 값**을 프롬프트로 쓴다.

## env 접근: `$env.X` vs `$env.X?`

- `$env.X` — 없으면 **에러** (`Cannot find column 'X'`).
- `$env.X?` — 없으면 **null** (안전).
- 함정: 프롬프트 클로저가 `$env.CMD_DURATION_MS`(? 없이)를 읽으면, 그 변수가 없는 `nu -c` 맥락에선 에러. 실제 인터랙티브에선 nushell이 항상 세팅하므로 무해 → **테스트 아티팩트**일 수 있으니 주의.

## `const` + `$nu.current-exe` (self-relative)

```nu
const TOOLS_DIR = ($nu.current-exe | path dirname | path dirname)
```

`$nu.current-exe`는 실행 중인 nu.exe 경로(여기선 `<tools>/bin/nu.exe`) → 2단계 상위 = `<tools>`.
**폴더를 옮겨도 따라온다.** `$nu.current-exe`는 const 안에서도 평가 가능(이미 `NU_PLUGIN_DIRS`가 그 패턴을 씀).

## `source`

```nu
source ([$ASSETS_DIR "configs/.oh-my-posh.nu"] | path join)
```

파일을 **현재 스코프로 로드**한다. 그 안의 `def`·`$env` 할당이 세션에 반영됨. config.nu가 `.oh-my-posh.nu`를 이렇게 끌어온다.

## vendor/autoload

- `<config>/vendor/autoload/*.nu`는 인터랙티브 시작 시 자동 source. `nu -c`에선 안 됨.
- ⚠️ 최신 nushell은 **autoload/source 스크립트 최상위의 `return`을 금지**한다: `Return used outside of custom command or closure`. oh-my-posh **공식 init이 최상위 `return`을 써서** 이 nushell과 충돌(사례 2 참조).

---

## Windows 경로 구분자 `/` vs `\`

- Windows API는 `/`·`\` 둘 다 경로 구분자로 받는다 → 혼용도 **대개 무해**(그래서 oh-my-posh는 혼용 경로로도 동작했음). 단 보기 싫고, 일부 엄격한 도구는 실패할 수 있다.

### 혼용이 생기는 전형

```nu
['C:/x/tools' 'a/b/c'] | path join
# => C:/x/tools\a/b/c   ← 포워드(문자열) + 백슬래시(path join 삽입) 혼용
```

- `path join`은 컴포넌트 사이에 **OS 네이티브 구분자(`\`)**를 끼운다.
- 컴포넌트 안에 이미 `/`가 있으면 그건 보존됨 → 혼용 발생.

### 정규화 방법

```nu
... | path join | str replace --all '\' '/'   # 포워드슬래시로 통일
... | path join | path expand                  # 네이티브(백슬래시)로 통일(+ ~,.. 해소)
```

- `str replace`는 기본 **리터럴**(정규식 아님). `'\'`는 단일 백슬래시.
- `path expand`는 절대경로화 + `~`/`..`/심볼릭링크까지 해소하므로 더 강한 정규화.

### 이 repo 적용

`.oh-my-posh.nu`의 `POSH_THEME` 생성에 `| str replace --all '\' '/'`를 붙여 포워드슬래시로 통일(TOOLS_DIR이 포워드슬래시 컨벤션이라 그에 맞춤).

```nu
$env.POSH_THEME = ([$env.WINDOW_ENV_CONFIG_TOOLS_DIR "assets/oh-my-posh/themes/peru.omp.json"]
    | path join | str replace --all '\' '/')
# => C:/Users/rody/WorkspacePrivate/window-env-config/tools/assets/oh-my-posh/themes/peru.omp.json
```

> 더 근본적으로는 "path join 결과는 항상 한 번 정규화한다"를 규칙으로 삼거나, 헬퍼를 두면 됨:
> `def joinpath [...p] { $p | path join | str replace --all '\' '/' }`
