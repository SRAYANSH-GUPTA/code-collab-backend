# Real-Time Code Linting Platform - Build Instructions for Claude Code

## Project Overview

Build a real-time code linting platform that analyzes code as users type and displays errors instantly, similar to VS Code's IntelliSense.

---

## Tech Stack

- **Backend**: Golang WebSocket server (to be deployed on EC2)
- **Frontend**: Next.js 14 (App Router) with Monaco Editor
- **Authentication**: Supabase
- **Linting Engine**: AWS Lambda functions (one per language)
- **Supported Languages**: TypeScript, Python, Dart, Go, C++

---

## Architecture Flow

```
User (Browser)
    ↓
Next.js Frontend (Monaco Editor)
    ↓
WebSocket Connection (with auth token)
    ↓
Golang WebSocket Server (EC2)
    ↓
Supabase Auth Verification
    ↓
AWS Lambda (Language-specific linters)
    ↓
Return Errors
    ↓
Display in Monaco Editor (red squiggles)
```

---

## Detailed Flow

1. User logs in via Supabase authentication in Next.js app
2. User opens Monaco Editor and starts coding
3. Frontend establishes WebSocket connection to Golang server with auth token
4. WebSocket server verifies token with Supabase API
5. As user types, code is sent via WebSocket to Golang server (debounced 300ms)
6. Golang server routes to appropriate Lambda function based on language
7. Lambda runs language-specific linter (static analysis only, no compilation)
8. Errors are returned to Golang server
9. Golang server sends errors back via WebSocket
10. Frontend displays errors in Monaco Editor with red squiggles and error panel

---

## Project Structure

```
project/
├── backend/                        # Golang WebSocket Server
│   ├── main.go                     # Entry point, HTTP server setup
│   ├── handlers/
│   │   ├── websocket.go            # WebSocket connection handler
│   │   ├── auth.go                 # Supabase auth verification
│   │   └── lambda.go               # AWS Lambda invocation logic
│   ├── models/
│   │   └── types.go                # Structs for requests/responses
│   ├── middleware/
│   │   └── ratelimit.go            # Rate limiting (60 req/min per user)
│   ├── config/
│   │   └── config.go               # Environment variable loading
│   ├── utils/
│   │   └── logger.go               # Logging utilities
│   ├── go.mod
│   ├── go.sum
│   ├── .env.example
│   └── README.md
│
├── frontend/                       # Next.js Application
│   ├── app/
│   │   ├── layout.tsx              # Root layout
│   │   ├── page.tsx                # Landing/home page (VERY SIMPLE)
│   │   ├── editor/
│   │   │   └── page.tsx            # Main editor page (SIMPLE UI)
│   │   └── login/
│   │       └── page.tsx            # Authentication page (SIMPLE)
│   ├── components/
│   │   ├── CodeEditor.tsx          # Monaco Editor wrapper component
│   │   ├── ErrorPanel.tsx          # Display linting errors (SIMPLE LIST)
│   │   ├── LanguageSelector.tsx    # Dropdown to select language
│   │   └── ConnectionStatus.tsx    # Show WebSocket status (SIMPLE)
│   ├── hooks/
│   │   └── useWebSocket.ts         # Custom hook for WebSocket connection
│   ├── lib/
│   │   ├── supabase.ts             # Supabase client configuration
│   │   └── types.ts                # TypeScript interfaces
│   ├── styles/
│   │   └── globals.css             # MINIMAL styling, mostly defaults
│   ├── package.json
│   ├── .env.local.example
│   ├── next.config.js
│   ├── tsconfig.json
│   └── README.md
│
└── lambda/                         # AWS Lambda Functions
    ├── typescript-linter/
    │   ├── index.js                # TypeScript linter logic
    │   ├── package.json
    │   └── README.md
    ├── python-linter/
    │   ├── lambda_function.py      # Python linter logic
    │   ├── requirements.txt
    │   └── README.md
    ├── dart-linter/
    │   ├── main.dart               # Dart linter logic
    │   ├── pubspec.yaml
    │   └── README.md
    ├── go-linter/
    │   ├── main.go                 # Go linter logic
    │   ├── go.mod
    │   └── README.md
    └── cpp-linter/
        ├── main.cpp                # C++ linter logic
        └── README.md
```

