/**
 * Local test script for TypeScript linter
 * Run with: node test.js
 */

const { handler } = require('./index.js');

async function test() {
  console.log('Testing TypeScript Linter...\n');

  
  console.log('Test 1: Type mismatch error');
  const result1 = await handler({
    language: 'typescript',
    code: `const x: number = 'hello';`
  });
  console.log('Result:', JSON.stringify(JSON.parse(result1.body), null, 2));
  console.log('');

  
  console.log('Test 2: Valid code');
  const result2 = await handler({
    language: 'typescript',
    code: `const x: number = 42;\nconsole.log(x);`
  });
  console.log('Result:', JSON.stringify(JSON.parse(result2.body), null, 2));
  console.log('');

  
  console.log('Test 3: Syntax error');
  const result3 = await handler({
    language: 'typescript',
    code: `const x: number = ;`
  });
  console.log('Result:', JSON.stringify(JSON.parse(result3.body), null, 2));
  console.log('');

  
  console.log('Test 4: Multiple errors');
  const result4 = await handler({
    language: 'typescript',
    code: `const x: number = 'hello';\nconst y: string = 42;`
  });
  console.log('Result:', JSON.stringify(JSON.parse(result4.body), null, 2));
  console.log('');

  console.log('All tests completed!');
}

test().catch(console.error);
