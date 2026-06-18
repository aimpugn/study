# nushell 활용법 · 문법 레퍼런스

일반적인 nushell 사용/문법 정리. 프롬프트·oh-my-posh 특화 내용은 [prompt-rendering.md](prompt-rendering.md)·[troubleshooting-oh-my-posh.md](troubleshooting-oh-my-posh.md), 설정에 실제 쓰인 구문은 [nushell-syntax.md](nushell-syntax.md) 참고.

## 1. 핵심 철학 — 모든 게 구조적 데이터

bash는 텍스트(문자열) 스트림, nushell은 **구조적 데이터(table / record / list) 스트림**. 명령 출력이 표라서 컬럼으로 다룬다.

```nu
ls | where size > 1mb | sort-by modified | select name size
ps | where cpu > 10 | get name
```

## 2. 변수 / 상수

```nu
let x = 42              # 불변
mut y = 0; $y = $y + 1  # 가변
const C = "..."         # 파스타임 상수 — use/source/플러그인 경로에 쓸 수 있음
```

## 3. 파이프라인 & 데이터 명령

| 목적 | 명령 |
|------|------|
| 필터 | `where <조건>`, `filter {\|x\| ... }` |
| 컬럼 | 선택 `select a b`, 추출 `get a`, 안전추출 `get a?` |
| 변환 | `each {\|x\| ...}`, `update col {...}`, `insert col {...}` |
| 정렬/제한 | `sort-by col`, `reverse`, `first n`, `last n`, `skip n`, `take n` |
| 집계 | `length`, `math sum`, `group-by`, `uniq`, `reduce {\|it,acc\| ...}` |
| 펼침/모음 | `flatten`, `compact`(null 제거), `append`, `prepend` |

```nu
[1 2 3 4] | where $it mod 2 == 0      # [2 4]
$env | columns | where $it =~ POSH    # POSH* env 키 목록
```

## 4. 문자열 / 경로

```nu
"a,b,c" | split row ","                # [a b c]
"Hello" | str downcase | str contains "ell"
$"값은 ($x) 입니다"                     # 문자열 보간
[$dir sub file.txt] | path join        # 경로 결합 (네이티브 구분자 \ 삽입)
$path | path dirname                    # 상위, basename/parse 도 있음
$path | path expand                     # 절대화 + 정규화(~, .. 해소)
$path | path exists
```

> Windows `/` vs `\` 혼용은 `path join`이 `\`를 끼우기 때문. `str replace --all '\' '/'`(포워드 통일) 또는 `path expand`(네이티브 통일)로 정규화 — [nushell-syntax.md](nushell-syntax.md) 마지막 절.

## 5. 환경변수 `$env`

```nu
$env.FOO          # 접근 (없으면 에러)
$env.FOO?         # 없으면 null
$env.FOO = "bar"  # 설정 (스코프 한정)
$env | columns    # 모든 키
```

- ⚠️ `$env` 키는 **대소문자 구분**(Windows PATH는 보통 `$env.Path`).
- 블록/클로저 안의 `$env` 변경을 밖으로 내보내려면 `def --env` / `do --env`.

## 6. 커스텀 명령 `def`

```nu
def greet [name: string, --loud] {
    let m = $"hi ($name)"
    if $loud { $m | str upcase } else { $m }
}
def --wrapped run [...args] { ^some-tool ...args }  # 미정의 플래그까지 받아 포워딩
def --env cd-up [] { cd .. }                         # $env 변경을 호출자에 반영
```

## 7. 클로저 `{|args| ... }`

```nu
let f = {|x| $x * 2 }
[1 2 3] | each $f                                    # [2 4 6]
$env.PROMPT_COMMAND = {|| date now | format date "%H:%M" }
```

## 8. 모듈 / 스크립트 로드

```nu
use mymod.nu        # 'mymod sub-cmd' 형태
use mymod.nu *      # 모든 export를 현재 스코프로
source script.nu    # 파일을 현재 스코프에서 실행 (def/$env 반영)
overlay use ovl     # 켜고 끌 수 있는 오버레이
```

- `vendor/autoload/*.nu` 는 **인터랙티브 시작 시 자동 로드**(`nu -c`엔 안 됨).
- ⚠️ source/autoload 스크립트 **최상위 `return` 금지**(커스텀 명령/클로저 안에서만 가능).

## 9. 외부 명령 연동

```nu
^git status                          # ^ = 외부 명령 강제 실행
^$exe_path arg1 arg2                  # 변수에 든 경로를 명령으로
^cmd | complete                      # {stdout, stderr, exit_code} 캡처
git log --oneline | lines | first 5  # 외부 텍스트 출력 → 구조화
$"--flag=($val)"                     # 값 있는 플래그(한 토큰), 또는 `--flag $val`(두 토큰)
```

## 10. config 구조 & `$nu`

| 파일/변수 | 역할 |
|----------|------|
| `$nu.env-path` (env.nu) | 부팅 초기 env |
| `$nu.config-path` (config.nu) | 메인 설정 (env.nu 다음) |
| `$nu.current-exe` | 실행 중인 nu.exe 경로 — self-locating에 활용 |
| `$nu.default-config-dir` | 설정 디렉터리 |
| `<config>/vendor/autoload/*.nu` | 자동 로드 |

```nu
# self-locating 예: nu.exe 가 <tools>/bin/nu.exe 이면
const TOOLS_DIR = ($nu.current-exe | path dirname | path dirname)
```

## 11. 흔한 함정

- `$env.X`는 없으면 에러 → 안전하게 `$env.X?` 또는 `... | default <v>`.
- source/autoload 최상위 `return` 금지.
- `const`는 const-evaluable 식만(대부분 `path`/`str` 명령 OK, 사용자 `def`는 불가).
- `path join`은 네이티브 `\`를 끼움 → 포워드슬래시 베이스와 혼용. 필요시 정규화.
- 외부 명령은 `^`로 호출. 값 플래그는 `--k=v`(한 토큰) 또는 `--k v`(두 토큰).
- nushell 버전 간 API 변동이 잦다 — config 헤더 버전과 실제 nu.exe 버전이 다르면 비호환 가능(예: 최상위 `return` 금지 시점).
