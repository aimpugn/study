<?php

require_once dirname(__DIR__) . '/autoload.php';


function initialPrivateKey($merchantKey)
{

    $needKey = substr($merchantKey, 0, 32);
    $hexBytes = str_split($needKey);

    $length = sizeof($hexBytes) / 2;

    $rawBytes = array();

    for($i = 0 ; $i < $length ;$i++) {
        $high = base_convert((string)$hexBytes[$i * 2], 16, 10);
        $low = base_convert((string)$hexBytes[$i * 2 + 1], 16, 10);

        if($high == 0) {
            $high = ord($hexBytes[$i * 2]) % 16;
        }

        if($low == 0) {
            $low = ord($hexBytes[$i * 2 + 1]) % 16;
        }

        $value = (($high << 4) | $low);

        if($value > 127) {
            $value -= 256;
        }
        $rawBytes[$i] = chr($value);

    } // for end
    return implode($rawBytes);
}

# 결제정보 암호화 데이터
# 암호화 알고리즘 : AES/ECB/PKCS5padding
# 암호결과 인코딩 : Hex Encoding
# 암호 key : MID에 부여된 MerchantKey 앞 16자리
# 결제정보 암호화 생성 규칙 Hex(AES(CardNo=value&ExpYear=YY&ExpMonth=MM&IDNo=value&CardPw=value))
/*
 카드번호 : 1234567890123456 / 유효기간(년) : 25년 / 유효기간(월) : 12월 / 생년월일(6) 또는 사업자번호(10) : 800101 / 비밀번호(앞2자리) : 12
- 상점키 : b+zhZ4yOZ7FsH8pm10safdaspidfWERsad10fijw86yLc6BJeFVrZFXhAoJ3gEWgr+wN123MV0W4hvDdbe4Sjw==
- 평문 : CardNo=1234567890123456&ExpYear=25&ExpMonth=12&IDNo=800101&CardPw=12
- 암호 Key : b+zhZ4yOZ7FsH8pm(상점키 앞16자리)
- 암호화결과 : 7b23e8b9e9e144228d4c288fbedb570ec6e6466c9b59a0e1670204550cc1954a9e638e3b4daa5cef1f4f238539e28181782196f43b3b72b9dad 3956bbd7117b41204cd479bcd2afc55e790b5bc121855
 */
$mid = 'someMID05m';
$ediDate = date('YmdHis');
$moid = "moid_$ediDate";
$merchantKey = 'bJIP7QTPWgXVfZUJnrQr1cMGDTVAiCxAjSdvLZsukPeZwWOSx55hQoNugA9kQfrLx0zfqFoNcxY1QMVb/OjJKA==';
$cardNo = '5361487600310859';
$expYear = '23';
$expMonth = '10';
$idNo = '861202';
$cardPw = '10';
$passwordKey = substr($merchantKey, 0, 16);
$plainText = "CardNo=$cardNo&ExpYear=$expYear&ExpMonth=$expMonth&IDNo=$idNo&CardPw=$cardPw";

$signData = bin2hex(hash('sha256', $mid . $ediDate . $moid . $merchantKey, true));
$aes = new Crypt_AES(CRYPT_AES_MODE_ECB);
$aes->setKey($passwordKey);
$encData = bin2hex($aes->encrypt($plainText));
// CardNo=1234567890123456&ExpYear=25&ExpMonth=12&IDNo=800101&CardPw=12
print <<<EOF
====== EncData
mid: $mid
editDate: $ediDate
moid: $moid
passwordKey: $passwordKey
plainText: $plainText
encData: $encData

====== SignData
signData: $signData

EOF;
// SignData: 위변조 검증 Data, Hex(SHA256(MID + EdiDate + Moid + 상점키))
/*
# Body
====== EncData
mid: someMID05m
editDate: 20221128190306
moid: moid_20221128190306
passwordKey: bJIP7QTPWgXVfZUJ
plainText: CardNo=5361487600310859&ExpYear=23&ExpMonth=10&IDNo=861202&CardPw=10
encData: d019f08088e327a8b1aab8be36fc5c5f1e9cb81bb01e7c5a8f675af8e6a0cde668127af14d25b253533a377af0545164a97fffb0d8851c19d039efaf43d40e86ea5fa94ed9bfeba55bbffdd654353483

====== SignData
signData: daefc45ffcefc526944b45715601bdb76ab671e0f107d1a8960a52eaef4958cf

# Response
{
  "ResultCode": "F100",
  "ResultMsg": "빌키가 정상적으로 생성되었습니다.",
  "BID": "BIKYsomeMID05m2211281903264568",
  "AuthDate": "20221128",
  "CardCode": "04",
  "CardName": "[삼성]",
  "TID": "someMID05m01162211281903267425",
  "CardCl": "0",
  "AcquCardCode": "04",
  "AcquCardName": "[삼성]"
}
 */
