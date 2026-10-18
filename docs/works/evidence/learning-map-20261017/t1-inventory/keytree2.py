"""Key tree for embedded screen data: treat top-level dict values (RUNS/DATA) or list items (QUERIES) as documents.

Usage: python3 -I keytree2.py FILE.json [FILE.json ...] [--fixed=PATH ...]
"""
import json, sys, runpy, os
here = os.path.dirname(os.path.abspath(__file__))
docs = []
flags = [a for a in sys.argv[1:] if a.startswith('--')]  # passed through to keytree.py (e.g. --fixed=PATH)
for p in [a for a in sys.argv[1:] if not a.startswith('--')]:
    d = json.load(open(p))
    docs.extend(d.values() if isinstance(d, dict) else d)
tmp = os.path.join(here, '_docs.tmp.json')
paths = []
for i, d in enumerate(docs):
    q = os.path.join(here, f'_doc{i}.tmp.json')
    json.dump(d, open(q, 'w'))
    paths.append(q)
sys.argv = ['keytree.py'] + paths + flags
runpy.run_path(os.path.join(here, 'keytree.py'), run_name='__main__')
for q in paths: os.remove(q)
