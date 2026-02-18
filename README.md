# Real-Time Code Linting Platform


A real-time code linting platform that analyzes code as users type and displays errors instantly, similar to VS Code's IntelliSense.

## Architecture

```
User (Browser)
    ↓
Next.js Frontend (Monaco Editor)
    ↓
WebSocket Connection (with auth token)
    ↓
Golang WebSocket Server (localhost:8080)
    ↓
Mock/Real Linting Engine
    ↓
Display Errors in Monaco Editor
```

## Tech Stack

- **Backend**: Golang with WebSocket (Gorilla WebSocket)
- **Frontend**: Next.js 14 (App Router) with Monaco Editor
- **Authentication**: Mock auth for local testing (Supabase ready)
- **Linting**: Mock linters built-in, AWS Lambda support ready
- **Supported Languages**: TypeScript, Python, Dart, Go, C++

## Project Structure

```
codecollab/
├── backend/                    # Golang WebSocket Server
│   ├── handlers/               # WebSocket, auth, lambda handlers
│   ├── models/                 # Data structures
│   ├── middleware/             # Rate limiting
│   ├── config/                 # Configuration
│   ├── utils/                  # Logging utilities
│   ├── main.go                 # Entry point
│   ├── go.mod                  # Go dependencies
│   └── .env                    # Environment variables
│
├── frontend/                   # Next.js Application
│   ├── app/                    # Next.js 14 app router
│   │   ├── page.tsx            # Landing page
│   │   ├── login/              # Login page
│   │   └── editor/             # Main editor page
│   ├── components/             # React components
│   │   ├── CodeEditor.tsx      # Monaco editor wrapper
│   │   ├── ErrorPanel.tsx      # Error display
│   │   ├── LanguageSelector.tsx
│   │   └── ConnectionStatus.tsx
│   ├── hooks/                  # Custom React hooks
│   │   └── useWebSocket.ts     # WebSocket connection
│   ├── lib/                    # Utilities and types
│   └── package.json
│
└── lambda/                     # AWS Lambda Functions
    └── typescript-linter/      # TypeScript linter
        ├── index.js            # Lambda handler
        ├── test.js             # Local testing
        └── package.json
```

## Quick Start

### 1. Start the Backend

```bash
cd backend

# Install Go dependencies
go mod download

# Run the server
go run main.go
```

The backend will start on `http:

### 2. Start the Frontend

```bash
cd frontend

# Install dependencies
npm install

# Run development server
npm run dev
```

The frontend will start on `http:

### 3. Use the Application

1. Open http:
2. Click "Go to Editor" or "Login"
3. Start typing code in any supported language
4. See errors appear as red squiggles in real-time!

## Features

### Current Features ✅

- [x] WebSocket-based real-time communication
- [x] Monaco Editor integration with syntax highlighting
- [x] Multi-language support (TypeScript, Python, Go, Dart, C++)
- [x] Mock authentication for local testing
- [x] Mock linters with basic error detection
- [x] Rate limiting (60 requests/minute per user)
- [x] Error display with line numbers and severity
- [x] Connection status indicator
- [x] Graceful error handling
- [x] TypeScript Lambda function (ready for deployment)

### Ready for Production 🚀

- [ ] Supabase authentication integration
- [ ] AWS Lambda deployment for real linters
- [ ] Additional language linters (Python, Go, Dart, C++)
- [ ] Database for session storage
- [ ] Metrics and analytics
- [ ] Enhanced UI/UX

## How It Works

### 1. Authentication
- User logs in (mock mode accepts any credentials)
- Auth token stored in localStorage
- Token sent with WebSocket connection

