# tag

## 매력 포인트 태그 입력 위젯

```dart
class CharmPointTags extends StatefulWidget {
  @override
  _CharmPointTagsState createState() => _CharmPointTagsState();
}

class _CharmPointTagsState extends State<CharmPointTags> {
  final TextEditingController _controller = TextEditingController();
  List<String> _tags = [];

  void _addTag(String tag) {
    if (tag.isNotEmpty && !_tags.contains(tag)) {
      setState(() {
        _tags.add(tag);
      });
      _controller.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TextField(
          controller: _controller,
          decoration: InputDecoration(hintText: 'Enter charm points'),
          onSubmitted: _addTag,
        ),
        Wrap(
          spacing: 8,
          children: _tags.map((tag) => Chip(
            label: Text(tag),
            onDeleted: () {
              setState(() {
                _tags.remove(tag);
              });
            },
          )).toList(),
        ),
      ],
    );
  }
}
```
