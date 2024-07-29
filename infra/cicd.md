# CI/CD

- [CI/CD](#cicd)
    - [Continuous Integration, Delivery, and Deployment](#continuous-integration-delivery-and-deployment)
    - [Continuous Integration (CI)](#continuous-integration-ci)
    - [Continuous Delivery (CD)](#continuous-delivery-cd)
    - [Continuous Deployment](#continuous-deployment)
    - [기타](#기타)

## Continuous Integration, Delivery, and Deployment

1. Continuous Integration (CI):

    코드 변경사항을 자주 메인 브랜치에 병합하여 통합하는 프로세스입니다.
    이를 통해 코드의 일관성을 유지하고, 버그를 조기에 발견할 수 있습니다.

    코드 통합과 초기 오류 감지에 중점을 둡니다.
    자동화된 빌드 및 테스트를 통해 *모든 개발자의 코드가 함께 작동하는지 확인*합니다.

2. Continuous Delivery (CD):

    소프트웨어를 언제든 릴리스할 수 있는 상태로 유지하는 방식입니다.
    CI의 결과물을 안정적인 배포 가능 상태로 유지하며, 실제 프로덕션 배포는 수동으로 결정합니다.

    코드를 신뢰할 수 있는 방식으로 지속적으로 배포 준비 상태로 유지하는 것입니다.
    이는 소프트웨어 출시의 신속성과 안정성을 보장합니다.

3. Continuous Deployment:

    코드 변경사항을 *자동으로 프로덕션 환경에 배포*하는 프로세스입니다.
    CI와 CD의 확장된 형태로, 모든 과정이 자동화되어 인간의 개입 없이 프로덕션 배포가 이루어집니다.

    변경사항을 가능한 한 빨리 프로덕션 환경에 배포하여, 실시간으로 사용자에게 새로운 기능과 버그 수정을 제공하는 것입니다.

일반적으로 CI/CD에서 CD는 Continuous Delivery를 가리킵니다.
Continuous Delivery 경우 자동화된 테스트와 준비 환경까지의 배포를 포함하지만,
최종 프로덕션 배포는 수동으로 이루어지기 때문에 더 안전하고 통제 가능한 방식으로 여겨집니다.

## Continuous Integration (CI)

CI는 개발자들이 코드 변경사항을 자주 메인 브랜치(예: main, master)에 병합하는 소프트웨어 개발 방식입니다.

주기적으로 작은 변경 사항을 통합하면 다음과 같은 이점이 있습니다:
- 코드 통합과정에서 발생할 수 있는 충돌을 최소화
- 소규모 변경의 문제를 조기에 발견

보통 다음과 같은 단계들을 거칩니다:
1. 개발자가 코드를 변경하고 버전 관리 시스템(예: Git)에 푸시합니다.
2. CI 서버(예: Jenkins, GitLab CI, GitHub Actions 등)가 변경을 감지하고 자동으로 빌드를 시작합니다.
3. 단위 테스트, 통합 테스트 등이 실행됩니다.
4. 빌드와 테스트 결과가 팀에게 보고됩니다.

## Continuous Delivery (CD)

CD는 소프트웨어를 언제든지 안전하게 릴리스할 수 있는 상태로 유지하는 방식입니다.
CI의 확장으로, *테스트를 통과한 코드를 자동으로 준비 환경(Staging)에 배포*합니다.

1. CI 프로세스를 통과한 코드가 Staging 환경에 자동으로 배포됩니다.
2. 준비 환경에서 추가 테스트(예: 성능 테스트, 사용자 수용 테스트)가 실행됩니다.
3. 프로덕션 배포는 수동으로 트리거됩니다.

## Continuous Deployment

Continuous Deployment는 CD의 한 단계 더 나아간 형태로, 코드 변경사항이 자동으로 프로덕션 환경에 배포됩니다.
인간의 개입이 필요 없으므로 매우 빠른 피드백 주기를 가능하게 하며, 지속적인 개선을 촉진합니다.
하지만 잘못된 코드가 프로덕션에 반영될 위험이 있으므로, 높은 수준의 테스트 자동화가 필수적입니다.

1. 개발자가 코드를 커밋합니다.
2. CI/CD 파이프라인이 자동으로 실행됩니다.
3. 모든 테스트를 통과하면 코드가 자동으로 프로덕션에 배포됩니다.

## 기타

- [Continuous Delivery vs. Continuous Deployment: A Comparison](https://katalon.com/resources-center/blog/continuous-delivery-vs-continuous-deployment)
- [Continuous integration vs. delivery vs. deployment](https://www.atlassian.com/continuous-delivery/principles/continuous-integration-vs-delivery-vs-deployment)
