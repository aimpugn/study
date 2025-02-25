package springframework.annoconfig.config;

import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.config.BeanDefinition;
import org.springframework.beans.factory.config.BeanFactoryPostProcessor;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;
import springframework.annoconfig.model.Book;
import springframework.annoconfig.model.NameRequiredBeanFactoryPostProcessor;
import springframework.annoconfig.repository.BookRepository;

import java.util.List;

@Configuration
@ComponentScan("springframework.*")
// @ComponentScans({
//     @ComponentScan("springframework.annoconfig.repository"),
//     @ComponentScan("springframework.annoconfig.model"),
// })
public class AppConfig {
    private final BookRepository bookRepository;

    public AppConfig(BookRepository bookRepository) {
        System.out.printf("AppConfig 초기화시 주입되는 BookRepository: %s\n", bookRepository);
        // AppConfig 초기화시 주입되는 BookRepository: springframework.annoconfig.repository.BookRepositoryImpl@125290e5
        this.bookRepository = bookRepository;
    }

    /**
     * 스프링은 빈 객체를 실제로 생성하기 전에 설정 메타 정보를 변경하기 위한 용도로 {@link BeanFactoryPostProcessor} 인터페이스를 사용합니다.
     * {@link BeanFactoryPostProcessor}를 사용하여 빈 설정 추가 및 수정 등이 가능합니다.
     * <p>
     * {@link BeanFactoryPostProcessor}는 빈 객체들이 생성되기 전 단계에서 실행되어야 합니다.
     * 빈의 메타데이터({@link BeanDefinition})를 수정하는 역할을 하기 때문에, 일반적인 빈과 실행 타이밍이 다릅니다.
     * <p>
     * 이를 위해 다른 빈 인스턴스가 생성되기 전에 실행될 수 있도록 `static` 키워드를 사용하면,
     * 스프링 컨테이너가 `@Configuration` 클래스({@link AppConfig})의 인스턴스를 만들지 않아도 접근 가능해집니다.
     *
     * @return {@link springframework.annoconfig.model.NameRequired} 인터페이스를 구현한 객체에 이름이 없다면 기본 이름을 설정해주는 빈 팩토리 후처리기를 리턴합니다.
     * @see <a href="https://docs.spring.io/spring-framework/reference/core/beans/factory-nature.html">Customizing the Nature of a Bean</a>
     * @see <a href="https://docs.spring.io/spring-framework/reference/core/beans/factory-scopes.html">Bean Scopes</a>
     * @see <a href="https://medium.com/@TheTechDude/spring-bean-lifecycle-full-guide-f865966e89ce">Spring Bean Lifecycle Full Guide</a>
     */
    @Bean
    public static NameRequiredBeanFactoryPostProcessor nameRequiredBeanFactoryPostProcessor() {
        NameRequiredBeanFactoryPostProcessor processor = new NameRequiredBeanFactoryPostProcessor();
        processor.setDefaultName("Default");
        return processor;
    }

    @PostConstruct
    public void init() {
        bookRepository.save(new Book(
            "사장학개론",
            List.of("김승호"),
            "스노우폭스북스",
            "9791188331888",
            "1188331884"));
        bookRepository.save(new Book(
            "혼자 공부하는 C 언어",
            List.of("서현우"),
            "한빛미디어",
            "9791169210911",
            "1169210910"));
        bookRepository.save(new Book(
            "이것이 자바다",
            List.of("신용권", "임경균"),
            "한빛미디어",
            "9791169212298",
            "1169212298"));
    }
}
