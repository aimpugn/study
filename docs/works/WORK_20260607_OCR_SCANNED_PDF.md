# WORK_20260607_OCR_SCANNED_PDF

## 0. Meta

- 작업 제목: Scanned Book PDF OCR Open-Source Guide
- WORK 파일 경로: `docs/works/WORK_20260607_OCR_SCANNED_PDF.md`
- 저장소: `/Users/rody/VscodeProjects/study`
- 작업 유형: `explain`
- 작업 깊이: `standard`
- 관련 요청: 책을 스캔한 PDF OCR 방법 추천 및 오픈소스(PaddleOCR 등) 활용법 안내
- 원문 사용자 요청: "책을 스캔한 pdf가 있는데, OCR 전 원본 상태. https://github.com/PaddlePaddle/PaddleOCR 등 오픈소스 활용해서 스스로 OCR 해보고 싶은데, 어떻게 하는 게 최선일지?"
- 대상 경로 / 자산: `/Users/rody/VscodeProjects/study/ai/ocr_scanned_pdf.md`
- 실행자: Antigravity
- 시작 일시: 2026-06-07T13:00:00+09:00
- 종료 일시: 2026-06-07T13:30:00+09:00
- 현재 상태: `COMPLETE`
- 완료 게이트: `ALLOW_COMPLETE`
- finish: `report`

## 1. Request Normalization

- goal: 스캔본 PDF를 텍스트 및 마크다운(Markdown)으로 복원하기 위한 오픈소스 OCR(PaddleOCR, Docling, Marker 등)의 최선책과 구체적인 활용 워크플로우를 정리한다.
- refs: PaddleOCR GitHub, Docling GitHub, Marker GitHub
- scope: 이미지 전처리(Pre-processing)부터 레이아웃 분석(Layout Analysis), OCR 엔진 선택, 그리고 LLM을 활용한 교정(Post-processing)까지 전체 파이프라인의 실무 감각과 예제 코드를 문서로 작성한다.
- mode: `explain`
- run_mode: `normal`
- finish: `report`
- must_keep: 기존 `ai/` 디렉터리 내 아키텍처 문서들과의 일관성 및 저장소의 deep study monograph 수준을 준수한다.
- extra_checks: N/A

### 1.1 Explicit Deliverables

- 사용자가 명시한 필수 요구: PaddleOCR 등 오픈소스 활용 방법, 스스로 OCR을 구축하고 실행하는 최선의 워크플로우 제시.
- 사용자가 명시한 금지 사항: 없음.
- path / naming / format / finish 관련 요구: 저장소의 mono-repo 성격에 따라 `/Users/rody/VscodeProjects/study/ai/ocr_scanned_pdf.md`에 단일 학습 가이드 파일로 저장.
- 내가 추가한 누락 방지 항목: 
  1. 한국어 서적 OCR 시 언어 지원과 성능 비교 (PaddleOCR vs Docling/Tesseract/EasyOCR)
  2. 스캔본 PDF 특유의 문제(왜곡, 2페이지 분할, 노이즈) 해결을 위한 이미지 전처리 기법 안내
  3. 실무에서 바로 실행할 수 있는 파이썬(Python) 예제 스크립트 작성

### 1.2 Non-Goals

- 이번 작업의 비범위: OCR 모델의 추가 fine-tuning이나 도커(Docker) 기반 대형 클러스터 구축은 제외한다.
- 지금 하지 않는 이유: 사용자의 첫 질문은 OCR을 스스로 하기 위한 최적의 도구 선정과 워크플로우 이해에 초점이 맞춰져 있으므로, 복잡한 커스텀 모델 학습보다는 검증된 오픈소스 도구를 빠르게 결합하는 파이프라인에 집중한다.

## 2. Root-First Framing

