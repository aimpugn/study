<?php

function processPayment($orderId, $uuid, $paymentData)
{
    $conn = new PDO('mysql:host=localhost;dbname=testdb', 'dbuser', 'dbpass');
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    try {
        $conn->beginTransaction();

        // 결제 시도 기록 삽입
        $stmt = $conn->prepare(
            'INSERT INTO payments (uuid, order_id, payment_status) VALUES (:uuid, :order_id, :status)'
        );
        $stmt->execute(['uuid' => $uuid, 'order_id' => $orderId, 'status' => 'PENDING']);

        // 현재 order_id의 최종 상태를 가져옵니다.
        $stmt = $conn->prepare(
            'SELECT * 
            FROM payments 
            WHERE order_id = :order_id AND final_status = TRUE 
            FOR UPDATE'
        );
        $stmt->execute(['order_id' => $orderId]);
        $finalOrder = $stmt->fetch(PDO::FETCH_ASSOC);

        // 결제 승인 요청 로직
        $paymentResult = requestPaymentApproval($paymentData);

        if ($paymentResult['status'] === 'success') {
            // 결제 승인 성공 시
            $stmt = $conn->prepare(
                'UPDATE payments 
                SET 
                    payment_status = :status, 
                    version = version + 1, 
                    updated_at = NOW() 
                WHERE uuid = :uuid'
            );
            $stmt->execute(['status' => 'SUCCESS', 'uuid' => $uuid]);

            // 이미 최종 상태가 설정되지 않았다면 업데이트
            if (!$finalOrder) {
                $stmt = $conn->prepare(
                    'UPDATE payments 
                    SET final_status = TRUE, 
                    updated_at = NOW() 
                    WHERE uuid = :uuid'
                );
                $stmt->execute(['uuid' => $uuid]);
            }
        } else {
            // 결제 승인 실패 시
            $stmt = $conn->prepare(
                'UPDATE payments 
                SET 
                    payment_status = :status, 
                    version = version + 1, 
                    updated_at = NOW() 
                WHERE uuid = :uuid'
            );
            $stmt->execute(['status' => 'FAILURE', 'uuid' => $uuid]);

            // 만약 이미 성공된 결제가 없다면, 최종 상태 업데이트
            if (!$finalOrder) {
                $stmt = $conn->prepare(
                    'UPDATE payments 
                    SET 
                        final_status = TRUE, 
                        updated_at = NOW() 
                    WHERE uuid = :uuid'
                );
                $stmt->execute(['uuid' => $uuid]);
            }
        }

        // 트랜잭션 커밋
        $conn->commit();

        return $paymentResult;

    } catch (Exception $e) {
        // 예외 발생 시 트랜잭션 롤백
        $conn->rollBack();
        throw $e;
    }
}

function requestPaymentApproval($paymentData)
{
    // PG사와 통신하여 결제 승인 요청을 처리하는 로직 구현
    // 예를 들어, HTTP API 호출 등을 통해 처리
    // 이 예제에서는 간단히 성공 응답을 반환
    return ['status' => 'success'];
}
