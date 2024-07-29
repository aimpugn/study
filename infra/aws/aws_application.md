# AWS application

- [AWS application](#aws-application)
    - [myApplication](#myapplication)
    - [기능](#기능)
    - [관련 서비스](#관련-서비스)
    - [생성하기](#생성하기)
        - [1. Specify application details](#1-specify-application-details)
        - [2. (Optional) Add resources](#2-optional-add-resources)
        - [3. Review and create](#3-review-and-create)
    - [기타](#기타)

## [myApplication](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/aws-myApplications.html)

`myApplication`이란 AWS에서 애플리케이션을 시작하고, 더 적은 노력으로 애플리케이션을 운영하며, 대규모로 더 빠르게 옮길 수 있도록 지원하는 새로운 기능 세트입니다.
애플리케이션의 비용, 상태, 보안 상태 및 성능을 보다 쉽게 관리 및 모니터링할 수 있습니다.

## [기능](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/aws-myApplications.html)

- Create applications

    애플리케이션을 생성하고 리소스를 조직합니다.
    Console, API, CLI, SDK 등을 통해 작업을 수행할 수 있습니다.
    애플리케이션을 생성하면 IaC가 생성되고 myApplication 대시보드에서 접근 가능합니다.
    IaC는 AWS CloudFormation과 Terraform 같은 IaC 도구에서 사용 가능합니다.

- Access your applications

    `myApplication` 위젯에서 선택하여 빠르게 접근 가능합니다.

- Compare application metrics

    비용 및 중요 보안 발견 수 같은 주요 지표를 비교할 수 있습니다.

- Monitor and manage applications

    Assess application health and performance using alarms, canaries, and service level objectives from Amazon CloudWatch, findings from AWS Security Hub, and cost trends from AWS Cost Explorer Service. You can also find compute metrics summaries and optimizations and manage resource compliance and configuration status from AWS Systems Manager.

    다음 같은 도구들을 활용하여 애플리케이션 상태 및 성능을 평가할 수 있습니다.
    - Amazon CloudWatch의 알람, 카나리아, 서비스 수준 목표
    - AWS Security Hub의 결과
    - AWS Cost Explorer Service의 비용 추세 등

    또한 AWS Systems Manager 통해 컴퓨팅 메트릭 요약 및 최적화를 확인하고 리소스 규정 준수 및 구성 상태를 관리할 수 있음

## [관련 서비스](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/aws-myApplications.html#myApp-related-services)

- AppRegistry
- AppManager
- Amazon CloudWatch
- Amazon EC2
- AWS Lambda
- AWS Resource Explorer
- AWS Security Hub
- Systems Manager
- AWS Service Catalog
- Tagging

## 생성하기

### 1. Specify application details

애플리케이션 이름(필수), 애플리케이션에 대한 설명(옵션), 태그(옵션), 속성 그룹(옵션)을 설정합니다.

### 2. (Optional) Add resources

> NOTE:
>
> AWS Resource Explorer가 활성화 되어 있어야 합니다.

리소스를 선택합니다. 필수는 아니어서 생성시에는 넘어가도 됩니다.

### 3. Review and create

## 기타

- [AWS 관리 콘솔의 myApplications 애플리케이션 리소스 관리 단순화](https://aws.amazon.com/ko/blogs/korea/new-myapplications-in-the-aws-management-console-simplifies-managing-your-application-resources/)
- [Create Application](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/myApp-getting-started.html#myApp-step1?icmpid=docs_console_home_create_application_help_panel)
