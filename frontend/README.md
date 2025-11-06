# Code Linting Platform - Frontend

Next.js frontend with Monaco Editor for real-time code linting.

## Features

- Monaco Editor integration with syntax highlighting
- Real-time WebSocket connection to backend
- Multi-language support (TypeScript, Python, Go, Dart, C++)
- Error display with red squiggles in editor
- Mock authentication for local testing
- Simple, functional UI

## Setup

1. Install dependencies:
```bash
npm install
```

2. Configure environment variables:
```bash
cp .env.local.example .env.local
# Edit .env.local if needed (default values work for local testing)
```

3. Run the development server:
```bash
npm run dev
```

Open [http:

## Usage

1. Click "Login" or "Go to Editor"
2. If prompted, enter any email/password (mock auth)
3. Select a language from the dropdown
4. Start typing code
5. Errors will appear as red squiggles in the editor
6. Error details are shown in the panel below the editor

## Pages

- `/` - Home page with navigation
- `/login` - Mock login page
- `/editor` - Main code editor with linting

## Mock Authentication

For local testing, the app uses mock authentication:
- Enter any email/password
- A mock token is generated and stored in localStorage
- The token is sent to the backend WebSocket connection
- Backend accepts any token when `USE_MOCK_AUTH=true`

## WebSocket Connection

The editor establishes a WebSocket connection on mount:
- Connection URL: `ws:
- Sends analysis requests on code change (300ms debounce)
- Receives error responses and displays them

## Building for Production

```bash
npm run build
npm start
```

## Tech Stack

- Next.js 14 (App Router)
- TypeScript
- Monaco Editor (VS Code editor component)
- Native WebSocket API
- Supabase auth helpers (for future real auth)
