"""E-05: compare STOP counts per function with END counts, list value expressions per file, BT stack groups.

Usage: python3 -I stats3.py EVIDENCE_DIR
"""
import os, sys, re
from collections import Counter
ev = sys.argv[1]
for f in sorted(os.listdir(ev)):
    if not f.endswith('.trace'): continue
    stops, end, exprs, bt_groups, bt_lines = Counter(), {}, Counter(), 0, 0
    prev = None; nopath = 0
    for line in open(os.path.join(ev, f), encoding='utf-8'):
        if line.startswith('--- '): break
        p = line.rstrip('\n').split(' ')
        if p[0] == 'STOP':
            stops[p[2]] += 1
            if ':' not in p[3]: nopath += 1
            for m in re.finditer(r'(?:^| )([^ =]+(?: \*\)[^=]*)?)=', ' '.join(p[4:])): pass
            rest = ' '.join(p[4:])
            if rest: exprs[p[2]] += 1
            prev = 'STOP'
        elif p[0] == 'BT':
            bt_lines += 1
            if prev != 'BT': bt_groups += 1
            prev = 'BT'
        elif p[0] == 'END':
            for kv in p[1:]:
                k, v = kv.rsplit('=', 1); end[k] = int(v)
    diff = {k: (stops.get(k, 0), end.get(k, 0)) for k in set(stops) | set(end) if stops.get(k, 0) != end.get(k, 0)}
    print(f'{f}: stops={sum(stops.values())} end_sum={sum(end.values())} bt_groups={bt_groups} bt_lines={bt_lines} stops_with_values={sum(exprs.values())} nopath={nopath} diff(stop,end)={diff}')
