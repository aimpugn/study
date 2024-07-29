# Design pagination

## 페이지네이션 개발 설계

```bash
infra(repository)
service?
presentation
```

이렇게 세 개의 레이어가 있습니다.
전체 목록을 조회하고, 페이지를 나누고, 전체 몇 건이고, 현재 페이지가 몇 페이지인지를 판단하는 건 아마도 서비스 레이어라고 생각하는데요. 왜냐하면 프리젠테이션은 요청을 받아서 검증하고 응답 형식에 맞게 가공하는 역할만 해야 하고, 인프라는 데이터를 가져오는 역할만 해야 한다면, 소거법으로 그 중간 레이어인 서비스 레이어에 있어야 하는 게 아닌가 싶습니다.
1. 우선 이렇게 소거해가며 레이어를 두는 게 맞을까요? 아니면 이런 데이터 변환의 의미에 맞게 레이어를 찾아가는 게 맞을까요?
페이지네이션의 로직은 서비스 레이어에서 이뤄지는데, 이를 응답하는 것은 프리젠테이션 레이어입니다.
2. 그렇다면 페이지네이션의 로직과 이를 응답하는 응답 형식을 분리해야 하는 걸까요? 그럼 서비스에서의 페이지네이션 타입과 프리젠테이션 레이어에서의 페이지 타입을 구별해야 할까요?
3. 그리고 이런 방식으로 생각하는 게 맞을까요? 이런 때에 어떤 논리적 사고 과정을 거쳐서 판단해야 하는 게 맞을지도 궁금합니다. 즉, 서비스 아키텍처를 설계하면서 다양한 선택의 과정이 있는데, 이때 무엇을 근거로, 어떻게, 어떤 논리적 사고 과정을 통해 판단을 내려야 저 뿐만 아니라 같은 팀원들도 납득을 할 수 있는 선택을 할 수 있을지 배우고 싶습니다.

이 질문들은 소프트웨어 아키텍처와 설계에 대한 매우 중요하고 깊이 있는 고민을 담고 있습니다. 각 질문에 대해 상세히 답변하고, 이런 결정을 내릴 때 사용할 수 있는 논리적 사고 과정을 설명하겠습니다.

1. 레이어 결정 방식에 대한 질문:

소거법으로 레이어를 결정하는 것은 시작점으로는 괜찮지만, 궁극적으로는 각 레이어의 책임과 역할에 기반하여 결정해야 합니다. 이는 단일 책임 원칙(Single Responsibility Principle)과 관심사의 분리(Separation of Concerns) 원칙에 근거합니다.

논리적 사고 과정:
a) 각 레이어의 주요 책임 정의:
- Presentation: 사용자 인터페이스, 요청 처리, 응답 형식화
- Service: 비즈니스 로직, 트랜잭션 관리, 도메인 규칙 적용
- Infrastructure: 데이터 접근, 외부 시스템 통합

b) 페이지네이션 관련 작업 분류:
- 데이터 조회: Infrastructure (Repository)
- 페이지 계산 로직: Service
- 응답 형식 결정: Presentation

c) 결정:
페이지네이션의 핵심 로직(전체 목록 조회, 페이지 나누기, 전체 건수 및 현재 페이지 계산)은 비즈니스 로직에 해당하므로 Service 레이어에 위치시키는 것이 적절합니다.

2. 페이지네이션 로직과 응답 형식 분리에 대한 질문:

페이지네이션 로직과 응답 형식을 분리하는 것은 좋은 접근 방식입니다. 이는 관심사의 분리 원칙과 인터페이스 분리 원칙(Interface Segregation Principle)에 부합합니다.

논리적 사고 과정:
a) 서비스 레이어의 책임:
- 비즈니스 로직 처리 (페이지네이션 계산)
- 도메인 중심의 데이터 구조 사용

b) 프레젠테이션 레이어의 책임:
- API 응답 형식 결정
- 클라이언트 요구사항에 맞는 데이터 변환

c) 결정:
서비스 레이어에서는 도메인 중심의 페이지네이션 타입을 사용하고, 프레젠테이션 레이어에서는 이를 API 응답에 적합한 형태로 변환하는 것이 좋습니다.

예시:

```kotlin
// Service Layer
data class PagedResult<T>(
    val items: List<T>,
    val totalItems: Long,
    val pageNumber: Int,
    val pageSize: Int
)

// Presentation Layer
data class PagedResponse<T>(
    val content: List<T>,
    val totalElements: Long,
    val totalPages: Int,
    val pageNumber: Int,
    val pageSize: Int,
    val first: Boolean,
    val last: Boolean
)

// Conversion in Presentation Layer
fun <T> PagedResult<T>.toPagedResponse(): PagedResponse<T> {
    val totalPages = (totalItems + pageSize - 1) / pageSize
    return PagedResponse(
        content = items,
        totalElements = totalItems,
        totalPages = totalPages.toInt(),
        pageNumber = pageNumber,
        pageSize = pageSize,
        first = pageNumber == 1,
        last = pageNumber == totalPages
    )
}
```

3. 아키텍처 설계 시 논리적 사고 과정:

아키텍처 설계 시 다음과 같은 단계와 원칙을 따르는 것이 도움이 됩니다:

a) 요구사항 분석:
- 기능적 요구사항과 비기능적 요구사항을 명확히 정의합니다.
- 예: 페이지네이션 기능, 성능 요구사항, 확장성 등

b) 설계 원칙 적용:
- SOLID 원칙 (단일 책임, 개방-폐쇄, 리스코프 치환, 인터페이스 분리, 의존성 역전)
- 관심사의 분리
- DRY (Don't Repeat Yourself)
- YAGNI (You Aren't Gonna Need It)

c) 트레이드오프 분석:
- 각 설계 결정의 장단점을 평가합니다.
- 예: 복잡성 vs 유연성, 성능 vs 유지보수성

d) 팀 합의:
- 설계 결정에 대해 팀 내 토론을 진행합니다.
- 결정 이유와 고려사항을 문서화합니다.

e) 반복 및 피드백:
- 초기 설계를 구현하고 피드백을 수집합니다.
- 필요에 따라 설계를 조정합니다.

논리적 사고 과정 예시 (페이지네이션 설계):

1. 요구사항 정의:
   - 대량의 데이터를 효율적으로 처리해야 함
   - 클라이언트에게 일관된 응답 형식 제공 필요
   - 다양한 데이터 소스에 대해 동일한 페이지네이션 로직 적용 가능해야 함

2. 설계 원칙 적용:
   - 단일 책임 원칙: 페이지네이션 로직을 서비스 레이어에 배치
   - 인터페이스 분리 원칙: 서비스 레이어와 프레젠테이션 레이어의 페이지네이션 모델 분리

3. 트레이드오프 분석:
   - 장점: 관심사 분리로 유지보수성 향상, 재사용성 증가
   - 단점: 약간의 복잡성 증가, 변환 로직 필요

4. 결정 및 근거:
   - 서비스 레이어에 PagedResult 모델 도입
   - 프레젠테이션 레이어에 PagedResponse 모델 도입
   - 근거: 비즈니스 로직과 표현 로직의 명확한 분리, 향후 확장성 고려

5. 검증 및 피드백:
   - 팀 내 코드 리뷰를 통한 설계 검증
   - 실제 구현 후 성능 및 사용성 테스트

이러한 체계적인 접근 방식을 통해 팀원들이 설계 결정을 이해하고 납득할 수 있는 근거를 제공할 수 있습니다. 또한, 이 과정을 문서화하여 향후 참조할 수 있도록 하는 것도 중요합니다.
