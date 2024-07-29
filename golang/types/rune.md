# rune

- [rune](#rune)
    - [rune, unicode, unicode point, utf-8](#rune-unicode-unicode-point-utf-8)

## rune, unicode, unicode point, utf-8

UTF-8, Unicode, Unicode Code Point, 그리고 Go 언어의 `rune`은 문자 인코딩과 문자 표현에 관련된 개념들입니다. 이들의 관계를 이해하려면 각 용어의 의미와 용도를 알아야 합니다.

1. Unicode:
   - Unicode는 전 세계의 모든 문자 시스템을 포괄하는 표준 문자 집합입니다. 각 문자에는 고유한 식별 번호가 할당되어 있으며, 이를 '코드 포인트'라고 합니다.
   - Unicode는 문자를 추상적인 방식으로 정의하며, 실제 파일이나 메모리에 어떻게 저장될지는 결정하지 않습니다. 이는 인코딩 방식에 따라 달라집니다.

2. Unicode Code Point:
   - Unicode Code Point는 Unicode 문자 집합에서 각 문자에 할당된 고유한 번호입니다.
   - 이 번호는 일반적으로 *U+XXXX*의 형태로 표현되며, 여기서 XXXX는 16진수 값입니다. 예를 들어, 영문 대문자 'A'의 Unicode Code Point는 U+0041입니다.

3. UTF-8:
   - UTF-8은 Unicode 문자 집합을 인코딩하는 방법 중 하나입니다. 이 인코딩 방식은 가변 길이 문자 인코딩 방식으로, 각 Unicode Code Point를 1바이트에서 4바이트까지의 길이로 인코딩합니다.
   - UTF-8의 중요한 특징 중 하나는 ASCII 문자에 대해 하위 호환성을 제공한다는 점입니다. ASCII 문자는 UTF-8에서도 동일한 1바이트로 표현됩니다.

4. Rune:
   - Go 언어에서 `rune`은 `int32`의 별칭입니다. 이는 Unicode Code Point를 나타내는 데 사용됩니다.
   - Go 문자열은 기본적으로 UTF-8로 인코딩되어 있습니다. `rune`을 사용하여 문자열을 순회하면, 각 문자(Unicode 문자)에 대한 Code Point를 얻을 수 있습니다.
   - `rune`을 사용하면 UTF-8 인코딩된 문자열에서 각 문자를 적절히 해석하고 처리할 수 있습니다. 예를 들어, UTF-8 인코딩에서 한 문자가 여러 바이트로 표현될 수 있으므로, 단순히 바이트 단위로 문자열을 처리하면 잘못된 결과를 얻을 수 있습니다. `rune`을 사용하면 이런 문제를 방지할 수 있습니다.

5. ASCII
    - ASCII는 128개의 문자를 나타내기 위한 7비트 코드 체계입니다.
    - ASCII 코드는 기본 영문자, 숫자, 특수 문자 및 제어 문자를 포함합니다.
    - 모든 ASCII 문자는 Unicode에서 동일한 코드 포인트를 갖습니다. 예를 들어, ASCII의 'A'(65)는 Unicode에서 U+0041입니다.
