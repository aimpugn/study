package springframework.annoconfig.model;

import java.util.Optional;

public interface BookShelf {
    Optional<Book> getBookByIsbn13(String isbn);

    void setBookByIsbn13(Book book);
}
