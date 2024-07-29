<?php
class Utf8ize {
    public static function convert($mixed, $fromEncoding = 'auto') {
        if (is_array($mixed)) {
            foreach ($mixed as $key => $value) {
                $mixed[$key] = self::convert($value, $fromEncoding);
            }
        } elseif (is_string($mixed)) {
            return mb_convert_encoding($mixed, 'UTF-8', $fromEncoding);
        }

        return $mixed;
    }
}