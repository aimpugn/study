# Posh 테마 렌더링 검증

nushell 프롬프트가 의도한 oh-my-posh 테마(`peru`)로 그려지는지를, nushell 세션을 실제로 띄우지 않고 한 번의 비대화형 실행으로 확인하는 명령이다. `config.nu`/`env.nu`를 정상 경로로 로드시킨 뒤 프롬프트를 만드는 클로저 `$env.PROMPT_COMMAND`를 직접 호출하고, 그 출력 앞부분만 떼어 본다. 프롬프트는 직전 명령의 실행 시간(`CMD_DURATION_MS`)과 종료 코드(`LAST_EXIT_CODE`)에 따라 모양이 달라지므로, 두 값을 더미로 주입해 렌더링 조건을 고정한다. 출력에 `~\` 형태의 경로와 청록/마젠타 세그먼트가 보이면 `peru` 테마가 적용된 것이다.

PowerShell에서 실행한다.

```powershell
$nu='C:\Users\rody\WorkspacePrivate\window-env-config\tools\bin\nu.exe'
Write-Output "=== 수정된 PROMPT_COMMAND 클로저가 peru 를 그리나 (config 로드 후 실행) ===" & $nu --config 'C:\Users\rody\AppData\Roaming\nushell\config.nu' --env-config 'C:\Users\rody\AppData\Roaming\nushell\env.nu' -c '$env.CMD_DURATION_MS = "0823"; $env.LAST_EXIT_CODE = 0; do $env.PROMPT_COMMAND' 2>&1 | Select-Object -First 4
Write-Output "`n=== (대조) peru 의 특징 세그먼트가 맞는지 — 위 출력에 '~\' 경로 + 청록/마젠타가 보이면 peru ==="
```
