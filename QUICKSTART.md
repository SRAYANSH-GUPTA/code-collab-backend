# Quick Start Guide

## Your servers are running! 🚀

### Backend Server
- **URL**: http:
- **WebSocket**: ws:
- **Health Check**: http:
- **Status**: ✅ Running with mock auth and mock linters

### Frontend Application
- **URL**: http:
- **Status**: ✅ Running in development mode

## How to Use

### Step 1: Open the Application
Open your browser and navigate to: http:

### Step 2: Login (Mock Authentication)
1. Click "Go to Editor" or "Login"
2. Enter any email and password (e.g., `test@example.com` / `password`)
3. Click "Login"

### Step 3: Try the Code Editor
1. You'll see the Monaco Editor with TypeScript code
2. Try modifying the code to introduce an error:
   ```typescript
   const x: number = 'hello';  
   ```
3. Watch for:
   - Red squiggly lines in the editor
   - Error details in the panel below
   - Connection status indicator (should be green "● Connected")

### Step 4: Test Different Languages
1. Use the language dropdown to select:
   - TypeScript
   - Python
   - Go
   - Dart
   - C++
2. Each language has different default code with intentional errors
3. Try editing the code to see real-time error detection

## Example Code to Test

### TypeScript (type error)
```typescript
const x: number = 'hello';  
```

### Python (syntax error)
```python
def greet(name)  # Error: Expected ':' at end of function definition
    print(f"Hello, {name}")
```

### Go (syntax error)
```go
func main {  
    fmt.Println("Hello")
}
```

## Troubleshooting

### WebSocket Not Connected (Red Indicator)
1. Make sure backend is running on port 8080
2. Check browser console for errors
3. Try refreshing the page

### No Errors Appearing
1. Check that WebSocket is connected (green indicator)
2. Try the example code above with known errors
3. Check browser console for WebSocket messages

### Can't Access Localhost
- Backend: http:
- Frontend: http:

## What's Happening Behind the Scenes

1. **Authentication**: Mock token generated and stored in localStorage
2. **WebSocket**: Connection established with auth token: `ws:
3. **Code Analysis**:
   - You type code → 300ms debounce → WebSocket sends code
   - Backend receives → Mock linter analyzes → Returns errors
   - Frontend displays errors as red squiggles in Monaco Editor
4. **Rate Limiting**: 60 requests per minute per user

## Next Steps

### To Stop the Servers
- Backend: Press Ctrl+C in the backend terminal
- Frontend: Press Ctrl+C in the frontend terminal

### To Deploy for Production
1. Set `USE_MOCK_AUTH=false` and configure real Supabase
2. Set `USE_MOCK_LAMBDA=false` and deploy Lambda functions
3. Deploy backend to AWS EC2
4. Deploy frontend to Vercel or Netlify

See the main [README.md](README.md) for detailed deployment instructions.

## Testing the Lambda Function Locally

```bash
cd lambda/typescript-linter
npm install
npm test
```

This will test the TypeScript linter with various code samples.

## Project Structure

```
codecollab/
├── backend/          # Go WebSocket server (running on :8080)
├── frontend/         # Next.js app (running on :3000)
└── lambda/           # AWS Lambda functions (for production)
```

## Features Working Out of the Box

✅ Real-time WebSocket connection
✅ Monaco Editor with syntax highlighting
✅ Error detection for 5 languages
✅ Mock authentication
✅ Rate limiting
✅ Error display with line numbers
✅ Connection status indicator
✅ Language switching
✅ Graceful error handling

## Have Fun! 🎉

You now have a fully functional real-time code linting platform running locally!

Try breaking some code and watch the magic happen! ✨
