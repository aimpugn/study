# Generic

- [Generic](#generic)
    - [제네릭 in dart](#제네릭-in-dart)
    - [함수형으로 추상화](#함수형으로-추상화)
    - [`extends` 제네릭 사용](#extends-제네릭-사용)

## 제네릭 in dart

- Dart에서 제네릭은 타입의 재사용성을 높이고, 타입 안정성을 강화하기 위해 사용된다
- 제네릭을 사용하면
    - 컴파일 시에 타입 정보를 보존하고,
    - 런타임에 타입 체크를 할 수 있다

## 함수형으로 추상화

```dart
import 'package:dilectio_app/src/subprojects/user/domain/repositories/repository_interface.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fpdart/fpdart.dart';

mixin DomainHandler<RepositoryType extends RepositoryInterface, StateType,
    ErrorType, SuccessType> on AutoDisposeAsyncNotifier<StateType> {

  ReaderTask<RepositoryType, void> taskHandler(
    Future<TaskEither<ErrorType, SuccessType>> Function(RepositoryType)
        function,
    AsyncValue<StateType> Function(ErrorType) onError,
    AsyncValue<StateType> Function(SuccessType) onSuccess,
  ) =>
      ReaderTask<RepositoryType, void>.Do((_) async {
        state = AsyncValue<StateType>.loading();

        final taskEither = await _(ReaderTask(function));

        final either = await taskEither.run();

        state = either.fold(
          (l) => onError(l),
          (r) => onSuccess(r),
        );
      });
}
```

## `extends` 제네릭 사용

```dart
class GreyFilledSeparatingTagsInput<TagType extends Tag> extends ConsumerStatefulWidget {
                                        ^^^^^^^^ 제네릭 타입 매개변수
  @override
  ConsumerState<ConsumerStatefulWidget> createState() =>
      GreyFilledSpaceSeparatingTagsState<TagType>();
}

class GreyFilledSpaceSeparatingTagsState<TagType extends Tag>
                                        ^^^^^^^^^^^^^^^^^^^^^ 
                                        상위 위젯인 `GreyFilledSeparatingTagsInput`의 타입과 일치해야 한다
    extends ConsumerState<GreyFilledSeparatingTagsInput<TagType>> {
            ^                                                ^^^^^ TagType 사용 가능
            ^ `GreyFilledSeparatingTagsInput` 상태를 관리하는 클래스
  final List<TagType> _tags = [];

  @override
  void dispose() {
    super.dispose();
    _controller.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (widget.labelName != null && widget.labelName is String)
          Text(widget.labelName!, style: widget.labelTextStyle),
        gap8,
        Autocomplete<TagType>(
          optionsBuilder: (TextEditingValue textEditingValue) {
            if (textEditingValue.text == '') {
              return const Iterable<TagType>.empty();
            }
            return _tags.where((TagType option) {
              option
              return option.contains(textEditingValue.text.toLowerCase());
            });
          },
        ),
```

1. **'TagType' doesn't conform to the bound 'Object' 에러**:
   - Dart에서 모든 타입은 기본적으로 `Object`를 상속받습니다. 그러나 제네릭에서 특정 타입을 사용할 때, Dart의 타입 체커는 명시적으로 타입이 `Object`의 하위 타입임을 확인해야 합니다. 여기서 발생한 에러는 `TagType`이 `Object`의 하위 타입임이 명확하지 않다는 것을 의미합니다.
   - 이 문제를 해결하기 위해서는 `TagType`이 `Object`를 상속받는다는 것을 명시적으로 선언해야 합니다. Dart에서는 모든 클래스가 기본적으로 `Object`를 상속받으므로, 이 문제는 일반적으로 발생하지 않습니다. 그러나 코드의 다른 부분에서 `TagType`에 대한 타입 제한이 잘못 설정되었을 수 있습니다.

이 두 문제를 해결하기 위한 코드 수정은 다음과 같습니다:

1. `GreyFilledSpaceSeparatingTagsState` 클래스를 `ConsumerState<GreyFilledSeparatingTagsInput<TagType>>`로 변경합니다.
2. `TagType`에 대해 `Object`의 하위 타입임을 명시적으로 선언합니다. 예를 들어, `Tag` 클래스가 이미 `Object`를 상속받고 있으므로, `TagType extends Tag` 선언은 충분합니다.

수정된 코드는 다음과 같습니다:

```dart
class GreyFilledSeparatingTagsInput<TagType extends Tag> extends ConsumerStatefulWidget {
  // 기존 필드와 생성자...

  @override
  ConsumerState<ConsumerStatefulWidget> createState() =>
      GreyFilledSpaceSeparatingTagsState<TagType>();
}

class GreyFilledSpaceSeparatingTagsState<TagType extends Tag>
    extends ConsumerState<GreyFilledSeparatingTagsInput<TagType>> {
  // 클래스 구현...
}
```

이렇게 하면 `GreyFilledSpaceSeparatingTagsState` 클래스는 `GreyFilledSeparatingTagsInput` 클래스와 동일한 제네릭 타입 매개변수를 사용하며, `TagType`은 `Tag`의 하위 타입임이 보장됩니다.
