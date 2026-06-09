package lld.parkinglot;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.Optional;
import lld.parkinglot.ParkingLot.Ticket;
import lld.parkinglot.ParkingLot.Vehicle;
import lld.parkinglot.ParkingLot.VehicleSize;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

/** 주차장의 공용 계약 테스트. 정답지와 내 구현이 같은 동작 계약을 통과해야 한다. */
abstract class ParkingLotContractTest {

    protected abstract ParkingLot newLot(int small, int medium, int large);

    @Test
    @DisplayName("기본: 빈 자리에 입차하면 티켓 발급")
    void park_returns_ticket() {
        ParkingLot lot = newLot(1, 0, 0);
        Optional<Ticket> ticket = lot.park(new Vehicle("11가1111", VehicleSize.SMALL));
        assertTrue(ticket.isPresent());
        assertEquals(0, lot.availableSpots());
    }

    @Test
    @DisplayName("만차면 빈 값을 돌려준다")
    void full_lot_returns_empty() {
        ParkingLot lot = newLot(1, 0, 0);
        lot.park(new Vehicle("a", VehicleSize.SMALL));
        assertTrue(lot.park(new Vehicle("b", VehicleSize.SMALL)).isEmpty());
    }

    @Test
    @DisplayName("best-fit: 작은 차가 큰 자리를 낭비하지 않는다")
    void best_fit_preserves_larger_spots() {
        ParkingLot lot = newLot(1, 0, 1); // 작은 자리 1 + 큰 자리 1
        // 작은 차가 큰 자리를 차지하면 아래 큰 차가 못 들어온다. best-fit이면 둘 다 성공.
        assertTrue(lot.park(new Vehicle("small", VehicleSize.SMALL)).isPresent());
        assertTrue(lot.park(new Vehicle("large", VehicleSize.LARGE)).isPresent());
    }

    @Test
    @DisplayName("맞는 자리가 없으면(차가 너무 큼) 빈 값")
    void vehicle_too_large_returns_empty() {
        ParkingLot lot = newLot(1, 0, 0); // 작은 자리뿐
        assertTrue(lot.park(new Vehicle("bus", VehicleSize.LARGE)).isEmpty());
    }

    @Test
    @DisplayName("출차하면 자리가 다시 빈다")
    void leave_frees_spot() {
        ParkingLot lot = newLot(1, 0, 0);
        Ticket ticket = lot.park(new Vehicle("a", VehicleSize.SMALL)).orElseThrow();
        assertTrue(lot.leave(ticket));
        assertEquals(1, lot.availableSpots());
        assertTrue(lot.park(new Vehicle("b", VehicleSize.SMALL)).isPresent());
    }

    @Test
    @DisplayName("잘못된 티켓 출차는 거부(이중 출차·위조 포함)")
    void leave_invalid_ticket_returns_false() {
        ParkingLot lot = newLot(1, 0, 0);
        Ticket ticket = lot.park(new Vehicle("a", VehicleSize.SMALL)).orElseThrow();
        assertTrue(lot.leave(ticket));
        assertFalse(lot.leave(ticket), "이미 나간 차는 다시 못 나간다");
        assertFalse(lot.leave(new Ticket("T999", "x", 0)), "위조 티켓 거부");
    }

    @Test
    @DisplayName("실패: 자리 수가 음수면 예외")
    void invalid_counts_rejected() {
        assertThrows(IllegalArgumentException.class, () -> newLot(-1, 0, 0));
    }
}
