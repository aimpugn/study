package proxy.springaop;

import org.aopalliance.intercept.MethodInterceptor;
import org.aopalliance.intercept.MethodInvocation;
import org.jetbrains.annotations.NotNull;
import org.jetbrains.annotations.Nullable;
import org.springframework.aop.framework.ProxyFactory;
import org.springframework.cglib.proxy.Enhancer;
import org.springframework.cglib.proxy.FixedValue;
import org.springframework.cglib.proxy.InvocationHandler;
import org.springframework.cglib.proxy.NoOp;

interface Pojo {
    void foo();

    void bar();

    Object baz();
}

class SimplePojo implements Pojo {

    @Override
    public void foo() {
        System.out.printf("\t%s#foo method\n", this.getClass().getName());
        this.bar();
    }

    @Override
    public void bar() {
        System.out.printf("\t%s#bar method\n", this.getClass().getName());
    }

    @Override
    public Object baz() {
        System.out.printf("\t%s#baz method\n", this.getClass().getName());

        return "Return from baz";
    }
}

/**
 * {@link MethodInterceptor}는 Spring AOP에서 Around Advice를 구현하는 데 사용됩니다.
 * <p>
 * 반면 {@link org.springframework.cglib.proxy.MethodInterceptor}는 Spring 내부에서 사용되는 CGLIB 라이브러리입니다.
 * CGLIB 기반 프록시에서 메서드 호출을 가로채고 추가 동작을 정의하는 데 사용됩니다.
 * JDK 동적 프록시와 달리 클래스 기반 프록시를 생성하며, 인터페이스가 아닌 구체 클래스도 프록시화할 수 있습니다.
 * <p>
 * 인터페이스 뒤에 프록시를 구현하는 JDK 동적 프록시와 달리,
 * CGLIB는 클래스 기반 프록시를 생성하므로 인터페이스가 아닌 구체 클래스도 프록시화할 수 있습니다.(프록시된 클래스의 하위 클래스를 동적으로 구현)
 *
 * @see <a href="https://jcs.ep.jhu.edu/ejava-springboot/coursedocs/content/html_single/aop-notes.html">Spring AOP and Method Proxies</a>
 */
class SimpleAdvice implements MethodInterceptor {
    /**
     * {@link MethodInvocation} 객체를 통해 메서드 호출 정보를 제공하며,
     * 원래 메서드를 호출하기 위해 {@link org.aopalliance.intercept.Joinpoint#proceed() proceed} 메서드를 사용합니다.
     *
     * @param invocation the method invocation joinpoint
     * @return {@link org.aopalliance.intercept.Joinpoint#proceed() proceed} 호출의 결과
     * @throws Throwable 메서드 호출 과정중 발생할 수 있는 익셉션
     */
    @Override
    public @Nullable Object invoke(@NotNull MethodInvocation invocation) throws Throwable {
        System.out.printf("%s#invoke() is called\n", this.getClass().getName());
        Object result = invocation.proceed();
        System.out.printf("After %s is called\n", invocation.getMethod().getName());

        return result;
    }
}

