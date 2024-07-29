# Flutter Version Management

## [fvm](https://fvm.app/)

- Multiple Flutter SDKs
- Project Versioning
- Advanced Tooling

## [overview](https://fvm.app/docs/getting_started/overview/)

## [설치](https://fvm.app/docs/getting_started/installation)

### MacOS

```bash
# install
brew tap leoafarias/fvm
brew install fvm

# uninstall
brew uninstall fvm
brew untap leoafarias/fvm
```

## 사용

## 전역 설정

```bash
# Set beta channel as global
fvm global beta

# Check version
flutter --version # Will be beta release

# Set stable channel as global
fvm global stable

# Check version
flutter --version # Will be stable release
```

```bash
╰─ fvm global 3.3.0          
Flutter "3.3.0" is not installed.
Would you like to install it? Y/n: y
```
