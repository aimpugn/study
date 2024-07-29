# fpdart Do notation

- [fpdart Do notation](#fpdart-do-notation)
    - [typical chain of methods -\> Do notation](#typical-chain-of-methods---do-notation)

## typical chain of methods -> Do notation

> Note:
>
> We recommend using the Do notation whenever possible to improve the legibility of your code 🤝

```dart
/// Without the Do notation
String goShopping() => goToShoppingCenter()
    .alt(goToLocalMarket)
    .flatMap(
      (market) => market.buyBanana().flatMap(
            (banana) => market.buyApple().flatMap(
                  (apple) => market.buyPear().flatMap(
                        (pear) => Option.of('Shopping: $banana, $apple, $pear'),
                      ),
                ),
          ),
    )
    .getOrElse(
      () => 'I did not find 🍌 or 🍎 or 🍐, so I did not buy anything 🤷‍♂️',
    );
```

```dart
/// Using the Do notation
String goShoppingDo() => Option.Do(
      (_) {
        final market = _(goToShoppingCenter().alt(goToLocalMarket));
        final amount = _(market.buyAmount());

        final banana = _(market.buyBanana());
        final apple = _(market.buyApple());
        final pear = _(market.buyPear());

        return 'Shopping: $banana, $apple, $pear';
      },
    ).getOrElse(
      () => 'I did not find 🍌 or 🍎 or 🍐, so I did not buy anything 🤷‍♂️',
    );
```

- Do 표기법은 `Do()` 생성자를 사용해서 초기화할 수 있다
- `_` 함수에 액세스할 수 있고, `flatMap` 없이 각`Option` 안의 값을 추출하여 사용할 수 있다.
