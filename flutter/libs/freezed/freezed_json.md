# Freezed json

## `json_serializable`

> Note:
>
> Freezed will only generate a fromJson if the factory is using `=>`.

```yaml
#  pubspec.yaml
dev_dependencies:
  json_serializable:
```

## `build.yaml`

```yaml
targets:
  $default:
    builders:
      json_serializable:
        options:
          # Options configure how source code is generated for every
          # `@JsonSerializable`-annotated class in the package.
          #
          # The default value for each is listed.
          any_map: false
          checked: false
          constructor: ""
          create_factory: true
          create_field_map: false
          create_per_field_to_json: false
          create_to_json: true
          disallow_unrecognized_keys: false
          explicit_to_json: false
          field_rename: none
          generic_argument_factories: false
          ignore_unannotated: false
          include_if_null: true
```

## fromJSON - classes with multiple constructors

### What about `@JsonSerializable` annotation?

```dart
@freezed
class Example with _$Example {
  // You can pass `@JsonSerializable` annotation by placing it over constructor e.g.
  @JsonSerializable(explicitToJson: true)
  factory Example(@JsonKey(name: 'my_property') SomeOtherClass myProperty) = _Example;

  factory Example.fromJson(Map<String, dynamic> json) => _$ExampleFromJson(json);
}
```
