# platform

## 플랫폼별 디자인

플랫폼별로 다른 디자인을 적용하고 싶다면, Flutter의 Platform-aware 기능을 사용하여 OS에 따라 다른 스타일을 적용할 수 있다

```dart
//  현재 플랫폼을 확인하고, 그에 맞는 스타일을 적용할 수 있다
Theme.of(context).platform;
```

- Material Design
    - 주로 Android 앱에서 사용
    - Flutter는 Material 위젯을 제공하여 Android 스타일을 손쉽게 구현할 수 있도록 해준다
- Cupertino Design
    - iOS 스타일을 구현하기 위한 위젯들을 제공
    - iOS 사용자에게 익숙한 인터페이스를 만들 수 있다
- Responsive Design
    - 웹이나 데스크톱 애플리케이션을 고려할 때는 반응형 디자인을 적용하는 것이 중요하다
