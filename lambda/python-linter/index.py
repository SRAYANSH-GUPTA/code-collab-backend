import io
import json
import re
from pyflakes.api import check
from pyflakes.reporter import Reporter


def _response(status_code, errors):
    return {
        "statusCode": status_code,
        "body": json.dumps({"errors": errors}),
    }


def _syntax_error_to_lint(error):
    return {
        "line": error.lineno or 1,
        "column": error.offset or 1,
        "message": error.msg,
        "severity": "error",
        "length": 1,
    }


def _parse_pyflakes_output(output):
    errors = []
    pattern = re.compile(r"^.*?:(\d+):(?:(\d+):)?\s*(.*)$")

    for line in output.splitlines():
        line = line.strip()
        if not line:
            continue

        match = pattern.match(line)
        if not match:
            continue

        line_no = int(match.group(1))
        column = int(match.group(2) or 1)
        message = match.group(3).strip()

        errors.append(
            {
                "line": line_no,
                "column": column,
                "message": message,
                "severity": "warning",
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

    try:
        compile(code, "main.py", "exec")
    except SyntaxError as error:
        return _response(200, [_syntax_error_to_lint(error)])

    stdout = io.StringIO()
    stderr = io.StringIO()
    reporter = Reporter(stdout, stderr)
    check(code, "main.py", reporter)

    combined = "\n".join(part for part in (stdout.getvalue(), stderr.getvalue()) if part)
    return _response(200, _parse_pyflakes_output(combined))
