# pprof

## pprof?

**P**erformance **prof**iles.

> **Profiling?**
>
> 메모리 사용량, 함수별 CPU 점유 시간, tracing 등 어플리케이션을 동적으로 분석하는 것

## 분석 과정

1. 어플리케이션에 `profile.proto` 생성 로직 작성
2. 어플리케이션 실행 후 체크하고 싶은 지점에서 `profile.proto`를 생성\
3. `pprof`로 `profile.proto` 분석

## 분석 항목

- `allocs`: 과거에 할당된 모든 메모리의 샘플링
- `block`: 동기화 메커니즘(synchronization primitives)에서 발생한 blocking을 추적
- `cmdline`: 프로그램에서 발생한 command line 호출을 알 수 있다
- `goroutine`: 모든 현재 고루틴들을 stack trace
- `heap`: 메모리에 할당된 살아있는 오브젝트들을 샘플링.
- `mutex`: 충돌된 뮤텍스 홀더들의 stack trace
- `profile`: CPU를 프로파일링
- `threadcreate`: 새로 만들어진 OS thread에 대한 stack trace
- `trace`: 현재 프로그램 실행에 대한 trace

## `pprof`s

### [go fiber pprof](https://docs.gofiber.io/api/middleware/pprof/)

> Pprof middleware for Fiber that serves via its HTTP server runtime profiling data in the format expected by the pprof visualization tool.
> The package is typically only imported for the side effect of registering its HTTP handlers.
> The handled paths all begin with `/debug/pprof/`.

## graphviz

```bash
sudo yum install graphviz
```

## 실행 방법

## pprof commands & options

### Commands

- `callgrind`: Outputs a graph in callgrind format
- `comments`: Output all profile comments
- `disasm`: Output assembly listings annotated with samples
- `dot`: Outputs a graph in DOT format
- `eog`: Visualize graph through eog
- `evince`: Visualize graph through evince
- `gif`: Outputs a graph image in GIF format
- `gv`: Visualize graph through gv
- `kcachegrind`: Visualize report in KCachegrind
- `list`: Output annotated source for functions matching regexp
- `pdf`: Outputs a graph in PDF format
- `peek`: Output callers/callees of functions matching regexp
- `png`: Outputs a graph image in PNG format
- `proto`: Outputs the profile in compressed protobuf format
- `ps`: Outputs a graph in PS format
- `raw`: Outputs a text representation of the raw profile
- `svg`: Outputs a graph in SVG format
- `tags`: Outputs all tags in the profile
- `text`: Outputs top entries in text form
- `top`: Outputs top entries in text form
- `topproto`: Outputs top entries in compressed protobuf format
- `traces`: Outputs all profile samples in text form
- `tree`: Outputs a text rendering of call graph
- `web`: Visualize graph through web browser
- `weblist`: Display annotated source in a web browser
- `o/options`: List options and their current values
- `q/quit/exit/^D`: Exit pprof

### Options

- `call_tree`: Create a context-sensitive call tree
- `compact_labels`: Show minimal headers
- `divide_by`: Ratio to divide all samples before visualization
- `drop_negative`: Ignore negative differences
- `edgefraction`: Hide edges below `<f>*total`
- `focus`: Restricts to samples going through a node matching regexp
- `hide`: Skips nodes matching regexp
- `ignore`: Skips paths going through any nodes matching regexp
- `intel_syntax`: Show assembly in Intel syntax
- `mean`: Average sample value over first value (count)
- `nodecount`: Max number of nodes to show
- `nodefraction`: Hide nodes below `<f>*total`
- `noinlines`: Ignore inlines.
- `normalize`: Scales profile based on the base profile.
- `output`: Output filename for file-based outputs
- `prune_from`: Drops any functions below the matched frame.
- `relative_percentages` Show percentages relative to focused subgraph
- `sample_index`: Sample value to report (0-based index or name)
- `show`: Only show nodes matching regexp
- `show_from`: Drops functions above the highest matched frame.
- `source_path`: Search path for source files
- `tagfocus`: Restricts to samples with tags in range or matched by regexp
- `taghide`: Skip tags matching this regexp
- `tagignore`: Discard samples with tags in range or matched by regexp
- `tagleaf`: Adds pseudo stack frames for labels key/value pairs at the callstack leaf.
- `tagroot`: Adds pseudo stack frames for labels key/value pairs at the callstack root.
- `tagshow`: Only consider tags matching this regexp
- `trim`: Honor nodefraction/edgefraction/nodecount defaults
- `trim_path`: Path to trim from source paths before search
- `unit`: Measurement units to display

### Option groups (only set one per group)

- granularity
    - `functions`: Aggregate at the function level.
    - `filefunctions`: Aggregate at the function level.
    - `files`: Aggregate at the file level.
    - `lines`: Aggregate at the source code line level.
    - `addresses`: Aggregate at the address level.
- sort
    - `cum`: Sort entries based on cumulative weight
    - `flat`: Sort entries based on own weight

:   Clear focus/ignore/hide/tagfocus/tagignore

type "help <cmd|option>" for more information
