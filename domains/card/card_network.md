# card network

- [card network](#card-network)
    - [3당사자 모델과 4당사자 모델](#3당사자-모델과-4당사자-모델)
        - [3당사자 모델(Three-Party Model)](#3당사자-모델three-party-model)
        - [4당사자 모델(Four-Party Model)](#4당사자-모델four-party-model)
    - [대한민국의 카드 결제 시스템](#대한민국의-카드-결제-시스템)
    - [참고 자료](#참고-자료)

## 3당사자 모델과 4당사자 모델

카드 3당사자 모델과 4당사자 모델은 카드 결제 시스템에서 참여자들의 역할과 구성에 따라 구분되는 개념입니다.

### 3당사자 모델(Three-Party Model)

[3당사자 모델](../resources/3당사자%20모델.jpeg)

3당사자 모델은 발급사(`Issuer`), 카드회원(`Cardholder`), 가맹점(`Merchant`)으로 구성됩니다.

- 발급사는 카드를 발급하고, 가맹점 모집 및 관리를 직접 수행합니다.
- 카드회원은 발급사로부터 카드를 발급받아 사용합니다.
- *가맹점은 발급사와 직접 계약*을 맺고 카드 결제를 받습니다.

대표적인 예로는 American Express, Discover 등이 있습니다.

### 4당사자 모델(Four-Party Model)

[3당사자 모델](../resources/4당사자%20모델.jpeg)

4당사자 모델은 발급사(`Issuer`), 카드회원(`Cardholder`), 가맹점(`Merchant`), 매입사(`Acquirer`)로 구성됩니다.

- 발급사는 카드를 발급하고 카드회원을 관리합니다.
- *매입사는 가맹점을 모집*하고 카드 거래를 중계/정산하는 역할을 합니다.
- 카드 브랜드사(Card Network)는 카드 결제 네트워크를 제공하고 수수료를 받습니다.

대표적인 예로는 Visa, Mastercard 등이 있습니다.

## 대한민국의 카드 결제 시스템

대한민국은 대부분 3당사자 모델을 따르고 있습니다.

- 카드사(예: 신한카드, 삼성카드 등)가 발급사와 매입사 역할을 동시에 수행합니다.
- 일부 카드사는 자체 가맹점 인프라가 부족하여 VAN사(Value-Added Network)를 통해 가맹점 관리와 카드 승인을 처리하기도 합니다.

하지만 최근에는 간편결제사(예: 토스, 네이버페이 등)가 선불전자지급수단을 발행하면서 4당사자 모델도 점차 확대되고 있는 추세입니다.

이처럼 카드 3당사자/4당사자 모델은 카드 결제 생태계에서 참여자들의 역할 분담을 나타내는 개념으로, 각국의 금융 인프라와 규제 환경에 따라 다양한 형태로 발전해왔습니다. 대한민국은 전통적으로 3당사자 모델 중심이었으나, 최근 핀테크 기업의 성장으로 4당사자 모델도 확산되는 과도기에 있다고 볼 수 있겠습니다.

## 참고 자료

- [한국에는 없지만 글로벌 시장에는 존재하는 사업/산업 1편: 카드 네트워크 시스템 (관련 주식 종목: Mastercard, Visa)](https://blog.naver.com/brjhlee/222071914218)
- [카드의 생태계(3당사자 vs 4당사자)](https://konairecruit.oopy.io/e55ba02c-a790-4a0a-a057-8371f3d93951)
- [카드산업, 새로운 패러다임을 준비하라](https://assets.kpmg.com/content/dam/kpmg/kr/pdf/2019/kr-issuemonitor-card-new-paradigm-20191104.pdf)