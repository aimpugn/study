# gRPC troubleshooting

- [gRPC troubleshooting](#grpc-troubleshooting)
    - [Import ... was not found or had errors](#import--was-not-found-or-had-errors)
        - [문제](#문제)
        - [원인](#원인)
        - [해결](#해결)
    - [Cannot resolve import 'entity/pg\_provider.proto'](#cannot-resolve-import-entitypg_providerproto)
        - [문제](#문제-1)
        - [원인](#원인-1)
        - [해결](#해결-1)
    - [INTERNAL: http2 exception](#internal-http2-exception)
        - [문제](#문제-2)
        - [원인](#원인-2)
        - [해결](#해결-2)
    - [protoc-gen-grpc-java not available on apple m1](#protoc-gen-grpc-java-not-available-on-apple-m1)
        - [문제](#문제-3)
        - [원인](#원인-3)
        - [해결](#해결-3)
    - [failed to connect to all addresses](#failed-to-connect-to-all-addresses)
        - [문제](#문제-4)
        - [원인](#원인-4)
            - [grpcurl 설치](#grpcurl-설치)
        - [해결](#해결-4)
    - [DEADLINE\_EXCEEDED when gRPC call from PHP client to MSA in k8s](#deadline_exceeded-when-grpc-call-from-php-client-to-msa-in-k8s)
        - [문제](#문제-5)
        - [원인](#원인-5)
            - [디버깅 과정](#디버깅-과정)
            - [참고](#참고)
        - [해결](#해결-5)
    - [GPBDecodeException: Error occurred during parsing: Fail to push limit](#gpbdecodeexception-error-occurred-during-parsing-fail-to-push-limit)
        - [문제](#문제-6)
        - [원인](#원인-6)
        - [해결](#해결-6)
    - ["google.api.Http.rules" is already defined in file "google/api/http.proto" 등](#googleapihttprules-is-already-defined-in-file-googleapihttpproto-등)
        - [문제](#문제-7)
        - [원인](#원인-7)
        - [해결](#해결-7)

## Import ... was not found or had errors

### 문제

```log
entity/allowed_pg_merchant_id_prefix.proto: File not found.
entity/core_certification_channel.proto: File not found.
entity/core_default_channel.proto: File not found.
entity/core_extra_channels.proto: File not found.
entity/core_subscription_channel.proto: File not found.
entity/pg_provider.proto: File not found.
v2/channel-service/backoffice_core_channel_service.proto:7:1: Import "entity/allowed_pg_merchant_id_prefix.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:8:1: Import "entity/core_certification_channel.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:9:1: Import "entity/core_default_channel.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:10:1: Import "entity/core_extra_channels.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:11:1: Import "entity/core_subscription_channel.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:12:1: Import "entity/pg_provider.proto" was not found or had errors.
v2/channel-service/backoffice_core_channel_service.proto:141:3: "CoreDefaultChannel" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:149:12: "CoreExtraChannel" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:157:12: "CoreCertificationChannel" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:165:12: "CoreSubscriptionChannel" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:173:12: "AllowedByPgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:178:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:184:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:191:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:197:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:210:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:218:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:225:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:233:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:280:3: "UpdatePgSpecificCredentialOverrideRule" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:281:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:285:3: "PgSpecificCredentialOverrideRule" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:286:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:290:3: "PgProvider" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:292:5: "InicisOverrideRuleRequest" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:297:3: "PgSpecificCredentialOverrideRule" is not defined.
v2/channel-service/backoffice_core_channel_service.proto:298:3: "PgProvider" is not defined.
```

### 원인

protoc로 protobuf 생성하는 프로젝트에 `proto`를 관리하는 디렉토리에 여러 메시지 정의들이 존재

```tree
 rody  ~/IdeaProjects/internal-interface   main ±  tree proto
proto
├── cardinfo
│   └── v1
│       └── cardinfo.proto
├── google
│   ├── api
│   │   ├── annotations.proto
│   │   ├── http.proto
│   │   └── httpbody.proto
│   └── protobuf
│       ├── any.proto
│       ├── any_test.proto
│       ├── api.proto
│       ├── descriptor.proto
│       ├── duration.proto
│       ├── empty.proto
│       ├── field_mask.proto
│       ├── map_lite_unittest.proto
│       ├── map_proto2_unittest.proto
│       ├── map_unittest.proto
│       ├── source_context.proto
│       ├── struct.proto
│       ├── test_messages_proto2.proto
│       ├── test_messages_proto3.proto
│       ├── timestamp.proto
│       ├── type.proto
│       ├── unittest.proto
│       ├── unittest_arena.proto
│       ├── unittest_custom_options.proto
│       ├── unittest_drop_unknown_fields.proto
│       ├── unittest_embed_optimize_for.proto
│       ├── unittest_empty.proto
│       ├── unittest_enormous_descriptor.proto
│       ├── unittest_import.proto
│       ├── unittest_import_lite.proto
│       ├── unittest_import_public.proto
│       ├── unittest_import_public_lite.proto
│       ├── unittest_lazy_dependencies.proto
│       ├── unittest_lazy_dependencies_custom_option.proto
│       ├── unittest_lazy_dependencies_enum.proto
│       ├── unittest_lite.proto
│       ├── unittest_lite_imports_nonlite.proto
│       ├── unittest_mset.proto
│       ├── unittest_mset_wire_format.proto
│       ├── unittest_no_field_presence.proto
│       ├── unittest_no_generic_services.proto
│       ├── unittest_optimize_for.proto
│       ├── unittest_preserve_unknown_enum.proto
│       ├── unittest_preserve_unknown_enum2.proto
│       ├── unittest_proto3.proto
│       ├── unittest_proto3_arena.proto
│       ├── unittest_proto3_arena_lite.proto
│       ├── unittest_proto3_lite.proto
│       ├── unittest_proto3_optional.proto
│       ├── unittest_well_known_types.proto
│       └── wrappers.proto
├── heimdall
│   └── rate_limit.proto
├── payments
│   └── v1
│       └── payments.proto
├── paymentsinfo
│   └── v1
│       └── paymentsinfo.proto
├── payport
│   └── v1
│       ├── authorize
│       │   └── authorize.proto
│       ├── basic
│       │   └── basic.proto
│       ├── card
│       │   └── card.proto
│       ├── keyboard
│       │   └── keyboard.proto
│       ├── txs
│       │   └── txs.proto
│       └── user
│           └── user.proto
├── pgintegration
│   └── v1
│       └── pgintegration.proto
├── promotion
│   └── v1
│       └── promotion.proto
├── txnotifier
│   └── v1
│       └── txnotifier.proto
├── v2
│   ├── basis
│   │   └── basis.proto
│   ├── channel-service
│   │   ├── backoffice_core_channel_service.proto
│   │   ├── channel_service.proto
│   │   ├── entity
│   │   │   ├── allowed_pg_merchant_id_prefix.proto
│   │   │   ├── application.proto
│   │   │   ├── channel.proto
│   │   │   ├── channel_attribute.proto
│   │   │   ├── channel_type.proto
│   │   │   ├── core_certification_channel.proto
│   │   │   ├── core_channel.proto
│   │   │   ├── core_default_channel.proto
│   │   │   ├── core_extra_channels.proto
│   │   │   ├── core_pg_credential.proto
│   │   │   ├── core_subscription_channel.proto
│   │   │   ├── merchant_form.proto
│   │   │   ├── pg_provider.proto
│   │   │   ├── pg_setting.proto
│   │   │   └── status.proto
│   │   ├── google
│   │   │   └── api
│   │   │       ├── annotations.proto
│   │   │       ├── http.proto
│   │   │       └── httpbody.proto
│   │   ├── healthcheck_service.proto
│   │   ├── onboarding_service.proto
│   │   └── protoc-gen-openapiv2
│   │       └── options
│   │           ├── annotations.proto
│   │           └── openapiv2.proto
│   ├── merchant-service
│   │   ├── merchant
│   │   │   └── merchant_service.proto
│   │   ├── pgapply
│   │   │   └── pg_application_service.proto
│   │   ├── store
│   │   │   └── store_service.proto
│   │   ├── term
│   │   │   └── terms.proto
│   │   └── user
│   │       └── user_service.proto
│   ├── payments
│   │   └── card
│   │       └── card.proto
│   └── txs
│       └── txs.proto
└── van
    └── v1
        └── van.proto

43 directories, 95 files
```

여기서 생성하려는 프로토의 디렉토리가 `proto/v2/channel-service`에 있는데, 내부에서 `import "entity/..."` 식으로 되어 있는데, 이때

### 해결

- 경로를 변경한 후에 generate한다. 이때 `--proto_path`(`-I` 옵션)은 proto 디렉토리가 된다

```tree
proto
│
...
│
├── v2
│   ├── basis
│   │   └── basis.proto
│   ├── channel-service
│   │   ├── backoffice_core_channel_service.proto
│   │   ├── channel_service.proto
│   │   ├── entity
│   │   │   ├── allowed_pg_merchant_id_prefix.proto
```

```diff
- import "entity/allowed_pg_merchant_id_prefix.proto";
+ import "v2/channel-service/entity/allowed_pg_merchant_id_prefix.proto";
```

## Cannot resolve import 'entity/pg_provider.proto'

### 문제

- proto 파일에서 다른 proto 파일 임포트 시에 경로를 리졸브하지 못하는 문제

### 원인

- [기본적으로, `intellij-protobuf-editor`는 프로젝트에 구성된 소스 루트를 protobuf 임포트 경로로 사용한다](https://stackoverflow.com/a/62843504)고 한다. 만약 이게 올바르지 않다면, `Settings > Languages & Frameworks > Protocol Buffers`에서 `Configure automatically`를 체크 해제하고 오버라이드하거나 추가할 수 있다.
- `.../src/main/java/proto` (where ... means whatever your project's base path is).

### 해결

- 임시 방편
    - 'entity/pg_provider.proto' -> '{SERVICE}/entity/pg_provider.proto' 같은 식으로 경로를 치환
- 근본적인 해결법?

## INTERNAL: http2 exception

### 문제

gRPC 호출 시에 아래와 같은 에러가 발생

```json
{"@timestamp":"2023-01-11T18:09:38.754+09:00","logger_name":"finance.chai.gateway.transaction.repository.ChannelServiceChannelRepository","thread_name":"DefaultDispatcher-worker-2","level":"info","msg":"INTERNAL: http2 exception","context":{}}
```

```log
Caused by: io.netty.handler.codec.http2.Http2Exception: First received frame was not SETTINGS. Hex dump for first 5 bytes: 485454502f
```

### 원인

- MSA 내부적으로는 평문 통신을 하지만, 외부에서 URL 통해서 접근 시에는 TLS 통신하도록 되어 있다. 따라서 gRPC 채널 생성 옵션에서 TLS 사용하도록 추가해야 함
- `485454502f` hex를 string으로 변환하면 `HTTP/`이 된다. HTTPS가 아니라서 발생하는 문제로 보임

### 해결

```kotlin
// ASIS
@Bean(name = ["channelServiceGrpcClient"])
fun channelServiceGrpcClient(
    @Value("\${internal.grpc.channel-service.host}")
    host: String,
    @Value("\${internal.grpc.channel-service.port}")
    port: Int
): ChannelServiceGrpcKt.ChannelServiceCoroutineStub {
    return ChannelServiceGrpcKt.ChannelServiceCoroutineStub(
        ManagedChannelBuilder
            .forAddress(host, port)
            .intercept(GrpcClientInterceptor())
            .usePlaintext()
            .build()
    )
}
```

```kotlin
// TOBE
@Bean(name = ["channelServiceGrpcClient"])
fun channelServiceGrpcClient(
    @Value("\${internal.grpc.channel-service.host}")
    host: String,
    @Value("\${internal.grpc.channel-service.port}")
    port: Int
): ChannelServiceGrpcKt.ChannelServiceCoroutineStub {
    return ChannelServiceGrpcKt.ChannelServiceCoroutineStub(
        ManagedChannelBuilder
            .forAddress(host, port)
            .intercept(GrpcClientInterceptor())
            .useTransportSecurity() // <------------- 변경
            .build()
    )
}
```

## [protoc-gen-grpc-java not available on apple m1](https://github.com/grpc/grpc-java/issues/7690)

### 문제

Apple silicon m1 chip 맥북에서 gRPC 코드 생성 실패

### 원인

arm64용 gRPC 생성이 불가

### 해결

인텔 `osx-x86_64`으로 명시하여 생성

```kts
protobuf {
  protoc {
    artifact = 'com.google.protobuf:protoc:3.14.0:osx-x86_64'
  }
}
```

```xml
<protocArtifact>com.google.protobuf:protoc:3.14.0:exe:osx-x86_64</protocArtifact>
```

[`osdetector.os`를 사용하여 분기처리](https://github.com/grpc/grpc-java/issues/7690#issuecomment-760332746)할 수 있다

```gradle
protobuf {
  protoc {
    if (osdetector.os == "osx") {
      artifact = 'com.google.protobuf:protoc:3.14.0:osx-x86_64'
    } else {
      artifact = 'com.google.protobuf:protoc:3.14.0'
    }
  }
}
```

```gradle
protobuf {
    protoc {
        if ("aarch64" == System.getProperty("os.arch")) {
            // mac m1
            artifact = "com.google.protobuf:protoc:3.10.0:osx-x86_64"
        } else {
            // other
            artifact = "com.google.protobuf:protoc:3.10.0"
        }
    }
}
```

## failed to connect to all addresses

### 문제

local에서 TGS로 gRPC 콜을 하는데 Exception 발생

```log
Array
(
    [title] => ############# TGS Exception ############
    [data] => Array
        (
            [status] => 14
            [code] => UNAVAILABLE
            [message] => failed to connect to all addresses
        )

)
```

### 원인

인증서 없는데, TLS handshake 시도해서 문제였다!

#### [grpcurl 설치](https://princepereira.medium.com/install-grpccurl-in-ubuntu-6ad71fd3ed31)

go는 다운로드 받아서 volume으로 연결되어 있는 디렉토리에 이동시킨 상태에서 진행

```shell
tar --directory="/usr/local" -xzf go1.20.linux-amd64.tar.gz
mkdir -p $HOME/GoProjects/src -p $HOME/GoProjects/pkg -p $HOME/GoProjects/bin
# PATH variable 추가
echo "export PATH=$PATH:/usr/local/go/bin:$HOME/GoProjects/bin" >> $HOME/.bashrc
# GOPATH 추가
echo "export GOPATH=$HOME/GoProjects" >> $HOME/.bashrc
# 현재 터미널에 반영
source $HOME/.bashrc
```

go version 확인

```shell
root@e8f9fc287ba6:/home/ubuntu/test-resources# go version
go version go1.20 linux/amd64
```

go 프로젝트 경로로 이동해서 설치

```shell
cd $GOPATH
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
```

```shell
grpcurl -help

The 'address' is only optional when used with 'list' or 'describe' and a
protoset or proto flag is provided.

If 'list' is indicated, the symbol (if present) should be a fully-qualified
service name. If present, all methods of that service are listed. If not
present, all exposed services are listed, or all services defined in protosets.

If 'describe' is indicated, the descriptor for the given symbol is shown. The
symbol should be a fully-qualified service, enum, or message name. If no symbol
is given then the descriptors for all exposed or known services are shown.

If neither verb is present, the symbol must be a fully-qualified method name in
'service/method' or 'service.method' format. In this case, the request body will
be used to invoke the named method. If no body is given but one is required
(i.e. the method is unary or server-streaming), an empty instance of the
method's request type will be sent.

The address will typically be in the form "host:port" where host can be an IP
address or a hostname and port is a numeric port or service name. If an IPv6
address is given, it must be surrounded by brackets, like "[2001:db8::1]". For
Unix variants, if a -unix=true flag is present, then the address must be the
path to the domain socket.

... 생략 ...
```

### 해결

php 코드에서 TLS 통신하려는 코드를 수정

```diff
$checkoutTransactionServiceClient = new CheckoutTransactionServiceClient(
    HostUtil::getTransactionGatewayServiceHost(),
-   ['credentials' => ChannelCredentials::createSsl()]
+   ['credentials' => ChannelCredentials::createInsecure()]
);
```

## DEADLINE_EXCEEDED when gRPC call from PHP client to MSA in k8s

### 문제

PHP 클라이언트에서 k8s 내에 있는 MSA에 gRPC 콜을 하면 DEADLINE_EXCEEDED 발생

### 원인

#### 디버깅 과정

tmux

grpcurl 설치

로컬 -> 체크아웃서비스 -> some-A-serivce -> merchant

grpcurl --help | less `2>&1`

git clone --bare

git bundle create --all

scp `bundle` remote_server:/path/to/save

`shasum`으로 같은 파일인지 체크

fpm/pool.d/log/www.log.slow 슬로우 로그...

풀의 워커가 죽고 재생성되는 건 그렇게 큰 부하가 아니다.

`cgi-fcgi`

idle 프로세스의 수를 센다?

`watch` jq 'del()'

[php & grpc LB](https://github.com/grpc/grpc/blob/master/doc/load-balancing.md)
[gRPC Name Resolution](https://github.com/grpc/grpc/blob/master/doc/naming.md)

apm과 log는 다르다

systemctl php-fpm

`degraded` 상태

systemctl list-units --failed

`ntp`가 실패?

`synchrozie-panes on | off`

`systemctl-timesynd`, `systemd-timedated`

`timedatectl`

`chronyd`

datadog apm 설치 과정

```shell
mkdir -p /tmp/removeme
cd /tmp/removeme
curl -LO 'https://github.com/DataDog/dd-trace-php/releases/latest/download/datadog-setup.php'
sudo php datadog-setup.php --php-bin=all
sudo vim /etc/php/5.6/fpm/pool.d/www.conf
# add below
# env[DD_SERVICE] = 'core-pay'
# env[DD_ENV] = 'prod'
sudo systemctl reload php5.6-fpm
```

#### 참고

- [Getting StatusCode.DEADLINE_EXCEEDED when trying to make gRPC call to Tensorflow Serving Kubernetes pod](https://stackoverflow.com/questions/61182590/getting-statuscode-deadline-exceeded-when-trying-to-make-grpc-call-to-tensorflow)
- [All GRPC calls failed with DEADLINE_EXCEEDED](https://github.com/grpc/grpc-java/issues/9072)
- [gRPC PHP - Keepalive User Guide for gRPC Core (and dependents)](https://grpc.github.io/grpc/php/md_doc_keepalive.html)
- [how to keep alive connection with php?](https://github.com/grpc/grpc/issues/23098)
    - [keep alive options on github.com/grpc/grpc](https://github.com/grpc/grpc/blob/ca08fd9f9b8043c6ce18071698d8d98b5b7ea1b0/include/grpc/impl/codegen/grpc_types.h#L220-L230)
- [NodeJS gRPC server in k8s QUEUE_TIMEOUT](https://groups.google.com/g/grpc-io/c/ZuGBYYjGEuk)

### 해결

아직까지는 딱히 원인 파악 못했고, 해결법도 못 찾음

datadog apm 설치는 했고, dd agent(?) 추가도 해야할 듯?

## GPBDecodeException: Error occurred during parsing: Fail to push limit

### 문제

gRPC와 함께 동작하는 유닛 테스트 실행하는데 에러 발생

```log
GPBDecodeException: Error occurred during parsing: Fail to push limit
```

### 원인

The solution was found at [Fail to push limit #566](https://github.com/googleads/google-ads-php/issues/566):

Fail to push limit. can occur if you modify the auto generated protobuf metadata files.

For example, if you upload files via an FTP client. Some versions of FileZilla modify line endings, see [filezilla Data Type](https://wiki.filezilla-project.org/Data_Type).

Removal of any characters from the binary data results in unexpected byte length comparisons. Think of it like serialized data: s:5:"123" would fail because the string length is 3 when it was expected to be 5.

You can change the transfer data type in FileZilla via the main menu under Transfer -> Transfer type and selecting Binary. Then reupload all the vendor files.

### 해결

다시 생성해서 기존 protobuf 덮어쓰기

## "google.api.Http.rules" is already defined in file "google/api/http.proto" 등

### 문제

```bash
MyPkg/google/api/http.proto:34:21: "google.api.Http.rules" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:42:8: "google.api.Http.fully_decode_reserved_expansion" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:30:9: "google.api.Http" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto: "google.api.HttpRule.pattern" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:266:10: "google.api.HttpRule.selector" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:273:12: "google.api.HttpRule.get" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:276:12: "google.api.HttpRule.put" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:279:12: "google.api.HttpRule.post" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:282:12: "google.api.HttpRule.delete" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:285:12: "google.api.HttpRule.patch" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:291:23: "google.api.HttpRule.custom" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:298:10: "google.api.HttpRule.body" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:303:10: "google.api.HttpRule.response_body" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:308:21: "google.api.HttpRule.additional_bindings" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:262:9: "google.api.HttpRule" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:314:10: "google.api.CustomHttpPattern.kind" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:317:10: "google.api.CustomHttpPattern.path" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:312:9: "google.api.CustomHttpPattern" is already defined in file "google/api/http.proto".
MyPkg/google/api/http.proto:34:12: "google.api.HttpRule" seems to be defined in "google/api/http.proto", which is not imported by "MyPkg/google/api/http.proto".  To use it here, please add the necessary import.
MyPkg/google/api/http.proto:291:5: "google.api.CustomHttpPattern" seems to be defined in "google/api/http.proto", which is not imported by "MyPkg/google/api/http.proto".  To use it here, please add the necessary import.
MyPkg/google/api/http.proto:308:12: "google.api.HttpRule" seems to be defined in "google/api/http.proto", which is not imported by "MyPkg/google/api/http.proto".  To use it here, please add the necessary import.
```

### 원인

이미 google~ 패키지가 있는데 중복 설정돼서 에러가 발생

### 해결
