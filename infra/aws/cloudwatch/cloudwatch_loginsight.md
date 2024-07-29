# log insight

## parse access log

```sql
fields @timestamp, @message
| parse @message "* * * [* *] * \"* * *\" * * \"*\" * * * \"*\" *" 
    as httpXForwardedFor, empty, remoteUser, requestTimeIso8601, msec, connectionRequests, method, url, version, status, bodyBytesSent, referer, agent, requestTime, upstreamResponseTime, body, httpXAmznTraceId
| display httpXForwardedFor, empty, remoteUser, requestTimeIso8601, msec, connectionRequests, method, url, version, status, bodyBytesSent, referer, agent, requestTime, upstreamResponseTime, body, httpXAmznTraceId
| sort @timestamp desc
| limit 20
```
