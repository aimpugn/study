# out parameter

- [out parameter](#out-parameter)
    - [Out Parameter의 의미](#out-parameter의-의미)
    - [유래](#유래)
    - [예제](#예제)
        - [C# 예제](#c-예제)
        - [C++ 예제](#c-예제-1)
        - [Python 예제](#python-예제)
    - [참고사항](#참고사항)

## Out Parameter의 의미

"Out Parameter" 또는 "Output Parameter"는 프로그래밍에서 함수나 메서드의 매개변수 중 하나 이상의 값을 함수 외부로 반환하는 데 사용되는 매개변수를 의미합니다. 일반적으로 함수는 반환 값으로 하나의 데이터만을 반환하지만, Out Parameter를 사용하면 여러 값을 반환하거나, 함수 내부에서 계산된 값을 외부 변수에 직접 할당할 수 있습니다.

## 유래

Out Parameter의 개념은 프로그래밍 언어가 발전하면서 다중 값을 반환하거나, 함수 내에서 처리된 결과를 외부로 전달하는 메커니즘으로 발전했습니다. 초기의 프로그래밍 언어에서는 한 번에 하나의 값만 반환할 수 있었기 때문에, 여러 값을 반환하거나, 함수 외부의 상태를 변경하기 위해 Out Parameter가 사용되었습니다.

## 예제

### C# 예제

C# 언어에서는 `out` 키워드를 사용하여 Out Parameter를 선언합니다.

```csharp
public void GetCoordinates(out int x, out int y) {
    x = 10;
    y = 20;
}

public void Example() {
    int x, y;
    GetCoordinates(out x, out y);
    Console.WriteLine($"X: {x}, Y: {y}"); // X: 10, Y: 20
}
```

### C++ 예제

C++에서는 참조(&)를 사용하여 Out Parameter를 구현할 수 있습니다.

```cpp
void GetCoordinates(int &x, int &y) {
    x = 10;
    y = 20;
}

int main() {
    int x, y;
    GetCoordinates(x, y);
    cout << "X: " << x << ", Y: " << y << endl; // X: 10, Y: 20
    return 0;
}
```

### Python 예제

Python은 다중 반환 값을 지원하기 때문에, 튜플을 사용하여 여러 값을 반환하는 것이 일반적입니다. 따라서, Python에서는 Out Parameter 대신 다중 반환 값을 사용합니다.

```python
def get_coordinates():
    x = 10
    y = 20
    return x, y

x, y = get_coordinates()
print(f"X: {x}, Y: {y}") # X: 10, Y: 20
```

## 참고사항

- Out Parameter는 함수의 사이드 이펙트를 발생시키기 때문에, 함수형 프로그래밍에서는 일반적으로 권장되지 않습니다.
- 많은 현대 프로그래밍 언어는 다중 반환 값을 지원하여 Out Parameter의 필요성을 줄여줍니다.
