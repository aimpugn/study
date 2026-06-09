package lld.parkinglot;

import java.util.ArrayDeque;
import java.util.Deque;
import java.util.EnumMap;
import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 주차장 정답지.
 *
 * <p>배정은 best-fit: 필요한 크기 이상 자리 중 '가장 작은' 자리를 준다. 그래서 작은 차가 큰 자리를
 * 낭비하지 않는다. 크기별 '빈 자리' 목록({@code freeBySize})을 두어 검색을 O(자리 종류 수)로 끝낸다.
 *
 * <p>적합 규칙은 {@link VehicleSize}의 순서(ordinal)로 표현한다. 새 크기를 추가해도 분기 나열이 늘지 않는다(OCP).
 */
public final class ReferenceParkingLot implements ParkingLot {

    private static final class Spot {
        final int id;
        final VehicleSize size;
        boolean occupied;
        String occupantPlate;

        Spot(int id, VehicleSize size) {
            this.id = id;
            this.size = size;
        }
    }

    private final Map<Integer, Spot> spotsById = new HashMap<>();
    private final Map<VehicleSize, Deque<Spot>> freeBySize = new EnumMap<>(VehicleSize.class);
    private final AtomicLong ticketSeq = new AtomicLong();
    private int available;

    public ReferenceParkingLot(int small, int medium, int large) {
        if (small < 0 || medium < 0 || large < 0) {
            throw new IllegalArgumentException("spot counts must be >= 0");
        }
        for (VehicleSize size : VehicleSize.values()) {
            freeBySize.put(size, new ArrayDeque<>());
        }
        int nextId = 0;
        nextId = createSpots(nextId, VehicleSize.SMALL, small);
        nextId = createSpots(nextId, VehicleSize.MEDIUM, medium);
        createSpots(nextId, VehicleSize.LARGE, large);
        this.available = small + medium + large;
    }

    private int createSpots(int startId, VehicleSize size, int count) {
        int id = startId;
        for (int i = 0; i < count; i++) {
            Spot spot = new Spot(id, size);
            spotsById.put(id, spot);
            freeBySize.get(size).push(spot);
            id++;
        }
        return id;
    }

    @Override
    public synchronized Optional<Ticket> park(Vehicle vehicle) {
        Spot spot = takeBestFitSpot(vehicle.size());
        if (spot == null) {
            return Optional.empty();
        }
        spot.occupied = true;
        spot.occupantPlate = vehicle.licensePlate();
        available--;
        Ticket ticket = new Ticket("T" + ticketSeq.incrementAndGet(), vehicle.licensePlate(), spot.id);
        return Optional.of(ticket);
    }

    private Spot takeBestFitSpot(VehicleSize needed) {
        // needed 이상 크기 중 가장 작은 자리부터 본다: SMALL 차는 SMALL -> MEDIUM -> LARGE.
        for (VehicleSize size : VehicleSize.values()) {
            if (size.ordinal() < needed.ordinal()) {
                continue; // 차보다 작은 자리는 못 쓴다.
            }
            Deque<Spot> free = freeBySize.get(size);
            if (!free.isEmpty()) {
                return free.pop();
            }
        }
        return null;
    }

    @Override
    public synchronized boolean leave(Ticket ticket) {
        Spot spot = spotsById.get(ticket.spotId());
        if (spot == null || !spot.occupied || !spot.occupantPlate.equals(ticket.licensePlate())) {
            return false; // 없는 자리거나, 이미 비었거나, 다른 차의 티켓이면 거부.
        }
        spot.occupied = false;
        spot.occupantPlate = null;
        freeBySize.get(spot.size).push(spot);
        available++;
        return true;
    }

    @Override
    public synchronized int availableSpots() {
        return available;
    }
}