---

## Backend Requirements (Golang)

### Dependencies

Required Go packages:
```go
github.com/gorilla/websocket       
github.com/aws/aws-sdk-go-v2       
github.com/aws/aws-sdk-go-v2/config
github.com/aws/aws-sdk-go-v2/service/lambda
github.com/joho/godotenv           
```

### main.go

**Should:**
- Load environment variables from .env file
- Start HTTP server on port specified in .env (default: 8080)
- Set up route: `GET /ws` for WebSocket upgrade
- Set up route: `GET /health` for health check
- Implement graceful shutdown on SIGINT/SIGTERM
- Add CORS headers if needed

**Example structure:**
```go
package main

import (
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    
    "your-module/handlers"
    "your-module/config"
)

func main() {
    
    cfg := config.Load()
    
    
    http.HandleFunc("/ws", handlers.HandleWebSocket)
    http.HandleFunc("/health", handlers.HandleHealth)
    
    
    
}
```

### handlers/websocket.go

**Should:**
- Upgrade HTTP connection to WebSocket using gorilla/websocket
- Extract auth token from query parameter: `?token=xxx`
- Call `handlers/auth.go` to verify token with Supabase
- If auth fails: close connection with error message
- If auth succeeds: store connection in map `connectionID -> userID`
- Handle incoming messages in this format:
  ```json
  {
    "action": "analyze",
    "language": "typescript",
    "code": "const x: number = 'hello';"
  }
  ```
- Validate message structure
- Check rate limit before processing
- Call `handlers/lambda.go` to invoke appropriate Lambda
- Send response back via WebSocket:
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
- Handle connection cleanup on disconnect
- Use goroutines for concurrent connection handling
- Add proper error handling and logging

**Connection Storage:**
```go
type Connection struct {
    UserID    string
    WebSocket *websocket.Conn
    LastSeen  time.Time
}

var connections = make(map[string]*Connection)
var mu sync.RWMutex
```

### handlers/auth.go

**Should:**
- Receive auth token as parameter
- Make HTTP POST request to: `https:
- Include headers:
  ```
  Authorization: Bearer {token}
  apikey: {SUPABASE_ANON_KEY}
  ```
- Parse response to get user information
- Return user ID if valid
- Return error if invalid token
- Handle network errors and timeouts (5 second timeout)
- Cache valid tokens temporarily (optional, 5 minute TTL)

**Function signature:**
```go
func VerifyToken(token string) (userID string, err error)
```

### handlers/lambda.go

**Should:**
- Accept language and code as parameters
- Map language to Lambda ARN from environment variables:
  ```
  typescript → LAMBDA_ARN_TYPESCRIPT
  python    → LAMBDA_ARN_PYTHON
  dart      → LAMBDA_ARN_DART
  go        → LAMBDA_ARN_GO
  cpp       → LAMBDA_ARN_CPP
  ```
- Use AWS SDK to invoke Lambda synchronously (RequestResponse)
- Create Lambda payload:
  ```json
  {
    "language": "typescript",
    "code": "..."
  }
  ```
- Set 10 second timeout for Lambda invocation
- Parse Lambda response
- Handle Lambda errors (timeout, execution errors)
- Return errors array

**Function signature:**
```go
func InvokeLinter(language, code string) (errors []LintError, err error)
```

### middleware/ratelimit.go

**Should:**
- Implement in-memory rate limiter
- Track requests per user ID
- Limit: 60 requests per minute per user
- Use sliding window algorithm
- Thread-safe using sync.Map or mutex
- Return true if allowed, false if rate limited

**Function signature:**
```go
func CheckRateLimit(userID string) bool
```

### models/types.go

**Should define:**
```go
type AnalyzeRequest struct {
    Action   string `json:"action"`
    Language string `json:"language"`
    Code     string `json:"code"`
}

