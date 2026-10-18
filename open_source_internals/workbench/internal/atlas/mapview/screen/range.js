// 범위 규칙(브라우저 쪽). 정본은 Go의 mapview/ranges.go이고, check-map.mjs가 Go가 쓴 대조표로 같은 결과인지 봅니다.
// 브라우저에서는 window.AtlasRange, node에서는 이 파일의 글을 함수로 감싸 돌려받습니다.
(function (root) {
  "use strict";

  // normalize 함수는 드래그한 가로 구간 [x0, x1]을 칸 경계에 맞춘 [first, last]로 바꿉니다.
  // 조금이라도 덮은 칸은 모두 넣고(바깥쪽으로 맞춤), 고를 수 있는 칸은 [lo, hi]뿐입니다. 덮은 칸이 없으면 null입니다.
  function normalize(cells, x0, x1, lo, hi) {
    if (x1 < x0) { const t = x0; x0 = x1; x1 = t; }
    let first = -1, last = -1;
    for (let i = Math.max(lo, 0); i <= hi && i < cells.length; i++) {
      const c = cells[i];
      if (c.x1 < x0 || c.x0 > x1) continue;
      if (first < 0) first = i;
      last = i;
    }
    return first < 0 ? null : [first, last];
  }

  // rangeLeaves 함수는 [first, last] 칸들의 잎 집합의 합집합을 정렬해 돌려줍니다.
  function rangeLeaves(cellLeaves, first, last) {
    const m = new Set();
    for (let i = first; i <= last && i < cellLeaves.length; i++) for (const id of cellLeaves[i]) m.add(id);
    return [...m].sort((a, b) => a - b);
  }

  // relate 함수는 범위 a가 범위 b에 대해 어떤 관계인지 돌려줍니다(same, same_leaves, contains, inside, overlap, apart).
  function relate(a, b) {
    const inB = new Set(b.leaves);
    let common = 0;
    for (const id of a.leaves) if (inB.has(id)) common++;
    if (common === a.leaves.length && common === b.leaves.length) return a.zoom === b.zoom ? "same" : "same_leaves";
    if (common === b.leaves.length) return "contains";
    if (common === a.leaves.length) return "inside";
    if (common > 0) return "overlap";
    return "apart";
  }

  const api = { normalize, rangeLeaves, relate };
  root.AtlasRange = api;
  return api;
})(typeof window !== "undefined" ? window : globalThis);
