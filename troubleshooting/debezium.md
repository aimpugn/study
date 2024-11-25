# debezium

## table ${TABLE} whose schema isn't known to this connector. One possible cause is an incomplete database history topic

### 문제

```log
INFO: Connected to 10.16.130.42:3306 at mysql-bin.019927/1232584 (sid:5868, cid:151659)
2023-12-15 03:30:53,725 INFO   MySQL|service_db_main|binlog  Connected to MySQL binlog at 10.16.130.42:3306, starting at MySqlOffsetContext [sourceInfoSchema=Schema{io.debezium.connect
or.mysql.Source:STRUCT}, sourceInfo=SourceInfo [currentGtid=null, currentBinlogFilename=mysql-bin.019927, currentBinlogPosition=1232584, currentRowNumber=0, serverId=0, sourceTime=n
ull, threadId=-1, currentQuery=null, tableIds=[], databaseName=null], snapshotCompleted=false, transactionContext=TransactionContext [currentTransactionId=null, perTableEventCount={
}, totalEventCount=0], restartGtidSet=null, currentGtidSet=null, restartBinlogFilename=mysql-bin.019927, restartBinlogPosition=1232584, restartRowsToSkip=1, restartEventsToSkip=2, c
urrentEventLengthInBytes=0, inTransaction=false, transactionId=null, incrementalSnapshotContext =IncrementalSnapshotContext [windowOpened=false, chunkEndPosition=null, dataCollectio
nsToSnapshot=[], lastEventKeySent=null, maximumKey=null]]   [io.debezium.connector.mysql.MySqlStreamingChangeEventSource]
2023-12-15 03:30:53,726 INFO   MySQL|service_db_main|streaming  Waiting for keepalive thread to start   [io.debezium.connector.mysql.MySqlStreamingChangeEventSource]
2023-12-15 03:30:53,726 INFO   MySQL|service_db_main|binlog  Creating thread debezium-mysqlconnector-service_db_main-binlog-client   [io.debezium.util.Threads]
2023-12-15 03:30:53,727 ERROR  MySQL|service_db_main|binlog  Encountered change event 'Event{header=EventHeaderV4{timestamp=1702369405000, eventType=TABLE_MAP, serverId=11, headerLengt
h=19, dataLength=55, nextPosition=1232717, flags=0}, data=TableMapEventData{tableId=50, database='service_name', table='some_service_table_name', columnTypes=3, 3, 3, 3, 15, -2, -4, -4, -4, 12,
columnMetadata=0, 0, 0, 0, 384, 63233, 2, 2, 2, 0, columnNullability={2, 3, 4, 5, 6, 7, 8}, eventMetadata=null}}' at offset {transaction_id=null, file=mysql-bin.019927, pos=1232584,
 server_id=11, event=1} for table service_name.some_service_table_name whose schema isn't known to this connector. One possible cause is an incomplete database history topic. Take a new snapshot
 in this case.
Use the mysqlbinlog tool to view the problematic event: mysqlbinlog --start-position=1232643 --stop-position=1232717 --verbose mysql-bin.019927   [io.debezium.connector.mysql.MySqlS
treamingChangeEventSource]
2023-12-15 03:30:53,727 ERROR  MySQL|service_db_main|binlog  Error during binlog processing. Last offset stored = {transaction_id=null, file=mysql-bin.019927, pos=1232584, server_id=94
51255, event=1}, binlog reader near position = mysql-bin.019927/1232643   [io.debezium.connector.mysql.MySqlStreamingChangeEventSource]
2023-12-15 03:30:53,727 ERROR  MySQL|service_db_main|binlog  Producer failure   [io.debezium.pipeline.ErrorHandler]
io.debezium.DebeziumException: Error processing binlog event
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.handleEvent(MySqlStreamingChangeEventSource.java:369)
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.lambda$execute$25(MySqlStreamingChangeEventSource.java:860)
        at com.github.shyiko.mysql.binlog.BinaryLogClient.notifyEventListeners(BinaryLogClient.java:1125)
        at com.github.shyiko.mysql.binlog.BinaryLogClient.listenForEventPackets(BinaryLogClient.java:973)
        at com.github.shyiko.mysql.binlog.BinaryLogClient.connect(BinaryLogClient.java:599)
        at com.github.shyiko.mysql.binlog.BinaryLogClient$7.run(BinaryLogClient.java:857)
        at java.base/java.lang.Thread.run(Thread.java:829)
Caused by: io.debezium.DebeziumException: Encountered change event for table service_name.some_service_table_name whose schema isn't known to this connector
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.informAboutUnknownTableIfRequired(MySqlStreamingChangeEventSource.java:654)
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.handleUpdateTableMetadata(MySqlStreamingChangeEventSource.java:633)
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.lambda$execute$13(MySqlStreamingChangeEventSource.java:831)
        at io.debezium.connector.mysql.MySqlStreamingChangeEventSource.handleEvent(MySqlStreamingChangeEventSource.java:349)
        ... 6 more
```