type LintError struct {
    Line     int    `json:"line"`
    Column   int    `json:"column"`
    Message  string `json:"message"`
    Severity string `json:"severity"` 
    Length   int    `json:"length"`
}

type AnalyzeResponse struct {
    Type          string      `json:"type"` 
    Errors        []LintError `json:"errors,omitempty"`
    ErrorMessage  string      `json:"message,omitempty"`
    ExecutionTime int         `json:"executionTime,omitempty"` 
}
```

### Environment Variables (.env)

```bash
# Supabase
SUPABASE_URL=https:
SUPABASE_ANON_KEY=your-anon-key

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key

# Lambda ARNs
LAMBDA_ARN_TYPESCRIPT=arn:aws:lambda:us-east-1:123456789:function:typescript-linter
LAMBDA_ARN_PYTHON=arn:aws:lambda:us-east-1:123456789:function:python-linter
LAMBDA_ARN_DART=arn:aws:lambda:us-east-1:123456789:function:dart-linter
LAMBDA_ARN_GO=arn:aws:lambda:us-east-1:123456789:function:go-linter
LAMBDA_ARN_CPP=arn:aws:lambda:us-east-1:123456789:function:cpp-linter

# Server
PORT=8080
ENV=development
```

---

## Frontend Requirements (Next.js)

### IMPORTANT: Keep Frontend VERY SIMPLE

**Design Philosophy:**
- **NO fancy UI** - Use basic HTML elements
- **NO CSS frameworks** - Just minimal inline styles or basic CSS
- **NO complex layouts** - Simple vertical stack
- **Focus on functionality** - Testing WebSocket and Monaco integration

### Dependencies

Required packages:
```json
{
  "dependencies": {
    "next": "^14.0.0",
    "react": "^18.0.0",
    "react-dom": "^18.0.0",
    "@monaco-editor/react": "^4.6.0",
    "@supabase/auth-helpers-nextjs": "^0.8.0",
    "@supabase/supabase-js": "^2.38.0"
  },
  "devDependencies": {
    "@types/node": "^20.0.0",
    "@types/react": "^18.0.0",
    "typescript": "^5.0.0"
  }
}
```

### app/page.tsx (Landing Page)

**Should be EXTREMELY SIMPLE:**
```tsx
'use client'

export default function Home() {
  return (
    <div style={{ padding: '20px', maxWidth: '600px', margin: '0 auto' }}>
      <h1>Code Linter Platform</h1>
      <p>Real-time code linting for multiple languages</p>
      <div style={{ marginTop: '20px' }}>
        <a href="/login" style={{ marginRight: '10px' }}>Login</a>
        <a href="/editor">Go to Editor</a>
      </div>
    </div>
  )
}
```

### app/login/page.tsx (Auth Page)

**Should be VERY SIMPLE:**
- Just email/password input fields (plain HTML inputs)
- Login button
- Sign up button
- NO styling beyond basic margins/padding
- Use Supabase auth helpers
- Redirect to /editor on successful login

```tsx
'use client'

