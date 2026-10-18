"""gdb 안에서 실행해, 지정한 함수에 멈출 때마다 함수·소스 위치·주요 값을 한 줄씩 기록합니다.

사용:
    gdb -p <백엔드 PID> -batch -x trace.py
설정(환경 변수):
    TRACE_WATCH  "함수<TAB>gdb 식" 줄로 된 파일. 식이 비어 있으면 위치만 기록합니다.
                 같은 함수에 여러 줄을 두면 식을 차례로 평가합니다.
    TRACE_BT_AT  이 함수에 처음 멈출 때 전체 호출 스택을 기록합니다(선택).
    TRACE_END    이 함수에 멈추면 기록을 끝내고 프로세스에서 분리합니다(기본 ReadyForQuery).
    TRACE_READY  중단점 설정을 마친 뒤 만들 표시 파일. 쿼리를 보내는 쪽이 이 파일을 기다립니다.
    PG_SRC       경로를 상대경로로 줄일 소스 트리 위치(기본 /src/postgresql-18.6).
출력:
    STOP <순번> <함수> <소스 상대경로>:<줄> [식=값 ...]
    BT   <깊이> <함수> <소스 상대경로>:<줄>
    END  <함수별 멈춘 횟수>
실패:
    TRACE_WATCH의 함수가 바이너리에 없으면 gdb가 중단점을 만들지 못합니다.
    이 경우 "NOBREAK <함수>"를 기록하고 나머지 함수로 계속합니다.
    식을 평가하지 못하면 값 자리에 <unavailable: 이유>를 남깁니다.
"""

import os
from collections import Counter, defaultdict

import gdb

SRC = os.environ.get("PG_SRC", "/src/postgresql-18.6").rstrip("/") + "/"
END = os.environ.get("TRACE_END", "ReadyForQuery")
BT_AT = os.environ.get("TRACE_BT_AT", "")
hits = Counter()
step = 0


def where(frame):
    """프레임의 함수 이름과 소스 위치를 반환합니다.

    Args:
        frame: gdb.Frame.
    Returns:
        (함수 이름, "상대경로:줄") 튜플. 소스 정보가 없으면 위치는 "?".
    """
    sal = frame.find_sal()
    name = frame.name() or "?"
    if sal.symtab is None:
        return name, "?"
    path = sal.symtab.fullname()
    if path.startswith(SRC):
        path = path[len(SRC):]
    return name, f"{path}:{sal.line}"


def evaluate(expr):
    """현재 프레임에서 gdb 식을 평가해 한 줄 문자열로 반환합니다.

    Args:
        expr: gdb 식(예: "query_string").
    Returns:
        값의 문자열 표현. 실패하면 "<unavailable: 이유>".
    """
    try:
        return str(gdb.parse_and_eval(expr)).replace("\n", " ")
    except gdb.error as err:
        return f"<unavailable: {err}>"


class TraceBreakpoint(gdb.Breakpoint):
    """멈출 때마다 기록만 하고 실행을 계속하는 중단점입니다."""

    def __init__(self, function, exprs):
        super().__init__(function, internal=False)
        self.function = function
        self.exprs = exprs

    def stop(self):
        """멈춘 지점을 기록합니다.

        Returns:
            False(실행을 계속함). 끝 함수에서는 분리를 예약하고 True를 반환해 멈춥니다.
        """
        global step
        step += 1
        hits[self.function] += 1
        frame = gdb.newest_frame()
        name, loc = where(frame)
        values = " ".join(f"{e}={evaluate(e)}" for e in self.exprs)
        print(f"STOP {step} {name} {loc} {values}".rstrip(), flush=True)

        # 깊은 지점에서 한 번만 전체 호출 스택을 남기면, 바깥 함수들의 위치까지 같은 실행에서 확인됩니다.
        if self.function == BT_AT and hits[self.function] == 1:
            depth, f = 0, frame
            while f is not None:
                fname, floc = where(f)
                print(f"BT {depth} {fname} {floc}", flush=True)
                depth += 1
                f = f.older()

        return self.function == END


# 감시 목록을 읽어 함수마다 평가할 식을 모읍니다.
watch = defaultdict(list)
with open(os.environ["TRACE_WATCH"], encoding="utf-8") as fh:
    for raw in fh:
        line = raw.rstrip("\n")
        if not line or line.startswith("#"):
            continue
        func, _, expr = line.partition("\t")
        if expr:
            watch[func].append(expr)
        else:
            watch.setdefault(func, [])
watch.setdefault(END, [])

# 중단점을 만들고, 만들지 못한 함수는 기록에 남깁니다.
gdb.execute("set pagination off")
gdb.execute("set confirm off")
# 없는 함수를 "나중에 생길 수도 있는" 보류 중단점으로 만들지 않고 바로 오류로 받습니다.
gdb.execute("set breakpoint pending off")
for func, exprs in watch.items():
    try:
        TraceBreakpoint(func, exprs)
    except gdb.error:
        print(f"NOBREAK {func}", flush=True)

# 준비가 끝났음을 알린 뒤, 끝 함수에 멈출 때까지 실행을 계속합니다.
ready = os.environ.get("TRACE_READY")
if ready:
    open(ready, "w").close()
gdb.execute("continue")
print("END " + " ".join(f"{k}={v}" for k, v in sorted(hits.items())), flush=True)
gdb.execute("detach")
