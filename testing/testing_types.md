# testing types

- [testing types](#testing-types)
    - [테스트의 종류](#테스트의-종류)
    - [1. 단위 테스트 (Unit Testing)](#1-단위-테스트-unit-testing)
    - [2. 통합 테스트 (Integration Testing)](#2-통합-테스트-integration-testing)
    - [3. 기능 테스트 (Functional Testing)](#3-기능-테스트-functional-testing)
    - [4. 인수 테스트 (Acceptance Testing)](#4-인수-테스트-acceptance-testing)
    - [5. 성능 테스트 (Performance Testing)](#5-성능-테스트-performance-testing)

## 테스트의 종류

소프트웨어 테스팅에는 여러 유형이 있으며, 각각 다른 목적과 범위를 가지고 있습니다.

## 1. 단위 테스트 (Unit Testing)

목적: 개별 컴포넌트(함수, 메서드, 클래스 등)의 기능을 검증합니다.
범위: 가장 작은 테스트 단위로, 일반적으로 하나의 함수나 메서드를 대상으로 합니다.
특징: 외부 의존성을 최소화하거나 모의 객체(mock)를 사용합니다.

예시 (Go):

```go
func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d; want 5", result)
    }
}
```

## 2. 통합 테스트 (Integration Testing)

목적: 여러 컴포넌트나 모듈이 함께 올바르게 작동하는지 검증합니다.
범위: 두 개 이상의 모듈이나 서브시스템 간의 상호작용을 테스트합니다.
특징: 실제 의존성을 사용하거나, 더 복잡한 모의 객체를 사용할 수 있습니다.

예시 (Go):

```go
type UserService struct {
    DB *sql.DB
}

func (s *UserService) CreateUser(name string) error {
    _, err := s.DB.Exec("INSERT INTO users (name) VALUES (?)", name)
    return err
}

func TestCreateUser(t *testing.T) {
    // 실제 테스트 데이터베이스 연결
    db, err := sql.Open("mysql", "user:password@/testdb")
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    defer db.Close()

    service := &UserService{DB: db}
    err = service.CreateUser("John Doe")
    if err != nil {
        t.Errorf("Failed to create user: %v", err)
    }

    // 데이터베이스에서 사용자 확인
    var name string
    err = db.QueryRow("SELECT name FROM users WHERE name = ?", "John Doe").Scan(&name)
    if err != nil {
        t.Errorf("User not found in database: %v", err)
    }
    if name != "John Doe" {
        t.Errorf("Got name %s, want John Doe", name)
    }
}
```

## 3. 기능 테스트 (Functional Testing)

목적: 소프트웨어의 기능적 요구사항을 검증합니다.
범위: 전체 기능이나 사용 사례를 대상으로 합니다.
특징: 사용자의 관점에서 테스트하며, UI를 통한 테스트가 포함될 수 있습니다.

예시 (Python with Selenium):

```python
from selenium import webdriver
from selenium.webdriver.common.by import By

def test_login_functionality():
    driver = webdriver.Chrome()
    driver.get("https://example.com/login")

    username_input = driver.find_element(By.ID, "username")
    password_input = driver.find_element(By.ID, "password")
    login_button = driver.find_element(By.ID, "login-button")

    username_input.send_keys("testuser")
    password_input.send_keys("password123")
    login_button.click()

    # 로그인 후 대시보드로 리다이렉션되었는지 확인
    assert "dashboard" in driver.current_url

    driver.quit()
```

## 4. 인수 테스트 (Acceptance Testing)

목적: 소프트웨어가 비즈니스 요구사항과 사용자의 요구를 충족하는지 검증합니다.
범위: 전체 시스템을 대상으로 하며, 실제 사용 시나리오를 테스트합니다.
특징: 종종 클라이언트나 제품 소유자가 참여하며, 사용자 스토리나 시나리오 기반으로 수행됩니다.

예시 (Cucumber with Ruby):

```gherkin
Feature: User Registration

Scenario: Successful user registration
  Given I am on the registration page
  When I enter "johndoe@example.com" as email
  And I enter "password123" as password
  And I enter "password123" as password confirmation
  And I click the "Register" button
  Then I should see a welcome message "Welcome, John Doe!"
  And I should receive a confirmation email
```

```ruby
Given("I am on the registration page") do
  visit "/register"
end

When("I enter {string} as email") do |email|
  fill_in "Email", with: email
end

When("I enter {string} as password") do |password|
  fill_in "Password", with: password
end

When("I enter {string} as password confirmation") do |password|
  fill_in "Password confirmation", with: password
end

When("I click the {string} button") do |button_text|
  click_button button_text
end

Then("I should see a welcome message {string}") do |message|
  expect(page).to have_content(message)
end

Then("I should receive a confirmation email") do
  # 이메일 수신 확인 로직
end
```

## 5. 성능 테스트 (Performance Testing)

목적: 시스템의 성능, 응답성, 안정성을 평가합니다.
범위: 특정 기능이나 전체 시스템의 성능을 테스트합니다.
특징: 부하 테스트, 스트레스 테스트, 확장성 테스트 등이 포함됩니다.

예시 (사용 Apache JMeter):

```xml
<?xml version="1.0" encoding="UTF-8"?>
<jmeterTestPlan version="1.2" properties="5.0" jmeter="5.4.1">
  <hashTree>
    <TestPlan guiclass="TestPlanGui" testclass="TestPlan" testname="Test Plan" enabled="true">
      <stringProp name="TestPlan.comments"></stringProp>
      <boolProp name="TestPlan.functional_mode">false</boolProp>
      <boolProp name="TestPlan.tearDown_on_shutdown">true</boolProp>
      <boolProp name="TestPlan.serialize_threadgroups">false</boolProp>
      <elementProp name="TestPlan.user_defined_variables" elementType="Arguments" guiclass="ArgumentsPanel" testclass="Arguments" testname="User Defined Variables" enabled="true">
        <collectionProp name="Arguments.arguments"/>
      </elementProp>
      <stringProp name="TestPlan.user_define_classpath"></stringProp>
    </TestPlan>
    <hashTree>
      <ThreadGroup guiclass="ThreadGroupGui" testclass="ThreadGroup" testname="Thread Group" enabled="true">
        <stringProp name="ThreadGroup.on_sample_error">continue</stringProp>
        <elementProp name="ThreadGroup.main_controller" elementType="LoopController" guiclass="LoopControlPanel" testclass="LoopController" testname="Loop Controller" enabled="true">
          <boolProp name="LoopController.continue_forever">false</boolProp>
          <stringProp name="LoopController.loops">100</stringProp>
        </elementProp>
        <stringProp name="ThreadGroup.num_threads">50</stringProp>
        <stringProp name="ThreadGroup.ramp_time">10</stringProp>
        <boolProp name="ThreadGroup.scheduler">false</boolProp>
        <stringProp name="ThreadGroup.duration"></stringProp>
        <stringProp name="ThreadGroup.delay"></stringProp>
        <boolProp name="ThreadGroup.same_user_on_next_iteration">true</boolProp>
      </ThreadGroup>
      <hashTree>
        <HTTPSamplerProxy guiclass="HttpTestSampleGui" testclass="HTTPSamplerProxy" testname="HTTP Request" enabled="true">
          <elementProp name="HTTPsampler.Arguments" elementType="Arguments" guiclass="HTTPArgumentsPanel" testclass="Arguments" testname="User Defined Variables" enabled="true">
            <collectionProp name="Arguments.arguments"/>
          </elementProp>
          <stringProp name="HTTPSampler.domain">example.com</stringProp>
          <stringProp name="HTTPSampler.port"></stringProp>
          <stringProp name="HTTPSampler.protocol"></stringProp>
          <stringProp name="HTTPSampler.contentEncoding"></stringProp>
          <stringProp name="HTTPSampler.path">/api/users</stringProp>
          <stringProp name="HTTPSampler.method">GET</stringProp>
          <boolProp name="HTTPSampler.follow_redirects">true</boolProp>
          <boolProp name="HTTPSampler.auto_redirects">false</boolProp>
          <boolProp name="HTTPSampler.use_keepalive">true</boolProp>
          <boolProp name="HTTPSampler.DO_MULTIPART_POST">false</boolProp>
          <stringProp name="HTTPSampler.embedded_url_re"></stringProp>
          <stringProp name="HTTPSampler.connect_timeout"></stringProp>
          <stringProp name="HTTPSampler.response_timeout"></stringProp>
        </HTTPSamplerProxy>
        <hashTree/>
      </hashTree>
    </hashTree>
  </hashTree>
</jmeterTestPlan>
```

이 외에도 보안 테스트, 사용성 테스트, 회귀 테스트 등 다양한 유형의 테스트가 있습니다. 각 테스트 유형은 소프트웨어의 다른 측면을 검증하며, 종합적인 테스트 전략은 이러한 다양한 테스트 유형을 적절히 조합하여 수립됩니다.
