# read and watch

- [read and watch](#read-and-watch)
    - [`Provider`와 `Provider.notifier`의 차이점](#provider와-providernotifier의-차이점)
        - [`Provider`](#provider)
        - [`Provider.notifier`](#providernotifier)
    - [`read`와 `watch`](#read와-watch)
    - [`read`](#read)
    - [`watch` 동작 방식](#watch-동작-방식)

## `Provider`와 `Provider.notifier`의 차이점

### `Provider`

- Riverpod의 Provider는 주어진 BuildContext에 대한 Provider의 현재 상태를 제공한다
- Provider는 값이 변경될 때마다 위젯을 다시 빌드한다

### `Provider.notifier`

- `Provider.notifier`는 값이 변경될 때마다 위젯을 다시 빌드하지 않는다. 따라서, **`Provider.notifier`를 사용하여 state를 변경하면, 이 변경 사항이 즉시 반영되지 않을 수 있다**
    - 즉, `read`를 사용하여 `Provider.notifier`를 가져와 `state`를 변경하려고 하면, 이 변경 사항이 즉시 반영되지 않을 수 있다
    - 반면 `watch`를 사용하여 `Provider.notifier`를 가져오면, `state`의 변경 사항이 즉시 반영되며, 이 변경 사항을 반영하여 위젯을 다시 빌드

## `read`와 `watch`

- `read`?
    - 프로바이더의 현재 값을 동기적으로 가져온다
- `watch`?
    - 프로바이더의 값이 **변경될 때마다 위젯을 다시 빌드**하도록 지시한다
- Riverpod는 상태 관리를 명확하고 예측 가능하게 만들려고 하며, `read`와 `watch`의 차이점은 이러한 목표를 지원한다

## `read`

- `read`는 프로바이더의 현재 값을 동기적으로 읽어온다
- `read`는 호출 시점의 프로바이더 값에 접근할 수 있으며, 이 값은 `read` 호출 이후에 변경되더라도 업데이트되지 않는다
- `read`는 프로바이더의 값이 변경될 때마다 새로운 값을 받아오기 위해 다시 호출해야 한다

## `watch` 동작 방식

- `watch`를 사용할 때, 해당 위젯이 화면에 존재하는 한, Provider의 상태가 변경될 때마다 리스닝이 발생(감지)하고 연결된 **위젯을 다시 빌드**
    - 따라서 GoRouter로 페이지를 이동하더라도, 이전 페이지의 위젯이 아직 메모리에 존재하고 완전히 파기되지 않았다면, 그 위젯 내의 watch는 여전히 상태 변화를 감지할 수 있다
    - 로그인 상태와 같은 **전역적인 상태**는 앱의 여러 부분에서 중요할 수 있다. 따라서 각 페이지에서 별도로 `watch`를 사용하여 상태를 감시하는 것이 일반적이며, 이를 통해 각 페이지가 독립적으로 상태 변화에 반응할 수 있으며, 더 명확하고 관리하기 쉬운 코드 구조가 된다.
- 리액티브하게 동작하여, 프로바이더의 값이 변경될 때마다 새로운 값을 받아올 수 있다
- `watch`는 프로바이더의 값을 관찰하고, 값이 변경될 때마다 자동으로 반응한다
