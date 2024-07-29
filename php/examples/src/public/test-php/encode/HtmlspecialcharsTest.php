<?php
require_once dirname(__DIR__) . '/autoload.php';

function h($text, $double = true, $charset = null) {
    if (is_array($text)) {
        $texts = array();
        foreach ($text as $k => $t) {
            $texts[$k] = h($t, $double, $charset);
        }
        return $texts;
    } elseif (is_object($text)) {
        if (method_exists($text, '__toString')) {
            $text = (string)$text;
        } else {
            $text = '(object)' . get_class($text);
        }
    } elseif (is_bool($text)) {
        return $text;
    }

    static $defaultCharset = false;
    if ($defaultCharset === false) {
        $defaultCharset = 'UTF-8';
        if ($defaultCharset === null) {
            $defaultCharset = 'UTF-8';
        }
    }
    if (is_string($double)) {
        $charset = $double;
    }
    return htmlspecialchars($text, ENT_QUOTES, ($charset) ? $charset : $defaultCharset, $double);
}


println(h("가나다라"), 'h(가나다라)');
println(h("abcdefg"), 'h(abcdefg)');
println(h("<div>가나다라</div>>"), 'h(<div>가나다라</div>)');
println(h("<가나다라>"), 'h(<가나다라>)');
println(h("[가나다라]"), 'h([가나다라])');
println(h("[가나다라]&"), 'h([가나다라]&)');