# time

- [time](#time)
    - [brief guide to time for developers](#brief-guide-to-time-for-developers)
    - [Unix 타임스탬프](#unix-타임스탬프)
    - [Epoch](#epoch)
    - [ISO 8601와 RFC 3339](#iso-8601와-rfc-3339)
        - [ISO 8601](#iso-8601)
        - [RFC 3339](#rfc-3339)
        - [ISO 8601과 RFC 3339의 관계](#iso-8601과-rfc-3339의-관계)

## [brief guide to time for developers](https://www.willem.dev/projects/brief-guide-to-time-for-developers/)

![brief-guide-to-time](../resources/brief-guide-to-time.svg)

## Unix 타임스탬프

Unix 타임스탬프는 일반적으로 1970년 1월 1일 00:00:00 UTC부터 현재까지의 시간을 초 단위로 나타낸 것.
이를 Epoch 시간이라고도 한다. 초 단위의 Unix 타임스탬프는 대부분의 Unix 시스템과 Unix 기반 언어에서 널리 사용된다.

밀리초 단위의 Unix 타임스탬프는 초 단위 타임스탬프에 1000을 곱한 값으로, 더 정밀한 시간 표현이 필요할 때 사용된다.
예를 들어, 자바스크립트에서는 기본적으로 밀리초 단위의 타임스탬프를 사용한다.

Unix 타임스탬프는 초 단위와 밀리초 단위 외에도 마이크로초(10^-6), 나노초(10^-9) 등 더 정밀한 단위로 표현할 수 있다.
이러한 정밀한 단위는 특정 프로그래밍 언어나 시스템에서 성능 측정, 고정밀 타이밍 등을 위해 사용됩니다.

Unix 타임스탬프와 Epoch 시간은 사실상 같은 개념을 나타낸다.
Epoch 시간은 Unix 타임스탬프의 기준점인 '에포크(Epoch)'를 의미하며, 이는 1970년 1월 1일 00:00:00 UTC를 가리킨다.
따라서 Unix 타임스탬프는 Epoch 시간을 기준으로 한 시간의 흐름을 초 단위로 나타낸 것이다.

## Epoch

"Epoch"이라는 단어는 그리스어에서 유래한 것으로, "시대"나 "특정 시점"을 의미합니다. 컴퓨팅 분야에서 "Epoch"은 특정 시점을 기준으로 시간을 측정하는 시작점을 가리키는 용어로 사용됩니다.

Unix 시스템에서 Epoch는 1970년 1월 1일 00:00:00 UTC를 기준으로 합니다. 이 시점을 "Unix Epoch" 또는 "Epoch Time"이라고 하며, Unix 타임스탬프는 이 시점으로부터 경과한 시간을 초 단위로 나타냅니다. Unix 타임스탬프는 이러한 방식으로 시간을 측정하기 때문에 Epoch 시간이라고도 불립니다.

Unix Epoch의 유래는 Unix 운영 체제의 초기 개발과 관련이 있습니다. Unix 시스템의 개발자들은 시간을 표현하는 효율적인 방법을 찾아야 했고, 1970년 1월 1일을 기준점으로 선택했습니다. 이 날짜는 특별한 사건이나 기념일과 관련이 없지만, Unix 시스템이 개발되기 시작한 시기와 가까웠고, 시간대(Time Zone)의 영향을 받지 않는 UTC를 사용하여 전 세계적으로 일관된 시간 기준을 제공할 수 있었습니다.

Unix Epoch를 기준으로 하는 시간 측정 방식은 Unix와 Unix 기반 시스템에서 널리 채택되었으며, 이후 다양한 프로그래밍 언어와 시스템에서 표준적인 시간 표현 방법으로 사용되고 있습니다. Epoch 시간은 시간을 간단하고 일관된 방식으로 표현하고, 시간대 변환 없이 전 세계적으로 동일한 시간을 나타낼 수 있는 장점이 있습니다.

ISO 8601과 RFC 3339는 모두 날짜와 시간을 나타내는 국제 표준 형식을 제공하지만, 서로 다른 문맥과 목적을 가지고 있습니다. 이 둘 사이에는 밀접한 관련이 있으며, RFC 3339는 ISO 8601을 기반으로 하면서 인터넷 프로토콜과 어플리케이션에서의 사용에 특화된 몇 가지 구체적인 규칙을 추가한 것입니다.

## ISO 8601와 RFC 3339

RFC 3339는 ISO 8601의 규칙을 인터넷 환경에 맞게 조정하고 특화시킨 것으로, 두 표준은 상호 보완적인 관계에 있다.

### ISO 8601

ISO 8601은 국제 표준화 기구(International Organization for Standardization)에 의해 개발된 날짜와 시간의 표현에 대한 국제 표준입니다. 이 표준은 다양한 날짜와 시간 형식을 정의하며, 전 세계적으로 널리 사용됩니다. ISO 8601의 형식은 `YYYY-MM-DDTHH:MM:SS`와 같이 연, 월, 일을 나타내는 날짜 부분과 시, 분, 초를 나타내는 시간 부분으로 구성됩니다.

### RFC 3339

RFC 3339는 인터넷 표준 문서인 RFC(Request for Comments) 시리즈의 일부로, 날짜와 시간 형식에 대한 프로토콜을 정의합니다. RFC 3339는 ISO 8601의 서브셋으로, 인터넷 기반 어플리케이션과 프로토콜에서 날짜와 시간 데이터를 교환하기 위한 보다 엄격한 규칙을 제시합니다. 특히, RFC 3339는 UTC(협정 세계시)와 지역 시간대의 사용을 명확히 규정하며, ISO 8601보다 구체적인 시간대 표현과 초의 정밀도에 대한 규칙을 제공합니다.

### ISO 8601과 RFC 3339의 관계

RFC 3339는 ISO 8601을 기반으로 하여, 인터넷 통신과 어플리케이션에서의 특정 요구 사항을 충족시키기 위해 개발되었다.
많은 인터넷 기반 기술과 프로토콜은 날짜와 시간 데이터의 표현을 위해 RFC 3339 형식을 사용한다.
이는 ISO 8601 표준의 범위 내에 있으면서도, 특정한 사용 사례에 맞춰진 형식을 제공한다.
RFC 3339는 ISO 8601의 규칙을 따르면서도, 웹, 이메일, 기타 인터넷 프로토콜에서 날짜와 시간 데이터를 명확하고 일관되게 교환하기 위한 구체적인 가이드라인을 제공한다.

coreapi/payments/forensic/repository: Implement functions and add tests

- Add prepared statement
- Add mock forensic repository for testing
- Move `Forensic` type from `forensic` package to `model`
- Fix test codes affected by refactoring package of `Forensic` type