### 2. WebSocket Connection
- Frontend connects to `ws:
- Backend verifies token
- Connection maintained for real-time updates

### 3. Code Analysis Flow
1. User types code in Monaco Editor
2. After 300ms debounce, code is sent via WebSocket
3. Backend receives analysis request
4. Backend routes to appropriate linter (mock or Lambda)
5. Linter analyzes code and returns errors
6. Backend sends errors back via WebSocket
7. Frontend displays errors as markers in Monaco Editor

### 4. Rate Limiting
- 60 requests per minute per user
- Sliding window algorithm
- Prevents abuse and ensures fair usage

## Mock Linters

For local testing, the backend includes mock linters that detect common errors:

**TypeScript:**
- Type mismatches (e.g., `const x: number = 'hello'`)
- Missing function parameters
- Duplicate declarations

**Python:**
- Missing colons in if/function statements
- Invalid function syntax

**Go:**
- Missing function signatures
- Duplicate variable declarations

**Dart:**
- Type declaration with `var` keyword

**C++:**
- Missing function parameter lists

## Environment Variables

### Backend (.env)
```bash
PORT=8080
ENV=development
USE_MOCK_LAMBDA=true    # Use built-in mock linters
USE_MOCK_AUTH=true      # Accept any auth token
```

### Frontend (.env.local)
```bash
NEXT_PUBLIC_WS_URL=ws:
NEXT_PUBLIC_SUPABASE_URL=https:
NEXT_PUBLIC_SUPABASE_ANON_KEY=mock-key
```

## Testing

### Test Backend WebSocket

Using `wscat`:
```bash
npm install -g wscat
wscat -c "ws:
> {"action":"analyze","language":"typescript","code":"const x: number = 'hello';"}
```

### Test Lambda Function Locally

```bash
cd lambda/typescript-linter
npm install
npm test
```

### Health Check

```bash
curl http:
```

## Deployment

### Backend Deployment (AWS EC2)

1. Build the Go binary:
```bash
cd backend
GOOS=linux GOARCH=amd64 go build -o server main.go
```

2. Copy to EC2 and run:
```bash
./server
```

3. Set environment variables:
```bash
export USE_MOCK_LAMBDA=false
export USE_MOCK_AUTH=false
export SUPABASE_URL=<your-url>
export SUPABASE_ANON_KEY=<your-key>
```

### Frontend Deployment (Vercel/Netlify)

1. Build the frontend:
```bash
cd frontend
npm run build
```

2. Deploy to Vercel:
```bash
npm install -g vercel
vercel
```

3. Set environment variables in Vercel dashboard

### Lambda Deployment

See [lambda/typescript-linter/README.md](lambda/typescript-linter/README.md) for detailed instructions.

## API Reference

### WebSocket Endpoint

**Connect:** `ws:

**Send Analysis Request:**
```json
{
  "action": "analyze",
  "language": "typescript",
  "code": "const x: number = 'hello';"
}
```

**Receive Response:**
```json
{
  "type": "analysis_result",
  "errors": [
    {
      "line": 1,
      "column": 20,
      "message": "Type 'string' is not assignable to type 'number'",
      "severity": "error",
      "length": 7
    }
  ],
  "executionTime": 234
}
```

### HTTP Endpoints

- `GET /` - API information
- `GET /health` - Health check
- `GET /ws?token=<token>` - WebSocket upgrade

## Development

### Backend Development

```bash
cd backend
go run main.go
```

Logs will show:
- Server startup
- WebSocket connections
- Analysis requests
- Lambda invocations

### Frontend Development

```bash
cd frontend
npm run dev
```

Features:
- Hot reload
- TypeScript checking
- Console logging for WebSocket events

## Troubleshooting

### WebSocket Connection Failed

1. Check backend is running on port 8080
2. Check `NEXT_PUBLIC_WS_URL` in frontend `.env.local`
3. Check browser console for errors

### No Errors Appearing

1. Check WebSocket connection status (should be green)
2. Check browser console for messages
3. Try the example code with intentional errors

### Rate Limit Error

Wait 60 seconds or restart the backend to reset rate limits.

## Future Enhancements

- [ ] Real-time collaboration
- [ ] Code execution support
- [ ] Syntax highlighting customization
- [ ] Dark/light theme toggle
- [ ] Error fixing suggestions
- [ ] Code formatting
- [ ] File upload/download
- [ ] Session persistence
- [ ] User preferences
- [ ] Admin dashboard

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test locally
5. Submit a pull request

## License

MIT License - feel free to use this project for learning or commercial purposes.

## Support

For issues or questions:
- Check the documentation in each subdirectory
- Review the code comments
- Test with the example code provided

---

**Built for learning and testing WebSocket-based real-time applications with Go and Next.js!**
