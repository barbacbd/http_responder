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


## Testing

A python script has been added to assist with testing. It was run with python3.8, but is promised to execute with python >= 3. It is possible that the urllib module will not properly import with python2.

The script does not force the user to a specfic version of python at the top of the file with `#!/bin/bash`.

The purpose of this script was NOT to unit test, but to provide with a simple test client that could provide the user with detailed information about the runs (quickly).

If you have any questions, please feel free to ask ...

### Args

1. host name [required] - for the purpose of this test, `localhost` is appropriate unless running from another computer.

2. port [required] - port to connect to

3. number of clients [optional - default = 1] - This is the number of threads and ultimately clients that will be created and potentially attempt to talk to the server simultaneously.

