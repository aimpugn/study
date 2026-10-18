"""Extract embedded datasets from the workbench concept screens into JSON files.

Usage: python3 -I extract.py DESIGN_DIR OUT_DIR
Writes concept-3.json (DATA), concept-4.json (QUERIES), concept-5..9.json (RUNS) and
prints, for each, the character and UTF-8 byte offset of the literal and its length.
"""
import json, re, sys, os
design, out = sys.argv[1], sys.argv[2]
specs = [
    ('concept-3.html', 'DATA'), ('concept-4.html', 'QUERIES'),
    ('concept-5-tx.html', 'RUNS'), ('concept-6-locks.html', 'RUNS'),
    ('concept-7-durability.html', 'RUNS'), ('concept-8-serializable.html', 'RUNS'),
    ('concept-9-spring.html', 'RUNS'),
]
for fname, var in specs:
    s = open(os.path.join(design, fname), encoding='utf-8').read()
    m = re.search(r'(?m)^\s*const\s+' + var + r'\s*=\s*', s)
    start = m.end()
    end = s.find('\n', start)
    lit = s[start:end].rstrip()
    if lit.endswith(';'):
        lit = lit[:-1]
    if var == 'DATA':
        # unquoted top-level keys: {update: {...}, select: {...}}
        lit = re.sub(r'^\{(\w+):', r'{"\1":', lit)
        lit = re.sub(r'\},\s*(\w+):\s*\{"endUs"', r'},"\1":{"endUs"', lit)
    data = json.loads(lit)
    byte_off = len(s[:start].encode('utf-8'))
    byte_len = len(lit.encode('utf-8'))
    line = s.count('\n', 0, start) + 1
    base = fname.replace('.html', '')
    with open(os.path.join(out, base + '.json'), 'w') as f:
        json.dump(data, f)
    keys = list(data.keys()) if isinstance(data, dict) else [d.get('key') for d in data]
    print(f'{fname}: var={var} line={line} char_off={start} byte_off={byte_off} byte_len={byte_len} top={type(data).__name__} n={len(data)} keys={keys}')
