# DB deep-dive source map

이 문서는 사람이 읽는 출처·보존 지도입니다.
기계가 검사하는 canonical 계약은 같은 디렉터리의 `source-map.tsv`입니다.

`source-map.tsv`는 각 DU마다 다음 필드를 반드시 채웁니다.

- `local_seed`: 기존 `database/` 또는 관련 study 문서 중 어떤 자료를 살릴지.
- `official_sources`: 본문 판단의 하중을 지탱할 공식 문서, 표준, vendor reference.
- `lab_or_observability`: 독자가 직접 확인할 실험, SQL trace, metric, log, 관측 경로.
- `preservation_disposition`: 기존 자료를 보존, 병합, 확장, 대체 포인터, 보안상 승격 금지 등으로 어떻게 처리할지.
- `source_status`: whole-complete 시점에는 `verified`여야 한다. 작성 전 계획 상태는 `planned`로 둔다.
- `notes`: `du-registry.tsv`의 source requirement를 역추적할 수 있게 남긴다.

현재 source coverage 대상 DU는 아래 56개입니다.

```text
DU01 DU02 DU03 DU04 DU05 DU06 DU07
DU08 DU09 DU10 DU11 DU12 DU13 DU14
DU15 DU16 DU17 DU18 DU19 DU20 DU21
DU22 DU23 DU24 DU25 DU26 DU27 DU28
DU29 DU30 DU31 DU32 DU33 DU34 DU35
DU36 DU37 DU38 DU39 DU40 DU41 DU42
DU43 DU44 DU45 DU46 DU47 DU48 DU49
DU50 DU51 DU52 DU53 DU54 DU55 DU56
```

## 보존 원칙

기존 문서의 문장을 그대로 복사하는 것이 보존은 아닙니다.
기존 자료가 좋은 질문, 좋은 예시, 실제 장애 감각, 실험 흔적을 담고 있으면 새 deep-dive 문서에서 의미를 살려 확장합니다.
반대로 오래된 표현, 부정확한 단정, 민감 정보, 실행 환경이 불명확한 로그는 본문에 그대로 승격하지 않고 `source-map.tsv`에서 처분을 기록한 뒤 안전한 synthetic 예제나 관측 경로로 바꿉니다.

`database/elasticsearch/tools/esdump/auth.ini`처럼 secret 후보가 있는 파일은 `sensitive-source-do-not-promote`로 유지합니다.
이런 파일은 본문에서 직접 인용하지 않고, secret hygiene와 redaction 원칙을 설명하는 반례나 점검 항목으로만 사용합니다.

## 상태 모델

작성 전 `planned`는 출처 후보와 검증 경로를 먼저 고정했다는 뜻입니다.
본문 작성 중 실제로 해당 공식 문서를 확인하고, local seed를 읽고, lab 또는 관측 경로가 본문과 맞는지 확인한 DU만 `verified`로 바꿀 수 있습니다.
따라서 `planned` 상태가 남아 있으면 validator는 whole-complete를 허용하지 않습니다.
