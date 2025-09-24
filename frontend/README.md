# Host Snapshot Manager Frontend

A React-based frontend application for managing and comparing host snapshots.

## Features

- Upload JSON snapshot files with drag-and-drop support
- View all available hosts from uploaded snapshots
- Browse available timestamps for each host
- Display JSON content of selected snapshots
- Compare differences between two snapshots of the same host

## Getting Started

### Prerequisites

- Node.js (v14 or higher)
- npm or yarn
- Backend server running on `http://localhost:8080`

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm start
```

The application will open in your browser at `http://localhost:3000`.

### Usage

1. Use the upload section to drag and drop or select JSON files. Files must follow the naming convention: `host_<ip>_<timestamp>.json`
2. After uploading, available hosts will appear in the left panel
3. Click on a host to see all available timestamps for that host
4. Click on a timestamp to view the JSON content of that snapshot
5. Click "Show Diff" and select another timestamp to compare differences between snapshots

## API Integration

The frontend communicates with the backend API at `http://localhost:8080/api`:

- `GET /api/host/all` - Get all available hosts
- `GET /api/host?ip=<host>` - Get timestamps for a specific host
- `GET /api/snapshot?ip=<host>&at=<timestamp>` - Get snapshot content
- `POST /api/snapshot` - Upload a new snapshot file
- `GET /api/snapshot/diff?ip=<host>&t1=<timestamp1>&t2=<timestamp2>` - Get differences between snapshots

To change the backend API location, update the value of the const `API_BASE_URL` in `App.js` and `FileUpload.js`.

## File Format

Uploaded files must follow this naming convention:
```
host_<ip_address>_<YYYY-MM-DD>T<HH-MM-SS>[.fraction](Z|±HH-MM).json
```
Example: `host_125.199.235.74_2025-09-15T08-49-45Z.json`
