"""Kind distributions and event-scale counts for the extracted screen data.

Usage: python3 -I stats.py DATA_DIR EVIDENCE_DIR
Prints, per input, counters of event kinds, hop kinds, effect kinds, ends, lanes,
kernel node totals and gdb trace line kinds. Reads only; writes nothing.
"""
import json, os, sys, re
from collections import Counter
data, ev = sys.argv[1], sys.argv[2]
def load(n): return json.load(open(os.path.join(data, n + '.json')))
def show(title, c):
    print(f'  {title}: total={sum(c.values())} ' + ', '.join(f'{k}={v}' for k, v in sorted(c.items(), key=lambda x: (-x[1], str(x[0])))))

# E-03 == concept-9 RUNS
R = load('concept-9-spring')
print('== E-03 / concept-9', len(R), 'runs')
c = {k: Counter() for k in ['sev', 'msgtype', 'msgdir', 'effect', 'end', 'lane', 'why', 'cap', 'snapkind', 'clog', 'reqstatus', 'notes']}
tot = Counter()
for key, r in sorted(R.items()):
    per = Counter()
    for cn in r.get('conns') or []:
        for m in cn.get('messages') or []:
            c['msgtype'][(m['dir'], m['type'])] += 1; per['messages'] += 1
            if m.get('tc') is not None: per['msg_tc'] += 1
        for q in cn.get('requests') or []:
            per['requests'] += 1
            c['reqstatus'][q.get('status')] += 1
            s = q.get('server')
            if s:
                per['req_server'] += 1
                for e in s.get('events') or []:
                    c['sev'][e['kind']] += 1; per['server_events'] += 1
                per['snapshots'] += len(s.get('snapshots') or [])
                per['checks'] += len(s.get('checks') or [])
                cs = s.get('catalogSnapshots')
                if cs: per['catalogSnapshots_sum'] += cs if isinstance(cs, int) else 0
    for t in r.get('transactions') or []:
        per['tx'] += 1
        for x in t.get('clog') or []: c['clog'][x['kind']] += 1
    j = r.get('java') or {}
    for k in ['spans', 'throws', 'parks', 'dups', 'stackOnly']:
        per['java_' + k] += len(j.get(k) or [])
    v = r.get('view') or {}
    for n in v.get('tree') or []:
        per['tree'] += 1
        c['end'][n.get('end')] += 1; c['lane'][n.get('lane')] += 1
        for w in n.get('why') or []: c['why'][w] += 1
        for e in n.get('effects') or []: c['effect'][e['kind']] += 1; per['effects'] += 1
        if n.get('hidden'): per['tree_hidden'] += 1
    per['captures'] += len(r.get('captures') or [])
    per['checkpoints'] += len(r.get('checkpoints') or [])
    per['borrows'] += len(r.get('borrows') or [])
    per['joins'] += len(r.get('joins') or [])
    per['notes'] += len(r.get('notes') or [])
    per['blocked'] += len(r.get('blocked') or [])
    per['waitCycles'] += len(v.get('waitCycles') or [])
    per['binding'] += len(v.get('binding') or [])
    tot.update(per)
    print(f'  {key}: ' + ' '.join(f'{k}={per[k]}' for k in ['messages','requests','server_events','snapshots','checks','java_spans','java_throws','tree','effects','captures','borrows','joins']))
print('  TOTAL', dict(sorted(tot.items())))
for k in c: show(k, c[k])

# tx runs concept-5..8
for name in ['concept-5-tx', 'concept-6-locks', 'concept-7-durability', 'concept-8-serializable']:
    R = load(name)
    print('==', name, len(R), 'runs', list(R.keys()))
    rk, sk, roles, spfn = Counter(), Counter(), Counter(), Counter()
    tot = Counter()
    for key, r in R.items():
        tot['steps'] += len(r.get('steps') or [])
        for st in r.get('steps') or []:
            for q in st.get('requests') or []:
                tot['requests'] += 1
                for e in q.get('events') or []: rk[e['kind']] += 1
                tot['snapshots'] += len(q.get('snapshots') or [])
                tot['checks'] += len(q.get('checks') or [])
                tot['out_msgs'] += len(q.get('out') or [])
                tot['in_msgs'] += 1 if q.get('in') else 0
        s = r.get('server') or {}
        for p in s.get('procs') or []: roles[p.get('role')] += 1
        for e in s.get('events') or []: sk[e['kind']] += 1
        for p in s.get('spans') or []: spfn[p.get('fn')] += 1
        tot['captures'] += len(r.get('captures') or [])
        tot['hintChanges'] += len(r.get('hintChanges') or [])
        tot['joins'] += len(r.get('joins') or [])
        tot['kills'] += len(r.get('kills') or [])
        tot['bystanders'] += len(r.get('bystanders') or [])
    print('  TOTAL', dict(sorted(tot.items())))
    show('request.events kind', rk); show('server.events kind', sk); show('server.procs role', roles); show('server.spans fn', spfn)

# concept-3 / concept-4 hops
def knodes(ks):
    n = h = g = cl = 0
    for k in ks or []:
        n += 1
        if k.get('hidden'): h += 1
        if k.get('guess'): g += 1
        if k.get('calls'): cl += 1
        a, b, c2, d = knodes(k.get('kids'))
        n += a; h += b; g += c2; cl += d
    return n, h, g, cl
for name in ['concept-3', 'concept-4']:
    D = load(name)
    items = list(D.items()) if isinstance(D, dict) else [(q['key'], q['data']) for q in D]
    print('==', name, len(items), 'queries')
    hk, stg = Counter(), Counter()
    tot = Counter()
    for key, q in items:
        per = Counter()
        for hp in q.get('hops') or []:
            per['hops'] += 1; per['count_sum'] += hp.get('count') or 1
            hk[hp.get('kind')] += hp.get('count') or 1; stg[hp.get('stage')] += 1
            per['msgs'] += len(hp.get('msgs') or [])
            per['frames'] += len(hp.get('stack') or [])
            if hp.get('t0') != hp.get('t1'): per['t0ne_t1'] += 1
            if (hp.get('count') or 1) > 1: per['merged'] += 1
            if hp.get('kernel'):
                a, b, c2, d = knodes(hp['kernel']); per['knodes'] += a; per['khidden'] += b; per['kguess'] += c2; per['kcalls'] += d
        tot.update(per)
        print(f'  {key}: ' + ' '.join(f'{k}={v}' for k, v in sorted(per.items())))
    print('  TOTAL', dict(sorted(tot.items())))
    show('hop kind (weighted by count)', hk); show('hop stage (rows)', stg)

# E-05 gdb traces
print('== E-05')
for f in sorted(os.listdir(ev)):
    if not f.endswith('.trace'): continue
    kc = Counter(); vals = 0
    in_psql = False
    for line in open(os.path.join(ev, f), encoding='utf-8', errors='replace'):
        w = line.split('\t', 1)[0].split(' ', 1)[0].rstrip('\n')
        if line.startswith('--- '): in_psql = True; kc['psql_hdr'] += 1; continue
        if in_psql: kc['psql_line'] += 1; continue
        if w in ('STOP', 'BT', 'END', 'NOBREAK'): kc[w] += 1
        elif line.startswith('#'): kc['comment'] += 1
        elif line.strip() == '': kc['blank'] += 1
        else: kc['other:' + w[:20]] += 1
    print(f'  {f}: ' + ' '.join(f'{k}={v}' for k, v in sorted(kc.items())))
