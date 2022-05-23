# A simple CRUD Example

## Quick Start

```shell
go run main.go
```

```shell
curl -X POST --location "http://localhost:8080/books" \
    -H "Content-Type: application/json" \
    -d "{
          \"title\": \"Book A\"
        }"
```
Response:

```json
{
  "id": 1,
  "title": "Book A",
  "date_created": "2022-05-23T21:12:13.720048+07:00"
}
```


