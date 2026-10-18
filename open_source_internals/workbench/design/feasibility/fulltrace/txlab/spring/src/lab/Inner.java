package lab;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.transaction.annotation.Propagation;
import org.springframework.transaction.annotation.Transactional;

/** 안쪽 트랜잭션입니다. 메서드마다 전파 방식이 다르고, 바깥({@link Outer})의 같은 이름 장면이 부릅니다. */
public class Inner {

    private final JdbcTemplate jdbc;

    public Inner(JdbcTemplate jdbc) {
        this.jdbc = jdbc;
    }

    /** 바깥 트랜잭션에 참여해(REQUIRED) 2번 행을 고친 뒤 실패합니다. */
    @Transactional
    public void requiredFails() {
        jdbc.update("UPDATE acct SET balance = balance + 10 WHERE id = 2");
        Lab.at("inner-after-update");
        throw new IllegalStateException("inner fails");
    }

    /** 바깥을 잠시 떼어 두고(REQUIRES_NEW) 새 트랜잭션으로 2번 행을 고친 뒤 커밋합니다. */
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void requiresNew() {
        jdbc.update("UPDATE acct SET balance = balance + 10 WHERE id = 2");
        Lab.at("inner-after-update");
    }

    /**
     * 새 트랜잭션으로 바깥이 이미 잠근 1번 행을 고치려 합니다. 락 기다림을 3초로 묶어 둡니다({@code SET LOCAL}은 이 트랜잭션에만
     * 적용). 묶지 않으면 바깥은 안쪽이 돌아오기를, 안쪽은 바깥의 커밋을 끝없이 기다립니다.
     */
    @Transactional(propagation = Propagation.REQUIRES_NEW)
    public void requiresNewSameRow() {
        jdbc.execute("SET LOCAL lock_timeout = '3s'");
        Lab.at("inner-before-update");
        jdbc.update("UPDATE acct SET balance = balance + 10 WHERE id = 1");
        Lab.at("inner-after-update");
    }

    /** 바깥 트랜잭션 안에 저장점을 만들고(NESTED) 2번 행을 고친 뒤 실패합니다. */
    @Transactional(propagation = Propagation.NESTED)
    public void nestedFails() {
        jdbc.update("UPDATE acct SET balance = balance + 10 WHERE id = 2");
        Lab.at("inner-after-update");
        throw new IllegalStateException("inner fails");
    }
}
