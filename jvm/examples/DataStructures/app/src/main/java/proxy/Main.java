package proxy;

import utils.ConsoleColors;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.Proxy;

// 프록시할 인터페이스
interface MyInterface {
    void doSomething();

    String getSomething();
}

// 실제 구현 클래스
class MyImplementation implements MyInterface {
    @Override
    public void doSomething() {
        System.out.println("Doing something");
    }

    @Override
    public String getSomething() {
        return "Something";
    }
}

// InvocationHandler 구현
class LoggingInvocationHandler implements InvocationHandler {
    private final Object target;

    public LoggingInvocationHandler(Object target) {
        this.target = target;
    }

    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        System.out.println(ConsoleColors.GREEN + "Before method: " + method.getName() + ConsoleColors.RESET);
        Object result = method.invoke(target, args);
        System.out.println(ConsoleColors.GREEN + "After method: " + method.getName() + ", result: " + result + ConsoleColors.RESET);
        return result;
    }
}

/**
 * MyInterface의 모든 메서드 호출을 가로채고 로깅하는 프록시를 생성합니다.
 * 동적 프록시는 런타임에 인터페이스를 구현하는 클래스를 생성하여 메서드 호출을 중개합니다.
 */
public class Main {
    public static void main(String[] args) {
        MyInterface realObject = new MyImplementation();

        MyInterface proxyObject = (MyInterface) Proxy.newProxyInstance(
            MyInterface.class.getClassLoader(),
            new Class<?>[]{MyInterface.class},
            new LoggingInvocationHandler(realObject)
        );

        // 프록시 객체를 통한 메서드 호출
        proxyObject.doSomething();
        String result = proxyObject.getSomething();
        System.out.println("Main got result: " + result);
    }
}