# Running Lambda Functions Locally

## ✅ Your Setup is Now Complete!

The backend is now configured to run the **real TypeScript linter** locally instead of using mock linters.

## How It Works

Instead of calling AWS Lambda, the Go backend now:
1. Spawns a Node.js process
2. Loads the Lambda function from `lambda/typescript-linter/index.js`
3. Passes the code to analyze
4. Receives real TypeScript compiler errors
5. Returns them to the frontend

## Current Configuration

**Backend (.env):**
```bash
USE_MOCK_LAMBDA=false  # ✅ Using real linter
USE_MOCK_AUTH=true     # Still using mock auth
```

## What's Working Now

✅ **TypeScript**: Real linter using TypeScript compiler API
❌ **Python**: Falls back to mock (not implemented yet)
❌ **Go**: Falls back to mock (not implemented yet)
❌ **Dart**: Falls back to mock (not implemented yet)
❌ **C++**: Falls back to mock (not implemented yet)

## Test the Real TypeScript Linter

1. **Go to the editor**: http:
2. **Select TypeScript** from the dropdown
3. **Try this code**:
```typescript
const x: number = 'hello';  
const y: string = 42;        
function test() {
    return "test" + 5;       
}
```

You should see **actual TypeScript compiler errors** with accurate line numbers and messages!

## Switching Between Modes

### Use Real Linter (Current)
```bash
# Edit backend/.env
USE_MOCK_LAMBDA=false
```

### Use Mock Linter
```bash
# Edit backend/.env
USE_MOCK_LAMBDA=true
```

Then restart the backend:
```bash
# Kill the backend
pkill -f "go run main.go"

# Start it again
cd backend && go run main.go
```

## Adding More Languages

To add real linters for other languages, you have two options:

### Option 1: Create More Lambda Functions

Create similar Lambda functions for each language:

```
lambda/
├── typescript-linter/  ✅ Done
├── python-linter/      📝 TODO
├── go-linter/         📝 TODO
├── dart-linter/       📝 TODO
└── cpp-linter/        📝 TODO
```

### Option 2: Call Native Tools Directly

Modify `backend/handlers/lambda.go` to call native linting tools:
- Python: `pylint` or `flake8`
- Go: `go vet` or `golangci-lint`
- Dart: `dart analyze`
- C++: `clang-tidy` or `cppcheck`

## Testing Lambda Locally (Without Backend)

You can test the Lambda function directly:

```bash
cd lambda/typescript-linter
npm test
```

This runs the Lambda with various test cases and shows the output.

## Advantages of Local Lambda

✅ No AWS setup required
✅ No API calls or latency
✅ Free to use
✅ Easy to debug
✅ Works offline
✅ Real linting accuracy

## Disadvantages

❌ Requires Node.js installed
❌ Spawns a new process for each request (slower than mock)
❌ Only TypeScript implemented currently

## Performance

- **Mock Linter**: ~10-50ms
- **Local Lambda**: ~100-300ms (includes Node.js startup)
- **AWS Lambda**: ~200-500ms (includes network latency)

## Next Steps

1. ✅ Test TypeScript with real errors
2. Create Python linter Lambda function
3. Create other language linters
4. Deploy to AWS for production

## Troubleshooting

### "Command 'node' not found"
Install Node.js: `sudo pacman -S nodejs` (Arch Linux)

### "Cannot find module"
Run `npm install` in the Lambda directory:
```bash
cd lambda/typescript-linter
npm install
```

### Backend shows errors
Check the backend logs for detailed error messages. The backend will fall back to mock mode if the Lambda fails.

## Summary

🎉 **You now have a real TypeScript linter running locally!**

- Open http:
- Select TypeScript
- Type code with errors
- See real TypeScript compiler errors!

No AWS account needed, no deployment required!
