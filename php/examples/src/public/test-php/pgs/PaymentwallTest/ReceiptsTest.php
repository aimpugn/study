<?php


require_once dirname(dirname(__DIR__)) . '/autoload.php';

$params = [
    'key' => '790d26a46214d76957edaf9aaa7bb6db',
    'ref' => 'b289413823',
    'type' => '0',
    'format' => 'html',
    'sign_version' => 3
];

function generateSignature($params = array(), $privateKey = '')
{
    ksort($params);
    $baseString = '';

    foreach ($params as $key => $value) {
        if (!isset($value)) {
            continue;
        }
        if (is_array($value)) {
            foreach ($value as $k => $v) {
                $baseString .= $key . '[' . $k . ']' . '=' . ($v === false ? '0' : $v);
            }
        } else {
            $baseString .= $key . '=' . ($value === false ? '0' : $value);
        }
    }


    return hash('sha256', $baseString . $privateKey);
}

$signature = generateSignature($params, '47f69d659ad6433296e006aa1334dac2');


# https://api.paymentwall.com/api/rest/receipt?key=790d26a46214d76957edaf9aaa7bb6db&ref=b289413823&type=0&format=html&sign_version=3&sign=abde9f334cd0748f0bc9cabeebf7355e7b0876cb48224f468a1091e3244257aa


function checkPaymentStatus($pgId, $ref, $pgSecret)
{
    $checkUrl = 'https://api.paymentwall.com/api/rest/payment';
    $params = [
        'key' => $pgId,
        'ref' => $ref,
        'sign_version' => 3,
    ];
    $sign = generateSignature($params, $pgSecret);
    $params['sign'] = $sign;

    $checkUrl .= '?' . http_build_query($params);

    try {
        $client = new \GuzzleHttp\Client();

        $response = $client->get($checkUrl);


        $body = $response->getBody()->getContents();

        return json_decode($body);
    } catch (Exception $e) {
        $this->log('PaymentwallCheckStatus(' . $ref . ')' . $e->getMessage(), LOG_ERR);
        throw new Exception('결제 검증에 실패하였습니다.');
    }
}

print_r(json_encode(checkPaymentStatus('790d26a46214d76957edaf9aaa7bb6db', 'b289413823', '47f69d659ad6433296e006aa1334dac2'), JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_PRETTY_PRINT));