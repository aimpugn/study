"""Checks that were first run as one-off commands: identity of copies, coverage, and odd values.

Usage: python3 -I stats4.py DATA_DIR REPO_ROOT
Prints one labelled line per check. Reads only; writes nothing.
"""
import json, os, re, sys
from collections import Counter
data, root = sys.argv[1], sys.argv[2]
OSI = os.path.join(root, 'open_source_internals')
DESIGN = os.path.join(OSI, 'workbench/design')
PG = os.path.join(OSI, 'database/postgresql')
def load(n): return json.load(open(os.path.join(data, n + '.json')))
TX = ['concept-5-tx', 'concept-6-locks', 'concept-7-durability', 'concept-8-serializable']
R9 = load('concept-9-spring')

# 1. Screens equal their templates once the data line is put back to the placeholder.
for names, tmpl, var in [(TX, 'tx-view.tmpl.html', 'RUNS'), (['concept-9-spring'], 'spring-view.tmpl.html', 'RUNS')]:
    t = open(os.path.join(DESIGN, tmpl), encoding='utf-8').read()
    for n in names:
        s = open(os.path.join(DESIGN, n + '.html'), encoding='utf-8').read()
        back = re.sub(r'(?m)^const ' + var + r' = .*$', 'const ' + var + ' = __RUNS__;', s, count=1)
        print(f'[template] {n}.html == {tmpl} after restoring __RUNS__: {back == t}')

# 2. Replay page script equals web/app.js.
h = open(os.path.join(PG, 'replays/select-where-id-2.html'), encoding='utf-8').read()
app = open(os.path.join(OSI, 'workbench/web/app.js'), encoding='utf-8').read()
for m in re.finditer(r'<script([^>]*)>(.*?)</script>', h, re.S):
    if 'bundle' not in m.group(1):
        print(f'[replay] script == app.js (strip): {m.group(2).strip() == app.strip()} chars={len(m.group(2))} line={h.count(chr(10), 0, m.start()) + 1}')
print(f'[replay] pages in replays/: {sorted(os.listdir(os.path.join(PG, "replays")))}')

# 3. E-03 files: count and sizes.
d = os.path.join(PG, 'scripts/.lab-tmp/spring/decoded-r5')
fs = sorted(f for f in os.listdir(d) if f.endswith('.json'))
sz = [os.path.getsize(os.path.join(d, f)) for f in fs]
print(f'[e03 files] n={len(fs)} sum={sum(sz)} min={min(sz)} max={max(sz)}')

# 4. xip shapes in every snapshot.
xs = Counter()
for n in TX:
    for r in load(n).values():
        for st in r['steps']:
            for q in st['requests']:
                for sn in q.get('snapshots') or []: xs[(n, 'null' if sn.get('xip') is None else len(sn['xip']))] += 1
for r in R9.values():
    for cn in r['conns']:
        for q in cn['requests']:
            for sn in (q.get('server') or {}).get('snapshots') or []: xs[('e03', 'null' if sn.get('xip') is None else len(sn['xip']))] += 1
print('[xip]', dict(xs))

# 5. E-03 odd values: empty-conn borrows, overlapping captures, per-run leftAfter and leftFns, sessions, ambiguous, uprobe totals.
eb = [(k, b['result'], b['idleMs'], b['validated']) for k, r in sorted(R9.items()) for b in r['borrows'] if not b.get('conn')]
print('[e03 borrow conn=""]', len(eb), eb)
ov = [(k, c.get('label'), len(c['overlap'])) for k, r in sorted(R9.items()) for c in r['captures'] if c.get('overlap')]
print('[e03 overlap]', ov, 'total', sum(x[2] for x in ov))
la = sorted(r['siteRepair']['leftAfter'] for r in R9.values())
lf = Counter(fn for r in R9.values() for fn in r['siteRepair']['leftFns'])
print('[e03 leftAfter]', la, 'sum', sum(la), '| runs having each leftFn', dict(lf))
print('[e03 sessions per run]', {k: len(r['sessions']) for k, r in sorted(R9.items())})
print('[e03 ambiguous]', Counter(r['view']['ambiguous'] for r in R9.values()))
print('[e03 uprobe] events', sum(r['clock']['uprobe']['events'] for r in R9.values()), 'joined', sum(r['clock']['uprobe']['joined'] for r in R9.values()))
print('[e03 stackOnly events]', sum(s['events'] for r in R9.values() for s in r['java']['stackOnly']),
      'spans lab.*', sum(1 for r in R9.values() for s in r['java']['spans'] if s['fn'].startswith('lab.')),
      'tree lab.*', sum(1 for r in R9.values() for n in r['view']['tree'] if n['fn'].startswith('lab.')))
