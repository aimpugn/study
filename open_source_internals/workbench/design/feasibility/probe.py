"""탐색용 코드입니다(2026-10-13, 작업대 관측 재설계의 가능성 시험 P1~P11의 재현 경로). 사용자 지시에 따라 작업대 도구는 Go로 통일하므로, 같은 기록을 Go(GDB/MI)로 다시 만든 뒤 이 파일을 지웁니다.

가능성 시험용 gdb 스크립트: 한 실행 안에서 함수 멈춤, 시스템 콜, 공유 버퍼와 페이지, 메모리 컨텍스트를 함께 기록합니다.

사용: gdb -p <백엔드 PID> -batch -x probe.py
설정: PROBE_READY  중단점과 캐치포인트를 다 건 뒤 만들 표시 파일
출력: STOP / SYSCALL / SYSRET / BUF / PAGE / LP / BYTES / MCX / MCXN / CUR / SHM / MAP / OTHER / END 줄
"""

import os
import re

import gdb

READY = os.environ["PROBE_READY"]
PID = gdb.selected_inferior().pid
# ReadyForQuery는 결과를 보내는 pq_flush를 안에서 부르므로, 그 뒤 다음 명령을 기다리는 recvfrom 호출에서 끝냅니다.
seen_rfq = [False]


def fdpath(fd):
    """백엔드의 파일 디스크립터가 가리키는 경로를 /proc에서 읽어 반환합니다."""
    try:
        return os.readlink(f"/proc/{PID}/fd/{fd}")
    except OSError as err:
        return f"<{err.strerror}>"


def ev(expr):
    """현재 프레임에서 gdb 식을 평가해 정수로 반환합니다."""
    return int(gdb.parse_and_eval(expr))


def dump_buffer_and_page():
    """heap_prepare_pagescan 입구에서 현재 버퍼의 태그, 페이지 머리, 줄 포인터, 행 머리를 함수 호출 없이 읽어 출력합니다.

    Returns:
        (페이지 시작 주소, 줄 포인터 개수) 튜플.
    """
    buf = ev("((HeapScanDesc) sscan)->rs_cbuf")
    d = f"BufferDescriptors[{buf - 1}].bufferdesc"
    print(f"BUF buffer={buf} spcOid={ev(d + '.tag.spcOid')} dbOid={ev(d + '.tag.dbOid')} "
          f"relNumber={ev(d + '.tag.relNumber')} fork={ev(d + '.tag.forkNum')} block={ev(d + '.tag.blockNum')} "
          f"state=0x{ev(d + '.state.value'):08x}", flush=True)
    page = f"((PageHeader) (BufferBlocks + (Size) {buf - 1} * 8192))"
    lower, upper, special = ev(page + "->pd_lower"), ev(page + "->pd_upper"), ev(page + "->pd_special")
    lsn = (ev(page + "->pd_lsn.xlogid") << 32) | ev(page + "->pd_lsn.xrecoff")
    print(f"PAGE addr=0x{ev('(long) ' + page):x} lsn={lsn >> 32:X}/{lsn & 0xffffffff:X} "
          f"pd_lower={lower} pd_upper={upper} pd_special={special}", flush=True)
    nlp = (lower - 24) // 4
    for i in range(nlp):
        lp = f"{page}->pd_linp[{i}]"
        off, flags, length = ev(lp + ".lp_off"), ev(lp + ".lp_flags"), ev(lp + ".lp_len")
        tup = f"((HeapTupleHeader) ((char *) {page} + {off}))"
        print(f"LP {i + 1} lp_off={off} lp_flags={flags} lp_len={length} "
              f"t_xmin={ev(tup + '->t_choice.t_heap.t_xmin')} t_xmax={ev(tup + '->t_choice.t_heap.t_xmax')} "
              f"t_infomask=0x{ev(tup + '->t_infomask'):04x}", flush=True)
    return ev("(long) " + page), nlp


def dump_structure(page_addr, nlp):
    """메모리 구조 그림에 필요한 배치 자료를 출력합니다: 주소 공간, 공유 메모리의 주요 배열, 행 바이트, 메모리 컨텍스트 나무 전체.

    Args:
        page_addr: 현재 페이지의 시작 주소(정수).
        nlp: 페이지의 줄 포인터 개수.
    """
    # 주소 공간은 커널이 보여 주는 매핑 표 그대로 남기고, 구역 분류는 화면 쪽에서 합니다.
    with open(f"/proc/{PID}/maps", encoding="utf-8") as fh:
        for line in fh:
            print("MAP " + line.rstrip("\n"), flush=True)
    print(f"SHM BufferDescriptors=0x{ev('(long) BufferDescriptors'):x} BufferBlocks=0x{ev('(long) BufferBlocks'):x} "
          f"NBuffers={ev('NBuffers')} debug_query_string=0x{ev('(long) debug_query_string'):x} "
          f"CurrentMemoryContext=0x{ev('(long) CurrentMemoryContext'):x}", flush=True)

    # 행 바이트는 함수를 부르지 않고 프로세스 메모리를 그대로 읽습니다.
    inferior = gdb.selected_inferior()
    for i in range(nlp):
        lp = f"((PageHeader) {page_addr})->pd_linp[{i}]"
        off, length = ev(lp + ".lp_off"), ev(lp + ".lp_len")
        data = bytes(inferior.read_memory(page_addr + off, length))
        print(f"BYTES {i + 1} {data.hex()}", flush=True)

    # 나무 전체를 주소와 부모 주소로 남겨, 화면이 부모 관계와 크기로 그림을 그리게 합니다.
    stack = [(gdb.parse_and_eval("TopMemoryContext"), 0)]
    while stack:
        node, depth = stack.pop()
        while int(node) != 0:
            print(f"MCXN addr=0x{int(node):x} parent=0x{int(node['parent']):x} depth={depth} "
                  f"mem_allocated={int(node['mem_allocated'])} name={node['name'].string()}", flush=True)
            child = node["firstchild"]
            if int(child) != 0:
                stack.append((child, depth + 1))
            node = node["nextchild"]


