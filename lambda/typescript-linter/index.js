const ts = require('typescript');
const fs = require('fs');
const path = require('path');

/**
 * AWS Lambda handler for TypeScript linting
 *
 * Event format:
 * {
 *   "language": "typescript",
 *   "code": "const x: number = 'hello';"
 * }
 *
 * Response format:
 * {
 *   "errors": [
 *     {
 *       "line": 1,
 *       "column": 20,
 *       "message": "Type 'string' is not assignable to type 'number'",
 *       "severity": "error",
 *       "length": 7
 *     }
 *   ]
 * }
 */
exports.handler = async (event) => {
  console.log('TypeScript linter invoked');
  console.log('Code length:', event.code?.length || 0);

  try {
    const { code } = event;

    if (!code) {
      return {
        statusCode: 400,
        body: JSON.stringify({
          errors: [{
            line: 1,
            column: 1,
            message: 'No code provided',
            severity: 'error',
            length: 1
          }]
        })
      };
    }

    
    const tempFile = path.join('/tmp', 'temp.ts');
    fs.writeFileSync(tempFile, code);

    
    const program = ts.createProgram([tempFile], {
      noEmit: true,
      target: ts.ScriptTarget.ES2015,
      module: ts.ModuleKind.CommonJS,
      strict: true,
      esModuleInterop: true,
      skipLibCheck: true,
    });

    
    const diagnostics = [
      ...program.getSyntacticDiagnostics(),
      ...program.getSemanticDiagnostics(),
    ];

    
    const errors = diagnostics.map((diagnostic) => {
      const message = ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n');

      let line = 1;
      let column = 1;
      let length = 1;

      if (diagnostic.file && diagnostic.start !== undefined) {
        const { line: lineNum, character } = diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start);
        line = lineNum + 1; 
        column = character + 1; 
        length = diagnostic.length || 1;
      }

      return {
        line,
        column,
        message,
        severity: diagnostic.category === ts.DiagnosticCategory.Error ? 'error' : 'warning',
        length,
      };
    });

    console.log('Analysis complete:', errors.length, 'errors found');

    
    try {
      fs.unlinkSync(tempFile);
    } catch (e) {
      console.warn('Failed to delete temp file:', e.message);
    }

    return {
      statusCode: 200,
      body: JSON.stringify({ errors })
    };

  } catch (error) {
    console.error('Error during linting:', error);

    return {
      statusCode: 500,
      body: JSON.stringify({
        errors: [{
          line: 1,
          column: 1,
          message: 'Internal linter error: ' + error.message,
          severity: 'error',
          length: 1
        }]
      })
    };
  }
};
