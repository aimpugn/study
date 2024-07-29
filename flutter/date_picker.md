# date picker

## 캘린더 예시

```dart
Future<void> _selectDate(BuildContext context) async {
  final DateTime? picked = await showDatePicker(
    context: context,
    initialDate: DateTime.now(),
    firstDate: DateTime(2000),
    lastDate: DateTime(2025),
  );
  if (picked != null) {
    // Handle the selected date
  }
}

// 호출 예시:
// ElevatedButton(
//   onPressed: () => _selectDate(context),
//   child: Text('Select date'),
// )
```
