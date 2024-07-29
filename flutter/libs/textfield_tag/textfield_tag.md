# Textfield Tag

## 예제

```dart
import 'package:dilectio_app/src/subprojects/common/themes/app_colors.dart';
import 'package:dilectio_app/src/subprojects/common/themes/app_sizes.dart';
import 'package:dilectio_app/src/subprojects/common/widgets/inputs/gray_filled_text_form_field.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:textfield_tags/textfield_tags.dart';

class GreyFilledSpaceSeparatingTagsInput extends ConsumerStatefulWidget {
  final String? labelName;
  final TextStyle? labelTextStyle;
  final String? helperText;
  final String? hintText;

  const GreyFilledSpaceSeparatingTagsInput({
    super.key,
    this.labelName,
    this.labelTextStyle,
    this.helperText,
    this.hintText,
  });

  @override
  ConsumerState<ConsumerStatefulWidget> createState() =>
      GreyFilledSpaceSeparatingTagsState();
}

class GreyFilledSpaceSeparatingTagsState
    extends ConsumerState<GreyFilledSpaceSeparatingTagsInput> {
  final _controller = TextfieldTagsController();
  // double _distanceToField = 500.0;
  final List<String> _tags = [];

  @override
  void dispose() {
    super.dispose();
    _controller.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (widget.labelName != null && widget.labelName is String)
          Text(widget.labelName!, style: widget.labelTextStyle),
        gap8,
        TextFieldTags(
          textfieldTagsController: _controller,
          initialTags: const [],
          textSeparators: const [' '],
          letterCase: LetterCase.normal,
          inputfieldBuilder: (context, tec, fn, error, onChanged, onSubmitted) {
            return ((context, sc, tags, onTagDelete) {
              return TextField(
                controller: tec,
                onChanged: onChanged,
                onSubmitted: onSubmitted,
                decoration: InputDecoration(
                  // isDense: true,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(5.0),
                    borderSide: BorderSide.none,
                  ),
                  focusedBorder: const OutlineInputBorder(
                    borderSide: BorderSide(
                      color: primaryBackgroundColor,
                      width: 3.0,
                    ),
                  ),
                  filled: true, // `filled`가 true 값이어야 채워진다
                  fillColor: inputGrayColor,
                  helperText: widget.helperText,
                  helperStyle: const TextStyle(
                    color: primaryBackgroundColor,
                  ),
                  hintText: _controller.hasTags ? '' : widget.hintText,
                  errorText: error,
                  prefixIconConstraints: BoxConstraints(
                      maxWidth: MediaQuery.of(context).size.width * 0.74),
                  prefixIcon: tags.isNotEmpty
                      ? SingleChildScrollView(
                          controller: sc,
                          scrollDirection: Axis.horizontal,
                          child: Row(
                              children: tags.map((String tag) {
                            return Container(
                              decoration: const BoxDecoration(
                                borderRadius: BorderRadius.all(
                                  Radius.circular(20.0),
                                ),
                                color: primaryBackgroundColor,
                              ),
                              margin:
                                  const EdgeInsets.symmetric(horizontal: 5.0),
                              padding: const EdgeInsets.symmetric(
                                  horizontal: 10.0, vertical: 5.0),
                              child: Row(
                                mainAxisAlignment:
                                    MainAxisAlignment.spaceBetween,
                                children: [
                                  InkWell(
                                    child: Text(
                                      tag,
                                      style: const TextStyle(
                                        color: Colors.white,
                                      ),
                                    ),
                                    onTap: () {
                                      print("$tag selected");
                                    },
                                  ),
                                  const SizedBox(width: 4.0),
                                  InkWell(
                                    child: const Icon(
                                      Icons.cancel,
                                      size: 14.0,
                                      color: Color.fromARGB(255, 233, 233, 233),
                                    ),
                                    onTap: () {
                                      onTagDelete(tag);
                                    },
                                  )
                                ],
                              ),
                            );
                          }).toList()),
                        )
                      : null,
                ),
              );
            });
          },
        ),
      ],
    );
  }
}

```
