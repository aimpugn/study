"""Draft the structural-role table for T1 from the inventory mark files.

Usage: python3 -I roles.py marks_e05.txt e05 marks_c3.txt c3 ... > ../t1-structural-roles.tsv

Every mark line whose role is 장소, 연결, 시각 or 사건 becomes one row: the input, the key path,
the mark's role and provenance (or the fixed provenance from PROV_FIXES), a decision, and the rule that made it. A decision is either a
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

# 2026-10-18 Spring 적재기 앞의 결정(설계 메모 5의 4). 근거와 수는 ../t1-e03-evidence.md이고, 괄호 안의 C·R 번호는 그 수를 다시 세는
# t1-inventory/e03/count.py의 검사입니다(모두 n/n). 미정 44행과, 규칙이 정했지만 뜻이 맞지 않거나(정규식이 잘못 맞음) 구현하기에 모호한 행을
# 덮어씁니다. attrs에 남는 그릇(모음) 아래의 구조 역할은 검사 3이 막으므로 X 부류여야 하고, 다시 세는 칸은 attrs에 두지 않는 칸입니다.
_E03 = {
    # 멈춤 자리·막힘 관찰·관찰. 멈춤 자리는 [t, tGo] 구간(둘 다 맥 시계), 관찰과는 txlab이 남긴 captureNs로 same 연결.
    "$.checkpoints[].tGo": ("end_ns", "M 멈춤 자리 구간의 끝 = go를 보낸 맥 시각(txlab/spring.go 327). 68행 모두 t보다 늦음(C2b)"),
    "$.checkpoints[].tCapture": ("recompute", "M 다시 셈: same 연결로 잇는 관찰의 t(txlab이 관찰기에게 받은 captureNs, C1a 68/68). 같은 순간을 사건 둘로 만들지 않음"),
    "$.blocked[].tCapture": ("recompute", "M 다시 셈: same 연결로 잇는 관찰의 t(C1b 6/6)"),
    "$.blocked[].after": ("link", "M 막힘 관찰 → 그 앞 멈춤 자리의 after 연결(근거 txlab, txlab/spring.go 353. n이 같은 자리 하나, C1c 6/6)"),
    "$.blocked[].waiters": ("target", "M 막힘 관찰의 target = 락을 기다린 백엔드 프로세스(txlab이 pid로 적고 해석기가 이름으로 바꿈, spring.go 627~633). 0 또는 1개(C21 6/6), 둘 이상이면 적재기가 멈춤. 이름은 그 프로세스의 연결에서 다시 셈"),
    "$.blocked[].waiters[]": ("target", "M 위와 같음"),
    "$.captures[].afterStep": ("recompute", "M 다시 셈: same 연결로 잇는 멈춤 자리의 n, 막힘 관찰이면 그 after(C1a·C1b 74/74)"),
    "$.captures[].span": ("end_ns", "M span[1] = 관찰 sample의 끝(관찰용 접속이 마지막 부분의 답을 보낸 sendto, spring.go 772~803 captureSpans). span[0]은 관찰 t와 같아 다시 셈(C5 74/74). 관찰은 sample이라 자식 읽기 구간과 어긋남을 따지지 않음"),
    "$.captures[].span[]": ("end_ns", "M 위와 같음"),
    "$.captures[].overlap[]": ("link", "M 관찰 → 관찰 span 안에 컨테이너 쪽 중계기가 읽은 메시지(컨테이너 쪽 사건)의 overlap 연결(근거 해석기, spring.go 601~605). 규칙으로 다시 센 목록과 차례까지 같고(C7 74/74) 원소마다 메시지 하나와 같음(C7b 17/17)"),
    "$.captures[].overlap[].t": ("recompute", "M 다시 셈: overlap 연결이 가리키는 메시지에서"),
    "$.captures[].overlap[].tc": ("recompute", "M 다시 셈: overlap 연결이 가리키는 메시지에서"),
    "$.captures[].overlap[].dir": ("recompute", "M 다시 셈: overlap 연결이 가리키는 메시지에서"),
    "$.captures[].overlap[].type": ("recompute", "M 다시 셈: overlap 연결이 가리키는 메시지에서"),
    "$.captures[].overlap[].name": ("recompute", "M 다시 셈: overlap 연결이 가리키는 메시지에서"),
    # 연결과 연결 위 메시지. 메시지 하나는 맥 쪽(t)과 컨테이너 쪽(tc) 사건 둘이고, 백엔드 쪽 끝은 target(F)·from(B)입니다.
    "$.conns[].conn": ("recompute", "M 다시 셈: 연결 장소 이름 C<n>의 n(spring.go 473, C9a 24/24)"),
    "$.conns[].name": ("place", "M 연결 장소(kind connection)의 이름"),
    "$.conns[].pid": ("from/target", "M 연결 위 메시지의 target(F)·from(B) = 그 연결의 백엔드 프로세스 장소(이름 pid). txlab이 BackendKeyData로 안 값(txlab/spring.go 224~226, C9b 24/24). 내보내기는 그 장소 이름"),
    "$.conns[].thread": ("recompute", "M 다시 셈: 그 연결의 첫 Q(맥 시각)를 감싼 QueryExecutorImpl.execute JFR 구간 가운데 구간 차례로 첫 것의 스레드(spring.go 511~514, spring_jfr.go 414~421, C13 24/24). 시계가 다른 두 값을 견준 해석기의 짝짓기라 구조로 두지 않음"),
    "$.conns[].messages[].t": ("start_ns", "M 맥 쪽 중계기 사건의 시각(mac 시계)"),
    "$.conns[].messages[].tc": ("start_ns", "M 컨테이너 쪽 중계기 사건의 시각(container 시계). 두 사건은 message 연결(근거 해석기: 같은 순번에 방향·번호·종류·길이가 같음, spring.go 204~220, C8 24/24)"),
    "$.conns[].messages[].dir": ("from/target", "M F면 target이, B면 from이 그 연결의 백엔드 프로세스 장소. 사건 이름은 type"),
    "$.conns[].messages[].name": ("X2", "X2 준비 문장 이름(Parse·Bind의 name, 값은 S_1뿐 38행). 실행 위치가 아닌 데이터 대상"),
    "$.conns[].requests[].seq": ("recompute", "M 다시 셈: 그 연결의 요청 차례(0은 접속 인증, C10c 166/166). 요청 사건은 contains로 품은 메시지에서 모든 칸을 다시 세어 attrs가 없음"),
    "$.conns[].requests[].server.events[].f.at": ("recompute", "M 다시 셈: hint·ereport 사건의 부른 줄 = 스택(안쪽부터) 1번 칸의 코드 자리(해석기는 scenario.go 908·1012에서 따로 셈하지만 16/16 같음, C14). 코드 자리는 장소라 X2(데이터 대상의 이름)가 아님. 그래서 f를 통째로 두지 않고 f.<칸>으로 나눔(코드 검토 지적 2)"),
    "$.conns[].requests[].server.checks[].table": ("X2", "X2 데이터 대상(표)의 이름"),
    "$.conns[].requests[].server.checks[].path": ("recompute", "M 다시 셈: 판정 경로 사건(이름 path)이 없으면 null. 빈 목록은 원본에 없음"),
    "$.conns[].requests[].server.checks[].path[]": ("row", "R1 판정 경로의 호출 하나 = 줄 판정 사건의 자식 instant(이름 path, 시각 미기록)"),
    "$.sessions.{*}": ("recompute", "M 다시 셈: 연결 장소마다 {백엔드 pid: 연결 이름}(spring.go 474, C9e 14/14)"),
    # 트랜잭션. ReadyForQuery 상태로 해석기가 정한 구간이고, 품은 요청과 끝난 뒤 첫 앱 메시지를 연결로 둡니다.
    "$.transactions[].conn": ("place", "M 트랜잭션 구간의 장소 = 연결 장소"),
    "$.transactions[].first": ("recompute", "M 다시 셈: contains로 품은 요청 가운데 seq가 가장 작은 것(spring.go 311~338, C12b 20/20)"),
    "$.transactions[].last": ("recompute", "M 다시 셈: contains로 품은 요청 가운데 seq가 가장 큰 것(C12b 20/20)"),
    "$.transactions[].xids[]": ("recompute", "M 다시 셈: 품은 요청의 서버 창 사건에서 xid(그 앞 assign_xid의 level, 요청마다 1부터)와 commit_clog·abort_clog의 사본(spring.go 537~553, C12c 20/20, 44행)"),
    "$.transactions[].xids[].xid": ("recompute", "M 위와 같음"),
    "$.transactions[].xids[].t": ("recompute", "M 위와 같음"),
    "$.transactions[].clog[]": ("recompute", "M 위와 같음"),
    "$.transactions[].clog[].kind": ("recompute", "M 위와 같음"),
    "$.transactions[].clog[].t": ("recompute", "M 위와 같음"),
    "$.transactions[].clog[].xid": ("recompute", "M 위와 같음"),
    "$.transactions[].next": ("link", "M 트랜잭션 → 끝난 뒤 그 연결에 앱이 처음 보낸 메시지(맥 쪽 사건)의 next 연결(근거 해석기, spring.go 326~332, C12b 20/20)"),
    "$.transactions[].next.t": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    "$.transactions[].next.tc": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    "$.transactions[].next.dir": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    "$.transactions[].next.type": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    "$.transactions[].next.len": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    "$.transactions[].next.text": ("recompute", "M 다시 셈: next 연결이 가리키는 메시지에서"),
    # JFR. 장소는 스레드 → 메서드(fn) → 서술자(desc)이고, 서술자가 없는 프레임 글(fn:줄)은 메서드 아래 프레임입니다.
    "$.java.spans[].fn": ("place", "M 구간 장소(서술자)의 부모 = 메서드 장소(스레드 아래)"),
    "$.java.spans[].desc": ("place", "M 구간의 장소 = 서술자 장소(메서드 아래)"),
    "$.java.spans[].thread": ("place", "M 메서드 장소의 부모 = 스레드 장소(앱 장소 아래)"),
    "$.java.throws[].thread": ("place", "M 예외 사건의 장소(at[0] 프레임)에서 메서드를 거쳐 올라간 스레드 장소"),
    "$.java.parks[].thread": ("place", "M 잠듦 사건의 장소(at[0] 프레임)에서 메서드를 거쳐 올라간 스레드 장소"),
    "$.java.parks[].under": ("X5", "X5 해석기가 스택 전체에서 고른 첫 추적 대상 칸(spring_jfr.go 299~308). 4행 모두 at[] 앞 6칸 밖이라(C22) 구조에서 다시 셀 수 없음"),
    # 호출 트리. 마디는 해석기 출처의 구간이고 장소는 main 스레드 아래 서술자, 스택은 [마디, 숨은 칸(안쪽부터), 부모 메서드의 부른 줄]입니다.
    "$.view.tree[].fn": ("place", "M 마디 장소(서술자)의 부모 = 메서드 장소(main 스레드 아래, spring_tree.go 22 sceneThread)"),
    "$.view.tree[].desc": ("place", "M 마디의 장소 = 서술자 장소"),
    "$.view.tree[].lane": ("recompute", "M 다시 셈: fn의 앞부분 규칙(spring_tree.go 265~279 laneOf, C15d 1,388/1,388)"),
    "$.view.tree[].site": ("stack", "M 스택 마지막 칸 = 부모 메서드 장소 아래 그 줄의 프레임. 뿌리는 0(C15f 16/16)이고 부모 칸이 없음"),
    "$.view.tree[].hidden": ("stack", "M 스택 가운데 칸 = 부모와 이 마디 사이의 추적하지 않은 JFR 칸(원본은 바깥부터라 뒤집음)"),
    "$.view.tree[].hidden[]": ("stack", "M 위와 같음"),
    "$.view.tree[].t0": ("start_ns", "M 기록된 마디의 시작. 스택에만 보인 54마디(t1 null)는 처음 본 시각이라 시작을 비우고 기록된 자손의 t0 최솟값으로 다시 셈(C15k 54/54)"),
    "$.view.tree[].seen": ("recompute", "M 다시 셈: 기록된 자손 마디의 t1 최댓값(spring_tree.go 213~214, C15k 54/54)"),
    "$.view.tree[].kids": ("parent", "L1 다시 셈: parent가 이 마디인 마디를 t0 차례로, 같으면 id 차례(spring_tree.go 222~240, C15b 1,388/1,388)"),
    "$.view.tree[].kids[]": ("parent", "L1 위와 같음(목록 원소 규칙이 잘못 맞았음)"),
    "$.view.tree[].branch": ("recompute", "M 다시 셈: key = fn + desc(lab. 마디 24개는 없음, C15g), alt = run attrs의 view.branches[key].alts 가운데 자식 차례·끝난 꼴·마지막 자식 꼴이 맞는 길(spring_compare.go 82~111). 맞는 길이 하나가 아니면 적재기가 멈춤(C15h·C15i 1,364/1,364)"),
    "$.view.tree[].branch.key": ("recompute", "M 위와 같음"),
    "$.view.tree[].branch.alt": ("recompute", "M 위와 같음"),
    # 효과. 종류마다 자리가 다릅니다(C16a: 한 마디 안의 차례는 묶임 효과 → 종류 효과 → 잠듦 → 예외, 1,388/1,388).
    "$.view.tree[].effects[]": ("row", "R1 종류마다: autocommit·validate·rollbackOnly는 마디의 자식 instant(해석기 출처), return·suspend·resume은 추정 출처 instant(마디 → contains), wire·borrow·park·throw는 마디에서 요청·빌림·JFR 잠듦·예외로 가는 연결(사본, C16b~f)"),
    "$.view.tree[].effects[].conn": ("target", "M return·suspend·resume 효과 사건의 target = 옮긴 연결(빈 글이면 없음). borrow·wire 효과는 연결한 빌림·요청에서 다시 셈"),
    "$.view.tree[].effects[].seq": ("link", "M 마디(가장 바깥 QueryExecutorImpl.execute) → 그 호출이 열려 있는 동안 보낸 요청의 wire 연결(근거 추정: 맥 시계의 보낸 시각을 JFR 시계의 구간에 견줌, spring_tree.go 852~863). conn·seq는 연결한 요청에서 다시 셈(C16b 1,388/1,388)"),
    "$.view.tree[].effects[].seq[]": ("link", "M 위와 같음"),
    "$.view.tree[].effects[].t0": ("recompute", "M 다시 셈: park·throw 효과는 연결한 JFR 잠듦·예외의 시각(C16e 4/4, C16f 30/30)"),
    "$.view.tree[].effects[].t1": ("recompute", "M 다시 셈: park 효과는 연결한 JFR 잠듦의 끝(C16e 4/4)"),
    "$.view.tree[].effects[].at": ("recompute", "M 다시 셈: park는 연결한 잠듦의 under(X5), throw는 연결한 예외의 at[]에서 마디의 메서드와 이름이 같은 첫 칸(C16e 4/4, C16f 30/30)"),
    "$.view.tree[].effects[].idleMs": ("recompute", "M 다시 셈: borrow 효과는 연결한 빌림 사건의 idleMs(C16c 24/24)"),
    # 빌림, 연결 묶임, 슬롯, 기다림, 기다림 고리.
    "$.borrows[].thread": ("place", "M 빌림 사건의 장소 = 앱 스레드 장소"),
    "$.borrows[].conn": ("target", "M 빌림 사건의 target = 받은 연결 장소(100 ms 창으로 붙인 추정, spring_jfr.go 435~479). 빈 글 4행은 모두 시간 초과라(C20a) target 없음과 ambiguity note"),
    "$.view.binding[]": ("row", "R1 묶임 상태 하나 = 추정 출처 instant(main 스레드). 첫 상태(장면 시작, 맥 시계)는 늘 빈 상태라 다시 셈(C17d 14/14)"),
    "$.view.binding[].bound": ("target", "M 묶임 상태 사건의 target = 묶인 연결 장소(빈 글이면 target 없음)"),
    "$.view.binding[].held": ("recompute", "M 다시 셈: 앞 상태부터 by 마디가 doSuspend면 앞 상태의 bound를 쌓고 doResume이면 뺌(C17f 56/56)"),
    "$.view.binding[].held[]": ("recompute", "M 위와 같음"),
    "$.view.binding[].by": ("link", "M 묶임 상태 → 그 상태를 만든 트리 마디의 by 연결(근거 추정). -1(빌림으로 묶음)이면 연결 없음(spring_tree.go 403~465, C17a 14/14)"),
    "$.view.slots[].conn": ("target", "M 슬롯 구간(main 스레드, 추정 출처)의 target = 쓴 연결 장소. t1 0(돌려주지 않음)이면 끝을 비움(지금 0행)"),
    "$.view.waits[].conn": ("place.parent", "M 기다림 구간(해석기 출처)의 장소 = 그 연결의 백엔드 프로세스(서버 창이 contains로 품음). 이름은 장소에서 다시 셈"),
    "$.view.waits[].seq": ("recompute", "M 다시 셈: 기다림을 품은 서버 창에 server 연결로 이어진 요청의 seq(X1이 아니라 요청을 가리킴, C18a 14/14). 끝이 요청의 done(맥 시계)인 기다림은 컨테이너 시계 구간이 될 수 없어 적재기가 멈춤(지금 0행, C18b)"),
    "$.view.waitCycles[]": ("recompute", "M 다시 셈: 기다림·트리 마디와 효과·묶임 상태·슬롯·xid 사건에서(spring_tree.go 532~626, C19 14/14). 시각이 종류마다 다른 시계(server는 컨테이너, pool은 JFR)이고 칸이 모두 다른 행의 사본"),
    "$.view.waitCycles[].kind": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].t0": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].t1": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].xid": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].thread": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].waiter": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].holder": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].call": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].heldBy": ("recompute", "M 위와 같음"),
    "$.view.waitCycles[].cycle": ("recompute", "M 위와 같음"),
    "$.view.window": ("recompute", "M 다시 셈: [start 멈춤 자리의 tGo, end 멈춤 자리의 t](spring.go 714~720, C3 14/14)"),
    "$.view.window[]": ("recompute", "M 위와 같음"),
    "$.view.pauses": ("recompute", "M 다시 셈: 멈춤 자리마다 [t, tGo](spring.go 721~723, C2 14/14)"),
    "$.view.pauses[]": ("recompute", "M 위와 같음(목록 원소 규칙이 잘못 맞았음)"),
    "$.view.pauses[][]": ("recompute", "M 위와 같음"),
    "$.view.branches.{*}.lines[]": ("X4", "X4 교차 기록 비교(목록 원소 규칙이 잘못 맞았음)"),
    # 표 정의는 트랜잭션 실험과 같게 데이터 대상입니다.
    "$.tables[]": ("X2", "X2 데이터 대상(표)의 정의 = run attrs의 값(txlab 출처)"),
    "$.tables[].name": ("X2", "X2 데이터 대상(표)의 이름"),
    "$.tables[].columns[]": ("X2", "X2 데이터 대상(열)의 정의"),
    "$.tables[].columns[].name": ("X2", "X2 데이터 대상(열)의 이름"),
    "$.tables[].indexes[]": ("X2", "X2 데이터 대상(색인)의 정의"),
    "$.tables[].indexes[].name": ("X2", "X2 데이터 대상(색인)의 이름"),
}
# 줄 판정 안의 세 목록은 트랜잭션 실험과 같은 꼴입니다(fn은 부른 함수, at은 부른 줄).
for _path in ("$.conns[].requests[].server.checks[].lines[]", "$.conns[].requests[].server.checks[].path[]", "$.conns[].requests[].server.checks[].path[].inner[]"):
    _E03.update({
        _path + ".fn": ("place", "M 사건의 장소 = 부른 함수(백엔드 프로세스 아래)"),
        _path + ".at": ("stack", "M 부른 줄 = 스택 1번의 코드 자리(함수 이름 없음)"),
    })
# 2026-10-18 결정 검토(하위 에이전트, 「수정 후 승인」)의 지적으로 고친 행. 근거는 ../t1-e03-evidence.md 8절.
_E03.update({
    # 지적 1: 서버 창은 두 ReadyForQuery 입구 사이를 요청 하나로 보는 해석기 규칙의 창이라 도출(142개 가운데 122개가 첫 F가 닿기 전에 시작).
    "$.conns[].requests[].server.t0": ("start_ns", "M 서버 창 = 해석기 출처(spring-cont)의 구간의 시작: 앞 ReadyForQuery 입구(spring.go 516~535). 두 입구 사이를 요청 하나로 보는 해석기 규칙의 창이라 도출이고, 142개 가운데 122개가 그 요청의 첫 F가 컨테이너 쪽에 닿기 전에 시작함(백엔드가 앱을 기다린 시간이 듦)"),
    "$.conns[].requests[].server.t1": ("end_ns", "M 서버 창의 끝: 이번 ReadyForQuery 입구"),
    # 지적 2: span[1]은 해석기가 관찰용 접속의 strace에서 고른 값이라 관찰 사건의 끝이 아니라 해석기의 순간.
    "$.captures[].span": ("row", "M span[1] = 해석기 출처(spring-cont)의 순간 하나(관찰용 접속이 마지막 부분의 답을 보낸 sendto, spring.go 772~803 captureSpans)이고 관찰 → 그 순간의 end 연결. span[0]은 관찰 t와 같아 다시 셈(C5 74/74). 관찰 sample은 끝을 두지 않음(읽기 구간 370개 가운데 60개가 span[1]보다 늦게 끝남)"),
    "$.captures[].span[]": ("start_ns", "M span[1]은 그 해석기 순간의 시각, span[0]은 다시 셈"),
    # 지적 3: 연결 위 메시지의 f는 기록기 칸이 아니고, f.pid는 BackendKeyData가 알린 백엔드 pid(장소 식별자)라 X2가 아님.
    "$.conns[].messages[].f.pid": ("recompute", "M 다시 셈: B·K 메시지의 from(백엔드 프로세스 장소)의 이름(C9f 24/24). 기록기 칸 f의 X2 규칙이 연결 위 메시지의 f(wireText, spring.go 222~249)에 잘못 맞았음. f의 나머지 칸은 f.auth·f.code·f.severity·f.where 열쇠로 나눠 둠"),
    # 지적 4: caller는 해석기가 붙인 관계가 아니라 관찰한 JFR 스택의 0번 칸이고, 트리 마디가 없는 구간 258개에서는 유일한 호출 사슬 기록.
    "$.java.spans[].caller": ("stack", "M 구간 스택 1번 = JFR 스택 0번 칸(부른 쪽, spring_jfr.go 265~267)의 프레임(메서드 아래). 트리 마디가 없는 구간 258개(장면 밖 main 238, 다른 스레드 20)에서는 이것이 유일한 호출 사슬 기록이고, 1,592개 모두 비지 않음"),
    "$.java.spans[].depth": ("recompute", "M 다시 셈: 같은 스레드의 구간을 (t0 오름, t1 내림) 차례로 놓고 아직 끝나지 않은 호출의 수(spring_jfr.go 325~341, 구간의 스레드·시작·끝만 씀, C27 1,592/1,592). 첫 판의 X5는 트리 깊이로 다시 셀 수 없다는 것만 보았고(1,334개 가운데 750개가 다름) JFR 구간 자체로 셀 수 있는지는 보지 않았음(코드 검토 지적 1)"),
})
OVERRIDES.update({("e03", p): d for p, d in _E03.items()})

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


# Provenance fixes, keyed by (input, path): the mark file's 출처 was set before the format was designed, and a review
# showed it wrong for these paths. The table then carries the fixed 출처 (the Go loader topics/pgsql/roles reads it, so
# check 4 expects it), and the rule column keeps the mark's value and the reason. Each value is (fixed 출처, reason).
_READS = ("M 읽기 구간은 해석기 captureReads(decode/scenario_check.go 395~433)가 그 질의를 받은 recvfrom부터 관찰용 접속이 다음 질의를 "
          "받은 때(없으면 다음 관찰의 시작, 마지막 관찰이면 시작)까지로 정한 위쪽 경계라 도출. 끝은 그 읽기의 사건이 아님"
          "(t1-inventory/reads.py: tx 1,708개는 마지막 관찰 44, 그 밖은 모두 다음 질의를 받은 때. e03 370개는 다음 관찰의 시작 54, 마지막 14, "
          "다음 질의 302). Spring 코드 검토 지적 3a")
PROV_FIXES = {}
for _inp in ("tx", "e03"):
    PROV_FIXES[(_inp, "$.captures[].reads")] = ("도출", _READS)
    for _part in ("activity", "locks", "pages", "running", "xacts"):
        PROV_FIXES[(_inp, "$.captures[].reads." + _part)] = ("도출", _READS)
        PROV_FIXES[(_inp, "$.captures[].reads." + _part + "[]")] = ("도출", _READS)
PROV_FIXES[("e03", "$.java.spans[].caller")] = ("관찰", "M caller는 JFR 스택의 0번 칸(decode/spring_jfr.go 265~267)을 해석기가 "
                                                "'fn:줄' 글로 옮긴 것이라 관찰. Spring 결정 검토 지적 4, 코드 검토 지적 7")


def classify(inp, path):
    """Return (decision, rule) for one key path of one input: a hand decision if any, else the first matching rule, else TBD."""
    if (inp, path) in OVERRIDES:
        return OVERRIDES[(inp, path)]
    for only, pat, dec, rule in RULES:
        if only and inp not in only:
            continue
        if re.search(pat, path):
            return dec, rule
    return "TBD", "규칙 없음"


def main(args):
    """Print the table for the (mark file, input name) pairs in args, one row per structural-role mark line."""
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
            if (inp, path) in PROV_FIXES:
                fixed, why = PROV_FIXES[(inp, path)]
                rule = f"{rule} / 출처 고침(표시 {prov} → {fixed}): {why}"
                prov = fixed
            print(f"{inp}\t{path}\t{role}\t{prov}\t{dec}\t{rule}")


if __name__ == "__main__":
    main(sys.argv[1:])
