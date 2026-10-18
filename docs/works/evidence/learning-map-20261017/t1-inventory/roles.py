"""Draft the structural-role table for T1 from the inventory mark files.

Usage: python3 -I roles.py marks_e05.txt e05 marks_c3.txt c3 ... > ../t1-structural-roles.tsv

Every mark line whose role is 장소, 연결, 시각 or 사건 becomes one row: the input, the key path,
the mark's role and provenance, a decision, and the rule that made it. A decision is either a
structural column of the common format (the row then belongs to the anti-stuffing scan) or an
exception class from the T1 design memo, section 6 (the key may stay inside attrs). Paths that no
rule settles get 'TBD'; they must be decided by hand before that input's importer is written.
Rules are tried in order; the first match wins.
"""
import re
import sys

ROLES = {"장소", "연결", "시각", "사건"}

# (input set or None for all, path regex, decision, rule name). Order matters.
# Decisions made by hand for paths no rule settles, keyed by (input, path). Each carries its reason.
OVERRIDES = {
    ("e05", "$.steps[].place"): ("place", "M 멈춘 자리 = 프레임 장소(함수와 줄)"),
    ("e05", "$.steps[].place.loc"): ("place.loc", "M 프레임 장소의 소스 위치"),
    ("e05", "$.steps[].place.loc.path"): ("place.loc", "M 프레임 장소의 소스 경로"),
    ("e05", "$.steps[].place.loc.line"): ("place.loc", "M 프레임 장소의 줄"),
    ("e05", "$.steps[].values[]"): ("X7", "X7 멈춘 순간 읽은 값 목록(사건이 아니라 값)"),
    ("c3", "$.endUs"): ("end_ns", "M 질의 구간(받은 순간 0부터)의 끝"),
    ("c4", "$.data.endUs"): ("end_ns", "M 질의 구간(받은 순간 0부터)의 끝"),
    ("c4", "$.data.stages[].caller"): ("stack", "M 단계가 시작한 함수의 부른 쪽 = 스택 둘째 칸"),
    ("c4", "$.data.hops[].msgs[].cut"): ("link", "M 다시 셈: 나가는 continues 연결이 있거나 그런 메시지의 copy 사본이면 참(이 송신의 보이는 바이트 밖으로 이어짐, journey.go 190·219행)"),
    ("c4", "$.data.hops[].msgs[].cont"): ("link", "M 앞 송신의 잘린 메시지에서 오는 continues 연결(journey.go 346행)"),
    ("c4", "$.data.hops[].sysUs"): ("dur_ns", "M strace가 잰 시스템 콜 길이 = strace 출처의 시스템 콜 사건(길이만), 이동과 syscall 연결"),
    ("c4", "$.data.stages[].caller"): ("stack", "M 스택 0번 = 단계 함수를 부른 쪽 프레임"),
    ("c4", "$.data.kernelPid"): ("place.parent", "M 커널 마디 장소(커널 함수·프레임)의 프로세스"),
}

# 2026-10-18 구현 중 고침: 이동 하나가 장소 값을 여섯 가져(pid, proc, site, from, to, files) place/target 두 칸으로
# 정해지지 않았다. 사건의 place는 실행된 자리(gdb와 같은 뜻)로 두고 형식에 from 칸을 더했다(설계 메모 5의 2).
for _inp, _pre in (("c3", "$"), ("c4", "$.data")):
    OVERRIDES.update({
        (_inp, _pre + ".pid"): ("place", "M 질의 구간 사건의 장소 = 그 질의의 백엔드 프로세스"),
        (_inp, _pre + ".hops[].from"): ("from", "P1 이동의 출발 칸(형식의 from 칸)"),
        (_inp, _pre + ".hops[].pid"): ("place.parent", "M 이동을 실행한 프로세스 = 이동 장소(코드 자리)의 부모"),
        (_inp, _pre + ".hops[].proc"): ("recompute", "M 다시 셈: 이동의 프로세스가 질의 구간의 장소(백엔드)면 backend, 아니면 other(1,194행 모두 맞음)"),
        (_inp, _pre + ".hops[].site"): ("place", "M 이동의 장소 = 코드 자리(함수 이름 없는 경로:줄). 빈 글(5행)이면 프로세스"),
        (_inp, _pre + ".hops[].files"): ("from/target", "M 파일 장소(칸 아래)가 출발 칸이 파일 칸이면 from을, 아니면 target을 좁힘(198행 모두 하나)"),
        (_inp, _pre + ".hops[].files[]"): ("from/target", "M 파일 장소(칸 아래)가 출발 칸이 파일 칸이면 from을, 아니면 target을 좁힘(198행 모두 하나)"),
        (_inp, _pre + ".stages[].fn"): ("place", "M 단계 사건의 장소 = 단계 함수"),
        (_inp, _pre + ".stages[].at"): ("stack", "M 부른 줄 = 스택 0번 프레임의 위치(concept-3은 caller가 없어 코드 자리)"),
    })
