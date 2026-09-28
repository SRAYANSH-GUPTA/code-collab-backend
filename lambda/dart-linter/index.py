import json
import os
import subprocess


TMP_FILE = "/tmp/main.dart"


def _response(status_code, errors):
    return {
        "statusCode": status_code,
        "body": json.dumps({"errors": errors}),
    }


def _parse_dart_output(output):
    errors = []

    for line in output.splitlines():
        line = line.strip()
        if not line:
            continue

        # --format machine emits SEVERITY|TYPE|CODE|FILE|LINE|COLUMN|LENGTH|MESSAGE.
        parts = line.split("|", 7)
        if len(parts) != 8 or not all(value.isdigit() for value in parts[4:7]):
            continue
        severity = parts[0].lower()
        line_no, column, length, message = parts[4:8]
        errors.append(
            {
                "line": int(line_no),
                "column": int(column),
                "message": message.strip(),
                "severity": "warning" if severity == "info" else severity,
                "length": max(1, int(length)),
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
        ["dart", "analyze", "--format", "machine", TMP_FILE],
        capture_output=True,
        text=True,
        env={**os.environ, "CI": "true"},
    )

    output = "\n".join(part for part in (result.stdout, result.stderr) if part)
    errors = _parse_dart_output(output)

    if result.returncode not in (0, 2) and not errors:
        return _response(
            200,
            [
                {
                    "line": 1,
                    "column": 1,
                    "message": output.strip() or "dart analyze failed",
                    "severity": "error",
                    "length": 1,
                }
            ],
        )

    return _response(200, errors)
