# HTTP Responder

The executable http_responder was built with the following steps:

* All instructions assume that the user has moved to the http_responder directory.

```bash
go build .
```

The executable should be named `http_responder`

To run the executable:

```bash
./http_responder
```

The following arguments can be used:

1. a single port

```bash
./http_responder 8080
```

2. Short hand for port

```bash
./http_responder -p 8080
```

3. Long name parameter

```bash
./http_responder --port 8080
```

All other arguments will be ignored. If the port is not an integer it will fail and the program will exit immediately. Providing no arguments will also cause the program to exit immediately.

