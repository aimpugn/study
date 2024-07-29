# Pivot

## Pivot이란?

- **용어 정의**: 'Pivot'은 원래 물리학, 기계학, 공학에서 사용되던 용어로, 어떤 물체가 회전하는 축이나 중심점을 의미합니다.
- **어원**: 프랑스어 'pivot'에서 유래되었으며, 이는 다시 라틴어 'pivus'에서 파생되었습니다.
- **데이터베이스와 프로그래밍에서의 의미**: 데이터를 재구조화하거나 변환하는 것을 의미하며, 특히 데이터의 행과 열의 위치를 바꾸거나, 복잡한 데이터 집합을 분석하기 쉬운 형태로 변환하는 것을 지칭합니다.

## pivot의 어원 및 의미의 확장

라틴어의 "pivus"에서 파생된 것으로, "pivus"는 회전하거나 중심을 의미하는 단어입니다.
이 라틴어 단어는 중세 라틴어에서 "힌지(hinge)" 또는 "핀(pin)"의 의미로 사용되기도 했으며, 이는 회전 축이나 지점을 나타내는 데 사용됩니다.

- **라틴어 "pivus"**

    이 단어는 '회전하는 축' 또는 '피벗 포인트'를 의미하며, 회전이나 중심을 나타내는 데 사용됩니다.
    이러한 의미는 물리학, 기계학에서 매우 중요한 개념입니다.

    "pivot"은 기계 부품이나 물체가 회전하는 축을 지칭하는 데 사용되어 왔습니다.
    이 개념은 다양한 기계의 디자인과 기능에서 핵심적인 요소로 사용됩니다.

- **프랑스어 "pivot"**

    라틴어에서 유래된 이 단어는 중세 프랑스어를 거쳐 현대 프랑스어로 이어졌으며, '회전축'이라는 의미로 꾸준히 사용되었습니다.

    또한, '핵심적인 사람이나 사물'을 의미하는 데에도 사용되며, 이는 어떤 시스템에서 중심적인 역할을 한다는 의미에서 비유적으로 사용됩니다.

- **비즈니스 및 전략적 용어로의 발전**

    시간이 지남에 따라 'pivot'은 변화나 전략적인 방향 전환을 의미하는 비즈니스 용어로도 발전했습니다.

    이는 기업이나 프로젝트가 새로운 방향이나 방법론으로 전환하는 것을 의미하며, 이는 원래의 '회전' 또는 '방향 전환'의 의미에서 파생된 것입니다.

- **데이터베이스에서의 Pivot**

    데이터베이스 관리 및 데이터 분석에서 'pivot'은 데이터를 재구조화하여 다양한 방식으로 정보를 요약하고 분석할 수 있게 하는 기술을 의미합니다.

    이는 특정 축을 중심으로 데이터를 회전시켜서 다양한 차원에서 데이터를 분석하고 이해하는 데 도움을 줍니다.

네, 맞습니다. Google 스프레드시트의 `QUERY` 함수에서 `PIVOT A`를 사용하는 경우, `A` 열의 고유 값들을 새로운 열 헤더로 변환하고, 이 열 헤더 아래에 각각의 값(여기서는 `SUM(C)`)을 배치하게 됩니다. 이러한 방식으로 데이터를 재구성하면, 데이터의 표현 방식이 변경되어 보다 명확하고 분석하기 쉬운 형태로 정보를 확인할 수 있습니다.

## Pivot의 역할 및 목적

1. **데이터 재구조화**: `PIVOT`는 기존의 데이터 구조를 변경하여 새로운 관점에서 정보를 볼 수 있게 합니다. 특히, 데이터베이스의 행과 열을 바꾸는 것은 많은 데이터 분석 시나리오에서 유용합니다.

2. **시각적 명확성 및 접근성 향상**: `A` 열이 날짜를 포함하고 있는 경우, 각 날짜에 대한 데이터를 열 형태로 나열하면 시간에 따른 데이터 변화를 한눈에 파악하기 쉽습니다. 이는 특히 시계열 데이터 분석에 유리합니다.

3. **분석 및 보고 용이성**: 날짜별로 데이터를 열 방향으로 확장함으로써, 특정 기간 동안의 데이터 변화나 추세를 쉽게 분석하고 보고할 수 있습니다. 각 열이 특정 날짜를 대표하므로, 날짜별로 집계된 데이터를 바로 비교하고 분석할 수 있습니다.

## 왜 Pivot을 사용하는가?

`PIVOT`을 사용하는 주된 이유는 복잡한 데이터 세트에서 특정 변수(여기서는 날짜)를 중심으로 데이터를 쉽게 비교하고 분석하기 위함입니다. 예를 들어, 여러 날짜에 걸친 유저별 활동 데이터가 있을 때, `PIVOT`을 사용하여 각 날짜를 열로 설정하고, 각 유저의 활동을 행으로 설정하면, 유저별 활동 패턴의 시간에 따른 변화를 효과적으로 분석할 수 있습니다. 이는 시간의 흐름에 따른 데이터의 변화를 시각적으로 더 잘 파악하고, 해당 데이터에 기반한 의사 결정을 보다 명확하게 할 수 있게 도와줍니다.

이렇게 `PIVOT`을 사용하는 것은 데이터를 보다 목적에 맞게 배열하여 효율적으로 정보를 추출하고, 결과적으로 데이터의 가치를 최대한 활용할 수 있도록 합니다.

## Pivot Table

### 개념 및 사용

- **Pivot Table**: 데이터 분석에 사용되는 도구로, 특히 스프레드시트 프로그램에서 자주 볼 수 있습니다.
- **기능**: Pivot Table은 복잡한 데이터를 요약, 정렬, 그룹화 및 집계하여 보다 쉽게 이해할 수 있도록 도와줍니다.
- **데이터 '회전'**: 데이터를 'pivot'하면 데이터를 여러 방면에서 분석할 수 있게 하여 다양한 통찰을 제공합니다.

### 유래와 역사

- **개발 배경**: 데이터 분석과 보고서 작성이 점점 복잡해지면서 pivot table의 개념이 발전했습니다.
- **최초 도입**: 1980년대에 Lotus Development Corporation이 스프레드시트 프로그램인 Lotus Improv에서 처음으로 Pivot Table을 소개했습니다.
- **표준화**: 이후 Microsoft Excel을 포함한 다수의 스프레드시트 프로그램에서 pivot table 기능을 통합하면서 데이터 분석 방법에 혁신을 가져왔습니다.

### 데이터베이스에서의 Pivot Table 사용

- **다중 테이블 연결**: 데이터베이스에서의 'Pivot Table'은 여러 테이블에 걸쳐 있는 데이터를 연결하고, 이를 하나의 테이블로 요약하는데 사용됩니다.
- **연결 및 분석**: Pivot table은 다른 데이터 소스를 중심축처럼 연결하며, 이를 통해 새로운 통찰이나 분석을 할 수 있습니다.
- **예시**: "빌링키 발급 정보"와 "트랜잭션 정보"를 연결하기 위한 pivot 테이블은 이 두 데이터 세트 사이의 관계를 명확하게 하며, 분석을 용이하게 합니다.

### Pivot Table의 역할

- **데이터 세트 연결**: Pivot table은 두 개 이상의 데이터 세트를 연결하는 중심 역할을 하며, 여러 데이터 소스간의 관계를 보다 명확하게 해줍니다.
- **분석 도구로서의 기능**: 다양한 데이터 세트를 쉽게 분석하고, 중요한 데이터 포인트를 빠르게 식별할 수 있도록 도와줍니다.