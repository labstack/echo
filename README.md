# Bolt
Multi transport, multi format REST style network library

## Socket Specification

### Trasport
- WebSocket
- TCP

### Command (*1-byte*)
- INIT (**1**)
- AUTH (**2**)
- HTTP (**3**)
- PUB (**4**)
- MPUB (**5**)
- SUB (**6**)
- USUB (**7**)

### INIT
> Request

```sh
1  # Command (1-byte)
99 # Correlation ID (4-byte)
8  # Config length (2-byte)
{} # Config as JSON (n-byte)
```
> Config
```js
{
	"Format": 
}
```
> Response

```sh
99 # Correlation ID (4-byte)
200 # Status code (2-byte)
```

### AUTH
> Request

```sh
2                              # Command (1-byte)
99                             # Correlation ID (4-byte)
30                             # Token length (2-byte)
1g42*jMG!a?D3eF>Xwt!dI05]Y9egP # Token (n-byte)
```
> Response

```sh
99  # Correlation ID (4-byte)
200 # Status code (2-byte)
```

### HTTP
> Request

```sh
3        # Command (1-byte)
99       # Correlation ID (4-byte)
GET\n    # Method (n-byte)
/users\n # Path (n-byte)
-        # Headers
64       # Body length (8-byte)
-        # Body (n-byte)
```
> Response

```sh
3   # Command (1-byte)
200 # Status code (2-byte)

# For POST, PUT & PATCH
64  # Body length (8-byte)
-   # Body (n-byte)
```
