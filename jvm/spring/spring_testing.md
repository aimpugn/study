# Spring Testing

- [Spring Testing](#spring-testing)
    - [테스트 종류](#테스트-종류)
        - [유닛 테스트](#유닛-테스트)
        - [통합 테스트](#통합-테스트)
        - [E2E 테스트](#e2e-테스트)
    - [라이브러리](#라이브러리)
        - [`spring-boot-test` vs `spring-boot-starter-test`](#spring-boot-test-vs-spring-boot-starter-test)
    - [예제](#예제)
        - [`@Component`의 유닛 테스트](#component의-유닛-테스트)
        - [`@Service`의 유닛 테스트](#service의-유닛-테스트)
        - [중첩 테스트 클래스와 외부 클래스의 공유 불가](#중첩-테스트-클래스와-외부-클래스의-공유-불가)
        - [`Service`, `Component`, `Mapper` 구조에 대한 테스트](#service-component-mapper-구조에-대한-테스트)
            - [`SomeService`에 대한 테스트](#someservice에-대한-테스트)
            - [`AComponent`에 대한 테스트](#acomponent에-대한-테스트)
            - [`BComponent`에 대한 테스트](#bcomponent에-대한-테스트)
        - [외부 통신에 대한 테스트](#외부-통신에-대한-테스트)

## 테스트 종류

### 유닛 테스트

```java
@ExtendWith(MockitoExtension.class)
class UserServiceUnitTest {
    @Mock
    private UserMapper userMapper;

    @InjectMocks
    private UserService userService;

    @Test
    void findUserById() {
        User mockUser = new User(1L, "user@example.com");
        Mockito.when(userMapper.findById(1L)).thenReturn(Optional.of(mockUser));

        Optional<User> user = userService.findUserById(1L);

        assertTrue(user.isPresent());
        assertEquals("user@example.com", user.get().getEmail());
    }
}
```

가장 빠르게 수행할 수 있고, 리팩토링할 때의 안전망이 됩니다.
DB 같은 외부 자원을 직접 사용하지 않아야 하고, 빠르고 독립적이어야 합니다.

`@Component`, `@Service`, `@Mapper` 등 개별 로직을 검증합니다.
이때 최소한의 의존성만 사용합니다.
Unit 5, Mockito(외부 의존성을 Mocking), MyBatis Test 등을 사용합니다.

### 통합 테스트

```java
@MybatisTest
@AutoConfigureTestDatabase(replace = Replace.NONE)
class UserMapperIntegrationTest {
    @Autowired
    private UserMapper userMapper;

    @Test
    void insertAndFindById() {
        // 실제 DB를 통해 데이터를 삽입하고 조회하여 Mapper와 DB 스키마가 정확히 일치하는지 검증합니다
        User user = new User(null, "user@example.com");

        userMapper.insert(user);
        assertNotNull(user.getId());

        Optional<User> found = userMapper.findById(user.getId());
        assertTrue(found.isPresent());
        assertEquals("user@example.com", found.get().getEmail());
    }
}
```

`@Service`가 여러 `@Component`와 협력하여 비즈니스 로직을 제대로 수행하는지, `@Component`가 *실제 `@Mapper`와 연결된 DB 접근*을 포함하여 데이터 접근 로직이 의도한 대로 동작하는지 등을 검증합니다.

통합 테스트에서 Docker 기반 DB를 사용하거나 H2 같은 인메모리 DB를 사용합니다.

```java
@Testcontainers
@MybatisTest
class UserMapperIntegrationTest {

    @Container
    static MySQLContainer<?> mysql = new MySQLContainer<>("mysql:8.0.28")
                                        .withDatabaseName("testdb")
                                        .withUsername("user")
                                        .withPassword("password");

    @DynamicPropertySource
    static void databaseProperties(DynamicPropertyRegistry registry) {
        registry.add("spring.datasource.url", mysql::getJdbcUrl);
        registry.add("spring.datasource.username", mysql::getUsername);
        registry.add("spring.datasource.password", mysql::getPassword);
    }

    @Autowired
    private UserMapper userMapper;

    @Test
    void insertAndFindUser() {
        User user = new User(null, "user@test.com");
        userMapper.insert(user);

        Optional<User> found = userMapper.findById(user.getId());
        assertTrue(found.isPresent());
        assertEquals("user@test.com", found.get().getEmail());
    }
}
```

### E2E 테스트

외부에서 요청을 하는 것처럼 전체 계층을 테스트합니다.
실제 운영환경 시나리오 기반 통합 검증합니다.

MockMvc, Spring Boot Test 등을 사용합니다.

```java
@ExtendWith(SpringExtension.class)
@ContextConfiguration(classes = {AppConfig.class, WebConfig.class})
@WebAppConfiguration
public class UserControllerTest {
    @Autowired
    private WebApplicationContext context;

    // MockMvc를 통해 실제 `DispatcherServlet` 동작을 시뮬레이션합니다.
    // 이를 통해 Servlet 환경에서 HTTP 요청/응답 테스트가 가능합니다.
    private MockMvc mockMvc;

    @BeforeEach
    void setup() {
        this.mockMvc = MockMvcBuilders
            .webAppContextSetup(context)
            .build();
    }

    @Test
    void testGetUsers() throws Exception {
        mockMvc.perform(get("/users").accept(MediaType.APPLICATION_JSON))
            .andExpect(status().isOk())
            .andExpect(content().contentType(MediaType.APPLICATION_JSON));
    }
}
```

## 라이브러리

### `spring-boot-test` vs `spring-boot-starter-test`

- `spring-boot-test`
    - Spring Boot 테스트의 핵심 기능만 포함하는 모듈로, 기본적인 Spring Boot 테스트 기능을 제공합니다.
    - `@SpringBootTest`와 같은 테스트 어노테이션을 제공합니다.

- [`spring-boot-starter-test`](https://docs.spring.io/spring-boot/reference/testing/test-scope-dependencies.html#testing.test-scope-dependencies)
    - 통합 테스트 패키지로, 여러 테스트 라이브러리를 한 번에 가져오는 "스타터 팩"입니다.
    - `spring-boot-test` 및 다양한 테스트 라이브러리를 함께 제공합니다.
        - JUnit (JUnit 5 및 JUnit 4 호환을 위한 빈티지 엔진)
        - Spring Test & Spring Boot Test
        - AssertJ
        - Hamcrest
        - Mockito
        - JSONassert
        - JsonPath

Spring Boot 경우 `spring-boot-starter-test`만 의존성에 추가하면 테스트에 필요한 대부분의 라이브러리를 사용할 수 있습니다.

## 예제

### `@Component`의 유닛 테스트

```java
@Component
class UserComponent {
    private final UserMapper userMapper;

    public UserComponent(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    public Optional<User> findById(Long id) {
        return userMapper.findById(id);
    }
}
```

Mock 객체를 이용해 Mapper의 결과를 모의(Mock) 처리합니다.

```java
@ExtendWith(MockitoExtension.class)
class UserComponentTest {

    @Mock
    private UserMapper userMapper;

    @InjectMocks
    private UserComponent userComponent;

    @Test
    void findById_shouldReturnUser_whenExists() {
        Long userId = 1L;
        User mockUser = new User(userId, "user@example.com");

        when(userMapper.findById(userId)).thenReturn(Optional.of(mockUser));

        Optional<User> result = userComponent.findById(userId);

        assertTrue(result.isPresent());
        assertEquals("user@example.com", result.get().getEmail());
    }
}
```

### `@Service`의 유닛 테스트

여러 Component들의 협력과 비즈니스 로직 정확성을 검증합니다.
Component 의존성은 Mock으로 처리합니다.

```java
@ExtendWith(MockitoExtension.class)
class UserServiceTest {

    @Mock
    private UserComponent userComponent;

    @InjectMocks
    private UserService userService;

    @Test
    void deactivateUser_shouldMarkUserInactive() {
        Long userId = 1L;
        User mockUser = new User(userId, "user@example.com");
        mockUser.setActive(true);

        when(userComponent.findById(userId)).thenReturn(Optional.of(mockUser));

        userService.deactivateUser(userId);

        assertFalse(mockUser.isActive());
        verify(userComponent).updateUser(mockUser);
    }
}
```

### 중첩 테스트 클래스와 외부 클래스의 공유 불가

```java
package com.example.testing.controller;

import com.example.testing.AppConfig;
import com.example.testing.mapper.UserMapper;
import com.example.testing.model.User;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.BDDMockito;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.mock.mockito.MockBean;
import org.springframework.http.MediaType;
import org.springframework.test.annotation.DirtiesContext;
import org.springframework.test.context.ContextConfiguration;
import org.springframework.test.context.junit.jupiter.SpringExtension;
import org.springframework.test.context.web.WebAppConfiguration;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;
import org.springframework.web.context.WebApplicationContext;

import javax.sql.DataSource;
import java.util.List;

import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

@ExtendWith(SpringExtension.class)
@WebAppConfiguration
@ContextConfiguration(classes = AppConfig.class)
public class UserControllerTest {
    @Autowired
    WebApplicationContext context;
    MockMvc mockMvc;

    @Nested
    class UsingDb {
        @Autowired
        DataSource dataSource;

        @BeforeEach
        void setup() throws Exception {
            try (var conn = dataSource.getConnection();
                 var stmt = conn.createStatement()) {
                stmt.execute("CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY AUTO_INCREMENT, name VARCHAR(100));");
                stmt.execute("DELETE FROM users");
                stmt.execute("INSERT INTO users(name) VALUES ('hello'), ('world');");
            }

            mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
        }

        @Test
        @DirtiesContext(methodMode = DirtiesContext.MethodMode.AFTER_METHOD)
        void testGetUsersUsingDb() throws Exception {
            mockMvc.perform(get("/users").accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(result ->
                    System.out.println("DB 응답: " + result.getResponse().getContentAsString()));
        }
    }

    @Nested
    class UsingMock {
        @MockBean
        UserMapper userMapper;

        @BeforeEach
        void setup() {
            mockMvc = MockMvcBuilders.webAppContextSetup(context).build();
            BDDMockito.given(userMapper.findAll()).willReturn(List.of(
                new User(1, "a"),
                new User(2, "b"),
                new User(3, "c")
            ));
        }

        @Test
        @DirtiesContext(methodMode = DirtiesContext.MethodMode.AFTER_METHOD)
        void testGetUsersUsingMockito() throws Exception {
            mockMvc.perform(get("/users").accept(MediaType.APPLICATION_JSON))
                .andExpect(status().isOk())
                .andExpect(result ->
                    System.out.println("Mock 응답: " + result.getResponse().getContentAsString()));
        }

    }

}
```

`@Nested` 클래스가 상위 클래스(`UserControllerTest`)의 필드를 "상속처럼 공유"하지 않습니다.
반면, 상위 클래스의 설정(`@WebAppConfiguration`, `@ContextConfiguration`, `@ExtendWith`) 등은 `@Nested`(중첩) 클래스에 상속됩니다.

따라서 빈 주입(`@Autowired`)은 `@Nested` 클래스마다 별도의 인스턴스를 주입해야 합니다.

```java
@ExtendWith(SpringExtension.class)
@WebAppConfiguration
@ContextConfiguration(classes = AppConfig.class)
public class UserControllerTest {
    // Nested 클래스마다 독립적으로 주입되어야 합니다.
    @Autowired
    WebApplicationContext context;

    @Nested
    class UsingDb {
        // 상위의 context, mockMvc를 공유할 수 없음 (별도의 독립 인스턴스)
    }

    @Nested
    class UsingMock {
        // 상위의 context, mockMvc를 공유할 수 없음 (별도의 독립 인스턴스)
    }
```

JUnit 5의 `@Nested` 테스트 클래스는 실제로는 독립적인 테스트 인스턴스입니다.
`UsingDb`와 `UsingMock` 클래스는 각각 별도의 인스턴스로 생성되며, 필요한 필드는 각각의 클래스 내부에서 `@Autowired`로 독립적으로 주입되어야 합니다.

### `Service`, `Component`, `Mapper` 구조에 대한 테스트

```java

import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import org.springframework.stereotype.Service;

import java.io.BufferedWriter;
import java.io.FileOutputStream;
import java.io.IOException;
import java.io.OutputStreamWriter;
import java.nio.charset.StandardCharsets;
import java.util.HashMap;
import java.util.HashSet;

interface ServiceBase<IN, OUT> {

}

interface ComponentBase<IN, OUT> {

}

@Mapper
interface ThatMapper {
    HashMap<String, String> findAll(@Param("some_key") String someKey);
}

@Mapper
interface ThisMapper {
    HashMap<String, String> findById(@Param("some_key") String someKey);
}

@Mapper
interface OtherMapper {

}

@Service
class SomeService implements ServiceBase<String, HashMap<String, String>> {
    @Autowired
    AComponent aComponent;
    @Autowired
    BComponent bComponent;

    public String doSomething(String key) {
        var componentResult = aComponent.doComplexThing(key);
        try {
            bComponent.generateFile(componentResult, key);
        } catch (Exception e) {
            e.printStackTrace();
            return "Error occurred";
        }

        return componentResult.toUpperCase(); // 예시 로직
    }
}

@Component
class AComponent {
    private final ThatMapper thatMapper;
    private final WebClient webClient;

    public AComponent(ThatMapper thatMapper, WebClient webClient) {
        this.thatMapper = thatMapper;
        this.webClient = webClient;
    }

    public String doComplexThing(String someKey) {
        HashMap<String, String> dbData = thatMapper.findAll(someKey);

        StringBuilder sb = new StringBuilder();
        dbData.forEach((key, value) -> {
            String externalData = webClient.get()
                .uri("/external/api/" + key)
                .retrieve()
                .bodyToMono(String.class)
                .block();
            sb.append(value).append(":").append(externalData).append(";");
        });

        return sb.toString();
    }
}

@Component
class BComponent implements ComponentBase<String, HashSet<String>> {
    ThisMapper thisMapper;
    OtherMapper otherMapper;

    public void generateFile(String content, String key) throws IOException {
        var row = thisMapper.findById(key);
        try (var writer = new BufferedWriter(new OutputStreamWriter(
            new FileOutputStream(row.get("some_file_path")), StandardCharsets.UTF_8))) {
            writer.write(content);
        }
        // 또는
        // var file = Paths.get(row.get("some_file_path"));
        // Files.write(file, key.getBytes(StandardCharsets.UTF_8));
    }
}
```

JUnit(`junit-jupiter-api`, `junit-jupiter-engine`), Mockito(`mockito-core`, `mockito-junit-jupiter`) 등을 사용합니다.

#### `SomeService`에 대한 테스트

`SomeService`는 `AComponent`, `BComponent` 컴포넌트를 사용하므로, 두 컴포넌트를 모두 Mock으로 처리합니다.

```java
package com.example.testing;

import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class SomeServiceTest {
    @Mock
    AComponent aComponent;

    @Mock
    BComponent bComponent;

    @InjectMocks
    SomeService someService;

    @Test
    void doSomething_successfulScenario() throws Exception {
        // given
        String inputKey = "some-key";
        String componentResult = "result from a component";

        when(aComponent.doComplexThing(inputKey)).thenReturn(componentResult);
        doNothing().when(bComponent).generateFile(componentResult, inputKey);

        // when
        String result = someService.doSomething(inputKey);

        // then
        assertEquals(componentResult.toUpperCase(), result);
        verify(aComponent).doComplexThing(inputKey);
        verify(bComponent).generateFile(componentResult, inputKey);
    }

    @Test
    void doSomething_bComponentThrowsException() throws Exception {
        // given
        String inputKey = "some-key";
        String componentResult = "result from a component";

        when(aComponent.doComplexThing(inputKey)).thenReturn(componentResult);
        doThrow(new RuntimeException("I/O Exception")).when(bComponent).generateFile(componentResult, inputKey);

        // when
        String result = someService.doSomething(inputKey);

        // then
        assertEquals("Error occurred", result);
        verify(aComponent).doComplexThing(inputKey);
        verify(bComponent).generateFile(componentResult, inputKey);
    }
}
```

#### `AComponent`에 대한 테스트

`AComponent`는 `Mapper`와 `WebClient`라는 외부 자원을 사용하므로 둘 다 Mock 처리합니다.

```java
package com.example.testing;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.web.reactive.function.client.WebClient;
import reactor.core.publisher.Mono;

import java.util.HashMap;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.anyString;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AComponentTest {

    @Mock
    ThatMapper thatMapper;

    @Mock
    WebClient webClient;

    @Mock
    WebClient.RequestHeadersUriSpec requestHeadersUriSpec;

    @Mock
    WebClient.RequestHeadersSpec requestHeadersSpec;

    @Mock
    WebClient.ResponseSpec responseSpec;

    AComponent aComponent;

    @BeforeEach
    void setup() {
        aComponent = new AComponent(thatMapper, webClient);
    }

    @Test
    void doComplexThing_successful() {
        // given
        String inputKey = "test-key";
        HashMap<String, String> mapperResult = new HashMap<>();
        mapperResult.put("key1", "value1");
        mapperResult.put("key2", "value2");

        when(thatMapper.findAll(inputKey)).thenReturn(mapperResult);

        when(webClient.get()).thenReturn(requestHeadersUriSpec);
        when(requestHeadersUriSpec.uri(anyString())).thenReturn(requestHeadersSpec);
        when(requestHeadersSpec.retrieve()).thenReturn(responseSpec);
        when(responseSpec.bodyToMono(String.class))
                .thenReturn(Mono.just("externalData1"), Mono.just("externalData2"));

        // when
        String result = aComponent.doComplexThing(inputKey);

        // then
        String expected = "value1:externalData1;value2:externalData2;";
        assertEquals(expected, result);

        verify(thatMapper).findAll(inputKey);
        verify(webClient, times(2)).get();
    }
}
```

#### `BComponent`에 대한 테스트

`Mapper`와 파일 시스템이라는 외부 자원을 사용합니다.
하지만 이때 파일시스템 접근을 완전히 배제하려면, File 생성 로직 분리이 필요합니다.

```java
interface FileWriter {
    void write(String filePath, String content) throws IOException;
}

@Component
class FileWriterImpl implements FileWriter {
    @Override
    public void write(String filePath, String content) throws IOException {
        try (BufferedWriter writer = new BufferedWriter(
                new OutputStreamWriter(new FileOutputStream(filePath), StandardCharsets.UTF_8))) {
            writer.write(content);
        }
        // 또는
        // var file = Paths.get(row.get("some_file_path"));
        // Files.write(file, key.getBytes(StandardCharsets.UTF_8));
    }
}
```

그리고 `FileWriter`를 컴포넌트로 주입하여 사용합니다.

```java
@Component
class BComponent implements ComponentBase<String, HashSet<String>> {
    private final ThisMapper thisMapper;
    private final OtherMapper otherMapper;
    private final FileWriter fileWriter;

    public BComponent(ThisMapper thisMapper, OtherMapper otherMapper, FileWriter fileWriter) {
        this.thisMapper = thisMapper;
        this.otherMapper = otherMapper;
        this.fileWriter = fileWriter;
    }

    public void generateFile(String content, String key) throws IOException {
        var row = thisMapper.findById(key);
        fileWriter.write(row.get("some_file_path"), content);
    }
}
```

`Mapper`를 Mocking하고, 파일시스템 접근을 최소화하도록 로컬 파일 시스템에서 접근 가능한 임시 파일을 생성하는 방식으로 테스트할 수 있습니다.

```java
@ExtendWith(MockitoExtension.class)
class BComponentTest {

    @Mock
    ThisMapper thisMapper;

    @Mock
    OtherMapper otherMapper;

    @Mock
    FileWriter fileWriter; // 파일시스템을 Mock 처리

    BComponent bComponent;

    @BeforeEach
    void setup() {
        bComponent = new BComponent(thisMapper, otherMapper, fileWriter);
    }

    @Test
    void generateFile_successful() throws IOException {
        // given
        String content = "test-content";
        String key = "test-key";
        String filePath = "/tmp/testfile.txt";

        HashMap<String, String> mapperResult = new HashMap<>();
        mapperResult.put("some_file_path", filePath);

        when(thisMapper.findById(key)).thenReturn(mapperResult);

        // when
        bComponent.generateFile(content, key);

        // then
        verify(thisMapper).findById(key);
        verify(fileWriter).write(filePath, content); // 파일 시스템 직접 접근이 없음
    }

    @Test
    void generateFile_whenMapperReturnsNoPath_shouldThrowException() {
        // given
        String content = "test-content";
        String key = "test-key";
        when(thisMapper.findById(key)).thenReturn(new HashMap<>());

        // when, then
        Assertions.assertThrows(NullPointerException.class, () -> {
            bComponent.generateFile(content, key);
        });

        verify(thisMapper).findById(key);
        verifyNoInteractions(fileWriter);
    }
}
```

### 외부 통신에 대한 테스트

테스트에서는 *외부 호출 자체는 Mock 처리*하고, `RestTemplate`의 요청과 응답에 대한 처리를 확인해야 합니다.

```java
@Component
public class ExternalApiClient {

    private final RestTemplate restTemplate;

    public ExternalApiClient(RestTemplate restTemplate) {
        this.restTemplate = restTemplate;
    }

    public UserData fetchUserData(String userId) {
        String url = "http://external-service.com/api/users/{id}";

        // 요청 바디 예시 (POST 요청의 경우, 본 예시는 GET이라 요청 바디는 없지만 설명 목적상 예시 제공)
        HttpEntity<RequestPayload> requestEntity = new HttpEntity<>(
            new RequestPayload(userId),
            createHeaders()
        );

        ResponseEntity<UserData> response = restTemplate.exchange(
            url,
            HttpMethod.POST,
            requestEntity,
            UserData.class,
            userId
        );

        return response.getBody();
    }

    private HttpHeaders createHeaders() {
        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.set("X-API-KEY", "some-api-key");
        return headers;
    }

    // DTO 클래스 예시
    public static class RequestPayload {
        private String id;

        public RequestPayload(String id) {
            this.id = id;
        }

        public String getId() {
            return id;
        }

        public void setId(String id) {
            this.id = id;
        }
    }

    public static class UserData {
        private String id;
        private String name;
        // getters, setters 생략
    }
}
```

이렇게 외부 서비스와 통신하는 경우, 다음과 같은 케이스들에 대한 테스트를 수행할 수 있습니다:
- URL이 올바른지
- HTTP 메서드(POST, GET 등)가 올바른지
- 헤더가 제대로 설정되는지
- 요청 본문(Request Body)의 내용이 정확한지

```java
@ExtendWith(MockitoExtension.class)
class ExternalApiClientTest {

    @Mock
    RestTemplate restTemplate;

    ExternalApiClient externalApiClient;

    @BeforeEach
    void setup() {
        externalApiClient = new ExternalApiClient(restTemplate);
    }

    @InjectMocks
    ExternalApiClient externalApiClient;

    @Test
    void someTest() {
        when(restTemplate.exchange(
            eq("url"),
            eq(HttpMethod.GET),
            any(HttpEntity.class),
            eq(Response.class))
        ).thenReturn(new ResponseEntity<>(new Response(), HttpStatus.OK));

        externalApiClient.callApi();

        verify(restTemplate).exchange(
            eq("url"),
            eq(HttpMethod.GET),
            any(HttpEntity.class),
            eq(Response.class));
    }

    @Test
    void fetchUserData_successfulRequest() {
        // given
        String userId = "123";
        String expectedUrl = "http://external-service.com/api/users/{id}";

        ExternalApiClient.UserData mockUserData = new ExternalApiClient.UserData();
        mockUserData.setId("123");
        mockUserData.setName("John Doe");

        ResponseEntity<ExternalApiClient.UserData> responseEntity =
                new ResponseEntity<>(mockUserData, HttpStatus.OK);

        // RestTemplate.exchange 호출 시 예상되는 응답 설정
        when(restTemplate.exchange(
                eq(expectedUrl),
                eq(HttpMethod.POST),
                any(HttpEntity.class),
                eq(ExternalApiClient.UserData.class),
                eq(userId))
        ).thenReturn(responseEntity);

        // when
        ExternalApiClient.UserData result = externalApiClient.fetchUserData(userId);

        // then
        assertNotNull(result);
        assertEquals("123", result.getId());
        assertEquals("John Doe", result.getName());

        // `ArgumentCaptor`를 사용하여 실제로 만들어진 요청을 상세히 들여다보고 검증할 수 있습니다.
        ArgumentCaptor<HttpEntity<ExternalApiClient.RequestPayload>> requestCaptor =
                ArgumentCaptor.forClass(HttpEntity.class);

        // verify(mock객체).호출할메서드(인자);
        // - Mock 객체의 특정 메서드가 호출되었는지 확인하는 기능입니다.
        // - 호출된 횟수도 확인할 수 있습니다.
        //
        // 특정 인자의 실제 값이 중요하지 않을 경우, 다음과 같이 `any`를 사용하여 호출되는 메서드의 인자값을 구체적인 값과 관계없이 어떤 값이든 받아들일 수 있습니다.
        // ```
        // verify(restTemplate).exchange(
        //     any(String.class), // 문자열이면 어떤 값이든
        //     eq(HttpMethod.GET),
        //     any(HttpEntity.class),
        //     eq(Response.class)
        // );
        // ```
        verify(restTemplate).exchange(
                eq(expectedUrl),
                eq(HttpMethod.POST),
                requestCaptor.capture(), // 실제 요청을 캡처
                eq(ExternalApiClient.UserData.class),
                eq(userId)
        ); // 지정한 인자들과 함께 한 번 호출되었는지 확인

        HttpEntity<ExternalApiClient.RequestPayload> capturedRequest = requestCaptor.getValue();

        // 요청 헤더 검증
        assertNotNull(capturedRequest.getHeaders());
        assertEquals(MediaType.APPLICATION_JSON, capturedRequest.getHeaders().getContentType());
        assertEquals("some-api-key", capturedRequest.getHeaders().getFirst("X-API-KEY"));

        // 요청 바디 검증
        assertNotNull(capturedRequest.getBody());
        assertEquals(userId, capturedRequest.getBody().getId());
    }
}
```

만약 나중에 HTTP 클라이언트 구현체를 `RestTemplate`에서 `WebClient` 등으로 바꾸더라도 테스트 코드가 변경에 영향을 받지 않을 뿐더러, 로직 정확성을 유지하는 데 도움을 줍니다.

```java
@ExtendWith(MockitoExtension.class)
class ExternalApiClientTest {


}
```
