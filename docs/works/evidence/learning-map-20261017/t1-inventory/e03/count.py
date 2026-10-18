"""Re-count the facts behind the E-03 (Spring) decisions of the T1 structural-role table.

Usage: python3 -I count.py <path to open_source_internals/workbench/design/concept-9-spring.html>

The script reads the RUNS literal (line 129 of concept-9) and re-applies, in Python, the decoder rules that
produced each field (decode/spring.go, spring_jfr.go, spring_tree.go, spring_compare.go). Every check prints
"<id> ok/total description". A check passes when ok == total; the last line says how many failed and the exit
code is 1 if any did. The ids are cited in ../../t1-e03-evidence.md. Nothing here is imported from the
workbench. Most rules are the Go code rewritten step by step (the docstring names the lines); two are not
(C15k and C16f): they are different rules that give the same value on these records, and the importer
uses them as recompute rules that stop the import when they do not hold.
"""
import collections
import json
import re
import sys

RESULTS = []
INFO = []


def info(cid, desc, value):
    """Record a count that is reported, not judged (for example how many copies a recompute replaces)."""
    INFO.append((cid, desc, value))


def check(cid, desc, ok, total, note=""):
    """Record one check. ok and total are counts over the 14 runs (or over the items named in desc)."""
    RESULTS.append((cid, desc, ok, total, note))


def load(path):
    """Return the RUNS object of concept-9 (key "5-<scene>/<run>" -> decoded record)."""
    text = open(path, encoding="utf-8").read()
    m = re.search(r"(?m)^\s*const RUNS = (.*);\s*$", text)
    return json.loads(m.group(1))


def ref_split(ref):
    """Split a JFR frame text "pkg.Class::method:line" into (method key, line)."""
    i = ref.rindex(":")
    return ref[:i], int(ref[i + 1:])


# ---------------------------------------------------------------- decoder rules re-written in Python

def split_requests(msgs):
    """decode/spring.go 258-308 splitRequests over the mac-side view of one connection.

    Returns the answered requests and, per request, the indexes of the messages that belong to it
    (F messages of its batch, B messages up to and including its ReadyForQuery).
    """
    reqs = [{"seq": 0, "sql": ["(접속 인증)"], "types": [], "sent": 0, "members": []}]
    cur = None
    for i, m in enumerate(msgs):
        if m["dir"] != "F" or m["type"] == "X":
            continue
        if m["type"] in ("ssl", "startup"):
            if reqs[0]["sent"] == 0:
                reqs[0]["sent"] = m["t"]
            reqs[0]["members"].append(i)
            continue
        if cur is None:
            cur = {"seq": len(reqs), "sent": m["t"], "types": [], "sql": None, "stmt": "", "members": []}
        cur["members"].append(i)
        cur["types"].append(m["type"])
        if m["type"] in ("P", "B") and cur["stmt"] == "":
            cur["stmt"] = m.get("name", "")
        if m["type"] in ("Q", "P"):
            cur["sql"] = (cur["sql"] or []) + [m.get("text", "")]
        if m["type"] in ("S", "Q"):
            reqs.append(cur)
            cur = None
    k = 0
    for i, m in enumerate(msgs):
        if m["dir"] != "B" or k >= len(reqs):
            continue
        reqs[k]["members"].append(i)
        if m["type"] == "C":
            reqs[k].setdefault("tags", []).append(m.get("text", ""))
        elif m["type"] == "E":
            f = m.get("f") or {}
            reqs[k]["error"] = {"code": f.get("code", ""), "message": m.get("text", ""), "where": f.get("where", "")}
        elif m["type"] == "Z":
            reqs[k]["done"], reqs[k]["status"] = m["t"], m.get("text", "")
            k += 1
    return reqs[:k]


def derive_txs(conn, reqs, msgs):
    """decode/spring.go 311-338 deriveTxs, plus the xid and clog copies of 537-553."""
    out, cur = [], None
    for r in reqs:
        if r["seq"] == 0:
            continue
        if cur is None and r["status"] != "I":
            cur = {"conn": conn, "first": r["seq"], "t0": r["sent"], "xids": [], "clog": [], "end": ""}
        if cur is not None and r["status"] == "I":
            cur["last"], cur["t1"] = r["seq"], r["done"]
            if r.get("tags"):
                cur["end"] = r["tags"][-1]
            for m in msgs:
                if m["dir"] == "F" and m["t"] > r["done"]:
                    cur["next"] = m
                    break
            out.append(cur)
            cur = None
    for tx in out:
        for r in reqs[tx["first"]:tx["last"] + 1]:
            if not r.get("server"):
                continue
            level = "1"
            for e in r["server"]["events"]:
                if e["kind"] == "assign_xid":
                    level = e["f"]["level"]
                elif e["kind"] == "xid":
                    tx["xids"].append({"xid": e["f"]["xid"], "level": level, "t": e["t"]})
                elif e["kind"] in ("commit_clog", "abort_clog"):
                    tx["clog"].append({"kind": e["kind"], "xid": e["f"]["xid"], "nsub": e["f"]["nsub"], "t": e["t"]})
    return out


