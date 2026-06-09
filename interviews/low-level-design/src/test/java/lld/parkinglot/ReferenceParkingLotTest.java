package lld.parkinglot;

/** 정답지를 공용 계약 테스트에 연결한다. 항상 초록이어야 한다. */
class ReferenceParkingLotTest extends ParkingLotContractTest {

    @Override
    protected ParkingLot newLot(int small, int medium, int large) {
        return new ReferenceParkingLot(small, medium, large);
    }
}
