<?php
/*
 * ```sql
 * CREATE TABLE payments (
 *     id SERIAL PRIMARY KEY,
 *     uuid VARCHAR(36) NOT NULL,
 *     order_id VARCHAR(50) NOT NULL,
 *     payment_status VARCHAR(20) NOT NULL,
 *     idempotency_key VARCHAR(36) NOT NULL,
 *     created_at TIMESTAMP DEFAULT NOW(),
 *     updated_at TIMESTAMP DEFAULT NOW(),
 *     UNIQUE (uuid),
 *     UNIQUE (idempotency_key)
 * );
 * ```
 */

/**
 * 각 결제 요청마다 고유한 Idempotency Key를 생성하여 중복된 요청이 발생할 경우, 처음 처리된 요청만을 유효하게 처리하고 나머지는 무시합니다.
 * - 장점: 중복 요청을 쉽게 처리할 수 있으며, API 설계에서 많이 사용됨.
 * - 단점: 모든 요청에 대해 고유 키를 관리해야 하며, 키의 중복 관리를 신경 써야 함.
 */
function processPayment($orderId, $uuid, $paymentData, $idempotencyKey)
{
    $conn = new PDO('mysql:host=localhost;dbname=testdb', 'dbuser', 'dbpass');
    $conn->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    try {
        $conn->beginTransaction();

        // Idempotency Key를 사용하여 중복 요청을 방지합니다.
        $stmt = $conn->prepare('SELECT * FROM payments WHERE idempotency_key = :idempotency_key');
        $stmt->execute(['idempotency_key' => $idempotencyKey]);
        $existingPayment = $stmt->fetch(PDO::FETCH_ASSOC);

        if ($existingPayment) {
            // 이미 처리된 요청이면 기존 결과를 반환합니다.
            $conn->rollBack();
            return ['status' => $existingPayment['payment_status']];
        }

        // 결제 승인 요청 로직
        $paymentResult = requestPaymentApproval($paymentData);

        if ($paymentResult['status'] === 'success') {
            // 결제 승인 성공 시
            $stmt = $conn->prepare('INSERT INTO payments (uuid, order_id, payment_status, idempotency_key) VALUES (:uuid, :order_id, :status, :idempotency_key)');
            $stmt->execute(['uuid' => $uuid, 'order_id' => $orderId, 'status' => 'SUCCESS', 'idempotency_key' => $idempotencyKey]);
        } else {
            // 결제 승인 실패 시
            $stmt = $conn->prepare('INSERT INTO payments (uuid, order_id, payment_status, idempotency_key) VALUES (:uuid, :order_id, :status, :idempotency_key)');
            $stmt->execute(['uuid' => $uuid, 'order_id' => $orderId, 'status' => 'FAILURE', 'idempotency_key' => $idempotencyKey]);
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
