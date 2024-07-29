# Assets & media

- [Assets \& media](#assets--media)
    - [Assets \& media](#assets--media-1)
        - [assets 지정하기](#assets-지정하기)
    - [Updating the app icon](#updating-the-app-icon)
        - [Android](#android)
        - [iOS](#ios)
    - [Updating the launch screen](#updating-the-launch-screen)

## [Assets & media](https://docs.flutter.dev/ui/assets/assets-and-images)

### assets 지정하기

- 디렉토리에 직접 위치한(located directly) 파일들만 포함된다
- 특정 파일 포함하기

    ```yaml
    flutter:
        assets:
            - assets/my_icon.png
            - assets/background.png
    ```

- 디렉토리 하위의 모든 assets 포함하기

    ```yaml
    flutter:
        assets:
            - directory/
            - directory/subdirectory/
    ```

    - 하위 디렉터리에 있는 파일을 추가하려면 디렉터리별로 항목을 만든다

## [Updating the app icon](https://docs.flutter.dev/ui/assets/assets-and-images#updating-the-app-icon)

### Android

- In your Flutter project’s root directory, navigate to `.../android/app/src/main/res`.
- The various bitmap resource folders such as `mipmap-hdpi` already contain placeholder images named `ic_launcher.png`.
- Replace them with your desired assets respecting the recommended icon size per screen density as indicated by the [Android Developer Guide](https://developer.android.com/training/multiscreen/screendensities).
- 만일 `.png` 파일명을 바꾼다면, `AndroidManifest.xml`의 `<application>` 태그의 `android:icon` 속성에서 해당 이름도 변경해야 한다

### iOS

- In your Flutter project’s root directory, navigate to `.../ios/Runner`.
- The A`ssets.xcassets/AppIcon.appiconset` directory already contains placeholder images.
- Replace them with the appropriately sized images as indicated by their filename as dictated by the [Apple Human Interface Guidelines](https://developer.apple.com/design/human-interface-guidelines/app-icons). **Keep the original file names**.

## [Updating the launch screen](https://docs.flutter.dev/ui/assets/assets-and-images#updating-the-launch-screen)

- Flutter also uses native platform mechanisms to draw transitional launch screens to your Flutter app while the Flutter framework loads.
- This launch screen persists until Flutter renders the first frame of your application.
- Note: This implies that
    - if you don’t call [`runApp()`](https://api.flutter.dev/flutter/widgets/runApp.html) in the `main()` function of your app
    - or more specifically, if you don’t call [`FlutterView.render()`](https://api.flutter.dev/flutter/dart-ui/FlutterView/render.html) in response to [`PlatformDispatcher.onDrawFrame`](https://api.flutter.dev/flutter/dart-ui/PlatformDispatcher/onDrawFrame.html),
    - the launch screen persists forever.
