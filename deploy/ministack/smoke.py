#!/usr/bin/env python3
"""Run inside MiniStack: python /opt/codecollab/smoke.py."""
import json
import boto3


client = boto3.client(
    "lambda", endpoint_url="http://localhost:4566", region_name="ap-south-1",
    aws_access_key_id="test", aws_secret_access_key="test",
)
samples = {
    "typescript": ("const x: number = 1;", "const x: number = 'wrong';"),
    "python": ("x = 1\n", "x = unknown_variable\n"),
    "dart": ("void main() { print('hello'); }", "void main() { int x = 'wrong'; print(x); }"),
    "go": ('package main\nimport "fmt"\nfunc main() { fmt.Println("hello") }', 'package main\nfunc main() { var x int = "wrong"; _ = x }'),
    "cpp": ("int main() { return 0; }", 'int main() { int x = "wrong"; return x; }'),
}
for language, cases in samples.items():
    for invalid, code in enumerate(cases):
        result = client.invoke(
            FunctionName=f"codecollab-{language}-linter",
            Payload=json.dumps({"language": language, "code": code}).encode(),
        )
        payload = json.load(result["Payload"])
        assert "FunctionError" not in result, (language, payload)
        assert payload["statusCode"] == 200, (language, payload)
        errors = json.loads(payload["body"])["errors"] or []
        assert bool(errors) == bool(invalid), (language, "invalid" if invalid else "valid", errors)
        assert all(error["line"] > 0 and error["column"] > 0 for error in errors)
        print(f"{language}: {'invalid' if invalid else 'valid'} sample passed ({len(errors)} diagnostics)", flush=True)
