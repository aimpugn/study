# riverpod generator

- [riverpod generator](#riverpod-generator)
    - [`@riverpod` on Class](#riverpod-on-class)
    - [`@Riverpod` on function](#riverpod-on-function)
    - [syntax](#syntax)
    - [기타](#기타)

## `@riverpod` on Class

![async notifier void](../resources/riverpod/async-notifier-void.webp)

- 위 이미지를 보면 `void`, `FutureOr<void>`, 그리고 `AsyncValue<void>`를 다룬다.
- `state`의 타입은 `AsyncValue<void>`이 된다
    - 왜? `build` 메서드의 리턴 타입이 `FutureOr<void>`이기 때문
    - 따라서 상태(`state`)를 `AsyncData`, `AsyncLoading`, `AsyncError`로 설정할 수 있다.
- `ProviderFor`는 클래스가 된다

- class 정의

    ```dart
    part 'sign_in_controller.g.dart';

    @riverpod
    class SignInController extends _$SignInController {
        final log = Logger();

        @override
        FutureOr<void> build() {}

        Future<bool> loginByEmail({
            required String email,
            required String password,
        }) async {
            state = const AsyncValue.loading();
        //  ^^^^^ `void set state(AsyncValue<void> newState)`

            final usersService = ref.watch(appUserApplicationProvider);
            state = await AsyncValue.guard(
                () async => usersService.getJWT(email, password));

            return state.hasError == false;
        }
    }
    ```

- 생성된 프로바이더

    ```dart
    // GENERATED CODE - DO NOT MODIFY BY HAND

    part of 'sign_in_controller.dart';

    // **************************************************************************
    // RiverpodGenerator
    // **************************************************************************

    String _$signInControllerHash() => r'c4088e3dfcbc32aa1733131c27a496be3b21821e';

    /// See also [SignInController].
    @ProviderFor(SignInController)
    final signInControllerProvider =
        AutoDisposeAsyncNotifierProvider<SignInController, void>.internal(
    SignInController.new,
    name: r'signInControllerProvider',
    debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
        ? null
        : _$signInControllerHash,
    dependencies: null,
    allTransitiveDependencies: null,
    );

    typedef _$SignInController = AutoDisposeAsyncNotifier<void>;
    // ignore_for_file: type=lint
    // ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member
    ```

## `@Riverpod` on function

- 이때 `SignInController`는 그저 `ref`를 갖고 있는 클래스일 뿐이다. 따라서 자체적인 `state`가 없다.
- `ref.state`는 `SignInController`가 된다
- `ProviderFor`는 함수가 된다

- 함수 정의

    ```dart
    class SignInController {
        final SignInControllerRef ref;

        const SignInController({required this.ref});

        Future<bool> loginByEmail({
            required String email,
            required String password,
        }) async {
            ref.state; // SignInController
            return true;
        }
    }

    @Riverpod()
    SignInController signInController(SignInControllerRef ref) {
        return SignInController(ref: ref);  
    }
    ```

- 생성된 프로바이더

    ```dart
    // GENERATED CODE - DO NOT MODIFY BY HAND

    part of 'sign_in_controller.dart';

    // **************************************************************************
    // RiverpodGenerator
    // **************************************************************************

    String _$signInControllerHash() => r'96996def040252f66d7e8dbf7b695aea45586848';

    /// See also [signInController].
    @ProviderFor(signInController)
    final signInControllerProvider = AutoDisposeProvider<SignInController>.internal(
    signInController,
    name: r'signInControllerProvider',
    debugGetCreateSourceHash: const bool.fromEnvironment('dart.vm.product')
        ? null
        : _$signInControllerHash,
    dependencies: null,
    allTransitiveDependencies: null,
    );

    typedef SignInControllerRef = AutoDisposeProviderRef<SignInController>;
    // ignore_for_file: type=lint
    // ignore_for_file: subtype_of_sealed_class, invalid_use_of_internal_member
    ```

## [syntax](https://docs-v2.riverpod.dev/docs/concepts/about_code_generation#the-syntax)

## 기타

- [How to use Notifier and AsyncNotifier with the new Flutter Riverpod Generator](https://codewithandrea.com/articles/flutter-riverpod-async-notifier/)
