# cache

- [cache](#cache)
    - [전역 변수로 사용하기 vs 구조체 속성으로 사용하기](#전역-변수로-사용하기-vs-구조체-속성으로-사용하기)
    - [한번 업로드되어 수정되지 않는 파일 캐싱할 때 락이 필요한지?](#한번-업로드되어-수정되지-않는-파일-캐싱할-때-락이-필요한지)
        - [잠재적 사이드 이펙트](#잠재적-사이드-이펙트)
        - [추천 솔루션](#추천-솔루션)
        - [예시 코드](#예시-코드)

## 전역 변수로 사용하기 vs 구조체 속성으로 사용하기

아래와 같이 글로벌 캐시를 하나 선언합니다.
이때 이 캐시를 그대로 전역 변수로 사용하는 게 좋을지, 아니면 구조체 속성으로 주입하여

```go
var cacheInstance *lru.Cache[string, []byte]

func init() {
  // 캐시 초기화
  var err error
  cacheInstance, err = lru.New[string, []byte](128)
  if err != nil {
    panic(err)
  }
}
```

1. 글로벌 변수로 사용하는 것

    **장점**

    - 코드가 단순하고, 어디서나 쉽게 접근할 수 있습니다.
    - 초기화 코드가 간단하고 특정 초기화 함수를 호출할 필요가 없습니다.

    **단점**

    - 글로벌 상태는 테스트하기 어렵고, 테스트 간에 상태가 공유되기 때문에 테스트가 상호 의존적이 될 수 있습니다.
    - 여러 고루틴이 동시에 접근할 경우 동시성 문제를 일으킬 수 있습니다. 이를 위해 추가적인 동기화 메커니즘이 필요합니다.
    - 글로벌 변수를 사용하면 유연성이 떨어지고, 의존성 주입이 어렵습니다.

    ```go
    var cacheInstance *lru.Cache[string, []byte]

    func init() {
        var err error
        cacheInstance, err = lru.New 
        if err != nil {
            panic(err)
        }
    }

    func GetFromCache(key string) ([]byte, bool) {
        return cacheInstance.Get(key)
    }
    ```

2. 구조체의 속성으로 사용하는 것

    **장점**

    - 구조체를 통해 필요한 곳에 의존성을 주입할 수 있습니다.
    - 모킹을 통해 테스트할 수 있고, 구조체 인스턴스를 생성하여 독립적인 테스트가 가능합니다.
    - 캐시 인스턴스를 사용하는 로직이 구조체 내부에 캡슐화되어 관리됩니다.

    **단점**

    - 초기화와 관리가 다소 복잡해질 수 있습니다.
    - 구조체 정의와 초기화 코드가 필요합니다.

    ```go
    type CacheWrapper struct {
        Cache *lru.Cache[string, []byte]
    }

    func NewCacheWrapper(size int) (*CacheWrapper, error) {
        cache, err = lru.New[string, []byte](size)
        if err != nil {
            return nil, err
        }
        return &CacheWrapper{Cache: cache}, nil
    }

    func (cw *CacheWrapper) GetFromCache(key string) ([]byte, bool) {
        return cw.Cache.Get(key)
    }
    ```

차이가 발생하는 이유는 다음과 같습니다:
- **글로벌 상태**는 프로그램 전체에서 공유되기 때문에 언제든지 접근할 수 있지만, 이는 의존성이 명시적으로 드러나지 않게 합니다.
- **구조체 속성**으로 사용하는 경우, 의존성을 명시적으로 드러내고 주입할 수 있습니다. 이는 코드의 테스트 가능성과 유연성을 높입니다.

어떤 방법이 더 좋을까?
- **단순한 프로그램**이거나 **프로토타입**을 빠르게 개발할 때는 글로벌 변수가 더 나을 수 있습니다. 코드가 간단해지고, 빠르게 구현할 수 있기 때문입니다.
- **확장 가능하고 유지보수 가능한 프로그램**에서는 구조체 속성으로 사용하는 것이 더 좋습니다. 이는 테스트 가능성과 코드의 유연성을 높이며, 캡슐화를 통해 코드의 구조를 더 명확하게 만들 수 있습니다.

## 한번 업로드되어 수정되지 않는 파일 캐싱할 때 락이 필요한지?

캐시되는 데이터는 s3 파일로, s3에 한번 업로드되면 해당 파일은 수정되지 않습니다.

그렇다면,
1. 캐시 추가할 때 다른 고루틴이 먼저 조회해서 없다고 생각해서 s3 가져오고, 그 동안 캐시됐다 하더라도, 같은 파일이니 또 업데이트 해도 사실 문제는 없다.
2. 캐시가 이미 있을 때는 문제가 되지 않고
3. LRU로 사용되지 않아서 사라졌다고 해도 다시 필요해지면 1번으로 동작한다.

이러면 굳이 락이 필요한가?

락을 사용해야 하는지에 대한 결정은 데이터의 일관성, 성능 요구사항, 그리고 코드의 복잡성 간의 균형을 맞추는 문제입니다. 주어진 상황에서 락을 사용해야 하는지 판단하기 위해 몇 가지 중요한 사항을 고려해보겠습니다.

1. 데이터 일관성

    만약 여러 고루틴이 동시에 캐시에 접근하여 데이터를 추가하거나 업데이트한다면,
    데이터 일관성을 보장하기 위해 락을 사용하는 것이 일반적입니다.

    그러나, 캐시되는 데이터가 S3에서 가져온 파일이며, 파일이 업로드된 후 수정되지 않는다고 가정할 때,
    동일한 파일을 여러 번 캐시에 추가하는 것은 큰 문제가 되지 않을 수 있습니다.

2. 성능

    락을 사용하면 여러 고루틴이 동시에 캐시에 접근하는 것을 제어할 수 있어 데이터 일관성을 보장할 수 있지만,
    락으로 인해 성능이 저하될 수 있습니다.

    락이 필요 없는 경우, 성능을 최적화할 수 있습니다.

3. 코드의 복잡성

    락을 사용하면 코드가 복잡해지고, 디버깅과 유지보수가 어려워질 수 있습니다.
    락을 사용하지 않는다면, 코드가 더 간단해질 수 있습니다.

### 잠재적 사이드 이펙트

- **Race Condition**:

    두 개 이상의 고루틴이 동시에 캐시에 접근하여 데이터를 추가하거나 삭제할 때, 예상치 못한 동작이 발생할 수 있습니다.

    예를 들어, 하나의 고루틴이 데이터를 추가하려고 할 때, 다른 고루틴이 해당 데이터를 이미 추가했거나 삭제했을 수 있습니다.

- **Data Corruption**

    데이터가 일관되지 않게 저장될 수 있습니다.

- **중복된 작업**

    여러 고루틴이 동일한 데이터를 S3에서 가져오고, 이를 캐시에 추가하려고 할 때, 중복된 작업이 발생할 수 있습니다.

### 추천 솔루션

**락 없이 사용**:
주어진 조건에서 락 없이 사용하는 것이 성능상 더 유리할 수 있습니다. 단, 이를 위해서는 몇 가지 주의사항이 필요합니다:
- **Idempotent Operations**:

    데이터 추가 및 업데이트가 idemopent(멱등) 하도록 보장해야 합니다.
    즉, 동일한 데이터를 여러 번 추가하거나 업데이트해도 결과가 동일해야 합니다.

- **Double-Checked Locking**:

    데이터가 이미 캐시에 있는지 확인하고, 없을 때만 데이터를 추가하는 방식을 사용합니다.

### 예시 코드

락 없이 사용하는 경우:

```go
var cacheInstance *lru.Cache[string, []byte]

func init() {
    var err error
    cacheInstance, err = lru.New 
    if err != nil {
        panic(err)
    }
}

func GetFromCache(key string) ([]byte, bool) {
    value, ok := cacheInstance.Get(key)
    if ok {
        return value, true
    }
    return nil, false
}

func AddToCache(key string, value []byte) {
    // Double-Checked Locking
    if _, ok := cacheInstance.Get(key); !ok {
        cacheInstance.Add(key, value)
    }
}

func FetchAndCache(key string) ([]byte, error) {
    if value, ok := GetFromCache(key); ok {
        return value, nil
    }

    // Fetch from S3
    value, err := FetchFromS3(key)
    if err != nil {
        return nil, err
    }

    AddToCache(key, value)
    return value, nil
}

func FetchFromS3(key string) ([]byte, error) {
    // S3에서 데이터를 가져오는 로직
    // ...
    return []byte("example data"), nil
}
```

이 예제에서는 `FetchAndCache` 함수를 사용하여 데이터를 캐시에서 가져오거나, 없을 경우 S3에서 가져와 캐시에 추가합니다. 중복된 작업이 발생하더라도, 데이터의 일관성이 보장되므로 큰 문제가 되지 않습니다.
