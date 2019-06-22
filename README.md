# beep-permissions

Beep backend handling user permissions. Currently, permissions are defined as user-scope (i.e. userid in conversationid). If no such pairing exists, permission is denied. Might consider moving to searchms style user-scope-action system later.

Relations are cached in redis to avoid excessive querying time. A listener updates the cache on database changes.

## Environment variables

Supply environment variables by either exporting them or editing `.env`.

| ENV | Description | Default |
| --- | ----------- | ------- |
| LISTEN | Host and port for service to listen on | :80 |
| POSTGRES | URL of postgres | postgresql://root@pg:5432/core?sslmode=disable |
| REDIS | URL of redis | redis:6379 |

## API

| Contents |
| -------- |
| Get Permission |

---

### Get Permission

```
GET /user/:userid/conversation/:conversationid
```

Query to see if userid-conversationid is permissable.

#### Params

#### Success (200 OK)

Empty body.

#### Errors
