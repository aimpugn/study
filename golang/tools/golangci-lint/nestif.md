# nestif

- [nestif](#nestif)
    - [복잡성 측정 원리](#복잡성-측정-원리)
    - [복잡성 계산](#복잡성-계산)
        - [복잡도 12의 예제](#복잡도-12의-예제)

## 복잡성 측정 원리

1. [Cyclomatic Complexity (순환 복잡도)](../../../terminology/function_complexity.md):

    코드의 복잡성을 측정하는 지표입니다.
    독립적인 실행 경로의 수를 기반으로 계산됩니다.

2. Nesting Level (중첩 수준): 코드 블록의 중첩 깊이도 복잡성에 영향을 줍니다.

## 복잡성 계산

자세한 복잡도 계산 방법은 [`nestif`의 복잡도 계산 규칙](https://github.com/nakabonne/nestif?tab=readme-ov-file#rules)을 참고합니다.

### 복잡도 12의 예제

```go
func (l *SomStruct) SomeFunc(command Command, response *Response) {
    if !recv.needMoreInfo {
        return
    }

    provider := command.GetProvider()
    var bInfoCode *string

    if slices.Contains([]string{"a", "b"}, provider) { // 0

        if command.GetACode() == "AAAA" { // +1

            if someInfo := command.GetSomeInfo("BBBB"); someInfo != nil { // +2
                response.SomeInfoName = &someInfo.Name
            }
        }

        if tryBCode := command.GetBCode(); convutil.IsNotEmptyString(tryBCode) { // +1
            bInfoCode = &tryBCode

            if len(tryBCode) == 4 { // +2

                if bInfo := command.GetBInfo(); bInfo != nil { // +3
                    bInfoCode = &info.Code
                } else { // +1 (else 분기)
                    bInfoCode = nil
                }
            }
        }
    } else if provider == "c" { // +1 (else if 분기)

        if someInfo, err := bbbutil.GetSomeInfo(response.field1); err == nil && someInfo.ThatNumber > 0 { // +1 (조건문)
            bInfoCode = ptrutils.Ptr(fmt.Sprintf("%03d", someInfo.ThatNumber))
        }
    }

    response.BInfoCode = bInfoCode
}
```
