# Template

## Template Pattern (템플릿 패턴)

행동을 클래스로부터 분리하는 데 사용되는 기법.
객체의 행동을 다양화하고 확장할 수 있는 메커니즘을 제공함으로써 소프트웨어의 유연성과 재사용성을 증가시킨다.

템플릿 패턴은 알고리즘의 구조를 정의하고 일부 단계를 서브클래스로 연기하는 디자인 패턴이다.
즉, 알고리즘의 일부를 구현하면서 전체 구조의 변경을 방지하는 방법이다.

- **핵심 로직과 필수 로직의 분리**: 템플릿 패턴을 사용하면 알고리즘의 핵심적인 부분은 상위 클래스에서 정의하고, 그 핵심 로직에 해당되지 않는 선택적 혹은 변화가 필요한 부분은 하위 클래스에서 구현할 수 있습니다. 이를 통해 상위 클래스에서는 변하지 않는 필수적인 로직의 흐름을 정의하고, 하위 클래스에서는 이러한 흐름에 따른 구체적인 실행 로직을 제공하게 됩니다.

- **필수적인 로직의 강제**: 템플릿 메소드 내에서 정의된 단계(예: 알고리즘의 특정 단계를 실행하는 메소드 호출)는 오버라이드할 수 없는 `final` 메소드로 선언되어, 서브클래스가 이를 변경할 수 없게 함으로써 필수적인 실행 순서나 로직을 강제할 수 있습니다.

## Example

### (Java) 핵심 로직 외의 필수적인 로직들을 강제

```java
// 추상 클래스 - 템플릿 메소드를 정의
abstract class Game {
    // 템플릿 메소드 - final로 선언하여 오버라이드를 방지
    public final void play() {
        initialize();
        startPlay();
        endPlay();
    }

    // 각 단계별 메소드 - 서브클래스에서 구현
    abstract void initialize();
    abstract void startPlay();
    abstract void endPlay();
}

// 서브클래스 1
class Soccer extends Game {
    @Override
    void initialize() {
        System.out.println("Soccer Game Initialized! Start playing.");
    }

    @Override
    void startPlay() {
        System.out.println("Soccer Game Started. Enjoy the game!");
    }

    @Override
    void endPlay() {
        System.out.println("Soccer Game Finished!");
    }
}

// 메인 클래스
public class TemplatePatternDemo {
    public static void main(String[] args) {
        Game game = new Soccer();
        game.play(); // 템플릿 메소드 호출
    }
}
```

### (Go) 핵심 로직 외의 필수적인 로직들을 강제

Go에서는 템플릿 패턴을 구현하기 위해 인터페이스와 구조체 임베딩을 사용할 수 있다.
구조체 임베딩을 통해 '상속'과 유사한 효과를 낼 수 있으며, 인터페이스를 통해 필수 메서드를 강제할 수 있다.

```go
package main

import "fmt"

// Game 인터페이스는 play 메서드를 정의합니다.
type Game interface {
    play()
}

// BaseGame 구조체는 공통 로직을 가지고 있습니다.
type BaseGame struct {}

// 공통 로직을 실행하는 메서드들
func (g *BaseGame) initialize() {
    fmt.Println("Game is initializing.")
}
func (g *BaseGame) startPlay() {
    fmt.Println("Game is starting.")
}
func (g *BaseGame) endPlay() {
    fmt.Println("Game is finishing.")
}

// Soccer는 BaseGame을 임베딩하여 Game 인터페이스를 구현합니다.
type Soccer struct {
    BaseGame
}

func (s *Soccer) play() {
    s.initialize()
    s.startPlay()
    // Soccer 특화 로직
    fmt.Println("Soccer Game is playing.")
    s.endPlay()
}

func main() {
    var game Game = &Soccer{}
    game.play()
}
```
