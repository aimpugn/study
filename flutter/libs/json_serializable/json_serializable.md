# json_serializable

- [json\_serializable](#json_serializable)
    - [`explicitToJson`, `explicit_to_json`](#explicittojson-explicit_to_json)
        - [by `build.yaml`](#by-buildyaml)
    - [왜 `json_serializable` package는 자동으로 자식 오브젝트의 `toJson()`을 호출 안하는가?](#왜-json_serializable-package는-자동으로-자식-오브젝트의-tojson을-호출-안하는가)
    - [기타](#기타)

## `explicitToJson`, `explicit_to_json`

### by `build.yaml`

```yaml
targets:
  $default:
    builders:
      json_serializable:
        options:
          explicit_to_json: true
```

## 왜 `json_serializable` package는 자동으로 자식 오브젝트의 `toJson()`을 호출 안하는가?

`flutter pub run build_runner build`를 실행하여 생성된 `*.g.dart`가 자동으로 `toJson()`를 호출하지 않는다

> This is by design. The JSON encoding logic in `dart:convert` calls `toJson` for you, which means calls to `toJson` are lazy.

```dart

part 'app_user.g.dart';
part 'app_user.freezed.dart';

@freezed
class AppUser with _$AppUser {
  const AppUser._();

  const factory AppUser({
    required DefaultInfo defaultInfo,
    required Option<EducationInfo> education,
    required Option<WorkplaceInfo> workplace,
  }) = _AppUser;

  factory AppUser.fromJson(Map<String, dynamic> json) =>
      _$AppUserFromJson(json);

  Future<void> sendEmailVerification() async {}
}
```

```dart
part 'default_info.freezed.dart';
part 'default_info.g.dart';

@freezed
class DefaultInfo with _$DefaultInfo {
  @JsonSerializable(explicitToJson: true)
  const factory DefaultInfo({
    required String uid,
    required String email,
    required bool emailVerified,
    required Option<String> name,
    required Option<String> phone,
    required Option<bool> phoneVerified,
    required Option<List<Uri>>? profileUris,
  }) = _DefaultInfo;

  factory DefaultInfo.fromJson(Map<String, dynamic> json) =>
      _$DefaultInfoFromJson(json);
}
```

이 상태에서 `@JsonSerializable(explicitToJson: true)`가 없다면, `DefaultInfo`를 json으로 변하는 데 실패한다

```log
Left(DomainErrors(message: Internal Server Error, error: Invalid argument (dartObject): Could not convert: Instance of '_$36_DefaultInfo', stackTrace: dart-sdk/lib/_internal/js_dev_runtime/private/ddc_runtime/errors.dart 294:49))
```

[`explicitToJson`](https://pub.dev/documentation/json_annotation/latest/json_annotation/JsonSerializable/explicitToJson.html) 속성을 사용하여 이를 오버라이딩할 수 있다

## 기타

- [Automatically call .toJson() of children objects in _$<NameOfObject>ToJson(<Object>)](https://github.com/google/json_serializable.dart/issues/561)
- [Why is explicitToJson not default? #620](https://github.com/google/json_serializable.dart/issues/620)