for _path in ("$.data.hops[].kernel[]", "$.data.hops[].kernel[].kids[]", "$.data.hops[].kernel[].kids[]{2+}"):
    OVERRIDES.update({
        ("c4", _path + ".fn"): ("place", "M 커널 마디의 장소 = 커널 함수(at이 있으면 그 아래 프레임)"),
        ("c4", _path + ".at"): ("place.loc", "M 커널 프레임 장소의 위치(at이 없으면 함수 장소)"),
    })

# 2026-10-18 트랜잭션 실험 적재기 앞의 결정(설계 메모 5의 3). 수는 concept-5~8의 44벌에서 셌다(scratchpad t1/txcheck*.py와 같은 셈).
# 세션은 txlab 장소 아래의 장소이고, 세션과 백엔드 프로세스의 짝은 '단계 contains 요청'의 두 장소로 남는다(세션마다 백엔드 하나,
# kill 뒤에도 같은 세션이 다시 접속한 기록은 없음). 그래서 세션의 pid, 프로세스의 session, kill의 session은 다시 센다.
OVERRIDES.update({
    ("tx", "$.sessions[]"): ("place", "M 세션 장소(kind session, txlab 장소 아래). 차례는 장소 id 차례"),
    ("tx", "$.sessions[].name"): ("place", "M 세션 장소의 이름"),
    ("tx", "$.sessions[].pid"): ("recompute", "M 다시 셈: 그 세션의 단계가 contains로 품은 요청 구간의 장소(백엔드 프로세스). 세션마다 하나가 아니면 적재기가 멈춤(88행 모두 하나)"),
    ("tx", "$.steps[].session"): ("place", "M 단계 구간의 장소 = 세션 장소"),
    ("tx", "$.server.procs[].pid"): ("place", "M 프로세스 장소의 이름(pid). 역할은 역할 장소(부모)로 둠(여정과 같게)"),
    ("tx", "$.server.procs[].session"): ("recompute", "M 다시 셈: 그 프로세스에 요청이 놓인 세션(88행 모두 sessions[].pid와 같고 모두 backend)"),
    ("tx", "$.server.events[].pid"): ("place.parent", "M 사건 장소(가장 안쪽 프레임, 스택이 없으면 그 자신)의 프로세스"),
    ("tx", "$.server.spans[].pid"): ("place.parent", "M 구간 장소(함수 입구 프레임)의 프로세스"),
    ("tx", "$.server.spans[].fn"): ("place", "M 구간의 장소 = 함수 입구 프레임(함수 장소 아래, 위치는 at)"),
    ("tx", "$.server.spans[].at"): ("place.loc", "M 함수 입구 프레임의 위치 = 함수가 정의된 줄(scenario_server.go 293행 serverSpan.At)"),
    ("tx", "$.bystanders[].pid"): ("place", "M 해석기 요약 사건(sample, 시각 미기록)의 장소 = 그 프로세스(110행 모두 server.procs에 있음)"),
    ("tx", "$.kills[].pid"): ("target", "M kill의 target = 죽인 백엔드 프로세스"),
    ("tx", "$.kills[].session"): ("recompute", "M 다시 셈: kill target 프로세스의 세션(8행 모두 맞음). kill의 장소는 txlab 장소"),
    ("tx", "$.kills[].tReady"): ("end_ns", "M kill 구간의 끝 = txlab이 server.log에서 접속 재개 줄을 본 시각(scenario.go 660행, 50 ms 간격으로 봄)"),
    ("tx", "$.kills[].walWrittenT"): ("link", "M kill → kill 전에 끝난 pg_wal 쓰기 가운데 쓴 범위의 끝(LSN)이 가장 먼 것의 wal 연결(근거 해석기, walExtentBefore). 시각은 그 사건의 끝으로 다시 셈(6행 모두 pg_wal 파일의 pwrite64 하나와 t1이 정확히 같고, 그 쓰기의 lsn1이 walWritten과 같음)"),
    ("tx", "$.kills[].beforeStep"): ("link", "M kill → 그 뒤에 보낸 단계의 before 연결(근거 txlab). 시나리오가 kill로 끝나면 단계 수 + 1이라 이을 단계가 없어 적재기가 멈춤(txlab/run.go 42행, 지금 기록에는 없음)"),
    ("tx", "$.pauses[].beforeStep"): ("link", "M 멈춤 → 그 뒤에 보낸 단계의 before 연결(근거 txlab, txlab/run.go 35행)"),
    ("tx", "$.captures[].afterStep"): ("link", "M 관찰 → 단계의 after 연결(근거 txlab). 0(첫 단계 앞, 44행)은 연결 없음이고 내보내기가 0으로 되살림"),
    ("tx", "$.captures[].afterKill"): ("link", "M 관찰 → kill의 after 연결(근거 txlab). 이 관찰은 kill 앞 마지막 단계와도 after로 잇는다(txlab/run.go 80행, 8행 모두 beforeStep - 1)"),
    ("tx", "$.hintChanges[].at"): ("place", "M 힌트 사건의 장소 = 코드 자리(SetHintBits를 부른 경로:줄, 함수 이름 없음)"),
    ("tx", "$.hintChanges[].by"): ("place.parent", "M 힌트를 바꾼 세션의 백엔드 = 힌트 장소(코드 자리)의 부모. 사건 출처가 추정이라 이 장소 선택도 추정. 내보내기는 그 프로세스의 세션 이름"),
    ("tx", "$.hintChanges[].step"): ("link", "M 단계 → 힌트 사건의 contains 연결(근거 힌트 출처, 추정). 18행은 by와 그 단계의 세션이 달라 by로 다시 셀 수 없음"),
    ("tx", "$.hintChanges[].table"): ("X2", "X2 데이터 대상(표)의 이름"),
    ("tx", "$.steps[].requests[].checks[].table"): ("X2", "X2 데이터 대상(표)의 이름"),
    ("tx", "$.steps[].requests[].checks[].path"): ("recompute", "M 다시 셈: 판정 경로 사건(이름 path)이 없으면 null(4행). 빈 목록은 원본에 없음"),
    ("tx", "$.steps[].requests[].checks[].path[]"): ("row", "R1 판정 경로의 호출 하나 = 줄 판정 사건의 자식 instant(이름 path, 시각 미기록)"),
    ("tx", "$.tables[]"): ("X2", "X2 데이터 대상(표)의 정의 = run attrs의 값(txlab 출처)"),
    ("tx", "$.tables[].name"): ("X2", "X2 데이터 대상(표)의 이름"),
    ("tx", "$.tables[].columns[]"): ("X2", "X2 데이터 대상(열)의 정의"),
    ("tx", "$.tables[].columns[].name"): ("X2", "X2 데이터 대상(열)의 이름"),
    ("tx", "$.tables[].indexes[]"): ("X2", "X2 데이터 대상(색인)의 정의"),
    ("tx", "$.tables[].indexes[].name"): ("X2", "X2 데이터 대상(색인)의 이름"),
    ("tx", "$.tupleAddrs"): ("X2", "X2 데이터 대상(행)의 공유 버퍼 주소 = run attrs의 값(해석기 출처)"),
    ("tx", "$.tupleAddrs[]"): ("X2", "X2 데이터 대상(행)의 공유 버퍼 주소"),
    ("tx", "$.tupleAddrs[].addrs"): ("X2", "X2 데이터 대상(행)의 공유 버퍼 주소"),
    ("tx", "$.tupleAddrs[].addrs[]"): ("X2", "X2 데이터 대상(행)의 공유 버퍼 주소"),
    ("tx", "$.tupleAddrs[].table"): ("X2", "X2 데이터 대상(표)의 이름"),
    ("tx", "$.captures[].pages.parent[].values.name"): ("X2", "X2 행 값(표 parent의 name 열)"),
    ("tx", "$.captures[].pages.doctors[].values.name"): ("X2", "X2 행 값(표 doctors의 name 열)"),
})
# 줄 판정 안의 세 목록(줄, 판정 경로, 경로 안의 호출)은 fn이 부른 함수이고 at이 부른 줄이다(heapam_visibility.c:977의
# HeapTupleHeaderXminCommitted). 여정 concept-3의 단계와 같은 꼴: 장소는 부른 함수, 스택은 [그 함수, 부른 줄의 코드 자리].
for _path in ("$.steps[].requests[].checks[].lines[]", "$.steps[].requests[].checks[].path[]", "$.steps[].requests[].checks[].path[].inner[]"):
    OVERRIDES.update({
        ("tx", _path + ".fn"): ("place", "M 사건의 장소 = 부른 함수(백엔드 프로세스 아래)"),
        ("tx", _path + ".at"): ("stack", "M 부른 줄 = 스택 1번의 코드 자리(함수 이름 없음)"),
    })

