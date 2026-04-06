# 학습 자산 구조 기준

이 문서는 `learn-netty` 프로젝트에서 문서, 예제 코드, 실험, 나중의 시각화 자산을 **어떤 축으로 배치할지** 정하는 기준 문서입니다.

결론부터 먼저 말하면, **지금 기본값은 서브프로젝트 분리가 아니라 "학습 단계별 패키지 + 공통 지원 패키지"** 입니다.

그 이유는 지금 이 프로젝트의 중심이 `배포 단위`보다 `학습 단위`이기 때문입니다. 지금 단계에서 가장 중요한 질문은 "이 코드를 어떤 artifact로 나눌까?"보다, **"이 강의가 이전 강의에서 무엇을 이어받고, 다음 강의로 무엇을 넘기는가?"** 입니다.

## 왜 지금은 서브프로젝트보다 lesson 패키지인가

서브프로젝트는 분명 장점이 있습니다. 빌드 경계를 분리하고, 의존성을 다르게 가져가고, 실행 앱과 공유 라이브러리를 깔끔하게 나눌 수 있습니다.

하지만 지금 바로 그렇게 가면, 초반 학습 자산은 오히려 아래처럼 흐려질 가능성이 큽니다.

1. 학습 단계보다 빌드 구조가 먼저 눈에 들어옵니다.
2. "Lesson 01에서 Lesson 02로 어떻게 발전했는가"보다 "이 모듈은 어디를 의존하는가"가 더 먼저 보입니다.
3. 강의용 비교 예제가 모듈 경계로 찢어져, 한눈에 비교하기 어려워집니다.

이 프로젝트의 현재 목적은 제품 아키텍처를 설계하는 것이 아니라, **Netty를 단계적으로 이해하는 데 가장 유리한 구조를 고르는 것**입니다. 그래서 지금 기본값은 `lesson 중심 구조`가 더 맞습니다.

## 현재 기본 구조

현재와 가까운 기본 구조는 아래처럼 봅니다.

```text
learn-netty/
  docs/
    lessons/
      lesson-01-why-netty.md
      lesson-02-...
    works/
      20260406_....md
  src/main/java/io/aimpugn/learn/netty/
    support/
    lesson1/
    lesson2/
    lesson3/
  src/test/java/io/aimpugn/learn/netty/
    lesson1/
    lesson2/
    lesson3/
```

핵심은 최상위 축이 `lesson1`, `lesson2`, `lesson3`처럼 **학습 단계**라는 점입니다.

## 패키지 배치 원칙

### 1. lesson이 최상위 축이다

새로운 주제는 먼저 "이게 몇 번째 단계 학습인가?"를 묻고, 그 답을 패키지에 반영합니다.

예를 들어 첫 강의는 이렇게 놓습니다.

```text
src/main/java/io/aimpugn/learn/netty/
  lesson1/
    blocking/
    nio/
    netty/
    common/
```

이 구조의 장점은 단순합니다. 비교하는 대상이 서로 멀리 흩어지지 않습니다. 첫 강의에서 보고 싶은 것은 `blocking`, `nio`, `netty`의 차이인데, 그 셋이 한 lesson 아래 모여 있어야 눈으로 비교하기 쉽습니다.

### 2. lesson 내부에서는 "비교 대상"이나 "실험 종류"로 나눈다

lesson 아래의 하위 패키지는 기술 이름이나 실험 종류처럼 **이번 단계에서 실제로 비교하는 축**으로 나눕니다.

예를 들어 다음처럼 나누는 것이 자연스럽습니다.

```text
lesson2/
  eventloop/
  experiment/
  netty/
```

또는

```text
lesson3/
  inbound/
  outbound/
  pipeline/
```

중요한 것은 "항상 같은 모양으로 강제한다"가 아니라, **이번 lesson에서 독자가 무엇을 비교하면서 배워야 하는지**가 패키지 이름에 드러나야 한다는 점입니다.

### 3. 여러 lesson이 함께 쓰는 것만 support로 올린다

둘 이상의 lesson에서 반복해서 쓰고, lesson에 종속되지 않은 지원 코드는 `support`로 올립니다.

대표적으로 이런 것들이 여기에 들어갑니다.

1. 공통 관측 유틸
2. 테스트 클라이언트
3. 샘플 실행 helper
4. 시각화용 공통 event 모델
5. lesson 간 공통으로 쓰는 작은 DTO나 formatter

반대로, **해당 lesson에서만 쓰는 설명 보조 코드나 실험 코드는 쉽게 `support`로 올리지 않습니다.** support가 커지기 시작하면, 나중에는 lesson 코드보다 support가 더 먼저 눈에 들어오게 됩니다. 학습 프로젝트에서는 이게 구조적 잡음이 됩니다.

### 4. "공통화"보다 "학습 맥락 유지"를 먼저 본다

보통 제품 코드에서는 중복 제거를 빨리 생각합니다. 하지만 이 프로젝트는 학습 코드이기 때문에, 중복이 조금 있더라도 lesson 맥락이 더 선명하면 그 편이 낫습니다.

예를 들어 첫 강의와 둘째 강의가 모두 작은 서버 bootstrap을 쓴다고 해서 바로 하나의 거대한 helper로 뽑는 것은 좋은 기본값이 아닙니다. 독자가 lesson 안에서 코드를 읽으며 구조를 파악해야 하는데, 핵심 흐름이 support 쪽으로 숨어 버릴 수 있기 때문입니다.

여기서는 **제품 수준의 DRY보다 학습 수준의 가시성**을 먼저 봅니다.

## 문서 구조 원칙

문서도 같은 기준으로 움직입니다.

1. 질문에 답하는 강의 문서는 `docs/lessons`
2. 집행 기록은 `docs/works`
3. 프로젝트 전체에 적용되는 기준 문서는 `docs` 루트 또는 별도 guide 문서

