# Api Gateway

## Krakend

![how to Krakend works](https://raw.githubusercontent.com/castmetal/krakend-private-auth-server-response/main/KrakendFlow.png)

> KrakenD is an extensible, declarative, high-performance open-source API Gateway.

- Its core functionality is to create an API that acts as an aggregator of many microservices into single endpoints, doing the heavy-lifting automatically for you: aggregate, transform, filter, decode, throttle, auth, and more.

- KrakenD needs no programming as it offers a declarative way to create the endpoints. It is well structured and layered, and open to extending its functionality using plug-and-play middleware developed by the community or in-house.

## Summary

This project was built to solve some problems with the security of APIs mainly in the microservices environment, which you can need more control with the flow of information and the access of it.

Look at this: imagine you need to build a new APP, and this app needs to be made with microservices or distributed monolithic services. This app will need more control over the route exposition, who can be allowed to access them, observability with trace id, rate limit, and others. Note it can be a big problem and your application will not be exposed to these gaps for your information.

The solution proposed is a plugin attached to Krakend API Gateway, which will control each request and provide access control requesting another personal API as an authorization provider like oAuth2, or Auth0, but with your rules about the security. You can be able to put some information for your services and isolate your API environment/ecosystem from the public internet. Your services won't have concerns about the access, and it will collect some information about the users, rather than decipher the token for each request. Therefore, all the developers will focus on the problem and how to solve it in the business domain they belong to.

## Install

### Docker build example

- RUN:

```sh
docker build --build-arg ENV=prod --build-arg AUTHORIZER_SERVICE_URL="{your authorizer service url example: http://localhost:8000}" --build-arg API_EXAMPLE_URL="{your service url example: http://localhost:4000}" --build-arg PUBLIC_FLAG="{your endpoint private flag example: /public/}" --build-arg TOKEN_HEADER="example: x-auth" -t mykrakend .
```

- Run docker exec listening 8001 port tcp and exposing

### Without docker

> Install Krakend on link: [Krakend Install](https://www.krakend.io/download/)

For local test, run it on your terminal with krakend install, example:

```sh
ERROR_FLAG="myerror_flag" PUBLIC_FLAG="/public/" FC_ENABLE=1 TOKEN_HEADER="token" AUTHORIZER_SERVICE_URL="http://localhost:8088" API_EXAMPLE_URL="http://localhost:8888" KRAKEND_PORT=8001 krakend run -d -c ./krakend.json -p 8001
```

## How this API Gateway Works

This API Gateway is working with a private server auth provider.

This plugin validates any endpoint with a private flag on URL and send a default request to your private auth service and creates a new header called x-user containing the payload information about your profile customer service.

![how it auth provider plugin works](https://raw.githubusercontent.com/castmetal/krakend-private-auth-server-response/main/autho-provider-plugin.png)

## Configure Endpoints

> To create an endpoint you only need to add an endpoint object under the endpoints list with the resource you want to expose. If there is no method is declared, it’s assumed to be read-only (GET).

The endpoints section looks like this:

```json
{
    "endpoints": [
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
    ]
}
```

- `{{ env "{your_env}" }}` - Read your env file value
- `backend` - your service configuration: host, url_pattern, etc
- `input_headers` - Headers for send to your backend service
- `return_error_details` - Necessary if your output_encoding is json
- `sd` - service discovery, change to dns if you have

> For further information access this link: [https://www.krakend.io/docs/endpoints/creating-endpoints/](https://www.krakend.io/docs/endpoints/creating-endpoints/)

## Run KrakenD with FC_ENABLE=1

Run your krakend:

- `ERROR_FLAG="myerror_flag" PUBLIC_FLAG="/public/" FC_ENABLE=1 TOKEN_HEADER="token" AUTHORIZER_SERVICE_URL="http://localhost:8088" API_EXAMPLE_URL="http://localhost:8888" KRAKEND_PORT=8001 krakend run -d -c ./krakend.json -p 8001`

### Send a request to your private endpoint.

Example, if x-auth is your auth service provider header:

- `curl -H "x-auth: YourToken" -H "api_id: YourApiId" -H "client_id: YourApiId" http://localhost:8001/clients`
