package springframework.annoconfig.model;

import org.springframework.beans.BeansException;
import org.springframework.beans.MutablePropertyValues;
import org.springframework.beans.factory.config.BeanDefinition;
import org.springframework.beans.factory.config.BeanFactoryPostProcessor;
import org.springframework.beans.factory.config.ConfigurableListableBeanFactory;

public class NameRequiredBeanFactoryPostProcessor implements BeanFactoryPostProcessor {
    private String defaultName;

    public void setDefaultName(String defaultName) {
        this.defaultName = defaultName;
    }

    @Override
    public void postProcessBeanFactory(ConfigurableListableBeanFactory beanFactory) throws BeansException {
        for (String name : beanFactory.getBeanDefinitionNames()) {
            System.out.printf("Bean definition name: %s\n", name);
            // `BeanDefinition`에는 해당 빈이 어떤 인터페이스를 구현하는지에 대한 정보가 없습니다.
            Class<?> beanClass = beanFactory.getType(name);

            // `instanceof`로 체크하려면 인스턴스가 필요합니다.
            // 하지만 빈이 생성되기 전이므로 `isAssignableFrom`을 사용해야 합니다.
            if (
                beanClass != null &&
                    NameRequired.class.isAssignableFrom(beanClass)
            ) {
                System.out.printf("%s bean is assignable to NameRequired\n", name);
                BeanDefinition bd = beanFactory.getBeanDefinition(name);
                MutablePropertyValues mpv = bd.getPropertyValues();
                System.out.printf("\tMutablePropertyValues: %s\n", mpv);
                if (!mpv.contains("name")) {
                    mpv.add("name", defaultName);
                    // 만약 `name` 프로퍼티를 주입하지 않으면 다음과 같이 출력됩니다.
                    // BookShelf: DefaultBookShelf[bookRepository=springframework.annoconfig.repository.BookRepositoryImpl@3406472c,name=null]
                    //
                    // 하지만 `name` 프로퍼티를 주입하면, `"Default"`라는 이름이 주입되고, 다음과 같이 출력됩니다.
                    // BookShelf: DefaultBookShelf[bookRepository=springframework.annoconfig.repository.BookRepositoryImpl@5717c37,name=Default]

                    // mpv.add("unknown", defaultName);
                    // 만약 위와 같이 존재하지 않는 프로퍼티를 설정하려고 하면, 다음과 같은 에러가 발생합니다.
                    // ```
                    // Exception in thread "main" org.springframework.beans.factory.BeanCreationException: Error creating bean with name 'defaultBookShelf' defined in file [/Users/rody/VscodeProjects/study/jvm/examples/HowWorks/app/build/classes/java/main/springframework/model/DefaultBookShelf.class]: Invalid property 'unknown' of bean class [springframework.annoconfig.model.DefaultBookShelf]: Bean property 'unknown' is not writable or has an invalid setter method. Does the parameter type of the setter match the return type of the getter?
                    // ... 생략 ...
                    // Caused by: org.springframework.beans.NotWritablePropertyException: Invalid property 'unknown' of bean class [springframework.annoconfig.model.DefaultBookShelf]: Bean property 'unknown' is not writable or has an invalid setter method. Does the parameter type of the setter match the return type of the getter?
                    // ... 생략 ...
                    // ```
                }
            }
        }
    }
}