def lane_of(fn):
    """decode/spring_tree.go 265-279 laneOf."""
    for prefix, lane in (("lab.", "app"), ("org.springframework.jdbc.core.", "jdbc"), ("org.springframework.", "tx"),
                         ("com.zaxxer.hikari.", "hikari"), ("org.postgresql.", "pg")):
        if fn.startswith(prefix):
            return lane
    return "other"


def last(n):
    """decode/spring_tree.go 78-83 callNode.last."""
    return n["t1"] if n["t1"] is not None else n["seen"]


def binding_replay(tree, borrows, start):
    """decode/spring_tree.go 403-465 bindingReplay. Returns (states, slots, effects by node id)."""
    steps = []
    for b in borrows:
        if b["thread"] == "main" and b["result"] == "ok" and b["conn"] != "" and b["t0"] >= start:
            steps.append((b["t1"], "bind", b["conn"], -1))
    for n in tree:
        for suffix, kind in (("DataSourceTransactionManager::doSuspend", "suspend"),
                             ("DataSourceTransactionManager::doResume", "resume"),
                             ("DataSourceTransactionManager::doCleanupAfterCompletion", "unbind"),
                             ("ProxyConnection::close", "close")):
            if n["fn"].endswith(suffix):
                steps.append((n["t0"], kind, "", n["id"]))
                break
    steps.sort(key=lambda s: s[0])  # Python's sort is stable, like sort.SliceStable
    states = [{"t": start, "bound": "", "held": [], "by": -1}]
    slots, open_, effects = [], {}, collections.defaultdict(list)
    bound, released, held = "", "", []
    for t, kind, conn, node in steps:
        if kind == "bind":
            bound = conn
            open_[conn] = len(slots)
            slots.append({"conn": conn, "t0": t, "t1": 0})
        elif kind == "suspend":
            held.append(bound)
            effects[node].append({"kind": "suspend", "conn": bound})
            bound = ""
        elif kind == "resume":
            if held:
                bound, held = held[-1], held[:-1]
            effects[node].append({"kind": "resume", "conn": bound})
        elif kind == "unbind":
            released, bound = bound, ""
        elif kind == "close":
            effects[node].append({"kind": "return", "conn": released})
            if released in open_:
                slots[open_.pop(released)]["t1"] = t
            released = ""
            continue
        states.append({"t": t, "bound": bound, "held": list(held), "by": node})
    return states, slots, effects


def server_waits(conns):
    """decode/spring_tree.go 478-500 serverWaits."""
    out = []
    for c in conns:
        for r in c["requests"]:
            if not r.get("server"):
                continue
            w = None
            for e in r["server"]["events"]:
                if e["kind"] == "xact_wait" and w is None:
                    w = {"conn": c["name"], "seq": r["seq"], "t0": e["t"], "t1": r["done"], "xid": e["f"]["xid"], "end": "done"}
                elif e["kind"] == "ereport" and w is not None and w["end"] == "done":
                    w["t1"], w["end"] = e["t"], "error"
            if w is not None:
                out.append(w)
    return out


def wait_cycles(tree, waits, binding, conns, slots, pool):
    """decode/spring_tree.go 532-626 waitCycles + poolWaitCycles (keys as the JSON writes them, zero values omitted)."""
    out = []
    for w in waits:
        c = {"kind": "server", "t0": w["t0"], "t1": w["t1"], "waiter": w["conn"], "xid": w["xid"], "holder": "",
             "thread": "", "call": -1, "heldBy": "", "cycle": False}
        for cn in conns:
            for r in cn["requests"]:
                for e in (r.get("server") or {}).get("events", []):
                    if e["kind"] == "xid" and e["f"]["xid"] == w["xid"]:
                        c["holder"] = cn["name"]
        for n in tree:
            if n["t0"] > w["t0"] or n["t1"] is None or n["t1"] < w["t1"]:
                continue
            for e in n["effects"]:
                if e["kind"] == "wire" and e["conn"] == w["conn"] and w["seq"] in e["seq"]:
                    c["call"], c["thread"] = n["id"], "main"
        st = None
        for b in binding:
            if b["t"] <= w["t0"]:
                st = b
        if st is not None and c["holder"] and c["holder"] in st["held"]:
            c["heldBy"] = "main"
        c["cycle"] = c["call"] >= 0 and c["heldBy"] != "" and c["heldBy"] == c["thread"]
        out.append(c)
    for n in tree:
        if not any(e["kind"] == "borrow" and e.get("result") == "timeout" for e in n["effects"]) or n["t1"] is None:
            continue
        c = {"kind": "pool", "pool": pool, "t0": n["t0"], "t1": n["t1"], "waiter": "", "xid": "", "holder": "",
             "thread": "main", "call": n["id"], "heldBy": "", "cycle": False}
        in_use = [s["conn"] for s in slots if s["t0"] <= n["t0"] and (s["t1"] == 0 or s["t1"] > n["t0"])]
        st = None
        for b in binding:
            if b["t"] <= n["t0"]:
                st = b
        c["holder"] = ", ".join(in_use)
        all_held = bool(in_use) and st is not None and all(x in st["held"] for x in in_use)
        if all_held:
            c["heldBy"] = "main"
        c["cycle"] = pool > 0 and len(in_use) >= pool and all_held
        out.append(c)
    # omitempty on waitCycle: pool(int). The other fields are always written.
    for c in out:
        if not c.get("pool"):
            c.pop("pool", None)
    return out


