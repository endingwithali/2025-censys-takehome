# Backend

## Prerequisites

Before running the backend service, ensure you have:

1. **Go installed** (version 1.24 or higher)
   - Download from [golang.org](https://golang.org/dl/)
   - Verify installation: `go version`

2. **PostgreSQL installed and running**
   - Download from [postgresql.org](https://www.postgresql.org/download/)
   - Start the PostgreSQL service

3. **Database setup completed** (see Database section below)

## Installation & Setup

### 1. Install Dependencies
```bash
go mod download
```

### 2. Create Required Directories

Files will be written to disk in `/snapshot`. 

```bash
mkdir snapshot
```

### 3. Configure Environment

Update the configuration in `cmd/config/config.go` if needed:
- Database connection string
- Snapshot storage path
- Server port (default: 8080)

## To Run

### Commands
To run the backend service 
```bash
go run cmd/main.go
```

### Running Tests
```bash
go test -v ./... 
```

### Running Tests with Coverage
```bash
go test -cover ./...
```


## Database
You will need to create two databases:
- TestDB 
    - Expected Values:
        - `host=localhost`
        - `user=test` 
        - `password=testpassword` 
        - `dbname=censys_testdb` 
        - `port=5432`
        - `sslmode=disable`
        - `TimeZone=UTC`
    - If you wish to change these values for the test DB be sure to update *line 35* in `/internal/repo/snapshot_test.go`
- Production DB
    - Values stored in `cmd/config/config.go`

The values for the DB String are expected to be formatted in the standard [Postgres Connection String Format](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING).

### DB Creation

PostgreSQL must be running on your system. [Installation instructions here](https://www.postgresql.org/download/).

1) Create the DBs:
```bash
$ createdb {censys2025 or censys_testdb}
```
2) Log into the databases
```bash
$ psql {censys2025 or censys_testdb}
```

3. Create the users:
```bash
{dbname}=# CREATE USER backend WITH PASSWORD 'backendpassword';
{dbname}=# GRANT ALL PRIVILEGES ON DATABASE censys2025 TO backend;
```

### SQL Migration
Once you've created a user for the database, run the following command to create the required tables in the databases, for each database you've created. 
```bash 
psql -U {user you created in previous section} -d {censys2025 or censys_testdb} -f internal/repo/schema/schema.sql
```

### Resetting the DB

To clear the DB of existing files:
1) Delete all files in `/snapshot`. 
2) Clear the table in the DB manually
```bash
$ psql {censys2025 or censys_testdb}
{dbname}=# TRUNCATE TABLE snapshot;
```

## Endpoints

### ▶️ GET `/api/health`

Summary: Check if the server is running.

Responses:
- 200: OK
- 500: Internal Server Error   

### ▶️ GET `/api/host/all`

Summary: Get all possible hosts.

Example:
```
GET /api/host/all
```

Responses:
- 200: ListSnapshotsResponse
- 500: Internal Server Error (Unable to get list of hosts)

Response Body:
```json 
{
    [List of all possible hosts IPs as strings]
}
```

### ▶️ GET `/api/host?host={host}`

Summary: Get all timestamps of all snapshots available for a host.

Path Params:
- `host`: string (IPv4/IPv6 of Host)

Example:
```
GET /api/host?host=125.199.235.74
```

Responses:
- 200: ListSnapshotsResponse
- 204: APIError (Snapshot not found in DB or on disk)
- 500: API Error (Unable to create difference)

Response Body:
```json
{
    [List of all timestamps for the snapshots available for a host as strings]
}
```

### ▶️ GET `/api/snapshot?host={host}&at={timestamp}`

Summary: Get snapshot at specific timestamp for a host.

Path Params:
- `host`: string (IPv4/IPv6 of Host)
- `at`: string (timestamp of file)

Example:
```
GET /snapshots?host=125.199.235.74&at=2025-09-10T03:00:00Z
```

Responses:
- 200: ListSnapshotsResponse
- 204: APIError (Snapshot not found in DB or on disk)
- 500: API Error (Unable to create difference)

Response Body:
```json
{
    JSON string of contents of snapshot file
}
```

### ▶️ POST `/api/snapshot`

Summary: Create a snapshot for a host.

Body Params:
- `file`: string (JSON String of body of snapshot file)

Example:
```json
POST /api/snapshot
Content-Type: multipart/form-data
Body:
{
    file: (File)
}
```

Responses:
- 200: Success
- 209: API Error (Snapshot failed to be created by DB)
- 400: API Error (Invalid file format)
- 500: Server Error (Unable to create snapshot)

### ▶️ GET `/api/snapshot/diff?ip={host}&t1={timestamp}&t2={timestamp}`

Summary: Get snapshot differences for a host.
Path Params:
- `host`: string (IPv4/IPv6)
- `t1`: string (timestamp of file 1)
- `t2`: string (timestamp of file 2)

Example:
```
    GET /api/snapshot/diff?ip=125.199.235.74&t1=2025-09-10T03:00:00Z&t2=2025-09-10T03:00:00Z
```

Responses:
- 200: ListSnapshotsResponse
- 204: APIError (Snapshot not found in DB or on disk)
- 500: Internal Server Error (Unable to create difference)

 Response Body:
```json
	{
	  "diffStatus": "FullMatch"|"SupersetMatch"|"NoMatch"|"FirstArgIsInvalidJson"|"SecondArgIsInvalidJson"|"BothArgsAreInvalidJson"|"Invalid"
	  "differences": {Color-coded differences string}
	}
```



