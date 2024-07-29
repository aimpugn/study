# mapping

## update field

```json
PUT <INDEX_NAME>/_mappings
{
    "properties": {
        "transactions": {
            "type": "nested",
            "properties": {
                "channel.pg_company": {
                    "type": "keyword"
                }
            }
        }
    }
}
```
