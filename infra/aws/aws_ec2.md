# AWS EC2

- [AWS EC2](#aws-ec2)
    - [AMI](#ami)
    - [인스턴스 패밀리](#인스턴스-패밀리)
        - [1. T 시리즈 (T2, T3, T4g)](#1-t-시리즈-t2-t3-t4g)
        - [2. C 시리즈 (C4, C5, C6g)](#2-c-시리즈-c4-c5-c6g)
        - [3. R 시리즈 (R4, R5, R6g)](#3-r-시리즈-r4-r5-r6g)
        - [4. M 시리즈 (M4, M5, M6g)](#4-m-시리즈-m4-m5-m6g)
        - [5. G 시리즈 (G3, G4)](#5-g-시리즈-g3-g4)
        - [6. I 시리즈 (I3, I3en)](#6-i-시리즈-i3-i3en)
    - [생성](#생성)
        - [애플리케이션 및 OS 이미지](#애플리케이션-및-os-이미지)
        - [인스턴스 유형](#인스턴스-유형)
        - [키페어(로그인)](#키페어로그인)
        - [네트워크 설정](#네트워크-설정)
            - [웹 서버 또는 애플리케이션 서버 생성](#웹-서버-또는-애플리케이션-서버-생성)
            - [웹 서버 또는 애플리케이션 서버를 생성하여 RDS DB 인스턴스에 연결](#웹-서버-또는-애플리케이션-서버를-생성하여-rds-db-인스턴스에-연결)
            - [RDS 데이터베이스에 대한 Bastion Host 생성](#rds-데이터베이스에-대한-bastion-host-생성)
        - [스토리지 구성](#스토리지-구성)
            - [`IOPS`](#iops)
        - [API](#api)
            - [RunInstances](#runinstances)
            - [새 보안 그룹 설정](#새-보안-그룹-설정)
        - [고급](#고급)
    - [로그인](#로그인)
        - [Ubuntu AMI 설치한 경우](#ubuntu-ami-설치한-경우)
        - [로그인 후 root 유저 되기](#로그인-후-root-유저-되기)
    - [EC2 root 접근 관련 AWS 문서](#ec2-root-접근-관련-aws-문서)
    - [RDS 데이터베이스 연결](#rds-데이터베이스-연결)
    - [배포(Deployment) 방법](#배포deployment-방법)
        - [직접 배포](#직접-배포)
        - [AWS에 특화된 방식](#aws에-특화된-방식)
        - [이미 실행중인 EC2 인스턴스에 배포하기](#이미-실행중인-ec2-인스턴스에-배포하기)
    - [AWS 공식 가이드 문서](#aws-공식-가이드-문서)
    - [초보자를 위한 권장 사항](#초보자를-위한-권장-사항)
    - [참고 자료](#참고-자료)

## AMI

- Amazon Linux 2023 AMI

    Amazon Linux 2023는 5년의 장기 지원을 제공하는 최신 범용 Linux 기반 OS입니다. AWS에 최적화되어 있으며 클라우드 애플리케이션을 개발 및 실행할 수 있는 안전하고 안정적인 고성능 실행 환경을 제공하도록 설계되었습니다.

- AMAAmazonZON Linux 2 AMI (HVM) - Kernel 5.10, SSD Volume Type

    Amazon Linux 2는 5년간 지원을 제공합니다. Amazon EC2에 성능 최적화된 Linux kernel 5.10와 systemd 219, GCC 7.3, Glibc 2.26, Binutils 2.29.1, 최신 소프트웨어 패키지를 추가적으로 제공합니다.

- Amazon Linux 2 with .NET6, PowerShell, Mono, and MATE Desktop Environment
- Red Hat Enterprise Linux 9 (HVM), SSD Volume Type

    Red Hat Enterprise Linux version 9 (HVM), EBS General Purpose (SSD) Volume Type

- Ubuntu Server 24.04 LTS (HVM), SSD Volume

    Ubuntu Server 24.04 LTS (HVM),EBS General Purpose (SSD) Volume Type. Support available from Canonical (<http://www.ubuntu.com/cloud/services>).

- Ubuntu Server 22.04 LTS (HVM), SSD Volume

    Ubuntu Server 22.04 LTS (HVM),EBS General Purpose (SSD) Volume Type. Support available from Canonical (<http://www.ubuntu.com/cloud/services>).

- SUSE Linux Enterprise Server 15 SP6 (HVM), SSD Volume Type

    SUSE Linux Enterprise Server 15 Service Pack 6 (HVM), EBS General Purpose (SSD) Volume Type. Amazon EC2 AMI Tools preinstalled.

## 인스턴스 패밀리

EC2 인스턴스는 여러 '패밀리'로 그룹화되며, 각 패밀리는 특정 사용 사례에 최적화되어 있습니다.
- T (범용)
- C (컴퓨팅 최적화)
- R (메모리 최적화)
- M (균형)
- G (GPU)
- I (스토리지 최적화) 등

[instances.vantage.sh](https://instances.vantage.sh/?region=ap-northeast-2) 통해 한 눈에 비교할 수 있습니다.

### 1. T 시리즈 (T2, T3, T4g)

- 정의: 버스트 가능한 성능 인스턴스
- 특징:
    - 기본 수준의 CPU 성능 제공
    - CPU 크레딧을 사용하여 필요 시 성능 버스트
    - T2는 이전 세대, T3는 현재 세대, T4g는 ARM 기반 프로세서 사용
- 사용 사례:
    - 웹 서버, 개발 환경, 소규모 데이터베이스
    - 변동성 있는 워크로드를 가진 애플리케이션
- 성능 메트릭:
    - vCPU: 2-8
    - 메모리: 0.5-32 GiB
    - 네트워크 성능: 최대 5 Gbps

### 2. C 시리즈 (C4, C5, C6g)

- 정의: 컴퓨팅 최적화 인스턴스
- 특징:
    - 높은 CPU 대 메모리 비율
    - C4는 이전 세대, C5는 현재 세대, C6g는 ARM 기반
- 사용 사례:
    - 고성능 웹 서버, 과학적 모델링
    - 배치 처리, 분산 분석, 고성능 컴퓨팅 (HPC)
- 성능 메트릭:
    - vCPU: 2-96
    - 메모리: 3.75-192 GiB
    - 네트워크 성능: 최대 100 Gbps

### 3. R 시리즈 (R4, R5, R6g)

- 정의: 메모리 최적화 인스턴스
- 특징:
    - 높은 메모리 대 CPU 비율
    - R4는 이전 세대, R5는 현재 세대, R6g는 ARM 기반
- 사용 사례:
    - 인메모리 데이터베이스, 분산 메모리 캐싱
    - 실시간 빅 데이터 분석, 대규모 SAP HANA 배포
- 성능 메트릭:
    - vCPU: 2-96
    - 메모리: 15.25-768 GiB
    - 네트워크 성능: 최대 100 Gbps

### 4. M 시리즈 (M4, M5, M6g)

- 정의: 범용 인스턴스
- 특징:
    - 컴퓨팅, 메모리, 네트워킹 리소스의 균형
    - M4는 이전 세대, M5는 현재 세대, M6g는 ARM 기반
- 사용 사례:
    - 중소규모 데이터베이스, 백엔드 서버
    - 기업용 애플리케이션, 게임 서버
- 성능 메트릭:
    - vCPU: 2-96
    - 메모리: 8-384 GiB
    - 네트워크 성능: 최대 25 Gbps

### 5. G 시리즈 (G3, G4)

- 정의: GPU 최적화 인스턴스
- 특징:
    - NVIDIA Tesla GPU 탑재
    - G3는 그래픽 집약적 애플리케이션용, G4는 기계 학습용
- 사용 사례:
    - 3D 시각화, 비디오 인코딩
    - 기계 학습 추론, 게임 스트리밍
- 성능 메트릭:
    - GPU: 1-4
    - vCPU: 4-96
    - 메모리: 16-384 GiB
    - GPU 메모리: 8-32 GiB

### 6. I 시리즈 (I3, I3en)

- 정의: 스토리지 최적화 인스턴스
- 특징:
    - 초고속 NVMe SSD 스토리지
    - I3은 고성능, I3en은 향상된 네트워킹 성능
- 사용 사례:
    - 고성능 NoSQL 데이터베이스
    - 데이터 웨어하우징, 분산 파일 시스템
- 성능 메트릭:
    - vCPU: 2-96
    - 메모리: 15.25-768 GiB
    - NVMe 스토리지: 1.9-60 TB
    - 네트워크 성능: 최대 100 Gbps

## 생성

### 애플리케이션 및 OS 이미지

### 인스턴스 유형

### 키페어(로그인)

> NOTE: 키 페어를 생성해서 해당 키 페어로만 접속할 수 있도록 합니다.

- [RSA](../../protocols/encryption/encryption.md)
- ED 25519

### 네트워크 설정

- VPC(Virtual Private Cloud, required)

    인스턴스를 시작할 VPC입니다.

- 새 보안 그룹

    22 포트 대신 별도 포트로 SSH 허용합니다.
    사용자 지정 프로토콜 & TCP & 포트 범위: 22022

#### 웹 서버 또는 애플리케이션 서버 생성

- 기본 VPC 또는 서브넷을 변경하지 않은 경우:
    인스턴스 시작 마법사에서 기본 네트워크 설정을 사용할 수 있습니다.

- 기본 VPC 또는 서브넷을 변경한 경우:

    1. (웹 서버)  인스턴스를 시작할 VPC에는 인터넷 게이트웨이가 연결되어 있어야 합니다. 기본 VPC에는 인터넷 게이트웨이가 자동으로 설정되어 있습니다.

    2. (웹 서버)인스턴스에 퍼블릭 IP 주소가 할당되어야 합니다.

    3. (웹 서버 또는 애플리케이션 서버 – 선택 사항)로그인할 디바이스의 IP 주소에서 SSH 또는 RDP를 허용하는 보안 그룹 규칙을 추가합니다(예: 업무용 노트북).

    4. (웹 서버)인터넷으로부터의 HTTP 및 HTTPS 트래픽을 허용하는 보안 그룹 규칙을 추가합니다(0.0.0.0/0).

#### 웹 서버 또는 애플리케이션 서버를 생성하여 RDS DB 인스턴스에 연결

1. [웹 서버 또는 애플리케이션 서버 생성](#웹-서버-또는-애플리케이션-서버-생성)에 따라 웹 서버 또는 애플리케이션 서버 생성 단계에 따라 서버를 생성합니다.

2. EC2 인스턴스가 DB 인스턴스와 동일한 VPC에 있는지 확인합니다. 기본적으로 VPC는 1개만 있습니다.

3. DB 인스턴스의 보안 그룹에는 데이터베이스 통신 포트에 대한 TCP 규칙이 있어야 합니다.

    - 유형:
        - Aurora, MySQL, Maria DB – 3306
        - PostgreSQL – 5432
        - Oracle – 1521
        - MSSQL – 1433
    - 소스
        - Bastion Host의 프라이빗 IP 주소

#### RDS 데이터베이스에 대한 Bastion Host 생성

- 기본 VPC 또는 서브넷을 변경하지 않은 경우:
    인스턴스 시작 마법사에서 기본 네트워크 설정을 사용할 수 있습니다.

- 기본 VPC 또는 서브넷을 변경한 경우:

    1. 인스턴스는 데이터베이스와 동일한 VPC에 있어야 합니다. 기본적으로 VPC는 1개뿐입니다.

    2. EC2 인스턴스를 시작하는 VPC에는 인터넷 게이트웨이가 연결되어 있어야 합니다. 기본 VPC는 인터넷 게이트웨이와 함께 자동으로 설정됩니다.

    3. EC2 인스턴스에 퍼블릭 IP 주소가 할당되어야 합니다.

    4. 로그인할 디바이스의 IP 주소에서 SSH 또는 RDP를 허용하는 보안 그룹 규칙을 추가합니다(예: 업무용 노트북).

    5. DB 인스턴스의 보안 그룹에는 데이터베이스 통신 포트에 대한 TCP 규칙이 있어야 합니다.

        - 유형:
            - Aurora, MySQL, Maria DB – 3306
            - PostgreSQL – 5432
            - Oracle – 1521
            - MSSQL – 1433
        - 소스
            - EC2 인스턴스의 프라이빗 IP 주소

### 스토리지 구성

- gp3
- gp2
- 프로비저닝된 IOPS SSD(io1)
- 프로비저닝된 IOPS SSD(io2)

#### `IOPS`

SSD 지원 볼륨은 I/O 크기가 작고 읽기/쓰기 작업이 빈번하게 발생하는 트랜잭션 워크로드에 최적화되어 있으며,
주요 성능 속성(dominant performance attribute)은 IOPS입니다.

볼륨이 지원할 수 있는 요청된 초당 I/O 작업 수입니다.
프로비저닝된 IOPS SSD(io1 및 io2) 및 범용 SSD(gp2 및 gp3) 볼륨에만 적용됩니다.

- 프로비저닝된 IOPS SSD(io1 및 io2) 볼륨

    볼륨 크기에 따라 다음과 같은 IOPS를 지원합니다.
    - io1 볼륨의 경우 100~64,000IOPS
    - io2 볼륨의 경우 100~256,000IOPS

    io1 볼륨의 경우 GiB당 최대 50IOPS까지 프로비저닝할 수 있으며,
    io2 볼륨의 경우 GiB당 최대 1,000IOPS까지 프로비저닝할 수 있습니다.

- 범용 SSD(gp2) 볼륨

    기준 성능은 GiB당 3IOPS에서 최소 100IOPS(33.33GiB 이하)부터 최대 16,000IOPS(5,334GiB 이상)까지 선형적으로 확장됩니다.

- 범용 SSD(gp3) 볼륨

    3,000IOPS의 기준 성능을 지원하며, GiB당 500IOPS(최대 16,000IOPS)까지 프로비저닝할 수 있습니다.

- 마그네틱(표준) 볼륨

    평균적으로 약 100IOPS를 제공하며 최대 수백 IOPS의 버스트 기능을 제공합니다.

- 처리량 최적화 HDD(st1) 및 콜드 HDD(sc1) 볼륨

    이 경우 성능은 처리량(MiB/s)으로 측정됩니다.

### API

#### RunInstances

```json
{
  "MaxCount": 1,
  "MinCount": 1,
  "ImageId": "ami-05d2438ca66594916",
  "InstanceType": "t2.medium",
  "KeyName": "ec2_aimpugn",
  "EbsOptimized": false,
  "BlockDeviceMappings": [
    {
      "DeviceName": "/dev/sda1",
      "Ebs": {
        "Encrypted": false,
        "DeleteOnTermination": true,
        "Iops": 3000,
        "SnapshotId": "snap-06c992bdc2741bc46",
        "VolumeSize": 8,
        "VolumeType": "gp3",
        "Throughput": 125
      }
    }
  ],
  "NetworkInterfaces": [
    {
      "AssociatePublicIpAddress": true,
      "DeviceIndex": 0,
      "Groups": [
        "<groupId of the new security group created below>"
      ]
    }
  ],
  "CreditSpecification": {
    "CpuCredits": "standard"
  },
  "TagSpecifications": [
    {
      "ResourceType": "instance",
      "Tags": [
        {
          "Key": "Name",
          "Value": "aimpugn"
        }
      ]
    }
  ],
  "MetadataOptions": {
    "HttpEndpoint": "enabled",
    "HttpPutResponseHopLimit": 2,
    "HttpTokens": "required"
  },
  "PrivateDnsNameOptions": {
    "HostnameType": "ip-name",
    "EnableResourceNameDnsARecord": true,
    "EnableResourceNameDnsAAAARecord": false
  }
}
```

#### 새 보안 그룹 설정

- [CreateSecurityGroup](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateSecurityGroup.html#API_CreateSecurityGroup_Examples)

    ```json
    {
        "GroupName": "sgr-ec2-aimpugn",
        "Description": "2024-08-30 16:25:00 created Security Group",
        "VpcId": "vpc-fe789c95"
    }
    ```

- [AuthorizeSecurityGroupIngress](https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_AuthorizeSecurityGroupIngress.html)

    ```json
    {
        "GroupId": "<groupId of the security group created above>",
        "IpPermissions": [
            {
                "IpProtocol": "6",
                // 22 포트로 마구 요청하는 것을 막기 위해 임의 포트
                "FromPort": 22022,
                "ToPort": 22022,
                "IpRanges": [
                    {
                        "CidrIp": "내.IP.주.소/32",
                        "Description": "personal SSH"
                    }
                ]
            }
        ]
    }
    ```

### 고급

- 도메인 조인 디렉터리

    도메인 조인을 사용하면 인스턴스를 AWS Directory Service에서 정의한 디렉터리에 조인할 수 있습니다.
    이를 통해 Windows 및 Linux 인스턴스 네트워크에서 단일 로그인 및 중앙 집중식 관리 환경을 제공할 수 있습니다.
    도메인을 조인하려면 필요한 권한을 포함한 IAM 역할이 있어야 합니다.
    활성 디렉터리 또는 손상된 디렉터리만 목록에 표시됩니다.

    조인할 디렉터리를 선택하면 인스턴스 시작 시 Amazon EC2에서 자동으로 디렉터리 속성을 지정하는 SSM 문서를 생성합니다.

- IAM 인스턴스 프로파일

    인스턴스의 IAM 인스턴스 프로파일 값을 지정하지 않으면 원본 템플릿의 값이 계속 사용됩니다.
    템플릿 값을 지정하지 않으면 기본 API 값이 사용됩니다.

## 로그인

> NOTE: 키 페어를 생성해서 해당 키 페어로만 접속할 수 있도록 합니다.

키페어를 다운받았으면, `~/.ssh` 디렉토리로 옮기고, 연결하기 편하게 `~/.ssh/config`를 설정합니다.

### Ubuntu AMI 설치한 경우

```config
Host ec2
    User ubuntu
    HostName ec2-w-x-y-z.<ap-region>.compute.amazonaws.com
    Port <Port>
    IdentityFile ~/.ssh/<NAME>.pem
    LogLevel INFO
```

### 로그인 후 root 유저 되기

임시 root 셸을 사용합니다.

```bash
# 현재 로그인한 사용자(ubuntu)가 sudoer인지 판단합니다.
# sodoer라면 사용자(root)의 비밀번호 데이터베이스 항목에서 로그인 셸로 지정한 셸을 실행합니다.
sudo -i

# 또는
sudo su -
```

또는 AWS Systems Manager Session Manager 사용합니다.
- EC2 인스턴스에 대한 보안 원격 관리를 제공합니다.
- root 권한으로 명령을 실행할 수 있습니다.

참고로 `su`로 root 계정으로 전환할 수 없습니다.
기본적으로 EC2 인스턴스에서는 root 계정의 직접 로그인이 비활성화되어 있습니다.

## EC2 root 접근 관련 AWS 문서

1. [EC2 User Guide - 루트 사용자 접근](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/managing-users.html#root-user-access)
2. [AWS 보안 모범 사례](https://docs.aws.amazon.com/wellarchitected/latest/security-pillar/sec_securely_operate_workload.html)
3. [EC2 인스턴스에 대한 Systems Manager 세션 관리자 설정](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-getting-started.html)

## RDS 데이터베이스 연결

- [Amazon EC2 인스턴스를 Amazon RDS DB 인스턴스에 연결](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/tutorial-connect-ec2-instance-to-rds-database.html)

RDS 데이터베이스를 EC2 인스턴스에 빠르게 연결하여 인스턴스 간 트래픽을 허용할 수 있습니다.

트래픽은 보안 그룹에 의해 허용되며, 보안 그룹은 다음과 같이 자동으로 생성되어 인스턴스 및 데이터베이스에 추가됩니다.

- Amazon EC2

    `ec2-rds-x`(ex: `ec2-rds-1`)라는 보안 그룹을 생성하여 EC2 인스턴스에 추가합니다.

    여기에는 `rds-ec2-x`(데이터베이스 보안 그룹, ex: `rds-ec2-1`)를 대상으로 지정하여 데이터베이스로의 트래픽을 허용하는 *아웃바운드* 규칙이 하나 있습니다.

- Amazon RDS

    `rds-ec2-x`(ex: `rds-ec2-1`)라는 보안 그룹을 생성하여 데이터베이스에 추가합니다.

    여기에는 `ec2-rds-x`(EC2 인스턴스 보안 그룹, ex: `ec2-rds-1`)를 소스로 지정하여 EC2 인스턴스에서의 트래픽을 허용하는 *인바운드* 규칙이 하나 있습니다.

보안 그룹은 *서로를 대상 및 소스로 참조하고 데이터베이스 포트에서의 트래픽만 허용*합니다.
`rds-ec2-x` 보안 그룹이 있는 모든 데이터베이스가 `ec2-rds-x` 보안 그룹과 함께 모든 EC2 인스턴스와 통신할 수 있도록 이러한 보안 그룹을 재사용할 수 있습니다.

`x`는 새 보안 그룹이 자동으로 생성될 때마다 1씩 증가하는 숫자입니다.

## 배포(Deployment) 방법

### 직접 배포

- 직접 설치 및 구성
    - 방법: SSH를 통해 EC2에 접속하여 필요한 소프트웨어를 직접 설치하고 구성합니다.
    - 장점: 세부적인 제어가 가능하며, 시스템에 대한 이해도를 높일 수 있습니다.
    - 단점: 시간이 많이 소요되며, 오류 가능성이 높습니다.

- 사용자 데이터 스크립트 사용
    - 방법: EC2 인스턴스 시작 시 사용자 데이터 스크립트를 통해 자동으로 소프트웨어를 설치하고 구성합니다.
    - 장점: 인스턴스 생성 시 자동화된 설정이 가능합니다.
    - 단점: 복잡한 설정의 경우 스크립트 작성이 어려울 수 있습니다.

- Docker 컨테이너 사용
    - 방법: Docker를 설치하고 애플리케이션을 컨테이너화하여 실행합니다.
    - 장점: 환경 일관성, 쉬운 배포 및 스케일링이 가능합니다.
    - 단점: Docker에 대한 학습이 필요할 수 있습니다.

- CI/CD 파이프라인 구축
    - 방법: AWS CodePipeline, CodeBuild, CodeDeploy 등을 사용하여 자동화된 배포 파이프라인을 구축합니다.
    - 장점: 지속적인 통합 및 배포가 가능하며, 개발 프로세스를 개선할 수 있습니다.
    - 단점: 초기 설정이 복잡할 수 있습니다.

- Configuration Management 도구 사용
    - 방법: Ansible, Chef, Puppet 등의 도구를 사용하여 인프라를 코드로 관리합니다.
    - 장점: 대규모 인프라 관리와 일관된 구성이 가능합니다.
    - 단점: 도구 학습에 시간이 필요할 수 있습니다.

### AWS에 특화된 방식

- [AWS Elastic Beanstalk](https://docs.aws.amazon.com/ko_kr/elasticbeanstalk/latest/dg/Welcome.html) 사용
    - 방법: AWS Elastic Beanstalk 서비스를 통해 애플리케이션을 배포합니다.
    - 장점: 인프라 관리의 복잡성을 줄이고 쉽게 배포할 수 있습니다.
    - 단점: 세부적인 제어가 제한될 수 있습니다.

- [AWS CodeDeploy](https://docs.aws.amazon.com/codedeploy/latest/userguide/welcome.html)
    - AWS의 완전 관리형 배포 서비스입니다.
    - EC2 인스턴스, 온프레미스 서버, Lambda 함수 등에 애플리케이션을 자동으로 배포합니다.
    - 자동화된 롤백, 점진적 배포 등 다양한 배포 옵션을 제공합니다.
    - GitHub, Bitbucket 등과 통합이 가능합니다.

    구현 단계:
    1. EC2 인스턴스에 CodeDeploy 에이전트 설치
    2. 애플리케이션 및 배포 그룹 생성
    3. appspec.yml 파일 작성
    4. 소스 코드 저장소와 연동

- AWS OpsWorks
    - Chef와 Puppet을 사용하여 애플리케이션을 구성하고 배포하는 구성 관리 서비스입니다.
    - 장점: 인프라 관리 자동화, 다양한 환경 지원

- AWS CloudFormation
    - 인프라를 코드로 정의하고 프로비저닝하는 서비스입니다.
    - EC2 인스턴스 생성부터 애플리케이션 배포까지 전체 스택을 관리할 수 있습니다.
    - 장점: 인프라의 버전 관리 가능, 재사용성 높음

- AWS Systems Manager
    - EC2 인스턴스와 온프레미스 서버를 관리하는 서비스입니다.
    - 패치 관리, 구성 관리, 자동화 등의 기능을 제공합니다.
    - 장점: 중앙 집중식 관리, 보안 강화

- AWS App Runner

    > AWS App Runner is not available in 아시아 태평양 (서울). Please select another region.

    - 컨테이너화된 웹 애플리케이션을 쉽게 배포하고 실행할 수 있는 완전 관리형 서비스입니다.
    - 장점: 인프라 관리 필요 없음, 자동 스케일링
    - [ESC + Fargate vs App Runner](https://cloudonaut.io/fargate-vs-apprunner/)

- [AWS Amplify](https://ap-northeast-2.console.aws.amazon.com/amplify/apps)

    > Amplify는 AWS CDK를 기반으로 구축되므로 모든 AWS 서비스를 추가, 확장 및 사용자 지정할 수 있습니다.

    - 웹 및 모바일 애플리케이션의 개발과 배포를 위한 완전 관리형 서비스입니다.
    - 프론트엔드 및 백엔드 배포를 모두 지원합니다.
    - 장점: CI/CD 파이프라인 자동 구성, 다양한 프레임워크 지원
    - [작동 방식](https://docs.amplify.aws/react/how-amplify-works/concepts/) 참고

- Serverless 접근 방식 (AWS Lambda + API Gateway)
    - 방법: 서버리스 아키텍처를 사용하여 애플리케이션을 구축합니다.
    - 장점: 인프라 관리가 필요 없으며, 자동 스케일링이 가능합니다.
    - 단점: 특정 사용 사례에만 적합할 수 있으며, 콜드 스타트 문제가 있을 수 있습니다.

- [AWS CodeStar](https://docs.aws.amazon.com/ko_kr/codestar/latest/userguide/welcome.html)

    > 2024년 7월 31일부터 Amazon Web Services (AWS) 는 프로젝트 생성 및 보기에 AWS CodeStar 대한 지원을 중단합니다.
    > 2024년 7월 31일 이후에는 더 이상 AWS CodeStar 콘솔에 액세스하거나 새 프로젝트를 생성할 수 없습니다.
    > 하지만 소스 리포지토리 AWS CodeStar, 파이프라인, 빌드를 포함하여 에서 생성한 AWS 리소스는 이번 변경의 영향을 받지 않고 계속 작동합니다.
    > AWS CodeStar 연결 및 AWS CodeStar 알림은 이번 중단으로 인해 영향을 받지 않습니다.

    - [GitHub 소스 리포지토리를 사용하여 프로젝트 만들기](https://docs.aws.amazon.com/ko_kr/codestar/latest/userguide/console-tutorial.html)

### 이미 실행중인 EC2 인스턴스에 배포하기

- GitHub Actions + SSH 배포
    - GitHub에서 제공하는 CI/CD 서비스를 활용합니다.
    - SSH를 통해 EC2 인스턴스에 직접 배포합니다.
    - 배포 프로세스를 코드로 관리할 수 있습니다.

    구현 단계:
    1. GitHub Actions workflow 파일 생성
    2. EC2 SSH 키를 GitHub Secrets에 저장
    3. Workflow에서 SSH를 통한 배포 스크립트 실행

    ```yaml
    name: Deploy to EC2
    on:
        push:
        branches: [ main ]
    jobs:
        deploy:
        runs-on: ubuntu-latest
        steps:
        - uses: actions/checkout@v2
        - name: Deploy to EC2
            env:
            PRIVATE_KEY: ${{ secrets.EC2_SSH_KEY }}
            HOST: ${{ secrets.EC2_HOST }}
            run: |
            echo "$PRIVATE_KEY" > private_key && chmod 600 private_key
            ssh -o StrictHostKeyChecking=no -i private_key ubuntu@${HOST} '
                cd /path/to/your/app &&
                git pull origin main &&
                npm install &&
                pm2 restart your-app
            '
    ```

- Jenkins + EC2 플러그인
    - Jenkins를 사용하여 CI/CD 파이프라인을 구축합니다.
    - EC2 플러그인을 통해 EC2 인스턴스와 쉽게 연동할 수 있습니다.
    - 다양한 플러그인과 확장성을 제공합니다.

    구현 단계:
    1. Jenkins 서버 설정 (EC2 또는 로컬)
    2. EC2 플러그인 설치 및 구성
    3. Jenkins 파이프라인 작성

- Ansible + Crontab
    - Ansible을 사용하여 배포 프로세스를 자동화합니다.
    - Crontab을 통해 주기적으로 Ansible 플레이북을 실행합니다.
    - 인프라 관리와 애플리케이션 배포를 통합할 수 있습니다.

    구현 단계:
    1. 로컬 머신이나 별도의 서버에 Ansible 설치
    2. 배포를 위한 Ansible 플레이북 작성
    3. Crontab에 Ansible 플레이북 실행 명령 추가

- Docker + Watchtower
    - 애플리케이션을 Docker 컨테이너로 패키징합니다.
    - Watchtower를 사용하여 자동으로 컨테이너 업데이트를 관리합니다.
    - 컨테이너 기반 배포로 환경 일관성을 유지할 수 있습니다.

    구현 단계:
    1. EC2 인스턴스에 Docker 설치
    2. 애플리케이션 Dockerize
    3. Docker 이미지를 저장소(예: Docker Hub)에 푸시
    4. Watchtower 설정 및 실행

## AWS 공식 가이드 문서

1. [EC2 사용 설명서](https://docs.aws.amazon.com/ec2/index.html)
2. [EC2에서 애플리케이션 배포하기](https://docs.aws.amazon.com/toolkit-for-visual-studio/latest/user-guide/deployment-ecs-aspnet-core.html)
3. [Elastic Beanstalk 개발자 안내서](https://docs.aws.amazon.com/ko_kr/elasticbeanstalk/latest/dg/concepts.html)
4. [AWS CodeDeploy 사용 설명서](https://docs.aws.amazon.com/codedeploy/latest/userguide/welcome.html)
5. [Docker 컨테이너 배포 가이드](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/docker-basics.html)
6. [AWS Lambda 개발자 안내서](https://docs.aws.amazon.com/lambda/latest/dg/welcome.html)

## 초보자를 위한 권장 사항

1. 시작하기: 직접 설치 및 구성 방법으로 시작하여 EC2의 기본 개념을 이해합니다.
2. 자동화 도입: 사용자 데이터 스크립트를 사용하여 기본적인 자동화를 경험합니다.
3. 컨테이너화 학습: Docker를 학습하고 애플리케이션을 컨테이너화해봅니다.
4. 관리형 서비스 활용: Elastic Beanstalk를 사용하여 더 쉽게 애플리케이션을 배포해봅니다.
5. CI/CD 도입: 간단한 CI/CD 파이프라인을 구축하여 자동화된 배포 프로세스를 경험합니다.

이러한 방법들을 단계적으로 학습하고 적용하면서, 프로젝트의 규모와 복잡성에 따라 적절한 방법을 선택할 수 있습니다. AWS의 공식 문서와 튜토리얼을 참고하면서 실습을 진행하는 것이 효과적인 학습 방법이 될 것입니다.

## 참고 자료

- [Amazon EC2 인스턴스 유형](https://aws.amazon.com/ec2/instance-types/)
- [AWS 성능 최적화 백서](https://docs.aws.amazon.com/whitepapers/latest/cost-optimization-leveraging-ec2-spot-instances/introduction.html)
- [Amazon EBS volume types](https://docs.aws.amazon.com/ebs/latest/userguide/ebs-volume-types.html?icmpid=docs_ec2_console)
