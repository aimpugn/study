<?php

class HttpParser
{
    /**
     * http 원문을 `headers`와 `body`로 파싱
     * @param $rawText
     * @return array ['headers'=>Header, 'body'=>Body]
     */
    public function parse($rawText)
    {
        $result = [
            'headers' => null,
            'body' => null
        ];
        $rawText = trim($rawText);
        $headerAndBody = $this->splitHeaderAndBody($rawText);
        if (!empty($headerAndBody)) {
            $result['headers'] = $this->getHeaders($headerAndBody[0]);
            $result['body'] = !empty($headerAndBody[1]) ? trim($headerAndBody[1]) : '';
        }

        return $result;
    }

    private function splitHeaderAndBody($rawText)
    {
        # 비표준 HTTP 메시지 경우
        $splitText = explode("\n\n", $rawText);
        if(count($splitText) == 2) {
            return $splitText;
        }

        return explode("\r\n\r\n", $rawText);
    }

    private function getHeaders($headerText)
    {
        $headers = [];
        $headerLines = array_map('trim', explode("\n", $headerText));
        $lineCnt = count($headerLines);
        $cnt = 0;
        if ($lineCnt > 0) {
            $firstLine = $headerLines[$cnt++];
            if (strpos($firstLine, 'HTTP') === false) {
                return $headers;
            }

            $headers['http'] = $firstLine;
            while ($cnt < $lineCnt && ($line = $headerLines[$cnt++]) && ! empty($line)) {
                $splitLine = array_map('trim', explode(":", $line, 2));
                $headers[$splitLine[0]] = $splitLine[1];
            }
        }

        return $headers;
    }
}
