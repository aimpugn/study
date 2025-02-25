package springframework.annoconfig.model;

import java.util.List;
import java.util.Objects;

public record Book(
    String title,
    List<String> authors,
    String publisher,
    String isbn13,
    String isbn10
) {
    public Book {
        Objects.requireNonNull(title, "title은 null이 될 수 없습니다.");
        Objects.requireNonNull(authors, "authors은 null이 될 수 없습니다.");
        Objects.requireNonNull(publisher, "publisher은 null이 될 수 없습니다.");
        Objects.requireNonNull(isbn13, "ISBN-13은 null이 될 수 없습니다.");
        Objects.requireNonNull(isbn10, "ISBN-10은 null이 될 수 없습니다.");

        if (title.isBlank()) {
            throw new IllegalArgumentException("책 타이틀이 비었습니다.");
        }
        if (authors.isEmpty()) {
            throw new IllegalArgumentException("저자 목록이 비었습니다.");
        }
        if (publisher.isBlank()) {
            throw new IllegalArgumentException("출판사가 비었습니다.");
        }
        if (isbn13.isBlank()) {
            throw new IllegalArgumentException("ISBN-13이 비었습니다.");
        }
        if (isbn13.length() != 13) {
            throw new IllegalArgumentException("ISBN-13 길이에 맞지 않습니다.");
        }
        if (isbn10.isBlank()) {
            throw new IllegalArgumentException("ISBN-10이 비었습니다.");
        }
        if (isbn10.length() != 10) {
            throw new IllegalArgumentException("ISBN-10 길이에 맞지 않습니다.");
        }
    }
}
