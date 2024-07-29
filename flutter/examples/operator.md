# operator

- [operator](#operator)
    - [@override `===`](#override-)

## @override `===`

```dart
@override
bool operator ==(dynamic other) {
return identical(this, other) ||
    (other.runtimeType == runtimeType &&
        other is _$_DefaultInfo &&
        (identical(other.uid, uid) || other.uid == uid) &&
        (identical(other.email, email) || other.email == email) &&
        (identical(other.emailVerified, emailVerified) ||
            other.emailVerified == emailVerified) &&
        (identical(other.name, name) || other.name == name) &&
        (identical(other.phone, phone) || other.phone == phone) &&
        (identical(other.phoneVerified, phoneVerified) ||
            other.phoneVerified == phoneVerified) &&
        (identical(other.profileUris, profileUris) ||
            other.profileUris == profileUris));
}
```
