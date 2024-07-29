# AWS IAM

## IAM

## 역할 생성

### 신뢰할 수 있는 엔터티 선택

*이 계정에서 작업을 수행하도록 허용할* 엔터티를 선택합니다.

- AWS 서비스
- AWS 계정

    사용자 또는 써드 파티에 속한 다른 AWS 계정의 엔터티가
    이 계정에서 작업을 수행하도록 허용합니다.

- 웹 자격 증명

    지정된 외부 웹 자격 증명 고급자와 연동된 사용자가 이 역할을 맡아
    이 계정에서 작업을 수행하도록 허용합니다.

- SAML 2.0 연동
- 사용자 지정 신뢰 정책

### 권한 추가

### 이름 지정, 검토 및 생성

- 역할 세부 정보
    - 역할 이름
    - 설명
- 1단계: 신뢰할 수 있는 엔터티 선택

    신뢰 정책:

    ```json
    {
        "Version": "2012-10-17",
        "Statement": [
            {
                "Effect": "Allow",
                "Action": "sts:AssumeRole",
                "Principal": {
                    "AWS": "<My AWS ID>"
                },
                "Condition": {}
            }
        ]
    }
    ```

- 2단계: 권한 추가

    [SystemAdministrator](https://us-east-1.console.aws.amazon.com/iam/home?region=ap-northeast-2#/policies/details/arn%3Aaws%3Aiam%3A%3Aaws%3Apolicy%2Fjob-function%2FSystemAdministrator)
