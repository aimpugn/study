# php w/ gRPC

- [php w/ gRPC](#php-w-grpc)
    - [debug](#debug)
    - [`protoc` 명령어로 프로토콜 버퍼 메시지 생성](#protoc-명령어로-프로토콜-버퍼-메시지-생성)
    - [클라이언트 생성](#클라이언트-생성)
        - [사용 가능한 옵션들](#사용-가능한-옵션들)
        - [예제](#예제)
    - [응답 처리](#응답-처리)
    - [Read](#read)

## debug

- [Tracing and Logging](https://github.com/grpc/grpc/blob/master/src/php/README.md#tracing-and-logging)
- [gRPC environment variables](https://github.com/grpc/grpc/blob/master/doc/environment_variables.md#grpc-environment-variables)

```ini
# php.ini
; grpc extension added by promotion grpc
extension=grpc.so

# grpc.grpc_verbosity=info
grpc.grpc_verbosity=debug
#grpc.grpc_trace=all,-timer_check
grpc.grpc_trace=all,-timer_check,-polling,-polling_api,-pollable_refcount,-timer,-timer_check
grpc.log_filename=/var/log/php/grpc.log

;grpc.grpc_verbosity=info
;grpc.grpc_trace=all,-timer_check,-polling,-polling_api,-pollable_refcount,-timer,-timer_check
;grpc.log_filename=/var/log/php/grpc.log
```

```shell
# reload
service php5.6-fpm reload
```

컨테이너의 fpm 리로드

```shell
docker exec -it php5.6 /bin/bash -c "service php5.6-fpm reload"
```

## `protoc` 명령어로 프로토콜 버퍼 메시지 생성

```shell
protoc --proto_path=./proto/v2/channel-service \
    --grpc_out=./gen_src/php \
    --plugin=protoc-gen-grpc=./supplements/php/grpc_php_plugin \
    $(find ./proto/v2/channel-service -iname "*.proto")
```

## 클라이언트 생성

PHP gRPC 클라이언트를 초기화할 때 타임아웃 옵션을 설정하려면 `ChannelCredentials` 생성 시 옵션 배열에 타임아웃 값을 추가하면 됩니다.
gRPC PHP 확장에서는 `'grpc.primary_user_agent'`, `'grpc.ssl_target_name_override'`와 같은 채널 옵션을 지원합니다.

가령 `new \Grpc\Channel('localhost:32400', [])` 코드를 사용하여 gRPC 채널을 생성할 때, `localhost`의 포트 `32400`에 연결을 시도합니다. 이 코드 조각은 gRPC 클라이언트가 특정 주소 (`localhost:32400`)에 있는 gRPC 서버와 통신할 수 있도록 채널을 설정하는 것입니다.

```php
// gRPC 채널 생성
// 1. `new \Grpc\Channel('localhost:32400', [])`를 통해 `localhost:32400` 주소로의 채널을 생성합니다.
// 2. 생성된 채널은 지정된 주소로 연결을 시도합니다. 이 과정에서 네트워크 연결이 설정되고, gRPC 서버와의 통신을 준비합니다.
// 3. 이후 생성된 채널을 gRPC 클라이언트 인스턴스에서 사용하여 서버와 통신할 수 있습니다.
$channel = new \Grpc\Channel('localhost:32400', []);

// gRPC 클라이언트 생성
$client = new MyServiceClient('localhost:32400', [
    // 전하지 않은 (비암호화) 채널을 생성하는 옵션
    'credentials' => \Grpc\ChannelCredentials::createInsecure(),
]);

// gRPC 호출
$request = new MyRequest();
list($response, $status) = $client->SomeFunc($request)->wait();

// 응답 처리
if ($status->code === \Grpc\STATUS_OK) {
    echo "Response received: ", $response->getMessage(), "\n";
} else {
    echo "Error: ", $status->details, "\n";
}

// 채널 종료
$channel->close();
```

### 사용 가능한 옵션들

gRPC 클라이언트 초기화 시 설정할 수 있는 다양한 옵션들은 클라이언트와 서버 간의 통신을 최적화하고 안정성을 보장하기 위해 존재합니다.
각 옵션의 역할과 단위를 구체적인 사례를 통해 설명하겠습니다.

1. **credentials**
   - **역할**: 클라이언트와 서버 간의 보안 연결을 설정합니다. `Grpc\ChannelCredentials::createSsl()`은 SSL/TLS를 사용하여 보안 연결을 설정합니다.
   - **단위**: N/A (객체)

2. **grpc.max_receive_message_length**
   - **역할**: 클라이언트가 수신할 수 있는 최대 메시지 크기를 설정합니다. 기본값은 -1로, 이는 무제한을 의미합니다.
   - **단위**: 바이트 (Bytes)
   - **사례**: 큰 파일을 전송하는 경우, 이 값을 적절히 설정하여 메모리 초과 오류를 방지할 수 있습니다.

3. **grpc.max_send_message_length**
   - **역할**: 클라이언트가 보낼 수 있는 최대 메시지 크기를 설정합니다. 기본값은 -1로, 이는 무제한을 의미합니다.
   - **단위**: 바이트 (Bytes)
   - **사례**: 큰 데이터를 서버로 전송할 때, 이 값을 설정하여 전송 가능한 최대 크기를 제한할 수 있습니다.

4. **grpc.keepalive_time_ms**
   - **역할**: 클라이언트가 서버에 keepalive ping을 보내는 간격을 설정합니다. 이 값은 연결이 유휴 상태일 때도 연결을 유지하기 위해 사용됩니다.
   - **단위**: 밀리초 (Milliseconds)
   - **사례**: 클라이언트와 서버 간의 연결이 장시간 유휴 상태일 때도 연결을 유지하고자 할 때 사용합니다. 예를 들어, 60초(60000밀리초)로 설정하면 60초마다 ping을 보냅니다.

5. **grpc.keepalive_timeout_ms**
   - **역할**: 클라이언트가 keepalive ping에 대한 응답을 기다리는 시간을 설정합니다. 이 시간이 초과되면 연결이 끊어집니다.
   - **단위**: 밀리초 (Milliseconds)
   - **사례**: 20초(20000밀리초)로 설정하면, 클라이언트는 20초 동안 서버의 응답을 기다립니다. 응답이 없으면 연결을 끊습니다.

6. **grpc.keepalive_permit_without_calls**
   - **역할**: 활성 호출이 없어도 keepalive ping을 보낼 수 있도록 설정합니다.
   - **단위**: 부울 (Boolean)
   - **사례**: `1`로 설정하면, 활성 호출이 없어도 keepalive ping을 보낼 수 있습니다. 이는 장시간 유휴 상태에서도 연결을 유지하는 데 유용합니다.

7. **grpc.http2.min_time_between_pings_ms**
   - **역할**: 두 개의 keepalive ping 사이의 최소 시간을 설정합니다.
   - **단위**: 밀리초 (Milliseconds)
   - **사례**: 10초(10000밀리초)로 설정하면, 두 개의 ping 사이에 최소 10초가 지나야 합니다. 이는 ping의 빈도를 제한하여 네트워크 부하를 줄이는 데 유용합니다.

8. **grpc.http2.max_pings_without_data**
   - **역할**: 데이터 없이 보낼 수 있는 최대 ping 수를 설정합니다.
   - **단위**: 정수 (Integer)
   - **사례**: `0`으로 설정하면, 데이터 없이 무제한으로 ping을 보낼 수 있습니다. 이는 연결을 유지하는 데 유용하지만, 네트워크 부하를 증가시킬 수 있습니다.

9. **grpc.enable_http_proxy**
   - **역할**: HTTP 프록시를 사용할지 여부를 설정합니다.
   - **단위**: 부울 (Boolean)
   - **사례**: `0`으로 설정하면, HTTP 프록시를 사용하지 않습니다. 이는 프록시를 사용하지 않는 환경에서 유용합니다.

10. **timeout**
    - **역할**: RPC 호출의 전체 타임아웃을 설정합니다. 이 시간 내에 응답이 없으면 호출이 실패합니다.
    - **단위**: 마이크로초 (Microseconds)
    - **사례**: 3초(3000000마이크로초)로 설정하면, 클라이언트는 3초 동안 서버의 응답을 기다립니다. 응답이 없으면 호출이 실패합니다.

gRPC PHP 확장에서 지원하는 모든 채널 옵션은 [gRPC PHP Channel Options](https://github.com/grpc/grpc/blob/master/src/php/ext/grpc/channel.c)에서 확인할 수 있습니다.

### 예제

예를 들어, 다음과 같은 상황을 가정해 보겠습니다:

- 클라이언트는 서버와의 연결을 유지하기 위해 60초마다 keepalive ping을 보내고, 20초 동안 응답을 기다립니다.
- 클라이언트는 큰 파일을 전송할 수 있도록 최대 메시지 크기를 무제한으로 설정합니다.
- 클라이언트는 3초 동안 서버의 응답을 기다립니다.

```php
$this->client = new PromotionServiceClient('some.host.internal', [
    'credentials' => Grpc\ChannelCredentials::createSsl(),
    'grpc.max_receive_message_length' => -1,
    'grpc.max_send_message_length' => -1,
    'grpc.keepalive_time_ms' => 60000, // 60 seconds
    'grpc.keepalive_timeout_ms' => 20000, // 20 seconds
    'grpc.keepalive_permit_without_calls' => 1,
    'grpc.http2.min_time_between_pings_ms' => 10000, // 10 seconds
    'grpc.http2.max_pings_without_data' => 0,
    'grpc.enable_http_proxy' => 0,
    'timeout' => 3000000, // 3 seconds in microseconds. 서버 응답을 기다리는 최대 시간을 나타냅니다.
]);
```

## 응답 처리

```php
// app/Vendor/grpc/grpc/src/lib/AbstractCall.php
protected function _deserializeResponse($value)
{
    if ($value === null) {
        return;
    }

    // Proto3 implementation
    if (is_array($this->deserialize)) {
        // ex: ['\Finance\Chai\Channel\ListChannelsResponse', 'decode']
        list($className, $deserializeFunc) = $this->deserialize;
        $obj = new $className();
        if (method_exists($obj, $deserializeFunc)) {
            $obj->$deserializeFunc($value);
        } else {
            $obj->mergeFromString($value);
        }

        return $obj;
    }

    // Protobuf-PHP implementation
    return call_user_func($this->deserialize, $value);
}
```

`decode` 함수가 정의되어 있지 않은 경우 `mergeFromString`가 호출된다.

## Read

PHP에서 프로토콜 버퍼 메시지의 속성들은 `$this->$oneof_name` 처럼 기본적으로 문자열 이름 기반으로 동작한다.

```php
/**
 * Generated from protobuf field <code>.path.to.entity.SomeOneOf some_one_of = 5;</code>
 * @return \Path\To\Entity\SomeOneOf|null
 */
public function getSomeOneOf()
{
    return $this->readOneof(5);
}
```

```php
// app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Message.php
protected function readOneof($number)
{
    $field = $this->desc->getFieldByNumber($number);
    $oneof = $this->desc->getOneofDecl()[$field->getOneofIndex()];
    $oneof_name = $oneof->getName();
    $oneof_field = $this->$oneof_name;
    if ($number === $oneof_field->getNumber()) {
        return $oneof_field->getValue();
    } else {
        return $this->defaultValue($field);
    }
}
```

```php
// app/Vendor/google/protobuf/src/Google/Protobuf/Internal/Descriptor.php
// $this->desc 의 타입
namespace Google\Protobuf\Internal;

class Descriptor
{
    use HasPublicDescriptorTrait;

    private $full_name;
    private $field = [];
    private $json_to_field = [];
    private $name_to_field = [];
    private $index_to_field = [];
    private $nested_type = [];
    private $enum_type = [];
    private $klass;
    private $legacy_klass;
    private $options;
    private $oneof_decl = [];

    ... 생략 ...

    public function getOneofDecl()
    {
        return $this->oneof_decl;
    }

    ... 생략 ...
```