export default function Login() {
  return (
    <div style={{ padding: '20px', maxWidth: '400px', margin: '0 auto' }}>
      <h2>Login</h2>
      <input type="email" placeholder="Email" style={{ display: 'block', width: '100%', marginBottom: '10px', padding: '8px' }} />
      <input type="password" placeholder="Password" style={{ display: 'block', width: '100%', marginBottom: '10px', padding: '8px' }} />
      <button style={{ marginRight: '10px', padding: '8px 16px' }}>Login</button>
      <button style={{ padding: '8px 16px' }}>Sign Up</button>
    </div>
  )
}
```

### app/editor/page.tsx (Main Editor Page)

**Should be VERY SIMPLE:**
- Check if user is authenticated (redirect to /login if not)
- Initialize WebSocket connection
- Render components in simple vertical layout:
  1. Connection status at top (simple text)
  2. Language selector (plain select dropdown)
  3. Monaco Editor (full width)
  4. Error panel at bottom (simple list)
- NO complex CSS, NO flexbox gymnastics
- Just stack elements vertically with basic spacing

### components/CodeEditor.tsx

**Should:**
- Be a client component (`'use client'`)
- Use `@monaco-editor/react`
- Props: `language`, `onChange`, `onMount`
- Default theme: `vs-dark`
- Basic configuration:
  ```tsx
  options={{
    minimap: { enabled: true },
    fontSize: 14,
    lineNumbers: 'on',
    automaticLayout: true,
  }}
  ```
- Debounce onChange by 300ms before calling parent
- Handle editor mount to get editor reference
- Provide method to display errors as markers

**SIMPLE implementation - NO fancy features yet**

### components/ErrorPanel.tsx

**Should be EXTREMELY SIMPLE:**
- Receive `errors` array as prop
- Display as plain unordered list (`<ul>`)
- Each error shows: `Line {line}: {message}`
- NO styling beyond basic list styles
- Click on error to scroll editor to that line (optional)

```tsx
export default function ErrorPanel({ errors }) {
  if (!errors || errors.length === 0) {
    return <div>No errors found</div>
  }
  
  return (
    <div style={{ padding: '10px', border: '1px solid #ccc' }}>
      <h3>Errors ({errors.length})</h3>
      <ul>
        {errors.map((err, i) => (
          <li key={i}>Line {err.line}: {err.message}</li>
        ))}
      </ul>
    </div>
  )
}
```

### components/LanguageSelector.tsx

**Should be EXTREMELY SIMPLE:**
- Plain HTML `<select>` dropdown
- Options: TypeScript, Python, Dart, Go, C++
- Props: `value`, `onChange`
- NO custom styling

```tsx
export default function LanguageSelector({ value, onChange }) {
  return (
    <select value={value} onChange={(e) => onChange(e.target.value)} style={{ padding: '8px', fontSize: '14px' }}>
      <option value="typescript">TypeScript</option>
      <option value="python">Python</option>
      <option value="dart">Dart</option>
      <option value="go">Go</option>
      <option value="cpp">C++</option>
    </select>
  )
}
```

### components/ConnectionStatus.tsx

**Should be EXTREMELY SIMPLE:**
- Just text showing "Connected" or "Disconnected"
- Green text if connected, red if not
- NO animations, NO fancy indicators

```tsx
export default function ConnectionStatus({ isConnected }) {
  return (
    <div style={{ color: isConnected ? 'green' : 'red', padding: '10px' }}>
      {isConnected ? '● Connected' : '○ Disconnected'}
    </div>
  )
}
```

### hooks/useWebSocket.ts

**Should:**
- Accept WebSocket URL and auth token as parameters
- Create WebSocket connection on mount
- Append token to URL: `${url}?token=${token}`
- Handle connection lifecycle:
  - `onopen`: Set connected state to true
  - `onclose`: Set connected state to false
  - `onerror`: Log error
  - `onmessage`: Parse JSON and update errors state
- Provide `sendMessage` function to send analysis requests
- Clean up connection on unmount
- Implement auto-reconnect with exponential backoff (optional for v1)

**Return:**
```ts
{
  isConnected: boolean,
  errors: LintError[],
  sendMessage: (message: object) => void
}
```

### lib/supabase.ts

**Should:**
- Create Supabase client using environment variables
- Export `createClientComponentClient` helper
- Simple setup, no complex logic

```ts
import { createClientComponentClient } from '@supabase/auth-helpers-nextjs'

