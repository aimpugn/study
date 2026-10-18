"""Compare concept-9 embedded RUNS with the decoded-r5 JSON files (deep equality, list diff paths)."""
import json, sys, os
runs = json.load(open(sys.argv[1]))
d = sys.argv[2]
def diff(a, b, path, out, limit=10):
    if len(out) >= limit: return
    if type(a) != type(b):
        if isinstance(a,(int,float)) and isinstance(b,(int,float)) and a == b: return
        out.append(f'{path}: type {type(a).__name__} vs {type(b).__name__}'); return
    if isinstance(a, dict):
        for k in set(a) | set(b):
            if k not in a: out.append(f'{path}.{k}: missing in screen'); continue
            if k not in b: out.append(f'{path}.{k}: missing in file'); continue
            diff(a[k], b[k], path + '.' + k, out, limit)
    elif isinstance(a, list):
        if len(a) != len(b): out.append(f'{path}: len {len(a)} vs {len(b)}'); return
        for i, (x, y) in enumerate(zip(a, b)): diff(x, y, f'{path}[{i}]', out, limit)
    elif a != b:
        out.append(f'{path}: {str(a)[:60]!r} vs {str(b)[:60]!r}')
for name, v in runs.items():
    scene, run = name.split('/')
    fn = 'r' + scene + '-' + run + '.json'
    f = json.load(open(os.path.join(d, fn)))
    out = []
    diff(v, f, '$', out)
    print(name, fn, 'IDENTICAL' if not out else 'DIFF', out[:10])
