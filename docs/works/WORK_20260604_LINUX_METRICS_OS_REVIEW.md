# WORK 20260604 Linux Metrics OS Review

## 0. Meta

- Status: COMPLETE after verification; commit to be created after this ledger snapshot and reported in final response
- Work root: `/Users/rody/VscodeProjects/study`
- Requested whole objective: review `linux/commands/metrics` as Linux/OS/kernel learning material, run `$review-kernel + $multi-agent + $dialectic-kernel`, and immediately improve bounded findings when safe.
- Requested closure scope: metrics docs/scripts review, bounded repairs, verification, final audit, scoped commit.
- Active skills: `review-kernel`, `multi-agent`, `dialectic-kernel`, coupled `rigorous-task`.
- Multi-agent topology: compatibility fallback to single-agent separated lanes because this runtime did not expose independent background sub-agent delegation. Roster: `Orchestrator`, `Builder`, `Critic`, `Protocol Sentinel`.
- Primary target paths: `linux/commands/metrics/*`, `interviews/linux-kernel-hardware-practical-internals.md` as related owner/drift reference.
- Dirty-scope note: repository had many unrelated modified/untracked files before this work. This run only touches `linux/commands/metrics/*` and this WORK file.

## 1. Instruction Stack

- Global AGENTS: resolved via `/Users/rody/.codex/AGENTS.md` symlink and applied. Relevant rules: protected-intent normalization, checklist freeze, evidence-first review, no bare completion, review-to-repair handoff, commit after repo-changing verification unless excluded.
- Global WORK template: resolved via `/Users/rody/.codex/AGENTS_WORK_TEMPLATE.md`, superseded by repo-local template for this repo.
- Repo AGENTS: `AGENTS.md` read and applied. Relevant rules: study assets must restore understanding, connect explanation to replayable evidence, use exemplar floor, keep WORK under `docs/works/`, and commit after final review/test for repo-changing work.
- Repo WORK template: `AGENTS_WORK_TEMPLATE.md` read and used in compressed form.
- Project facts: `PROJECT_INTENT.md` and `USECASE.md` read and applied. No root `TERMINOLOGY.md` exists.
- Nested overlay: no `linux/**/AGENTS.md`, `PROJECT_INTENT.md`, `USECASE.md`, or `TERMINOLOGY.md` found.

## 2. Request Normalization

- Surface request: "현재 프로젝트를 study로 보고, `linux/commands/metrics`의 VMware/OS monitoring scripts가 커널·OS 공부를 실제 문제 의미까지 이해시키는지 review/검수하고, 필요하면 바로 개선/수정."
- Protected intent: avoid command-cheatsheet drift; make metrics explain `request path -> kernel object -> queue/cache/buffer/wait -> observable metric -> failure mode -> escalation/verification`.
- Risky literal interpretation: only report "good/bad" or only run script tests while missing whether the material teaches why and when the metric matters.
- Quality floor: repo exemplar/deep-study contract plus prior capstone rule from `interviews/linux-kernel-hardware-practical-internals.md`: scripts are evidence, not the teaching center.
- Finish: review + bounded repair + verification + commit, because the user did not request report-only/no-commit.

## 3. Review Routing Surface

- review target: `linux/commands/metrics` docs/scripts, especially `storage_io_vmware.md`, `vm_os_watch.sh`, `vm_os_lab.sh`, `vm_io_health.sh`, `vm_stat.md`, `io.md`, tests.
- requested outcome / action intent: review-and-improve.
- protected purpose: durable OS/kernel learning and VMware troubleshooting reasoning, not tool syntax memorization.
- activated common gates: purpose fit, requested outcome fidelity, claim fidelity, evidence strength, counterexample resistance, omission/distortion, verification, downstream impact, review-to-repair handoff.
- activated domain packs: Linux `/proc` and block I/O metrics, VMware ESXi storage latency, study-repo explanation quality, shell script regression tests.
- omitted expertise: live ESXi/vCenter and real Oracle Linux VM runtime verification.
- why omitted expertise is safe: current repairs do not claim live host performance; live runtime remains bounded by script tests and official docs.
- evidence boundary: repo evidence + official docs + script regression tests; no live production behavior claim.
- conditional gates activated: dialectic handoff, cost/rigor receipt, review-to-repair handoff, downstream impact.
- verification surface: `bash linux/commands/metrics/tests/run_tests.sh`, `npx --no-install markdownlint-cli2 ...`, `git diff --check`.

## 4. Frozen Checklist

- C-01 USER: Review whether `linux/commands/metrics` explains OS/kernel meaning, not only scripts/commands. PASS requires concrete evidence from docs/scripts.
- C-02 USER: Pay special attention to VMware framing. PASS requires host-vs-guest boundary and GAVG/KAVG/DAVG escalation to be checked.
- C-03 USER: Use review-kernel + multi-agent + dialectic-kernel. PASS requires separated critic/sentinel lanes and claim attack/repair records.
- C-04 USER: If material bounded improvements are needed, apply them immediately. PASS requires accepted findings converted to patch and verified.
- C-05 AI: Preserve unrelated dirty WIP. PASS requires scoped diff/stage/commit.
- C-06 AI: Verify script and Markdown surfaces. PASS requires metrics tests, lint, and diff check or explicit blocker.

## 5. Evidence Ledger