def dump_memory_contexts():
    """TopMemoryContext부터 자식 목록을 따라가며 깊이 2까지의 이름과 mem_allocated, 전체 개수와 합을 출력합니다."""
    count, total = 0, 0
    stack = [(gdb.parse_and_eval("TopMemoryContext"), 0)]
    while stack:
        node, depth = stack.pop()
        while int(node) != 0:
            count += 1
            alloc = int(node["mem_allocated"])
            total += alloc
            if depth <= 1:
                print(f"MCX depth={depth} name={node['name'].string()} mem_allocated={alloc}", flush=True)
            child = node["firstchild"]
            if int(child) != 0:
                stack.append((child, depth + 1))
            node = node["nextchild"]
    print(f"MCX total_contexts={count} total_mem_allocated={total}", flush=True)
    cur = gdb.parse_and_eval("CurrentMemoryContext")
    chain = []
    while int(cur) != 0:
        chain.append(cur["name"].string())
        cur = cur["parent"]
    print("CUR " + " <- ".join(chain), flush=True)


class Stop(gdb.Breakpoint):
    """기록만 하고 계속하는 중단점입니다. ReadyForQuery에서는 멈춰 기록을 끝냅니다."""

    def stop(self):
        """멈춘 함수와 위치를 기록하고, heap_prepare_pagescan에서는 메모리 상태도 기록합니다.

        Returns:
            False(실행을 계속함).
        """
        frame = gdb.newest_frame()
        sal = frame.find_sal()
        print(f"STOP {frame.name()} {sal.symtab.filename if sal.symtab else '?'}:{sal.line}", flush=True)
        if self.location == "heap_prepare_pagescan":
            page_addr, nlp = dump_buffer_and_page()
            dump_memory_contexts()
            dump_structure(page_addr, nlp)
        if self.location == "ReadyForQuery":
            seen_rfq[0] = True
        return False


gdb.execute("set pagination off")
gdb.execute("set confirm off")
gdb.execute("handle SIGUSR1 SIGUSR2 SIGURG SIGALRM nostop noprint pass")
for loc in ["exec_simple_query", "heap_prepare_pagescan", "printtup", "ReadyForQuery"]:
    Stop(loc)
gdb.execute("catch syscall pread64 preadv pwrite64 openat sendto recvfrom epoll_pwait")
open(READY, "w").close()

# 캐치포인트의 멈춤 문구는 to_string으로 잡히지 않으므로(1차 시험에서 관측), 시스템 콜 번호($x8)와
# "입구를 본 뒤 아직 출구를 못 본 호출" 표로 입구와 출구를 가립니다. gdb도 스레드마다 같은 방식으로 번갈아 셉니다.
NAMES = {67: "pread64", 69: "preadv", 68: "pwrite64", 56: "openat", 206: "sendto", 207: "recvfrom", 22: "epoll_pwait"}
pending = {}
while True:
    gdb.execute("continue", to_string=False)
    try:
        frame = gdb.newest_frame()
    except gdb.error:
        print("OTHER 프로세스가 끝남", flush=True)
        break
    nr = ev("$x8")
    name = NAMES.get(nr)
    if name is None:
        print(f"OTHER x8={nr} at {frame.name()}", flush=True)
        continue
    if name not in pending:
        if name == "recvfrom" and seen_rfq[0]:
            print("SYS recvfrom (다음 명령 대기, 여기서 끝냄)", flush=True)
            break
        x0, x1, x2, x3 = (ev(f"$x{i}") for i in range(4))
        if name in ("pread64", "pwrite64", "preadv"):
            pending[name] = f"fd={x0}({fdpath(x0)}) count_or_iovcnt={x2} offset={x3}"
        elif name in ("sendto", "recvfrom"):
            pending[name] = f"fd={x0}({fdpath(x0)}) len={x2}"
        elif name == "openat":
            pending[name] = "path=" + gdb.parse_and_eval(f"(char *) {x1}").string()
        else:
            pending[name] = f"timeout_ms={x3}"
        print(f"SYSCALL {name} {pending[name]}", flush=True)
    else:
        print(f"SYSRET {name} -> {ev('(long) $x0')}", flush=True)
        del pending[name]

print("END", flush=True)
gdb.execute("detach")
