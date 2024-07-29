# Image class

- [Image class](#image-class)
    - [Image classes](#image-classes)
        - [`Image`](#image)
        - [`Image.new`](#imagenew)
        - [`Image.asset`](#imageasset)
        - [`Image.network`](#imagenetwork)
        - [`Image.file`](#imagefile)
        - [`Image.memory`](#imagememory)
    - [Memory usage](#memory-usage)
    - [Web considerations](#web-considerations)
    - [See also](#see-also)
        - [Cookbook: Display images from the internet](#cookbook-display-images-from-the-internet)
            - [Bonus: animated gifs](#bonus-animated-gifs)
        - [Cookbook: Work with cached images](#cookbook-work-with-cached-images)
        - [Cookbook: Fade in images with a placeholder](#cookbook-fade-in-images-with-a-placeholder)
            - [In Memory Example](#in-memory-example)
            - [From Asset Example](#from-asset-example)

## Image classes

### `Image`

- The default constructor can be used with any `ImageProvider`, such as a `NetworkImage`, to display an image from the internet.

```dart
const Image(
  image: NetworkImage('https://flutter.github.io/assets-for-api-docs/assets/widgets/owl.jpg'),
)
```

### `Image.new`

- for obtaining an image from an [`ImageProvider`](https://api.flutter.dev/flutter/painting/ImageProvider-class.html).

### `Image.asset`

- for obtaining an image from an [`AssetBundle`](https://api.flutter.dev/flutter/services/AssetBundle-class.html) using a key.

### [`Image.network`](https://api.flutter.dev/flutter/widgets/Image/Image.network.html)

- to display an image from the internet.
- for obtaining an image from a URL.

```dart
Image.network('https://flutter.github.io/assets-for-api-docs/assets/widgets/owl-2.jpg')
```

### `Image.file`

- for obtaining an image from a [`File`](https://api.flutter.dev/flutter/dart-io/File-class.html).

### `Image.memory`

- for obtaining an image from a [`Uint8List`](https://api.flutter.dev/flutter/dart-typed_data/Uint8List-class.html).

## Memory usage

- The image is stored in memory in uncompressed form (so that it can be rendered).
    - Large images will use a lot of memory
    - ex: a 4K image (3840×2160) will use over 30MB of RAM (assuming 32 bits per pixel).
- This problem is exacerbated by the images being cached in the `ImageCache`, so large images can use memory for even longer than they are displayed.
- The `Image.asset`, `Image.network`, `Image.file`, and `Image.memory` constructors allow
    - a custom decode size to be specified through `cacheWidth` and `cacheHeight` parameters.
    - The engine will then decode and store the image at the specified size, instead of the image's natural size.
- This can significantly reduce the memory usage. For example,
    - a 4K image that will be rendered at only 384×216 pixels (one-tenth the horizontal and vertical dimensions) would only use 330KB if those dimensions are specified using the `cacheWidth` and `cacheHeight` parameters
    - a 100-fold reduction in memory usage.

## Web considerations

- In the case where a network image is used on the Web platform, the `cacheWidth` and `cacheHeight` parameters are **only supported when the application is running with the `CanvasKit renderer`**.
- When the application is using the `HTML renderer`, the web engine delegates image decoding of network images to the Web, which does not support custom decode sizes.

## See also

- [`Icon`](https://api.flutter.dev/flutter/widgets/Icon-class.html), which shows an image from a font.
- [`Ink.image`](https://api.flutter.dev/flutter/material/Ink/Ink.image.html),
    - which is the preferred way to show an image in a material application
    - (especially if the image is in a [`Material`](https://api.flutter.dev/flutter/material/Material-class.html) and will have an [`InkWell`](https://api.flutter.dev/flutter/material/InkWell-class.html) on top of it).
- [`Image`](https://api.flutter.dev/flutter/dart-ui/Image-class.html), the class in the `dart:ui` library.

### [Cookbook: Display images from the internet](https://flutter.dev/docs/cookbook/images/network-image)

```dart
import 'package:flutter/material.dart';

void main() => runApp(const MyApp());

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    var title = 'Web Images';

    return MaterialApp(
      title: title,
      home: Scaffold(
        appBar: AppBar(
          title: Text(title),
        ),
        body: Image.network('https://picsum.photos/250?image=9'),
      ),
    );
  }
}
```

#### Bonus: animated gifs

- 애니메이션 GIF를 지원

```dart
Image.network('https://docs.flutter.dev/assets/images/dash/dash-fainting.gif');
```

### [Cookbook: Work with cached images](https://flutter.dev/docs/cookbook/images/cached-images)

- [`cached_network_image`](https://pub.dev/packages/cached_network_image) package 사용해야 한다
- 플레이스홀더와 이미지가 로드될 때 페이드 인되는 이미지를 사용할 수 있다

```dart
import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    const title = 'Cached Images';

    return MaterialApp(
      title: title,
      home: Scaffold(
        appBar: AppBar(
          title: const Text(title),
        ),
        body: Center(
          child: CachedNetworkImage(
            /// 어떤 위젯이든 `placeholder`로 사용할 수 있다
            placeholder: (context, url) => const CircularProgressIndicator(), // spinner while the image loads
            imageUrl: 'https://picsum.photos/250?image=9',
          ),
        ),
      ),
    );
  }
}
```

### [Cookbook: Fade in images with a placeholder](https://flutter.dev/docs/cookbook/images/fading-in-images)

- 기본 이미지 위젯을 사용하여 이미지를 표시할 때 이미지가 로드되는 즉시 화면에 튀어나오는 것을 볼 수 있는데, 이는 사용자에게 시각적으로 어색하게 느껴질(feel visually jarring) 수 있다
- 대신 처음에는 플레이스홀더를 표시하고 이미지가 로드될 때 페이드 인하면 좋지 않을까? 바로 이런 용도로 `FadeInImage` 위젯을 사용
- `FadeInImage`는 어떤 이미지 유형과도 함께 쓸 수 있다
    - in-memory
    - local assets
    - or images from the internet.

#### In Memory Example

```dart
import 'package:flutter/material.dart';
import 'package:transparent_image/transparent_image.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    const title = 'Fade in images';

    return MaterialApp(
      title: title,
      home: Scaffold(
        appBar: AppBar(
          title: const Text(title),
        ),
        body: Stack(
          children: <Widget>[
            const Center(child: CircularProgressIndicator()),
            Center(
              child: FadeInImage.memoryNetwork(
                placeholder: kTransparentImage,
                image: 'https://picsum.photos/250?image=9',
              ),
            ),
          ],
        ),
      ),
    );
  }
}
```

#### From Asset Example

```yaml
# pubspec.yaml 
 flutter:
   assets:
+    - assets/loading.gif
```

```dart
import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    const title = 'Fade in images';

    return MaterialApp(
      title: title,
      home: Scaffold(
        appBar: AppBar(
          title: const Text(title),
        ),
        body: Center(
          child: FadeInImage.assetNetwork(
            placeholder: 'assets/loading.gif',
            image: 'https://picsum.photos/250?image=9',
          ),
        ),
      ),
    );
  }
}
```
