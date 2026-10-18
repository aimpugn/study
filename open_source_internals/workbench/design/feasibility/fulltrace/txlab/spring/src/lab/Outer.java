package lab;

import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.transaction.annotation.Transactional;

/**
 * 바깥 트랜잭션입니다. 장면마다 메서드 하나이고, 모두 기본 전파(REQUIRED)로 트랜잭션을 새로 엽니다.
 *
 * <p>{@link #once()}를 뺀 모든 장면이 같은 꼴입니다. 1번 행을 고치고 멈춘 뒤 안쪽 빈({@link Inner})을 부르고, 돌아오면 다시
 * 멈춥니다. 안쪽은 다른 빈이어야 합니다. 같은 객체 안에서 부르면 프록시를 거치지 않아 {@code @Transactional}이 아무 일도
 * 하지 않습니다.
 */
public class Outer {

    private final JdbcTemplate jdbc;
    private final Inner inner;

    public Outer(JdbcTemplate jdbc, Inner inner) {
        this.jdbc = jdbc;
        this.inner = inner;
    }

    /**
     * 안쪽 REQUIRED가 실패하고 바깥이 그 예외를 삼키는 장면입니다. 바깥은 정상으로 끝나지만 커밋 대신 롤백되고
     * {@code UnexpectedRollbackException}을 받는다고 예상합니다(WORK 실험 5 사전 등록).
     */
    @Transactional
    public void required() {
        jdbc.update("UPDATE acct SET balance = balance - 10 WHERE id = 1");
        Lab.at("outer-after-update");
        try {
            inner.requiredFails();
        } catch (RuntimeException e) {
            Lab.note("outer-caught", e);
        }
        Lab.at("outer-after-inner");
    }

    /** 안쪽이 REQUIRES_NEW로 다른 행을 고치고 먼저 커밋하는 장면입니다. 풀 크기 1의 장면도 이 메서드를 씁니다. */
    @Transactional
    public void requiresNew() {
        jdbc.update("UPDATE acct SET balance = balance - 10 WHERE id = 1");
        Lab.at("outer-after-update");
        inner.requiresNew();
        Lab.at("outer-after-inner");
    }

    /** 안쪽이 REQUIRES_NEW로 바깥이 잠근 같은 행을 고치려는 장면입니다. 안쪽은 바깥의 커밋을, 바깥은 안쪽의 반환을 기다립니다. */
    @Transactional
    public void requiresNewSameRow() {
        jdbc.update("UPDATE acct SET balance = balance - 10 WHERE id = 1");
        Lab.at("outer-after-update");
        inner.requiresNewSameRow();
        Lab.at("outer-after-inner");
    }

    /** 안쪽 NESTED가 실패하고 바깥이 그 예외를 삼키는 장면입니다. 안쪽만 저장점까지 되돌아가고 바깥은 커밋된다고 예상합니다. */
    @Transactional
    public void nested() {
        jdbc.update("UPDATE acct SET balance = balance - 10 WHERE id = 1");
        Lab.at("outer-after-update");
        try {
            inner.nestedFails();
        } catch (RuntimeException e) {
            Lab.note("outer-caught", e);
        }
        Lab.at("outer-after-inner");
    }

    /**
     * 1번 행에서 10을 빼는 트랜잭션 하나입니다. 안에 멈춤 자리가 없습니다. 5-reborrow 장면은 이것을 멈춤 없이 두 번 잇달아 불러,
     * 둘째 빌리기가 첫째가 연결을 돌려준 뒤 0.5초 안에 오게 합니다(HikariCP는 그 안이면 연결이 살아 있는지 검사하지 않음).
     */
    @Transactional
    public void once() {
        jdbc.update("UPDATE acct SET balance = balance - 10 WHERE id = 1");
    }
}