public class Main {
    /**
     * @see <a href="https://docs.spring.io/spring-framework/reference/core/aop/proxying.html">Spring Framework / Core Technologies / Aspect Oriented Programming with Spring / Proxying Mechanisms</a>
     */
    public static void testSpringAOP() {
        System.out.println("===== SimplePojo start =====");
        Pojo pojo = new SimplePojo();
        // this is a direct method call on the 'pojo' reference
        pojo.foo();
        System.out.println("===== SimplePojo end =====");

        System.out.println("===== ProxyFactory - SimplePojo start =====");
        // `ProxyFactory`는 AOP 기능을 제공하기 위해 프록시 객체를 생성하는 유틸리티 클래스입니다.
        // 다음 두 가지 방식으로 프록시를 생성합니다.
        // - 'JDK 동적 프록시':
        //    - 타겟 클래스가 하나 이상의 인터페이스를 구현한 경우
        //    - 생성된 프록시는 인터페이스만 구현하며, 구체 클래스에 의존하지 않습니다.
        // - 'CGLIB 기반 프록시':
        //    - 타겟 클래스가 인터페이스를 구현하지 않았거나, 구체 클래스를 기반으로 프록시가 필요할 경우
        //    - CGLIB은 바이트코드 조작을 통해 구체 클래스를 상속받아 프록시 객체를 생성합니다.
        ProxyFactory factory = new ProxyFactory(new SimplePojo());
        factory.addInterface(Pojo.class);
        factory.addAdvice(new SimpleAdvice());
        Pojo proxiedPojo = (Pojo) factory.getProxy();
        // 클라이언트 코드는 프록시에 대한 참조를 갖습니다.
        // 즉, `proxiedPojo` 참조에 대한 메서드 호출은, 프록시에 대한 호출입니다.
        // 그 결과, 프록시는 특정 메서드 호출과 관련된 모든 인터셉터(Advice)에게 위임할 수 있습니다.
        // 하지만, 그 호출이 결국 타겟 오브젝트인 `SimplePojo`에 도달하게 되면,
        // `this.foo()`, `this.bar()` 같이 자기 자신에게 수행하는 모든 메서드 호출은
        // 프록시가 아닌 `this` 참조에 대해 호출이 됩니다.
        // 즉, 여기서 "자신"(`this`)이라 함은 프록시가 아닌 `SimplePojo`가 됩니다.
        // 자기 자신(`this`)을 통한 메서드 호출(self-invocation)의 경우, 해당 메서드 호출에 연결된 어드바이스(Advice)가 실행되지 않음을 의미합니다.
        // - Avoid self invocation: 추천. 침습성이 가장 낮은 방법.
        // - Inject a self reference:
        // - Use AopContext.currentProxy(): 매우 비추천. 코드가 Spring AOP에 완전히 결합하게 됨.
        System.out.println("1. proxiedPojo.foo()");
        proxiedPojo.foo();
        // 1. proxiedPojo.foo()
        // proxy.SimpleAdvice#invoke() is called
        //      proxy.SimplePojo#foo method
        //      proxy.SimplePojo#bar method  <===== 이 경우 Advice가 실행되지 않습니다.
        // After foo is called
        System.out.println("2. proxiedPojo.bar()");
        proxiedPojo.bar();
        // 2. proxiedPojo.bar()
        // proxy.springaop.SimpleAdvice#invoke() is called
        //      proxy.springaop.SimplePojo#bar method
        // After bar is called
        System.out.println("3. proxiedPojo.baz()");
        proxiedPojo.baz();
        // 2. proxiedPojo.baz()
        // proxy.SimpleAdvice#invoke() is called
        // 	    proxy.SimplePojo#baz method
        // After baz is called
        System.out.println("===== ProxyFactory - SimplePojo end =====");
    }

