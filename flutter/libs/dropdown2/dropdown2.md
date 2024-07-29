# [dropdown2](https://pub.dev/packages/dropdown_button2)

- [dropdown2](#dropdown2)
    - [example](#example)

## example

```dart
Row(
    children: [
        DropdownButtonHideUnderline(
            child: DropdownButton2(
                value: l10n.labelDomestic,
                items: [l10n.labelDomestic, l10n.labelForeign]
                    .map((String value) {
                return DropdownMenuItem<String>(
                    value: value,
                    child: Row(
                        children: [
                        Text(
                            value,
                            style: const TextStyle(
                            color: Colors.black,
                            ),
                        ),
                        ],
                    ));
                }).toList(),
                onChanged: _onCitizenshipChanged,
                dropdownStyleData: DropdownStyleData(
                padding: const EdgeInsets.symmetric(vertical: 6),
                decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(4),
                    color: Colors.white,
                ),
                ),
            ),
        ),
    ]
)
```
