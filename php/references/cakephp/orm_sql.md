# Snippets - PHP

- [Snippets - PHP](#snippets---php)
    - [pretty print SQL](#pretty-print-sql)
    - [`find('count', [])`](#findcount-)

## pretty print SQL

```php
public function execute($sql, $options = array(), $params = array()) {
  $options += array('log' => $this->fullDebug);

  $t = microtime(true);
  $this->_result = $this->_execute($sql, $params);

  if ($options['log']) {
   $this->took = round((microtime(true) - $t) * 1000, 0);
   $this->numRows = $this->affected = $this->lastAffected();
   $this->logQuery($sql, $params);
  }
//        $this->printLog();
  return $this->_result;
 }

function printSQL()
{
    if ($this->_result->queryString) {
        App::uses('CakeLog', 'Log');
        CakeLog::write(LOG_INFO, "#####################################################################");
        CakeLog::write(LOG_INFO, "############################### QUERY START #########################");
        $sqlTemp = $this->_result->queryString;
        $sqlTemp = preg_replace("/SELECT\s+/", "SELECT\n  ", $sqlTemp);
        $sqlTemp = str_replace(", `", ",\n  `", $sqlTemp);
        $sqlTemp = str_replace("` FROM", "`\nFROM", $sqlTemp);
        $sqlTemp = preg_replace("/\s*\w*JOIN(.*)/", "\nLEFT JOIN$1", $sqlTemp);
        $sqlTemp = preg_replace("/`\s*WHERE/", "`\nWHERE", $sqlTemp);
        $sqlTemp = preg_replace("/\)\s*WHERE/", ")\nWHERE", $sqlTemp);
        $sqlTemp = preg_replace("/\s*LIMIT/", "\nLIMIT", $sqlTemp);
        $sqlTemp = preg_replace("/\s*\w*ORDER BY(.*)/", "\nORDER BY$1", $sqlTemp);
        CakeLog::write(LOG_INFO, "SQL:\n" . print_r($sqlTemp, true) . "\n");
        CakeLog::write(LOG_INFO, "################################ QUERY END ################################");
        CakeLog::write(LOG_INFO, "###########################################################################\n");
    }
}

function generateCallTrace(){
  $e = new Exception();
  $trace = explode("\n", $e->getTraceAsString());
  // reverse array to make steps line up chronologically
  $trace = array_reverse($trace);
  array_shift($trace); // remove {main}
  array_pop($trace); // remove call to this method
  $length = count($trace);
  $result = array();

  for ($i = 0; $i < $length; $i++)
  {
   $result[] = ($i + 1)  . ')' . substr($trace[$i], strpos($trace[$i], ' ')); // replace '#someNum' with '$i)', set the right ordering
  }

  return "\t" . implode("\n\t", $result);
 }
```

```php
function printSQL()
{
    if ($this->_result->queryString) {
        App::uses('CakeLog', 'Log');
        CakeLog::write(LOG_INFO, "#####################################################################");
        CakeLog::write(LOG_INFO, "############################### QUERY START #########################");
        $sqlTemp = $this->_result->queryString;
        $sqlTemp = preg_replace("/SELECT\s+/", "SELECT\n  ", $sqlTemp);
        $sqlTemp = str_replace(", `", ",\n  `", $sqlTemp);
        $sqlTemp = str_replace("` FROM", "`\nFROM", $sqlTemp);
        $sqlTemp = preg_replace("/\s*(\w*)\s*JOIN(.*)/", "\n$1 JOIN$2", $sqlTemp);
        $sqlTemp = preg_replace("/`\s*WHERE/", "`\nWHERE", $sqlTemp);
        $sqlTemp = preg_replace("/\)\s*WHERE/", ")\nWHERE", $sqlTemp);
        $sqlTemp = preg_replace("/\s*LIMIT/", "\nLIMIT", $sqlTemp);
        $sqlTemp = preg_replace("/\s*\w*ORDER BY(.*)/", "\nORDER BY$1", $sqlTemp);
        CakeLog::write(LOG_INFO, "SQL:\n" . print_r($sqlTemp, true) . "\n");
        CakeLog::write(LOG_INFO, "################################ QUERY END ################################");
        CakeLog::write(LOG_INFO, "###########################################################################\n");
    }
}
```

## `find('count', [])`

```php
// 불필요한 JOIN 제외하고 싶은 경우
// $this->MyModel->unbindModel([
//     'belongsTo' => [
//         'SbcrSchedule',
//         'Payment',
//     ],
// ]);
$result = $this->MyModel->find('count', [
    // `AS count`가 있어야 숫자로 리턴 된다
    // `AS COUNT`처럼 대문자로 사용하면 안된다
    'fields' => 'COUNT(1) AS count',
    'conditions' => [
        'id' => 123456,
    ],
]);
```

```sql
SELECT
  COUNT(1) AS count
FROM
  `some_schema`.`my_models` AS `MyModel`
  INNER JOIN `some_schema`.`another_model` AS `AnotherModel` ON (`MyModel`.`another_id` = `AnotherModel`.`id`)
WHERE
  `id` = 123456
```
