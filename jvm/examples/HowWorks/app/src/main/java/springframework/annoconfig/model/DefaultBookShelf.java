package springframework.annoconfig.model;

import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
import springframework.annoconfig.repository.BookRepository;

import java.util.Optional;

@Component
public class DefaultBookShelf implements BookShelf, NameRequired {
    private final BookRepository bookRepository;
    private String name;

    @Autowired
    public DefaultBookShelf(BookRepository bookRepository) {
        this.bookRepository = bookRepository;
    }

    @Override
    public void setName(String name) {
        this.name = name;
    }

    @Override
    public Optional<Book> getBookByIsbn13(String isbn) {
        return bookRepository.findByIsbn13(isbn);
    }

    @Override
    public void setBookByIsbn13(Book book) {
        bookRepository.save(book);
    }

    @Override
    public String toString() {
        return String.format("DefaultBookShelf[bookRepository=%s,name=%s]", bookRepository, name);
    }
}
