# PHPUnit 3.7에서 mocking 하기

- [PHPUnit 3.7에서 mocking 하기](#phpunit-37에서-mocking-하기)
    - [PHPUnit 3.7.38 Dockerfile](#phpunit-3738-dockerfile)
    - [gRPC 클라이언트 모킹하기](#grpc-클라이언트-모킹하기)
        - [1. gRPC 클라이언트 생성](#1-grpc-클라이언트-생성)
        - [2. PHPUnit 3.7.38 설치](#2-phpunit-3738-설치)
        - [3. gRPC 클라이언트 모킹](#3-grpc-클라이언트-모킹)
    - [기타](#기타)
    - [Closure 사용하여 mocking 된 메서드 호출간 데이터 전달](#closure-사용하여-mocking-된-메서드-호출간-데이터-전달)

## PHPUnit 3.7.38 Dockerfile

```Dockerfile
FROM php:5

RUN apt-get update -y

RUN apt-get install -y wget

RUN sh -c 'wget -c https://phar.phpunit.de/phpunit-3.7.38.phar -O phpunit.phar && chmod +x phpunit.phar && mv phpunit.phar /usr/local/bin/phpunit'
```

## gRPC 클라이언트 모킹하기

PHP 5.6, CakePHP 2.5, 그리고 PHPUnit 3.7.38 환경에서 gRPC 클라이언트를 모킹하는 방법을 설명하겠습니다. 이 환경에서는 최신 PHPUnit과 Mockery를 사용할 수 없으므로, PHPUnit 3.7.38의 기본 모킹 기능을 사용해야 합니다.

### 1. gRPC 클라이언트 생성

먼저, `protoc`를 사용하여 gRPC 클라이언트를 생성합니다.

예를 들어, `example.proto` 파일을 사용하여 클라이언트를 생성합니다.

```proto
syntax = "proto3";

package example;

service ExampleService {
  rpc SayHello (HelloRequest) returns (HelloReply);
}

message HelloRequest {
  string name = 1;
}

message HelloReply {
  string message = 1;
}
```

이 파일을 사용하여 PHP 클라이언트를 생성합니다.

```bash
protoc --php_out=./generated \
    --grpc_out=./generated \
    --plugin=protoc-gen-grpc=/path/to/grpc_php_plugin \
    example.proto
```

### 2. PHPUnit 3.7.38 설치

PHPUnit 3.7.38을 설치합니다. CakePHP 2.5는 이 버전을 지원합니다.

```bash
composer require --dev phpunit/phpunit:3.7.38
```

### 3. gRPC 클라이언트 모킹

테스트 파일 구조는 다음과 같습니다.

```bash
├─ app
│   │ Test
│   └─ Case
│       └─ ExampleServiceTest.php
└─ generated
   └─ ExampleServiceClient.php
   └─ ...
```

테스트 파일을 작성합니다.

```php
<?php

App::uses('ControllerTestCase', 'TestSuite');
App::uses('ExampleServiceClient', 'Vendor/Generated');

class ExampleServiceTest extends ControllerTestCase
{
    public function testSayHello()
    {
        // `getMock` 메서드를 사용하여 `ExampleServiceClient`를 모킹합니다.
        $mockClient = $this->getMock(
            'ExampleServiceClient',
            array('SayHello'), // `SayHello` 메서드를 모킹하여 원하는 값을 반환하도록 설정합니다.
        );

        // `HelloReply` 객체를 생성하고, 반환할 메시지를 설정합니다.
        $reply = new HelloReply();
        $reply->setMessage('Hello, World!');

        // `ExampleServiceClient`의 `SayHello` 메서드를 모킹하여
        // `HelloReply` 객체를 반환하도록 설정합니다.
        $mockClient->expects($this->once())
                   ->method('SayHello')
                   ->will($this->returnValue($reply));

        // `HelloRequest` 객체를 생성하고
        $request = new HelloRequest();
        $request->setName('World');

        // `SayHello` 메서드를 호출합니다.
        $response = $mockClient->SayHello($request);

        // 응답 메시지가 예상대로 반환되는지 확인합니다.
        $this->assertEquals('Hello, World!', $response->getMessage());
    }
}
```

## 기타

- [How to mock a CakePHP behavior for unit testing](https://stackoverflow.com/questions/19833495/how-to-mock-a-cakephp-behavior-for-unit-testing)

## Closure 사용하여 mocking 된 메서드 호출간 데이터 전달

```php
<?php

App::uses(ComponentCollection::class, 'Controller');
App::uses(DelegateComponentMockingUtil::class, 'Test/Util');
App::uses(BigComponent::class, 'Controller/Component');
App::uses(InicisDelegateComponent::class, 'Controller/Component');
App::uses(PromotionTestUtil::class, 'Test/Util');
App::uses(ThatModel::class, 'Model');
App::uses(ThisModel::class, 'Model');
App::uses(DboMock::class, 'Test/Case/Model');

class BigComponentTest extends CakeTestCase
{
    public const ANY_STRING = 'ANY_STRING';
    public const INT_0 = 0;
    public const INT_1000 = 1000;
    public const STATUS_A = 'a';
    public const STATUS_B = 'b';
    public const STATUS_C = 'c';
    public const VALID_MINUS_ID = 'minus-id-';

    /**
     * @var callable(): ResultFromBigFunction
     */
    private $closureToCapture;

    public function setUp()
    {
        parent::setUp();
        // Setup our component and fake test controller
        $this->BigComponent = new BigComponent(new ComponentCollection());

        $this->TheOtherModel = ClassRegistry::init('TheOtherModel');
    }

    public function tearDown()
    {
        parent::tearDown();
        // Clean up after we're done
        unset($this->BigComponent);
        unset($this->Controller);
        unset($this->closureToCapture);
    }

    /**
     * 프로모션 적용 없이 {@see BigComponent::someBigFunction()} 통해 결제 성공하는 경우입니다.
     *
     * @return void
     */
    public function testWhenSomeBigFunctionSuccessWithoutPromotion()
    {
        // 모의 InicisDelegateComponent 인스턴스를 가져옵니다.
        $mockInicisDelegate = DelegateComponentMockingUtil::aProvider(
            $this,
            $this->BigComponent->_Collection
        );

        // 모의 InicisDelegateComponent::sbcrPay() 함수 결과가 리턴할 값을 설정합니다.
        $expectedPaidAt = '2024-07-01 17:10:00';
        $aProviderWillReturn = DelegateComponentMockingUtil::aProviderWillReturn(
            true,
            [
                'paid_at' => $expectedPaidAt,
            ]
        );
        $mockInicisDelegate
            ->expects($this->once())
            ->method(DelegateComponentMockingUtil::METHOD_SBCR_PAY)
            ->will($this->returnValue($aProviderWillReturn));

        // 모의 InicisDelegateComponent::sbcrPay() 함수 결과를 기반으로
        // BigComponent::someBigFunction 함수의 절차를 모의합니다.
        $this->mockSomeBigFunctionFlow($aProviderWillReturn);

        $result = $this->BigComponent->someBigFunction(
            self::INT_1000,
            $this->getGateway([
                'pg_provider' => DelegateComponentMockingUtil::PG_PROVIDER_HTML5_INICIS,
            ]),
            $this->getCardInfo(),
            $this->getOrderRequest(),
            $this->getExtra()
        );

        $resultTheOtherModel = $result->data['some_data_key'];
        $actualTheOtherModel = $resultTheOtherModel[TheOtherModel::class];
        $actualTheOtherModelExtension = $resultTheOtherModel[TheOtherModelExtension::class];

        $this->assertTrue($result->success);
        $this->assertNull($result->error);
        $this->assertEquals(
            self::STATUS_B,
            $actualTheOtherModel['status'],
            'Failed to match some_table.status'
        );
        $this->assertEquals(
            $aProviderWillReturn->data['apply_num'],
            $actualTheOtherModel['apply_num'],
            'Failed to match some_table.apply_num'
        );
    }

    /**
     * {@see BigComponent::someBigFunction()} 메서드 내에서 호출되는 모델들의
     * 호출 과정을 정리하고 순서대로 모의합니다.
     * {@see $BigComponent}가 먼저 초기화가 되어 있어야 정상적으로 동작합니다.
     *
     * 모의된 각 메서드가 리턴하는 배열의 키는 기준이 되고, 각 인자로 전달되는 배열에
     * 해당하는 키가 존재할 경우에만 덮어씁니다.
     *
     * @param ImpResult $impResult {@see PGDelegator::sbcrPay() 결과를 나타냅니다.
     * @param bool $isPromotion 프로모션 적용 여부를 지정합니다.
     * @param int $securedAmount 확보됐다고 반환될 프로모션 비용을 덮어씁니다.(기본값: 100)
     * @param bool $paymentUpdateAll 성공/실패 결제 업데이트 성공 여부를 덮어씁니다.
     * @param bool $paymentExtensionUpdateAll
     *
     * @return void
     */
    private function mockSomeBigFunctionFlow(
        $impResult,
        $isPromotion = false,
        $securedAmount = 100,
        $paymentUpdateAll = true,
        $paymentExtensionUpdateAll = true
    ) {
        /* NOTE: BigComponent::someBigFunction 내에서 호출되는 함수들을 모두 미리 모의합니다. */
        /** @var PHPUnit_Framework_MockObject_MockObject $mockTheOtherModel */
        $mockTheOtherModel = $this->getMockForModel(TheOtherModel::class, [
            'find', 'findById', 'findByImpUid',
            'getDataSource', 'getLastInsertID',
            'saveAll', 'updateAll',
        ]);
        /** @var PHPUnit_Framework_MockObject_MockObject $mockTheOtherModelExtension */
        $mockTheOtherModelExtension = $this->getMockForModel(TheOtherModelExtension::class, [
            'updateAll',
        ]);

        /** @var PHPUnit_Framework_MockObject_MockObject $mockUser */
        $mockUser = $this->getMockForModel(User::class, [
            'findById',
        ]);

        /** @var PHPUnit_Framework_MockObject_MockObject $mockThatModelTheOtherModel */
        $mockThatModelTheOtherModel = $this->getMockForModel(ThatModelTheOtherModel::class, [
            'getFirstIssuedTheOtherModel',
        ]);

        /** @var PHPUnit_Framework_MockObject_MockObject $mockOtherModel */
        $mockOtherModel = $this->getMockForModel(OtherModel::class, [
            'saveMap',
        ]);

        /* 모의 메서드를 가진 모의 인스턴스를 설정합니다. */
        $this->BigComponent->TheOtherModel = $mockTheOtherModel;
        $this->BigComponent->TheOtherModelExtension = $mockTheOtherModelExtension;
        $this->BigComponent->OtherModel = $mockOtherModel;
        $this->BigComponent->ThatModelTheOtherModel = $mockThatModelTheOtherModel;
        $this->BigComponent->User = $mockUser;

        /*
         * BigComponent::someBigFunction 로직 모의를 시작합니다.
         * 크게 4개의 단계를 거칩니다
         * 1. 결제 데이터 ready
         * 2. sbcrPay 성공/실패 따른 후처리
         * 3. 성공/실패 결제 결과 저장
         * 4. 저장 내역 조회하여 ImpResult 타입으로 반환
         */

        /* 1. ready 결제 데이터 준비 과정 */
        // 1.1. ready: generateBillingUID 과정중 사용 가능한 imp_uid 여부를 판단합니다.
        $this->BigComponent->TheOtherModel
            ->expects($this->any())
            ->method('find')
            ->will($this->returnValue([]));

        // 1.2. ready: 서비스/테스트 환경에 따라 웹훅 발송할 주소를 판단할 때 설정을 조회합니다.
        $this->BigComponent->User
            ->expects($this->any())
            ->method('findById')
            ->will($this->returnValue([
                'User' => [
                    'id' => self::INT_1000,
                    'payment_notification_url' => self::ANY_STRING,
                    'payment_notification_sandbox_url' => self::ANY_STRING,
                ],
            ]));

        // 1.3. ready: 결제 수단을 저장하기 위해, 빌링키 발급 당시 저장했던 결제건을 조회합니다.
        $this->BigComponent->ThatModelTheOtherModel
            ->expects($this->any())
            ->method('getFirstIssuedTheOtherModel')
            ->will($this->returnValue([
                'pay_method' => self::ANY_STRING,
            ]));

        // 1.4. ready: 프로모션 적용 조건에 해당할 경우 해당 프로모션 비용을 반환합니다.
        if ($isPromotion) {
            $this->assertTrue(is_numeric($securedAmount));

            if ($impResult->success) {
                // 추후 결제가 성공하면 $securedAmount 된 값이 있다면 확정짓습니다.
                $commitOrReleaseMethodName = PromotionTestUtil::METHOD_NAME_COMMIT;
                $promotionCommitOrReleaseResponse = PromotionTestUtil::getCommitResponseWithData();
            } else {
                // 추후 결제가 실패하면 프로모션 비용을 풀어줍니다.
                $commitOrReleaseMethodName = PromotionTestUtil::METHOD_NAME_RELEASE;
                $promotionCommitOrReleaseResponse = PromotionTestUtil::getReleaseResponseWithData();
            }

            // 프로모션
            PromotionTestUtil::mockPromotionServiceClient($this, [
                [
                    'methodName' => PromotionTestUtil::METHOD_NAME_TRY_SECURE,
                    'unaryCallWillReturn' => [
                        PromotionTestUtil::getTrySecureResponseWithData($securedAmount),
                        PromotionTestUtil::getStatusObject(),
                    ],
                ],
                [
                    'methodName' => $commitOrReleaseMethodName,
                    'unaryCallWillReturn' => [
                        $promotionCommitOrReleaseResponse,
                        PromotionTestUtil::getStatusObject(),
                    ],
                ],
            ]);
        }

        // 1.5. ready: 결제 준비 데이터를 저장합니다.
        // ready 데이터중 검증하고 싶은 값들을 use 로 넘겨야 콜백 안에서 사용이 가능합니다.
        $saveDataValidations = [
            'isPromotion' => $isPromotion,
            'securedAmount' => $securedAmount,
        ];
        $this->BigComponent->TheOtherModel
            ->expects($this->once())
            ->method('saveAll')
            ->with(
                $this->callback(function ($saveData) use ($saveDataValidations) {
                    /** @var ResultFromBigFunction $saveData */
                    if ($saveDataValidations['isPromotion'] === true) {
                        if ($saveData['TheOtherModelExtension']['int_col_1'] !== $saveDataValidations['securedAmount']) {
                            return false;
                        }
                        if (empty($saveData['TheOtherModelExtension']['str_col_4'])) {
                            return false;
                        }
                    }
                    $saveData['id'] = self::INT_1000;

                    // `saveAll`시 저장되는 결제 데이터를 이후 로직에서 재사용하기 위해
                    // 현재 결제 데이터를 캡처하도록 클로저를 생성하여 클래스 속성으로 저장합니다.
                    $this->closureToCapture = $this->captureData($saveData);
                    return true;
                })
            )
            ->will($this->returnValue(true));

        if ($isPromotion) {
            // 1.6. 프로모션에 해당할 경우 commit 또는 release 위해 조회하게 됩니다.
            $this->BigComponent->TheOtherModel
                ->expects($this->once())
                ->method('findByImpUid')
                ->will($this->returnValue([
                    'TheOtherModel' => [
                        'promotion_id' => self::VALID_MINUS_ID,
                        'promotion_amount' => $securedAmount,
                    ],
                ]));
        }

        /* 2. 각 DelegateComponent::sbcrPay() 실행됩니다. */
        // 2.1. ready 로 저장했던 some_table 테이블의 id 값을 리턴합니다.
        $this->BigComponent->TheOtherModel
            ->expects($this->once())
            ->method('getLastInsertID')
            ->will($this->returnCallback(function () {
                $this->assertTrue(is_callable($this->closureToCapture));
                $paymentData = call_user_func($this->closureToCapture);

                return $paymentData['id'];
            }));

        // 2.2. 결제 건을 조회합니다.
        // 4. 최종적으로 모든 로직을 거치고 캡처된 데이터를 반환합니다.
        $this->BigComponent->TheOtherModel
            ->expects($this->atLeastOnce())
            ->method('findById')
            ->will($this->returnCallback(function () {
                // returnValue 메서드 사용하여 즉시 평가하는 대신,
                // returnCallback 메서드를 사용하여 실제 테스트 코드가 실행될 때 lazy evaluation 합니다.
                return call_user_func($this->closureToCapture);
            }));

        // 3. 성공/실패 결제 결과를 저장합니다.
        if ($paymentUpdateAll) {
            // 3.1. 결제 결과를 some_table 테이블에 업데이트합니다.
            $this->BigComponent->TheOtherModel
                ->expects($this->once())
                ->method('updateAll')
                ->with(
                    $this->callback(function ($paymentToSave) use ($impResult) {
                        $this->assertTrue(is_callable($this->closureToCapture));

                        if (!$impResult->success) {
                            // 3.1.1. sbcrPay 실패시
                            //        DB에 안전하게 저장하도록 값들을 따옴표로 감싸고 이스케이프하기 위해
                            //        데이터 소스를 가져오는 것을 mocking 합니다.
                            $this->BigComponent->TheOtherModel
                                ->expects($this->once())
                                ->method('getDataSource')
                                ->will($this->returnValue($this->getMock(
                                    /**
                                     * {@see Mysql} 클래스로 모의 인스턴스 만드는 대신,
                                     * 테스트시 사용하는 {@see DboMock} 클래스로 모의 인스턴스를 생성합니다.
                                     *
                                     * {@see DboSource::__construct()} 생성자의 두번째 인자로 false 값을
                                     * 전달함으로써 {@see Mysql}로 실제 DB 연결 없이 테스트용으로 모의 인스턴스를
                                     * 생성할 수 있습니다.
                                     *
                                     * 하지만 로직상 {@see DboSource::value()} 메서드 사용하여
                                     * DB에 안전하게 저장하기 위해서 따옴표로 감싸거나 이스케이프 처리만 수행하므로,
                                     * 테스트용 {@see DboMock} 클래스로 모의 인스턴스를 생성합니다.
                                     */
                                    DboMock::class,
                                    ['value']
                                )));
                        }

                        // some_table 테이블에 저장할 데이터를 합쳐서 다시 캡처합니다.
                        $capturedTheOtherModelData = call_user_func($this->closureToCapture);

                        // DB 저장 위해 quote 처리되어 있는 것을 제거합니다.
                        $capturedTheOtherModelData[TheOtherModel::class] = array_merge(
                            $capturedTheOtherModelData[TheOtherModel::class],
                            $this->removeQuoted($paymentToSave, TheOtherModel::class)
                        );

                        /* 저장할 데이터로 업데이트 된 결제 데이터를 다시 캡처합니다. */
                        $this->closureToCapture = $this->captureData($capturedTheOtherModelData);
                        return true;
                    })
                )
                ->will($this->returnValue($paymentUpdateAll));

            // 3.2. DelegateComponent 에서 `some_field`을 반환했다면 한 번 호출되어야 합니다.
            if (!empty($impResult->data['some_field'])) {
                $this->BigComponent->OtherModel
                    ->expects($this->once())
                    ->method('saveMap')
                    ->will($this->returnValue(true));
            }

            // 3.3. 결제 결과를 payment_extensions 테이블에 업데이트합니다.
            $this->BigComponent->TheOtherModelExtension
                ->expects($this->once())
                ->method('updateAll')
                ->with($this->callback(function ($paymentExtensionToSave) {
                    $this->assertTrue(is_callable($this->closureToCapture));

                    // payment_extensions 테이블에 저장할 데이터를 합쳐서 다시 캡처합니다.
                    $capturedTheOtherModelData = call_user_func($this->closureToCapture);

                    // DB 저장 위해 quote 처리되어 있는 것을 제거합니다.
                    $capturedTheOtherModelData[TheOtherModelExtension::class] = array_merge(
                        $capturedTheOtherModelData[TheOtherModelExtension::class],
                        $this->removeQuoted($paymentExtensionToSave, TheOtherModelExtension::class)
                    );

                    /* 저장할 데이터로 업데이트 된 결제 데이터를 다시 캡처합니다. */
                    $this->closureToCapture = $this->captureData($capturedTheOtherModelData);

                    return true;
                }))
                ->will($this->returnValue($paymentExtensionUpdateAll));
        }
    }

    /**
     * @param array{
     *     pg_provider?: string,
     *     pg_id?: string,
     *     pg_secret?: string|null,
     *     cancel_password?: string|null,
     *     pg_ext_key?: string|null,
     *     pg_ext_priv?: string|null,
     *     sandbox?: bool,
     *     channel_id?: string,
     * } $sbcrCustomer gateway 설정에 사용되는 `ThatModel` 모델 값을 덮어씁니다.
     *
     * @return array{
     *     col1: string,
     *     col1: string,
     * }
     */
    private function getGateway($sbcrCustomer = [])
    {
        return ThatModel::getGatewayFrom([
            'ThatModel' => $this->replaceIfKeyExists(
                [
                    'col1' => self::ANY_STRING,
                    'col2' => self::ANY_STRING,
                ],
                $sbcrCustomer
            ),
        ]);
    }

    /**
     * @param array{
     *     other_key?: string,
     *     some_code?: string,
     *     some_issuer?: string,
     *     some_publisher?: string,
     *     card_name?: string,
     *     card_type?: string,
     *     card_num?: string,
     *     customer_id?: string,
     *     that_col6?: string,
     * } $sbcrCustomer
     * @param mixed $sbcrCustomerPhone
     *
     * @return array{
     *     other_key?: string,
     *     some_code?: string,
     *     some_issuer?: string,
     *     some_publisher?: string,
     *     card_name?: string,
     *     card_type?: string,
     *     card_num?: string,
     *     customer_id?: string,
     *     that_col6?: string,
     * }
     */
    private function getCardInfo(
        $sbcrCustomer = [],
        $sbcrCustomerPhone = []
    ) {
        return ThatModel::getBillRequest([
            'ThatModel' => $this->replaceIfKeyExists(
                [
                    'other_key' => self::ANY_STRING,
                    'some_code' => self::ANY_STRING,
                ],
                $sbcrCustomer
            ),
            'ThatModelPhone' => $this->replaceIfKeyExists(
                [
                    'id' => null,
                    'userkey' => null,
                    'other_key_date' => null,
                ],
                $sbcrCustomerPhone
            ),
        ]);
    }

    /**
     * @param array{
     *     that_col1: string,
     *     that_col2: string,
     *     that_col3: string,
     *     that_col4: string,
     *     that_col5: string,
     *     that_col6: string,
     * } $sbcrCustomer
     * @param array{
     *     amount: int,
     *     buyer_addr?: string|null,
     *     buyer_email?: string|null,
     *     buyer_name?: string|null,
     *     buyer_postcode?: string|null,
     *     buyer_tel?: string|null,
     *     this_col2: string,
     *     uid_from_client: string,
     *     name: string,
     *     this_at: string,
     *     free_amount: int|null,
     * } $sbcrSchedule
     * @param array{
     *     card_quota: int,
     * } $sbcrScheduleExtension
     *
     * @return array{
     *     amount: number,
     *     buyer_addr?: string|null,
     *     buyer_email?: string|null,
     *     buyer_name?: string|null,
     *     buyer_postcode?: string|null,
     *     buyer_tel?: string|null,
     *     card_quota: int,
     *     this_col2: string,
     *     that_col1?:string|null,
     *     that_col2?:string|null,
     *     that_col3?:string|null,
     *     that_col4?:string|null,
     *     that_col5?:string|null,
     *     that_col6: string,
     *     uid_from_client: string,
     *     name: string,
     *     promotion_amount: 0,
     *     free_amount: int,
     *     vat: null,
     * }
     */
    private function getOrderRequest(
        $sbcrCustomer = [],
        $sbcrSchedule = [],
        $sbcrScheduleExtension = []
    ) {
        return ThisModel::getOrderRequestForSbcrPay([
            'ThatModel' => $this->replaceIfKeyExists(
                [
                    'that_col1' => self::ANY_STRING,
                    'that_col2' => self::ANY_STRING,
                    'that_col3' => self::ANY_STRING,
                    'that_col4' => self::ANY_STRING,
                    'that_col5' => self::ANY_STRING,
                    'that_col6' => self::ANY_STRING,
                ],
                $sbcrCustomer
            ),
            'ThisModel' => $this->replaceIfKeyExists(
                [
                    'this_col1' => self::INT_1000,
                    'this_col2' => self::ANY_STRING,
                    'uid_from_client' => self::ANY_STRING,
                    'name' => self::ANY_STRING,
                    'this_at' => self::ANY_STRING,
                    'free_amount' => null,
                ],
                $sbcrSchedule
            ),
            'ThisModelExtension' => $this->replaceIfKeyExists(
                [
                    'card_quota' => self::INT_0,
                ],
                $sbcrScheduleExtension
            ),
        ]);
    }

    /**
     * @param array{
     *  user_id?: int,
     *  free_amount?: int|null,
     *  plus_amount?: int|null,
     * } $sbcrSchedule
     * @param array{
     *     custom_data?: string,
     *     col1?: string,
     *     col2?: string,
     *     col3?: bool,
     *     col4?: bool,
     *     col5?: string,
     * } $sbcrScheduleExtension
     * - `col5`는 JSON 형식이어야 합니다.
     *
     * @return array{
     *     amount_debug: true,
     *     custom_data: string,
     *     col3: bool,
     *     col1: string,
     *     payment_type: string,
     *     col2: string,
     *     free_amount_or_null: int|null,
     *     col4: bool,
     *     user_id: string,
     *     plus_amount: int|null,
     *     ...
     * }
     */
    private function getExtra(
        $sbcrSchedule = [],
        $sbcrScheduleExtension = []
    ) {
        return ThisModel::getExtraForSbcrPay([
            'ThisModel' => $this->replaceIfKeyExists(
                [
                    'user_id' => self::ANY_STRING,
                    'free_amount' => null,
                    'plus_amount' => null,
                ],
                $sbcrSchedule
            ),
            'ThisModelExtension' => $this->replaceIfKeyExists(
                [
                    'custom_data' => self::ANY_STRING,
                    'col1' => self::ANY_STRING,
                    'col2' => self::ANY_STRING,
                    'col3' => self::ANY_STRING,
                    'col4' => self::ANY_STRING,
                    'col5' => null, // json 으로 전달하면 json_decode 되어 결과에 추가됩니다.
                ],
                $sbcrScheduleExtension
            ),
        ]);
    }

    /**
     * `$arrToReplace` 배열에 기준이 되는 배열 `$mainArr`의 키가 존재하면
     * 해당 `$arrToReplace`의 값으로 `$mainArr`의 값을 덮어씁니다.
     *
     * {@see BigComponent}에서 각 결제 또는 취소 로직 단계별로 존재하는 키가 상이할 수 있습니다.
     * 기준이 되는 배열에 키가 존재하고 해당 키에 덮어쓸 값이 존재하는 경우에만 덮어쓰도록 합니다.
     *
     * @param array $mainArr 기준이 되고 덮어쓰여질 배열
     * @param array $arrToReplace 덮어쓸 값이 들어있는 배열
     * @return array 덮어쓰기가 완료된 기준 배열
     */
    private function replaceIfKeyExists($mainArr, $arrToReplace)
    {
        if (empty($arrToReplace)) {
            return $mainArr;
        }

        array_walk($mainArr, function (&$value, $key) use ($arrToReplace) {
            if (isset($arrToReplace[$key])) {
                $value = $arrToReplace[$key];
            }
        });

        return $mainArr;
    }

    /**
     * 다음과 같이 quote 처리 되어 있는 데이터에서 따옴표 제거하고 주어진 클래스명을 키에서 제거합니다.
     *
     *      ASIS: [TheOtherModel.some_code] = 'ANY_STRING'
     *      TOBE: [some_code] = ANY_STRING
     *
     * @param array $quotedData
     * @param string $className
     * @return array
     */
    private function removeQuoted($quotedData, $className = '')
    {
        $removed = [];
        array_walk($quotedData, function ($value, $key) use (&$removed, $className) {
            if (!empty($className)) {
                $key = str_replace($className . '.', '', $key);
            }
            if (is_string($value)) {
                $removed[$key] = str_replace("'", '', $value);
            } else {
                $removed[$key] = $value;
            }
        });

        return $removed;
    }

    private function captureData($data)
    {
        return function () use ($data) {
            return $data;
        };
    }
}
```