- E-01 repo: `storage_io_vmware.md` already opens with situation-based entrypoints, request path, `/proc/diskstats`, iowait caveats, VMware GAVG/KAVG/DAVG, `vm_os_watch.sh`, field dictionary, case decomposition, and lab experiments.
- E-02 repo: `vm_os_watch.sh` calculates deltas from `/proc/stat`, `/proc/diskstats`, `/proc/net/dev`, `/proc/vmstat`, `/proc/meminfo`, and labels values as kernel/user-facing meanings.
- E-03 repo: `vm_os_lab.sh` turns monitor values into prediction/observation experiments: iowait illusion, dirty/writeback, direct write, page cache, tmpfs/available memory.
- E-04 repo: `tests/run_tests.sh` has deterministic snapshots for CPU, load, D state, memory, swap, LV rows, disk await/aqu/util, layout, log behavior.
- E-05 official: Linux kernel iostats docs confirm `/proc/diskstats` and `/sys/block/*/stat` expose cumulative counters, including time spent doing I/O and weighted I/O time.
- E-06 official: Linux proc docs warn iowait from `/proc/stat` is not reliable.
- E-07 official: Broadcom/VMware docs define esxtop GAVG/KAVG/DAVG and queue/latency boundaries.
- E-08 repo/prior: `interviews/linux-kernel-hardware-practical-internals.md` sets the general OS/kernel owner; metrics docs are the VMware environment detail.

## 6. Claim / Reasoning Ledger

| Claim | Support tier | Admission | Critic attack | Response | Status |
| --- | --- | --- | --- | --- | --- |
| The metrics material is not merely a command cheat sheet; its core document and lab already teach kernel counters, queues, cache/writeback, wait states, and VMware boundaries. | T1 repo evidence + official docs | APPLY | The document may still be too VMware-centered for general OS learning. | ACCEPT_REPAIR: add explicit bridge to the general kernel/hardware monograph and teach-back targets near the top. | PASS |
| Script logic is regression-protected enough for this review lane. | T1 test suite after installing gawk | APPLY | Initial test failure could indicate broken script logic. | REBUT: failure was missing `gawk`; after installing it, deterministic tests pass 48/0, Linux live portion skipped on macOS `/proc` absence. | PASS |
| `vm_stat.md` is out of scope because the user emphasized VMware/Linux. | T2 inference | HOLD/REWORK | It is inside `linux/commands/metrics` and lint fails; it also has reversed Swapins/Swapouts descriptions. | ACCEPT_REPAIR: fix lint style and swap direction using local macOS manpage evidence. | PASS |
| This turn can safely repair without asking for plan approval. | T1 user request + repo policy | APPLY | Installing `gawk` changes local environment. | ACCEPT_REPAIR: install was a small reversible verification dependency but Homebrew auto-cleanup side effect must be disclosed. | PASS with disclosure |

## 7. Dialectic Rounds

- R1 definition: Orchestrator claim = review must judge learning transfer, not just script correctness. Critic attack = "metrics docs may already be fine; over-editing can bloat." Response = ACCEPT_REPAIR by using bounded bridge text, not wholesale rewrite. Completion marker: definition accepted.
- R2 evidence breadth: Orchestrator claim = official and repo evidence are enough for structural review. Critic attack = "VMware facts need primary source." Response = ACCEPT_REPAIR by checking Linux kernel docs and Broadcom VMware docs. Completion marker: evidence accepted.
- R3 checklist stress: Critic attack = "`vm_stat.md` is easy to miss because it is macOS, but it is in the requested directory." Response = ACCEPT_REPAIR. Completion marker: checklist includes all metrics docs.
- R4 execution result: Builder patches `storage_io_vmware.md`, `vm_stat.md`, and this WORK. Critic attack = "bridge text might be decorative and tests may only prove syntax." Response = REBUT/ACCEPT_REPAIR: bridge text freezes teach-back targets at the reading entrypoint, and tests verify script calculations while lint/diff verify doc mechanics. Completion marker: verification accepted.
- R5 closure: Protocol Sentinel checked final verification, scoped diff, and commit boundary. Completion marker: PASS, with live Linux `/proc` execution disclosed as skipped on macOS.

## 8. Repair Log

- Added a short bridge in `storage_io_vmware.md` from VMware-specific metrics to the general `interviews/linux-kernel-hardware-practical-internals.md` monograph and froze four teach-back targets.
- Converted `vm_stat.md` manpage excerpt from indented code to fenced `text` block for markdownlint.
- Corrected `vm_stat.md` Swapins/Swapouts descriptions using `man vm_stat` evidence.
- Installed `gawk` with Homebrew to make `vm_os_watch.sh` regression tests runnable in this local environment. Homebrew also auto-updated and auto-removed `python@3.13` during cleanup; this is an environment side effect, not a repo file change.

## 9. Verification Log

- `bash linux/commands/metrics/tests/run_tests.sh`: PASS 48 / FAIL 0. Linux live `/proc` execution checks were skipped because this verification ran on macOS.
- `npx --no-install markdownlint-cli2 linux/commands/metrics/storage_io_vmware.md linux/commands/metrics/io.md linux/commands/metrics/vm_stat.md docs/works/WORK_20260604_LINUX_METRICS_OS_REVIEW.md`: PASS, 0 errors.
- `git diff --check -- linux/commands/metrics/storage_io_vmware.md linux/commands/metrics/vm_stat.md docs/works/WORK_20260604_LINUX_METRICS_OS_REVIEW.md`: PASS.
- `man vm_stat | col -b`: used to verify `Swapins` means swapped back in from disk and `Swapouts` means swapped out to disk.

## 10. Closure Control

- Requested whole objective id: `WORK_20260604_LINUX_METRICS_OS_REVIEW`
- Requested closure scope: metrics review + bounded repair + verification + scoped commit.
- Achieved closure scope: metrics review + bounded repair + verification; scoped commit pending immediately after this ledger update.
- Whole-request completion verdict: WHOLE_COMPLETE after scoped commit.
- Remaining executable count: 0 after scoped commit.
- Next immediate target: scoped commit of this work's files.
- Unrelated open work: many pre-existing dirty/untracked files outside this write scope remain untouched.