def alt_index(node, nodes, alts):
    """decode/spring_compare.go 82-111: the index of the alt whose signature is this node's. None if not exactly one."""
    seq = ["lab.*" if nodes[k]["fn"].startswith("lab.") else nodes[k]["fn"] + nodes[k]["desc"] for k in node["kids"]]
    end = "exception" if node["end"] != "return" else "return"
    last_exc = bool(node["kids"]) and nodes[node["kids"][-1]]["end"] != "return"
    hits = [i for i, a in enumerate(alts) if a["seq"] == seq and a["end"] == end and a["lastExc"] == last_exc]
    return hits[0] if len(hits) == 1 else None


# ---------------------------------------------------------------- checks

def main(path):
    """Run every check over the RUNS literal of the file at path, print the table, and return the exit code.

    Args:
        path: concept-9-spring.html.

    Returns:
        0 when every judged check has ok == total, otherwise 1.

    """
    runs = load(path)
    check("R0", "concept-9 RUNS 기록 수(14)", len(runs), 14)
    acc = collections.Counter()
    tot = collections.Counter()

    def count(cid, ok):
        tot[cid] += 1
        acc[cid] += bool(ok)

    zero_times = 0
    time_values = 0
    copies = collections.Counter()
    for key in sorted(runs):
        r = runs[key]
        conns, view, java = r["conns"], r["view"], r["java"]
        caps, cps, blocked = r["captures"], r["checkpoints"], r["blocked"]

        # C1 관찰 ↔ 멈춤 자리·막힘 관찰: tCapture가 관찰 t와 같고 afterStep이 n·after와 같은 짝이 하나뿐
        for c in cps:
            m = [x for x in caps if x["t"] == c["tCapture"]]
            count("C1a", len(m) == 1 and m[0]["afterStep"] == c["n"] and m[0]["label"] == c["label"])
        for b in blocked:
            m = [x for x in caps if x["t"] == b["tCapture"]]
            count("C1b", len(m) == 1 and m[0]["afterStep"] == b["after"] and m[0]["label"] == "blocked")
            count("C1c", sum(1 for c in cps if c["n"] == b["after"]) == 1)
        paired = {c["tCapture"] for c in cps} | {b["tCapture"] for b in blocked}
        for x in caps:
            count("C1d", x["t"] in paired)

        # C2 멈춤 구간 = 멈춤 자리의 [t, tGo], 모든 자리에서 tGo > t
        count("C2", view["pauses"] == [[c["t"], c["tGo"]] for c in cps if c["tGo"] > c["t"]] and all(c["tGo"] > c["t"] for c in cps))
        for c in cps:
            count("C2b", c["tGo"] > c["t"])
        # C3 장면 구간 = [start의 tGo, end의 t]
        frm = [c["tGo"] for c in cps if c["label"] == "start"][-1]
        to = [c["t"] for c in cps if c["label"] == "end"][-1]
        count("C3", view["window"] == [frm, to])

        # C5 관찰 span[0] = t, span[1] ≥ span[0]
        for k, x in enumerate(caps):
            count("C5", x["span"][0] == x["t"] and x["span"][1] >= x["span"][0])
            for part, (a, b) in x["reads"].items():
                count("C6", a <= b)
                if a == b:
                    count("C6b-last", k == len(caps) - 1)
                copies["읽기 구간 가운데 끝이 관찰 span의 끝보다 늦은 것"] += b > x["span"][1]
                copies["읽기 구간"] += 1

        # C7 겹친 메시지 = 관찰 span 안의 tc를 가진 메시지(모든 연결, tc 차례), 메시지 사본과 하나씩 같음
        allc = sorted([m for c in conns for m in c["messages"]], key=lambda m: m["tc"])
        for x in caps:
            want = [m for m in allc if x["span"][0] <= m["tc"] <= x["span"][1]]
            count("C7", x["overlap"] == want)
            for o in x["overlap"]:
                count("C7b", sum(1 for m in allc if m == o) == 1)

        # C8 두 시계: 연결마다 같은 순번으로 잘렸고(same) 구간이 비지 않음, ms → ns가 정확
        for w in r["clock"]["wire"]:
            count("C8", w["same"] and w["lo"] <= w["hi"] and round(w["lo"] * 1e6) / 1e6 == w["lo"] and round(w["hi"] * 1e6) / 1e6 == w["hi"])
        # C8b clock.wire = 저장된 메시지의 (tc − t)로 다시 셈(compareClocks, spring.go 204~220): B의 최댓값이 lo, F의 최솟값이 hi
        ns = lambda us: round(us * 1000)
        for w in r["clock"]["wire"]:
            c = [x for x in conns if x["conn"] == w["conn"]][0]
            d = [(m["dir"], ns(m["tc"]) - ns(m["t"])) for m in c["messages"]]
            lo = max(v for k, v in d if k == "B") / 1e6
            hi = min(v for k, v in d if k != "B") / 1e6
            count("C8b", w["n"] == len(c["messages"]) and w["lo"] == lo and w["hi"] == hi)

        # C9 연결: 이름 = "C" + 번호, pid = BackendKeyData의 pid, sessions = {pid: 이름}, 마지막 메시지 종류
        sess = {}
        for c in conns:
            ks = [m for m in c["messages"] if m["dir"] == "B" and m["type"] == "K"]
            count("C9a", c["name"] == "C%d" % c["conn"])
            count("C9b", len(ks) == 1 and ks[0]["f"]["pid"] == str(c["pid"]))
            count("C9c", c["lastMac"] == c["messages"][-1]["type"] and c["lastCont"] == c["lastMac"])
            count("C9d", all(m["t"] != 0 and m["tc"] != 0 for m in c["messages"]))
            # C9f 연결 위 메시지의 f.pid는 BackendKeyData(B·K)에만 있고 그 연결의 백엔드 pid(장소 식별자)
            for m in c["messages"]:
                if "pid" in (m.get("f") or {}):
                    count("C9f", m["dir"] == "B" and m["type"] == "K" and m["f"]["pid"] == str(c["pid"]))
                elif m["type"] == "K":
                    count("C9f", False)
            sess[str(c["pid"])] = c["name"]
        count("C9e", r["sessions"] == sess)

        # C10 요청 = splitRequests(메시지): seq·sent·done·types·sql·stmt·status·tags·error가 모두 같고 seq = 차례
        for c in conns:
            got = split_requests(c["messages"])
            count("C10a", len(got) == len(c["requests"]))
            members = [i for q in got for i in q["members"]]
            count("C10b", len(members) == len(set(members)))
            for i, (q, g) in enumerate(zip(c["requests"], got)):
                same = (q["seq"] == i == g["seq"] and q["sent"] == g["sent"] and q["done"] == g["done"] and q["types"] == g["types"]
                        and q["sql"] == g["sql"] and q.get("stmt", "") == g.get("stmt", "") and q["status"] == g["status"]
                        and q.get("tags") == g.get("tags") and q.get("error") == g.get("error"))
                count("C10c", same)
                ms = [c["messages"][j] for j in g["members"]]
                count("C10d", ms[0]["t"] == q["sent"] and ms[-1]["type"] == "Z" and ms[-1]["t"] == q["done"])
                q["_members"] = g["members"]
            # C10f 앞 요청의 답이 오기 전에 보낸 요청(pgjdbc가 묶음 둘을 한 번에 보냄, spring.go 252~253)은 보낸 시각이 앞 요청과 같음.
            # 같으면 구간이 포개지고, 다르면(중계기가 두 묶음을 따로 읽은 기록) 같은 연결 장소에서 어긋납니다.
            rq = c["requests"]
            for a, b in zip(rq, rq[1:]):
                if b["sent"] < a["done"]:
                    count("C10f", b["sent"] == a["sent"])
            free = [m["dir"] + m["type"] for j, m in enumerate(c["messages"]) if j not in set(members)]
            tot["C10e-free"] += len(free)
            acc["C10e-free"] += sum(1 for x in free if x == "FX")

        # C11 서버 창: seq ≥ 1에만 있음, sql = 요청 sql을 "; "로 이은 것, in·out은 null, 창은 이어 붙음
        for c in conns:
            count("C11a", c["join"][0] == c["join"][1])
            prev = None
            for q in c["requests"]:
                s = q.get("server")
                count("C11b", (s is None) == (q["seq"] == 0))
                if s is None:
                    continue
                count("C11c", s["sql"] == "; ".join(q["sql"] or []))
                count("C11d", s["in"] is None and s["out"] is None and s["t0"] <= s["t1"])
                # 창이 그 요청의 첫 F가 컨테이너 쪽에 닿기 전에 시작하는지(같은 컨테이너 시계): 창에 백엔드가 앱을 기다린 시간이 듦
                first_f = min(c["messages"][j]["tc"] for j in q["_members"] if c["messages"][j]["dir"] == "F")
                copies["서버 창 가운데 첫 F가 컨테이너 쪽에 닿기 전에 시작하는 것"] += s["t0"] < first_f
                copies["서버 창"] += 1
                if prev is not None:
                    count("C11e", prev["t1"] == s["t0"])
                prev = s

        # C12 트랜잭션 = deriveTxs(요청): first·last·t0·t1·end·next, xids·clog = 서버 사건의 사본
        mine = []
        for c in conns:
            mine += derive_txs(c["conn"], c["requests"], c["messages"])
        count("C12a", len(mine) == len(r["transactions"]))
        for a, b in zip(r["transactions"], mine):
            count("C12b", all(a[k] == b[k] for k in ("conn", "first", "last", "t0", "t1", "end")) and a.get("next") == b.get("next"))
            count("C12c", a["xids"] == b["xids"] and a["clog"] == b["clog"])
            cn = [c for c in conns if c["conn"] == a["conn"]][0]
            count("C12d", a["t0"] == cn["requests"][a["first"]]["sent"] and a["t1"] == cn["requests"][a["last"]]["done"])
            copies["xids+clog"] += len(a["xids"]) + len(a["clog"])

        # C13 연결을 연 스레드 = 첫 Q를 감싼 QueryExecutorImpl.execute 구간의 스레드(spans 차례에서 처음)
        for c in conns:
            qs = [m for m in c["messages"] if m["dir"] == "F" and m["type"] == "Q"]
            th = ""
            if qs:
                for s in java["spans"]:
                    if s["fn"] == "org.postgresql.core.v3.QueryExecutorImpl::execute" and s["t0"] <= qs[0]["t"] <= s["t1"]:
                        th = s["thread"]
                        break
            count("C13", th == c["thread"])
        # 트리 마디가 없는 JFR 구간(장면 밖이거나 main이 아닌 스레드): caller가 그 호출 사슬의 유일한 기록
        keyed = {(n["fn"], n["desc"], n["t0"], n["t1"]) for n in view["tree"] if n["t1"] is not None}
        bare = [s for s in java["spans"] if not (s["thread"] == "main" and (s["fn"], s["desc"], s["t0"], s["t1"]) in keyed)]
        copies["트리 마디가 없는 JFR 구간"] += len(bare)
        w0, w1 = view["window"]
        copies["트리 마디가 없는 JFR 구간 가운데 장면 구간 밖의 main 스레드 구간"] += sum(1 for s in bare if s["thread"] == "main" and (s["t1"] < w0 or s["t0"] > w1))
        copies["트리 마디가 없는 JFR 구간 가운데 main이 아닌 스레드 구간"] += sum(1 for s in bare if s["thread"] != "main")
        copies["JFR 구간"] += len(java["spans"])
        copies["caller가 빈 JFR 구간"] += sum(1 for s in java["spans"] if s["caller"] == "")
        # C13b JFR 구간은 (t0 오름, t1 내림)으로 정렬돼 있음(buildSpringJava 326-331): id 차례 = spans 차례
        sp = java["spans"]
        count("C13b", all((sp[i]["t0"], -sp[i]["t1"]) <= (sp[i + 1]["t0"], -sp[i + 1]["t1"]) for i in range(len(sp) - 1)))
        # C27 depth = 스레드마다 spans 차례로 보며 시작보다 먼저 끝난 열린 호출을 닫고 남은 열린 호출의 수(spring_jfr.go 325-341)
        open_ = collections.defaultdict(list)
        for s in sp:
            st = open_[s["thread"]]
            while st and st[-1] < s["t0"]:
                st.pop()
            count("C27", s["depth"] == len(st))
            st.append(s["t1"])

        # C14 서버 사건의 f.at = 스택(안쪽부터) 1번 칸의 위치
        for c in conns:
            for q in c["requests"]:
                for e in (q.get("server") or {}).get("events", []):
                    if "at" in (e.get("f") or {}):
                        st = e.get("stack") or []
                        count("C14", len(st) >= 2 and st[-2]["at"] == e["f"]["at"])

        # C15 호출 트리
        tree = view["tree"]
        byid = {n["id"]: n for n in tree}
        count("C15a", [n["id"] for n in tree] == list(range(len(tree))))
        for n in tree:
            kids = sorted([m["id"] for m in tree if m["parent"] == n["id"]], key=lambda i: byid[i]["t0"])
            count("C15b", n["kids"] == kids)
            d, p = 0, n["parent"]
            while p >= 0:
                d, p = d + 1, byid[p]["parent"]
            count("C15c", n["depth"] == d)
            count("C15d", n["lane"] == lane_of(n["fn"]))
            count("C15e", (n["t1"] is None) == (n["end"] == "passed") == ("seen" in n))
            if n["parent"] < 0:
                count("C15f-root-site0", n["site"] == 0)
            if n["fn"].startswith("lab."):
                count("C15g-lab-nobranch", "branch" not in n)
            else:
                count("C15h-key", n.get("branch", {}).get("key") == n["fn"] + n["desc"])
                count("C15i-alt", alt_index(n, byid, view["branches"][n["fn"] + n["desc"]]["alts"]) == n["branch"]["alt"])
        roots = sorted([n["id"] for n in tree if n["parent"] < 0], key=lambda i: byid[i]["t0"])
        count("C15j", view["roots"] == roots)
        # 스택에만 보인 마디: t0 = 기록된 자손의 t0 최솟값, seen = 기록된 자손의 t1 최댓값. 해석기는 스택을 걸으며 그 마디를 지난
        # 구간마다 min·max를 갱신하므로(spring_tree.go 206~214) Go를 옮긴 것이 아니라 같은 값을 내는 다른 규칙입니다.
        for n in tree:
            if n["t1"] is not None:
                continue
            desc_, stack = [], list(n["kids"])
            while stack:
                m = byid[stack.pop()]
                desc_.append(m)
                stack += m["kids"]
            rec = [m for m in desc_ if m["t1"] is not None]
            count("C15k", bool(rec) and n["t0"] == min(m["t0"] for m in rec) and n["seen"] == max(m["t1"] for m in rec))
        # 기록된 마디마다 같은 (fn, desc, t0, t1)의 main 스레드 JFR 구간이 하나 이상(사본 여부만 봄, 짝은 만들지 않음)
        for n in tree:
            if n["t1"] is not None:
                count("C15l", any(s["thread"] == "main" and s["fn"] == n["fn"] and s["desc"] == n["desc"] and s["t0"] == n["t0"] and s["t1"] == n["t1"] for s in sp))
        count("C15m", view["ambiguous"] == 0)

        # C16 효과: 차례(묶임 효과 → 종류 효과 → 잠듦 → 예외)와 종류마다 사본 규칙
        order = {"suspend": 0, "resume": 0, "return": 0, "wire": 1, "borrow": 1, "autocommit": 1, "rollbackOnly": 1, "validate": 1, "park": 2, "throw": 3}
        for n in tree:
            ranks = [order[e["kind"]] for e in n["effects"]]
            count("C16a", ranks == sorted(ranks))
            parent_fn = byid[n["parent"]]["fn"] if n["parent"] >= 0 else ""
            wires = []
            if n["fn"] == "org.postgresql.core.v3.QueryExecutorImpl::execute" and parent_fn != n["fn"]:
                for c in conns:
                    seq = [q["seq"] for q in c["requests"] if q["seq"] > 0 and n["t0"] <= q["sent"] <= last(n)]
                    if seq:
                        wires.append({"kind": "wire", "conn": c["name"], "seq": seq})
            count("C16b-wire", [e for e in n["effects"] if e["kind"] == "wire"] == wires)
            if n["fn"] == "com.zaxxer.hikari.util.ConcurrentBag::borrow":
                want = []
                for b in r["borrows"]:
                    if b["thread"] == "main" and b["t0"] <= n["t0"] and last(n) <= b["t1"]:
                        e = {"kind": "borrow", "conn": b["conn"], "result": b["result"], "idleMs": b["idleMs"], "validated": b["validated"]}
                        want.append({k: v for k, v in e.items() if v not in ("", 0, False) or k == "kind"})
                got = [e for e in n["effects"] if e["kind"] == "borrow"]
                count("C16c-borrow", got == want and len(got) == 1)
            if n["fn"] == "org.postgresql.jdbc.PgConnection::setAutoCommit":
                sent = sum(1 for c in conns for q in c["requests"] if q["seq"] > 0 and n["t0"] <= q["sent"] <= last(n))
                got = [e for e in n["effects"] if e["kind"] == "autocommit"]
                count("C16d-autocommit", len(got) == 1 and sent == 0 and "sent" not in got[0] and got[0].get("value") in ("true", "false"))
            for e in n["effects"]:
                if e["kind"] == "park":
                    m = [p for p in java["parks"] if p["thread"] == "main" and p["t0"] == e["t0"] and p["t1"] == e["t1"]]
                    count("C16e-park", len(m) == 1 and m[0]["under"] == e["at"] and ref_split(e["at"])[0] == n["fn"])
                # 예외 효과의 at: 해석기는 예외 스택 전체에서 첫 추적 대상 칸을 고르지만(spring_tree.go 830~845·903~904), 그 칸은
                # 마디와 메서드가 같은 첫 칸이고 at[](앞 8칸) 안에 있으면 이 규칙으로 같은 값이 나옵니다(다른 규칙, 같은 값).
                if e["kind"] == "throw":
                    m = [t for t in java["throws"] if t["thread"] == "main" and t["t"] == e["t0"]]
                    firsts = [a for a in (m[0]["at"] if len(m) == 1 else []) if ref_split(a)[0] == n["fn"]]
                    count("C16f-throw", len(m) == 1 and m[0]["class"] == e["class"] and m[0]["message"] == e.get("message", "")
                          and bool(firsts) and firsts[0] == e["at"])
                if e["kind"] in ("rollbackOnly", "validate"):
                    count("C16g-kindonly", list(e) == ["kind"])
        # 예외·잠듦 효과는 main 스레드의 예외·잠듦마다 하나씩
        n_thr = sum(1 for n in tree for e in n["effects"] if e["kind"] == "throw")
        count("C16h", n_thr == sum(1 for t in java["throws"] if t["thread"] == "main"))
        n_park = sum(1 for n in tree for e in n["effects"] if e["kind"] == "park")
        count("C16i", n_park == sum(1 for p in java["parks"] if p["thread"] == "main"))

        # C17 연결 묶임·슬롯·묶임 효과 = bindingReplay(빌림, 트리, 장면 시작)
        states, slots, beffects = binding_replay(tree, r["borrows"], view["window"][0])
        count("C17a-binding", view["binding"] == states)
        count("C17b-slots", view["slots"] == slots)
        for n in tree:
            got = [e for e in n["effects"] if e["kind"] in ("suspend", "resume", "return")]
            count("C17c-beffects", got == beffects.get(n["id"], []))
        count("C17d-state0", view["binding"][0] == {"t": view["window"][0], "bound": "", "held": [], "by": -1})
        copies["slot t1 = 0(돌려주지 않음)"] += sum(1 for s in slots if s["t1"] == 0)
        copies["slots"] += len(slots)
        # held를 저장한 상태만으로 다시 셈: by 마디가 doSuspend면 앞 상태의 bound를 쌓고, doResume이면 뺌
        held = []
        for i, s in enumerate(view["binding"][1:], 1):
            fn = byid[s["by"]]["fn"] if s["by"] >= 0 else ""
            if fn.endswith("::doSuspend"):
                held = held + [view["binding"][i - 1]["bound"]]
            elif fn.endswith("::doResume") and held:
                held = held[:-1]
            count("C17f-held", s["held"] == held)

        # C18 서버 기다림 = serverWaits(요청), 끝은 모두 오류(ereport, 컨테이너 시계)
        waits = server_waits(conns)
        count("C18a", view["waits"] == waits)
        for w in view["waits"]:
            count("C18b-error", w["end"] == "error")

        # C19 기다림 고리 = waitCycles + poolWaitCycles
        count("C19", view["waitCycles"] == wait_cycles(tree, view["waits"], view["binding"], conns, view["slots"], r["pool"]))

        # C20 빌림: conn이 빈 4건은 모두 시간 초과, idleMs 0, 검사 안 함. 빌림의 시각은 getConnection(long) JFR 구간과 같음
        for b in r["borrows"]:
            if b["conn"] == "":
                count("C20a", b["result"] == "timeout" and b["idleMs"] == 0 and b["validated"] is False)
            count("C20b", any(s["fn"] == "com.zaxxer.hikari.pool.HikariPool::getConnection" and s["desc"].startswith("(J)")
                              and s["thread"] == b["thread"] and s["t0"] == b["t0"] and s["t1"] == b["t1"] for s in sp))

        # C21 막힘 관찰의 waiters는 연결 이름(sessions의 값) 하나 이하
        for b in blocked:
            count("C21", len(b["waiters"]) <= 1 and all(x in sess.values() for x in b["waiters"]))

        # C22 JFR 잠듦의 under는 at[](앞 6칸) 밖
        for p in java["parks"]:
            count("C22", p["under"] not in p["at"])

        # C24 시각 값 0이 없음(rel(0) = 0이 "값 없음"과 섞이는 함정, spring.go 408-411)
        def times(v, path):
            nonlocal zero_times, time_values
            if isinstance(v, dict):
                for k, x in v.items():
                    times(x, path + "." + k)
            elif isinstance(v, list):
                for x in v:
                    times(x, path + "[]")
            elif isinstance(v, (int, float)) and not isinstance(v, bool) and re.search(r"\.(t|t0|t1|tc|tCapture|tGo|sent|done|accept|close|seen)$|\.(span|window|pauses\[\])\[\]$|\.reads\.\w+\[\]$", path):
                time_values += 1
                zero_times += v == 0
        times(r, "$")

    for cid in sorted(tot):
        check(cid, "", acc[cid], tot[cid])
    check("C24", "시각 칸(t, t0, t1, tc, tCapture, tGo, sent, done, accept, close, seen, span, window, pauses, reads)의 값이 0이 아님",
          time_values - zero_times, time_values)

    # C23 갈림 모음과 compared가 14벌에서 같은 값
    first = runs[sorted(runs)[0]]["view"]
    same = sum(1 for k in runs if runs[k]["view"]["branches"] == first["branches"] and runs[k]["view"]["compared"] == first["compared"])
    check("C23", "view.branches·view.compared가 14벌 모두 같은 값", same, len(runs), "compared = %s" % (sorted(first["compared"]) == sorted(runs)))

    # C25 JFR 프레임 글: 줄 −1 수, 줄 0 수
    refs = []
    for r in runs.values():
        j = r["java"]
        refs += [s["caller"] for s in j["spans"] if s["caller"]]
        refs += [a for t in j["throws"] for a in t["at"]] + [a for p in j["parks"] for a in p["at"]] + [p["under"] for p in j["parks"]]
        refs += [h for n in r["view"]["tree"] for h in n["hidden"]] + [e["at"] for n in r["view"]["tree"] for e in n["effects"] if "at" in e]
    lines = collections.Counter(ref_split(x)[1] for x in refs)
    check("C25a", "JFR 프레임 글 가운데 줄이 −1인 것(나머지는 양수)", sum(v for k, v in lines.items() if k == -1),
          sum(v for k, v in lines.items() if k <= 0), "줄 0: %d, 전체 %d" % (lines.get(0, 0), len(refs)))
    # 오버로드: 이름이 같고 서술자가 다른 메서드(구간·트리 마디에서)와, 그 메서드를 가리키는 서술자 없는 프레임 글
    descs = collections.defaultdict(set)
    for r in runs.values():
        for s in r["java"]["spans"]:
            descs[s["fn"]].add(s["desc"])
        for n in r["view"]["tree"]:
            descs[n["fn"]].add(n["desc"])
    over = sorted(f for f, ds in descs.items() if len(ds) > 1)
    thr_over = sum(1 for r in runs.values() for t in r["java"]["throws"] for a in t["at"] if ref_split(a)[0] in over)
    check("C25b", "오버로드 메서드(이름 하나에 서술자 둘 이상) 수 / 그 메서드를 가리키는 예외 프레임 칸", len(over), len(over),
          "%s; 예외 at[] 칸 %d" % (", ".join(over), thr_over))

    # C26 위치를 남기지 않은 PostgreSQL 함수(줄 모름 "?" 스택 칸, 줄 판정 안의 부른 함수)마다, 같은 백엔드의 스택 칸에서
    # 그 이름이 나온 경로의 수. 트랜잭션 실험 적재기의 callee 규칙(0개면 위치 없는 함수, 2개 이상이면 멈춤)과 같습니다.
    kinds = collections.Counter()
    for r in runs.values():
        for c in r["conns"]:
            paths = collections.defaultdict(set)
            wins = [q["server"] for q in c["requests"] if q.get("server")]
            for s in wins:
                for lst in ("events", "snapshots", "checks"):
                    for e in s[lst]:
                        for f in e.get("stack") or []:
                            if f["at"] != "?":
                                paths[f["fn"]].add(f["at"].rsplit(":", 1)[0])
            for s in wins:
                for lst in ("events", "snapshots", "checks"):
                    for e in s[lst]:
                        for f in e.get("stack") or []:
                            if f["at"] == "?":
                                kinds["?칸 경로 %d개" % min(len(paths[f["fn"]]), 2)] += 1
                for ch in s["checks"]:
                    calls = list(ch["lines"]) + list(ch["path"] or []) + [i for p in ch["path"] or [] for i in p.get("inner", [])]
                    for f in calls:
                        kinds["판정 안 호출 경로 %d개" % min(len(paths[f["fn"]]), 2)] += 1
    many = sum(v for k, v in kinds.items() if k.endswith("2개"))
    check("C26", "위치 없는 PostgreSQL 함수 가운데 같은 백엔드에서 경로가 둘 이상인 것이 없음", sum(kinds.values()) - many, sum(kinds.values()),
          ", ".join("%s %d" % kv for kv in sorted(kinds.items())))
    # 기록된 트리 마디와 JFR 구간: 같은 (fn, desc, t0, t1, main)의 구간이 정확히 하나인 마디 수(짝을 만들지 않는 근거의 크기)
    uniq = sum(1 for r in runs.values() for n in r["view"]["tree"] if n["t1"] is not None and
               sum(1 for s in r["java"]["spans"] if s["thread"] == "main" and (s["fn"], s["desc"], s["t0"], s["t1"]) == (n["fn"], n["desc"], n["t0"], n["t1"])) == 1)
    info("I", "기록된 트리 마디 가운데 값이 같은 JFR 구간이 정확히 하나인 것", uniq)
    # JFR 구간의 depth(같은 스레드에서 감싼 기록된 호출 수, spring_jfr.go)가 짝인 트리 마디의 깊이와 다른 수: X5인 까닭의 크기
    differ = sum(1 for r in runs.values() for n in r["view"]["tree"] if n["t1"] is not None for s in r["java"]["spans"]
                 if s["thread"] == "main" and (s["fn"], s["desc"], s["t0"], s["t1"]) == (n["fn"], n["desc"], n["t0"], n["t1"]) and s["depth"] != n["depth"])
    info("I", "값이 같은 JFR 구간이 있는 트리 마디 가운데 구간의 depth가 트리 깊이와 다른 것", differ)

    for k, v in sorted(copies.items()):
        info("I", k, v)
    failed = [x for x in RESULTS if x[2] != x[3]]
    for cid, desc, ok, total, note in sorted(RESULTS):
        print(f"{cid}\t{ok}/{total}\t{desc}\t{note}".rstrip())
    for cid, desc, value in INFO:
        print(f"{cid}\t{value}\t{desc}")
    print(f"# failed {len(failed)}")
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1]))
