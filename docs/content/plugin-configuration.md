# Plugin Configuration

| Environment Variable        | Is Required? | Description                     | Value Example                            |
|:----------------------------|:-------------|:--------------------------------|:-----------------------------------------|
| GREENER_INGRESS_ENDPOINT    | *Yes*        | Server URL                      | `http://localhost:5096`                  |
| GREENER_INGRESS_API_KEY     | *Yes*        | API key                         | \[API key created in Greener\]           |
| GREENER_SESSION_ID          | *No*         | Session UUIDv4 ID               | `"b7e499fd-f6e1-435c-8ef7-624287ca2bd4"` |
| GREENER_SESSION_DESCRIPTION | *No*         | Session description             | `"My test session"`                      |
| GREENER_SESSION_LABELS      | *No*         | Labels to attach to the session | `"label1=value1,label2"`                 |
| GREENER_SESSION_BAGGAGE     | *No*         | JSON to attach to the session   | `'{"version": "2.0.0"}'`                 |
