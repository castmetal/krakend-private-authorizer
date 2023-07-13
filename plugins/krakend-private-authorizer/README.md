# krakend-private-authorizer

![how to auth provider plugin works](https://github.com/castmetal/krakend-private-auth-server-response/blob/main/autho-provider-plugin.png)

- Creating a default http-server plugin for any endpoints to request for a private auth server with your Auth Token Validation

This plugin validates any endpoint with a private flag on URL and send a default request to your private auth service and create a new header called x-user containing the payload information about your profile customer service.

Otherwise, this plugin manipulates the response with the backend header error and payload for any endpoints,

## Configuration

On your krakend.json add the name plugin and compile your .so version into plugins folder

```
{
    "$schema": "https://www.krakend.io/schema/v3.json",
    "version": 3,
    "name": "API Gateway",
    "plugin": {
        "pattern": ".so",
        "folder": "./plugins/"
    },
    "extra_config": {
        "plugin/http-server": {
          "name": [
            "krakend-private-authorizer"
          ],
          "krakend-private-authorizer": {
            "auth_url": "{{ env "AUTHORIZER_SERVICE_URL" }}/v1/token/access",
            "token_header": "{{ env "TOKEN_HEADER" }}",
            "api_id_header": "api_id",
            "client_id_header": "client_id",
            "default_error_message": "Service unavailable",
            "params": [
              "api_id_header",
              "client_id_header",
              "token_header"
            ],
            "modify_errors": true,
            "public_flag": "{{ env "PUBLIC_FLAG" }}",
            "error_flag": "{{ env "ERROR_FLAG" }}"
          }
      }
    }
....

```

### krakend-private-authorizer config:

`auth_url`: Your Auth Provider Service Url, to request private endpoints

`token_header`: Your Token Header, for sending to your Auth Provider Service

`api_id_header`: Your Api Id Header, for sending to your Auth Provider Service

`client_id_header`: Your Client Id Header, for sending to your Auth Provider Service

`params`: Params, an array of headers to send for your Auth Provider Service

`modify_errors`: If true, modify the error response to requesting resources with >=300 status Code, it will take the same response as is in your requested service

`public_flag`: Public Flag to allow endpoints that can’t need an Authorizer Service.

`error_flag`: Error flag name of your: return_error_details, on your Krakend backends configuration

## Compiling

Check your krakend compatibilities,

- `krakend version`
  This command gives to you the go and glibc version that'll you need.

If you received this response: <pre>No incompatibilities found!</pre>:

Run the example code, change with your local configuration.

- `krakend check-plugin --go 1.20.4 --libc GLIBC-2.31 --sum ./go.sum`
- 1.20.4 is needed
- GLIBC-2.31 is your libc version

If you received this response: <pre>No incompatibilities found!</pre>:

Run:

- `go build -x -buildmode=plugin -o krakend-private-authorizer.so .`

- Copy `krakend-private-authorizer.so` to your plugin folder configuration on krakend.

Example:

```
    "plugin": {
        "pattern": ".so",
        "folder": "./plugins/"
    },
```

## Endpoint config

- Add the private flag to set an endpoint as private, and intercept any request to your auth service provider
- Each endpoint with the private flag will validate with the header token passing ons request

Example:

```
{
    "endpoint": "/clients",
    "method": "GET",
    "output_encoding": "json",
    "input_headers": [
        "Authorization",
        "Content-Type",
        "{{ env "TOKEN_HEADER" }}",
        "api_id",
        "client_id",
        "x-request-id",
        "x-remote-ip",
        "x-origin",
        "Accept-Encoding"
    ],
    "backend": [
        {
            "url_pattern": "/v1/example",
            "encoding": "json",
            "method": "GET",
            "host": [
                "{{ env "API_EXAMPLE_URL" }}"
            ],
            "extra_config": {
                "backend/http": {
                    "return_error_details": "{{ env "ERROR_FLAG" }}"
                },
                "qos/http-cache": {
                    "shared": false
                }
            }
        }
    ]
}
```

- Set the input headers to send:
- x-auth: example of your auth header for send to your auth provider service
- x-origin: tracking the origin of request (url endpoint of krakend will send as a header)
- x-remote-ip: the client remote ip
- x-request-id: the unique request-id with an uuid v4

If your endpoint set with private flag:

- Before the endpoint execution, krakend-private-authorizer sends a request to your auth provider service url, with token header set in krakend.json config and intercept request.
- If your auth provider service status code is 200, request will execute in your backend and return the response
- Otherwise, the response the krakend-private-authorizer collect your error details if exists, and sends it with the same status code and payload received to your response
- If the execution of your back-end is ok, the same response will give by your client

### Example of Private Headers

This payload is an example about your private endpoint received, after the auth provider service validate the execution (localhost execution example):

```
{
  host: 'localhost:4000',
  'user-agent': 'KrakenD Version 2.4.1',
  'transfer-encoding': 'chunked',
  'accept-encoding': 'gzip, deflate, br',
  'content-type': 'application/x-www-form-urlencoded',
  'x-auth': '{{your_auth_payload}}',
  'x-b3-sampled': '1',
  'x-b3-spanid': '15d67401ab73f5ba',
  'x-b3-traceid': 'fe7808b00bdabf5893ae659d7e42ced2',
  'x-forwarded-for': '::1',
  'x-forwarded-host': 'localhost:8001',
  'x-origin': '/clients',
  'x-request-id': '88de71e7-c667-4292-8b5f-154b64810dab',
  'x-remote-ip': '',
  'token': ''
}
```

- For validation in your VPC endpoint service, get the JWT token and check the permissions or create more metadata as you need.

## Run KrakenD with FC_ENABLE=1

Run your krakend:

- `ERROR_FLAG="myerror_flag" PUBLIC_FLAG="/public/" FC_ENABLE=1 TOKEN_HEADER="token" AUTHORIZER_SERVICE_URL="http://localhost:8088" API_EXAMPLE_URL="http://localhost:8888" KRAKEND_PORT=8001 krakend run -d -c ./krakend.json -p 8001`
  `
