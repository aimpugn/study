// Command check-screens 는 pgsql 주제의 자료 파일이 화면 파일이 들고 있던 상수·글과 같은지 봅니다(설계 메모 8과 9의 검사 6).
//
// 화면(concept-3·4, 트랜잭션 실험과 Spring 템플릿, gdb 작업대)은 데이터 literal 밖에 구성도·장면 이름·고정 설명을 상수로 들고
// 있었습니다. 그 사실은 이제 이 폴더의 layout.json(구성도), scenes.json(장면 목록), fixed.json(고정 설명)이 가집니다.
// 이 명령은 두 방향을 봅니다.
//
//  1. 자료 파일의 값마다 화면 파일의 선언을 JavaScript 엔진으로 평가해 같은지 견줍니다. 색 같은 표시 방식은 PROJECT의 규칙대로
//     빼고 견줍니다. 선언이 아닌 고정 글("text")은 화면 파일에 글자 그대로 있는지 봅니다.
//  2. 화면 파일의 대문자 상수 선언마다 자료 파일에 있거나 EXCLUDED에 까닭과 함께 있는지 봅니다. 새 상수가 조용히 빠지지 않게 합니다.
//
// 쓰는 법(workbench 폴더에서):
//
//	node topics/pgsql/check-screens.mjs                         # 검사. 다르면 차이를 쓰고 종료 코드 1
//	node topics/pgsql/check-screens.mjs --print <파일> <이름>...  # 평가하고 PROJECT를 적용한 값만 씀(자료 파일을 처음 만들 때)
//
// JavaScript literal을 가장 정확히 읽는 것은 JavaScript 엔진이라 이 검사만 node로 합니다(작업대 도구는 Go라는 원칙의 예외, 검증 전용).
// 정규식은 {"$regex": 원문, "$flags": 깃발}로 씁니다. 함수는 자료가 아니라 옮기지 않습니다.
import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import vm from "node:vm";

const here = path.dirname(fileURLToPath(import.meta.url));
const workbench = path.resolve(here, "../..");

// DATA_FILES 는 화면에서 옮긴 사실을 가진 자료 파일입니다. 모양은 {"<workbench 기준 화면 파일>": {"<이름>": 값, "text": [글...]}}입니다.
const DATA_FILES = ["layout.json", "scenes.json", "fixed.json"];

// PROJECT 는 견주기 전에 화면의 값에서 뺄 표시 방식입니다(설계 메모 8: 색 값과 기호는 기록이나 소스에 대한 사실이 아님).
const PROJECT = {
  "design/concept-4.html": {
    // 메시지마다 [이름, 색]이라 이름만 옮깁니다.
    MSG: (v) => Object.fromEntries(Object.entries(v).map(([dir, m]) => [dir, Object.fromEntries(Object.entries(m).map(([t, x]) => [t, x[0]]))])),
  },
  "design/spring-view.tmpl.html": {
    // 칸의 색(CSS 변수 이름)만 뺍니다. 폭과 곁 칸 표시는 구성도의 배치라 옮깁니다.
    "TOPIC.lanes": (v) => v.map(({ color, ...rest }) => rest),
  },
};

// EXCLUDED 는 화면 파일의 대문자 상수 가운데 옮기지 않는 것과 그 까닭입니다. 여기에도 자료 파일에도 없는 상수는 검사가 실패합니다.
const EXCLUDED = {
  "design/concept-3.html": {
    DATA: "기록 데이터(여정 적재기가 옮김)",
    MSGCOLOR: "색(표시 방식)",
  },
  "design/concept-4.html": {
    QUERIES: "기록 데이터(여정 적재기가 옮김)",
    DATA: "QUERIES로 셈함",
  },
  "design/tx-view.tmpl.html": {
    RUNS: "기록 데이터(트랜잭션 실험 적재기가 옮김)",
    SCS: "RUNS의 열쇠로 셈함",
    ROW: "그림의 줄 높이(표시 방식)",
  },
  "design/spring-view.tmpl.html": {
    RUNS: "기록 데이터(Spring 적재기가 옮김)",
    LANES: "TOPIC.lanes로 셈함",
    PROC: "TOPIC.lanes로 셈함",
    SCS: "RUNS의 열쇠로 셈함",
    WHYGLYPH: "렌즈 기호(표시 방식)",
    WHYCOLOR: "렌즈 색(표시 방식)",
    SCRUB: "시간 막대의 크기(표시 방식)",
    "TOPIC.storeKey": "브라우저 저장 열쇠(표시 방식)",
    "TOPIC.effect": "함수(화면 규칙)",
    "TOPIC.effectDetail": "함수(화면 규칙)",
    "TOPIC.rows": "함수(화면 규칙)",
    "TOPIC.intervals": "함수(화면 규칙)",
    "TOPIC.effGlyph": "함수(화면 규칙)",
    "TOPIC.board": "함수(화면 규칙)",
    "TOPIC.ganttRows": "함수(화면 규칙)",
    "TOPIC.outcomeChips": "함수(화면 규칙)",
    "TOPIC.startRow": "함수(화면 규칙)",
    "TOPIC.runFlaw": "함수(화면 규칙)",
    "TOPIC.moments": "함수(화면 규칙)",
  },
  "web/app.js": {
    ROW: "그림의 줄 높이(표시 방식)",
    LANE: "그림의 칸 폭(표시 방식)",
    SEQ: "그림의 번호 칸 폭(표시 방식)",
  },
};

