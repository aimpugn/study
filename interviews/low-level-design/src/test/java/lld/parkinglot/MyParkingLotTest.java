package lld.parkinglot;

import org.junit.jupiter.api.Disabled;

/** 내 구현을 정답지와 같은 계약 테스트에 연결한다. 구현을 시작하면 아래 한 줄을 지운다. */
@Disabled("내 구현(MyParkingLot)을 시작하면 이 줄을 지우세요.")
class MyParkingLotTest extends ParkingLotContractTest {

    @Override
    protected ParkingLot newLot(int small, int medium, int large) {
        return new MyParkingLot(small, medium, large);
    }
}
