// 지도 화면 검사. node는 화면 검사에만 씁니다(작업 원칙). 플랫폼 검사라 주제를 모르고, 장면은 보기 모형에서 찾습니다.
//
//   node internal/atlas/mapview/screen/check-map.mjs <화면 파일.html> [사진 폴더]
//
// 1) 범위 대조: Go 정본(mapview/ranges.go)이 쓴 대조표 testdata/ranges.json을 range.js로 다시 셈해 모두 같은지 봅니다(브라우저 없이).
// 2) 화면 조작: headless Chromium으로 화면을 열어 줌, 후보 고르기, 출발·도착, 범위 드래그, 재생, 빗금 띠, 조작 기록을 조작하고 상태를 읽습니다.
//    playwright가 없으면 이 부분을 건너뛰지 않고 실패합니다.
// 사진 폴더를 주면 1440×900으로 세 줌과, 지면 아래를 기록하지 않은 경로의 함수 줌을 찍습니다. 다르면 차이를 쓰고 종료 코드 1입니다.
import { readFileSync } from "node:fs";
import { createRequire } from "node:module";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import vm from "node:vm";

const here = dirname(fileURLToPath(import.meta.url));
const [html, shots] = process.argv.slice(2);
if (!html) {
  console.error("쓰는 법: node check-map.mjs <화면 파일.html> [사진 폴더]");
  process.exit(2);
}
const fails = [];
const check = (ok, msg) => { if (!ok) fails.push(msg); };

// 1) 범위 대조
const sandbox = {};
vm.runInNewContext(readFileSync(join(here, "range.js"), "utf8"), sandbox);
const R = sandbox.AtlasRange;
const fx = JSON.parse(readFileSync(join(here, "..", "testdata", "ranges.json"), "utf8"));
let n = 0;
for (const c of fx.normalize) {
  const got = R.normalize(c.cells || [], c.x0, c.x1, c.lo, c.hi);
  const want = c.want.length ? c.want : null;
  check(JSON.stringify(got) === JSON.stringify(want), `normalize ${JSON.stringify([c.x0, c.x1, c.lo, c.hi])}: ${JSON.stringify(got)} ≠ Go ${JSON.stringify(want)}`);
  n++;
}
for (const c of fx.relate) {
  const got = R.relate({ zoom: c.a.zoom, leaves: c.a.leaves || [] }, { zoom: c.b.zoom, leaves: c.b.leaves || [] });
  check(got === c.want, `relate ${JSON.stringify(c.a)} ${JSON.stringify(c.b)}: ${got} ≠ Go ${c.want}`);
  n++;
}
for (const c of fx.leaves) {
  const got = R.rangeLeaves(c.cell_leaves.map((l) => l || []), c.first, c.last);
  check(JSON.stringify(got) === JSON.stringify(c.want || []), `rangeLeaves ${c.first}–${c.last}: ${JSON.stringify(got)} ≠ Go ${JSON.stringify(c.want)}`);
  n++;
}
console.log(`범위 대조 ${n}건`);

