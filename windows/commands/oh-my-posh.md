# oh-my-posh

## `oh-my-posh init nu --config 테마파일경로`

가능합니다. **현재 Oh My Posh 공식 문서 기준으로는 Nushell 설정 파일인 `$nu.config-path`의 마지막 줄에 `oh-my-posh init nu --config <테마파일경로>`를 넣으면 됩니다.** Nushell은 Oh My Posh에서 `v0.104.0` 이상을 요구하고, custom theme은 `--config`에 로컬 JSON 경로를 지정하는 방식입니다.

Nushell에서 먼저 설정 파일 위치를 확인합니다.

```nu
# Nushell 설정 파일의 실제 경로를 확인합니다.
# Windows에서는 보통 C:\Users\<사용자>\AppData\Roaming\nushell\config.nu 계열입니다.
$nu.config-path

# 메모장으로 엽니다.
notepad $nu.config-path
```

그리고 `config.nu` 맨 아래에 아래처럼 추가하면 됩니다. Windows 경로는 Nushell에서도 `/`를 쓰는 편이 덜 헷갈립니다.

```nu
# Oh My Posh가 Nushell용 prompt hook을 구성합니다.
# --config는 사용할 테마 JSON 파일을 지정합니다.
# 경로에 공백이 있을 수 있으므로 문자열로 감쌉니다.
oh-my-posh init nu --config 'C:/Users/rody/path/to/peru.omp.json'
```

예를 들어 `peru.omp.json`이 Oh My Posh 기본 테마 폴더에 있다면 보통 이런 형태입니다.

```nu
oh-my-posh init nu --config 'C:/Users/rody/AppData/Local/Programs/oh-my-posh/themes/peru.omp.json'
```

또는 테마 파일을 직접 관리하고 싶다면 이런 위치가 더 낫습니다.

```nu
# 예: 테마 파일을 사용자 홈 아래에 고정해 둔 경우
oh-my-posh init nu --config 'C:/Users/rody/.config/oh-my-posh/peru.omp.json'
```

이렇게 한 뒤 Nushell을 재시작하면 적용됩니다.

```nu
# 버전 확인: Oh My Posh 문서 기준 Nushell은 v0.104.0 이상이어야 합니다.
version

# Oh My Posh가 PATH에서 잡히는지 확인합니다.
which oh-my-posh

# 테마 파일이 실제로 존재하는지 확인합니다.
ls 'C:/Users/rody/.config/oh-my-posh/peru.omp.json'
```

만약 `config.nu`에 넣었는데 적용이 안 되거나, 사용하는 Oh My Posh/Nushell 조합이 예전 방식만 잘 먹는다면 생성 파일을 `source`하는 방식으로 바꿔도 됩니다. 이 방식은 `oh-my-posh init nu --print`가 Nushell용 초기화 스크립트를 만들고, `config.nu`가 그 파일을 읽는 구조입니다.

```nu
# config.nu 안에 넣는 대체 방식입니다.
# 매번 Nushell 시작 시 Oh My Posh 초기화 스크립트를 생성해 ~/.oh-my-posh.nu에 저장합니다.
# 그 다음 source가 그 스크립트를 읽어 prompt hook을 현재 Nushell 세션에 등록합니다.
oh-my-posh init nu --config 'C:/Users/rody/.config/oh-my-posh/peru.omp.json' --print | save -f ($nu.home-path | path join '.oh-my-posh.nu')
source ($nu.home-path | path join '.oh-my-posh.nu')
```

추천은 먼저 공식 문서의 현재 방식인 아래 한 줄입니다.

```nu
oh-my-posh init nu --config 'C:/Users/rody/.config/oh-my-posh/peru.omp.json'
```

실패하면 그때 `--print | save -f ...` + `source ...` 방식으로 내리면 됩니다. 핵심은 `peru.omp.json`을 **테마 이름이 아니라 실제 파일 경로**로 넘기는 것이고, Oh My Posh 공식 문서도 `--config`가 로컬 설정 파일 경로를 받을 수 있다고 설명합니다.