xe = sum(1 for r in R9.values() for cn in r['conns'] for q in cn['requests'] if q.get('server') for e in q['server']['events'] if e['kind'] == 'xid')
print('[e03 xid events vs transactions.xids]', xe, sum(len(t['xids']) for r in R9.values() for t in r['transactions']))
print('[e03 passed ends with t1 null]', sum(1 for r in R9.values() for n in r['view']['tree'] if n['end'] == 'passed' and n.get('t1') is None),
      'of passed', sum(1 for r in R9.values() for n in r['view']['tree'] if n['end'] == 'passed'))

# 6. tx: failing self-checks, raw-scale totals, leftAfter per run.
for n in TX:
    T = load(n)
    bad = [(k, s['n'], s['check']) for k, r in T.items() for s in r['steps'] if s.get('check') and not (s['check']['countMatch'] and s['check']['setMatch'])]
    print(f'[{n}] failing step.check {bad} | clock.events {sum(r["clock"]["events"] for r in T.values())} joined {sum(r["clock"]["joined"] for r in T.values())} | leftAfter {[r["siteRepair"]["leftAfter"] for r in T.values()]}')

# 7. concept-3 vs concept-4: same recordings?
c3 = load('concept-3'); c4 = {q['key']: q['data'] for q in load('concept-4')}
for k in ['update', 'select']:
    a, b = c3[k], c4[k]
    ka = Counter((x['kind'], x['t0']) for x in a['hops']); kb = Counter((x['kind'], x['t0']) for x in b['hops'])
    print(f'[c3~c4 {k}] pid {a["pid"]}/{b["pid"]} endUs {a["endUs"]}/{b["endUs"]} query_same {a["query"] == b["query"]} hops {len(a["hops"])}/{len(b["hops"])} common(kind,t0) {sum((ka & kb).values())} only_c3 {dict(ka - kb)}')
    print(f'    kinds only in c4: {sorted(set(x["kind"] for x in b["hops"]) - set(x["kind"] for x in a["hops"]))}')
print('[c4 same-sql runs]', [(k, c4[k]['pid'], c4[k]['endUs'], c4[k].get('kernelPid')) for k in ['insert', 'kernel_insert', 'kernel_insert2', 'create_index', 'kernel_index']])
print('[c3 src per query]', {k: len(v['src']) for k, v in c3.items()})

# 8. REL coverage of hop files.
for name, dd in [('concept-3', c3), ('concept-4', c4)]:
    s = open(os.path.join(DESIGN, name + '.html'), encoding='utf-8').read()
    rel = json.loads(re.search(r'const REL = (\{.*?\});', s).group(1))
    files = set(re.sub(r'(_fsm|_vm|_init|\.\d+)$', '', f) for q in dd.values() for x in q['hops'] for f in (x.get('files') or []))
    print(f'[{name} REL] entries {len(rel)} distinct hop files {len(files)} non-data {sorted(f for f in files if not f.startswith(("base/", "global/")))} data-not-in-REL {sorted(f for f in files if f.startswith(("base/", "global/")) and f not in rel)} REL-unused {len([k for k in rel if k not in files])}')

# 9. gdb traces: BT depth continuity and the STOP each BT block follows.
ev = os.path.join(PG, 'evidence')
for f in sorted(os.listdir(ev)):
    if not f.endswith('.trace'): continue
    blocks, prev, cur = [], None, None
    for line in open(os.path.join(ev, f), encoding='utf-8'):
        p = line.split(' ')
        if p[0] == 'BT':
            if cur is None: cur = {'after': prev, 'depths': []}; blocks.append(cur)
            cur['depths'].append(int(p[1]))
        else:
            cur = None
            if p[0] == 'STOP': prev = f'{p[1]} {p[2]}'
    print(f'[bt {f}]', [(b['after'], len(b['depths']), b['depths'] == list(range(len(b['depths'])))) for b in blocks])
