"""Second pass: inference-marker counts in the tx runs and concept-4 (hint attribution, ftrace-clock events, procsGuess).

Usage: python3 -I stats2.py DATA_DIR
"""
import json, os, sys
from collections import Counter
data = sys.argv[1]
def load(n): return json.load(open(os.path.join(data, n + '.json')))
for name in ['concept-5-tx', 'concept-6-locks', 'concept-7-durability', 'concept-8-serializable']:
    R = load(name)
    via, clk, chk, ab, cap = Counter(), Counter(), Counter(), Counter(), Counter()
    for r in R.values():
        for hc in r.get('hintChanges') or []: via[hc.get('via')] += 1; via['by=' + str(hc.get('by'))[:12]] += 0
        for e in (r.get('server') or {}).get('events') or []:
            clk[(e.get('f') or {}).get('clock', '-')] += 1
        for st in r.get('steps') or []:
            if st.get('check'): chk[json.dumps(st['check'].get('result'))[:40]] += 1
            if st.get('blocked'): ab['blocked'] += 1
            for q in st.get('requests') or []:
                if q.get('aborted'): ab['aborted'] += 1
                if q.get('t1') is None: ab['req_t1_null'] += 1
        for c in r.get('captures') or []:
            if c.get('afterKill'): cap['afterKill'] += 1
            if c.get('wal'): cap['wal'] += 1
    print(name, 'hint via', dict(via), '| server ev clock', dict(clk), '| step check', dict(chk), dict(ab), dict(cap))
Q = load('concept-4')
for q in Q:
    print(q['key'], 'typed=', q.get('typed'), 'x=', q.get('x'), 'procsGuess=', q.get('procsGuess'), 'kernelPid=', q['data'].get('kernelPid'), 'sysUs hops=', sum(1 for h in q['data']['hops'] if 'sysUs' in h), 'cut msgs=', sum(1 for h in q['data']['hops'] for m in (h.get('msgs') or []) if m.get('cut') or m.get('cont')))
