# security

## change security group

Actions > Security > Change security group

## 새로 띄운 서버가 기존 DB에 접근하고 싶다면?

`some-security-group`(Security group id: sg-2222)이라는 항목에 아래와 같은 설정이 있고,

| Security group rule ID | IP version | Type         | Protocol | Port range | Source                    | Description    |
|------------------------|------------|--------------|----------|------------|---------------------------|----------------|
| sgr-A                  | –          | MYSQL/Aurora | TCP      | 3306       | sg-1111 / stg-k8s-cluster | from k8s       |
| sgr-B                  | -          | SSH          | TCP      | 22         | sg-2222 / db              |                |
| sgr-C                  | IPv4       | MYSQL/Aurora | TCP      | 3306       | 10.0.0.0/16               | from legacy    |
| sgr-D                  | -          | MYSQL/Aurora | TCP      | 3306       | sg-3333 / A-service       | from A service |
| sgr-E                  | -          | MYSQL/Aurora | TCP      | 3306       | sg-2222 / db              | from self      |

그리고 DB inbound 규칙에 `some-security-group` 그룹이 추가되어 있다

이때 이 DB에 접근하고 싶다면, 새로 띄운 서버의 security groups에도 `some-security-group` 그룹을 추가한다.
