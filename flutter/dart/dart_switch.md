# dart switch

- [dart switch](#dart-switch)
    - [pattern type](#pattern-type)
        - [Object](#object)

## [pattern type](https://dart.dev/language/pattern-types)

### Object

- `switch ~ case`에서 클래스를 매칭할 수 있다

```dart
  TaskEither<ApplicationError, JWTApplicationSuccess> getJWT(String email, String password) =>
      TaskEither<ApplicationError, JWTApplicationSuccess>.Do((_) async {
        final appUserState =
            await getDefaultUserInfo(email, password).run(userRepository);

        return switch (appUserState) {
          AppUserError() =>
            _(TaskEither.left(const ApplicationError.notFound(""))),
          AppUserSuccess(appUser: AppUser appUser) =>
            _(TaskEither.right(JWTApplicationSuccess(JWT(
              {
                'id': appUser.defaultInfo.uid,
                'email': appUser.defaultInfo.email,
              },
              issuer: 'go_together',
            ))))
        };
      });
```
