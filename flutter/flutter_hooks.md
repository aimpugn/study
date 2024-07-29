# Flutter Hooks

## `useEffect`: Flutter와 Riverpod 사용시 `build` 메서드 내에서 Future 실행

- Flutter와 Riverpod를 사용할 때, `build` 메서드 내에서 직접적으로 Future를 실행하는 것은 권장되지 않는다
- `build` 메서드는 Flutter 프레임워크에 의해 빈번하게 호출되며, 이 곳에서 Future를 실행하면 예상치 못한 상태 변화나 성능 문제를 야기할 수 있다
- ~~`build` 메서드 내에서 직접적으로 Future를 실행~~하기보다는 아래와 같은 방식이 Flutter 앱 개발에서의 일반적인 베스트 프랙티스
    1. 위젯의 생명주기 메서드
    2. 이벤트 핸들링
    3. 혹은 Riverpod의 `useEffect`와 같은 기능을 활용

### 일반적인 접근 방법

1. State Lifecycle Method 사용: `initState`, `didChangeDependencies`와 같은 생명주기 메서드를 사용하여 Future를 실행한다. 이러한 메서드는 위젯의 생명주기와 연결되어 있으며, 위젯이 생성될 때 한 번만 호출된다

   ```dart
   // `initState` 자체는 동기적인 메서드이므로, async와 await를 사용할 수 없다
   // 위젯의 생명주기 동안 한 번만 호출되며, 위젯이 렌더링되기 전에 완료되어야 하는 초기 설정을 위한 곳이기 때문
   @override
   void initState() {
     super.initState();
     // 근데 initState에서 Riverpod의 프로바이더를 변경하려고 하면 안된다
     // Flutter의 위젯 라이프사이클 중 `build` 메서드 실행 중에는 프로바이더를 변경하는 것이 허용되지 않는다.
     // 이는 위젯 트리 구축 중에 상태를 변경하면 UI의 일관성을 해칠 수 있기 때문
     ref
       .read(userApplicationProvider.notifier)
       .updateLatestUser('Dzc0BpBCmQa74chhZix25HWTDII2');
   }
   ```

2. Effect 사용: Riverpod의 `useEffect` 혹은 비슷한 기능을 제공하는 다른 패키지를 사용하여 사이드 이펙트(side-effect)를 관리한다
3. Event Handling: 사용자 상호작용이나 특정 이벤트를 통해 Future를 실행한다. 예를 들어, 버튼 클릭 이벤트 핸들러에서 Future를 실행할 수 있다.

### `ConsumerWidget`에서의 접근 방법

`ConsumerWidget`을 사용하는 경우, 위젯의 상태를 감시하고 필요에 따라 Future를 실행하는 것이 좋습니다.

```dart
class MyWidget extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userState = ref.watch(userApplicationProvider);

    useEffect(() {
      ref
        .read(userApplicationProvider.notifier)
        .updateLatestUser('Dzc0BpBCmQa74chhZix25HWTDII2');
      return null; // useEffect clean-up function
    }, const []);

    // UI 구성
  }
}
```

`useEffect`는 컴포넌트가 마운트될 때 한 번만 사이드 이펙트를 실행하도록 보장한다. 이 방법은 상태가 변경될 때마다 Future를 실행하지 않고, 필요할 때만 실행되도록 관리한다.
