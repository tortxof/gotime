# ⌚ gotime

A minimalist HTTP microservice that provides timezone-aware current time with
automatic detection of upcoming timezone transitions (e.g., daylight saving time
changes).

## Features

- Returns current time in milliseconds since Unix epoch (compatible with
  JavaScript, Java, etc.)
- Provides timezone offset in seconds
- Detects next timezone transition within 4 weeks
- Zero external dependencies (Go standard library only)
- CORS-enabled for browser requests
- Docker support with minimal image size

## API Documentation

### Endpoint

```
GET /time
```

### Request Headers

| Header | Required | Default | Description |
|--------|----------|---------|-------------|
| `X-Timezone` | No | `UTC` | IANA timezone name (e.g., `America/New_York`, `Europe/London`) |

### Response Schema

**Content-Type**: `application/json`

**Format**: JSON array with 4 elements

```json
[
  <current time in milliseconds>,
  <current offset in seconds>,
  <next transition time in milliseconds>,
  <next offset in seconds>
]
```

| Index | Type | Description | Nullable |
|-------|------|-------------|----------|
| 0 | `number` | Current time in milliseconds since Unix epoch | No |
| 1 | `number` | Current timezone offset in seconds (negative for west of UTC) | No |
| 2 | `number` | Next timezone transition time in milliseconds since Unix epoch | Yes |
| 3 | `number` | New timezone offset after transition in seconds | Yes |

**Notes**:

- Elements at index 2 and 3 are `null` if no timezone transition is found within
  the next 4 weeks
- Timezone offsets are in seconds (e.g., `-18000` for UTC-5, `3600` for UTC+1)
- The service searches up to 4 weeks ahead for timezone transitions

### Response Headers

```
Content-Type: application/json
Cache-Control: no-store
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET,OPTIONS
Access-Control-Allow-Headers: X-Timezone
```

### Examples

#### Request with timezone (DST transition detected)

```bash
curl -H "X-Timezone: America/New_York" http://localhost:8080/time
```

**Response**: `200 OK`

```json
[1739934000000,-18000,1741503600000,-14400]
```

**Interpretation**:

- Current time: `1739934000000` ms (Tue Feb 18 2025 22:00:00 GMT-0500 (Eastern
  Standard Time))
- Current offset: `-18000` seconds (UTC-5, Eastern Standard Time)
- Next transition: `1741503600000` ms (Sun Mar 09 2025 03:00:00 GMT-0400
  (Eastern Daylight Time))
- New offset: `-14400` seconds (UTC-4, Eastern Daylight Time)

#### Request without timezone (defaults to UTC)

```bash
curl http://localhost:8080/time
```

**Response**: `200 OK`

```json
[1739934000000,0,null,null]
```

**Interpretation**:

- Current time: `1739934000000` ms
- Current offset: `0` seconds (UTC)
- No transition: `null` (UTC has no DST transitions)

#### Request with stable timezone (no upcoming transition)

```bash
curl -H "X-Timezone: America/Phoenix" http://localhost:8080/time
```

**Response**: `200 OK`

```json
[1739934000000,-25200,null,null]
```

**Interpretation**:

- Current time: `1739934000000` ms
- Current offset: `-25200` seconds (UTC-7, Arizona doesn't observe DST)
- No transition: `null` (no DST change in next 4 weeks)

### Error Responses

#### Invalid Timezone

**Status**: `400 Bad Request`

**Condition**: The `X-Timezone` header contains an invalid IANA timezone name

**Request**:

```bash
curl -H "X-Timezone: Invalid/Timezone" http://localhost:8080/time
```

**Response**:

```
invalid timezone
```

**Common causes**:

- Misspelled timezone name
- Non-existent timezone
- Invalid format (use IANA names like `America/New_York`, not abbreviations like
  `EST`)

#### Internal Server Error

**Status**: `500 Internal Server Error`

**Condition**: JSON encoding fails (extremely rare)

**Response**:

```
internal error
```

## Running the Service

### Locally

```bash
go run main.go
```

The service listens on `http://localhost:8080`

### With Docker

Build the image:

```bash
docker build -t gotime .
```

Run the container:

```bash
docker run -p 8080:8080 gotime
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `OVERRIDE_CURRENT_TIME` | Override current time (Unix timestamp in seconds) for testing | `1736091600` |

**Example**:

```bash
OVERRIDE_CURRENT_TIME=1739934000 go run main.go
```

## Development

### Run Tests

```bash
go test -v
```

### Run with Coverage

```bash
go test -cover
```

## How It Works

The service uses a binary search algorithm to efficiently find the exact moment
when a timezone transition occurs. When a request is received:

1. Load the requested timezone (or default to UTC)
2. Get the current time in that timezone
3. Check 4 weeks ahead for offset changes
4. If an offset change is detected, use binary search to pinpoint the exact
   second of transition
5. Return current time, current offset, transition time (if any), and new offset
   (if any)

The binary search ensures efficient detection even for timezones with complex
transition rules.

## License

Public domain. See [LICENSE.md](LICENSE.md).
