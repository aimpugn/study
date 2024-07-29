# dd

- [dd](#dd)
    - [dd?](#dd-1)
    - [원시 디스크 데이터 확인](#원시-디스크-데이터-확인)
    - [기타](#기타)

## dd?

> "data duplicator" or "convert and copy"

파일 변환 및 복사를 주 목적으로 하는 Unix 계열 운영 명령어입니다.
입력 파일이나 장치에서 읽고, 선택적 변환을 수행하고, 출력 파일이나 장치에 쓸 수 있습니다.

유닉스에서는 하드웨어용 장치 드라이버(예: 하드 디스크 드라이브)와 특수 장치 파일(예: `/dev/zero` 및 `/dev/random`)이 일반 파일처럼 파일 시스템에 나타납니다.

`dd`는 해당 기능이 각 드라이버에 구현되어 있는 경우, 하드웨어용 장치 드라이버나 특수 장치 파일을 읽거나 쓸 수 있습니다.

따라서 `dd`는 하드 드라이브의 부팅 섹터를 백업하거나 일정량의 임의 데이터를 얻는 등의 작업에 사용할 수 있습니다.

복사시 다음과 같은 변환을 수행할 수 있습니다:
- 바이트 순서 스왑
- ASCII 및 EBCDIC 테스트 인코디오가의 변환
또한 데이터를 복사할 때 바이트 순서 스왑, ASCII 및 EBCDIC(Extended Binary Coded Decimal Interchange Code) 텍스트 인코딩과의 변환 등 변환을 수행할 수 있습니다.

> EBCDIC?
>
> EBCDIC은 주로 IBM 메인프레임 운영 체제에서 사용되는 8비트 문자 인코딩입니다.
> 1960년대에 IBM에서 개발했으며 일부 레거시 시스템, 특히 은행 및 금융 부문에서 여전히 사용되고 있습니다.

## 원시 디스크 데이터 확인

```sh
sudo dd if=/dev/xvda bs=512 count=1 | hexdump -C
```

## 기타

- [Top 30 Linux commands Used In DevOps](https://medium.com/edureka/linux-commands-in-devops-73b5a2bcd007)
