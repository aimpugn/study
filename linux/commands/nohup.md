# nohup

- [nohup](#nohup)
    - [nohup?](#nohup-1)
    - [`nohup some_cmd &` 실행 시 일어나는 일](#nohup-some_cmd--실행-시-일어나는-일)
    - [세션을 아예 끊어 버린다면?](#세션을-아예-끊어-버린다면)
    - [`&`만으로는 백그라운드 프로세스 상주 불가](#만으로는-백그라운드-프로세스-상주-불가)
    - [예제](#예제)
        - [`nohup`으로 vector 프로세스 관리하기 by ChatGPT](#nohup으로-vector-프로세스-관리하기-by-chatgpt)

## nohup?

> The nohup utility invokes utility with its arguments and at this time sets the signal `SIGHUP` to be ignored.
> If the standard output is a terminal, the standard output is appended to the file `nohup.out` in the current directory.
> If standard error is a terminal, it is directed to the same place as the standard output.
>
> Some shells may provide a builtin `nohup` command which is similar or identical to this utility. Consult the builtin(1) manual page.

터미널(세션)의 종료와 무관하게 프로세스를 계속 살려 두기 위해 쉘이 제공하는 POSIX 유틸로, 터미널과 프로세스를 강하게 분리하는 방법입니다.

이름 그대로 no hang-up, 본래 모뎀 시절의 "전화가 끊김(hang-up)"을 알리기 위해 생겼던 `SIGHUP` 신호를 무시하도록 프로그램을 시작할 때 `signal(SIGHUP, SIG_IGN)`을 호출해 프로세스 내부 상태를 바꿉니다.

> 분리의 강함과 약함은 세션(session), 컨트롤 터미널(CTTY), 시그널 전파 관계를 끊는 정도가 얼마나 깊은지를 의미합니다.
>
> 다음과 같은 경우를 "약한 분리"라고 합니다.
>
> ```sh
> nohup some_cmd &
> ```
>
> 이 경우 `SIGHUP`만 무시하며 여전히 *부모 쉘과 같은 세션*에 속하며 *같은 터미널을 공유*합니다.
>
> 자식이 `setsid some_cmd` 또는 `daemon --some-options` 등을 호출해 스스로 세션 리더가 되면 컨트롤 터미널(CTTY)은 -1로 세팅되어 어떤 터미널과도 연결되지 않습니다.
> 세션 리더가 되었으므로 터미널 쪽에서 뿌리는 `SIGHUP` 대상에서 벗어나고 `SIG_IGN`으로 바꿀 필요조차 없습니다.
> 세션을 끊어 냈다는 점에서 약한 분리보다는 강하지만, 여전히 I/O 경로는 조정해 줘야 합니다.
>
> `systemd Type=forking` 서비스나 이중 포크의 경우 프로세스와 터미널 사이에 어떠한 커널 수준의 연결도 남아있지 않게 됩니다.
> 신호나 I/O 등에서 서로의 연결이 사라지므로 강하게 분리된다고 합니다.

터미널이 닫혀도 커널은 여전히 해당 PID로 `SIGHUP`을 전송하지만, 프로세스 안에서는 이미 "무시"로 등록돼 있으므로 아무 일도 일어나지 않습니다.

`nohup`은 대화형 입력을 모두 끊고 `stdin`이 `/dev/null`로 바뀌고 `stdout / stderr`가 지정 파일(기본은 `nohup.out`)로 리디렉션됩니다.

이렇게 한 번 무시로 설정된 `SIGHUP`은 *그 이후에도 무시*되지만, 애플리케이션이 직접 `signal(SIGHUP, handler)`를 다시 호출하면 우선순위가 덮어써집니다.

가령 `vector`처럼 설정 파일을 다시 읽기 위해 `SIGHUP`을 쓰는 데몬은 첫 `main()`에서 자신만의 핸들러를 다시 등록해 놓습니다.
자신만의 핸들러가 등록된 경우 `nohup ./vector … &` 형식으로 띄워도 외부에서 `kill -HUP <PID>`를 보내면 정상적으로 핸들러가 실행되어 `nohup`임에도 리로드가 일어납니다.

## `nohup some_cmd &` 실행 시 일어나는 일

로그인 쉘을 열면 해당 로그인 쉘이 세션 리더가 되고, `/dev/pts/N` 같은 터미널 디바이스가 세션 전체의 "컨트롤 터미널"이 됩니다.

`some_cmd &` 라고 입력하면 먼저 쉘은 `fork()`로 자신을 복제합니다.
그리고 자식 쪽에서 `execve()`를 호출하여 `some_cmd` 바이너리를 적재합니다.
이때 두 프로세스는 "세션(Session)"이라는 논리적 그룹으로 묶입니다.

> 세션이란 사용자가 로그인하거나 가상 터미널을 열 때마다 커널이 하나씩 만들어 주는 관할 구역 같은 것입니다.
> 해당 세션 안에서는 여러 개의 "프로세스 그룹"이 서로 협조하며 터미널을 공유할 수 있습니다.

`&` 연산자는 *세션을 분리하지 않은 채* 새로운 프로세스 그룹만 만듭니다.
쉘은 `setpgid()`를 호출해 방금 태어난 자식에게 독립된 `PGID`를 부여합니다.
이렇게 하면 나중에 `jobs`, `fg`, `bg` 같은 명령어로 해당 백그라운드 작업을 다룰 수 있습니다.

하지만 이것만으로는 세션 ID, 즉 "어느 터미널에서 태어났는지"에 해당하는 값은 변하지 않습니다.
결과적으로 `some_cmd`는 부모인 쉘과 *동일한 세션*과 *동일한 컨트롤 터미널(CTTY)* 을 공유합니다.

사용자가 터미널 창을 닫거나, 쉘이 `exit` 하거나, 네트워크 연결/GUI 탭이 닫히면, 커널은 해당 터미널의 세션 리더와 해당 터미널의 포그라운드 프로세스 그룹에만 "이제 더 이상 당신을 붙잡는 단말기는 없다"는 의미로 `SIGHUP`과 `SIGCONT`를 차례로 보냅니다.
이때 백그라운드 그룹은 자동 전파 대상이 아닙니다.

물리 직렬선에서는 모뎀이 Data-Carrier-Detect(DCD) 신호를 내려 버릴 때 드라이버가 `tty_hangup()`을 호출하는데, 이것이 하드웨어 레벨의 hang-up 이벤트입니다.
가상 PTY에서는 하드웨어 대신 *마스터 쪽 파일 디스크립터*가 마지막으로 `close()`될 때 `tty_hangup()`를 호출합니다.
POSIX `close()` 규격은 마스터가 마지막으로 닫히면 슬레이브를 `CTTY`로 쓰던 제어 프로세스에게 `SIGHUP`을 보낼 것을 요구합니다.

`SIGHUP`의 원래 의미는 "전화선이 끊겨서(hang-up) 더는 입출력이 불가능하다"는 통보이며, 아무 대비를 하지 않은 프로세스는 디폴트 동작에 따라 곧바로 종료됩니다.
이는 *백그라운드로 실행한 `some_cmd`도 마찬가지*입니다.
`PGID`가 다르더라도 같은 세션 안에 있는 이상 `CTTY`가 사라지는 순간 함께 `SIGHUP`을 받게 되고, 표준 시그널 핸들러는 곧장 프로세스를 종료시킵니다.

> 즉 `&`은 job control 편의를 위해 `PGID`만 바꿀 뿐입니다.
> 터미널과의 생살여탈권을 쥔 세션 및 `CTTY` 관계는 그대로 두기 때문에 터미널이 닫히면 그 운명을 같이하게 됩니다.

`nohup some_cmd &` 실행 시 `nohup` 명령은 자신이 `execve()`하기 직전에 두 가지를 조치합니다.
- `signal(SIGHUP, SIG_IGN)`을 호출해 `SIGHUP`을 무시하도록 프로세스 테이블의 시그널 디스패치 벡터를 수정합니다.

    => 세션과 `CTTY`는 여전히 공유하지만, 터미널이 사라져도 `SIGHUP`을 무시하므로 종료되지 않습니다.

- 표준 입력은 `/dev/null`로, 표준 출력과 표준 오류는 `nohup.out`(혹은 리다이렉션된 대상)으로 바꿉니다.

    => 표준 스트림이 더 이상 터미널 디바이스를 가리키지 않으므로, 커널이 해당 디바이스에 `EIO`를 반환하더라도 프로세스가 뜻밖의 쓰기 오류로 중단될 위험도 사라집니다.

이를 통해 세션 분리 대신 '시그널 무시 + I/O 리디렉션'을 통해 생존성과 로그 기록을 동시에 확보합니다.

## 세션을 아예 끊어 버린다면?

`setsid some_cmd`는 새로운 세션을 만들고 동시에 `CTTY`를 떼어 버립니다.
`nohup`보다 더 확실하지만, 표준 출력이 그대로 `TTY`를 바라보고 있다면 터미널이 닫히는 순간 해당 파일 디스크립터가 오류를 내기 때문에 여전히 리다이렉션을 함께 처리해줘야 합니다.

완전한 데몬화가 필요할 때는 "더블 포크" 기법을 이용합니다.
1. 첫 번째 `fork()` 후 부모가 종료
2. 자식이 `setsid()`로 세션 리더가 됨
3. 자식이 다시 `fork()`하여 자신도 종료함으로써 손자 프로세스는 세션 리더도 아니고 CTTY도 없는 상태가 됩

이런 패턴은 데몬 라이브러리(예: `libdaemon`)나 `systemd`의 `Type=forking` 서비스가 대신 처리해 줍니다.

## `&`만으로는 백그라운드 프로세스 상주 불가

결국 세션 및 `CTTY` 구조 때문입니다.
*커널이 세션 단위로 터미널 소멸을 전파*한다는 규칙이 변하지 않는 한, 터미널에 묶여 있는 프로세스는 `SIGHUP`의 기본 동작을 재정의하거나 세션 자체를 끊어야만 독립적으로 살아남을 수 있습니다.

`nohup`은 프로세스를 백그라운드에 상주시키는 가장 손쉬운 방법입니다.
만약 더 높은 수준의 요구(서비스 관리, 자원 제한, 자동 재시작 등)는 `systemd`, `tmux`, `screen`, 컨테이너 런타임, 혹은 데몬화 라이브러리 같은 별도의 도구로 처리해야 합니다.

## 예제

### `nohup`으로 vector 프로세스 관리하기 by ChatGPT

```sh
#!/usr/bin/env bash
#
#  ┌─────────────────────────────────────────────────────────────────────────┐
#  │ vectorctl.sh – Vector(Observability Agent)를 *nohup* 기반으로 제어     │
#  │ 학습 목적 : "각 줄이 실제로 무엇을 하고, 왜 그렇게 해야 하는가?"를     │
#  │            셸 스크립트 한눈에 파악하도록 **코드 옆**에 긴 설명 삽입     │
#  └─────────────────────────────────────────────────────────────────────────┘
#
#  ※ 프로덕션 서비스라면 systemd·supervisor 등을 권장하지만, systemd를
#    쓸 수 없는 제한 환경에서 *nohup*만으로 프로세스 생존 + 설정 리로드를
#    관리하는 실습 예제이다.
# ---------------------------------------------------------------------------

# ────────────────────────────────────────────────────────────────────────────
# 1) 경로·파일 변수 정의  (절대경로 권장 ─ '상대경로 + cd /' 조합은 위험)
# ────────────────────────────────────────────────────────────────────────────
VECTOR_BIN="/opt/vector/bin/vector"       # ❶ 실제 Vector 실행파일
VECTOR_CFG="/etc/vector/vector.toml"      # ❷ 설정 파일 (리로드 대상)
LOG_DIR="/var/log/vector"                 # ❸ 로그 전용 디렉터리
BOOT_LOG="$LOG_DIR/boot.log"              # ❹ nohup 첫머리 로그 (시작 실패 추적용)
RUNTIME_LOG="$LOG_DIR/runtime.log"        # ❺ Vector 내부 로거가 출력할 파일
PID_FILE="/run/vector.pid"                # ❻ '프로세스 ID'를 기록해 상태 추적

# ────────────────────────────────────────────────────────────────────────────
# 2) 디렉터리·권한 초기화
#    • /var/log/vector 같은 전용 폴더를 생성하여 root가 아닌 계정도 접근.
#    • PID 파일은 /run(=tmpfs)로 두어 재부팅 시 자동 정리.
# ────────────────────────────────────────────────────────────────────────────
mkdir -p "$LOG_DIR" "$(dirname "$PID_FILE")"
umask 077                # ❼ 새로 만드는 파일은 600/700 → 정보 노출 최소화

# ────────────────────────────────────────────────────────────────────────────
# 3) 헬퍼 함수 : "이미 실행 중인가?"
#    • PID_FILE 존재 + 실제 프로세스 생존( kill -0 ) 둘 다 만족해야 OK.
#      └─ kill -0 는 시그널을 보내지 않고 존재 여부만 검사하는 POSIX 규칙.
# ────────────────────────────────────────────────────────────────────────────
running() {
  [[ -f "$PID_FILE" ]] && kill -0 "$(cat "$PID_FILE")" 2>/dev/null
}

# ────────────────────────────────────────────────────────────────────────────
# 4) START : nohup + 백그라운드 기동 절차
#    핵심 포인트
#      a) cd /        → CWD 잠금 방지 (외장 FS, NFS 'device busy' 문제 예방)
#      b) nohup … & → SIGHUP 무시 + stdout/stderr → BOOT_LOG
#      c) echo $!    → 마지막 background PID를 파일에 기록
# ────────────────────────────────────────────────────────────────────────────
start() {
  if running; then
    echo "Vector already running (PID $(cat "$PID_FILE"))"
    return
  fi

  cd /                               # ⓫ 작업 디렉터리 잠금 해제
  [[ -f "$BOOT_LOG" ]] && mv "$BOOT_LOG" "$BOOT_LOG.$(date +%s)"  # ⓬ 로그 순환

  echo "Starting Vector ..."
  # ⓭ nohup:  SIGHUP → SIG_IGN, stdin=/dev/null, stdout/err→BOOT_LOG
  # ⓮ &   :  셸 job table에 백그라운드 등록
  nohup "$VECTOR_BIN" --config "$VECTOR_CFG" \
        >>"$BOOT_LOG" 2>&1 &
  echo $! > "$PID_FILE"              # ⓯ 백그라운드 PID( $! ) 저장

  sleep 1                            # ⓰ '찔끔' 대기 후 살아있는지 재확인
  if running; then
    echo "Vector started (PID $(cat "$PID_FILE"))"
  else
    echo "Vector failed – see $BOOT_LOG"
    rm -f "$PID_FILE"
    exit 1
  fi
}

# ────────────────────────────────────────────────────────────────────────────
# 5) STOP : 정상 종료 → 10초 대기 → 실패 시 SIGKILL
#    • 'kill PID'   : 프로세스에게 SIGTERM(15) 보냄 → Vector는 Graceful 종료
#    • 'kill -9'    : 10초 내 안 죽으면 최후통첩 (커널이 즉시 메모리 회수)
# ────────────────────────────────────────────────────────────────────────────
stop() {
  if ! running; then
    echo "Vector not running"
    return
  fi
  local pid=$(cat "$PID_FILE")
  echo "Stopping Vector (PID $pid) ..."
  kill "$pid"                        # ⓱ SIGTERM
  for _ in {1..10}; do               # ⓲ 최대 10초 기다림
    running || break
    sleep 1
  done
  if running; then
    echo "Graceful stop failed – sending SIGKILL"
    kill -9 "$pid"
  fi
  rm -f "$PID_FILE"
  echo "Vector stopped."
}

# ────────────────────────────────────────────────────────────────────────────
# 6) RELOAD : 설정 핫리로드 (SIGHUP)
#    • Vector는 main() 안에서 `signal(SIGHUP, reload_handler)` 식으로
#      *다시* 핸들러를 등록했기 때문에 nohup 초기 SIG_IGN을 덮어써둔 상태.
# ────────────────────────────────────────────────────────────────────────────
reload() {
  if ! running; then
    echo "Vector not running"
    exit 1
  fi
  echo "Reloading config with SIGHUP ..."
  kill -HUP "$(cat "$PID_FILE")"     # ⓳ '설정 다시 읽어!'
}

# ────────────────────────────────────────────────────────────────────────────
# 7) STATUS : PID 파일 + 프로세스 존재 여부 리포트
# ────────────────────────────────────────────────────────────────────────────
status() {
  if running; then
    echo "Vector running (PID $(cat "$PID_FILE"))"
  else
    echo "Vector not running"
  fi
}

# ────────────────────────────────────────────────────────────────────────────
# 8) 명령행 파서
# ────────────────────────────────────────────────────────────────────────────
case "$1" in
  start)   start   ;;
  stop)    stop    ;;
  reload)  reload  ;;
  status)  status  ;;
  *)
    echo "Usage: $0 {start|stop|reload|status}"
    ;;
esac

# ---------------------------------------------------------------------------
# 학습 체크포인트
#  • 'nohup'의 약한 분리 : SIGHUP 무시 + I/O 재배선, 세션은 그대로.
#  • Vector가 SIGHUP을 다시 잡아야 핫리로드가 작동한다.
#  • cd / : CWD 잠금을 풀어 파일시스템 관리(umount) 방해를 예방한다.
#  • PID + kill -0 : "파일만 남고 프로세스는 죽은" 좀비 PID 파일 검출.
#  • stop → SIGTERM → 유예 → SIGKILL : 리눅스 전통적인 데몬 종료 패턴.
# ---------------------------------------------------------------------------
```