즉, 문서도 lesson 중심으로 읽히게 만들어야 합니다. 독자가 먼저 보는 것은 `lesson-01`, `lesson-02`여야지, 내부 작업 메모나 잡다한 설계 조각이 아니어야 합니다.

## 프론트 시각화를 나중에 붙일 걸 고려하면 지금 무엇을 준비해야 하나

이 질문이 중요합니다. 지금 당장은 프론트를 만들지 않더라도, 나중에 요청/응답 흐름을 화면에 뿌리고 싶다면 지금부터 최소한의 방향은 맞춰 두는 편이 좋습니다.

추천하는 원칙은 이겁니다.

**관측 결과를 "그냥 문자열 로그"로만 다루지 말고, 나중에는 "시각화 가능한 이벤트"로 승격할 수 있게 구조를 잡는다.**

이 말은 처음부터 거대한 관측 프레임워크를 만들라는 뜻이 아닙니다. 오히려 반대입니다. 지금은 단순하게 가되, 나중에 아래 구조로 옮기기 쉬워야 한다는 뜻입니다.

```text
support/
  observability/
    ObservationEvent
    ObservationSink
    ConsoleObservationSink
```

그러면 나중에는 이렇게 확장할 수 있습니다.

```text
viewer/
  receives ObservationEvent
  renders timeline / handler flow / thread flow
```

즉, 지금은 콘솔 로그로 시작하되, **언젠가 같은 관측 데이터를 프론트에 보낼 수 있는 방향**을 미리 의식합니다.

다만 현재 lesson 1처럼 아직 단순한 단계에서는, 문자열 로그만으로도 충분하면 그대로 두는 것이 맞습니다. 중요한 것은 "당장 다 만들 것"이 아니라, **나중에 옮길 때 어디를 바꾸면 되는지 예측 가능한 구조를 택하는 것**입니다.

## 언제 서브프로젝트로 가는가

서브프로젝트는 아래 조건이 분명해질 때 도입하는 것이 좋습니다.

### 1. 프론트 시각화 앱이 실제로 생긴다

예를 들어 브라우저에서 요청 흐름 타임라인, handler 호출 순서, thread 이동을 보여 주는 viewer가 생기면, 그때는 `lesson examples`와 `viewer app`의 런타임 성격이 달라집니다.

그 시점에는 이런 구조가 자연스럽습니다.

```text
learn-netty/
  pom.xml
  lesson-examples/
  lesson-viewer/
  lesson-shared/
```

### 2. 공통 관측/event 모델이 여러 lesson에서 본격적으로 재사용된다

`ObservationEvent`, trace DTO, shared formatter, sample protocol 모델이 여러 lesson과 viewer에서 함께 쓰이기 시작하면, 그때는 `lesson-shared` 같은 공통 모듈을 분리할 가치가 생깁니다.

### 3. 벤치마크/실험 러너가 별도 의존성을 강하게 요구한다

예를 들어 JMH, 대량 트래픽 생성기, 별도 native 설정, 프론트 번들링 등이 붙기 시작하면, single-module 안에 계속 우겨 넣는 것보다 `benchmark`, `viewer`, `examples`를 나누는 편이 더 낫습니다.

## 언제까지 single-module로 버티는가

아래 상태라면 single-module이 더 낫습니다.

1. 예제와 강의가 아직 강하게 연결돼 있다
2. shared code가 작고 단순하다
3. 빌드 도구보다 lesson 비교가 더 중요하다
4. viewer나 benchmark가 아직 실제 artifact가 아니다

지금이 바로 이 상태에 가깝습니다. 그래서 **현재 기본값은 single-module 유지**가 맞습니다.

## 패키지 예시

앞으로 몇 단계까지는 아래 감각을 기본값으로 씁니다.

```text
src/main/java/io/aimpugn/learn/netty/
  support/
    observability/
    testkit/
  lesson1/
    blocking/
    nio/
    netty/
  lesson2/
    eventloop/
    bossworker/
    experiment/
  lesson3/
    pipeline/
    inbound/
    outbound/
  lesson4/
    framing/
    protocol/
    codec/
  lesson5/
    bytebuf/
    refcnt/
    leak/
```

이 예시는 완전한 강제가 아니라, **최상위 축은 lesson, 하위 축은 비교/실험 주제, 공통만 support**라는 원칙을 보여 주는 기준선입니다.

## 지금 당장 적용할 운영 규칙

앞으로 새 자산을 추가할 때는 먼저 아래 순서로 판단합니다.

1. 이 자산은 몇 번째 lesson에 속하는가
2. lesson 내부에서 어떤 비교 축이나 실험 축에 속하는가
3. 둘 이상의 lesson이 함께 쓰는가
4. viewer나 benchmark처럼 런타임 성격이 완전히 다른가

판정 규칙은 이렇게 정리할 수 있습니다.

1. 한 lesson 안에서만 쓰면 lesson 패키지에 둡니다.
2. 여러 lesson이 함께 쓰면 `support`를 검토합니다.
3. 프론트 앱이나 benchmark처럼 실행 세계가 다르면 서브프로젝트를 검토합니다.

## 현재 결론

현재 기준 문장은 이것으로 고정합니다.

**이 프로젝트는 지금 단계에서 "lesson 중심 single-module + 공통 support 패키지"를 기본 구조로 삼고, 프론트 viewer, 공통 event 모델, 벤치마크 러너가 실제 artifact로 자랄 때만 서브프로젝트로 승격한다.**

다음에 구조를 바꾸더라도, 바뀌지 않아야 하는 핵심은 하나입니다.

**독자가 lesson의 발전 과정을 눈으로 따라가기 쉬운 구조여야 한다.**
