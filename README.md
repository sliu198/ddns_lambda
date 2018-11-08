# DDNS Lambda Handler
An AWS Lambda-deployable DNS updater in Golang.
Intended to be used in conjunction with AWS API Gateway Lambda proxy integration.

## Build
```
git submodule update --init --recursive
./build.sh
```
Lambda deployment package will be created at target/handler.zip

## Deploy
Deploy to AWS Lambda and set the following environment variables:
```
HOSTED_ZONE_ID=[hosted zone id this function will modify]
ALLOWED_NAMES=[json array of names that this function is allowed to set]
SHARED_SECRET=[HMAC key used to authorize requests]
```
The Lambda function must have ChangeResourceRecordSets permission.

Create/modify and deploy an API Gateway API with Lambda proxy integration with this function.

## Usage
```
query="ip=[ip]&name=[name]&timestamp=$(date +%s)"
key=[SHARED_SECRET]
curl -s "[base url]?$query&hmac=$(sha256hmac $key $query)"
```

See [here](https://github.com/sliu198/hmac_go) for sha256hmac utility.