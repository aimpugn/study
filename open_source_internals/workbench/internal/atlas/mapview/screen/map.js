// 지도 화면. 보기 모형(atlas-data)만 읽고 그립니다. 주제를 모르며, 띠·구역·칸·나무의 이름과 상태는 모두 보기 모형에서 옵니다.
// 그림 규칙은 docs/works/WORK_20261017_LEARNING_MAP_PLATFORM.md 「T4 지도 화면 v1」의 설계 메모와 고친 규칙을 따릅니다.
(function () {
  "use strict";
  const V = JSON.parse(document.getElementById("atlas-data").textContent);
  const RANGE = window.AtlasRange;
  const $ = (id) => document.getElementById(id);
  const SVGNS = "http://www.w3.org/2000/svg";
  const BAND = Object.fromEntries(V.bands.map((b, i) => [b.id, Object.assign({ order: i }, b)]));
  const ZONE = Object.fromEntries(V.zones.map((z) => [z.id, z]));
  const ZOOMS = ["system", "process", "function"];
  const ZOOM_NAME = { system: "시스템", process: "프로세스 안", function: "함수·커널" };
  const REL_NAME = { same: "같은 자리", same_leaves: "같은 잎 · 줌만 다름", contains: "포함함", inside: "안에 듦", overlap: "겹침", apart: "따로" };
  const STATE_NAME = { events: "사건 있음", other: "경로 밖 사건만(흐림)", watched: "봤고 0번", unseen: "안 봄(빗금)" };
  const NATURE_NAME = { observed: "", derived: "규칙으로 셈", inferred: "≈ 추정", human: "사람이 정함" };

  // 화면의 상태입니다. 해설은 접힌 채로 시작합니다(참가자 시험의 시작 상태).
  const S = {
    route: V.default, zoom: "system", step: null, node: null, start: null, end: null, startNode: null, endNode: null,
    lens: true, drawer: false, legend: true, ranges: [], now: null, playing: null, history: [], log: [], t0: Date.now(), cells: [],
  };

  // 시간과 길이를 사람이 읽는 글로 씁니다.
  function dur(ns) {
    if (ns == null) return "";
    if (ns >= 1e6) return (ns / 1e6).toFixed(1) + " ms";
    if (ns >= 1e3) return (ns / 1e3).toFixed(1) + " µs";
    return ns + " ns";
  }
  // 출처 성격의 표지 글입니다. 관찰은 표지가 없습니다.
  function badge(nature) {
    const t = NATURE_NAME[nature] || "";
    return t ? `<span class="badge ${nature}">${t}</span>` : "";
  }
  // 그림 안 글에 붙일 짧은 성격 표지입니다(관찰은 없음).
  function mark(nature) { return nature === "inferred" ? " ≈" : nature === "derived" ? " (셈)" : ""; }
  // 접힌 칸의 하강 수입니다. 아래가 기록된 것이 일부뿐이면 「×k 중 r 기록」으로 써서 기록되지 않은 하강을 기록된 것처럼 보이지 않습니다.
  function descentCount(d) { return d.state === "events" && d.recorded < d.count ? `×${d.count} 중 ${d.recorded} 기록` : `×${d.count}`; }
  function esc(s) { return String(s == null ? "" : s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c])); }
  function el(tag, attrs, parent) {
    const e = document.createElementNS(SVGNS, tag);
    for (const k in attrs) if (attrs[k] != null) e.setAttribute(k, attrs[k]);
    if (parent) parent.appendChild(e);
    return e;
  }
  function text(parent, x, y, s, cls, extra) {
    const t = el("text", Object.assign({ x, y, class: cls }, extra || {}), parent);
    t.textContent = s;
    return t;
  }
  function route() { return V.routes[S.route]; }
  function group() { return V.groups[route().group]; }
  function shareClass(share) { return share > 0.5 ? "hot" : share > 0.1 ? "warm" : "cool"; }
  function thickness(ns) { return 2 + 2.2 * Math.log(1 + ns / 1e6); }

  // 함수 줌에서 펼칠 칸입니다. 고른 칸이 없으면 지면 아래를 기록한 칸 가운데 가장 오래 머문 칸, 그것도 없으면 가장 오래 머문 칸입니다.
  function focusStep() {
    const r = route();
    if (S.step != null) return S.step;
    let best = 0, bestScore = -1;
    r.steps.forEach((s, i) => {
      const score = (s.descent && s.descent.state === "events" ? 1e15 : 0) + s.dur_ns + s.tree.length;
      if (score > bestScore) { best = i; bestScore = score; }
    });
    return best;
  }
  function shownNodes(st) { return st.tree.map((n, i) => Object.assign({ idx: i }, n)).filter((n) => !S.lens || n.visible); }

  // 지금 줌의 칸(범위와 재생의 단위)과 고를 수 있는 칸 번호 [lo, hi]입니다.
  function cellsOf() {
    const r = route();
    const a = S.start == null ? 0 : S.start, b = S.end == null ? r.steps.length - 1 : S.end;
    if (S.zoom === "system") {
      const cells = r.system.map((s) => ({ leaves: s.leaves, label: s.zones.map((z) => ZONE[z].name).join(" → ") }));
      let lo = r.system.findIndex((s) => s.steps[1] >= a), hi = -1;
      r.system.forEach((s, i) => { if (s.steps[0] <= b) hi = i; });
      return { cells, lo: Math.max(lo, 0), hi };
    }
    if (S.zoom === "process") {
      return { cells: r.steps.map((s) => ({ leaves: s.leaves, label: s.name })), lo: a, hi: b };
    }
    const st = r.steps[focusStep()];
    const nodes = shownNodes(st);
    let lo = 0, hi = nodes.length - 1;
    if (S.startNode != null && S.startNode.step === focusStep()) lo = Math.max(0, nodes.findIndex((n) => n.idx >= S.startNode.node));
    if (S.endNode != null && S.endNode.step === focusStep()) { hi = -1; nodes.forEach((n, i) => { if (n.idx <= S.endNode.node) hi = i; }); }
    return { cells: nodes.map((n) => ({ leaves: n.leaves, label: n.name, node: n.idx })), lo, hi };
  }

  // 조작 하나를 기록하고 주소의 # 상태를 고칩니다(참가자 시험의 조작 기록, 고친 규칙 10).
  function act(op, extra) {
    const entry = Object.assign({ t: Date.now() - S.t0, op, zoom: S.zoom, route: route().id, step: S.step, node: S.node }, extra || {});
    S.log.push(entry);
    const h = new URLSearchParams({ r: S.route, z: S.zoom, s: S.step == null ? "" : S.step, n: S.node == null ? "" : S.node, a: S.start == null ? "" : S.start, b: S.end == null ? "" : S.end });
    try { history.replaceState(null, "", "#" + h.toString()); } catch (e) { /* file:// 등에서 막히면 기록만 남깁니다 */ }
  }
  function readHash() {
    const h = new URLSearchParams(location.hash.slice(1));
    const num = (k) => (h.get(k) === null || h.get(k) === "" ? null : Number(h.get(k)));
    if (num("r") != null && V.routes[num("r")]) S.route = num("r");
    if (ZOOMS.includes(h.get("z"))) S.zoom = h.get("z");
    S.step = num("s"); S.node = num("n"); S.start = num("a"); S.end = num("b");
    const r = route();
    if (S.step != null && !r.steps[S.step]) S.step = null;
  }

  function setZoom(z, why) {
    if (z === S.zoom) return;
    S.history.push({ zoom: S.zoom, step: S.step, node: S.node });
    S.zoom = z; S.now = null; stopPlay();
    act("zoom", { why: why || "" });
    render();
  }
  function selectCell(i) {
    const c = S.cells[i];
    if (!c) return;
    if (S.zoom === "process") { S.step = i; S.node = null; }
    else if (S.zoom === "function") { S.node = c.node; S.step = focusStep(); }
    else { S.step = route().system[i].steps[0]; S.sys = i; }
    S.now = i;
    act("select", { cell: i });
    render();
  }
  function setRoute(i, why) {
    const old = route();
    S.route = i;
    const r = route();
    if (S.step != null && !r.steps[S.step]) S.step = null;
    if (S.start != null && S.start >= r.steps.length) S.start = null;
    if (S.end != null && S.end >= r.steps.length) S.end = null;
    if (old.group !== r.group) { S.start = S.end = S.step = S.node = null; S.startNode = S.endNode = null; S.ranges = []; }
    S.node = null; S.now = null; stopPlay();
    act("route", { why: why || "" });
    render();
  }

  // 재생: 지금 줌의 칸을 「지금」 고리가 차례로 지나갑니다.
  function stopPlay() { if (S.playing) { clearInterval(S.playing); S.playing = null; } }
  function stepNow(delta) {
    const { lo, hi } = cellsOf();
    if (S.now == null) S.now = delta > 0 ? lo : hi;
    else S.now = Math.min(hi, Math.max(lo, S.now + delta));
    act("replay", { now: S.now });
    render();
  }
  function play() {
    if (S.playing) { stopPlay(); act("pause"); render(); return; }
    const { lo, hi } = cellsOf();
    if (S.now == null || S.now >= hi) S.now = lo - 1;
    act("play");
    S.playing = setInterval(() => {
      const c = cellsOf();
      if (S.now >= c.hi) { stopPlay(); render(); return; }
      S.now++;
      render();
    }, 650);
    render();
  }

  // ---- 그리기 ----

  function render() {
    renderTop();
    renderLeft();
    renderCrumbs();
    const c = cellsOf();
    S.cells = c.cells;
    S.lo = c.lo; S.hi = c.hi;
    const box = $("canvas");
    box.innerHTML = "";
    const W = box.clientWidth || 900, H = box.clientHeight || 560;
    const svg = el("svg", { viewBox: `0 0 ${W} ${H}`, "data-zoom": S.zoom }, box);
    defs(svg);
    let xs = [];
    if (S.zoom === "system") xs = drawSystem(svg, W, H);
    else if (S.zoom === "process") xs = drawProcess(svg, W, H);
    else xs = drawFunction(svg, W, H);
    S.xs = xs;
    renderBrackets(xs);
    renderRight();
    renderDrawer();
    renderMinimap();
    document.querySelectorAll("#zoombar button").forEach((b) => b.classList.toggle("on", b.dataset.zoom === S.zoom));
    $("play").textContent = S.playing ? "⏸" : "▶";
    $("lens").checked = S.lens;
  }
  function defs(svg) {
    const d = el("defs", {}, svg);
    const p = el("pattern", { id: "hatch", width: 6, height: 6, patternUnits: "userSpaceOnUse", patternTransform: "rotate(45)" }, d);
    el("rect", { width: 6, height: 6, fill: "var(--panel)" }, p);
    el("line", { x1: 0, y1: 0, x2: 0, y2: 6, stroke: "var(--hatch)", "stroke-width": 2 }, p);
  }

  function renderTop() {
    const sel = $("group");
    if (!sel.options.length) {
      V.groups.forEach((g, i) => {
        const o = document.createElement("option");
        o.value = i;
        o.textContent = `${sceneLabel(g)} · ${g.label}${g.routes.length > 1 ? ` (후보 ${g.routes.length})` : ""}`;
        sel.appendChild(o);
      });
    }
    sel.value = route().group;
    $("typed").textContent = route().typed || "";
  }
  function sceneLabel(g) { const r = V.routes[g.routes[0]]; return r.scene_label; }

  function candName(i) {
    const r = V.routes[i];
    const g = V.groups[r.group];
    const same = g.routes.filter((j) => V.routes[j].run === r.run);
    const run = r.run.split("/").pop();
    let s = run;
    if (same.length > 1) s += ` · ${same.indexOf(i) === 0 ? "첫" : same.indexOf(i) + 1 + "번째"} 기록`;
    if (r.process && !same.length) s += ` · ${r.process}`;
    return s;
  }

  function renderLeft() {
    const r = route();
    const a = S.start == null ? 0 : S.start, b = S.end == null ? r.steps.length - 1 : S.end;
    $("startName").textContent = S.startNode ? `${r.steps[S.startNode.step].tree[S.startNode.node].name} (함수·커널)` : `${a + 1} · ${r.steps[a].name}`;
    $("endName").textContent = S.endNode ? `${r.steps[S.endNode.step].tree[S.endNode.node].name} (함수·커널)` : `${b + 1} · ${r.steps[b].name}`;

    // 장소로 고르기: 칸을 펼친 나무에 든 장소 이름마다 그 장소를 지난 칸들입니다.
    const pick = $("placePick");
    if (pick.dataset.route !== String(S.route)) {
      pick.dataset.route = String(S.route);
      pick.length = 1;
      const names = new Map();
      r.steps.forEach((s, i) => s.tree.forEach((n) => { if (!names.has(n.name)) names.set(n.name, new Set()); names.get(n.name).add(i); }));
      [...names.keys()].sort().forEach((nm) => {
        const o = document.createElement("option");
        o.value = nm; o.textContent = `${nm} (${names.get(nm).size}칸)`;
        pick.appendChild(o);
      });
      $("visits").innerHTML = "";
    }

    // 경로 후보
    const g = group();
    $("cands").innerHTML = g.routes.map((i) => {
      const x = V.routes[i];
      const total = x.end_ns - x.start_ns;
      const tops = x.steps.map((s) => s).sort((p, q) => q.dur_ns - p.dur_ns).slice(0, 3);
      const bar = tops.map((s) => `<span class="${shareClass(s.dur_ns / (total || 1))}" style="width:${Math.max(2, (100 * s.dur_ns) / (total || 1))}%;background:var(--${shareClass(s.dur_ns / (total || 1))})"></span>`).join("");
      const d = (r.divergence || []).find((v) => v.other === i);
      let div = "";
      if (i !== S.route && d) {
        const where = d.kind === "steps" ? `칸 ${d.step + 1} ${esc(d.name)}(프로세스 안)` : d.kind === "none" ? "견줄 나무 없음" : d.kind === "same" ? "나무의 구조와 시간이 같음" : `${esc(d.name)}(함수·커널, 칸 ${d.step + 1})`;
        const kind = d.kind === "tree" ? "갈리는 곳" : d.kind === "time" ? "구조 같음 · 시간 차 최대" : d.kind === "steps" ? "칸 차례가 갈리는 곳" : "";
        div = `<div class="div">${kind}: <b>${where}</b>${d.diff_ns ? ` · 차 ${dur(d.diff_ns)}` : ""} <span class="badge derived">규칙으로 셈</span> <button class="tiny" data-go="${i}">그 자리로</button></div>`;
      }
      return `<div class="cand ${i === S.route ? "on" : ""}" data-route="${i}"><div class="head"><span>${esc(candName(i))}</span><span class="ms">${dur(total)}</span></div>
        <div class="bar">${bar}</div><div class="tops">${tops.map((s) => `${esc(s.name)} ${dur(s.dur_ns)}${mark(s.dur_nature)}`).join(" · ")}</div>${div}</div>`;
    }).join("");

    // 단계별 안내
    $("guideZoom").textContent = ZOOM_NAME[S.zoom];
    const gd = $("guide");
    let html = "";
    if (S.zoom === "system") {
      r.system.forEach((s, i) => {
        const dim = s.steps[1] < a || s.steps[0] > b;
        html += `<div class="it ${S.now === i ? "on" : ""} ${dim ? "dim" : ""}" data-cell="${i}"><span class="n">${i + 1}</span><span>${s.zones.map((z) => esc(ZONE[z].name)).join(" → ")}</span><span class="t">${dur(s.dur_ns)}</span></div>`;
      });
    } else if (S.zoom === "process") {
      let col = null;
      r.steps.forEach((s, i) => {
        if (s.column && s.column !== col) {
          col = s.column;
          const c = (r.columns || []).find((x) => x.name === col);
          html += `<div class="col"><span>${esc(col)}</span><span>${c ? `${dur(c.dwell_ns)} ${badge(c.nature)}` : ""}</span></div>`;
        }
        const dim = i < a || i > b;
        const t = s.t0 == null ? "순서만" : "+" + dur(s.t0 - r.start_ns);
        html += `<div class="it ${S.step === i ? "on" : ""} ${dim ? "dim" : ""}" data-cell="${i}"><span class="n">${i + 1}</span><span>${esc(s.name)}${s.fold ? ` ×${s.fold} ${badge(s.fold_nature)}` : ""} ${badge(s.nature)}</span><span class="t">${t}</span></div>`;
      });
    } else {
      const st = r.steps[focusStep()];
      shownNodes(st).forEach((n, i) => {
        html += `<div class="it ${S.node === n.idx ? "on" : ""}" data-cell="${i}" style="padding-left:${4 + n.depth * 10}px"><span>${esc(n.name)}${n.folded ? ` <small>+${n.folded}</small>` : ""} ${badge(n.nature)}</span><span class="t">${dur(n.dur_ns)}</span></div>`;
      });
    }
    gd.innerHTML = html;
    const un = r.untimed || [];
    $("untimed").innerHTML = un.length ? `시각 없는 사건 ${un.length}개(칸이 되지 않음): ${un.map((u) => `${esc(u.name)}${u.text ? ` <code>${esc(String(u.text).slice(0, 40))}</code>` : ""}`).join(", ")}` : "";
  }

  function renderCrumbs() {
    const r = route();
    const parts = [["system", "시스템"]];
    if (S.zoom !== "system") parts.push(["process", r.process ? `${r.scene_label} ${r.process}` : r.scene_label]);
    if (S.zoom === "process" && S.step != null && r.steps[S.step].column) parts.push(["process", r.steps[S.step].column]);
    if (S.zoom === "function") {
      const st = r.steps[focusStep()];
      if (st.column) parts.push(["process", st.column]);
      parts.push(["function", st.name]);
    }
    $("crumbs").innerHTML = parts.map((p, i) => `${i ? '<span class="sep">›</span>' : ""}<button data-zoom="${p[0]}">${esc(p[1])}</button>`).join("");
  }

  // 띠 상태를 그림의 클래스로 옮깁니다.
  function stateOf(states, id) { return (states && states[id]) || "unseen"; }

  // ---- 시스템 줌: 구역 지도와 경로 ----
  function drawSystem(svg, W, H) {
    const r = route();
    const P = (z) => ({ x: 10 + (z.x / 100) * (W - 20), y: 10 + (z.y / 100) * (H - 20), w: (z.w / 100) * (W - 20), h: (z.h / 100) * (H - 20) });
    for (const z of V.zones) {
      const b = P(z);
      const st = stateOf(r.zones, z.id);
      el("rect", { x: b.x, y: b.y, width: b.w, height: b.h, rx: 10, class: `zone ${st}`, fill: `var(--zone-${z.id}, var(--panel2))`, "data-zone": z.id }, svg);
      text(svg, b.x + 8, b.y + 16, z.name, "zlabel");
      if (st !== "events") text(svg, b.x + 8, b.y + 30, STATE_NAME[st], "zsub");
    }
    const a = S.start == null ? 0 : S.start, bEnd = S.end == null ? r.steps.length - 1 : S.end;
    // 구역마다 닻 하나. 폭이 넓은 구역(가로로 긴 띠, 예: 경계)은 앞 점의 가로 자리에서 건너, 다리처럼 세로로 지나갑니다.
    const pathOf = (rt) => {
      let prev = null;
      return rt.system.map((s, k) => s.zones.map((zid) => {
        const b = P(ZONE[zid]);
        let x = b.x + b.w / 2 + ((k % 3) - 1) * 6;
        if (prev && ZONE[zid].w > 50) x = Math.min(b.x + b.w - 12, Math.max(b.x + 12, prev[0]));
        prev = [x, b.y + b.h / 2 + ((k % 3) - 1) * 6];
        return prev;
      }));
    };
    // 같은 무리의 다른 후보는 회색 선입니다(후보를 고르는 줌).
    for (const i of group().routes) {
      if (i === S.route) continue;
      const pts = pathOf(V.routes[i]).flat();
      el("polyline", { points: pts.map((p) => p.join(",")).join(" "), class: "route other" }, svg);
    }
    const segs = pathOf(r);
    const xs = [];
    segs.forEach((pts, k) => {
      const s = r.system[k];
      const prev = k > 0 ? segs[k - 1][segs[k - 1].length - 1] : null;
      const all = prev ? [prev, ...pts] : pts;
      const dim = s.steps[1] < a || s.steps[0] > bEnd;
      el("polyline", { points: all.map((p) => p.join(",")).join(" "), class: `route ${dim ? "dim" : ""}` }, svg);
    });
    segs.forEach((pts, k) => {
      const [x, y] = pts[0];
      const c = el("circle", { cx: x, cy: y, r: 9, class: `station ${S.now === k ? "now" : ""}`, "data-cell": k }, svg);
      c.style.cursor = "pointer";
      text(svg, x, y, String(k + 1), "stext").setAttribute("data-cell", k);
    });
    // 구역 지도는 가로가 시간이 아니라, 범위 줄은 시스템 칸을 같은 폭으로 늘어놓습니다.
    const n = r.system.length;
    for (let i = 0; i < n; i++) xs.push({ x0: (i * W) / n, x1: ((i + 1) * W) / n - 1 });
    return xs;
  }

  // 보이는 띠: 구역 상태가 기록 없음이 아닌 띠, 그리고 지면과 그 아래 띠(빗금을 보이려고 늘 그림)입니다.
  function visibleBands(states, zones, zoom) {
    return V.bands.filter((b) => {
      if (zoom === "process" && b.zoom !== "process") return false;
      if (b.ground || b.below) return zoom === "process" ? true : null;
      return zones[b.zone] !== "unseen";
    });
  }

  // ---- 프로세스 안 줌: 깊이 순 띠 위의 경로 한 줄 ----
  function drawProcess(svg, W, H) {
    const r = route();
    const LW = 118, top = r.columns && r.columns.length ? 44 : 22, bottom = 18;
    const bands = visibleBands(r.bands, r.zones, "process");
    // 띠 높이: 지면은 얇게, 나머지는 남는 높이를 나눕니다.
    const avail = H - top - bottom;
    const rowH = Math.min(30, (avail - 16) / Math.max(1, bands.length - 1));
    const Y = {};
    let y = top;
    for (const b of bands) {
      const h = b.ground ? 16 : rowH;
      Y[b.id] = { y, h, mid: y + h / 2 };
      el("rect", { x: 0, y, width: W, height: h - 1, class: `band ${stateOf(r.bands, b.id)}`, fill: b.ground ? "var(--zone-boundary)" : `var(--zone-${b.zone}, var(--panel2))`, "data-band": b.id }, svg);
      text(svg, 6, y + h / 2 + 3, b.name, "blabel");
      const st = stateOf(r.bands, b.id);
      if (st !== "events" && !b.ground) text(svg, 6, y + h / 2 + 12, STATE_NAME[st], "bstate");
      y += h;
    }
    // 칸의 가로 자리: 열(처리 단계)마다 칸을 놓고, 고른 칸의 열은 넓힙니다. 칸이 없는 열은 좁은 자리 하나입니다.
    const focusCol = S.step != null ? r.steps[S.step].column : null;
    const slots = [];
    const cols = r.columns && r.columns.length ? r.columns : [{ name: "", steps: [0, r.steps.length - 1] }];
    for (const c of cols) {
      if (c.steps[0] < 0) { slots.push({ col: c, w: 1.2, empty: true }); continue; }
      for (let i = c.steps[0]; i <= c.steps[1]; i++) slots.push({ col: c, step: i, w: focusCol && c.name === focusCol ? 4 : 1 });
    }
    const unit = (W - LW - 12) / slots.reduce((s, x) => s + x.w, 0);
    let x = LW;
    const xs = [];
    for (const s of slots) {
      s.x0 = x; s.x1 = x + s.w * unit; s.cx = (s.x0 + s.x1) / 2; x = s.x1;
      if (s.step != null) xs[s.step] = { x0: s.x0, x1: s.x1 - 1 };
    }
    // 열 머리: 처리 단계 이름과 머문 시간
    if (r.columns && r.columns.length) {
      for (const c of r.columns) {
        const ss = slots.filter((s) => s.col === c);
        const x0 = ss[0].x0 + 1, x1 = ss[ss.length - 1].x1 - 1;
        el("rect", { x: x0, y: 4, width: Math.max(4, x1 - x0), height: 30, rx: 8, class: `colpill ${focusCol === c.name ? "on" : ""}` }, svg);
        if (x1 - x0 > c.name.length * 11 + 8) {
          text(svg, (x0 + x1) / 2, 17, c.name, "coltext");
          text(svg, (x0 + x1) / 2, 29, dur(c.dwell_ns) + mark(c.nature), "coltime");
        } else el("title", {}, svg.lastChild).textContent = `${c.name} ${dur(c.dwell_ns)}${mark(c.nature)}`;
        if (c.steps[0] < 0 && x1 - x0 > 30) text(svg, (x0 + x1) / 2, top + 14, "이동 없음", "ctext muted", { "text-anchor": "middle" });
      }
    }
    const homeY = (Y[r.home] || Y[bands[0].id]).mid;
    const a = S.start == null ? 0 : S.start, bEnd = S.end == null ? r.steps.length - 1 : S.end;
    const total = r.end_ns - r.start_ns || 1;
    const ground = bands.find((b) => b.ground);
    // 지면 아래에서 칸의 하강이 닿는 띠: 펼친 나무에서 사건이 있는 지면 아래 띠(프로세스 안 줌의 띠로 합침).
    const belowRow = (st) => {
      const ids = bands.filter((b) => b.below && stateOf(st.bands, b.id) === "events");
      return ids.length ? ids[0] : bands.find((b) => b.below);
    };
    // 고른 칸·지금 칸의 배경
    if (S.step != null && xs[S.step]) el("rect", { x: xs[S.step].x0, y: top, width: xs[S.step].x1 - xs[S.step].x0, height: y - top, class: "selcol" }, svg);
    if (S.now != null && xs[S.now]) el("rect", { x: xs[S.now].x0, y: top, width: xs[S.now].x1 - xs[S.now].x0, height: y - top, class: "nowcol" }, svg);
    // 경로 한 줄(집 띠 위). 출발·도착 밖은 흐립니다.
    const pts = r.steps.map((s, i) => [ (xs[i].x0 + xs[i].x1) / 2, homeY ]);
    const seg = (i0, i1, cls) => { if (i1 >= i0) el("polyline", { points: pts.slice(i0, i1 + 1).map((p) => p.join(",")).join(" "), class: cls }, svg); };
    seg(0, a, "route dim"); seg(bEnd, r.steps.length - 1, "route dim"); seg(a, bEnd, "route");
    r.steps.forEach((s, i) => {
      const cx = pts[i][0], dim = i < a || i > bEnd;
      const g = el("g", { "data-cell": i, class: dim ? "dimmed" : "" }, svg);
      if (dim) g.setAttribute("opacity", "0.45");
      const share = s.dur_ns / total;
      const cls = shareClass(share);
      // 하강: 지면을 건너는 칸만 선으로 그립니다(고친 규칙 4). 추정 근거면 점선입니다.
      if (s.descent && ground) {
        const gy = Y[ground.id].mid, row = belowRow(s), ry = row ? Y[row.id] : null;
        el("line", { x1: cx, y1: homeY, x2: cx, y2: gy, class: `descent ${s.descent.nature}` }, g);
        el("rect", { x: cx - 5, y: gy - 5, width: 10, height: 10, transform: `rotate(45 ${cx} ${gy})`, class: "diamond" }, g);
        if (ry) {
          if (s.descent.state === "events") {
            el("line", { x1: cx, y1: gy, x2: cx, y2: ry.mid, class: `descent ${s.descent.nature}` }, g);
            const th = Math.max(8, thickness(s.dur_ns) * 2.2), wdt = Math.max(10, (xs[i].x1 - xs[i].x0) * 0.6);
            el("rect", { x: cx - wdt / 2, y: ry.mid - th / 2, width: wdt, height: th, rx: th / 2, class: `cap ${cls}` }, g);
            if (s.sleep_ns) {
              const f = s.sleep_ns / s.dur_ns;
              el("rect", { x: cx - wdt / 2 + wdt * (1 - f), y: ry.mid - th / 2 + 1, width: wdt * f - 1, height: th - 2, rx: (th - 2) / 2, class: `cap sleep ${cls}` }, g);
            }
          } else {
            // 기록하지 않은 층 안으로는 선을 긋지 않습니다(X1). 선은 지면에서 멈추고 그 아래는 빗금 칸만 둡니다.
            el("rect", { x: cx - 7, y: ry.y + 2, width: 14, height: ry.h - 5, class: "hatchcell" }, g);
          }
        }
        if (s.descent.count > 1) text(g, cx + 6, gy - 6, descentCount(s.descent), "mark");
      }
      // 목적지 띠의 점(짧은 왕복은 선 없이 점으로)
      const dest = Y[s.band];
      if (dest && s.band !== r.home) el("circle", { cx, cy: dest.mid, r: 3.5, class: `dot ${s.nature}` }, g);
      // 칸 표지: 접힌 칸은 「×N」 알약, 머문 시간이 있으면 두께로, 나머지는 작은 점
      if (s.fold) {
        const wdt = Math.max(22, (xs[i].x1 - xs[i].x0) * 0.8);
        el("rect", { x: cx - wdt / 2, y: homeY - 8, width: wdt, height: 16, rx: 8, class: `cap ${cls}` }, g);
        text(g, cx, homeY, `×${s.fold}${mark(s.fold_nature)}`, "fold");
      } else if (s.dur_ns > 0 && !(s.descent && s.descent.state === "events")) {
        const th = thickness(s.dur_ns) * 2, wdt = Math.max(8, (xs[i].x1 - xs[i].x0) * 0.5);
        el("rect", { x: cx - wdt / 2, y: homeY - th / 2, width: wdt, height: th, rx: th / 2, class: `cap ${cls}` }, g);
      } else {
        el("circle", { cx, cy: homeY, r: S.step === i ? 6 : 4, class: "station" }, g);
      }
      text(g, cx, y + 12, String(i + 1), "depth");
    });
    // 출발·도착 깃발(세로선 없이 머리 줄 위)
    const flag = (i, label, color) => { if (xs[i]) { const fx = (xs[i].x0 + xs[i].x1) / 2; el("circle", { cx: fx, cy: top - 4, r: 3, fill: color }, svg); text(svg, fx + 5, top - 1, label, "ctext"); } };
    if (r.columns && r.columns.length) { flag(a, "출발", "var(--flag-start)"); flag(bEnd, "도착", "var(--flag-end)"); }
    divergenceMark(svg, xs, "process", homeY);
    return xs;
  }

  // 지금 경로와 다른 후보가 갈리는 곳의 표지입니다. 비교할 후보는 같은 무리의 첫 다른 후보입니다.
  function divergenceMark(svg, xs, zoom, y) {
    const r = route();
    // 비교할 후보: 구조가 갈리는 후보, 그다음 시간만 다른 후보, 그다음 칸 차례가 다른 후보입니다.
    const ds = (r.divergence || []).filter((v) => v.kind !== "none" && v.kind !== "same");
    const d = ["tree", "time", "steps"].map((k) => ds.find((v) => v.kind === k)).find(Boolean);
    if (!d) return;
    const other = candName(d.other);
    if (zoom === "process" && xs[d.step]) {
      const cx = (xs[d.step].x0 + xs[d.step].x1) / 2;
      el("circle", { cx, cy: y, r: 11, class: "divmark" }, svg);
      text(svg, cx + 13, y - 12, d.kind === "steps" ? `${other}와 칸 차례가 갈림` : `${other}와 갈림 · ${dur(d.diff_ns)} ▸ 함수·커널`, "divtext");
    }
    if (zoom === "function" && d.step === focusStep() && d.node >= 0) {
      const i = S.cells.findIndex((c) => c.node === d.node);
      if (i >= 0 && xs[i]) {
        const cx = (xs[i].x0 + xs[i].x1) / 2;
        el("circle", { cx, cy: y(i), r: 12, class: "divmark" }, svg);
        text(svg, cx + 14, y(i) - 12, `${other}와 ${d.kind === "tree" ? "갈림" : "구조 같음 · 시간 차 최대"} · ${dur(d.diff_ns)}`, "divtext");
      }
    }
  }

  // ---- 함수·커널 줌: 칸 하나를 펼친 나무의 계단 ----
  function drawFunction(svg, W, H) {
    const r = route();
    const si = focusStep();
    const st = r.steps[si];
    const nodes = shownNodes(st);
    const used = new Set(st.tree.map((n) => n.band));
    const hasLayers = st.tree.some((n) => BAND[n.band] && BAND[n.band].zoom === "function");
    // 보이는 띠: 나무가 쓴 띠, 함수 줌 층을 쓰면 같은 합쳐 보일 띠 아래의 모든 층, 지면과 그 아래 프로세스 띠, 봤고 0번인 띠입니다.
    const under = new Set(st.tree.map((n) => BAND[n.band] && BAND[n.band].under).filter(Boolean));
    const bands = V.bands.filter((b) => {
      if (used.has(b.id)) return true;
      if (b.zoom === "function") return under.has(b.under);
      if (b.ground || b.below) return !(hasLayers && under.has(b.id));
      return stateOf(st.bands, b.id) === "watched";
    });
    const LW = 118, top = 26, bottom = 22;
    const rowH = Math.min(30, (H - top - bottom) / Math.max(1, bands.length));
    const Y = {};
    bands.forEach((b, k) => {
      const y = top + k * rowH;
      Y[b.id] = y + rowH / 2;
      const stt = stateOf(st.bands, b.id);
      el("rect", { x: 0, y, width: W, height: rowH - 1, class: `band ${stt}`, fill: b.ground ? "var(--zone-boundary)" : `var(--zone-${b.zone}, var(--panel2))`, "data-band": b.id }, svg);
      text(svg, 6, y + rowH / 2 + 3, b.name, "blabel");
      if (stt !== "events" && !b.ground) text(svg, 6, y + rowH / 2 + 12, STATE_NAME[stt], "bstate");
    });
    text(svg, LW, 16, `${st.name} · ${dur(st.dur_ns)}${mark(st.dur_nature)}${st.sleep_ns ? ` · 그중 잠듦 ${dur(st.sleep_ns)}` : ""}`, "ctext");
    const unit = (W - LW - 16) / Math.max(1, nodes.length);
    const xs = nodes.map((n, i) => ({ x0: LW + i * unit, x1: LW + (i + 1) * unit - 1 }));
    const yOf = (i) => (Y[nodes[i].band] != null ? Y[nodes[i].band] : top + rowH / 2);
    const lo = S.lo, hi = S.hi;
    if (S.node != null) { const i = nodes.findIndex((n) => n.idx === S.node); if (i >= 0) el("rect", { x: xs[i].x0, y: top, width: unit, height: rowH * bands.length, class: "selcol" }, svg); }
    if (S.now != null && xs[S.now]) el("rect", { x: xs[S.now].x0, y: top, width: unit, height: rowH * bands.length, class: "nowcol" }, svg);
    // 계단 줄기: 나무 차례대로 잇습니다. 칸과 시계가 다른 마디로 가는 줄은 점선입니다(가로 차례가 시각이 아님).
    for (let i = 1; i < nodes.length; i++) {
      const p = [(xs[i - 1].x0 + xs[i - 1].x1) / 2, yOf(i - 1)], q = [(xs[i].x0 + xs[i].x1) / 2, yOf(i)];
      const cls = `route ${i <= lo || i > hi ? "dim" : ""}`;
      const line = el("polyline", { points: `${p[0]},${p[1]} ${q[0]},${p[1]} ${q[0]},${q[1]}`, class: cls }, svg);
      if (nodes[i].nature === "inferred" || nodes[i].other_clock) line.setAttribute("stroke-dasharray", "6 4");
    }
    const base = st.dur_ns || 1;
    nodes.forEach((n, i) => {
      const cx = (xs[i].x0 + xs[i].x1) / 2, cy = yOf(i);
      const g = el("g", { "data-cell": i }, svg);
      if (i < lo || i > hi) g.setAttribute("opacity", "0.45");
      // 칸과 시계가 다른 마디는 칸 길이에 대한 몫을 셈하지 않습니다(다른 시계·다른 뜻의 길이라 100%를 넘을 수 있음).
      const share = n.dur_ns != null && !n.other_clock ? n.dur_ns / base : 0;
      if (n.sleep && n.dur_ns) {
        const wdt = Math.max(16, unit * 0.9);
        el("rect", { x: cx - wdt / 2, y: cy - 7, width: wdt, height: 14, rx: 7, class: `cap sleep ${shareClass(share)}` }, g);
        text(g, cx, cy + 20, "잠듦", "ctext", { "text-anchor": "middle" });
      } else {
        el("circle", { cx, cy, r: S.node === n.idx ? 7 : 5, class: `node ${n.nature} ${S.node === n.idx ? "sel" : ""}` }, g);
      }
      // 글은 보조입니다: 이름은 머문 시간이 칸의 5% 이상이거나 잠든 마디, 뿌리, 고른 마디에만 쓰고 나머지는 마우스를 올리면 보입니다.
      el("title", {}, g).textContent = `${n.name} ${dur(n.dur_ns)}`;
      if (share >= 0.05 || n.sleep || n.parent < 0 || S.node === n.idx || nodes.length <= 14) {
        const label = n.name.length > 22 ? n.name.slice(0, 21) + "…" : n.name;
        const lab = text(g, cx + 6, cy - 8, label, `nlabel ${share > 0.5 ? "hot" : ""}`);
        if (unit < 70) lab.setAttribute("transform", `rotate(-24 ${cx + 6} ${cy - 8})`);
      }
      if (n.dur_ns != null && share >= 0.05) text(g, cx + 6, cy + 13, dur(n.dur_ns), "ctext muted");
      if (n.folded && S.lens) { el("rect", { x: cx - 10, y: cy + 15, width: 20, height: 12, rx: 6, class: "plus" }, g); text(g, cx, cy + 21, `+${n.folded}`, "plustext"); }
      if (n.other_clock) text(g, cx - 4, cy + 30, "시계 다름", "mark");
      if (n.nature === "inferred") text(g, cx - 12, cy + 4, "≈", "mark");
      text(g, cx, H - 6, String(n.depth), "depth");
    });
    text(svg, 6, H - 6, "나무 깊이", "depth", { "text-anchor": "start" });
    divergenceMark(svg, xs, "function", yOf);
    return xs;
  }

  // ---- 범위 괄호 줄 ----
  function renderBrackets(xs) {
    const box = $("brackets");
    box.innerHTML = "";
    xs.forEach((c, i) => {
      if (!c) return;
      const d = document.createElement("div");
      d.className = "bcell";
      d.style.left = c.x0 + "px"; d.style.width = (c.x1 - c.x0) + "px";
      if (i < S.lo || i > S.hi) d.style.background = "repeating-linear-gradient(45deg, transparent 0 4px, var(--panel2) 4px 6px)";
      box.appendChild(d);
    });
    S.ranges.forEach((rg, k) => {
      if (rg.zoom !== S.zoom || rg.route !== S.route || (rg.zoom === "function" && rg.step !== focusStep())) return;
      const a = xs[rg.first], b = xs[rg.last];
      if (!a || !b) return;
      const d = document.createElement("div");
      d.className = "bracket" + (k === S.ranges.length - 1 ? "" : " prev");
      d.style.left = a.x0 + "px"; d.style.width = (b.x1 - a.x0) + "px";
      d.innerHTML = `<span class="lbl">${String.fromCharCode(65 + (k % 26))} ${rg.first + 1}–${rg.last + 1}</span>`;
      box.appendChild(d);
    });
  }

  // 드래그 → 칸 경계에 맞춘 범위(바깥쪽으로). 클릭은 칸 하나. 흐린 칸(출발·도착 밖)은 넣지 않습니다.
  function bindBrackets() {
    const box = $("brackets");
    let x0 = null, ghost = null;
    const px = (ev) => ev.clientX - box.getBoundingClientRect().left;
    box.addEventListener("mousedown", (ev) => {
      x0 = px(ev);
      ghost = document.createElement("div");
      ghost.className = "dragbox"; ghost.style.left = x0 + "px"; ghost.style.width = "0px";
      box.appendChild(ghost);
    });
    window.addEventListener("mousemove", (ev) => {
      if (x0 == null) return;
      const x = px(ev);
      ghost.style.left = Math.min(x0, x) + "px"; ghost.style.width = Math.abs(x - x0) + "px";
    });
    window.addEventListener("mouseup", (ev) => {
      if (x0 == null) return;
      const x1 = px(ev);
      const start = x0;
      x0 = null;
      if (ghost) ghost.remove();
      chooseRange(start, x1);
    });
  }
  function chooseRange(x0, x1) {
    const res = RANGE.normalize(S.xs.map((c) => c || { x0: -1, x1: -2 }), x0, x1, S.lo, S.hi);
    if (!res) return;
    const [first, last] = res;
    const rg = { zoom: S.zoom, route: S.route, step: S.zoom === "function" ? focusStep() : null, first, last, leaves: RANGE.rangeLeaves(S.cells.map((c) => c.leaves), first, last), drag: [x0, x1] };
    // 같은 자리를 다시 고르면 새 자리를 만들지 않고 그 자리를 엽니다(범위 해설의 규칙 3).
    const same = S.ranges.findIndex((p) => p.route === rg.route && RANGE.relate(rg, p) === "same");
    if (same >= 0) { const [p] = S.ranges.splice(same, 1); S.ranges.push(p); }
    else S.ranges.push(rg);
    act("range", { first, last, zoomOf: rg.zoom });
    S.drawer = S.drawer || false;
    render();
  }

  // ---- 오른쪽: 읽는 법, 고른 칸, 미니맵 ----
  function renderRight() {
    $("legend").className = "legend" + (S.legend ? "" : " closed");
    $("legendToggle").textContent = S.legend ? "접기" : "펼치기";
    $("legend").innerHTML = [
      ['<svg width="30" height="10"><line x1="0" y1="5" x2="30" y2="5" stroke="var(--route)" stroke-width="3"/></svg>', "고른 경로 한 줄"],
      ['<svg width="30" height="10"><line x1="0" y1="5" x2="30" y2="5" stroke="var(--other)" stroke-width="2"/></svg>', "다른 후보(시스템 줌만)"],
      ['<svg width="30" height="12"><circle cx="15" cy="6" r="3.5" fill="var(--dot)"/></svg>', "짧은 왕복: 목적지 띠의 점"],
      ['<svg width="30" height="16"><line x1="15" y1="0" x2="15" y2="16" stroke="var(--route)" stroke-width="2" stroke-dasharray="5 3"/></svg>', "지면을 건너는 하강(점선 = 추정 짝)"],
      ['<svg width="30" height="12"><rect x="2" y="1" width="26" height="10" fill="url(#hatchL)" stroke="var(--hatch)"/><defs><pattern id="hatchL" width="5" height="5" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><line x1="0" y1="0" x2="0" y2="5" stroke="var(--hatch)" stroke-width="2"/></pattern></defs></svg>', "빗금: 지켜본 기록이 없는 층(안 봄)"],
      ['<svg width="30" height="12"><rect x="2" y="1" width="26" height="10" fill="none" stroke="var(--watched)" stroke-dasharray="3 3"/></svg>', "점선 테두리: 봤고 0번"],
      ['<svg width="30" height="12"><rect x="2" y="2" width="26" height="8" rx="4" fill="var(--hot)"/></svg>', "두께 = 머문 시간, 색 = 경로 시간 몫(10%·50%)"],
      ['<svg width="30" height="12"><rect x="2" y="2" width="26" height="8" rx="4" fill="none" stroke="var(--hot)"/></svg>', "빈 칸 = 잠든 시간"],
      ["<b>×N</b>", "잇단 칸 접음(N = 줄 수)"],
      ['<span class="badge inferred">≈ 추정</span>', "틀릴 수 있는 짝짓기·짐작"],
      ['<span class="badge derived">규칙으로 셈</span>', "해석기가 기록에서 셈"],
    ].map(([g, t]) => `<span>${g}</span><span>${t}</span>`).join("");

    const r = route();
    const det = $("detail");
    let html = "";
    if (S.zoom === "system") {
      const i = S.now != null ? S.now : r.system.findIndex((s) => S.step != null && s.steps[0] <= S.step && S.step <= s.steps[1]);
      const s = r.system[i];
      if (s) html = `<div class="name">${s.zones.map((z) => esc(ZONE[z].name)).join(" → ")}</div><div class="kv"><span class="k">칸</span><span>프로세스 안 ${s.steps[0] + 1}–${s.steps[1] + 1}</span><span class="k">머문 시간</span><span>${dur(s.dur_ns)}</span><span class="k">성격</span><span>${badge(s.nature) || "관찰"}</span></div><div class="acts"><button data-zoomto="process" data-step="${s.steps[0]}">프로세스 안으로</button></div>`;
      else html = `<div class="muted">시스템 칸(번호)을 고르면 그 구간이 보입니다.</div>`;
    } else if (S.zoom === "process") {
      const s = S.step != null ? r.steps[S.step] : null;
      if (s) {
        const fixed = V.fixed.bands[s.band] || V.fixed.names[s.place] || "";
        html = `<div class="name">${S.step + 1} · ${esc(s.name)}${s.fold ? ` ×${s.fold}` : ""} ${badge(s.nature)}</div>
          <div class="kv"><span class="k">띠</span><span>${esc(BAND[s.band].name)}</span><span class="k">자리</span><span class="mono">${esc(s.place)}</span>
          ${s.loc ? `<span class="k">코드</span><span class="mono">${esc(s.loc)}</span>` : ""}
          ${s.column ? `<span class="k">처리 단계</span><span>${esc(s.column)}</span>` : ""}
          <span class="k">시각</span><span>${s.t0 == null ? "순서만 기록" : "+" + dur(s.t0 - r.start_ns)}</span>
          <span class="k">머문 시간</span><span>${dur(s.dur_ns)} ${badge(s.dur_nature)}${s.sleep_ns ? ` (잠듦 ${dur(s.sleep_ns)}, ${Math.round((100 * s.sleep_ns) / s.dur_ns)}%)` : ""}</span>
          ${s.count > 1 ? `<span class="k">횟수</span><span>원래 이동 ${s.count}번</span>` : ""}
          ${s.descent ? `<span class="k">지면 아래</span><span>${s.descent.state === "events" ? "기록됨" : "안 봄(빗금)"}${s.descent.count > 1 ? ` · ${descentCount(s.descent)}` : ""} ${badge(s.descent.nature)}</span>` : ""}
          </div>
          ${fixed ? `<div class="fixed"><small>고정 설명 · ${esc(V.fixed.source)}</small><br>${esc(fixed)}</div>` : ""}
          ${s.stack ? `<div class="stack">${s.stack.slice(0, 7).map(esc).join("<br>")}</div>` : ""}
          ${(s.attrs || []).slice(0, 8).map((a) => `<div class="mono"><small>${esc(a.key)}</small> ${esc(a.value)} ${badge(a.nature)}</div>`).join("")}
          <div class="acts"><button data-end="start">출발로</button><button data-end="end">도착으로</button>${s.tree.length > 1 ? `<button data-zoomto="function">함수·커널로</button>` : ""}</div>`;
      } else html = `<div class="muted">칸(번호나 점)을 누르면 그 사건이 보입니다. 위 줄을 드래그하면 범위를 고릅니다.</div>`;
    } else {
      const st = r.steps[focusStep()];
      const n = S.node != null ? st.tree[S.node] : null;
      if (n) {
        const fixed = V.fixed.names[n.name] || V.fixed.bands[n.band] || "";
        html = `<div class="name">${esc(n.name)} ${badge(n.nature)}</div>
          <div class="kv"><span class="k">층</span><span>${esc(BAND[n.band].name)}</span>
          ${n.loc ? `<span class="k">코드</span><span class="mono">${esc(n.loc)}</span>` : ""}
          <span class="k">머문 시간</span><span>${dur(n.dur_ns)}${n.dur_ns && st.dur_ns && !n.other_clock ? ` · 칸의 ${Math.round((100 * n.dur_ns) / st.dur_ns)}%` : ""}</span>
          <span class="k">깊이</span><span>${n.depth}${n.via && n.via !== "parent" ? ` · ${esc(n.via)} 연결로 옴` : ""}</span>
          ${n.sleep ? `<span class="k">상태</span><span>잠듦(CPU를 내줌)</span>` : ""}
          ${n.other_clock ? `<span class="k">시계</span><span>칸과 다른 시계라 가로 차례는 시각이 아님</span>` : ""}
          ${n.folded ? `<span class="k">접힌 호출</span><span>${n.folded}개(렌즈)</span>` : ""}</div>
          ${fixed ? `<div class="fixed"><small>고정 설명 · ${esc(V.fixed.source)}</small><br>${esc(fixed)}</div>` : ""}
          <div class="acts"><button data-endnode="start">출발로</button><button data-endnode="end">도착으로</button></div>`;
      } else html = `<div class="muted">${esc(st.name)}을(를) 펼친 나무입니다. 마디를 누르면 자세히 보입니다.</div>`;
    }
    det.innerHTML = html;
  }

  function renderDrawer() {
    const d = $("drawer");
    d.classList.toggle("open", S.drawer);
    $("drawerToggle").textContent = S.drawer ? "접기" : "펼치기";
    const r = route();
    const cur = S.ranges[S.ranges.length - 1];
    $("drawerSummary").innerHTML = `<small>해설 ${S.drawer ? "펼침" : "접힘"} · 이 세션에서 고른 범위 ${S.ranges.length}개 · 공급자 없음</small>`;
    if (!S.drawer) { $("drawerBody").innerHTML = ""; return; }
    let rels = "";
    if (cur) {
      rels = S.ranges.slice(0, -1).filter((p) => p.route === cur.route).map((p, k) => `<div class="rel"><span class="k">${String.fromCharCode(65 + (k % 26))} ${ZOOM_NAME[p.zoom]} ${p.first + 1}–${p.last + 1}</span><span>${REL_NAME[RANGE.relate(cur, p)]}</span></div>`).join("") || `<div class="muted">견줄 다른 범위가 없습니다.</div>`;
    }
    const sel = S.zoom === "process" && S.step != null ? r.steps[S.step] : null;
    const fixed = sel ? (V.fixed.bands[sel.band] || V.fixed.names[sel.place] || "") : "";
    $("drawerBody").innerHTML = `
      <div><h3>AI 해설</h3><div class="none">공급자 설정 없음. 이 화면은 정적 재생본이라 공급자를 부르지 않고 「받기」 단추가 없습니다(T5에서 붙임).<br>${cur ? `고른 범위: ${ZOOM_NAME[cur.zoom]} ${cur.first + 1}–${cur.last + 1}, 잎 사건 ${cur.leaves.length}개. 받은 해설이 없어 해설 자리는 아직 없습니다(범위 해설의 규칙 6).` : "위 줄을 드래그해 범위를 고르세요."}</div></div>
      <div><h3>고정 설명</h3>${fixed ? esc(fixed) + `<br><small>${esc(V.fixed.source)}</small>` : '<div class="muted">고른 칸의 고정 설명이 없습니다.</div>'}</div>
      <div><h3>이 범위와 다른 범위</h3>${rels || '<div class="muted">범위를 고르지 않았습니다.</div>'}</div>`;
  }

  function renderMinimap() {
    const r = route();
    const box = $("minimap");
    box.innerHTML = "";
    const W = box.clientWidth || 250, H = 118;
    const svg = el("svg", { viewBox: `0 0 ${W} ${H}` }, box);
    defs(svg);
    const P = (z) => ({ x: 2 + (z.x / 100) * (W - 4), y: 2 + (z.y / 100) * (H - 24), w: (z.w / 100) * (W - 4), h: (z.h / 100) * (H - 24) });
    const step = S.step != null ? r.steps[S.step] : null;
    const here = step ? BAND[step.band].zone : null;
    for (const z of V.zones) {
      const b = P(z);
      el("rect", { x: b.x, y: b.y, width: b.w, height: b.h, rx: 3, class: `zone ${stateOf(r.zones, z.id)}`, fill: `var(--zone-${z.id}, var(--panel2))`, stroke: z.id === here ? "var(--sel)" : null, "stroke-width": z.id === here ? 2 : null }, svg);
    }
    // 아래 띠: 시스템 칸들, 지금 자리 칠함
    const n = r.system.length;
    r.system.forEach((s, i) => {
      const on = step ? s.steps[0] <= S.step && S.step <= s.steps[1] : S.zoom === "system" && S.now === i;
      el("rect", { x: 2 + (i * (W - 4)) / n, y: H - 16, width: (W - 4) / n - 1, height: 10, fill: on ? "var(--route)" : "var(--panel2)", stroke: "var(--rule)" }, svg);
    });
    const where = [ZOOM_NAME[S.zoom], step ? step.column || "" : "", step ? step.name : ""].filter(Boolean).join(" · ");
    box.title = "지금 · " + where;
  }

  // ---- 이벤트 ----
  function bind() {
    $("group").addEventListener("change", (e) => setRoute(V.groups[Number(e.target.value)].routes[0], "group"));
    $("cands").addEventListener("click", (e) => {
      const go = e.target.closest("[data-go]");
      if (go) {
        const d = route().divergence.find((v) => v.other === Number(go.dataset.go));
        if (d && d.step >= 0) { S.step = d.step; S.node = d.node >= 0 ? d.node : null; act("divergence", { other: d.other }); setZoom(d.kind === "steps" ? "process" : "function", "divergence"); }
        return;
      }
      const c = e.target.closest("[data-route]");
      if (c) setRoute(Number(c.dataset.route), "candidate");
    });
    $("guide").addEventListener("click", (e) => { const it = e.target.closest("[data-cell]"); if (it) selectCell(Number(it.dataset.cell)); });
    $("canvas").addEventListener("click", (e) => {
      const g = e.target.closest("[data-cell]");
      if (g) selectCell(Number(g.getAttribute("data-cell")));
    });
    $("canvas").addEventListener("dblclick", () => { if (S.zoom === "system") setZoom("process", "dblclick"); else if (S.zoom === "process" && S.step != null) setZoom("function", "dblclick"); });
    document.querySelectorAll("#zoombar button").forEach((b) => b.addEventListener("click", () => setZoom(b.dataset.zoom, "zoombar")));
    $("crumbs").addEventListener("click", (e) => { const b = e.target.closest("[data-zoom]"); if (b) setZoom(b.dataset.zoom, "crumb"); });
    $("back").addEventListener("click", () => { const h = S.history.pop(); if (h) { S.zoom = h.zoom; S.step = h.step; S.node = h.node; act("back"); render(); } });
    $("detail").addEventListener("click", (e) => {
      const b = e.target.closest("button");
      if (!b) return;
      if (b.dataset.zoomto) { if (b.dataset.step) S.step = Number(b.dataset.step); setZoom(b.dataset.zoomto, "detail"); return; }
      if (b.dataset.end) {
        if (b.dataset.end === "start") { S.start = S.step; S.startNode = null; if (S.end != null && S.end < S.start) [S.start, S.end] = [S.end, S.start]; }
        else { S.end = S.step; S.endNode = null; if (S.start != null && S.end < S.start) [S.start, S.end] = [S.end, S.start]; }
        act("endpoint", { which: b.dataset.end }); render(); return;
      }
      if (b.dataset.endnode) {
        const v = { step: focusStep(), node: S.node };
        if (b.dataset.endnode === "start") { S.startNode = v; S.start = v.step; } else { S.endNode = v; S.end = v.step; }
        if (S.start != null && S.end != null && S.end < S.start) { [S.start, S.end] = [S.end, S.start]; [S.startNode, S.endNode] = [S.endNode, S.startNode]; }
        act("endpoint", { which: b.dataset.endnode, node: S.node }); render();
      }
    });
    $("swap").addEventListener("click", () => { [S.start, S.end] = [S.end, S.start]; [S.startNode, S.endNode] = [S.endNode, S.startNode]; if (S.start != null && S.end != null && S.end < S.start) { [S.start, S.end] = [S.end, S.start]; [S.startNode, S.endNode] = [S.endNode, S.startNode]; } act("swap"); render(); });
    $("resetEnds").addEventListener("click", () => { S.start = S.end = null; S.startNode = S.endNode = null; act("reset-ends"); render(); });
    $("placePick").addEventListener("change", (e) => {
      const nm = e.target.value;
      const r = route();
      const steps = r.steps.map((s, i) => (s.tree.some((n) => n.name === nm) ? i : -1)).filter((i) => i >= 0);
      $("visits").innerHTML = steps.map((i, k) => `<button data-visit="${i}">${k + 1}번째 · 칸 ${i + 1}</button>`).join("");
      act("place", { place: nm });
    });
    $("visits").addEventListener("click", (e) => { const b = e.target.closest("[data-visit]"); if (b) { S.step = Number(b.dataset.visit); if (S.zoom === "system") S.zoom = "process"; act("visit", { step: S.step }); render(); } });
    $("first").addEventListener("click", () => { S.now = null; stepNow(1); });
    $("prev").addEventListener("click", () => stepNow(-1));
    $("next").addEventListener("click", () => stepNow(1));
    $("play").addEventListener("click", play);
    $("lens").addEventListener("change", (e) => { S.lens = e.target.checked; S.now = null; act("lens", { on: S.lens }); render(); });
    $("drawerToggle").addEventListener("click", () => { S.drawer = !S.drawer; act("drawer", { open: S.drawer }); render(); });
    $("legendToggle").addEventListener("click", () => { S.legend = !S.legend; render(); });
    $("theme").addEventListener("click", () => {
      const cur = document.documentElement.dataset.theme || (matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
      document.documentElement.dataset.theme = cur === "dark" ? "light" : "dark";
      try { localStorage.setItem("atlas-theme", document.documentElement.dataset.theme); } catch (e) { /* 저장을 못 하면 이번 화면에만 */ }
      render();
    });
    $("oplog").addEventListener("click", () => {
      const blob = new Blob([JSON.stringify({ route: route().id, log: S.log }, null, 1)], { type: "application/json" });
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob); a.download = "atlas-oplog.json"; a.click();
    });
    window.addEventListener("resize", () => render());
    bindBrackets();
  }

  try { const t = localStorage.getItem("atlas-theme"); if (t) document.documentElement.dataset.theme = t; } catch (e) { /* 없으면 시스템 설정 */ }
  readHash();
  bind();
  act("open");
  render();
  // 화면 검사(check-map.mjs)가 상태를 읽고 조작하는 자리입니다.
  window.__atlas = { S, V, render, setZoom, setRoute, selectCell, chooseRange, play, stepNow, cellsOf, focusStep };
})();