### 원인

- 스키마 체인지가 발생했고, 그걸 데베지움 커넥터에서 문제가 생긴 것으로 보임
- [Debezium MySQL connector error: Encountered change event for table whose schema isn't known to this connector](https://stackoverflow.com/questions/75518352/debezium-mysql-connector-error-encountered-change-event-for-table-whose-schema)
    - 발생 이유: This error happens when binlog fails to capture a DDL
    - 해결 방법: [A year and a half with Debezium: CDC With MySQL](https://tech.bigbasket.com/a-year-and-a-half-with-debezium-f4f323b4909d) 참고

### 해결

- Delete the crashed connector(s), use the command below:

    ```bash
    curl -i -X DELETE http://localhost:8084/connectors/mysql-connector-db3
    ```

- Delete or rename the history topic
- change the connector `snapshot.mode` to `schema_only_recovery`

    ```bash
    curl -i -X POST \
        -H "Accept:application/json" \
        -H  "Content-Type:application/json" \
        http://localhost:8084/connectors/ \
        -d '{
            "name": "mysql-connector-db3",
            "config": {
                "name": "mysql-connector-db3",
                "database.port": "3306",
                "database.user": "db_user_name",
                "connector.class": "io.debezium.connector.mysql.MySqlConnector",
                "rows.fetch.size": "2",
                "table.whitelist": "database1.countries,database1.state,database1.customers,database1.purchase",
                "database.hostname": "db_host1",
                "database.password": "password",
                "database.server.id": "1002",
                "database.whitelist": "database1",
                "database.server.name": "db3",
                "decimal.handling.mode": "string",
                "include.schema.changes": "true",
                "snapshot.select.statement.overrides": "database1.customers,database1.purchase",
                "snapshot.select.statement.overrides.database1.customers": "select *from customers where id < 0",
                "snapshot.select.statement.overrides.database1.purchase": "select* from purchase where id < 0",
                "database.history.kafka.topic": "dbhistory.db3",
                "database.history.kafka.bootstrap.servers": "localhost:9092",
                "database.history.kafka.recovery.attempts": "20",
                "database.history.kafka.recovery.poll.interval.ms": "10000000",
                "snapshot.mode": "schema_only_recovery"
                ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^ 여기를 변경
            }
        }'
    ```

## The db history topic is missing. You may attempt to recover it by reconfiguring the connector to SCHEMA_ONLY_RECOVERY

### 문제

```log
io.debezium.DebeziumException: The db history topic is missing. You may attempt to recover it by reconfiguring the connector to SCHEMA_ONLY_RECOVERY
    at io.debezium.connector.mysql.MySqlConnectorTask.validateAndLoadDatabaseHistory(MySqlConnectorTask.java:363)
    at io.debezium.connector.mysql.MySqlConnectorTask.start(MySqlConnectorTask.java:108)
    at io.debezium.connector.common.BaseSourceTask.start(BaseSourceTask.java:130)
    at org.apache.kafka.connect.runtime.WorkerSourceTask.execute(WorkerSourceTask.java:232)
    at org.apache.kafka.connect.runtime.WorkerTask.doRun(WorkerTask.java:186)
    at org.apache.kafka.connect.runtime.WorkerTask.run(WorkerTask.java:241)
    at java.base/java.util.concurrent.Executors$RunnableAdapter.call(Executors.java:515)
    at java.base/java.util.concurrent.FutureTask.run(FutureTask.java:264)
    at java.base/java.util.concurrent.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1128)
    at java.base/java.util.concurrent.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:628)
    at java.base/java.lang.Thread.run(Thread.java:829)

```

### 원인

- 커넥터를 재생성 해야 한다

### 해결

- 커넥터 삭제 후 재생성
