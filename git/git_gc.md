# Git GC

## Git GC (Garbage Collection)

`git gc` 명령어는 Git 저장소를 최적화하고 정리하는 데 사용되는 명령어입니다. "GC"는 "Garbage Collection"의 약자입니다. 이 명령어는 Git 저장소에서 사용되지 않는 객체를 제거하고, 저장소의 크기를 줄이며, 데이터 접근 성능을 향상시키기 위한 다양한 작업을 수행합니다.

`git gc` 명령어는 Git 저장소를 최적화하고, 사용하지 않는 객체를 제거하며, 저장소를 압축하여 공간을 절약하는 작업을 수행합니다. `gc`는 "Garbage Collection"의 약자입니다. 위 출력 결과는 `git gc` 명령어가 실행된 후의 상태를 보여줍니다. 각 단계의 의미를 자세히 설명하겠습니다.

`git gc` 명령어는 Git 저장소를 최적화하는 데 중요한 역할을 합니다. 저장소 내 모든 객체를 열거하고, 델타 압축을 수행하며, 최종적으로 압축된 객체들을 디스크에 기록합니다. 이를 통해 저장소의 크기를 줄이고, 불필요한 데이터를 제거하여 저장소의 성능과 효율성을 높일 수 있습니다.

`git gc` 명령어는 Git 저장소의 유지 보수와 최적화를 위해 매우 중요한 도구입니다. 정기적으로 실행하여 저장소의 크기를 줄이고 성능을 최적화하는 것이 좋습니다. 특히 큰 저장소나 오래된 저장소에서는 `git gc`를 통해 저장소를 정리하고 최적화하는 것이 필수적입니다.

## 주요 기능

1. **객체 압축**:
   - Git은 파일의 변경 이력과 메타데이터를 객체 형태로 저장합니다.
   - `git gc`는 이러한 객체들을 델타 압축하여 디스크 공간을 절약합니다. 델타 압축은 객체 간의 차이만을 저장하여 중복 데이터를 최소화합니다.

2. **패킹**:
   - Git 객체는 `.git/objects` 디렉토리에 저장됩니다. 많은 객체가 개별 파일로 저장되면 파일 시스템의 성능이 저하될 수 있습니다.
   - `git gc`는 여러 객체를 하나의 패키지 파일로 묶어 저장합니다. 이를 "packfile"이라고 합니다. 패킹을 통해 저장소 크기를 줄이고 파일 시스템의 I/O 성능을 향상시킵니다.

3. **사용되지 않는 객체 제거**:
   - 브랜치가 삭제되거나 리베이스 등의 작업으로 인해 사용되지 않는 객체가 남을 수 있습니다.
   - `git gc`는 이러한 사용되지 않는 객체를 제거하여 저장소를 정리합니다.

4. **리플로그 정리**:
   - Git은 브랜치와 기타 참조의 이동 기록을 리플로그에 저장합니다.
   - `git gc`는 오래된 리플로그 항목을 제거하여 리플로그를 정리합니다.

## 명령어 옵션

- **`git gc --auto`**:
    - 자동으로 가비지 컬렉션을 실행합니다. Git은 내부적으로 이 명령어를 사용하여 자동으로 가비지 컬렉션을 수행합니다.
    - 저장소의 크기나 특정 작업 후 일정 조건이 충족되면 실행됩니다.

- **`git gc --aggressive`**:
    - 더욱 철저하게 가비지 컬렉션을 수행합니다.
    - 더 많은 시간을 소요하지만, 저장소의 크기를 최대한 줄이고 성능을 최적화합니다.

- **`git gc --prune=<date>`**:
    - 지정된 날짜 이전의 모든 사용되지 않는 객체를 제거합니다.
    - 기본값은 `2주`입니다.

### 명령어 예제

1. **기본 가비지 컬렉션 실행**:

   ```bash
   git gc
   ```

2. **자동 가비지 컬렉션 실행**:

   ```bash
   git gc --auto
   ```

3. **공격적 가비지 컬렉션 실행**:

   ```bash
   git gc --aggressive
   ```

4. **지정된 날짜 이전의 객체 제거**:

   ```bash
   git gc --prune="30 days ago"
   ```

## `git gc` 명령어의 주요 기능

1. **객체 압축**:
   - Git은 파일 변경 이력을 저장하기 위해 많은 객체를 생성합니다. `git gc`는 이 객체들을 압축하여 디스크 공간을 절약합니다.
   - 델타 압축을 통해 유사한 객체들 간의 차이점만 저장함으로써 더 많은 공간을 절약할 수 있습니다.

2. **사용되지 않는 객체 제거**:
   - 삭제된 브랜치나 태그로 인해 남아 있는 사용되지 않는 객체들을 제거합니다.
   - 이를 통해 저장소의 크기를 줄이고, 불필요한 데이터를 제거합니다.

3. **패킹**:
   - 객체들을 하나의 큰 패키지 파일로 묶습니다. 이를 통해 파일 시스템의 I/O 성능을 향상시키고, 저장소의 크기를 줄일 수 있습니다.

## `git gc` 출력 예제

```sh
❯ git gc
Enumerating objects: 552, done.
Counting objects: 100% (552/552), done.
Delta compression using up to 10 threads
Compressing objects: 100% (500/500), done.
Writing objects: 100% (552/552), done.
Total 552 (delta 26), reused 6 (delta 0), pack-reused 0
```

1. **Enumerating objects: 552, done.**
   - Git이 저장소 내 모든 객체를 열거하고 있는 단계입니다.
   - 552개의 객체가 열거되었음을 나타냅니다.

2. **Counting objects: 100% (552/552), done.**
   - 모든 객체를 세는 과정입니다.
   - 552개의 객체를 모두 세었다는 것을 의미합니다.

3. **Delta compression using up to 10 threads**
   - 객체의 델타 압축을 수행하는 단계입니다.
   - 최대 10개의 스레드를 사용하여 델타 압축을 수행합니다.
   - 델타 압축은 유사한 객체들 간의 차이점만을 저장하여 공간을 절약하는 방법입니다.

4. **Compressing objects: 100% (500/500), done.**
   - 500개의 객체가 압축되었습니다.
   - 델타 압축된 객체들이 최종적으로 압축되는 단계입니다.

5. **Writing objects: 100% (552/552), done.**
   - 압축된 객체들을 디스크에 쓰는 단계입니다.
   - 552개의 객체가 최종적으로 디스크에 기록되었습니다.

6. **Total 552 (delta 26), reused 6 (delta 0), pack-reused 0**
   - 최종적으로 552개의 객체가 포함되었습니다.
   - 이 중 26개의 객체가 델타 압축되었습니다.
   - 6개의 객체가 재사용되었습니다 (이미 존재하는 객체를 다시 사용).
   - 패킹 과정에서 0개의 객체가 재사용되었습니다.
