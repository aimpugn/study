package springframework.annoconfig.repository;

import springframework.annoconfig.model.Book;

import java.util.Optional;

public interface BookRepository {
    /**
     * {@link Optional} 타입을 사용하여 `null` 반환을 방지합니다.
     *
     * @param isbn13 ISBN-13
     * @return ISBN-13으로 조회된 책을 반환합니다.
     */
    Optional<Book> findByIsbn13(String isbn13);

    void save(Book book);
}