    /**
     * CGLIB? Code Generation Library
     *
     * @see <a href="https://jcs.ep.jhu.edu/ejava-springboot/coursedocs/content/html_single/aop-notes.html#id-cglib">CGLIB</a>
     */
    public static void testCGLIBProxy() {
        Enhancer enhancerForMethodInterceptor = new Enhancer();
        enhancerForMethodInterceptor.setSuperclass(SimplePojo.class);
        // 1. org.springframework.cglib.proxy.MethodInterceptor 사용 방식
        enhancerForMethodInterceptor.setCallback((org.springframework.cglib.proxy.MethodInterceptor) (
                proxy, method, args1, methodProxy
        ) -> {
            System.out.printf("[MethodInterceptor] Before %s#%s\n", proxy.getClass().getName(), method.getName());
            Object result = methodProxy.invokeSuper(proxy, args1);
            System.out.printf("[MethodInterceptor] After %s#%s\n", proxy.getClass().getName(), method.getName());
            return result;
        });

        System.out.println("==== MethodInterceptor start ====");
        SimplePojo proxy1 = (SimplePojo) enhancerForMethodInterceptor.create();
        // 자식 클래스로 생성되므로, 이 경우 `this`는 프록시 오브젝트고,
        // 따라서 `this.foo`에서 호출하는 `this.bar`의 `this`도 프록시 오브젝트가 됩니다.
        proxy1.foo();
        // [MethodInterceptor] Before proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#foo
        // 	    proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#foo method
        // [MethodInterceptor] Before proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#bar
        //      proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#bar method
        // [MethodInterceptor] After proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#bar
        // [MethodInterceptor] After proxy.springaop.SimplePojo$$EnhancerByCGLIB$$e48d0e16#foo
        System.out.println("==== MethodInterceptor end ====");

        // 2. JDK Dynamic Proxy와 비슷한 방식
        Enhancer enhancerForInvocationHandler = new Enhancer();
        SimplePojo sp = new SimplePojo();
        enhancerForInvocationHandler.setSuperclass(SimplePojo.class);
        enhancerForInvocationHandler.setCallback((InvocationHandler) (
                proxy, method, args1
        ) -> {
            System.out.printf("[InvocationHandler] Before %s#%s\n", proxy.getClass().getName(), method.getName());

            // 이 방법의 경우 proxy 객체의 method를 호출하게 되고, 다시 콜백이 호출되어 무한 재귀 호출이 발생합니다.
            // Object result = method.invoke(obj, args1);
            //
            // 무한 재귀를 방지하기 위해 대상이 되는 객체(SimplePojo 인스턴스)를 직접 참조하여 리플렉션으로 원래 메서드를 호출합니다.
            Object result = method.invoke(sp, args1);
            System.out.printf("[InvocationHandler] After %s#%s\n", proxy.getClass().getName(), method.getName());

            return result;
        });

        System.out.println("==== InvocationHandler start ====");
        SimplePojo proxy2 = (SimplePojo) enhancerForInvocationHandler.create();
        proxy2.foo();
        // 반면 JDK 동적 프록시와 비슷하게, 여기서 `this`는 SimplePojo 인스턴스이므로,
        // `this.bar` 호출에 대한 인터셉터는 이뤄지지 않습니다.
        // [InvocationHandler] Before proxy.springaop.SimplePojo$$EnhancerByCGLIB$$8dc07f5e#foo
        //      proxy.springaop.SimplePojo#foo method
        //      proxy.springaop.SimplePojo#bar method
        // [InvocationHandler] After proxy.springaop.SimplePojo$$EnhancerByCGLIB$$8dc07f5e#foo
        System.out.println("==== InvocationHandler end ====");

        // 3. `FixedValue` 사용하여 항상 고정된 값을 리턴하도록 할 수도 있습니다.
        Enhancer enhancerForFixedValue = new Enhancer();
        enhancerForFixedValue.setSuperclass(SimplePojo.class);
        enhancerForFixedValue.setCallback((FixedValue) () -> "Fixed Value");

        System.out.println("==== FixedValue start ====");
        SimplePojo proxy3 = (SimplePojo) enhancerForFixedValue.create();
        Object proxy3Result = proxy3.baz();
        System.out.printf("\tproxy3Result: %s\n", proxy3Result);
        // proxy3Result: Fixed Value
        System.out.println("==== FixedValue end ====");

        // 4. super, 즉 `SimplePojo` 메서드를 직접 호출합니다.
        Enhancer enhancerForNoOp = new Enhancer();
        enhancerForNoOp.setSuperclass(SimplePojo.class);
        enhancerForNoOp.setCallback(NoOp.INSTANCE);

        System.out.println("==== NoOp start ====");
        SimplePojo proxy4 = (SimplePojo) enhancerForNoOp.create();
        proxy4.foo();
        //  proxy.springaop.SimplePojo$$EnhancerByCGLIB$$7a488dba#foo method
        //  proxy.springaop.SimplePojo$$EnhancerByCGLIB$$7a488dba#bar method
        System.out.println("==== NoOp end ====");
    }

    public static void main(String[] args) {
        System.out.println("======== testProxyPojo ========");
        testSpringAOP();
        System.out.println("======== testCGLIBProxy ========");
        testCGLIBProxy();
    }
}
