package lld.parkinglot;

import java.util.Optional;

/**
 * 직접 구현하는 공간. {@link ParkingLot} 헤더 카드와 테스트만 보고 모델을 채운다.
 * 막히면 그때만 {@link ReferenceParkingLot}을 연다.
 *
 * <p>핵심 질문: best-fit을 어떻게 표현하나? 새 차량/자리 타입 추가 시 어디만 바뀌어야 하나(OCP)?
 * 시작하려면 TODO를 채우고 {@code MyParkingLotTest}의 {@code @Disabled}를 지운다.
 */
public final class MyParkingLot implements ParkingLot {

    public MyParkingLot(int small, int medium, int large) {
        // TODO: 음수 검증 + 크기별 자리 생성, 빈 자리 목록 초기화
    }

    @Override
    public Optional<Ticket> park(Vehicle vehicle) {
        // TODO: best-fit 자리 찾기 -> 점유 표시 + 티켓 발급, 없으면 Optional.empty()
        throw new UnsupportedOperationException("아직 구현 전: MyParkingLot.park");
    }

    @Override
    public boolean leave(Ticket ticket) {
        // TODO: 티켓 검증(존재·점유·소유자 일치) -> 자리 비우기
        throw new UnsupportedOperationException("아직 구현 전: MyParkingLot.leave");
    }

    @Override
    public int availableSpots() {
        // TODO
        throw new UnsupportedOperationException("아직 구현 전: MyParkingLot.availableSpots");
    }
}
