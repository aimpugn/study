"""Attach a role mark and a provenance mark to every key-tree line, from explicit rules.

Usage: python3 -I marks.py TREE.txt PROFILE   (PROFILE: e03 | tx | c3 | c4)
Reads a keytree.py output. Collapses recursive kernel paths (.kids[] repeated) into one line
with summed counts. Prints "path  n=..  types  | role | provenance | rule".
Role: 장소 사건 연결 값 출처 시각 발췌 모음. Provenance: 관찰(recorded), 도출(decoder computed),
이름(human label text). First matching rule wins; every line gets exactly one rule id.
"""
import re, sys
from collections import OrderedDict, Counter
tree, prof = sys.argv[1], sys.argv[2]

# (id, path regex, role, provenance). Profile-specific rules come first.
P = {
 'e03': [
  ('E1', r'^\$\.view\.(java|jsrc)\b', '발췌', '관찰'),
  ('E2', r'^\$\.view\.tree\[\]\.(parent|kids(\[\])?|branch(\..*)?)$', '연결', '도출'),
  ('E3', r'^\$\.view\.tree\[\]\.effects\[\]\.seq(\[\])?$', '연결', '도출'),
  ('E4', r'^\$\.view\.tree\[\]\.(label|why(\[\])?)$', '이름', '도출'),
  ('E5', r'^\$\.view\.tree\[\]\.(end|hiddenKinds(\[\])?)$', '값', '도출'),
  ('E6', r'^\$\.view\.tree\[\]\.hidden(\[\])?$', '장소', '도출'),
  ('E7', r'^\$\.view\.tree\[\]$', '사건', '도출'),
  ('E8', r'^\$\.view\.tree\[\]\.effects\[\]$', '사건', '도출'),
  ('E9', r'^\$\.view\.branches\.\{\*\}\.alts\[\]\.(seq|records)(\[\])?$', '연결', '도출'),
  ('E9b', r'^\$\.view\.branches\.\{\*\}\.(alts\[\]\.sites|lines)(\[\])?$', '장소', '도출'),
  ('E9c', r'^\$\.view\.tree\[\]\.id$', '연결', '도출'),
  ('E10', r'^\$\.view\.branches\.\{\*\}\.(alts\[\]\.)?(label|labels(\[\])?)$', '이름', '도출'),
  ('E11', r'^\$\.view\.(compared|roots)(\[\])?$', '연결', '도출'),
  ('E12', r'^\$\.view\.waitCycles\[\]\.(waiter|holder|heldBy|call|cycle)$', '연결', '도출'),
  ('E13', r'^\$\.view\.(javaFrom|siteMismatch)$', '출처', '도출'),
  ('E14', r'^\$\.view\.poolConfig\.\w+At$', '발췌', '도출'),
  ('E15', r'^\$\.view\.poolConfig\.\w+Ms$', '값', '도출'),
  ('E16', r'^\$\.view\.binding\[\]\.(by|bound|held(\[\])?)$', '연결', '도출'),
  ('E17', r'^\$\.view\.(pauses|window)(\[\]){0,2}$', '시각', '도출'),
  ('E18', r'^\$\.view\b', None, '도출'),
  ('E19', r'^\$\.transactions\[\]\.(first|last)$', '연결', '도출'),
  ('E20', r'^\$\.transactions\[\]\.next(\..*)?$', '연결', '도출'),
  ('E21', r'^\$\.transactions\[\]\.end$', '값', '도출'),
  ('E22', r'^\$\.transactions\b', None, '도출'),
  ('E23', r'^\$\.borrows\[\]\.conn$', '연결', '도출'),
  ('E24', r'^\$\.borrows\b', None, '도출'),
  ('E25', r'^\$\.captures\[\]\.(visible|overlap|span)\b', None, '도출'),
  ('E26', r'^\$\.captures\[\]\.reads\b', '시각', '관찰'),
  ('E27', r'^\$\.captures\[\]\.label$', '이름', '관찰'),
  ('E28', r'^\$\.captures\[\]\.afterStep$', '연결', '관찰'),
  ('E29', r'^\$\.(clock|joins|siteRepair)\b', '출처', '도출'),
  ('E30', r'^\$\.java\.(dups|stackOnly|dropped)\b', '출처', '도출'),
  ('E31', r'^\$\.java\.spans\[\]\.twice$', '출처', '도출'),
  ('E32', r'^\$\.java\.spans\[\]\.(caller|depth)$', '연결', '도출'),
  ('E33', r'^\$\.conns\[\]\.requests\[\]\.server\.checks\[\]\.visible$', '값', '도출'),
  ('E34', r'^\$\.conns\[\]\.requests\[\]\.server\.(t0|t1)$', '시각', '도출'),
  ('E35', r'^\$\.conns\[\]\.requests\[\]\.(seq|sent|done)$', None, '도출'),
  ('E36', r'^\$\.conns\[\]\.join(\[\])?$', '출처', '도출'),
  ('E37', r'^\$\.conns\[\]\.(lastMac|lastCont)$', '출처', '관찰'),
  ('E38', r'^\$\.sessions\.\{\*\}$', '장소', '관찰'),
  ('E39', r'^\$\.(blocked|notes|checkpoints)\[\]\.(label|where|chain|message)$', '이름', '관찰'),
  ('E40', r'^\$\.blocked\[\]\.after$', '연결', '관찰'),
  ('E41', r'^\$\.blocked\[\]\.waiters(\[\])?$', '장소', '관찰'),
  ('E42', r'^\$\.(title|scene|pool|javaVersion|javaExit)$', '값', '관찰'),
 ],
 'tx': [
  ('T1', r'^\$\.steps\[\]\.check\b', '값', '도출'),
  ('T2', r'^\$\.(bystanders|hintChanges)\b', None, '도출'),
  ('T3', r'^\$\.hintChanges\[\]\.(by|via)$', '출처', '도출'),
  ('T4a', r'^\$\.tupleAddrs\[\]\.epoch$', '값', '도출'),
  ('T4', r'^\$\.tupleAddrs\b', '장소', '도출'),
  ('T5', r'^\$\.(clock|joins|siteRepair)\b', '출처', '도출'),
  ('T6', r'^\$\.server\.procs\[\]\.role$', '이름', '도출'),
  ('T7', r'^\$\.captures\[\]\.reads\b', '시각', '관찰'),
  ('T8', r'^\$\.captures\[\]\.(afterStep|afterKill)$', '연결', '관찰'),
  ('T9', r'^\$\.steps\[\]\.requests\[\]\.(t0|t1)$', '시각', '도출'),
  ('T10', r'^\$\.steps\[\]\.requests\[\]\.(in|out)(\[\])?$', '사건', '관찰'),
  ('T11', r'^\$\.steps\[\]\.requests\[\]\.checks\[\]\.visible$', '값', '도출'),
  ('T12', r'^\$\.steps\[\]\.(output|blocked)(\[\])?$', '값', '관찰'),
  ('T13', r'^\$\.steps\[\]\.n$', '연결', '관찰'),
  ('T14', r'^\$\.(title|scenario)$', '값', '관찰'),
  ('T15', r'^\$\.steps\[\]\.requests\[\]\.events\[\]\.f\.\{\*\}$', '값', '관찰'),
  ('T16', r'^\$\.server\.events\[\]\.f\.clock$', '출처', '도출'),
 ],
 'c3': [
  ('C1', r'^\$(\.data)?\.hops\[\]\.(count|t1)$', None, '도출'),
  ('C2', r'^\$(\.data)?\.hops\[\]\.kind$', '사건', '이름'),
  ('C3', r'^\$(\.data)?\.hops\[\]\.(stage|catalog)$', '값', '도출'),
  ('C4', r'^\$(\.data)?\.stages\[\]\.name$', '이름', '도출'),
  ('C5', r'^\$(\.data)?\.stages\[\]$', '사건', '도출'),
  ('C6', r'^\$(\.data)?\.hops\[\]\.n$', '연결', '도출'),
  ('C7', r'^\$(\.data)?\.hops\[\]\.(from|to)$', '장소', '도출'),
  ('C8', r'^\$(\.data)?\.hops\[\]\.proc$', '장소', '도출'),
  ('C9', r'^\$(\.data)?\.hops\[\]\.msgs\[\]\.(cut|cont)$', '연결', '도출'),
  ('C10', r'^\$\.data\.hops\[\]\.kernel.*\.(calls|callsOf)\b.*$', '발췌', '도출'),
  ('C11', r'^\$\.data\.hops\[\]\.kernel.*\.(guess|hidden)$', '출처', '도출'),
  ('C12', r'^\$\.data\.hops\[\]\.kernel.*\.layer$', '이름', '도출'),
  ('C13', r'^\$\.data\.hops\[\]\.kernel.*\.kids$', '연결', '관찰'),
  ('C13b', r'^\$\.data\.hops\[\]\.kernel.*\.kids\[\](\{2\+\})?$', '사건', '관찰'),
  ('C14', r'^\$\.data\.hops\[\]\.kernel(\[\])?$', '사건', '관찰'),
  ('C15', r'^\$\.procsGuess$', '출처', '도출'),
  ('C16', r'^\$\.procs\b', '장소', '관찰'),
  ('C17', r'^\$\.(key|label|typed|x)$', '값', '이름'),
  ('C18', r'^\$\.data\.kernelPid$', '장소', '관찰'),
  ('C19', r'^\$(\.data)?\.query$', '값', '관찰'),
  ('C20', r'^\$(\.data)?\.endUs$', '시각', '도출'),
 ],
}
P['c4'] = P['c3']
P['e05'] = [
  ('X1', r'^\$\.steps\[\]\.(place|stack\[\])\.(layer|lane)$', '이름', '도출'),
  ('X2', r'\.window\b', '발췌', '도출'),
  ('X3', r'^\$\.files\b', '발췌', '관찰'),
  ('X4', r'^\$\.lanes\b', '이름', '도출'),
  ('X5', r'^\$\.(subjectId|subjectTitle|sourceRoot|sourceNotice)$', '값', '이름'),
  ('X6', r'^\$\.meta\.sql$', '값', '도출'),
  ('X7', r'^\$\.meta\b', '출처', '이름'),
  ('X8', r'^\$\.steps\[\]\.seq$', '연결', '관찰'),
  ('X9a', r'^\$\.counts\[\]\.function$', '장소', '관찰'),
  ('X9', r'^\$\.counts\b', '값', '관찰'),
  ('X10', r'\.loc(\.(path|line))?$', '장소', '관찰'),
  ('X11', r'^\$\.steps\[\]\.values\[\]\.expr$', '값', '이름'),
  ('X12', r'^\$\.steps\[\]\.place$', '장소', '관찰'),
  ('X13', r'^\$\.output\b', '값', '관찰'),
]
# Generic rules on the last key name (after the profile rules).
G = [
  ('G1', r'(^|\.)(src|ksrc)\b.*$', '발췌', '관찰'),
  ('G2', r'\.stack(\[\])?(\.(fn|at))?$', '장소', '관찰'),
  ('G3', r'\.(t|t0|t1|tc|tCapture|tGo|sent|done|seen|accept|close|endUs|span|durs|idleMs|self|us|sysUs|durUs|maxDeviationUs|medianOffsetUs|lo|hi|first|last|tReady|walWrittenT)(\[\])?$', '시각', '관찰'),
  ('G4', r'\.(pid|conn|name|thread|fn|desc|site|at|path|table|rel|relation|ctid|block|lp|lane|proc|session|key|id|files|file|loc|function|layer|page|tuple|xid|vxid|txid|sxact|blk|buf|tup|off|rpid|wpid|rxid|wxid|xid1|xid2|who|to|from|caller|under)(\[\])?$', '장소', '관찰'),
  ('G5', r'\.(parent|kids|join|branch|next|by|seq|roots|records|compared|after|beforeStep|step)(\[\])?$', '연결', '관찰'),
  ('G5b', r'(\.f\.type|\.f\.kind|locks\[\]\.type|columns\[\]\.type|tupleAddrs\[\]\.epoch)$', '값', '관찰'),
  ('G5c', r'(^\$\.(tables|sessions|conns)\[\]|\.(columns|indexes|procs)\[\])$', '장소', '관찰'),
  ('G5d', r'(captures\[\]\.(pages|visible)\.\w+\[\]|captures\[\]\.(activity|locks)\[\])$', '값', '관찰'),
  ('G6', r'\.(kind|type|dir)$', '사건', '관찰'),
  ('G7', r'(\.(clock|recorder|uprobe|via|joined)|\.uprobe\.events)$', '출처', '관찰'),
  ('G8', r'\[\]$', '사건', '관찰'),
  ('G9', r'.*', '값', '관찰'),
]
def collapse(path):
    return re.sub(r'(\.kids\[\]){2,}', '.kids[]{2+}', path)
