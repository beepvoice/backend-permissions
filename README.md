# beep-permissions

Beep backend handling user permissions. Currently, permissions are defined as user-scope (i.e. userid in conversationid). If no such pairing exists, permission is denied. Might consider moving to searchms style user-scope-action system later.

Relations are cached in redis to avoid excessive querying time. A listener updates the cache on database changes.

This service is meant to be used internally. Otherwise, people can systematically query it finding out which conversation a said user is in.

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

| Name | Type | Description |
| ---- | ---- | ----------- |
| userid | String | User's ID |
| conversationid | Conversation ID |

#### Success (200 OK)

Empty body.

#### Errors

It is recommended to intrepet both as a rejection regardless of error type.

| Code | Description |
| ---- | ----------- |
| 401 | Matching userid-conversationid pair not found |
| 500 | Error accessing cache |