// OBJECTS 는 속성마다 따로 옮기는 상수입니다. 속성 하나하나가 자료 파일이나 EXCLUDED에 있어야 합니다.
const OBJECTS = { "design/spring-view.tmpl.html": ["TOPIC"] };

const sources = new Map();

/**
 * 화면 파일 하나의 글을 한 번만 읽어 돌려줍니다.
 * @param {string} file workbench 기준 경로
 * @returns {string}
 */
function source(file) {
  if (!sources.has(file)) sources.set(file, readFileSync(path.join(workbench, file), "utf8"));
  return sources.get(file);
}

/**
 * 줄 처음(들여쓰기 허용)의 `const <이름> = `로 시작하는 선언의 초기식 글을 돌려줍니다.
 * 함수 안의 선언(MASK, PART)도 받으려고 들여쓰기를 허용하므로, 같은 이름의 선언이 둘 이상이면 어느 것인지 가를 수 없어 오류입니다.
 * 끝은 글자로 세지 않고, 선언 뒤의 `;` 후보마다 그 앞까지를 식으로 읽어 보아 처음으로 문법이 맞는 자리를 씁니다.
 * 그래서 문자열·정규식 안의 `;`에 속지 않습니다.
 * @param {string} file workbench 기준 경로
 * @param {string} name 상수 이름
 * @returns {string} 초기식 글
 * @throws {Error} 선언이 없거나 여럿이거나 식으로 읽히는 끝이 없을 때
 */
function initializer(file, name) {
  const text = source(file);
  const found = [...text.matchAll(new RegExp(`^[ \\t]*const ${name} = `, "gm"))];
  if (found.length !== 1) throw new Error(`${file}: ${name} 선언 ${found.length}개`);
  const start = found[0].index + found[0][0].length;
  for (let i = text.indexOf(";", start); i >= 0; i = text.indexOf(";", i + 1)) {
    const src = text.slice(start, i);
    try {
      new vm.Script(`(${src}\n)`);
      return src;
    } catch {
      // 아직 식이 끝나지 않았습니다. 다음 `;`까지 늘립니다.
    }
  }
  throw new Error(`${file}: ${name}의 식으로 읽히는 끝이 없음`);
}

/**
 * JSON으로 옮길 수 없는 값을 표지로 바꿉니다. 정규식은 원문과 깃발로, 함수는 {"$function": true}로 씁니다.
 * @param {unknown} v
 * @returns {unknown}
 */
function plain(v) {
  if (Object.prototype.toString.call(v) === "[object RegExp]") return { $regex: v.source, $flags: v.flags };
  if (typeof v === "function") return { $function: true };
  if (Array.isArray(v)) return v.map(plain);
  if (v && typeof v === "object") return Object.fromEntries(Object.entries(v).map(([k, x]) => [k, plain(x)]));
  return v;
}

/**
 * 화면 파일의 상수 하나(또는 "상수.속성")를 평가하고 PROJECT를 적용한 값을 돌려줍니다.
 * 상수는 빈 맥락에서 평가합니다. 다른 상수를 읽는 초기식은 실패하므로, 그런 상수는 옮길 값이 아니라 셈 규칙입니다(EXCLUDED).
 * @param {string} file workbench 기준 경로
 * @param {string} key 상수 이름 또는 "상수.속성"
 * @returns {unknown}
 */
function evaluate(file, key) {
  const [name, prop] = key.split(".");
  let v = vm.runInContext(`(${initializer(file, name)}\n)`, vm.createContext({}), { filename: `${file}#${name}` });
  if (prop !== undefined) {
    if (!Object.hasOwn(v, prop)) throw new Error(`${file}: ${name}에 속성 ${prop}이 없음`);
    v = v[prop];
  }
  const project = PROJECT[file]?.[key];
  return plain(project ? project(v) : v);
}

/**
 * 두 JSON 값의 차이를 경로와 함께 모읍니다. 객체는 열쇠 차례를 보지 않고, 목록은 차례까지 봅니다.
 * @param {unknown} want 화면에서 평가한 값
 * @param {unknown} got 자료 파일의 값
 * @param {string} at 지금 경로
 * @param {string[]} out 차이를 쌓을 곳
 */
