# Cohesion and Coupling

- [Cohesion and Coupling](#cohesion-and-coupling)
    - [결합도와 응집도](#결합도와-응집도)
    - [결합도 낮추고 응집도 높이기](#결합도-낮추고-응집도-높이기)
    - [결합도의 유형 (낮은 결합도에서 높은 결합도 순)](#결합도의-유형-낮은-결합도에서-높은-결합도-순)
        - [메시지 결합도 (Message Coupling)](#메시지-결합도-message-coupling)
        - [데이터 결합도 (Data Coupling)](#데이터-결합도-data-coupling)
        - [스탬프 결합도 (Stamp Coupling)](#스탬프-결합도-stamp-coupling)
        - [제어 결합도 (Control Coupling)](#제어-결합도-control-coupling)
        - [외부 결합도 (External Coupling)](#외부-결합도-external-coupling)
        - [공통 결합도 (Common Coupling)](#공통-결합도-common-coupling)
        - [내용 결합도 (Content Coupling)](#내용-결합도-content-coupling)
        - [결합도를 낮추는 방법](#결합도를-낮추는-방법)
    - [응집도의 유형 (낮은 응집도에서 높은 응집도 순)](#응집도의-유형-낮은-응집도에서-높은-응집도-순)
        - [우연적 응집도 (Coincidental Cohesion)](#우연적-응집도-coincidental-cohesion)
        - [논리적 응집도 (Logical Cohesion)](#논리적-응집도-logical-cohesion)
        - [시간적 응집도 (Temporal Cohesion)](#시간적-응집도-temporal-cohesion)
        - [절차적 응집도 (Procedural Cohesion)](#절차적-응집도-procedural-cohesion)
        - [통신적 응집도 (Communicational Cohesion)](#통신적-응집도-communicational-cohesion)
        - [순차적 응집도 (Sequential Cohesion)](#순차적-응집도-sequential-cohesion)
        - [기능적 응집도 (Functional Cohesion)](#기능적-응집도-functional-cohesion)
        - [응집도를 높이는 방법](#응집도를-높이는-방법)

## 결합도와 응집도

결합도와 응집도의 개념은 시스템 이론과 정보 은닉 원칙에 기반합니다.
복잡한 시스템을 관리 가능한 단위로 분해하고, 각 단위의 내부 복잡성을 캡슐화하는 것을 목표로 합니다.

- 결합도(Coupling):

    결합도는 서로 다른 모듈 간의 상호 의존성 또는 연관 정도를 나타냅니다.
    한 모듈이 다른 모듈의 내부 작동 방식에 대해 얼마나 많이 알고 있는지를 측정합니다.

    낮은 결합도는 한 모듈의 변경이 다른 모듈에 미치는 영향을 최소화합니다.
    이를 통해 모듈의 독립성을 증가시키고, 시스템의 유지보수성과 확장성을 향상시킵니다.

- 응집도(Cohesion):

    응집도는 모듈 내부의 요소들이 서로 관련되어 있는 정도를 나타냅니다.
    모듈 내의 기능들이 단일 목적을 위해 얼마나 밀접하게 관련되어 있는지를 측정합니다.

    높은 응집도는 모듈이 명확하고 단일한 목적을 가지고 있음을 의미합니다.
    이를 통해 모듈의 이해와 유지보수를 용이하게 하며, 재사용성을 높입니다.

결합도와 응집도는 상호 보완적인 관계에 있습니다.
높은 응집도를 가진 모듈은 자연스럽게 다른 모듈과의 결합도가 낮아지는 경향이 있습니다.

## 결합도 낮추고 응집도 높이기

결합도(Coupling)를 낮추고 응집도(Cohesion)를 높이면 유지보수성, 확장성, 그리고 재사용성을 향상시킬 수 있습니다.

- 변경의 국소화:

   낮은 결합도는 모듈 간 의존성을 줄여, 한 모듈의 변경이 다른 모듈에 미치는 영향을 최소화합니다.
   높은 응집도는 관련 기능을 하나의 모듈에 집중시켜, 변경이 필요할 때 영향 범위를 해당 모듈로 제한합니다.

- 복잡성 관리:

   각 모듈이 독립적이고 집중된 기능을 가짐으로써, 개발자는 전체 시스템을 완벽히 이해하지 않고도 개별 모듈을 이해하고 수정할 수 있습니다.
   유지보수성과 확장성을 향상시키며, 개발 과정에서의 복잡성을 줄여줍니다.

- 테스트 용이성:

   독립적인 모듈은 격리된 환경에서 더 쉽게 테스트할 수 있습니다.
   버그 발견과 수정을 용이하게 하며, 새로운 기능 추가 시 기존 기능의 안정성을 보장하는 데 도움이 됩니다.

- 병렬 개발:

   모듈 간 의존성이 낮으면 여러 개발자가 동시에 다른 모듈을 작업할 수 있어, 개발 프로세스의 효율성이 향상됩니다.
   > 모듈: 독립적인 기능 단위로, 클래스, 함수, 또는 컴포넌트가 될 수 있습니다.

- 교체 가능성:

   잘 정의된 인터페이스를 통해 낮은 결합도를 유지하면, 필요시 모듈을 다른 구현으로 쉽게 교체할 수 있으므로 시스템의 유연성을 높입니다.
   > 인터페이스: 모듈 간 상호작용의 명세

## 결합도의 유형 (낮은 결합도에서 높은 결합도 순)

### 메시지 결합도 (Message Coupling)

모듈들이 메시지나 인터페이스를 통해 통신하며, 명확히 정의된 인터페이스를 통해 상호 작용합니다.
가장 낮은 수준의 결합도입니다.

메시지 결합도는 가장 낮은 수준의 결합도입니다.
모듈 간 통신이 잘 정의된 메시지 패싱 인터페이스를 통해 이루어집니다.
각 모듈은 다른 모듈의 내부 구조나 구현 방식을 알 필요 없이, 오직 공개된 인터페이스를 통해서만 상호작용합니다.

```kotlin
// 결제 처리 시스템

interface PaymentGateway {
    fun processPayment(amount: BigDecimal): PaymentResult
}

class StripeGateway : PaymentGateway {
    override fun processPayment(amount: BigDecimal): PaymentResult {
        // Stripe API를 사용한 결제 처리 로직
        return PaymentResult(success = true, transactionId = "STRIPE-12345")
    }
}

class PayPalGateway : PaymentGateway {
    override fun processPayment(amount: BigDecimal): PaymentResult {
        // PayPal API를 사용한 결제 처리 로직
        return PaymentResult(success = true, transactionId = "PAYPAL-67890")
    }
}

// OrderService는 구체적인 결제 게이트웨이 구현에 의존하지 않고,
// PaymentGateway 인터페이스를 통해 결제를 처리합니다.
// 이를 통해 새로운 결제 방식을 추가하거나 기존 방식을 변경할 때
// OrderService의 코드를 수정할 필요가 없어집니다.
class OrderService(private val paymentGateway: PaymentGateway) {
    fun placeOrder(order: Order) {
        val result = paymentGateway.processPayment(order.totalAmount)
        if (result.success) {
            // 주문 처리 로직
            println("Order placed successfully. Transaction ID: ${result.transactionId}")
        } else {
            println("Payment failed. Order cannot be processed.")
        }
    }
}

// 사용 예
fun main() {
    val stripeOrderService = OrderService(StripeGateway())
    stripeOrderService.placeOrder(Order(totalAmount = BigDecimal("100.00")))

    val paypalOrderService = OrderService(PayPalGateway())
    paypalOrderService.placeOrder(Order(totalAmount = BigDecimal("150.00")))
}
```

### 데이터 결합도 (Data Coupling)

데이터 결합도는 모듈 간에 단순한 데이터를 매개변수로 주고받는 형태의 결합입니다.
이 경우 모듈이 필요한 최소한의 데이터만을 매개변수로 전달받아 사용하며, 각 모듈은 자신이 필요로 하는 데이터만 알면 됩니다.

모듈 간의 인터페이스가 간단하고 명확하며, 각 모듈이 서로의 내부 구조에 대해 알 필요가 없다는 점에서 여전히 모듈 간의 결합도는 낮습니다.
그러나 데이터의 구조가 변경되면 영향을 받을 수 있습니다.

```kotlin
// 주식 거래 시스템

data class StockOrder(val symbol: String, val quantity: Int, val price: BigDecimal)

class StockExchange {
    fun executeOrder(order: StockOrder): Boolean {
        // 주식 거래 실행 로직
        println("Executing order: ${order.quantity} shares of ${order.symbol} at ${order.price}")
        return true
    }
}

class TradingBot(private val exchange: StockExchange) {
    fun placeBuyOrder(symbol: String, quantity: Int, maxPrice: BigDecimal) {
        // TradingBot과 StockExchange 사이의 결합도는
        // StockOrder 데이터 구조를 통해 이루어집니다.
        // 두 모듈은 서로의 내부 구현에 대해 알 필요 없이 단순히 데이터를 주고받습니다.
        val order = StockOrder(symbol, quantity, maxPrice)
        if (exchange.executeOrder(order)) {
            println("Buy order placed successfully")
        } else {
            println("Failed to place buy order")
        }
    }
}

fun main() {
    val exchange = StockExchange()
    val bot = TradingBot(exchange)

    bot.placeBuyOrder("AAPL", 100, BigDecimal("150.00"))
}
```

### 스탬프 결합도 (Stamp Coupling)

스탬프 결합도는 모듈들이 공유 데이터 구조를 매개변수로 전달하지만, 각 모듈이 그 구조의 일부만을 사용합니다.
데이터 결합도와 비슷해 보이지만, 모듈이 필요 이상의 데이터를 포함한 구조체나 객체를 전체로 전달받는 경우에 해당합니다.
모듈이 실제로 사용하지 않는 데이터에 대해서도 의존성을 가지게 되고 필요 이상의 지식을 모듈에 부여하게 되어 모듈의 독립성을 저해합니다.

데이터 결합도보다 약간 더 높은 결합도를 가지며, 불필요한 데이터 의존성을 만들 수 있습니다.

> 전달되어도 모듈이 사용하지 않는다면 의존성이 없는 게 아닌지?
> - 모듈이 특정 데이터 구조 전체를 매개변수로 받음으로써, 그 데이터 구조의 타입에 의존하게 됩니다.
> - 현재는 사용하지 않더라도 모듈 내에서 그 데이터에 접근할 수 있는 가능성이 열려 있기 때문에, 향후 코드 변경 시 의도치 않은 사용으로 이어질 수 있습니다.
> - 필요 이상의 데이터를 포함한 매개변수는 모듈의 인터페이스를 불필요하게 복잡하게 만듭니다.
> - 모듈에 필요하지 않은 데이터까지 노출됨으로써, 정보 은닉 원칙이 약화될 수 있습니다.
> - 불필요한 데이터까지 포함된 객체를 사용해 테스트를 작성해야 하므로, 테스트 코드가 복잡해질 수 있습니다.

```kotlin
// 고객 관리 시스템

data class Customer(
    val id: String,
    val name: String,
    val email: String,
    val address: Address,
    val creditScore: Int
)

data class Address(
    val street: String,
    val city: String,
    val country: String,
    val postalCode: String
)

class EmailService {
    fun sendWelcomeEmail(customer: Customer) {
        // 이메일 서비스는 Customer의 name과 email만 사용합니다.
        println("Sending welcome email to ${customer.name} at ${customer.email}")
    }
}

class ShippingService {
    fun scheduleDelivery(customer: Customer) {
        // 배송 서비스는 Customer의 name과 address만 사용합니다.
        val address = customer.address
        println("Scheduling delivery for ${customer.name} to ${address.street}, ${address.city}, ${address.country}")
    }
}

fun main() {
    val customer = Customer(
        id = "C12345",
        name = "John Doe",
        email = "john@example.com",
        address = Address("123 Main St", "New York", "USA", "10001"),
        creditScore = 750
    )

    // EmailService와 ShippingService는 모두 전체 Customer 객체를 받지만,
    // 각각 필요한 부분만 사용합니다.
    // Customer 클래스의 변경이 이 서비스들에 영향을 줄 수 있습니다.
    val emailService = EmailService()
    // name, email
    emailService.sendWelcomeEmail(customer)

    val shippingService = ShippingService()
    // name, address
    shippingService.scheduleDelivery(customer)
}
```

### 제어 결합도 (Control Coupling)

제어 결합도는 한 모듈이 다른 모듈의 내부 로직이나 동작 방식을 제어하는 정보를 전달할 때 발생합니다.
모듈 간의 결합도를 증가시키고, 모듈의 독립성과 재사용성을 저하시킬 수 있습니다.

```kotlin
// 로깅 시스템

enum class LogLevel { DEBUG, INFO, WARNING, ERROR }

class AdvancedLogger {
    fun log(message: String, level: LogLevel) {
        when (level) {
            LogLevel.DEBUG -> println("DEBUG: $message")
            LogLevel.INFO -> println("INFO: $message")
            LogLevel.WARNING -> println("WARNING: $message")
            LogLevel.ERROR -> println("ERROR: $message")
        }
    }
}

class UserService(private val logger: AdvancedLogger) {
    fun registerUser(username: String, email: String) {
        // UserService는 AdvancedLogger의 log 메서드를 호출할 때,
        // 로그 레벨을 지정함으로써 로거의 동작을 제어합니다.
        // 이는 UserService가 AdvancedLogger의 내부 동작 방식에 대해 알고 있어야 함을 의미하며,
        // 결합도를 증가시킵니다.
        logger.log("Attempting to register user: $username", LogLevel.DEBUG)

        // 사용자 등록 로직
        if (validateEmail(email)) {
            logger.log("User registered successfully: $username", LogLevel.INFO)
        } else {
            logger.log("Failed to register user due to invalid email", LogLevel.ERROR)
        }
    }

    private fun validateEmail(email: String): Boolean {
        // 이메일 유효성 검사 로직
        return email.contains("@")
    }
}

fun main() {
    val logger = AdvancedLogger()
    val userService = UserService(logger)

    userService.registerUser("john_doe", "john@example.com")
    userService.registerUser("jane_doe", "invalid-email")
}
```

### 외부 결합도 (External Coupling)

외부 결합도는 모듈이 외부 환경(예: 운영 체제, 하드웨어, 외부 시스템 인터페이스 등)에 의존할 때 발생합니다.
외부 결합도가 높을수록 시스템의 유연성과 이식성은 떨어지며, 외부 환경의 변화에 따라 시스템의 동작이 영향을 받을 수 있습니다.

```kotlin
import java.io.File
import java.nio.file.Paths
import java.sql.DriverManager

class DataProcessor {
    // 하드코딩된 데이터베이스 연결 문자열을 사용합니다.
    // MySQL 특정 JDBC 드라이버에 의존하여, 다른 데이터베이스로 변경하기 어렵습니다.
    // 예를 들어, PostgreSQL이나 SQLite로 변경하려면 드라이버 및 연결 문자열을 수정해야 합니다.
    private val dbConnection = DriverManager.getConnection("jdbc:mysql://localhost:3306/mydb", "user", "password")

    fun processData(inputFilePath: String, outputFilePath: String) {
        // 파일 시스템에 의존하여 데이터 처리하는 방식은 클라우드 스토리지(AWS S3, Google Cloud Storage 등)로의 전환을 어렵게 만듭니다.
        val inputFile = File(inputFilePath)
        if (!inputFile.exists()) {
            throw IllegalArgumentException("Input file does not exist: $inputFilePath")
        }

        val data = inputFile.readLines()

        // 데이터베이스에 직접 의존
        val statement = dbConnection.createStatement()
        data.forEach { line ->
            statement.executeUpdate("INSERT INTO processed_data (content) VALUES ('$line')")
        }

        // 시스템 환경 변수에 직접 의존
        val outputDir = System.getenv("OUTPUT_DIR") ?: "/tmp"
        // 출력 파일 경로가 하드코딩되어 있어 다양한 배포 환경(예: 클라우드 저장소)에서 유연하게 대응하기 어렵습니다.
        val outputFile = Paths.get(outputDir, outputFilePath).toFile()

        // 파일 시스템에 다시 직접 의존
        outputFile.writeText(data.joinToString("\n"))

        println("Data processed and saved to ${outputFile.absolutePath}")
    }

    fun cleanup() {
        dbConnection.close()
    }
}

fun main() {
    val processor = DataProcessor()
    processor.processData("input.txt", "output.txt")
    processor.cleanup()
}
```

### 공통 결합도 (Common Coupling)

공통 결합도는 여러 모듈이 전역 데이터 구조나 외부 자원을 공유할 때 발생합니다.
이는 모듈 간의 숨겨진 의존성을 만들어 내며, 한 모듈의 변경이 다른 모듈에 예기치 않은 영향을 줄 수 있어 시스템의 유지보수성을 크게 저하시킵니다.

```kotlin
// 글로벌 설정 및 캐시 시스템

// GlobalConfig와 GlobalCache는 여러 모듈에서 공유되는 전역 상태입니다.
object GlobalConfig {
    var databaseUrl: String = "jdbc:mysql://localhost:3306/mydb"
    var maxConnections: Int = 10
    var debugMode: Boolean = false
}

object GlobalCache {
    private val cache = mutableMapOf<String, Any>()

    fun put(key: String, value: Any) {
        cache[key] = value
    }

    fun get(key: String): Any? = cache[key]

    fun clear() {
        cache.clear()
    }
}

class UserRepository {
    fun getUser(id: String): User? {
        if (GlobalConfig.debugMode) {
            println("Debug: Fetching user with id $id")
        }

        // 전역 객체들에 의존하고 있어, 높은 결합도를 가집니다.
        // GlobalConfig나 GlobalCache의 변경이 여러 모듈에 영향을 줄 수 있습니다.
        val cachedUser = GlobalCache.get(id) as? User
        if (cachedUser != null) return cachedUser

        // 데이터베이스에서 사용자 조회 로직
        val user = User(id, "John Doe")
        GlobalCache.put(id, user)
        return user
    }
}

class AuthenticationService {
    fun authenticate(username: String, password: String): Boolean {
        // 전역 객체들에 의존하고 있어, 높은 결합도를 가집니다.
        // GlobalConfig나 GlobalCache의 변경이 여러 모듈에 영향을 줄 수 있습니다.
        if (GlobalConfig.debugMode) {
            println("Debug: Authenticating user $username")
        }

        // 인증 로직
        return true
    }
}

data class User(val id: String, val name: String)

fun main() {
    GlobalConfig.debugMode = true

    val userRepo = UserRepository()
    val authService = AuthenticationService()

    val user = userRepo.getUser("12345")
    println("Retrieved user: ${user?.name}")

    val authenticated = authService.authenticate("john_doe", "password123")
    println("Authentication result: $authenticated")
}
```

### 내용 결합도 (Content Coupling)

내용 결합도는 가장 높은 수준의 결합도로, 한 모듈이 다른 모듈의 내부 구현이나 데이터 구조에 직접 접근하거나 수정하는 경우를 말합니다.
모듈의 캡슐화를 완전히 깨뜨리며, 시스템의 유지보수성과 확장성을 심각하게 저하시킵니다. 내용 결합도는 일반적으로 피해야 할 안티패턴으로 간주됩니다.

```kotlin
// 데이터베이스 연결 관리 시스템

class DatabaseConnection {
    var isConnected = false
    private var connectionString = ""

    fun connect(connectionString: String) {
        this.connectionString = connectionString
        // 데이터베이스 연결 로직
        isConnected = true
        println("Connected to database: $connectionString")
    }

    fun disconnect() {
        // 데이터베이스 연결 해제 로직
        isConnected = false
        println("Disconnected from database")
    }
}

class QueryExecutor(private val dbConnection: DatabaseConnection) {
    fun executeQuery(query: String) {
        // QueryExecutor 클래스는 DatabaseConnection 클래스의 내부 구현에 직접 접근하고 있습니다.
        // DatabaseConnection 클래스의 내부 구현이 변경되면 QueryExecutor 클래스도 함께 수정해야 합니다.
        if (!dbConnection.isConnected) {
            // 직접적으로 DatabaseConnection의 내부 상태를 변경
            dbConnection.isConnected = true
            println("Forcefully set connection status to connected")
        }

        // 쿼리 실행 로직
        println("Executing query: $query")
    }

    fun forceReconnect() {
        // DatabaseConnection의 private 필드에 직접 접근 (실제로는 컴파일 에러가 발생하지만, 내용 결합도의 예시로 사용)
        dbConnection.connectionString = "new_connection_string"
        dbConnection.connect(dbConnection.connectionString)
    }
}

fun main() {
    val dbConnection = DatabaseConnection()
    dbConnection.connect("initial_connection_string")

    val queryExecutor = QueryExecutor(dbConnection)
    queryExecutor.executeQuery("SELECT * FROM users")

    dbConnection.disconnect()
    queryExecutor.executeQuery("SELECT * FROM products") // 연결이 끊어졌음에도 쿼리 실행 시도

    queryExecutor.forceReconnect() // 내부 구현에 직접 접근하여 재연결 시도
}
```

### 결합도를 낮추는 방법

1. 인터페이스 사용: 구체적인 구현 대신 추상화된 인터페이스에 의존
2. 의존성 주입: 외부에서 의존성을 주입받아 사용
3. 이벤트 기반 통신: 직접적인 메서드 호출 대신 이벤트를 통해 통신
4. 중재자 패턴: 직접적인 통신을 중재자 객체를 통해 간접화

## 응집도의 유형 (낮은 응집도에서 높은 응집도 순)

응집도가 높을수록 모듈이 단일 책임을 가지며, 유지보수성과 재사용성이 높아집니다.
기능적 응집도가 가장 이상적인 응집도로 간주되며, 그 외의 응집도는 점점 더 낮은 수준의 응집도를 나타냅니다.

### 우연적 응집도 (Coincidental Cohesion)

우연적 응집도는 가장 낮은 수준의 응집도로, 모듈 내의 기능들이 서로 아무런 연관성 없이 단순히 같은 모듈에 모여있는 경우입니다.
일반적으로 임의로 기능을 추가하면서 발생하며, 함수나 메서드들이 서로 관련 없이 존재하기 때문에 유지보수가 매우 어렵고, 모듈을 이해하기도 어렵습니다.

```kotlin
// 모듈 내에 전혀 연관성이 없는 기능들이 임의로 모여 있습니다.
class Utility {
    // 단순 문자열 출력
    fun printWelcomeMessage() {
        println("Welcome to our service!")
    }

    // 로그 파일 생성
    fun createLogFile(fileName: String) {
        println("Creating log file: $fileName")
    }

    // 난수 생성
    fun generateRandomNumber(): Int {
        return (1..100).random()
    }
}
```

### 논리적 응집도 (Logical Cohesion)

논리적 응집도는 비슷한 범주의 기능들이 모듈에 묶여 있지만, 각 기능이 서로 다른 작업을 수행하는 경우입니다.
즉, 논리적으로 비슷해 보이지만 기능적으로는 다릅니다.

예를 들어, 입력 처리라는 큰 범주 안에 여러 입력 장치(키보드, 마우스, 터치)의 처리가 함께 들어있는 경우입니다. 모듈 내 메서드들이 동일한 범주에 속하긴 하지만 실제로는 서로 다른 작업을 처리합니다.

```kotlin
// NotificationHandler 클래스는 알림을 처리하는 기능을 논리적으로 묶어놓았지만,
// 메서드마다 다른 유형의 알림을 전송합니다.
// 기능적으로 관련이 없는 메서드들이 단순히 같은 범주에 속해 있기 때문에 응집도가 낮습니다.
class NotificationHandler {
    // 이메일로 알림 전송
    fun sendEmailNotification(email: String, message: String) {
        println("Sending email to $email with message: $message")
    }

    // SMS로 알림 전송
    fun sendSmsNotification(phoneNumber: String, message: String) {
        println("Sending SMS to $phoneNumber with message: $message")
    }

    // 푸시 알림 전송
    fun sendPushNotification(deviceId: String, message: String) {
        println("Sending push notification to device $deviceId with message: $message")
    }
}
```

### 시간적 응집도 (Temporal Cohesion)

시간적 응집도는 특정 시점이나 이벤트에서 함께 실행되어야 하는 기능들이 모여 있는 경우를 의미합니다.

예를 들어, 애플리케이션의 초기화 단계에서 실행되어야 하는 여러 초기화 작업들이 함께 모듈에 포함되어 있는 경우입니다.
이들은 실행 시점이 동일하다는 이유로 모여 있지만, 기능 간의 직접적인 연관성은 적습니다.

```kotlin
// 애플리케이션이 시작될 때 실행되는 초기화 작업들을 포함하고 있습니다.
// 서로 기능적으로 연관성이 적지만, 실행 시점이 동일하므로 시간적 응집도를 가집니다.
class ApplicationInitializer {
    // 애플리케이션 시작 시 설정 파일 로드
    fun loadConfiguration() {
        println("Loading configuration...")
    }

    // 데이터베이스 초기화
    fun initializeDatabase() {
        println("Initializing database...")
    }

    // 로깅 시스템 설정
    fun setupLogging() {
        println("Setting up logging...")
    }
}
```

### 절차적 응집도 (Procedural Cohesion)

절차적 응집도는 모듈 내의 기능들이 특정 절차에 따라 순서대로 실행되는 경우에 해당합니다.
이때 각 기능은 어느 정도 연관성이 있더라도, 그들이 하나의 구체적인 책임을 수행하는 방식보다는 일정한 순서에 따라 묶여 있을 뿐이라는 특징을 가지고 있습니다.

하나의 명확한 목적을 달성하기 위해 협력하는 것이 아니라, 순서에 따라 진행되어야 하는 다양한 작업이 절차적으로 묶여 있을 뿐입니다.
즉, *작업들은 독립적으로 사용되기 어려우며, 순서에 대한 강한 의존성이 존재*합니다.

절차가 바뀌면 모듈이 제대로 동작하지 않을 수 있기 때문에 유지보수가 복잡할 수 있습니다.

```kotlin
class UserOnboardingProcessor {
    fun createUserAccount(userInfo: Map<String, String>) {
        println("Creating user account for ${userInfo["name"]}")
    }

    fun sendWelcomeEmail(userEmail: String) {
        println("Sending welcome email to $userEmail")
    }

    fun setupUserPreferences(userId: String) {
        println("Setting up preferences for user $userId")
    }
}
```

이를 응집도 높게 만들려면 단일 책임 원칙에 따라 각 책임을 분리하고, 관련된 작업을 모듈화 합니다.

```kotlin
// 사용자 계정 관련 작업을 담당하는 클래스
class UserAccountService {
    fun createUserAccount(userInfo: Map<String, String>): String {
        val userId = "generatedUserId" // 사용자 ID 생성 예시
        println("Creating user account for ${userInfo["name"]}")
        return userId
    }
}

// 이메일 전송 작업을 담당하는 클래스
class EmailService {
    fun sendWelcomeEmail(userEmail: String) {
        println("Sending welcome email to $userEmail")
    }
}

// 사용자 설정 초기화를 담당하는 클래스
class UserPreferencesService {
    fun setupUserPreferences(userId: String) {
        println("Setting up preferences for user $userId")
    }
}

// 온보딩 프로세스를 조정하는 모듈
class UserOnboardingProcessor(
    private val userAccountService: UserAccountService,
    private val emailService: EmailService,
    private val userPreferencesService: UserPreferencesService
) {
    fun onboardNewUser(userInfo: Map<String, String>, userEmail: String) {
        val userId = userAccountService.createUserAccount(userInfo)
        emailService.sendWelcomeEmail(userEmail)
        userPreferencesService.setupUserPreferences(userId)
        println("User onboarding completed for ${userInfo["name"]}")
    }
}

fun main() {
    val userAccountService = UserAccountService()
    val emailService = EmailService()
    val userPreferencesService = UserPreferencesService()
    val onboardingProcessor = UserOnboardingProcessor(userAccountService, emailService, userPreferencesService)

    onboardingProcessor.onboardNewUser(
        mapOf("name" to "John Doe"),
        "john.doe@example.com"
    )
}
```

각각의 클래스가 특정한 책임을 가지기 때문에 응집도가 높아지고, 각 클래스를 독립적으로 재사용하거나 유지보수하기 쉬워집니다.

### 통신적 응집도 (Communicational Cohesion)

통신적 응집도는 모듈 내의 메서드들이 동일한 입력 데이터를 사용하거나 동일한 출력 데이터를 생성하는 경우입니다.
이때 기능들은 데이터의 흐름을 기반으로 연결되어 있으며, 해당 데이터를 처리하는 모든 작업이 모듈에 모여 있습니다.

```kotlin
// `ReportGenerator` 클래스의 메서드들은 공통의 데이터(`fetchData`로 가져온 데이터)를 사용하여 작업을 수행합니다.
// 이 메서드들은 동일한 데이터를 기반으로 작업하기 때문에 통신적 응집도를 가집니다.
class ReportGenerator {
    // 데이터베이스에서 보고서에 필요한 데이터를 가져옴
    fun fetchData(): List<String> {
        println("Fetching data...")
        return listOf("Data1", "Data2", "Data3")
    }

    // 가져온 데이터를 기반으로 보고서 생성
    fun generateReport(data: List<String>): String {
        println("Generating report from data: $data")
        return "Report: $data"
    }

    // 생성된 보고서를 이메일로 전송
    fun sendReport(report: String) {
        println("Sending report: $report")
    }
}
```

### 순차적 응집도 (Sequential Cohesion)

순차적 응집도는 한 기능의 출력이 다른 기능의 입력으로 사용되는 경우입니다.
이때 모듈의 기능들은 서로 밀접하게 연결되어 있으며, 하나의 작업이 끝나면 다음 작업으로 결과가 넘어갑니다.

```kotlin
class OrderProcessor {

    // 주문이 유효한지 검증. 유효성 여부를 반환.
    fun validateOrder(order: String): Boolean {
        println("Validating order: $order")
        return order.isNotEmpty() // 예시: 빈 문자열이면 유효하지 않음
    }

    // 주문 처리: 주문을 처리하고 처리된 결과를 반환.
    fun processOrder(order: String): String {
        println("Processing order: $order")
        return "Processed Order: $order" // 처리된 주문의 결과를 반환
    }

    // 처리된 주문을 배송.
    fun shipOrder(processedOrder: String) {
        println("Shipping order: $processedOrder")
    }
}

fun main() {
    val orderProcessor = OrderProcessor()

    // 실제로 주문을 검증하고, 처리하고, 배송하는 일련의 순차적 처리 흐름을 보여줍니다.
    val order = "Order1234"

    if (orderProcessor.validateOrder(order)) {
        val processedOrder = orderProcessor.processOrder(order)
        orderProcessor.shipOrder(processedOrder)
    } else {
        println("Invalid order: $order")
    }
}
```

### 기능적 응집도 (Functional Cohesion)

기능적 응집도는 가장 높은 수준의 응집도로, 모듈 내 모든 기능이 단일 목적을 위해 협력하는 경우입니다.
모듈이 하나의 명확한 책임을 가지고 있으며, 이 책임을 달성하기 위한 모든 기능이 모듈 내에서 긴밀하게 협력합니다.
이러한 모듈은 단일 책임 원칙을 철저히 따르며, 재사용성과 유지보수성이 뛰어납니다.

```kotlin
// `AuthenticationService` 클래스는 사용자 인증이라는 명확한 하나의 목적을 가지고 있습니다.
// 이 클래스의 모든 메서드는 이 목적을 달성하기 위해 협력합니다.
class AuthenticationService {
    // 사용자 인증 처리
    fun authenticate(username: String, password: String): Boolean {
        println("Authenticating user: $username")
        return username == "admin" && password == "password"
    }

    // 인증 토큰 생성
    fun generateToken(username: String): String {
        println("Generating token for: $username")
        return "token_$username"
    }

    // 토큰 유효성 검증
    fun validateToken(token: String): Boolean {
        println("Validating token: $token")
        return token.startsWith("token_")
    }
}
```

### 응집도를 높이는 방법

1. 단일 책임 원칙 적용: 각 모듈이 하나의 책임만 가지도록 설계
2. 관련 기능 그룹화: 서로 관련된 기능들을 같은 모듈에 배치
3. 불필요한 기능 제거: 모듈의 주요 목적과 관련 없는 기능은 다른 모듈로 이동
4. 명확한 추상화: 모듈의 목적을 명확히 정의하고 그에 맞는 기능만 포함
