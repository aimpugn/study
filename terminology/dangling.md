# dangling

- [dangling](#dangling)
    - [dangling이란?](#dangling이란)
    - [Docker에서의 Dangling 이미지](#docker에서의-dangling-이미지)
    - [Git에서의 Dangling 커밋](#git에서의-dangling-커밋)
    - [프로그래밍 언어에서의 Dangling 포인터](#프로그래밍-언어에서의-dangling-포인터)
    - [데이터베이스에서의 Dangling 레퍼런스](#데이터베이스에서의-dangling-레퍼런스)
    - [HTML에서의 Dangling 앵커](#html에서의-dangling-앵커)

## dangling이란?

"dangling"이라는 용어는 다양한 컨텍스트에서 사용되며, 일반적으로 "어딘가에 연결되지 않은" 또는 "사용되지 않는" 상태를 의미합니다.
개발 및 IT 분야에서 "dangling"은 주로 리소스가 더 이상 참조되지 않거나 사용되지 않는 상태를 설명하는 데 사용됩니다.
이를 통해 리소스 관리와 참조 무결성을 유지하는 데 중요한 역할을 합니다.

- **Docker**: 더 이상 사용되지 않는 중간 이미지 레이어.
- **Git**: 브랜치나 태그로 참조되지 않는 커밋.
- **프로그래밍 언어**: 더 이상 유효하지 않은 메모리 주소를 가리키는 포인터.
- **데이터베이스**: 참조 무결성을 위반하는 외래 키.
- **HTML**: 존재하지 않는 앵커를 참조하는 링크.

## Docker에서의 Dangling 이미지

Docker에서 "dangling" 이미지는 더 이상 사용되지 않는 중간 이미지 레이어를 의미합니다.
- 특정 이미지 태그가 제거된 경우
- 새로운 이미지 빌드 과정에서 중간 단계로 생성된 이미지 레이어가 더 이상 참조되지 않는 경우

```sh
# 모든 "dangling" 이미지를 제거합니다.
docker image prune
```

"dangling" 이미지는 `docker images` 명령어를 실행했을 때 `<none>` 태그로 표시됩니다.

```sh
$ docker images -f "dangling=true"
REPOSITORY          TAG                 IMAGE ID            CREATED             SIZE
<none>              <none>              4e38e38c8ce0        2 weeks ago         1.23GB
```

## Git에서의 Dangling 커밋

Git에서 "dangling" 커밋은 브랜치나 태그로 참조되지 않는 커밋을 의미합니다.
이는 브랜치가 삭제되었거나, 리베이스 또는 체리픽 과정에서 발생할 수 있습니다.

```sh
git fsck --full --unreachable
```

위 명령어는 "dangling" 커밋을 포함한 모든 접근할 수 없는 객체를 찾습니다.

```sh
$ git fsck --full --unreachable
unreachable commit 4e38e38c8ce0
dangling commit 5f47e47d8df1
```

## 프로그래밍 언어에서의 Dangling 포인터

C나 C++ 같은 저수준 프로그래밍 언어에서 "dangling" 포인터는 더 이상 유효하지 않은 메모리 주소를 가리키는 포인터를 의미합니다.
이는 메모리가 해제된 후에도 포인터가 해당 메모리 주소를 참조하는 경우에 발생합니다.

```c
#include <stdio.h>
#include <stdlib.h>

int main() {
    int *ptr = (int *)malloc(sizeof(int));
    *ptr = 42;
    free(ptr);
    // ptr은 이제 dangling 포인터입니다.
    printf("%d\n", *ptr); // 정의되지 않은 동작
    return 0;
}
```

## 데이터베이스에서의 Dangling 레퍼런스

데이터베이스에서 "dangling" 레퍼런스는 참조 무결성을 위반하는 상태를 의미합니다.
이는 외래 키가 참조하는 레코드가 삭제되었거나 존재하지 않는 경우에 발생합니다.

```sql
CREATE TABLE orders (
    order_id INT PRIMARY KEY,
    customer_id INT,
    FOREIGN KEY (customer_id) REFERENCES customers(customer_id)
);

DELETE FROM customers WHERE customer_id = 1;
-- 이제 orders 테이블의 customer_id가 1인 레코드는 dangling 레퍼런스를 가집니다.
```

## HTML에서의 Dangling 앵커

HTML에서 "dangling" 앵커는 존재하지 않는 앵커를 참조하는 링크를 의미합니다.
이는 페이지 내에 정의되지 않은 앵커를 참조하는 경우에 발생합니다.

```html
<a href="#section1">Go to Section 1</a>
<!-- #section1 앵커가 정의되지 않은 경우, 이 링크는 dangling 앵커를 참조합니다. -->
```
