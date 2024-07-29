<?php
require_once dirname(__DIR__) . '/autoload.php';

$xml1 = new SimpleXMLElement('<xml/>');

print_r($xml1->asXML());

$rawXml1 = <<<XML
<?xml version="1.0" encoding="UTF-8"?><SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://domain.webpay.service.kcp.kr/xsd" xmlns:ns2="http://payment.domain.webpay.service.kcp.kr/xsd" xmlns:ns3="http://webservice.act.webpay.service.kcp.kr"><SOAP-ENV:Body><ns3:approve><ns3:req><ns2:accessCredentialType><ns1:accessLicense></ns1:accessLicense><ns1:signature></ns1:signature><ns1:timestamp></ns1:timestamp></ns2:accessCredentialType><ns2:baseRequestType><ns1:detailLevel>0</ns1:detailLevel><ns1:requestApp>WEB</ns1:requestApp><ns1:requestID>imp_20220610154040</ns1:requestID><ns1:userAgent>Mozilla/5.0 (Linux; Android 12; SM-A725M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.40 Mobile Safari/537.36</ns1:userAgent><ns1:version>0.1</ns1:version></ns2:baseRequestType><ns2:escrow>false</ns2:escrow><ns2:orderID>imp_20220610154040</ns2:orderID><ns2:paymentAmount>1000</ns2:paymentAmount><ns2:paymentMethod>AUTH</ns2:paymentMethod><ns2:productName>Order Name</ns2:productName><ns2:returnUrl>http://localhost:8001/kcp_payments/result</ns2:returnUrl><ns2:siteCode>BA001</ns2:siteCode></ns3:req></ns3:approve></SOAP-ENV:Body></SOAP-ENV:Envelope>
XML;


$xml2 = new SimpleXMLElement($rawXml1);
print_r($xml2->getDocNamespaces());
print_r($xml2->getNamespaces());

/*$xml2->registerXPathNamespace('prefix', 'http://domain.webpay.service.kcp.kr/xsd');
print_r($xml2->xpath('//prefix:requestID'));*/
print_r($xml2->xpath('//ns1:requestID')[0]);
$requestIdXml = $xml2->xpath('//ns1:requestID')[0];
print_r(strval($xml2->xpath('//ns1:requestID')[0]) . PHP_EOL);


$rawXml2 = '<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ns1="http://domain.webpay.service.kcp.kr/xsd" xmlns:ns2="http://payment.domain.webpay.service.kcp.kr/xsd" xmlns:ns3="http://webservice.act.webpay.service.kcp.kr"><SOAP-ENV:Body><ns3:approve><ns3:req><ns2:accessCredentialType><ns1:accessLicense></ns1:accessLicense><ns1:signature></ns1:signature><ns1:timestamp></ns1:timestamp></ns2:accessCredentialType><ns2:baseRequestType><ns1:detailLevel>0</ns1:detailLevel><ns1:requestApp>WEB</ns1:requestApp><ns1:requestID>imp_20220610162253</ns1:requestID><ns1:userAgent>Mozilla/5.0 (Linux; Android 12; SM-A725M) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.40 Mobile Safari/537.36</ns1:userAgent><ns1:version>0.1</ns1:version></ns2:baseRequestType><ns2:escrow>false</ns2:escrow><ns2:orderID>imp_20220610162253</ns2:orderID><ns2:paymentAmount>1000</ns2:paymentAmount><ns2:paymentMethod>AUTH</ns2:paymentMethod><ns2:productName>Order Name</ns2:productName><ns2:returnUrl>http://localhost:8001/kcp_payments/result</ns2:returnUrl><ns2:siteCode>BA001</ns2:siteCode></ns3:req></ns3:approve></SOAP-ENV:Body></SOAP-ENV:Envelope>';

$xml3 = new SimpleXMLElement($rawXml2);
print_r($xml3->xpath('//ns1:requestID')[0]);

$requestId = '12345';
$timestamp = date('Y-m-d\TH:i:s.u');

$rawXml = <<<XML
<?xml version='1.0' encoding='utf-8'?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
    <soapenv:Body>
        <ns:approveResponse xmlns:ns="http://webservice.act.webpay.service.kcp.kr">
            <ns:return xmlns:ax22="http://domain.webpay.service.kcp.kr/xsd"
                       xmlns:ax21="http://payment.domain.webpay.service.kcp.kr/xsd"
                       xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="ax21:ApproveRes">
                <ax21:approvalKey>wmB04F1cfu3U5V/zV/+cKAcHyKIPdQ/iE35VBPEo1cQ=</ax21:approvalKey>
                <ax21:baseResponseType xsi:type="ax22:BaseResponseType">
                    <ax22:detailLevel>0</ax22:detailLevel>
                    <ax22:error xsi:type="ax22:ErrorType">
                        <ax22:code>0000</ax22:code>
                        <ax22:detail></ax22:detail>
                        <ax22:message>Success</ax22:message>
                    </ax22:error>
                    <ax22:messageID>BA001L47SYUHVXX7</ax22:messageID>
                    <ax22:release>0.1</ax22:release>
                    <ax22:requestID>{$requestId}</ax22:requestID>
                    <ax22:responseType>SUCCESS</ax22:responseType>
                    <ax22:timestamp>$timestamp</ax22:timestamp>
                    <ax22:version>0.1</ax22:version>
                    <ax22:warningList xsi:nil="true"/>
                </ax21:baseResponseType>
                <ax21:payUrl>https://testsmpay.kcp.co.kr/pay/mobileGW.kcp</ax21:payUrl>
            </ns:return>
        </ns:approveResponse>
    </soapenv:Body>
</soapenv:Envelope>
XML;

print_r(trim($rawXml));