function diff(want, got, at, out) {
  if (Array.isArray(want) || Array.isArray(got)) {
    if (!Array.isArray(want) || !Array.isArray(got)) return void out.push(`${at}: 목록과 목록 아닌 것`);
    if (want.length !== got.length) out.push(`${at}: 길이 화면 ${want.length}, 자료 ${got.length}`);
    for (let i = 0; i < Math.min(want.length, got.length); i++) diff(want[i], got[i], `${at}[${i}]`, out);
    return;
  }
  if (want && got && typeof want === "object" && typeof got === "object") {
    for (const k of new Set([...Object.keys(want), ...Object.keys(got)])) {
      if (!Object.hasOwn(got, k)) out.push(`${at}.${k}: 자료에 없음`);
      else if (!Object.hasOwn(want, k)) out.push(`${at}.${k}: 화면에 없음`);
      else diff(want[k], got[k], `${at}.${k}`, out);
    }
    return;
  }
  if (want !== got) out.push(`${at}: 화면 ${JSON.stringify(want)}, 자료 ${JSON.stringify(got)}`);
}

/**
 * 화면 파일의 대문자 상수(세 글자 이상) 이름을 모두 돌려줍니다. 쉼표 뒤에 이어 선언한 이름(LAST_PROC)은 보지 못합니다.
 * @param {string} file workbench 기준 경로
 * @returns {string[]}
 */
function declared(file) {
  return [...source(file).matchAll(/^[ \t]*const ([A-Z][A-Z_]{2,}) = /gm)].map((m) => m[1]);
}

/**
 * 두 방향의 검사를 모두 돌려 문제 목록을 돌려줍니다. 비면 통과입니다.
 * @returns {string[]}
 */
function check() {
  const problems = [];
  const moved = {};
  for (const df of DATA_FILES) {
    const data = JSON.parse(readFileSync(path.join(here, df), "utf8"));
    for (const [file, entries] of Object.entries(data)) {
      for (const [key, got] of Object.entries(entries)) {
        // 선언이 아닌 고정 글은 화면 파일에 글자 그대로 있어야 합니다.
        if (key === "text") {
          for (const t of got) if (!source(file).includes(t)) problems.push(`${df} ${file} text: 화면에 없는 글 ${JSON.stringify(t)}`);
          continue;
        }
        if (moved[file]?.[key]) problems.push(`${df} ${file} ${key}: 자료 파일 ${moved[file][key]}에도 있음`);
        (moved[file] ??= {})[key] = df;
        try {
          const out = [];
          diff(evaluate(file, key), got, `${df} ${file} ${key}`, out);
          problems.push(...out);
        } catch (e) {
          problems.push(`${df} ${file} ${key}: ${e.message}`);
        }
      }
    }
  }

  // 화면의 상수(와 속성마다 옮기는 상수의 속성)가 모두 옮겨졌거나 까닭과 함께 빠졌는지 봅니다.
  for (const file of Object.keys(EXCLUDED)) {
    const keys = [];
    for (const name of declared(file)) {
      if (!OBJECTS[file]?.includes(name)) {
        keys.push(name);
        continue;
      }
      const v = vm.runInContext(`(${initializer(file, name)}\n)`, vm.createContext({}));
      keys.push(...Object.keys(v).map((p) => `${name}.${p}`));
    }
    for (const key of keys) {
      const isMoved = Boolean(moved[file]?.[key]), isExcluded = Object.hasOwn(EXCLUDED[file], key);
      if (isMoved === isExcluded) problems.push(`${file} ${key}: ${isMoved ? "옮겼는데 EXCLUDED에도 있음" : "옮기지도 빼지도 않음"}`);
    }
    for (const key of Object.keys(EXCLUDED[file])) if (!keys.includes(key)) problems.push(`${file} ${key}: EXCLUDED에 있는데 화면에 없음`);
  }
  return problems;
}

const args = process.argv.slice(2);
if (args[0] === "--print") {
  const [, file, ...keys] = args;
  process.stdout.write(JSON.stringify(Object.fromEntries(keys.map((k) => [k, evaluate(file, k)])), null, 1) + "\n");
} else if (args.length !== 0) {
  console.error("쓰는 법: node topics/pgsql/check-screens.mjs [--print <파일> <이름>...]");
  process.exit(2);
} else {
  const problems = check();
  for (const p of problems) console.log(p);
  console.log(problems.length === 0 ? "검사 6 통과: 자료 파일과 화면의 상수·글이 같음" : `검사 6 실패: ${problems.length}건`);
  process.exit(problems.length === 0 ? 0 : 1);
}
