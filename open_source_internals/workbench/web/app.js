// 오픈소스 내부 작업대의 화면 코드입니다. 실시간 실행과 재생본이 이 파일 하나를 함께 씁니다.
//
// 자료(Bundle)의 모양은 Go의 internal/replay/bundle.go가 정합니다. 이 파일은 그 자료를 그리기만 하고,
// 자료에 없는 소스 줄이나 설명을 지어내지 않습니다.
// 화면에 넣는 글자는 모두 textContent로 넣습니다. 소스에는 <, >, &가 흔하고, SQL은 학습자가 입력한 문자열이기 때문입니다.
(() => {
  'use strict';

  // 단계 목록의 고정 치수입니다. app.css의 --row, --lane, --seq와 같아야 경로 그림이 행과 점에 맞습니다.
  const ROW = 30;
  const LANE = 22;
  const SEQ = 36;

  const $ = (id) => document.getElementById(id);

  const state = {
    mode: 'replay', // 'replay'는 파일에 담긴 자료, 'live'는 서버에서 받은 자료
    bundle: null,
    index: 0, // 지금 보는 단계(0부터)
    frame: null, // 호출 스택의 칸을 보는 중이면 그 칸 번호, 아니면 null
    columns: [], // 이 기록에 실제로 나온 층만, 층 순서대로
    subjects: [],
  };

  /**
   * 요소 하나를 만듭니다. 문자열 자식은 글자로만 들어가므로 HTML로 해석되지 않습니다.
   * @param {string} tag 태그 이름
   * @param {Object} attrs 속성. on으로 시작하면 이벤트 처리기, false나 null이면 생략합니다.
   * @param {...(Node|string|number)} children 자식
   * @returns {HTMLElement}
   */
  function h(tag, attrs, ...children) {
    const node = document.createElement(tag);
    for (const [k, v] of Object.entries(attrs || {})) {
      if (v === false || v == null) continue;
      if (k === 'class') node.className = v;
      else if (k.startsWith('on')) node.addEventListener(k.slice(2), v);
      else node.setAttribute(k, v === true ? '' : String(v));
    }
    for (const c of children.flat()) {
      if (c == null || c === false) continue;
      node.append(c instanceof Node ? c : document.createTextNode(String(c)));
    }
    return node;
  }

  /** 페이지를 준비합니다. 파일에 자료가 들어 있으면 재생 모드, 없으면 서버에 붙는 실시간 모드입니다. */
  function init() {
    document.addEventListener('keydown', onKey);
    window.addEventListener('hashchange', () => selectFromHash());
    $('prev').addEventListener('click', () => go(state.index - 1));
    $('next').addEventListener('click', () => go(state.index + 1));
    $('frame-back').addEventListener('click', () => {
      state.frame = null;
      renderCurrent();
    });

    const embedded = $('bundle');
    if (embedded) {
      state.mode = 'replay';
      show(JSON.parse(embedded.textContent), []);
    } else {
      state.mode = 'live';
      initLive();
    }
  }

  /** 서버에서 주제 목록을 받아 실행 양식을 채웁니다. 실패하면 이유를 화면에 보입니다. */
  async function initLive() {
    $('run-form').hidden = false;
    $('run-form').addEventListener('submit', onRun);
    try {
      const data = await request('/api/subjects');
      state.subjects = data.subjects;
      const subj = data.subjects[0];
      $('title').textContent = `${subj.title} 작업대`;
      document.title = $('title').textContent;
      $('watch').replaceChildren(...subj.watches.map((w) => h('option', { value: w.id }, w.label)));
      $('watch').addEventListener('change', renderExamples);
      renderExamples();
      renderLabStatus(subj);
      $('provenance').textContent = subj.sourceNotice || '';
    } catch (err) {
      showMessage(`서버에서 주제 목록을 받지 못했습니다. ${err.message}`);
    }
  }

  /**
   * 서버에 요청하고 JSON 응답을 돌려줍니다. 서버의 다른 사이트 요청 방어가 요구하는 머리글을 붙입니다.
   * @param {string} url
   * @param {Object} [body] 있으면 POST로 보냅니다.
   * @returns {Promise<Object>} 응답 JSON
   * @throws {Error} 응답이 실패면 서버가 보낸 이유를 담습니다.
   */
  async function request(url, body) {
    const init = { headers: { 'X-Workbench': '1' } };
    if (body) {
      init.method = 'POST';
      init.headers['Content-Type'] = 'application/json';
      init.body = JSON.stringify(body);
    }
    const res = await fetch(url, init);
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || `서버 응답 ${res.status}`);
    return data;
  }

  /** 고른 감시 목록의 예시 SQL을 버튼으로 보입니다. 누르면 입력칸에 들어갑니다. */
  function renderExamples() {
    const subj = state.subjects[0];
    const watch = subj.watches.find((w) => w.id === $('watch').value);
    const box = $('examples');
    box.replaceChildren('예시: ');
    for (const sql of watch.examples || []) {
      box.append(h('button', { type: 'button', onclick: () => { $('sql').value = sql; $('sql').focus(); } }, sql));
    }
    if (!$('sql').value && watch.examples && watch.examples[0]) $('sql').value = watch.examples[0];
  }

  /** 실습 컨테이너가 켜져 있는지와, 꺼져 있으면 켜는 명령을 보입니다. */
  function renderLabStatus(subj) {
    const box = $('lab-status');
    if (subj.running) {
      box.replaceChildren(`실습 컨테이너 ${subj.container}이(가) 켜져 있습니다. SQL은 이 컨테이너 안에서만 실행됩니다.`);
      return;
    }
    box.replaceChildren(
      `실습 컨테이너 ${subj.container}이(가) 꺼져 있습니다. 저장소 루트에서 `,
      h('code', null, subj.startCommand),
      '을 실행한 뒤 다시 실행하세요.',
    );
  }

  /** 실행 양식을 보내고, 받은 기록을 화면에 그립니다. */
  async function onRun(ev) {
    ev.preventDefault();
    const subj = state.subjects[0];
    const button = $('run');
    button.disabled = true;
    button.textContent = '추적하는 중';
    hideMessage();
    try {
      const data = await request('/api/runs', { subject: subj.id, watch: $('watch').value, sql: $('sql').value });
      show(data.bundle, data.warnings || []);
    } catch (err) {
      showMessage(err.message);
    } finally {
      button.disabled = false;
      button.textContent = '실행하고 추적';
    }
    // 실행이 성공했든 실패했든 컨테이너 상태를 다시 받아, 실행 사이에 켜거나 끈 상태가 화면에 남지 않게 합니다.
    try {
      const data = await request('/api/subjects');
      renderLabStatus(data.subjects[0]);
    } catch (_) {
      // 상태를 다시 받지 못하면 이전 상태 문구를 그대로 둡니다. 실행 결과나 오류 문구는 이미 보이고 있습니다.
    }
  }

  /**
   * 실행 한 번의 자료를 화면 전체에 그리고, 주소의 #step-N 또는 첫 단계를 고릅니다.
   * @param {Object} bundle internal/replay.Bundle
   * @param {string[]} warnings 서버가 붙인 주의 문구
   */
  function show(bundle, warnings) {
    state.bundle = bundle;
    state.frame = null;
    state.columns = usedColumns(bundle);

    if (state.mode === 'replay') {
      $('title').textContent = `${bundle.subjectTitle} 실행 재생`;
      document.title = $('title').textContent;
      $('provenance').textContent = bundle.sourceNotice || '';
    }
    renderOrigin(bundle.meta);
    $('summary').hidden = false;
    $('main').hidden = false;
    $('summary-sql').textContent = bundle.meta.sql;

    renderLegend();
    renderSteps();
    renderCounts();
    renderOutput();

    const notes = [...warnings];
    if (bundle.noBreak && bundle.noBreak.length) {
      notes.push(`gdb가 중단점을 만들지 못한 함수가 있습니다: ${bundle.noBreak.join(', ')}.`);
    }
    if (!bundle.steps.length) {
      notes.push('기록에 멈춘 단계가 하나도 없습니다. 감시 목록의 함수를 이 SQL이 지나지 않았을 수 있습니다.');
    }
    if (notes.length) showMessage(notes.join(' '));
    else hideMessage();

    // 단계가 없으면 앞선 실행의 소스 창과 위치가 남아 이번 기록처럼 보이므로, 소스 창을 숨기고 위치를 0으로 둡니다.
    const empty = !bundle.steps.length;
    $('code-pane').hidden = empty;
    if (empty) {
      $('position').textContent = '0 / 0';
      $('prev').disabled = true;
      $('next').disabled = true;
      history.replaceState(null, '', window.location.pathname + window.location.search);
      return;
    }
    // 주소의 #step-N은 재생본을 처음 열 때만 따릅니다. 실시간 실행에서는 그 값이 앞선 실행의 단계 번호이기 때문입니다.
    if (state.mode === 'replay' && selectFromHash()) return;
    go(0);
  }

  /** 기록의 출처, 감시 목록, 두 번 실행된다는 사실을 제목 옆에 밝힙니다. */
  function renderOrigin(meta) {
    const box = $('origin');
    box.replaceChildren(`출처: ${meta.origin} · 감시 목록: ${meta.watchLabel}`);
    if (meta.recordedAt) box.append(` · ${meta.recordedAt}`);
    if (meta.guide) box.append(' · ', h('a', { href: meta.guide }, '이 실행을 설명하는 문서'));
    box.append(h('br'), '이 SQL은 두 번 실행됩니다. 첫 번째는 캐시를 데우는 실행이고, 아래 기록은 두 번째 실행의 것입니다.');
  }

  /**
   * 이 기록의 단계에 실제로 나온 층만 층 순서대로 골라 그림의 세로줄로 씁니다.
   * @returns {{lane: number, label: string}[]}
   */
  function usedColumns(bundle) {
    const used = new Set(bundle.steps.map((s) => s.place.lane));
    return bundle.lanes.filter((l) => used.has(l.index)).map((l) => ({ lane: l.index, label: l.label }));
  }

  /** 세로줄 번호와 층 이름을 범례로 보입니다. */
  function renderLegend() {
    $('lane-legend').replaceChildren(
      ...state.columns.map((c, i) => h('li', null, h('b', null, i + 1), c.label)),
    );
  }

  /** 단계 목록과, 단계가 층 사이를 오가는 경로 그림을 그립니다. */
  function renderSteps() {
    const box = $('steps');
    const steps = state.bundle.steps;
    const col = new Map(state.columns.map((c, i) => [c.lane, i]));
    const track = state.columns.length * LANE + 8;
    box.style.setProperty('--track', `${track}px`);

    // 경로 그림은 행 뒤에 깔리는 SVG 하나입니다. 행 높이가 고정이라 점의 세로 위치를 계산할 수 있습니다.
    const ns = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(ns, 'svg');
    svg.setAttribute('width', String(SEQ + track));
    svg.setAttribute('height', String(steps.length * ROW));
    svg.setAttribute('aria-hidden', 'true');
    const line = document.createElementNS(ns, 'polyline');
    line.setAttribute('points', steps.map((s, i) => `${SEQ + col.get(s.place.lane) * LANE + LANE / 2},${i * ROW + ROW / 2}`).join(' '));
    svg.append(line);

    const rows = steps.map((s, i) => h('button', {
      type: 'button',
      class: 'step',
      role: 'option',
      'aria-selected': 'false',
      title: `분류: ${s.place.layer} · ${s.place.loc.path || '위치 모름'}:${s.place.loc.line || ''}`,
      onclick: () => go(i),
    },
    h('span', { class: 'seq' }, s.seq),
    h('span', { class: 'track' }, dot(col.get(s.place.lane))),
    h('span', { class: 'name' }, s.place.function)));
    box.replaceChildren(svg, ...rows);
  }

  /** 세로줄 c에 놓이는 점 하나를 만듭니다. */
  function dot(c) {
    const d = h('span', { class: 'dot' });
    d.style.left = `${c * LANE + LANE / 2}px`;
    return d;
  }

  /** END 줄의 함수별 횟수를 표로 보입니다. */
  function renderCounts() {
    const rows = (state.bundle.counts || []).map((c) => h('tr', null, h('td', null, c.function), h('td', null, `${c.hits}회`)));
    $('counts').replaceChildren(...rows);
  }

  /** psql 출력이 있으면 보이고, 없으면 그 칸을 숨깁니다. */
  function renderOutput() {
    const out = state.bundle.output || [];
    $('output-box').hidden = out.length === 0;
    $('output').textContent = out.join('\n');
  }

  /** 단계 i로 옮겨 갑니다. 범위를 넘으면 가장 가까운 끝 단계로 맞춥니다. */
  function go(i) {
    if (!state.bundle || !state.bundle.steps.length) return;
    state.index = Math.max(0, Math.min(state.bundle.steps.length - 1, i));
    state.frame = null;
    renderCurrent();
  }

  /**
   * 주소의 #step-N을 읽어 그 단계로 옮겨 갑니다.
   * @returns {boolean} 주소에 단계가 있었으면 true
   */
  function selectFromHash() {
    const m = /^#step-(\d+)$/.exec(window.location.hash);
    if (!m || !state.bundle) return false;
    go(Number(m[1]) - 1);
    return true;
  }

  /** 지금 단계(또는 호출 스택의 칸)를 소스 창, 값, 호출 스택에 그립니다. */
  function renderCurrent() {
    const b = state.bundle;
    const step = b.steps[state.index];
    const rows = $('steps').querySelectorAll('.step');
    rows.forEach((r, i) => r.setAttribute('aria-selected', String(i === state.index)));
    keepVisible($('steps'), rows[state.index]);

    $('position').textContent = `${state.index + 1} / ${b.steps.length}`;
    $('prev').disabled = state.index === 0;
    $('next').disabled = state.index === b.steps.length - 1;

    const place = state.frame == null ? step.place : step.stack[state.frame];
    $('fn').textContent = place.function;
    $('loc').textContent = place.loc.path ? `${place.loc.path}:${place.loc.line}` : 'gdb가 소스 위치를 알지 못한 프레임';
    $('layer').textContent = `분류: ${place.layer}`;
    $('layer').title = '작업대가 소스 디렉터리를 보고 붙인 이름입니다. 소스가 정한 이름이 아닙니다.';
    $('frame-note').hidden = state.frame == null;
    if (state.frame != null) {
      $('frame-text').textContent = `${step.seq}번 단계의 호출 스택 ${state.frame}번 칸을 보는 중입니다.`;
    }

    renderCode(place);
    renderValues(step);
    renderStack(step);
    history.replaceState(null, '', `#step-${state.index + 1}`);
  }

  /** 소스 창에 위치 place의 구간을 그리고, 멈춘 줄이 보이도록 창 안에서만 스크롤합니다. */
  function renderCode(place) {
    const box = $('code');
    box.replaceChildren();
    const ranges = (place.window && place.window.ranges) || [];
    if (!ranges.length) {
      box.append(h('div', { class: 'empty' }, place.loc.path ? '이 위치의 소스 구간을 계산하지 못했습니다.' : 'gdb가 이 프레임의 소스 위치를 알지 못했습니다.'));
      return;
    }
    const lines = state.bundle.files[place.loc.path] || {};
    let current = null;
    let prevTo = 0;
    ranges.forEach((r, k) => {
      if (k > 0) box.append(h('div', { class: 'gap' }, `${prevTo + 1}~${r.from - 1}번 줄 생략`));
      for (let n = r.from; n <= r.to; n++) {
        const isCurrent = n === place.loc.line;
        const cls = ['ln', isCurrent ? 'current' : '', n === place.window.defLine ? 'def' : ''].filter(Boolean).join(' ');
        const text = lines[n];
        const row = h('div', { class: cls, 'data-n': n, 'aria-current': isCurrent ? 'true' : null },
          h('span', { class: 'no' }, n),
          h('span', { class: 'tx' }, text === undefined ? '(이 자료에 없는 줄)' : text));
        if (isCurrent) current = row;
        box.append(row);
      }
      prevTo = r.to;
    });
    if (current) box.scrollTop = Math.max(0, current.offsetTop - box.clientHeight / 3);
  }

  /** gdb가 멈춘 순간 읽은 값을 식마다 보입니다. 식으로 나눌 수 없었던 값은 기록 원문 그대로 보입니다. */
  function renderValues(step) {
    const rows = [];
    for (const v of step.values || []) rows.push(h('tr', null, h('td', null, v.expr), h('td', null, v.text)));
    if (step.rawValues) rows.push(h('tr', null, h('td', null, '기록 원문'), h('td', null, step.rawValues)));
    $('values').replaceChildren(...rows);
    $('values').hidden = rows.length === 0;
    $('no-values').hidden = rows.length !== 0;
  }

  /** 이 단계에 호출 스택이 기록되어 있으면 칸마다 버튼으로 보입니다. */
  function renderStack(step) {
    const stack = step.stack || [];
    $('stack-box').hidden = stack.length === 0;
    $('stack').replaceChildren(...stack.map((f, d) => h('li', null, h('button', {
      type: 'button',
      'aria-current': state.frame === d ? 'true' : null,
      onclick: () => {
        state.frame = d;
        renderCurrent();
      },
    },
    h('span', { class: 'd' }, d),
    h('span', { class: 'f' }, f.function, ' ', h('small', null, f.loc.path ? `${f.loc.path}:${f.loc.line} · 분류: ${f.layer}` : '위치 모름'))))));
  }

  /** 목록 box 안에서 row가 보이도록 box만 스크롤합니다. 페이지 전체는 움직이지 않습니다. */
  function keepVisible(box, row) {
    if (!row) return;
    const top = row.offsetTop;
    if (top < box.scrollTop || top + row.offsetHeight > box.scrollTop + box.clientHeight) {
      box.scrollTop = Math.max(0, top - box.clientHeight / 2);
    }
  }

  /** 오른쪽·왼쪽 화살표(또는 j·k)로 단계를 넘깁니다. 입력칸에 있을 때는 가로채지 않습니다. */
  function onKey(ev) {
    const tag = (ev.target && ev.target.tagName) || '';
    if (['INPUT', 'SELECT', 'TEXTAREA'].includes(tag) || ev.metaKey || ev.ctrlKey || ev.altKey) return;
    const moves = { ArrowRight: 1, j: 1, ArrowLeft: -1, k: -1 };
    if (ev.key in moves) {
      ev.preventDefault();
      go(state.index + moves[ev.key]);
    } else if (ev.key === 'Home') {
      ev.preventDefault();
      go(0);
    } else if (ev.key === 'End') {
      ev.preventDefault();
      go(Infinity);
    }
  }

  /** 주의나 오류 문구를 보입니다. */
  function showMessage(text) {
    $('message').textContent = text;
    $('message').hidden = false;
  }

  /** 주의나 오류 문구를 숨깁니다. */
  function hideMessage() {
    $('message').hidden = true;
  }

  init();
})();
