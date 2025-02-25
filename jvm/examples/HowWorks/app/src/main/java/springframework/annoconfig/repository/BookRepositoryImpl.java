package springframework.annoconfig.repository;

import org.springframework.stereotype.Repository;
import springframework.annoconfig.config.AppConfig;
import springframework.annoconfig.model.Book;

import java.util.Map;
import java.util.Optional;

/**
 * {@link Repository}는 {@link org.springframework.stereotype.Component}의 확장 개념입니다.
 * {@link Repository}는 Spring이 자동으로 감지하고 빈으로 관리할 리포지토리 클래스에 붙입니다.
 * 이 클래스가 스캔되고, {@link AppConfig#bookRepository} 속성에 주입됩니다.
 */
@Repository //
public class BookRepositoryImpl implements BookRepository {
    private final Map<String, Book> bookMap;

    public BookRepositoryImpl(Map<String, Book> bookMap) {
        this.bookMap = bookMap;
    }

    @Override
    public Optional<Book> findByIsbn13(String isbn13) {
        return Optional.ofNullable(bookMap.get(isbn13));
    }

    @Override
    public void save(Book book) {
        bookMap.put(book.isbn13(), book);
    }
}
