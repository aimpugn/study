# Image Picker Pick

## 이미지 선택하기

### image picker web 사용

```dart
// web_image_picker.dart
import 'dart:async';
import 'dart:html' as html;
import 'package:flutter/src/widgets/framework.dart';
import 'package:flutter/widgets.dart';
import 'package:image_picker/image_picker.dart';
import 'image_picker_service.dart';

class WebImagePickerService implements ImagePickerService {
  @override
  ImagePicker picker = ImagePicker();

  @override
  Widget getImageWidget(String imagePath) => Image.network(imagePath);

  @override
  Future<XFile?> pickImage() async =>
      await picker.pickImage(source: ImageSource.gallery);
}
```

그리고 이를 화면쪽에서 다음과 같이 사용하면

```dart
  Future<void> _pickImage() async {
    final pickedFile = await pickerService.pickImage();
    if (pickedFile != null) {
      print(pickedFile.path);
      // blob:http://localhost:57167/6460f49e-46cb-4501-9afd-a130cf69b5a7
      // 이렇게 임시 path가 나온다
      setState(() {
        _images.add(pickedFile);
      });
    }
  }
```

### 직접 핸들링하기

```dart
  Future pickImageFromGallery() async {
    // WEB 
    if (kIsWeb) {
      final uploadInput = html.FileUploadInputElement();
      uploadInput.click();
      // Listen for file selection
      uploadInput.onChange.listen((e) {
        final files = uploadInput.files;
        if (files != null && files.isNotEmpty) {
          final html.File file = files.first;
          final url = html.Url.createObjectUrl(file);
          widget.onPhotoChanged(file);

          final reader = html.FileReader();
          reader.readAsDataUrl(file);
          reader.onLoadEnd.listen((event) {
            setState(() {
              _imagePath = reader.result as String;
            });
          });
        }
      });
      return;
    }

    // Else
    XFile? pickedFile = await picker.pickImage(source: ImageSource.gallery);

    if (pickedFile != null) {
      widget.onPhotoChanged(pickedFile);

      setState(() {
        _imagePath = pickedFile.path;
      });
    }
  }
```

```dart
  @override
  Future<XFile?> pickImage() async {
    // Web 플랫폼에서 이미지 선택 로직 구현
    final completer = Completer<XFile>();
    final input = html.FileUploadInputElement()..accept = 'image/*';
    input.click();

    input.onChange.listen((event) {
      final files = input.files;
      if (files != null && files.isNotEmpty) {
        final file = files.first;
        final path = file.relativePath;
        if (path != null) {
          // 파일 처리 및 XFile 객체 생성
          completer.complete(XFile(path));
        }
      }
    });

    return completer.future;
  }
```
