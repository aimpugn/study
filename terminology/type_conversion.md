# Type Conversion

- [Type Conversion](#type-conversion)
    - [타입 변환 함수](#타입-변환-함수)
    - [atoi (ASCII to Integer)](#atoi-ascii-to-integer)
    - [itoa (Integer to ASCII)](#itoa-integer-to-ascii)
    - [atof (ASCII to Float)](#atof-ascii-to-float)
    - [sprintf와 sscanf](#sprintf와-sscanf)

## 타입 변환 함수

이 함수들은 문자열과 다른 데이터 타입 간의 변환을 쉽게 해주는 유틸리티 함수로서, 입력 및 출력, 데이터 처리 등에서 매우 유용하게 사용되고 있다. 프로그래밍 언어나 라이브러리에 따라 이러한 함수들의 이름이나 정확한 구현 방식은 다를 수 있으며, 일부 언어는 이러한 기능을 메서드나 연산자 형태로 제공하기도 한다.

## atoi (ASCII to Integer)

- **의미**: ASCII 문자열을 정수로 변환합니다.
- **사용 예시**: `atoi("123")`은 문자열 `"123"`을 정수 `123`으로 변환합니다.
- **언어 지원**: C언어의 표준 라이브러리 `<stdlib.h>`에 정의되어 있으며, 다른 언어들도 비슷한 기능을 제공합니다.

## itoa (Integer to ASCII)

- **의미**: 정수를 ASCII 문자열로 변환합니다.
- **사용 예시**: `itoa(123)`은 정수 `123`을 문자열 `"123"`으로 변환합니다.
- **언어 지원**: C 표준에는 직접적으로 포함되어 있지 않지만, 많은 컴파일러와 라이브러리에서 비표준 확장으로 제공됩니다. 다른 언어들은 자신들만의 방식으로 이 기능을 제공합니다.

## atof (ASCII to Float)

- **의미**: ASCII 문자열을 부동소수점 숫자로 변환합니다.
- **사용 예시**: `atof("123.45")`은 문자열 `"123.45"`를 부동소수점 숫자 `123.45`로 변환합니다.

## sprintf와 sscanf

- **의미**: `sprintf`는 다양한 데이터 타입을 문자열로 포맷팅하는 함수이고, `sscanf`는 문자열에서 다양한 데이터 타입으로 데이터를 스캔하는 함수입니다.
- **사용 예시**: `sprintf(buffer, "%d", 123)`는 정수 `123`을 문자열 `"123"`으로 변환하여 `buffer`에 저장합니다. `sscanf("123", "%d", &number)`는 문자열 `"123"`에서 정수 `123`을 추출하여 `number` 변수에 저장합니다.
