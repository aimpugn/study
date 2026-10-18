"""Print a merged key tree (path -> types, occurrence counts) for JSON documents.

Usage: python3 -I keytree.py FILE... [--depth N]
Arrays are shown as [] and dict keys under maps with dynamic keys are collapsed when
the key looks dynamic (digits, contains ':' or '/', or more than 40 distinct keys).
"""
import json, sys, collections

def typ(v):
    if v is None: return 'null'
    if isinstance(v, bool): return 'bool'
    if isinstance(v, int): return 'int'
    if isinstance(v, float): return 'float'
    if isinstance(v, str): return 'str'
    if isinstance(v, list): return 'list'
    return 'obj'

stats = collections.OrderedDict()
def note(path, v):
    s = stats.setdefault(path, {'types': collections.Counter(), 'n': 0, 'lens': []})
    s['types'][typ(v)] += 1
    s['n'] += 1
    if isinstance(v, (list, dict)):
        s['lens'].append(len(v))

DYN = set()
def walk(path, v, dynamic_paths):
    note(path, v)
    if isinstance(v, dict):
        dyn = path in dynamic_paths
        for k, x in v.items():
            walk(path + ('.{*}' if dyn else '.' + k), x, dynamic_paths)
    elif isinstance(v, list):
        for x in v:
            walk(path + '[]', x, dynamic_paths)

def find_dynamic(path, v, keysets):
    if isinstance(v, dict):
        keysets.setdefault(path, set()).update(v.keys())
        for k, x in v.items():
            find_dynamic(path + '.' + k, x, keysets)
    elif isinstance(v, list):
        for x in v:
            find_dynamic(path + '[]', x, keysets)

args = [a for a in sys.argv[1:] if not a.startswith('--')]
docs = []
for p in args:
    with open(p) as f:
        docs.append(json.load(f))
# Determine dynamic-key objects by iterating to fixed point (paths change once collapsed).
dynamic = set()
for _ in range(6):
    keysets = {}
    def fd(path, v):
        if isinstance(v, dict):
            dyn = path in dynamic
            keysets.setdefault(path, set()).update(v.keys())
            for k, x in v.items():
                fd(path + ('.{*}' if dyn else '.' + k), x)
        elif isinstance(v, list):
            for x in v: fd(path + '[]', x)
    for d in docs: fd('$', d)
    new = set(dynamic)
    for p, ks in keysets.items():
        if len(ks) > 40 or any(any(c.isdigit() for c in k[:1]) or ':' in k or '/' in k or '.' in k or ' ' in k for k in ks):
            new.add(p)
    if new == dynamic: break
    dynamic = new
for d in docs:
    walk('$', d, dynamic)
for p, s in stats.items():
    t = ','.join(f'{k}:{v}' for k, v in s['types'].items())
    extra = ''
    if s['lens']:
        extra = f' len[min={min(s["lens"])},max={max(s["lens"])},sum={sum(s["lens"])}]'
    dyn = ' (dynamic keys)' if p in dynamic else ''
    print(f'{p}  n={s["n"]} {t}{extra}{dyn}')