export const supabase = createClientComponentClient()
```

### Environment Variables (.env.local)

```bash
NEXT_PUBLIC_SUPABASE_URL=https:
NEXT_PUBLIC_SUPABASE_ANON_KEY=your-anon-key
NEXT_PUBLIC_WS_URL=ws:
```

### Frontend Summary

**KEEP IT SIMPLE:**
- ✅ Plain HTML elements
- ✅ Inline styles or minimal CSS
- ✅ No UI frameworks (no Tailwind, no Material UI, no shadcn)
- ✅ No complex layouts
- ✅ Focus on WebSocket + Monaco integration
- ✅ Just enough to test the system

**The goal is to test functionality, NOT to build a beautiful UI**

---

## Lambda Functions

### Start with TypeScript Linter Only

Build ONE working Lambda function first (TypeScript), then expand to others.

### typescript-linter/index.js

**Should:**
- Export handler function: `exports.handler = async (event) => {...}`
- Receive event:
  ```json
  {
    "language": "typescript",
    "code": "const x: number = 'hello';"
  }
  ```
- Write code to `/tmp/temp.ts`
- Run TypeScript compiler API or language server
- Parse diagnostics/errors
- Return response:
  ```json
  {
    "errors": [
      {
        "line": 1,
        "column": 20,
        "message": "Type 'string' is not assignable to type 'number'",
        "severity": "error",
        "length": 7
      }
    ]
  }
  ```
- Handle edge cases (empty code, syntax errors)
- Set timeout to 10 seconds

**Dependencies (package.json):**
```json
{
  "dependencies": {
    "typescript": "^5.0.0"
  }
}
```

**Implementation approach:**
- Use TypeScript's `ts.createProgram()` and `ts.getPreEmitDiagnostics()`
- Parse diagnostics into error format
- Keep it simple - just basic type checking

### Other Linters (Add Later)

**python-linter** - Use `pyright` or `pylint`  
**dart-linter** - Use `dart analyze`  
**go-linter** - Use `gopls` or `go vet`  
**cpp-linter** - Use `clangd` or `cppcheck`

**For initial build: Only create TypeScript linter**

---

## Development Workflow

### Phase 1: Backend Setup (Week 1)

1. Create backend/ directory with Go module
2. Implement basic HTTP server with WebSocket upgrade
3. Add mock auth (just accept any token for now)
4. Test WebSocket connection with a simple client
5. Add Lambda invocation (mock response first)
6. Test end-to-end flow

### Phase 2: Frontend Setup (Week 1)

1. Create Next.js app with TypeScript
2. Add Monaco Editor component
3. Add WebSocket hook
4. Connect to backend
5. Test typing and receiving mock responses
6. Add error display

### Phase 3: Lambda Integration (Week 2)

1. Build TypeScript linter Lambda
2. Deploy to AWS (or test locally with SAM)
3. Connect backend to real Lambda
4. Test real linting
5. Fix any bugs

### Phase 4: Auth Integration (Week 2)

1. Set up Supabase project
2. Add Supabase auth to frontend
3. Implement real auth verification in backend
4. Test complete flow

### Phase 5: Additional Languages (Week 3+)

1. Add Python linter Lambda
2. Add Go linter Lambda
3. Add remaining linters
4. Test each language

---

## Testing Strategy

### Manual Testing Checklist

**Backend:**
- [ ] Server starts without errors
- [ ] WebSocket connection established
- [ ] Auth token verified correctly
- [ ] Invalid token rejected
- [ ] Lambda invoked successfully
- [ ] Errors returned correctly
- [ ] Rate limiting works
- [ ] Connection cleanup on disconnect

**Frontend:**
- [ ] User can log in
- [ ] Monaco Editor loads
- [ ] WebSocket connects
- [ ] Code sends to backend
- [ ] Errors display in editor
- [ ] Language switching works
- [ ] Reconnection works

**Lambda:**
- [ ] Receives correct payload
- [ ] Returns errors in correct format
- [ ] Handles timeout correctly
- [ ] Works with various code samples

---

## What NOT to Include (Keep it Simple)

❌ **Don't add these yet:**
- Docker containers or Docker Compose
- CI/CD pipelines (GitHub Actions, etc.)
- Database for storing code or sessions
- User settings/preferences storage
- Collaborative editing features
- File upload/download
- Code execution (only linting)
- Syntax tree visualization
- Advanced error recovery
- Metrics/analytics
- Admin panel
- Complex styling or UI frameworks
- Server-side rendering optimizations
- Edge functions
- CDN setup
- Load balancers

---

## Success Criteria

The project is complete when:

1. ✅ User can log in via Supabase
2. ✅ Monaco Editor loads and accepts input
3. ✅ WebSocket connection established with auth
4. ✅ Typing TypeScript code triggers analysis
5. ✅ Errors appear in editor as red squiggles
6. ✅ Errors listed in error panel
7. ✅ Can switch between languages
8. ✅ Rate limiting prevents abuse
9. ✅ Connection survives page refresh (after re-auth)
10. ✅ All code is well-structured and documented

---

## Logging Requirements

**Backend should log:**
- Server startup/shutdown
- New WebSocket connections (with user ID)
- Disconnections
- Auth failures
- Lambda invocations (language, duration)
- Rate limit hits
- Errors

**Frontend should log (console):**
- WebSocket connection state changes
- Messages sent/received
- Errors

**Lambda should log:**
- Invocation start
- Code length received
- Analysis duration
- Number of errors found
- Any errors

---

## Code Quality Requirements

- ✅ Proper error handling everywhere (no panics in Go)
- ✅ Comments for complex logic
- ✅ TypeScript strict mode enabled
- ✅ Consistent naming conventions
- ✅ README.md in each major directory
- ✅ .env.example files with dummy values
- ✅ No hardcoded credentials
- ✅ Graceful degradation on errors

---

## Getting Started

### Priority Order

1. **First**: Build backend WebSocket server with mock responses
2. **Second**: Build simple Next.js frontend with Monaco
3. **Third**: Connect them and test WebSocket flow
4. **Fourth**: Build TypeScript Lambda and integrate
5. **Fifth**: Add real Supabase auth
6. **Sixth**: Add other language linters

### Running Locally

**Backend:**
```bash
cd backend
cp .env.example .env
# Edit .env with your credentials
go mod download
go run main.go
# Server runs on http:
```

**Frontend:**
```bash
cd frontend
cp .env.local.example .env.local
# Edit .env.local with your credentials
npm install
npm run dev
# App runs on http:
```

**Lambda (local testing):**
```bash
cd lambda/typescript-linter
npm install
# Test with SAM Local or create test script
```

---

## Final Notes

**Remember:**
- Start simple, iterate quickly
- Test each component independently first
- Don't over-engineer - KISS principle
- Focus on core functionality first
- UI can be improved later
- Document as you go
- Use descriptive commit messages

**The goal is a working prototype, not a production-ready application (yet).**

---

## Questions to Answer During Development

If you encounter these scenarios, here's what to do:

**Q: Should I use TypeScript or JavaScript for Lambda?**  
A: JavaScript for simplicity initially, TypeScript if you want type safety

**Q: Should I add authentication middleware in Go?**  
A: Yes, verify token on every WebSocket message

**Q: Should I cache Lambda responses?**  
A: Not initially - add if needed for performance

**Q: Should I use a WebSocket library in Next.js?**  
A: No, use native WebSocket API - it's sufficient

**Q: Should I deploy to AWS now?**  
A: No, get everything working locally first

**Q: What if Lambda times out?**  
A: Return a timeout error to user: "Analysis timed out, try shorter code"

**Q: What if WebSocket disconnects?**  
A: Auto-reconnect with exponential backoff (add later if needed)

---

## Success Checklist

When you're done, you should be able to:

- [ ] Clone the repo
- [ ] Copy .env files and add credentials
- [ ] Run `go run main.go` in backend/
- [ ] Run `npm run dev` in frontend/
- [ ] Visit http:
- [ ] Log in with Supabase
- [ ] Open editor
- [ ] Type some TypeScript code with an error
- [ ] See red squiggles appear
- [ ] See error in error panel below
- [ ] Switch to Python (even if not implemented yet)
- [ ] See connection status indicator
- [ ] Check browser console - no errors
- [ ] Check backend logs - see connection and analysis

---

## END OF REQUIREMENTS

**Claude Code: Please build this system following the specifications above. Start with the backend, then frontend with a VERY SIMPLE UI, then one Lambda function (TypeScript). Focus on getting a working end-to-end prototype before adding complexity.**

**Remember: SIMPLE UI - just make it functional, not beautiful!**