// 2) 화면 조작
let chromium;
try {
  const require = createRequire(join(process.execPath, "..", "..", "lib", "node_modules", "_"));
  ({ chromium } = require("playwright"));
} catch (e) {
  fails.push("playwright를 찾지 못해 화면 조작 검사를 못 함(npm i -g playwright)");
}
if (chromium) {
  const browser = await chromium.launch();
  try {
    const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
    const errs = [];
    page.on("pageerror", (e) => errs.push(String(e)));
    page.on("console", (m) => { if (m.type() === "error") errs.push(m.text()); });
    await page.goto("file://" + resolve(html));
    await page.waitForFunction(() => window.__atlas);
    const st = () => page.evaluate(() => { const S = window.__atlas.S; return { zoom: S.zoom, route: S.route, step: S.step, node: S.node, start: S.start, end: S.end, now: S.now, drawer: S.drawer, ranges: S.ranges.map((r) => ({ zoom: r.zoom, first: r.first, last: r.last, n: r.leaves.length })), log: S.log.map((l) => l.op) }; });

    // 시작 상태: 시스템 줌, 해설 접힘, 조작 기록에 open
    let s = await st();
    check(s.zoom === "system" && s.drawer === false && s.log[0] === "open", `시작 상태 ${JSON.stringify(s)}`);

    // 후보가 넷 이상인 무리(같은 요청의 다른 기록)로 가서 둘째 후보를 카드로 고릅니다.
    const g = await page.evaluate(() => window.__atlas.V.groups.findIndex((x) => x.routes.length >= 4));
    check(g >= 0, "후보 넷 이상인 무리가 없음");
    await page.selectOption("#group", String(g));
    const second = await page.evaluate((g) => window.__atlas.V.groups[g].routes[1], g);
    await page.click(`.cand[data-route="${second}"]`);
    s = await st();
    check(s.route === second, `후보 카드를 눌러도 경로가 ${s.route}(기대 ${second})`);
    check((await page.evaluate(() => location.hash)).includes(`r=${second}`), "주소의 # 상태에 경로가 없음");

    // 프로세스 안 줌: 칸을 고르고 출발·도착을 정하면 그 밖의 칸이 흐려집니다.
    await page.click('#zoombar button[data-zoom="process"]');
    const cells = await page.evaluate(() => window.__atlas.S.cells.length);
    await page.click('#guide [data-cell="1"]');
    await page.click('#detail button[data-end="start"]');
    await page.click(`#guide [data-cell="${Math.min(cells - 1, 3)}"]`);
    await page.click('#detail button[data-end="end"]');
    s = await st();
    check(s.start === 1 && s.end === Math.min(cells - 1, 3), `출발·도착 ${s.start}·${s.end}`);
    check((await page.locator('#canvas g[opacity="0.45"]').count()) > 0 || cells <= 4, "출발·도착 밖 칸이 흐려지지 않음");

    // 범위 드래그: 둘째 칸의 가운데에서 셋째 칸의 앞쪽 조금까지 → 바깥쪽으로 맞춰 [1, 2] 칸
    const xs = await page.evaluate(() => window.__atlas.S.xs);
    const box = await page.locator("#brackets").boundingBox();
    const y = box.y + box.height / 2;
    await page.mouse.move(box.x + (xs[1].x0 + xs[1].x1) / 2, y);
    await page.mouse.down();
    await page.mouse.move(box.x + xs[2].x0 + 1, y, { steps: 4 });
    await page.mouse.up();
    s = await st();
    const last = s.ranges[s.ranges.length - 1];
    check(last && last.first === 1 && last.last === 2 && last.n > 0, `드래그 범위 ${JSON.stringify(last)}(기대 1–2, 잎 > 0)`);
    // 흐린 칸(출발 앞)까지 끌어도 들어가지 않습니다.
    await page.mouse.move(box.x + 2, y); await page.mouse.down(); await page.mouse.move(box.x + xs[1].x1 - 1, y, { steps: 3 }); await page.mouse.up();
    s = await st();
    check(s.ranges[s.ranges.length - 1]?.first === 1, `흐린 칸이 범위에 들어감 ${JSON.stringify(s.ranges)}`);
    // 같은 자리를 다시 고르면 새 자리를 만들지 않습니다.
    const before = s.ranges.length;
    await page.mouse.move(box.x + (xs[1].x0 + xs[1].x1) / 2, y); await page.mouse.down(); await page.mouse.move(box.x + xs[2].x0 + 1, y, { steps: 3 }); await page.mouse.up();
    s = await st();
    check(s.ranges.length === before, `같은 범위를 다시 골라 자리가 늘어남 ${before} → ${s.ranges.length}`);
    // 해설 서랍: 공급자 없음 자리 표시와 범위 관계
    await page.click("#drawerToggle");
    const drawer = await page.textContent("#drawerBody");
    check(drawer.includes("공급자 설정 없음") && /(같은 자리|포함함|안에 듦|겹침|따로|같은 잎)/.test(drawer), "서랍에 공급자 없음 표시나 범위 관계가 없음");

    // 재생: 지금 고리가 칸을 차례로 지나갑니다.
    await page.click("#play");
    await page.waitForTimeout(1500);
    s = await st();
    check(s.now != null && s.now >= 1, `재생 뒤 지금 칸 ${s.now}`);
    await page.click("#play");

    // 커널까지 기록된 경로: 함수·커널 줌에 커널 층 띠가 사건 있음으로 보입니다.
    const kernelRoute = await page.evaluate(() => window.__atlas.V.routes.findIndex((r) => r.steps.some((s) => s.descent && s.descent.state === "events")));
    check(kernelRoute >= 0, "커널까지 기록된 경로가 없음");
    await page.evaluate((i) => { window.__atlas.setRoute(i); window.__atlas.setZoom("function"); }, kernelRoute);
    const layers = await page.locator('#canvas rect.band.events[data-band]').count();
    check(layers >= 3, `함수·커널 줌의 사건 있는 띠 ${layers}개`);
    // 추정 표시(고친 규칙 6): 추정 근거로 내려간 칸은 프로세스 안 줌에서 하강선이 점선 class(inferred)이고, 그 칸의 길이가 추정이면 함수 줌 머리 글에 ≈가 붙습니다.
    const infStep = await page.evaluate((i) => window.__atlas.V.routes[i].steps.findIndex((s) => s.descent && s.descent.state === "events" && s.dur_nature === "inferred"), kernelRoute);
    if (infStep >= 0) {
      const head = await page.evaluate((k) => { const A = window.__atlas; A.S.step = k; A.render(); return A.V.routes[A.S.route].steps[k].name + " · "; }, infStep);
      const heads = (await page.locator("#canvas text").allTextContents()).filter((t) => t.startsWith(head));
      check(heads.length === 1 && heads[0].includes("≈"), `길이가 추정인 칸의 함수 줌 머리에 ≈가 없음: ${JSON.stringify(heads)}`);
      await page.evaluate(() => window.__atlas.setZoom("process"));
      check((await page.locator("#canvas line.descent.inferred").count()) > 0, "추정 근거 하강선이 점선(inferred)으로 그려지지 않음");
    }
    // 기록하지 않은 하강(X1): 선은 지면에서 멈추고 빗금 칸만 있습니다. 그 칸의 하강선은 집 띠 → 지면 하나뿐입니다.
    const unseenStep = await page.evaluate((i) => window.__atlas.V.routes[i].steps.findIndex((s) => s.descent && s.descent.state === "unseen"), kernelRoute);
    if (unseenStep >= 0) {
      await page.evaluate(() => window.__atlas.setZoom("process"));
      check((await page.locator(`#canvas g[data-cell="${unseenStep}"] line.descent`).count()) === 1, "기록하지 않은 층 안으로 하강선이 들어감");
      check((await page.locator(`#canvas g[data-cell="${unseenStep}"] rect.hatchcell`).count()) === 1, "기록하지 않은 하강에 빗금 칸이 없음");
    } else fails.push("커널 경로에 기록하지 않은 하강이 없어 X1 하강 검사를 못 함");

    // 기록이 끝나는 층: 지면 아래를 하나도 기록하지 않은 경로는 함수 줌에서 그 띠가 빗금이고, 봤고 0번 띠가 있으면 점선 테두리입니다.
    const ended = await page.evaluate(() => window.__atlas.V.routes.findIndex((r) => r.steps.length && Object.entries(r.bands).every(([b, v]) => !window.__atlas.V.bands.find((x) => x.id === b).below || v === "unseen") && r.steps.some((s) => Object.values(s.bands).includes("watched"))));
    check(ended >= 0, "지면 아래를 기록하지 않고 봤고 0번 띠가 있는 경로가 없음");
    if (ended >= 0) {
      await page.evaluate((i) => { const A = window.__atlas; A.setRoute(i); const r = A.V.routes[i]; A.S.step = r.steps.findIndex((s) => Object.values(s.bands).includes("watched")); A.setZoom("process"); A.setZoom("function"); }, ended);
      check((await page.locator("#canvas rect.band.unseen").count()) > 0, "기록하지 않은 층의 빗금이 없음");
      check((await page.locator("#canvas rect.band.watched").count()) > 0, "봤고 0번 띠가 없음");
      if (shots) await page.screenshot({ path: join(shots, "screen-ended-function.png") });
    }

    // 테마와 조작 기록
    await page.click("#theme");
    check(["light", "dark"].includes(await page.evaluate(() => document.documentElement.dataset.theme)), "테마 단추가 듣지 않음");
    s = await st();
    for (const op of ["route", "zoom", "endpoint", "range", "play"]) check(s.log.includes(op), `조작 기록에 ${op}이 없음`);
    check(errs.length === 0, `화면 오류: ${errs.join(" | ")}`);

    if (shots) {
      // 기본 경로를 어두운 테마로 세 줌 찍습니다(시안과 견주기).
      await page.evaluate(() => { document.documentElement.dataset.theme = "dark"; const A = window.__atlas; A.S.ranges = []; A.S.start = A.S.end = null; A.S.drawer = false; A.setRoute(A.V.default); });
      for (const z of ["system", "process", "function"]) {
        await page.evaluate((z) => { const A = window.__atlas; A.S.step = null; A.S.node = null; A.S.now = null; if (A.S.zoom === z) A.render(); else A.setZoom(z); }, z);
        await page.waitForTimeout(150);
        await page.screenshot({ path: join(shots, `screen-${z}.png`) });
      }
    }
  } finally {
    await browser.close();
  }
}

if (fails.length) {
  console.error("화면 검사 실패:\n- " + fails.join("\n- "));
  process.exit(1);
}
console.log("화면 검사 통과: 범위 대조와 화면 조작");
