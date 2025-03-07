# Image Picker Widget

## 위젯으로 보여주기

```dart
class ImagePortrait extends StatelessWidget {
  final double height;
  final String imagePath;
  final ImageType imageType;

  const ImagePortrait({
    super.key,
    required this.imageType,
    required this.imagePath,
    this.height = 250.0,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: height * 0.65,
      height: MediaQuery.of(context).size.height / 3,
      decoration: BoxDecoration(
          border: Border.all(width: 2, color: kAccentColor),
          borderRadius: const BorderRadius.all(Radius.circular(25.0))),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(22.0),
        child: getImage(),
      ),
    );
  }

  Widget? getImage() {
    if (imageType == ImageType.NONE || imagePath == null) return null;
    if (imageType == ImageType.FILE_IMAGE) {
      return Image.file(File(imagePath), fit: BoxFit.fitHeight);
    } else if (imageType == ImageType.ASSET_IMAGE) {
      return Image.asset(imagePath, fit: BoxFit.fitHeight);
    } else if (imageType == ImageType.HTML_IMAGE) {
      Uint8List image = base64Decode(
          imagePath.replaceFirst("data:image/jpeg;base64,", "", 0));
      return Image.memory(image, fit: BoxFit.fitHeight);
    } else {
      return null;
    }
  }
}
```
