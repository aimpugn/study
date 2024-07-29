
# Spreadsheet

- [Spreadsheet](#spreadsheet)
    - [표준화된 인터페이스가 존재하는지](#표준화된-인터페이스가-존재하는지)
        - [1. 함수 구현의 일관성](#1-함수-구현의-일관성)
        - [2. 내부 API 및 함수 엔진](#2-내부-api-및-함수-엔진)
        - [3. 호환성을 위한 표준 스펙](#3-호환성을-위한-표준-스펙)
        - [4. 프로그래밍 언어와 API](#4-프로그래밍-언어와-api)
        - [5. 사용자 정의 함수(UDF)](#5-사용자-정의-함수udf)
        - [6. 데이터 및 파일 포맷 호환성](#6-데이터-및-파일-포맷-호환성)
        - [결론](#결론)
    - [스프레드시트 함수의 표준 스펙](#스프레드시트-함수의-표준-스펙)
        - [1. **Microsoft Excel의 사실상의 표준화**](#1-microsoft-excel의-사실상의-표준화)
        - [2. **OpenFormula 및 ODF (Open Document Format)**](#2-openformula-및-odf-open-document-format)
        - [3. **ECMA-376 및 ISO/IEC 29500 (Office Open XML)**](#3-ecma-376-및-isoiec-29500-office-open-xml)
        - [4. **기타 스프레드시트 포맷 표준**](#4-기타-스프레드시트-포맷-표준)
        - [5. **웹 기반 리소스 및 구현 가이드**](#5-웹-기반-리소스-및-구현-가이드)
        - [결론](#결론-1)

## 표준화된 인터페이스가 존재하는지

구글 스프레드시트, MS 엑셀, 폴라리스 오피스 등 여러 스프레드시트 프로그램이 `VLOOKUP`, SUM 등의 동일한 함수 기능을 제공하는 이유는 대부분의 스프레드시트 소프트웨어가 따르는 **표준화된 인터페이스**나 **스프레드시트 함수 표준**이 존재하기 때문입니다.
이러한 함수들은 **사용자 경험의 일관성**을 제공하며, 다양한 플랫폼에서 사용자가 동일한 기능을 수행할 수 있도록 합니다.
하지만 이를 구현하기 위한 직접적인 "인터페이스"가 표준화되어 있는 것은 아니며, 개발자와 소프트웨어 회사들은 다음과 같은 방식으로 이러한 기능을 구현합니다.

### 1. 함수 구현의 일관성

스프레드시트 소프트웨어에서 자주 사용되는 함수들(`VLOOKUP`, SUM 등)은 기본적으로 같은 기능을 제공합니다. 이는 사용자들이 다른 스프레드시트 프로그램으로 이동할 때 적응할 수 있는 일관된 경험을 제공하기 위해서입니다. 이런 함수들은 **수학적, 논리적 기능**을 표준화하여 제공하며, 각 소프트웨어가 이를 구현합니다.

### 2. 내부 API 및 함수 엔진

스프레드시트 프로그램들은 일반적으로 자체적인 **함수 엔진**을 사용하여 수식 계산을 처리합니다. 이 엔진은 사용자가 입력한 수식을 해석하고, 각 함수가 어떤 작업을 수행해야 하는지 결정합니다. 예를 들어, `SUM` 함수는 특정 셀 범위의 합을 계산하는 기능을 제공합니다. 이때 내부적으로 이 함수는 다음과 같은 인터페이스(내부 구현)를 가질 수 있습니다:

- `SUM(range: Range) -> Number`: 범위(`range`) 내의 모든 숫자를 합산한 결과를 반환하는 함수.

각 스프레드시트 프로그램은 이러한 함수를 자체 엔진에서 구현하여 실행합니다.

### 3. 호환성을 위한 표준 스펙

Excel은 매우 오랜 시간 동안 표준으로 자리잡았으며, 많은 스프레드시트 소프트웨어들이 Excel의 함수 구현을 참고하거나 호환되도록 설계되었습니다. 예를 들어, Excel의 `VLOOKUP` 함수는 기본적으로 4개의 인수를 받으며, 동일한 인수 구조와 동작을 가지도록 다른 스프레드시트 소프트웨어에서도 유사한 기능이 구현됩니다.

```text
`VLOOKUP`(lookup_value, table_array, col_index_num, [range_lookup])
```

이러한 표준화된 인터페이스를 따름으로써 다양한 스프레드시트 프로그램 간의 호환성이 높아지며, 사용자들은 다양한 플랫폼에서 동일한 기능을 사용할 수 있습니다.

### 4. 프로그래밍 언어와 API

스프레드시트 프로그램을 구현하는 데 사용되는 프로그래밍 언어와 API도 이러한 함수의 구현에 큰 영향을 미칩니다. 대부분의 스프레드시트 소프트웨어는 자체적인 프로그래밍 언어(C++ 또는 JavaScript 등)를 사용하여 이러한 함수를 구현하며, 각 함수의 동작 방식은 표준 수학 및 프로그래밍 개념에 기반합니다. 이러한 언어적 기반은 함수의 구현과 사용이 일관되도록 보장합니다.

### 5. 사용자 정의 함수(UDF)

사용자 정의 함수(UDF)는 사용자가 특정 기능을 필요에 맞게 정의할 수 있도록 지원하는 기능입니다. 이 또한 다양한 스프레드시트 소프트웨어에서 제공되며, 함수의 구조와 동작 방식을 사용자가 직접 지정할 수 있도록 합니다. 이는 소프트웨어가 제공하는 기본 함수 이외에 추가적인 기능을 필요로 하는 경우에 유용합니다.

### 6. 데이터 및 파일 포맷 호환성

스프레드시트 소프트웨어들은 일반적으로 Excel 파일 형식인 `.xlsx`, `.xls`와 같은 포맷을 지원합니다. 이러한 파일 포맷은 함수와 데이터 구조를 일정하게 유지하게 하여, 다양한 소프트웨어에서 동일한 기능을 사용할 수 있게 합니다. 이 포맷은 파일 내에서 수식을 정의하고 계산할 수 있는 구조를 포함하고 있습니다.

### 결론

스프레드시트 프로그램에서 동일한 함수를 구현하기 위해 특정한 "표준 인터페이스"가 존재하는 것은 아니지만, **표준화된 함수 구조**, **내부 함수 엔진의 구현**, **파일 포맷의 호환성**, **프로그래밍 언어와 API** 등의 요소들이 이러한 일관된 기능 제공을 가능하게 합니다. 이로 인해 사용자는 다양한 소프트웨어에서 동일한 경험을 할 수 있으며, 이는 소프트웨어 간 호환성을 높이는 데 중요한 역할을 합니다.

## 스프레드시트 함수의 표준 스펙

스프레드시트 함수의 표준 스펙은 엄밀한 의미에서 하나의 공식 표준으로 정의된 것은 아닙니다.
그러나 다음과 같은 요인들에 의해 사실상 표준화된 함수들이 존재하게 되었습니다:

### 1. **Microsoft Excel의 사실상의 표준화**

- Microsoft Excel은 1985년 출시 이후 스프레드시트 소프트웨어 시장에서 압도적인 점유율을 차지해왔습니다. Excel의 함수와 기능들이 널리 사용되면서, 다른 스프레드시트 소프트웨어들도 자연스럽게 Excel의 함수 구조와 동작 방식을 따라가게 되었습니다.
- Excel의 기능이 사실상의 표준으로 자리잡으면서, `SUM`, `VLOOKUP`, `IF` 등과 같은 함수들이 여러 스프레드시트 프로그램에서 동일한 방식으로 구현되었습니다.

### 2. **OpenFormula 및 ODF (Open Document Format)**

- OASIS(Open Applications Standards for Information Society)에서 관리하는 ODF(Open Document Format) 표준 내에는 OpenFormula라는 수식 표준이 포함되어 있습니다. OpenFormula는 스프레드시트 문서에서 수식을 작성하고 계산하는 방법을 정의합니다.
- OpenFormula는 Microsoft Excel, OpenOffice Calc, LibreOffice Calc 등의 스프레드시트 프로그램에서 공통적으로 사용될 수 있는 함수들을 정의하며, 이로 인해 일정한 함수 구현이 가능해졌습니다.

- ODF 및 OpenFormula의 공식 문서:
    - [ODF 표준 문서 (OASIS)](https://docs.oasis-open.org/office/OpenDocument/v1.3/cs02/part4-formula/OpenDocument-v1.3-cs02-part4-formula.html)
    - [OpenFormula (Wikipedia)](https://en.wikipedia.org/wiki/OpenFormula)

### 3. **ECMA-376 및 ISO/IEC 29500 (Office Open XML)**

- Microsoft가 주도하여 만든 Office Open XML 형식은 ECMA-376 및 ISO/IEC 29500 표준으로 제정되었습니다. 이 표준은 Microsoft Office에서 사용되는 파일 형식을 정의하며, Excel의 수식 및 함수 구조도 포함됩니다.
- 이 표준은 Excel 문서의 호환성을 유지하기 위한 스펙을 포함하고 있으며, 이에 따라 다양한 스프레드시트 프로그램에서 Excel과 호환되는 함수들을 구현하게 됩니다.

- ECMA-376 및 ISO/IEC 29500에 대한 정보:
    - [ECMA-376 (Office Open XML)](https://www.ecma-international.org/publications-and-standards/standards/ecma-376/)
    - [ISO/IEC 29500](https://www.iso.org/standard/71691.html)

### 4. **기타 스프레드시트 포맷 표준**

- Open Document Format(ODF) 및 Office Open XML 이외에도, 다양한 스프레드시트 포맷이 존재하며, 그에 따른 함수 구현 방식이 일정 부분 표준화되어 있습니다.

### 5. **웹 기반 리소스 및 구현 가이드**

- Google Sheets 및 Microsoft Excel은 자체 문서화와 개발자 리소스를 통해 함수의 구현 방식과 사용 방법을 상세히 설명하고 있습니다. 이러한 리소스는 각 플랫폼에서 함수의 일관성을 유지하고 개발자가 이를 쉽게 사용할 수 있도록 돕습니다.

- 예를 들어:
    - [Google Sheets 함수 리스트](https://support.google.com/docs/table/25273?hl=en)
    - [Microsoft Excel 함수 가이드](https://support.microsoft.com/en-us/office/excel-functions-by-category-5f91f4e9-7b42-46d2-9bd1-63f26a86c0eb)

### 결론

스프레드시트 함수들의 표준 스펙은 ODF의 OpenFormula와 같은 공식적인 표준, Microsoft Excel의 사실상의 표준화, 그리고 ECMA 및 ISO 표준을 통해 정의되고 있습니다. 각각의 스펙 문서는 OASIS, ECMA, ISO와 같은 국제 표준화 기구의 웹사이트에서 접근할 수 있으며, 주요 스프레드시트 소프트웨어의 공식 문서화에서도 상세한 정보를 얻을 수 있습니다.