rows = OrderedDict()
for line in open(tree, encoding='utf-8'):
    m = re.match(r'^(\S+)\s+n=(\d+)\s+(.*)$', line.rstrip('\n'))
    if not m: continue
    path, n, rest = collapse(m.group(1)), int(m.group(2)), m.group(3)
    types = Counter()
    for t, c in re.findall(r'(\w+):(\d+)', re.sub(r'len\[[^\]]*\]', '', rest)): types[t] += int(c)
    lenpart = re.search(r'len\[[^\]]*\]', rest)
    dyn = ' (dynamic keys)' if 'dynamic' in rest else ''
    if path in rows:
        r = rows[path]; r['n'] += n; r['types'].update(types); r['len'] = None; r['merged'] += 1
    else:
        rows[path] = {'n': n, 'types': types, 'len': lenpart.group(0) if lenpart else None, 'dyn': dyn, 'merged': 1}
roles = Counter()
done = {}
for path, r in rows.items():
    rule = None
    for rid, rx, role, prov in P[prof]:
        if re.search(rx, path):
            rule = (rid, role, prov); break
    g = next((rid, role, prov) for rid, rx, role, prov in G if re.search(rx, path))
    if rule is None: rule = g
    elif rule[1] is None: rule = (rule[0] + '+' + g[0], g[1], rule[2])
    is_container = all(t in ('obj', 'list', 'null') for t in r['types']) and (not path.endswith('[]') or path == '$')
    scalar_elem = path.endswith('[]') and not any(t in ('obj', 'list') for t in r['types'])
    role = '모음' if is_container and rule[1] in ('값', '사건') and not path.endswith('[]') else rule[1]
    if scalar_elem:
        parent = done.get(path[:-2])
        role = '값' if parent in (None, '모음') else parent
    if path == '$': role = '모음'
    done[path] = role
    roles[(role, rule[2])] += 1
    ty = ','.join(f'{t}:{c}' for t, c in r['types'].items())
    extra = (' ' + r['len']) if r['len'] else ''
    mg = f' (merged {r["merged"]} levels)' if r['merged'] > 1 else ''
    print(f'{path}  n={r["n"]}  {ty}{extra}{r["dyn"]}{mg}  | {role} | {rule[2]} | {rule[0]}')
print('# role x provenance line counts:', dict(sorted(roles.items(), key=lambda x: -x[1])))