- 근본 문제: 스캔본 PDF는 텍스트 레이어가 없는 이미지 묶음이며, 2단 레이아웃, 머리말/꼬리말 노이즈, 스캔 왜곡, 한국어 인식 정확도 저하 등의 문제가 얽혀 있어 단순 OCR 엔진 실행만으로는 읽기 좋은 구조화 문서(마크다운 등)를 얻기 어렵다.
- 왜 이 문제가 지금 중요한가: 스캔 서적을 RAG(검색 증강 생성)에 입력하거나 LLM 학습에 쓰기 위해서는 레이아웃이 깨지지 않은 정확한 원문 텍스트 복원이 필수적이다.
- 작업 목표: 스캔 PDF OCR 파이프라인의 각 단계별 역할과 한계를 규명하고, 오픈소스를 사용해 한국어 스캔본을 가장 깔끔하게 추출하는 파이프라인 가이드를 작성한다.
- 기대 이점: 사용자가 자신의 스캔 서적 상태에 맞춰 최적의 오픈소스 라이브러리를 선정하고, 시행착오 없이 전처리-추출-후처리 파이프라인을 파이썬으로 구현할 수 있게 된다.
- 이점이 닫혔다고 판단할 확인 기준: 생성되는 `ai/ocr_scanned_pdf.md` 문서가 단순 나열을 넘어 구체적인 의사결정 트리(Decision Tree)와 실행 가능한 파이썬 코드를 제공하는가.
- 하드 제약 / 호환성 경계: 한국어(Ko) 지원 여부, CPU/GPU 하드웨어 가속 리소스 제약.
- 성공 정의: PaddleOCR, Docling, Marker, Tesseract의 각 특성을 비교하고, 최고의 성능을 내는 한국어 도서 OCR 워크플로우가 단계별(전처리, 레이아웃 분석, 텍스트 추출, 교정)로 명확히 서술된 가이드 문서를 완성함.
- PARTIAL 조건: 일반론적인 OCR 툴 소개에 그치고 구체적인 파이썬 코드나 한국어 처리의 실무 팁이 누락된 경우.
- BLOCKED 조건: N/A

## 3. Reader & Internalization Contract

