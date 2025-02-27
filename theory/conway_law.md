# Conway's Law

- [Conway's Law](#conways-law)
    - [컨웨이의 법칙](#컨웨이의-법칙)
    - [예제](#예제)
        - [Golang package oriented design과 컨웨이의 법칙](#golang-package-oriented-design과-컨웨이의-법칙)

## 컨웨이의 법칙

컨웨이의 법칙(Conway's Law)은 1967년 Melvin Conway가 그의 논문 "How Do Committees Invent?"에서 제안한 개념으로, 소프트웨어 설계와 조직 구조 간의 상호작용을 설명하는 이론입니다.

이 법칙의 핵심 내용은 다음과 같습니다:

- "시스템을 설계하는 조직은 그 조직의 의사소통 구조와 일치하는 구조의 설계를 만들어낼 수밖에 없다."
- "조직의 의사소통 구조는 그 조직이 설계하는 시스템의 구조에 그대로 반영된다"

이는 소프트웨어 시스템의 구조가 그것을 개발하는 조직의 구조를 반영한다는 것을 의미합니다.
복잡한 시스템을 설계할 때 조직 내의 의사소통 경로와 구조가 시스템 설계의 모듈화 방식에 큰 영향을 미친다는 점을 강조했습니다.

특히 마이크로서비스 아키텍처와 같은 분산 시스템에서 그 중요성이 부각되었습니다.
조직 구조가 고정되어 있을 경우, 이를 반영한 시스템 구조가 발생할 수밖에 없다는 점은 시스템의 확장성과 유연성을 제한할 수 있다는 문제를 낳을 수 있습니다.

오늘날 컨웨이의 법칙은 소프트웨어 설계의 근본적인 원리 중 하나로 자리 잡았으며, 소프트웨어 아키텍처 설계 시 조직 구조를 함께 고려해야 하는 필요성을 일깨워줍니다.
이 법칙은 DevOps와 마이크로서비스 아키텍처와 같은 현대 소프트웨어 개발 패러다임에서 더욱 중요한 역할을 하고 있습니다.

컨웨이의 법칙은 다양한 기업과 오픈 소스 프로젝트에서 실제로 적용된 사례들이 많이 있습니다.
1. 아마존의 마이크로서비스 아키텍처

    아마존은 초기에는 모놀리식(monolithic) 구조의 시스템을 사용했으나, 조직의 확장에 따라 이를 마이크로서비스 아키텍처로 전환했습니다.
    이 과정에서 아마존은 각 비즈니스 도메인을 독립된 마이크로서비스로 나누었고, 이는 조직 내 팀들이 자율적으로 각 서비스의 개발과 운영을 담당할 수 있게 했습니다.
    결과적으로, 조직의 구조와 서비스 아키텍처가 일치하게 되면서 시스템의 유연성과 확장성이 크게 향상되었습니다.

2. 넷플릭스의 사례

    넷플릭스는 전 세계적으로 확장 가능한 스트리밍 서비스를 제공하기 위해, 조직을 마이크로서비스 중심으로 재편했습니다.
    이 과정에서 각 팀은 특정 서비스나 기능을 담당하며, 이 서비스들이 서로 독립적으로 배포 및 운영될 수 있게 설계되었습니다.
    이를 통해 넷플릭스는 빠르게 변화하는 시장 요구에 유연하게 대응할 수 있었고, 전 세계적으로 안정적인 서비스를 제공할 수 있었습니다.

## 예제

### Golang package oriented design과 컨웨이의 법칙

1. 구조적 일치성:
    - 컨웨이의 법칙: 조직 구조가 소프트웨어 구조에 반영됩니다.
    - 패키지 오리엔티드 설계: 비즈니스 도메인이나 기능을 중심으로 패키지를 구성합니다.
    - 연관성: 조직의 팀 구조가 특정 비즈니스 도메인이나 기능을 중심으로 구성되어 있다면, 패키지 구조도 이를 반영하게 됩니다.

2. 의사소통 패턴:
    - 컨웨이의 법칙: 팀 간 의사소통 패턴이 시스템 인터페이스에 반영됩니다.
    - 패키지 오리엔티드 설계: 각 패키지는 명확한 인터페이스를 통해 상호작용합니다.
    - 연관성: 팀 간 의사소통 방식이 패키지 간 인터페이스 설계에 영향을 미칩니다.

3. 모듈화와 독립성:
    - 컨웨이의 법칙: 독립적인 팀은 독립적인 모듈을 만듭니다.
    - 패키지 오리엔티드 설계: 각 패키지는 독립적으로 기능하는 단위입니다.
    - 연관성: 조직의 독립적인 팀 구조가 독립적인 패키지 구조로 이어집니다.

    각 패키지는 특정 기능이나 도메인에 집중된 모듈로 설계되며, 이는 독립적인 개발, 테스트, 배포를 가능하게 합니다.
    이러한 독립성은 시스템의 한 부분이 변경될 때 다른 부분에 미치는 영향을 최소화하는 데 매우 중요합니다.

    예를 들어, 사용자 관리(user), 제품 관리(product), 주문 처리(order)와 같은 패키지들은 각각 독립적으로 설계됩니다.
    이러한 패키지는 공통 인터페이스를 통해 상호작용하지만, 각각의 패키지는 자체적으로 모듈화되어 있어, 내부적으로 발생하는 변화가 다른 패키지에 영향을 미치지 않습니다.
    이는 대규모 시스템에서 기능 추가나 수정 시 전체 시스템을 재배포하지 않고도 특정 패키지만을 업데이트할 수 있는 장점을 제공합니다.

    이와 같은 모듈화는 팀 간의 작업 분할을 명확하게 할 수 있게 하며, 개발 효율성을 크게 향상시킵니다.
    각 팀은 자신의 도메인에 맞는 패키지를 관리하며, 독립적으로 기능을 확장하거나 최적화할 수 있습니다.
    결과적으로, Golang의 패키지 오리엔티드 설계는 소프트웨어의 구조적 복잡성을 줄이고, 시스템의 가독성과 유지보수성을 높이는 데 기여합니다.

4. 확장성:
    - 컨웨이의 법칙: 조직 구조의 변화가 시스템 구조의 변화를 유발합니다.
    - 패키지 오리엔티드 설계: 새로운 기능이나 도메인을 쉽게 추가할 수 있습니다.
    - 연관성: 조직의 성장이나 변화에 따라 새로운 팀이 추가되면, 이에 대응하는 새로운 패키지가 자연스럽게 생성됩니다.

5. 책임 분배:
    - 컨웨이의 법칙: 팀의 책임 영역이 소프트웨어 컴포넌트의 책임과 일치합니다.
    - 패키지 오리엔티드 설계: 각 패키지는 특정 도메인이나 기능에 대한 책임을 집니다.
    - 연관성: 조직 내 팀의 책임 영역이 소프트웨어의 패키지 구조에 직접적으로 반영됩니다.

```go
// 조직 구조:
// - 사용자 관리팀
// - 제품 관리팀
// - 주문 처리팀

// 패키지 구조:
project/
├── user/        // 사용자 관리팀의 책임 영역
│   ├── model/
│   ├── repository/
│   └── service/
├── product/     // 제품 관리팀의 책임 영역
│   ├── model/
│   ├── repository/
│   └── service/
└── order/       // 주문 처리팀의 책임 영역
    ├── model/
    ├── repository/
    └── service/
```

이 구조에서 각 패키지는 조직 내 특정 팀의 책임 영역을 반영합니다.
팀 간 의사소통은 각 패키지의 서비스 레이어를 통해 이루어집니다.

이론적 영향과 실무적 고려사항은 다음과 같습니다:

1. 조직 설계와 소프트웨어 설계의 상호작용:
    - 이론: 컨웨이의 법칙은 조직 설계가 소프트웨어 설계에 미치는 영향을 강조합니다.
    - 실무: 패키지 오리엔티드 설계를 적용할 때, 조직 구조를 고려하여 패키지를 설계해야 합니다.

2. 역컨웨이 매너버 (Reverse Conway Maneuver):
    - 이론: 원하는 시스템 아키텍처를 달성하기 위해 조직 구조를 변경하는 전략입니다.
    - 실무: 패키지 구조를 먼저 설계한 후, 이에 맞춰 조직 구조를 조정할 수 있습니다.

3. 마이크로서비스 아키텍처와의 연관성:
    - 이론: 컨웨이의 법칙은 마이크로서비스 아키텍처의 이론적 기반 중 하나입니다.
    - 실무: 패키지 오리엔티드 설계는 마이크로서비스로의 전환을 용이하게 만들 수 있습니다.

4. 팀 자율성과 책임:
    - 이론: 컨웨이의 법칙은 팀의 자율성이 소프트웨어 구조에 반영됨을 시사합니다.
    - 실무: 각 패키지에 대한 책임을 특정 팀에 할당하여 자율성과 책임을 부여합니다.

5. 시스템 경계 설정:
    - 이론: 컨웨이의 법칙은 조직 구조가 시스템 경계 설정에 영향을 미침을 보여줍니다.
    - 실무: 패키지 간 경계를 명확히 설정하여 팀 간 책임 영역을 구분합니다.

결론적으로, 패키지 오리엔티드 설계는 컨웨이의 법칙을 실제 소프트웨어 구조에 적용하는 방법 중 하나로 볼 수 있습니다.
이 설계 방식은 조직 구조와 소프트웨어 구조 간의 자연스러운 일치를 추구하며, 이를 통해 개발 효율성과 시스템의 유지보수성을 향상시킬 수 있습니다.
