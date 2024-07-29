<?php
require_once dirname(__DIR__) . '/autoload.php';

function to_utf8( $string ) {
// From http://w3.org/International/questions/qa-forms-utf-8.html
	if ( preg_match('%^(?:
      [\x09\x0A\x0D\x20-\x7E]            # ASCII
    | [\xC2-\xDF][\x80-\xBF]             # non-overlong 2-byte
    | \xE0[\xA0-\xBF][\x80-\xBF]         # excluding overlongs
    | [\xE1-\xEC\xEE\xEF][\x80-\xBF]{2}  # straight 3-byte
    | \xED[\x80-\x9F][\x80-\xBF]         # excluding surrogates
    | \xF0[\x90-\xBF][\x80-\xBF]{2}      # planes 1-3
    | [\xF1-\xF3][\x80-\xBF]{3}          # planes 4-15
    | \xF4[\x80-\x8F][\x80-\xBF]{2}      # plane 16
)*$%xs', $string) ) {
		return $string;
	} else {
		return iconv( 'CP1252', 'UTF-8', $string);
	}
}

function unicode_urldecode($url)
{
	preg_match_all('/%u([[:alnum:]]{4})/', $url, $a);

	foreach ($a[1] as $uniord)
	{
		$dec = hexdec($uniord);
		$utf = '';

		if ($dec < 128)
		{
			$utf = chr($dec);
		}
		else if ($dec < 2048)
		{
			$utf = chr(192 + (($dec - ($dec % 64)) / 64));
			$utf .= chr(128 + ($dec % 64));
		}
		else
		{
			$utf = chr(224 + (($dec - ($dec % 4096)) / 4096));
			$utf .= chr(128 + ((($dec % 4096) - ($dec % 64)) / 64));
			$utf .= chr(128 + ($dec % 64));
		}

		$url = str_replace('%u'.$uniord, $utf, $url);
	}

	return urldecode($url);
}

$query = "my=apples&are=green+and+red";

foreach(explode('&', $query) as $chunk) {
	$param = explode("=", $chunk);

	if($param) {
		printf("Value for parameter \"%s\" is \"%s\"<br/>\n", urldecode($param[0]), urldecode($param[1]));
	}
}

$data1 = [
	'sellerKey' => 'IM_YQFEGJ',
	'orderNo' => '202206072623965063',
	'requestMemo' => "[테스트] 테스트"
];
$decodedData1 = urldecode(stripslashes(json_encode($data1)));
//print_r($decodedData1);
$jsonData1 = json_decode($decodedData1, true);
print_r(unicode_urldecode($decodedData1));