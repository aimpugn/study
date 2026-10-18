"""List screen-owned constants (uppercase const) in screen files: line span, bytes, top-level entry count.

Usage: python3 -I consts.py FILE.html [FILE.html ...]
Skips the data literals (DATA in concept-3, QUERIES, RUNS). Entry count = items at bracket depth 1.
"""
import re, sys
SKIP = {'QUERIES', 'RUNS'}
for path in sys.argv[1:]:
    s = open(path, encoding='utf-8').read()
    print('==', path.rsplit('/', 1)[-1])
    for m in re.finditer(r'(?m)^(\s*)const ([A-Z][A-Z_]{2,})\s*=\s*', s):
        name = m.group(2)
        if name in SKIP or (name == 'DATA' and 'concept-3' in path): continue
        i = m.end(); depth = 0; items = 0; q = None; start = i; seen_open = False
        while i < len(s):
            c = s[i]
            if q:
                if c == '\\': i += 2; continue
                if c == q: q = None
            elif c in '"\'`': q = c
            elif c in '[{(':
                depth += 1
                if depth == 1: seen_open = True; items = 0 if s[i+1:].lstrip()[:1] in ']})' else 1
            elif c in ']})':
                depth -= 1
                if depth == 0: i += 1; break
            elif c == ',' and depth == 1: items += 1
            elif c in ';\n' and depth == 0: break
            i += 1
        lit = s[start:i]
        l0 = s.count('\n', 0, m.start()) + 1; l1 = s.count('\n', 0, i) + 1
        trailing = lit.rstrip().rstrip(';').rstrip()
        n = items if seen_open and trailing[:1] in '[{' else '-'
        print(f'  {name}: lines {l0}-{l1} bytes={len(lit.encode())} entries={n} head={lit[:70]!r}')
