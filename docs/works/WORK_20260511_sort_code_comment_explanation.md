# WORK 2026-05-11 Sort Code Comment Explanation

## Request

- 요청: 정렬 구현 코드 각각에 대해 `study-explanation`을 사용해 주석으로 상세 설명을 작성한다.
- 핵심 해석: 설명 본체는 README가 아니라 실제 정렬 구현 파일의 Javadoc, block comment, tight inline comment에 둔다.
- 범위: `algorithms/examples/programmers/jvm/src/main/java/sort` 아래 7개 정렬 구현 클래스.

## Project Overlay

- 적용 문서: `/Users/rody/VscodeProjects/study/AGENTS.md`, `PROJECT_INTENT.md`, `USECASE.md`
- 적용 이유: 이 저장소의 예제 코드는 나중에 다시 읽고 손으로 재생할 수 있는 학습 자산이어야 한다.
- 적용한 skill: `/Users/rody/VscodeProjects/myai/runtimes/codex/skills/study-explanation/SKILL.md`
- 선택한 route: code-first explanation
- 선택한 surface: source-near comments. README는 이번 요청의 주 설명 표면이 아니다.

## Frozen Checklist

- [x] Bubble sort, selection sort, insertion sort, merge sort, quick sort, heap sort, counting sort에 코드 근처 설명을 둔다.
- [x] 각 정렬의 첫 상태, 핵심 불변식, 데이터 이동 예시를 주석 안에서 확인할 수 있게 한다.
- [x] 핵심 루프나 재귀가 왜 안전한지, 무엇을 확정하거나 유지하는지 설명한다.
- [x] 기존 정렬 동작을 바꾸지 않는다.
- [x] Java 컴파일과 sort runner로 회귀를 확인한다.
- [x] diff whitespace 검사를 통과한다.

## Execution Notes

- `BubbleSort`: pass가 끝날 때마다 오른쪽 suffix가 확정되는 흐름과 early return 조건을 보강했다.
- `SelectionSort`: 왼쪽 prefix 확정 불변식, 최솟값 선택 과정, swap 기반 구현의 안정성 한계를 보강했다.
- `InsertionSort`: 값을 임시 보관하고 큰 값을 밀어 빈자리를 만드는 장면을 trace로 설명했다.
- `MergeSort`: half-open range, 재귀 바닥 조건, 두 sorted half를 buffer로 합치는 과정을 보강했다.
- `QuickSort`: pivot을 실제 고정 원소가 아니라 기준값으로 쓰는 분할 방식을 설명했다.
- `HeapSort`: 배열 인덱스로 표현한 max heap, root 교환, `siftDown` 복구 흐름을 보강했다.
- `CountingSort`: 값 범위를 count 배열 offset으로 바꾸는 방식과 값 범위가 비용에 들어가는 이유를 보강했다.

## Verification

- `git diff --check -- algorithms/examples/programmers/jvm/src/main/java/sort`: PASS
- `JAVA_HOME=/Users/rody/Library/Java/JavaVirtualMachines/jbr-21.0.10/Contents/Home ./gradlew -q compileJava`: PASS
- `java -cp build/classes/java/main sort.SortLearningRunner`: PASS, `All sort learning tests passed`

## Final Audit

- 사용자 요구인 "주석으로 상세 설명"은 각 구현 파일 안에 반영했다.
- `study-explanation`의 code-first route 기준인 first state, guided trace, data movement, replay path를 각 정렬 파일에 반영했다.
- 코드 실행 로직은 변경하지 않았고, 컴파일과 runner로 동작 회귀가 없음을 확인했다.
- 남은 작업: 없음. 기존 unrelated worktree 변경은 건드리지 않았다.
