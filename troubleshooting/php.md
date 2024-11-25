# php

- [php](#php)
    - [dyld\[18618\]: Library not loaded: /opt/homebrew/opt/libavif/lib/libavif.15.dylib](#dyld18618-library-not-loaded-opthomebrewoptlibavifliblibavif15dylib)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Error occurred during parsing: Fail to push limit](#error-occurred-during-parsing-fail-to-push-limit)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [`pq: canceling statement due to user request`](#pq-canceling-statement-due-to-user-request)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [expression is not allowed as field default value](#expression-is-not-allowed-as-field-default-value)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [Warning](#warning)
        - [문제](#문제-4)
        - [원인](#원인-4)
        - [해결](#해결-4)
    - [psalm: MixedArrayAccess: Cannot access array value on mixed variable $header](#psalm-mixedarrayaccess-cannot-access-array-value-on-mixed-variable-header)
        - [문제](#문제-5)
        - [원인](#원인-5)
        - [해결](#해결-5)
    - [CakePHP2.5의 네임스페이스와 intellij 프로젝트 설정](#cakephp25의-네임스페이스와-intellij-프로젝트-설정)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - [세그멘테이션 오류](#세그멘테이션-오류)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-7)

## dyld[18618]: Library not loaded: /opt/homebrew/opt/libavif/lib/libavif.15.dylib

### 문제

```log
❯ php-cs-fixer --help
dyld[18618]: Library not loaded: /opt/homebrew/opt/libavif/lib/libavif.15.dylib
  Referenced from: <87C1A268-34E4-396F-8BBC-D5591064E333> /opt/homebrew/Cellar/gd/2.3.3_5/lib/libgd.3.dylib
  Reason: tried: '/opt/homebrew/opt/libavif/lib/libavif.15.dylib' (no such file), '/System/Volumes/Preboot/Cryptexes/OS/opt/homebrew/opt/libavif/lib/libavif.15.dylib' (no such file), '/opt/homebrew/opt/libavif/lib/libavif.15.dylib' (no such file), '/usr/local/lib/libavif.15.dylib' (no such file), '/usr/lib/libavif.15.dylib' (no such file, not in dyld cache), '/opt/homebrew/Cellar/libavif/1.0.1/lib/libavif.15.dylib' (no such file), '/System/Volumes/Preboot/Cryptexes/OS/opt/homebrew/Cellar/libavif/1.0.1/lib/libavif.15.dylib' (no such file), '/opt/homebrew/Cellar/libavif/1.0.1/lib/libavif.15.dylib' (no such file), '/usr/local/lib/libavif.15.dylib' (no such file), '/usr/lib/libavif.15.dylib' (no such file, not in dyld cache)
[1]    18618 abort      php-cs-fixer --help
```

### 원인

### 해결

- [How to install PHP 8.3 on MacOS](https://yarnaudov.com/how-to-install-php83-on-macos.html)

```shell
brew install gd
```

## Error occurred during parsing: Fail to push limit

### 문제

```log
root@f177d0d69191:/var/www/api# ./app/Console/cake Webhook

Welcome to CakePHP v2.5.2 Console
---------------------------------------------------------------
App : app
Path: /var/www/api/app/
---------------------------------------------------------------
Error: Error occurred during parsing: Fail to push limit.
#0 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/CodedInputStream.php(363): Google\Protobuf\Internal\CodedInputStream->pushLimit(21979119)
#1 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/GPBWire.php(288): Google\Protobuf\Internal\CodedInputStream->incrementRecursionDepthAndPushLimit(21979119, 0, 0)
#2 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Message.php(431): Google\Protobuf\Internal\GPBWire::readMessage(Object(Google\Protobuf\Internal\CodedInputStream), Object(Google\Protobuf\Internal\FileDescriptorProto))
#3 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Message.php(511): Google\Protobuf\Internal\Message::parseFieldFromStreamNoTag(Object(Google\Protobuf\Internal\CodedInputStream), Object(Google\Protobuf\Internal\FieldDescriptor), Object(Google\Protobuf\Internal\FileDescriptorProto))
#4 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Message.php(807): Google\Protobuf\Internal\Message->parseFieldFromStream(10, Object(Google\Protobuf\Internal\CodedInputStream), Object(Google\Protobuf\Internal\FieldDescriptor))
#5 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Message.php(771): Google\Protobuf\Internal\Message->parseFromStream(Object(Google\Protobuf\Internal\CodedInputStream))
#6 /var/www/api/app/Vendor/google/protobuf/src/Google/Protobuf/Internal/DescriptorPool.php(61): Google\Protobuf\Internal\Message->mergeFromString('\n\xEF\xBF\xBD\n CoreWebho...')
#7 /var/www/api/app/Lib/Proto/GPBMetadata/CoreWebhook/Entity/Webhook.php(82): Google\Protobuf\Internal\DescriptorPool->internalAddGeneratedFile('\n\xEF\xBF\xBD\n CoreWebho...', true)
#8 /var/www/api/app/Lib/Proto/GPBMetadata/CoreWebhook/Service/Webhook.php(18): GPBMetadata\CoreWebhook\Entity\Webhook::initOnce()
#9 /var/www/api/app/Lib/Proto/Io/some-qwerty-org.io/Corewebhook/Grpc/SendBlockingWebhookRequest.php(41): GPBMetadata\CoreWebhook\Service\Webhook::initOnce()
#10 /var/www/api/app/Lib/CorewebhookService/SendBlockingWebhookRequest.php(21): Io\some-qwerty-org.io\Corewebhook\Grpc\SendBlockingWebhookRequest->__construct()
#11 /var/www/api/app/Lib/CorewebhookService/WebhookService.php(26): SendBlockingWebhookRequest::toCWValue(Array)
#12 /var/www/api/app/Console/Command/WebhookShell.php(28): WebhookService::SendBlockingWebhook(Array)
#13 /var/www/api/lib/Cake/Console/Shell.php(440): WebhookShell->main()
#14 /var/www/api/lib/Cake/Console/ShellDispatcher.php(207): Shell->runCommand(NULL, Array)
#15 /var/www/api/lib/Cake/Console/ShellDispatcher.php(66): ShellDispatcher->dispatch()
#16 /var/www/api/app/Console/cake.php(35): ShellDispatcher::run(Array)
#17 {main}
```

### 원인

- [생성된 protobuf 메타데이터 파일이 수정됐을 때 발생](https://stackoverflow.com/a/69382249)할 수 있다는 답변 있음

### 해결

다시 protobuf 생성하면 작동한다

## `pq: canceling statement due to user request`

### 문제

```php
public function main()
{
    try {
        $sent = WebhookService::SendBlockingWebhook([
            'v1PaymentNoticeRequestId' => 0,
            'v1UserId' => 424,
            'webhookRequest' => [
                'targetUrl' => 'http://sub_a.localhost:8001/temp',
                'contentType' => ContentType::JSON,
                'webhookPayload' => [
                    'impUid' => 'impUid',
                    'merchantUid' => 'merchantUid',
                    'paymentStatus' => 'paymentStatus',
                    'eventType' => '',
                ],
            ],
        ]);
        println($sent);
    } catch (Exception $exception) {
        println($exception->getMessage());
        println($exception->getCode());
    }
}
```

```log
pq: canceling statement due to user request
2
```

- message: pq: canceling statement due to user request
- code: 2

### 원인

- 요청을 받는 서버측에서 local database와 통신하는 것의 문제였다

### 해결

- corewebhook 코드를 수정

## expression is not allowed as field default value

### 문제

```php
class MyClass {
    private static $client = new Client();
                             ^^^^^^^^^^^^
                             expression is not allowed as field default valu
}
```

### 원인

- PHP 클래스의 속성(또는 멤버 변수)은 선언 시 간단한 값(정적 값)으로만 초기화할 수 있기 때문
- 클래스 속성은 ~~직접적인 표현식~~이나 ~~객체 생성과 같은 동적인 작업~~을 통해 초기화될 수 없다. 따라서, `private static $client = new Client();`와 같은 코드는 PHP에서 허용되지 않는다
- PHP에서 클래스 속성을 초기화할 때 다음과 같은 제약사항이 있다
    - 정적 값만 허용: 클래스 속성은 기본적으로 정적 값(예: 정수, 문자열, 배열 리터럴 등)으로만 초기화될 수 있다
    - 표현식 또는 함수 호출 불가: 객체 생성(`new ClassName()`)이나 함수 호출과 같은 표현식은 속성 선언 시 사용될 수 없다

### 해결

```php
class MyClass {
    private static $client;
    
    // 클래스 생성자에서 클라이언트 객체를 초기화
    public function __construct() {
        if (self::$client === null) {
            self::$client = new Client();
        }
    }
}
```

```php
class MyClass {
    private static $client;

    // 정적 메서드를 사용하여 클라이언트 객체를 초기화하고 반환
    public static function getClient() {
        if (self::$client === null) {
            self::$client = new Client();
        }
        return self::$client;
    }
}
```

## Warning

### 문제

CakePHP2.5 사용중 계속 경로가 도커 밖의 경로와 도커 안의 경로가 뒤섞여서 나옴

```log
╰─ php index.php
Array
(
    [Configure] => /Users/rody/IdeaProjects/some-B-serivce/lib/Cake/Core/Configure.php
    [Hash] => /Users/rody/IdeaProjects/some-B-serivce/lib/Cake/Utility/Hash.php
    [Cache] => /Users/rody/IdeaProjects/some-B-serivce/lib/Cake/Cache/Cache.php
    ... 생략 ...
    [ErrorHandler] => /var/www/api/lib/Cake/Error/ErrorHandler.php
    [autoload] => /var/www/api/app/Vendor/autoload.php
    [CakeLog] => /var/www/api/lib/Cake/Log/CakeLog.php
    ... 생략 ...
)
```

### 원인

`App::init()` 시에 파일 캐시로부터 경로를 불러오기 때문!

```php
public static function init() {
    self::$_map += (array)Cache::read('file_map', '_cake_core_');
    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^!!
    print_r(self::$_map);
    register_shutdown_function(array('App', 'shutdown'));
}
```

### 해결

- `app/tmp/cache/persistent/*` 파일 삭제

## psalm: MixedArrayAccess: Cannot access array value on mixed variable $header

### 문제

```log
psalm: MixedArrayAccess: Cannot access array value on mixed variable $header
```

```php
    /**
     * 외부 서비스 호출시 전달할 Trace header를 설정한다.
     *   - B3 Trace
     *   - Datadog Trace
     *
     * @param $header array{
     *  traceId: string,
     *  spanId: string,
     * }
     * @return void
     */
    public static function setOutgoingHeader($header)
    {
        self::$outgoingHeader = [
            'x-b3-traceid' => [$header['traceId']],
            'x-b3-spanid' => [bin2hex(openssl_random_pseudo_bytes(16))],
            'x-b3-parentspanid' => [$header['spanId']],
        ];
```

### 원인

`@param`에서 `$header`의 위치가 잘못되었었다

### 해결

```php
    /**
     * 외부 서비스 호출시 전달할 Trace header를 설정한다.
     *   - B3 Trace
     *   - Datadog Trace
     *
     * @param array{
     *  traceId: string,
     *  spanId: string,
     * } $header <------- 여기로 이동해야 추론 가능
     * @return void
     */
    public static function setOutgoingHeader($header)
    {
        self::$outgoingHeader = [
            'x-b3-traceid' => [$header['traceId']],
            'x-b3-spanid' => [bin2hex(openssl_random_pseudo_bytes(16))],
            'x-b3-parentspanid' => [$header['spanId']],
        ];
```

## CakePHP2.5의 네임스페이스와 intellij 프로젝트 설정

### 문제

1. CakePHP2.5를 사용중이고 /path/to/service/app/Lib/Some/Util/Converter.php 를 추가함.
2. project structure 설정
    1. /path/to/service 디렉토리를 source 디렉토리로 체크
    2. /path/to/service/app, /path/to/service/lib 두 디렉토리가 Package로 표시가 됨.
3. 그리고 Converter 의 클래스에 마우스를 올려 보면, 경고 발생

    > Namespace name doesn't match the PSR-0/PSR-4 project structure

    "Replace namespace '\' with 'app/Lib/Some/Util'"로 바꾸는 것을 추천함.

### 원인

CakePHP 2.5는 기본적으로 PSR-0 또는 PSR-4 네임스페이스 표준을 사용하지 않습니다.
대신 CakePHP는 자체적인 자동 로딩 방식과 파일 구조를 사용합니다.
따라서 IntelliJ와 같은 IDE에서 PSR-0/PSR-4 네임스페이스와 관련된 경고가 나타날 수 있습니다.
이 문제는 CakePHP의 파일 구조와 PSR-0/PSR-4 표준이 충돌하기 때문에 발생합니다.
CakePHP 2.5에서는 PSR-0/PSR-4 네임스페이스를 지원하지 않기 때문에, IDE가 파일 위치와 네임스페이스 간의 불일치를 감지합니다.

### 해결

CakePHP 2.5는 PSR-0/PSR-4 표준을 사용하지 않으며, 클래스 파일을 특정 디렉토리 구조에 배치하여 자동 로딩합니다.

예를 들어, `Lib` 디렉토리는 CakePHP에서 추가적인 라이브러리 파일을 저장하는 디렉토리입니다.

IntelliJ에서 CakePHP 프로젝트를 작업할 때는 PSR-0/PSR-4 네임스페이스 경고를 무시하거나 비활성화해야 합니다.
이를 통해 CakePHP의 파일 구조에 맞게 작업할 수 있습니다.

IntelliJ에서 특정 디렉토리를 네임스페이스로 인식하지 않도록 설정을 변경할 수 있습니다. 이는 PSR-0/PSR-4 네임스페이스 경고를 제거하는 데 도움이 됩니다.

1. **CakePHP 2.5 파일 구조 사용**:
    - `Lib` 디렉토리 내에 파일을 추가하고, 클래스 파일을 작성할 때 네임스페이스를 사용하지 않습니다.
    - 예를 들어, `Converter.php` 파일의 내용은 다음과 같아야 합니다:

    ```php
    <?php
    class Converter {
        // 클래스 내용
    }
    ```

2. **IntelliJ에서 네임스페이스 검사를 비활성화**:
    - IntelliJ 설정에서 네임스페이스 검사를 비활성화하여 PSR-0/PSR-4 경고를 무시할 수 있습니다.
    - `File` > `Settings` > `Editor` > `Inspections`로 이동합니다.
    - `PHP` > `PHP Inspections (EA Extended)` 항목을 찾습니다.
    - `Namespace` 관련 검사를 비활성화합니다.

3. **CakePHP 설정 유지**:
    - CakePHP 2.5 프로젝트는 PSR-0/PSR-4 표준을 따르지 않기 때문에, CakePHP의 파일 구조와 규칙을 따르는 것이 중요합니다. 클래스 파일을 작성할 때 CakePHP의 기존 규칙을 준수합니다.

## 세그멘테이션 오류

### 문제

```php
/**
 *
 * @return void
 */
public function testTrySecure($testCase)
{
    $mock = $this->getMock(
        'PromotionComponent',
        [],
        [new ComponentCollection(), []]
    );
    print_r($mock);
}
```

### 원인

mock 생성 코드를 따라가면, `newInstanceArgs` 호출시에 에러가 발생한다.

```php
protected static function getObject($code, $className, $originalClassName = '', $callOriginalConstructor = FALSE, $callAutoload = FALSE, array $arguments = array())
{
    if (!class_exists($className, FALSE)) {
        eval($code);
    }

    if ($callOriginalConstructor &&
        !interface_exists($originalClassName, $callAutoload)) {
        if (count($arguments) == 0) {
            $object = new $className;
        } else {
            $class = new ReflectionClass($className);
            $object = $class->newInstanceArgs($arguments);
        }
```

```php
const OPTION_CREDENTIALS = 'timeout';
const OPTION_TIMEOUT = 'credentials';
^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
이 부분이 문제였다

public function __construct(ComponentCollection $collection, $settings = [])
{
    parent::__construct($collection, $settings);

    if (
        !empty($settings[self::SETTINGS_CLIENT])
        && $settings[self::SETTINGS_CLIENT] instanceof PromotionServiceClient
    ) {
        $this->client = $settings[self::SETTINGS_CLIENT];
    } elseif (empty($this->client) || !$this->client instanceof PromotionServiceClient) {
        $this->client = new PromotionServiceClient(Envs::asString(
            Envs::CONF_HOST_TX_PROMOTION_SERVICE,
            'localhost:9090' // TODO prod 호스트명으로 수정하기
        ), [
            self::OPTION_CREDENTIALS => Grpc\ChannelCredentials::createSsl(),
            self::OPTION_TIMEOUT => 10 * 1000 * 1000, // 10s in microseconds
        ]);
        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        이 부분이 문제였다
    }
}
```

`credentials`에 `Grpc\ChannelCredentials::createSsl()`가 들어가야 하는데 `timeout` 키로 잘못 들어가고 있었다.

### 해결

변수 값을 수정...

```php
const OPTION_CREDENTIALS = 'credentials';
const OPTION_TIMEOUT = 'timeout';
```
