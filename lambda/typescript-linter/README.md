# TypeScript Linter Lambda Function

AWS Lambda function for linting TypeScript code using the TypeScript compiler API.

## Features

- Uses TypeScript's built-in compiler for accurate type checking
- Returns syntax and semantic errors
- Provides line, column, and error message details
- Handles edge cases (empty code, syntax errors, etc.)

## Local Testing

1. Install dependencies:
```bash
npm install
```

2. Run test script:
```bash
npm test
```

This will test the linter with various code samples.

## Deployment to AWS Lambda

### Option 1: Using AWS Console

1. Install dependencies:
```bash
npm install --production
```

2. Create a ZIP file:
```bash
zip -r typescript-linter.zip index.js package.json node_modules/
```

3. Upload to AWS Lambda:
   - Go to AWS Lambda Console
   - Create new function (Node.js 18.x or later)
   - Upload the ZIP file
   - Set handler to `index.handler`
   - Set timeout to 10 seconds
   - Set memory to 256 MB

### Option 2: Using AWS CLI

```bash
# Install dependencies
npm install --production

# Create ZIP
zip -r typescript-linter.zip index.js package.json node_modules/

# Create Lambda function
aws lambda create-function \
  --function-name typescript-linter \
  --runtime nodejs18.x \
  --role arn:aws:iam::YOUR_ACCOUNT_ID:role/lambda-execution-role \
  --handler index.handler \
  --zip-file fileb:
  --timeout 10 \
  --memory-size 256

# Update function (if already exists)
aws lambda update-function-code \
  --function-name typescript-linter \
  --zip-file fileb:
```

## Input Format

```json
{
  "language": "typescript",
  "code": "const x: number = 'hello';"
}
```

## Output Format

```json
{
  "statusCode": 200,
  "body": "{\"errors\":[{\"line\":1,\"column\":20,\"message\":\"Type 'string' is not assignable to type 'number'\",\"severity\":\"error\",\"length\":7}]}"
}
```

## Error Categories

- **Syntax Errors**: Missing semicolons, brackets, etc.
- **Type Errors**: Type mismatches, undefined variables, etc.
- **Semantic Errors**: Unreachable code, duplicate declarations, etc.

## Performance

- Typical execution time: 50-200ms
- Memory usage: ~50-100 MB
- Timeout set to 10 seconds for safety