RULES = [
    # Elements of record-forming lists become rows of a table.
    (None, r"\.(hops|msgs|kernel|kids|stages|steps|requests|in|out|events|snapshots|checks|lines|inner|captures|kills|pauses|spans|procs|bystanders|hintChanges|messages|borrows|checkpoints|notes|throws|parks|tree|effects|binding|slots|waits|waitCycles|blocked|transactions|clog)\[\](\{2\+\})?$", "row", "R1 행을 만드는 목록의 원소"),
    # Containers and lists of records are not leaves; their elements carry the decision.
    (None, r"\.stack(\[\])?$", "stack", "S1 스택 그릇"),
    (None, r"\.stack\[\]\.(fn|at|function|loc|loc\.path|loc\.line|depth)$", "stack", "S1 스택 칸 = 프레임 장소"),
    (None, r"\.(at)\[\]$", "stack", "S1 JFR 스택 칸"),
    (None, r"\.(at)$", "stack", "S1 JFR 스택 그릇(목록)"),
    # Data objects named inside records are values in the v0 map model (memo 6, class X2).
    (None, r"\.(xid|xids?|ctid|tup|tid|rel|relation|blk|buf|off|sxact|txid|page|tuple|vxid|block|lp|oid)$", "X2", "X2 데이터 대상의 이름"),
    (None, r"\.f\.[a-zA-Z]+$", "X2", "X2 기록기 칸 f 안의 대상 이름"),
    (None, r"\.values\.id$", "X2", "X2 행 값"),
    (None, r"\.(table|columns\[\]\.name|indexes\[\]\.name|name)$", "TBD", "이름 칸: 장소 이름인지 값인지 입력마다"),
    # Rows read by one observation stay with the sample (D10, class X3).
    (None, r"\.captures\[\]\.(pages|activity|locks|visible|xacts)", "X3", "X3 관찰 안의 행"),
    # Original numbering of a record itself (class X1); order is also kept by id order.
    (None, r"\.steps\[\]\.seq$", "X1", "X1 원래 번호(STOP 순번)"),
    (None, r"\.hops\[\]\.n$", "X1", "X1 원래 번호(이동)"),
    (None, r"\.view\.tree\[\]\.id$", "X1", "X1 원래 번호(트리 마디)"),
    (None, r"\.requests\[\]\.seq$", "X1", "X1 원래 번호(요청)"),
    (None, r"\.steps\[\]\.n$", "X1", "X1 원래 번호(단계)"),
    (None, r"\.view\.waits\[\]\.seq$", "X1", "X1 원래 번호(기다림의 요청 번호)"),
    # Cross-run comparisons produced by the decoder (class X4).
    (None, r"\.view\.(branches|compared)", "X4", "X4 교차 기록 비교"),
    # Decoder call relation on observed JFR spans; the same relation is structural in the tree (X5).
    (None, r"\.java\.spans\[\]\.(caller|depth)$", "X5", "X5 관찰 구간에 해석기가 붙인 호출 관계"),
    # Durations, settings and clock offsets are values, not event times (class X6).
    (None, r"\.(idleMs|self|durs\[\]|maxDeviationUs|medianOffsetUs|lo|hi|.*Ms|durUs)$", "X6", "X6 길이·설정·시계 차"),
    # Event times.
    (None, r"\.(t|t0|sent|accept|tCapture|tGo|tc)$", "start_ns", "T1 사건 시작(시계는 출처)"),
    (None, r"\.(t1|done|close)$", "end_ns", "T1 사건 끝"),
    (None, r"\.(us)$", "dur_ns", "T2 길이만 아는 사건"),
    (None, r"\.(sysUs)$", "TBD", "시스템 콜 시간: 짝 사건의 길이인지"),
    (None, r"\.reads\.[a-z]+(\[\])?$", "start_ns/end_ns", "T3 관찰 안의 읽기 구간 = 자식 구간"),
    (None, r"\.reads$", "start_ns/end_ns", "T3 관찰 안의 읽기 구간 그릇"),
    # Event kind.
    (None, r"\.(kind|type|dir)$", "name", "K1 사건 종류"),
    # Links.
    (None, r"\.view\.tree\[\]\.(parent|kids|kids\[\])$", "parent", "L1 트리 포함"),
    (None, r"\.view\.roots(\[\])?$", "parent", "L1 트리 뿌리 = 부모 없음"),
    (None, r"\.kernel\[\]\.(kids\[\]\.)*kids$|\.kids\[\]\{2\+\}\.kids$|\.kernel\[\]\.kids\[\]\.kids$", "parent", "L1 커널 마디 포함"),
    (None, r"\.(from)$", "place", "P1 이동의 출발 칸"),
    (None, r"\.(to)$", "target", "P2 이동의 도착 칸"),
    (None, r"\.(pid|thread|proc|conn|lane|under|waiters(\[\])?|files(\[\])?|function|fn|site|desc|kernelPid)$", "place/target", "P3 실행 위치"),
    (None, r"\.procs(\.[a-z_]+)?(\[\])?$", "place", "P3 서버 프로세스 역할표"),
    (None, r"\.sessions(\.\{\*\})?$", "place", "P3 세션"),
    (None, r"\.conns\[\]$", "place", "P3 연결"),
]


def classify(inp, path):
    if (inp, path) in OVERRIDES:
        return OVERRIDES[(inp, path)]
    for only, pat, dec, rule in RULES:
        if only and inp not in only:
            continue
        if re.search(pat, path):
            return dec, rule
    return "TBD", "규칙 없음"


def main(args):
    print("input\tpath\trole\tprov\tdecision\trule")
    for i in range(0, len(args), 2):
        f, inp = args[i], args[i + 1]
        for line in open(f, encoding="utf-8"):
            if not line.startswith("$") or "|" not in line:
                continue
            parts = [p.strip() for p in line.split("|")]
            path = parts[0].split()[0]
            role, prov = parts[1], parts[2]
            if role not in ROLES:
                continue
            dec, rule = classify(inp, path)
            print(f"{inp}\t{path}\t{role}\t{prov}\t{dec}\t{rule}")


if __name__ == "__main__":
    main(sys.argv[1:])
