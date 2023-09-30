# Wex Tag 

## How to run the project

```bash
cd src && go run main.go
```

A web server will start on port `3333`.

## Summary

My idea for this task was to keep it simple and try to separate the concerns as best as possible. Since only stdlib was used, I created a persistance module that locally saves the data to a json file. This module could be easily swapped by a db driver. On a real application I would probably use an external module to deal with monetary values so I tried to implement the integer logic as simple as possible.

The application is divided in four major modules: `application` where the business logic resides, `external` which implements the interaction with the treasury api, `persistance` that locally stores the transactions and the `main` package where we have the api handlers.


# Implemented endpoints:

## /

- Methods supported:
    - GET

Serves basic static form to submit request to `/registerTransaction` endpoint.

## /registerTransaction

- Methods supported:
    - POST

### Request

| Field Name  | Type   | About                 |
|-------------|--------|-----------------------|
| amount      | string | Comma-separated value |
| date        | string | / or - as separator   |
| description | string | Less than 50 ch.      |

Example request:

```bash
curl -X POST http://localhost:3333/registerTransaction -H "Content-Type: application/x-www-form-urlencoded"  -d "amount=2.56&date=30/09/2009&description=test" 
```

### Response

- `"Content-Type" : "application/json"`

| Field Name    | Type   | Constrains |
|---------------|--------|------------|
| transactionId | string |Transaction identifier|

Example response:
```json
{
    "transactionId":"08AADEDE-F0A7-A66A-B28C-A31D94A93C8D"
}
```


## /queryTransaction

- Methods supported:
    - GET

### Request

| Field Name    | Type   | Constrains |
|---------------|--------|------------|
| transactionId | string |            |

Example request:

```
http://localhost:3333/queryTransaction?transactionId=70ABEBB4-50F9-C36D-F524-A7C46B082B17
```

### Response

- `"Content-Type" : "application/json"`

| Field Name  | Type   | About                |
|-------------|--------|----------------------|
| description | string |                      |
| date        | string | YYYY-MM-DDThh:mm:ssZ |
| amount      | string | Value in USD         |
| uid         | string |Transaction identifier|

Example response:

```json
{
    "description": "Transaction Example",
    "date": "1998-08-01T00:00:00Z",
    "amount": "1.99",
    "uid": "70ABEBB4-50F9-C36D-F524-A7C46B082B17"
}
```

## /convertTransaction

- Methods supported:
    - GET

### Request

| Field Name    | Type   | About                  |
|---------------|--------|------------------------|
| transactionId | string | Transaction identifier |
| country       | string | Currency's country     |
| currency      | string | Desired currency       |

Example request:

```
http://localhost:3333/convertTransaction?transactionId=182D05C0-DCC8-3EEC-119A-FB708B0A6BB8&country=Mexico&currency=Peso
```


### Response

- `"Content-Type" : "application/json"`

| Field Name     | Type   | About                |
|----------------|--------|----------------------|
| description    | string |                      |
| transactionDate| string | YYYY-MM-DDThh:mm:ssZ |
| uid            | string | Transaction identifier                     |
| convertedValue | string | Value in requested currency                     |
| exchangeRate   | string | Exchange rate used|
| originalValue  | string | Value in USD         |

Example response:

```json
{
    "convertedValue": "1776.83",
    "description": "Sample Transaction",
    "exchangeRate": "17.77",
    "originalValue": "99.99",
    "transactionDate": "1998-05-01",
    "uid": "182D05C0-DCC8-3EEC-119A-FB708B0A6BB8"
}
```

## Remarks

- application suited for low request volume
- currently the file is save for each request, but could be adapted to save after a certain amount of requests received or after a time period

## Testing

Unit testing can be executed with:

```bash
cd src && go test ./...  -coverprofile=coverage.out
```

Results:
```
ok      wex/src 1.008s  coverage: 69.3% of statements
ok      wex/src/application     0.003s  coverage: 91.8% of statements
ok      wex/src/external        0.004s  coverage: 90.0% of statements
ok      wex/src/persistance     0.504s  coverage: 74.0% of statements
```

- the integration with the external treasury api would be better tested within an integration test.