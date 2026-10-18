"""Enumerate observed key sets behind dynamic-key nodes ({*}) and per-kind attribute (f) keys.

Usage: python3 -I dyn.py DATA_DIR
"""
import json, os, sys, re
from collections import Counter, defaultdict
data = sys.argv[1]
def load(n): return json.load(open(os.path.join(data, n + '.json')))
def fkeys(events, out):
    for e in events or []:
        out[e['kind']][0] += 1
        for k in (e.get('f') or {}): out[e['kind']][1][k] += 1
def show(title, out):
    print('==', title)
    for kind, (n, ks) in sorted(out.items(), key=lambda x: -x[1][0]):
        print(f'  {kind} ({n}): ' + ', '.join(f'{k}:{v}' for k, v in sorted(ks.items())))
R9 = load('concept-9-spring')
e3 = defaultdict(lambda: [0, Counter()]); xacts = Counter(); sess = Counter(); dropped = Counter(); before = Counter(); srcfmt = Counter(); jsrcfmt = Counter()
for r in R9.values():
    for cn in r['conns']:
        for q in cn['requests']:
            if q.get('server'): fkeys(q['server'].get('events'), e3)
    for c in r['captures']:
        for k, v in (c.get('xacts') or {}).items(): xacts['value=' + v] += 1
    for k, v in (r.get('sessions') or {}).items(): sess[f'{k}->{v}'] += 1
    for k, v in r['java']['dropped'].items(): dropped[k] += v
    for k, v in r['siteRepair']['before'].items(): before[k] += v
    for k in r['src']: srcfmt[re.sub(r':\d+$', ':N', k).split('/')[1] if '/' in k else k] += 1
    for k in r['view']['jsrc']: jsrcfmt['class:N' if re.match(r'^[\w.$]+:\d+$', k) else k] += 1
show('E-03 server.events f keys by kind', e3)
print('E-03 captures.xacts values', dict(xacts)); print('E-03 sessions', dict(sess)); print('E-03 java.dropped sum', dict(dropped)); print('E-03 siteRepair.before', dict(before))
print('E-03 src key form', dict(srcfmt), 'jsrc key form', dict(jsrcfmt))
rq = defaultdict(lambda: [0, Counter()]); sv = defaultdict(lambda: [0, Counter()]); xa = Counter(); bf = Counter(); pg = Counter()
for name in ['concept-5-tx', 'concept-6-locks', 'concept-7-durability', 'concept-8-serializable']:
    for r in load(name).values():
        for st in r['steps']:
            for q in st['requests']: fkeys(q.get('events'), rq)
        fkeys((r.get('server') or {}).get('events'), sv)
        for c in r['captures']:
            for k, v in (c.get('xacts') or {}).items(): xa['value=' + v] += 1
            for t in (c.get('pages') or {}): pg[t] += 1
        for k, v in (r.get('siteRepair') or {}).get('before', {}).items(): bf[k] += v
show('tx steps.requests.events f keys by kind', rq)
show('tx server.events f keys by kind', sv)
print('tx captures.xacts values', dict(xa)); print('tx siteRepair.before', dict(bf)); print('tx captures.pages tables (captures having)', dict(pg))
