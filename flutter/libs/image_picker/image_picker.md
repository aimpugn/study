# Image Picker

- [Image Picker](#image-picker)
    - [Image Picker plugin for Flutter](#image-picker-plugin-for-flutter)
        - [iOS](#ios)
        - [Android](#android)
            - [Handling MainActivity destruction](#handling-mainactivity-destruction)
    - [Permanently storing images and videos](#permanently-storing-images-and-videos)
    - [Using `launchMode: singleInstance`](#using-launchmode-singleinstance)
    - [Example](#example)
    - [for web](#for-web)
        - [`XFile`](#xfile)
        - [cross\_file](#cross_file)
        - [Web Limitations](#web-limitations)
        - [Usage](#usage)
            - [Import the package](#import-the-package)
            - [Use the plugin](#use-the-plugin)

## [Image Picker plugin for Flutter](https://pub.dev/packages/image_picker#image-picker-plugin-for-flutter)

### [iOS](https://pub.dev/packages/image_picker#ios)

- Starting with version 0.8.1 the iOS implementation uses PHPicker to pick (multiple) images on iOS 14 or higher.

### [Android](https://pub.dev/packages/image_picker#android)

- Starting with version 0.8.1 the Android implementation support to pick (multiple) images on Android 4.3 or higher.
- No configuration required
    - the plugin should work out of the box. It is however highly recommended to prepare for Android killing the application when low on memory.
    - How to prepare for this is discussed in the Handling MainActivity destruction on Android section.
- It is no longer required to add `android:requestLegacyExternalStorage="true"` as an attribute to the `<application>` tag in `AndroidManifest.xml`, as `image_picker` has been updated to make use of **scoped storage**.

#### Handling MainActivity destruction

- When under high memory pressure the Android system may kill the MainActivity of the application using the `image_picker`.
- On Android the `image_picker` makes use of the default `Intent.ACTION_GET_CONTENT` or `MediaStore.ACTION_IMAGE_CAPTURE` intents.
- This means that while the intent is executing the source application is moved to the background and becomes eligible for cleanup when the system is low on memory.
- When the intent finishes executing, Android will restart the application.
- Since the data is never returned to the original call use the `ImagePicker.retrieveLostData()` method to retrieve the lost data.

```dart
Future<void> getLostData() async {
  final ImagePicker picker = ImagePicker();
  final LostDataResponse response = await picker.retrieveLostData();
  if (response.isEmpty) {
    return;
  }
  final List<XFile>? files = response.files;
  if (files != null) {
    _handleLostFiles(files);
  } else {
    _handleError(response.exception);
  }
}
```

## Permanently storing images and videos

- Images and videos picked using the camera
    - are saved to your **application's local cache**,
    - and should therefore be expected to **only be around temporarily**.
- If you require your picked image to be stored permanently, it is your responsibility to move it to a more permanent location.

## Using `launchMode: singleInstance`

- Launching the image picker from an Activity with `launchMode: singleInstance` will always return `RESULT_CANCELED`.
- In this launch mode, **new activities are created in a separate Task**.
- As activities cannot communicate between tasks, the image picker activity cannot send back its eventual result to the calling activity.
- To work around this problem, consider using `launchMode: singleTask` instead.

## Example

```dart
final ImagePicker picker = ImagePicker();
```

```dart
// Pick an image.
final XFile? image = await picker.pickImage(source: ImageSource.gallery);
```

```dart
// Capture a photo.
final XFile? photo = await picker.pickImage(source: ImageSource.camera);
```

```dart
// Pick a video.
final XFile? galleryVideo = await picker.pickVideo(source: ImageSource.gallery);
```

```dart
// Capture a video.
final XFile? cameraVideo = await picker.pickVideo(source: ImageSource.camera);
```

```dart
// Pick multiple images.
final List<XFile> images = await picker.pickMultiImage();
```

```dart
// Pick singe image or video.
final XFile? media = await picker.pickMedia();
```

```dart
// Pick multiple images and videos.
final List<XFile> medias = await picker.pickMultipleMedia();
```

## [for web](https://pub.dev/packages/image_picker_for_web#limitations-on-the-web-platform)

### `XFile`

- This plugin uses `XFile` objects to abstract files picked/created by the user.
- Read more about `XFile` on the web in [`package:cross_file`'s README](https://pub.dev/packages/cross_file).

### cross_file

- An abstraction to allow working with files across multiple platforms
- Import `package:cross_file/cross_file.dart`,
    - **instantiate a `XFile` using a path or byte array**
    - and use its methods and properties to access the file and its metadata.

```dart
import 'package:cross_file/cross_file.dart';

final file = XFile('assets/hello.txt');

print('File information:');
print('- Path: ${file.path}');
print('- Name: ${file.name}');
print('- MIME type: ${file.mimeType}');

final fileContent = await file.readAsString();
print('Content of the file: ${fileContent}');  // e.g. "Moto G (4)"
```

### Web Limitations

- `XFile` on the web platform is backed by [`Blob`](https://api.dart.dev/be/180361/dart-html/Blob-class.html) objects and their URLs.
- It seems that Safari hangs when reading Blobs larger than 4GB (your app will stop without returning any data, or throwing an exception).
- This package will attempt to throw an `Exception` before a large file is accessed from Safari (if its size is known beforehand), so that case can be handled programmatically.

### Usage

#### Import the package

- This package is [endorsed](https://flutter.dev/docs/development/packages-and-plugins/developing-packages#endorsed-federated-plugin), which means
    - you can simply use `image_picker` normally.
    - This package will be automatically included in your app when you do, so **you do not need to add it to your `pubspec.yaml`.**
- However, if you import this package to use any of its APIs directly, you should add it to your `pubspec.yaml` as usual.

#### Use the plugin

- You should be able to use `package:image_picker` almost as normal.
- Once the user has picked a file, the returned `XFile` instance will contain a **network-accessible Blob URL** (pointing to a location within the browser).
- The instance will also let you retrieve the bytes of the selected file across all platforms.
- If you want to use the path directly, your code would need look like this:

```dart
...
if (kIsWeb) {
  Image.network(pickedFile.path);
} else {
  Image.file(File(pickedFile.path));
}
...
```

Or, using bytes

```dart
...
Image.memory(await pickedFile.readAsBytes())
...
```
