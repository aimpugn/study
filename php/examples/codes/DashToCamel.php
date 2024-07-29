<?php
# https://stackoverflow.com/questions/2791998/convert-string-with-dashes-to-camelcase
function dashesToCamelCase1($string, $capitalizeFirstCharacter = false)
{

    $str = str_replace('-', '', ucwords($string, '-'));

    if (!$capitalizeFirstCharacter) {
        $str = lcfirst($str);
    }

    return $str;
}

echo dashesToCamelCase1('this-is-a-string');

function dashesToCamelCase2($string, $capitalizeFirstCharacter = false)
{

    $str = str_replace('-', '', ucwords($string, '-'));

    if (!$capitalizeFirstCharacter) {
        $str = lcfirst($str);
    }

    return $str;
}

echo dashesToCamelCase2('this-is-a-string');