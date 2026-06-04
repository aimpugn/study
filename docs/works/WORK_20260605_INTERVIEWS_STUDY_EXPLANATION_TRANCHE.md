# WORK_20260605_INTERVIEWS_STUDY_EXPLANATION_TRANCHE

## 0. Meta

- 작업 제목: interviews 정리 문서 study-explanation 재작성 장기 tranche
- WORK 파일 경로: `docs/works/WORK_20260605_INTERVIEWS_STUDY_EXPLANATION_TRANCHE.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `refactor_docs | explain | audit | execute`
- 작업 깊이: `full`
- 현재 상태: `PARTIAL_FOR_WHOLE_REQUEST / FIRST_TRANCHE_VERIFIED`
- 완료 게이트: `BLOCK_COMPLETE`
- finish: `test+commit`

## 1. Request Normalization

- goal: `interviews/` 아래 raw source를 제외한 기존 정리 문서를 완전 초보 개발자도 이해할 수 있는 deep study 문서로 단계적으로 승격한다.
- protected intent: `OS Kernel Foundations`에서 좋아진 설명 기준을 전체 코퍼스에 확산한다. 문서가 길어지는 것이 목표가 아니라, 질문 -> 첫 벽돌 -> 메커니즘 -> trace -> 실패/검증 -> 면접 답변으로 이어지는 이해 복원력을 높이는 것이 목표다.
- risky literal interpretation: 4만 줄 이상의 정리 문서를 한 번에 모두 건드리고 완료로 선언하면 누락, 표면적 문장 교체, raw source 훼손, 검증 부재가 생긴다.
- scope: `interviews/` 하위 정식 학습 문서, hub/index 문서, deep-dive 문서.
- non-goals: `interviews/source/` raw question bank, `audit/` 검수 장부, 로컬 규칙/사실 문서의 내용 재작성.
- finish condition: 각 tranche마다 대상 파일, 수정/비수정/HOLD 사유, 검증 결과, 남은 대상 수를 남긴다. 전체 COMPLETE는 모든 정식 문서가 판정된 뒤에만 가능하다.

## 2. Active Instruction / Skill Stack

- 적용 규칙: `study/AGENTS.md`, `interviews/AGENTS.md`, `interviews/PROJECT_INTENT.md`, `interviews/USECASE.md`, `study/ai/authoring/LEARNING_DOC_GUIDE.md`
- 적용 스킬: `study-explanation`, `multi-agent`, `dialectic-kernel`, `review-kernel`, `humanize-korean`
- primary exemplar: `os-kernel-distributed-systems-deep-dive/01_os_kernel_foundations.md`
- reference principle: 문제를 먼저 열고 각 문제의 실제 커널 메커니즘과 역할 설명을 붙인다. 보이지 않는 커널 상태를 trace로 보여 준다.
- trait to avoid: 이름만 나열한 뒤 메타 안내문으로 설명 부족을 보완하는 방식.
- target quality lift: hub와 상세 문서가 `01_os_kernel_foundations.md`의 용어, 문제-해법 연결, beginner bridge를 따라가게 한다.

## 3. Corpus Inventory

- raw source excluded: `source/` 9개 Markdown, 약 29,953 lines.
- audit/provenance excluded from prose rewrite: `*/audit/*.md`.
- project rule/fact docs applied but not rewritten: `AGENTS.md`, `PROJECT_INTENT.md`, `USECASE.md`.
- organized docs candidate inventory after excluding source/audit: 53 Markdown files.
- content/support target after excluding project rule/fact docs: 50 Markdown files.
- first tranche: `os-kernel-distributed-systems-deep-dive/README.md`, `00_index_and_learning_path.md`, `01a` through `01e`.

## 4. Frozen Checklist

- [x] 대상 문서가 완전 초보 개발자에게 필요한 첫 벽돌을 초반에 둔다.
- [x] 문제를 나열하면 대표 메커니즘과 한 줄 역할 설명을 붙인다.
- [x] 숨은 커널 상태, queue/cache/buffer/wait는 가능한 한 trace나 table로 보인다.
- [x] `사건`처럼 OS/runtime 문맥에서 어색한 직역은 일반 한국 개발자 용례에 맞게 `이벤트(event)` 등으로 정리한다.
- [x] 질문형 replay 앞에는 기억을 복원하는 summary/checkpoint가 먼저 온다.
- [x] 제품 매핑은 제품 용어가 무엇인지와 앞 메커니즘의 어떤 비용/상태와 연결되는지 설명한다.
- [x] 비수정 파일은 "이미 충분함" 또는 "이번 tranche 밖" 사유를 남긴다.
- [x] 수정 후 링크, 금지 표현, 사용자 formatOnSave 스타일, git diff를 검증한다.

## 5. Multi-Agent Roster

- Orchestrator: Codex main agent, scope freeze, edits, synthesis, verification.
- Beginner Reviewer: complete beginner comprehension gaps for first OS tranche.
- Senior OS/Backend Reviewer: technical correctness, overclaim, missing caveat review.
- Korean Learning-Prose Reviewer: natural Korean learning flow and terminology review.

## 6. Dialectic Claim Cards

- C1: 첫 tranche를 OS hub/detail 묶음으로 시작하면 최근 개선된 kernel foundation 기준을 가장 빨리 주변 문서에 확산할 수 있다.
    - support tier: T2 strong inference from current user praise and adjacent doc topology.
    - admission lane: APPLY, reversible Markdown edits with same-loop review.
    - falsifier: agents or diff review show broader root README/database tranche is more urgent for coherence.
- C2: raw `source/`와 `audit/`는 이번 prose rewrite 대상이 아니라 evidence/provenance surface로 보존해야 한다.
    - support tier: T1 repo overlay and memory-backed project rule.
    - admission lane: APPLY.
    - falsifier: user explicitly asks to rewrite raw source files.
- C3: whole-request COMPLETE is blocked until every organized document is either improved, explicitly skipped with evidence, or held with a blocker.
    - support tier: T1 user request plus repo completion contract.
    - admission lane: APPLY.
    - falsifier: none inside current request.

## 7. Tranche Log

- 2026-06-05 first tranche started. Reviewer agents spawned for beginner, senior OS/backend, and Korean prose perspectives.
- 2026-06-05 first tranche edited:
    - `README.md`: `write(fd, bytes)`를 첫 벽돌로 두고, 요청이 kernel object, queue/cache/buffer/log, recovery path로 이동하는 기본 trace를 추가했다. 글자 수 기준과 `substantive` 표현은 제거했다.
    - `00_index_and_learning_path.md`: 전체 코퍼스의 문제-메커니즘-상태 표를 추가하고, replay 질문 앞에 실행/주소 공간/파일/네트워크/관측 기억 지도를 추가했다.
    - `01a_process_scheduling.md`: process, thread, run queue, context switch를 초보자용 첫 객체로 정의하고 Kafka/Cassandra/Spark 증상을 OS scheduling 상태로 내리는 trace를 추가했다. Linux scheduler 설명은 CFS 단정에서 CFS/EEVDF version caveat로 수정했다.
    - `01b_memory_and_address_space.md`: virtual address, page table, TLB, page fault 첫 trace와 VMA 표를 추가했다. `mmap()` 설명은 file-backed mapping, `MAP_SHARED`, `MAP_PRIVATE`, truncate/SIGBUS, DAX/MAP_SYNC 경계를 분리했다.
    - `01c_filesystem_page_cache_block_io.md`: path/fd/page cache/block queue 첫 trace와 path/dentry/inode/open file/fd owner-lifetime 표를 추가했다. Kafka ack, Cassandra CL, Spark checkpoint를 OS durability와 구분하는 caveat를 보강했다.
    - `01d_network_stack_and_io_multiplexing.md`: packet 도착과 request handler 실행 사이의 NIC RX queue, socket receive queue, event loop ready list를 초반에 추가했다. NAPI/RSS 설명은 누가 packet 작업을 수행하는지 초보자가 따라갈 수 있게 다시 썼다.
    - `01e_concurrency_isolation_observability.md`: 공유 상태와 대기 주체의 기억 지도, deadlock 오탈자 수리, 컨테이너/VM 용어 정리, Kafka/Cassandra/Spark 증상별 metric -> OS state -> distributed state -> 반증 확인 표를 추가했다.

## 8. Reviewer Findings And Repairs

- Beginner Reviewer:
    - finding: hub와 index가 `write()`에서 kernel state로 내려가는 첫 장면을 충분히 보여 주지 않았다.
    - repair: `README.md`와 `00_index_and_learning_path.md`에 첫 trace와 문제-메커니즘-상태 표를 추가했다.
    - finding: `01b`, `01c`, `01d`에서 VMA, file object lifetime, NAPI/RSS의 실행 주체가 초보자에게 숨어 있었다.
    - repair: 각 문서 초반에 객체 정의, owner/lifetime 표, packet 처리 주체 설명을 추가했다.
- Senior OS/Backend Reviewer:
    - finding: Linux 일반 scheduler를 CFS 중심으로만 말하면 최신 kernel 기준에서 과장될 수 있다.
    - repair: Linux 6.6 이후 EEVDF 전환 caveat와 kernel version/scheduling class/cgroup 설정 경계를 추가했다.
    - finding: Kafka/Cassandra/Spark 제품 매핑에서 OS 성공, product ack, distributed durability가 섞일 위험이 있었다.
    - repair: Kafka `acks=all`/ISR/min ISR, Cassandra CL/read repair/repair, Spark checkpoint/object storage 경계를 문서별로 나눠 적었다.
- Korean Learning-Prose Reviewer:
    - finding: `사건`, `substantive`, `20,000자`, `Container는`, `disk and memory heavy workload`, 비한국어 오탈자가 독자 흐름을 끊었다.
    - repair: 일반 한국 개발자가 자주 쓰는 표현을 우선하고 필요한 원어를 괄호 병기했다. proxy length 문장과 오탈자를 제거했다.

## 9. Verification Log

- `rg -n "사건|substantive|20,000|설명할 수 있는가입니다|caveat|subtle|Container|container는|Virtual machine|disk and memory heavy|대기열|아무도 ahead|आगे" ...first tranche files`
    - result: no matches.
- `git diff --check -- ...first tranche files docs/works/WORK_20260605_INTERVIEWS_STUDY_EXPLANATION_TRANCHE.md`
    - result: PASS, no whitespace errors.
- `git diff --stat -- ...first tranche files`
    - result before WORK update: 7 content files changed, 280 insertions, 51 deletions.
- one-off local Markdown link/anchor check over first tranche files
    - result: PASS, all local file targets and local anchors resolved.
- official source spot-check:
    - Linux scheduler docs confirm CFS is now making room for EEVDF and EEVDF uses lag/virtual deadline.
    - Kafka design docs confirm committed messages, ISR, `acks`, `min.insync.replicas`, and consumer visibility boundaries.
    - Cassandra official docs confirm commitlog/memtable/SSTable storage path and read repair consistency tradeoffs.
    - Spark official docs confirm executor memory overhead and Kubernetes/YARN container memory accounting boundaries.

## 10. Final Status

- first tranche verdict: `TRANCHE_COMPLETE_FOR_THIS_LOOP`.
- whole-request verdict: `PARTIAL`, because the full `interviews/` organized corpus is not yet fully rewritten or explicitly skipped.
- remaining disclosure: after excluding `source/`, `audit/`, and project rule/fact docs, 43 content/support Markdown files still need a tranche judgement. That judgement may be `improve`, `already sufficient`, `support/provenance skip`, or `HOLD`, but it must be recorded before whole-request COMPLETE.
- next immediate target: classify and improve the next highest-impact tranche. Candidate routes are the root interview hub docs or the database deep-dive suite.
