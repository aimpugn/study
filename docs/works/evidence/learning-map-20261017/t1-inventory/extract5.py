"""Extract the embedded replay bundle (script#bundle) from a workbench replay page and print its shape.

Usage: python3 -I extract5.py REPLAY.html OUT.json
"""
import json, re, sys
s = open(sys.argv[1], encoding='utf-8').read()
m = re.search(r'<script id="bundle" type="application/json">', s)
start = m.end(); end = s.index('</script>', start)
lit = s[start:end]
b = json.loads(lit)
json.dump(b, open(sys.argv[2], 'w'), ensure_ascii=False)
line = s.count('\n', 0, start) + 1
print('line', line, 'byte_off', len(s[:start].encode()), 'byte_len', len(lit.encode()))
print('top keys', list(b.keys()))
print('meta', b.get('meta'))
print('lanes', b.get('lanes'))
print('steps', len(b['steps']), 'stack frames', sum(len(x.get('stack') or []) for x in b['steps']), 'values', sum(len(x.get('values') or []) for x in b['steps']), 'raw', sum(1 for x in b['steps'] if x.get('rawValues')))
print('step0', json.dumps(b['steps'][0], ensure_ascii=False)[:600])
print('files', len(b.get('files') or {}), 'lines', sum(len(v) for v in (b.get('files') or {}).values()))
print('counts', len(b.get('counts') or []), 'output', b.get('output'), 'noBreak', b.get('noBreak'))
print('other keys', {k: type(v).__name__ for k, v in b.items()})
