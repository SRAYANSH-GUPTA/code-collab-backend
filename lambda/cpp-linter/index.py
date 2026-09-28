import json
import os
import re
import subprocess


TMP_FILE = "/tmp/main.cpp"


def _response(status_code, errors):
    return {
        "statusCode": status_code,
        "body": json.dumps({"errors": errors}),
    }


def _parse_compiler_output(output):
    errors = []
    pattern = re.compile(r"^.*?:(\d+):(?:(\d+):)?\s*(fatal error|error|warning):\s*(.*)$")

    for line in output.splitlines():
        line = line.strip()
        if not line:
            continue

        match = pattern.match(line)
        if not match:
            continue

        line_no, column, severity, message = match.groups()
        errors.append(
            {
                "line": int(line_no),
                "column": int(column or 1),
                "message": message.strip(),
                "severity": "error" if severity == "fatal error" else severity,
                "length": 1,
            }
        )

    return errors


def handler(event, context):
    code = (event or {}).get("code", "")
    if not code:
        return _response(
            400,
            [
                {
                    "line": 1,
                    "column": 1,
                    "message": "No code provided",
                    "severity": "error",
                    "length": 1,
                }
            ],
        )

    with open(TMP_FILE, "w", encoding="utf-8") as temp_file:
        temp_file.write(code)

    result = subprocess.run(
        ["g++", "-std=c++17", "-fsyntax-only", TMP_FILE],
        capture_output=True,
        text=True,
        env=os.environ,
    )

    output = "\n".join(part for part in (result.stdout, result.stderr) if part)
    errors = _parse_compiler_output(output)

    if result.returncode != 0 and not errors:
        return _response(
            200,
            [
                {
                    "line": 1,
                    "column": 1,
                    "message": output.strip() or "g++ failed",
                    "severity": "error",
                    "length": 1,
                }
            ],
        )

    return _response(200, errors)