- 주 독자: 한국어/영어 스캔 서적을 소유하고 있으며, 이를 디지털 텍스트 자산(마크다운, RAG용 청크 등)으로 변환하려는 개발자/연구자.
- 독자가 이미 알고 있다고 가정하는 것: 기초적인 파이썬 사용법, PDF가 이미지 형태로 되어 있다는 개념.
- 이 작업 결과를 통해 나중에 스스로 설명할 수 있어야 하는 것: 스캔 PDF OCR에서 전처리가 왜 결과 품질의 80%를 결정하는지, 레이아웃 분석(Layout Analysis)이 일반 OCR 엔진과 무엇이 다른지, 현대 문서 파싱(Docling, Marker)과 전통적 OCR(Tesseract)의 아키텍처적 차이.
- teach-back 목표: "스캔한 책을 OCR 할 때, 왜 Tesseract를 바로 돌리면 안 되고 PaddleOCR이나 Docling을 레이아웃 분석과 함께 엮어야 하는가?"에 대해 3단계(레이아웃 분할, 단일 이미지 OCR, 순서 복원)로 설명할 수 있어야 한다.
- 사용자의 현재 빈칸 / learner gap 진단: PaddleOCR이 단어 인식은 잘 하지만 책의 레이아웃(헤더/푸터 제거, 2단 편집 순서)을 자동으로 정렬해주지 않는다는 점에 대한 이해 부족. 이를 해결할 현대 도구(Docling/Marker)의 결합 필요성.
- 예상 독자 수준과 어린아이처럼 대하지 않는 기준: 친근하지만 전문적인 한국어를 사용하며, 단순 '추천'보다는 기술적 한계와 작동 원리를 함께 서술한다.
- 이 설명이 사용자를 더 높은 개발자 판단으로 끌어올려야 하는 지점: 단순 패키지 실행을 넘어, PDF 내 이미지 해상도(DPI) 제어, Binarization 임계치 조정, OCR 후 LLM 교정(Correction) 비용과의 타협 등 엔지니어링 의사결정을 스스로 할 수 있게 돕는다.
- 사용자가 내재화해야 할 사고 패턴: "Garbage In, Garbage Out". 이미지 노이즈가 제거되지 않은 상태에서 OCR 엔진 매개변수를 조정하는 것보다, 전처리 단계에서 이진화(Thresholding)와 왜곡 보정(Deskew)을 하는 것이 훨씬 효과적이다.
- 특히 막아야 하는 오해: "PaddleOCR만 쓰면 마크다운으로 표와 단락이 정렬된 예쁜 책이 뚝딱 나온다"는 오해 방지. PaddleOCR은 텍스트 감지(Detection)와 인식(Recognition) 모듈일 뿐, 문서 구조 복원은 상위 프레임워크나 추가 파이프라인이 필요하다.
- 오해를 깨기 위한 반례 / 대조쌍: Tesseract vs PaddleOCR (작동 방식 및 정확도 비교), Raw Text extraction vs Structured Parsing (Layout-aware).
- 기억 anchor 후보: "레이아웃(Layout)이 없는 OCR은 글자들의 무덤이다."
- active recall 질문 / 작은 실험 / replay 과제: 제공된 파이썬 스크립트를 사용하여 책의 1페이지를 이미지로 변환한 뒤 OpenCV로 Binarization을 적용하고 OCR 신뢰도 점수(Confidence score)의 변화를 비교해보기.
- 반드시 거쳐야 하는 추상화 계층: 이미지(픽셀) -> 바운딩 박스(좌표) -> 텍스트 라인 -> 구조화된 마크다운(블록).
- 핵심 대조쌍 / 혼동쌍: LayoutParser vs PP-Structure (레이아웃 분석 엔진 비교).
- 질문형으로 먼저 답해야 하는 핵심 질문: "오픈소스로 스캔 책을 OCR할 때, 어떤 순서로 파이프라인을 설계해야 최선의 품질을 얻는가?"
- 목차 필요 여부와 이유: 긴 장문 분석 문서이므로 가독성과 학습 편의를 위해 목차가 필수적임.
- `개요` 또는 서두에서 먼저 고정할 문장: "스캔 PDF OCR의 성공은 단순히 글자를 읽는 엔진의 성능이 아니라, 이미지를 다듬는 전처리와 문서의 흐름을 읽는 레이아웃 분석의 유기적 결합에 달려 있습니다."
- 이번 문서의 기본 전개 흐름: 
  1. 핵심 요약 및 의사결정 트리 (도구 선택)
  2. OCR 파이프라인의 4단계 아키텍처 (전처리 - 레이아웃 분석 - 텍스트 인식 - 후처리)
  3. 오픈소스 OCR 도구 비교분석 (PaddleOCR, Docling, Marker, Tesseract)
  4. 한국어 도서 최적화 워크플로우 및 파이썬 코드 예제
  5. 실무적인 한계와 실패 모드 (Troubleshooting)
- 이번 작업의 품질 기준 exemplar: `computer_architecture/threads/threads.md` (기초 배경 설명부터 코드, 비교, 실무 맥락까지 깊이 있게 연결하는 Monograph 스타일)

## 4. Depth Decision

- 선택한 깊이: `standard`
- 왜 이 깊이가 맞는가: 개념 문서 신설과 실무 코드 작성을 동시에 진행하며, 여러 툴의 정량적/정성적 비교를 다룬다.
- 축약 가능한 섹션과 그 근거: N/A

## 5. Project Overlay

- 적용한 로컬 AGENTS 경로: `/Users/rody/VscodeProjects/study/AGENTS.md`
- 활성화한 프로젝트 규칙: study 로컬 실행 규범 전체.

## 17. Frozen Checklist

- [x] C-01: OCR 도구 성능 비교 및 선택 가이드 (PaddleOCR, Tesseract, Docling, Marker 등) 작성.
- [x] C-02: 전체 OCR 파이프라인 아키텍처 설계도 및 단계별 설명 포함.
- [x] C-03: 한국어(Ko) 도서 처리에 맞춘 실무적 팁 제공.
- [x] C-04: 실행 가능한 파이썬(Python) 예제 코드(PaddleOCR 및 이미지 전처리) 포함.
- [x] C-05: 저장소 내 `/Users/rody/VscodeProjects/study/ai/ocr_scanned_pdf.md` 파일로 작성 완료 및 `docs/works/` 워크 로그 반영 완료.
