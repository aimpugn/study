package springframework.annoconfig;

import org.springframework.context.annotation.AnnotationConfigApplicationContext;
import springframework.annoconfig.config.AppConfig;
import springframework.annoconfig.model.Book;
import springframework.annoconfig.model.BookShelf;
import springframework.annoconfig.repository.BookRepository;

import java.util.List;
import java.util.Optional;
import java.util.function.BiConsumer;

public class SpringframeworkMain {

    private static void testAnnotationConfigApplicationContext() {
        BiConsumer<Optional<Book>, String> optionalBookConsumer = (Optional<Book> book, String isbn13) -> book.ifPresentOrElse(
            System.out::println,
            // 람다 표현식에서 사용하는 지역 변수는 `final`이거나 "effectively final"(사실상 final)이어야 합니다.
            () -> System.out.printf("ISBN-13(%s)으로 등록된 책을 찾을 수 없습니다.\n", isbn13)
        );

        try (AnnotationConfigApplicationContext ctx = new AnnotationConfigApplicationContext(AppConfig.class)) {
            BookRepository bookRepository = ctx.getBean(BookRepository.class);
            String firstIsbn13 = "9791188331888";
            Optional<Book> firstBook = bookRepository.findByIsbn13(firstIsbn13);
            optionalBookConsumer.accept(firstBook, firstIsbn13);

            BookShelf bookShelf = ctx.getBean(BookShelf.class);
            System.out.printf("BookShelf: %s\n", bookShelf);
            String secondIsbn13 = "9788966260959";
            Optional<Book> secondBook = bookShelf.getBookByIsbn13(secondIsbn13);
            optionalBookConsumer.accept(secondBook, secondIsbn13);

            bookShelf.setBookByIsbn13(new Book(
                "Clean Code(클린 코드)",
                List.of("로버트 C. 마틴"),
                "인사이트",
                "9788966260959",
                "8966260950"
            ));

            secondBook = bookShelf.getBookByIsbn13(secondIsbn13);
            optionalBookConsumer.accept(secondBook, secondIsbn13);
        }
    }

    public static void main(String[] args) {
        testAnnotationConfigApplicationContext();
    }
}
