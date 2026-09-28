import unittest
from index import _parse_dart_output


class DartDiagnosticsTest(unittest.TestCase):
    def test_machine_output_retains_position_length_and_severity(self):
        errors = _parse_dart_output(
            "Analyzing main.dart...\n"
            "ERROR|COMPILE_TIME_ERROR|INVALID_ASSIGNMENT|/tmp/main.dart|1|23|7|Wrong type\n"
            "WARNING|STATIC_WARNING|UNUSED_LOCAL_VARIABLE|/tmp/main.dart|2|7|1|Unused x\n"
        )
        self.assertEqual(len(errors), 2)
        self.assertEqual(errors[0], {
            "line": 1, "column": 23, "length": 7,
            "severity": "error", "message": "Wrong type",
        })
        self.assertEqual(errors[1]["severity"], "warning")

    def test_empty_or_unrelated_output_has_no_diagnostics(self):
        self.assertEqual(_parse_dart_output("No issues found!"), [])


if __name__ == "__main__":
    unittest.main()
