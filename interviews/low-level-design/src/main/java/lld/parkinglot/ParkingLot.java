package lld.parkinglot;

import java.util.Optional;

/**
 * 주차장 - 면접 암기 카드. 모델링(객체지향 설계) 대표 문제. 정답지는 {@link ReferenceParkingLot}.
 *
 * <p>직접 연습은 {@link MyParkingLot}. 둘 다 {@code ParkingLotContractTest}로 검증된다.
 * 알고리즘이 아니라 '확장 가능한 객체 모델'을 보는 문제다.
 *
 * <pre>
 * 한 줄    : 차량을 빈 자리에 배정/회수. 핵심은 자료구조가 아니라 책임이 분리된 클래스 모델.
 * 모델     : Vehicle(크기) · Spot(크기) · Ticket · ParkingLot. 적합 규칙: spot.size &gt;= vehicle.size(큰 자리는 작은 차 수용).
 * 배정 전략 : best-fit(맞는 가장 작은 자리)으로 큰 자리를 보존. 전략을 인터페이스로 빼면 first-fit 등으로 교체(전략 패턴).
 * OCP      : 새 차량/자리 타입은 enum과 적합 규칙만 늘리고 ParkingLot 본체는 그대로. if(type==..) 분기 나열을 피한다.
 * 합성     : 실제로는 Lot -&gt; Floor -&gt; Spot 계층. 여기서는 Lot이 Spot을 직접 보유하도록 단순화.
 * 동시성   : 같은 자리 이중 배정 방지 - 자리 점유는 원자적이어야 한다(락/CAS). 단골 꼬리질문.
 * 확장     : 요금 계산(시간×요율), 자리 검색을 크기별 가용 목록으로 O(1), 다층/구역, 전기차 전용 자리.
 * 꼬리질문 : 자리 검색을 O(1)로? 요금 정책 추가는? 동시 입차 경쟁은? 새 차량 타입 추가 비용은?
 * </pre>
 */
public interface ParkingLot {

    /** 작을수록 작은 차/자리. 큰 자리는 작은 차를 수용한다(LARGE 자리에 SMALL 차 OK). */
    enum VehicleSize { SMALL, MEDIUM, LARGE }

    record Vehicle(String licensePlate, VehicleSize size) { }

    /** 입차 증서. 회수할 때 어떤 자리의 어떤 차였는지 확인한다. */
    record Ticket(String id, String licensePlate, int spotId) { }

    /** 맞는 자리가 있으면 배정하고 티켓 발급, 없으면 빈 값. */
    Optional<Ticket> park(Vehicle vehicle);

    /** 티켓의 자리를 비운다. 티켓이 유효하고 점유 중이었으면 {@code true}. */
    boolean leave(Ticket ticket);

    /** 현재 비어 있는 자리 수. */
    int availableSpots();
}
