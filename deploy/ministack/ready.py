"""Register the linter images after MiniStack starts, including after reboot."""
from pathlib import Path
import boto3

marker = Path("/tmp/codecollab-linters-ready")
marker.unlink(missing_ok=True)
client = boto3.client(
    "lambda", endpoint_url="http://localhost:4566", region_name="ap-south-1",
    aws_access_key_id="test", aws_secret_access_key="test",
)
for language in ("typescript", "python", "dart", "go", "cpp"):
    name = f"codecollab-{language}-linter"
    image = f"codecollab/{language}-linter:local"
    try:
        client.get_function(FunctionName=name)
    except client.exceptions.ResourceNotFoundException:
        client.create_function(
            FunctionName=name, PackageType="Image", Code={"ImageUri": image},
            Role="arn:aws:iam::000000000000:role/codecollab-lambda",
            Timeout=60, MemorySize=512,
        )
    else:
        client.update_function_code(FunctionName=name, ImageUri=image)
        client.get_waiter("function_updated_v2").wait(FunctionName=name)
        client.update_function_configuration(FunctionName=name, Timeout=60, MemorySize=512)
    client.get_waiter("function_active_v2").wait(FunctionName=name)
    client.put_function_concurrency(FunctionName=name, ReservedConcurrentExecutions=1)
    print(f"Ready: {name}", flush=True)
marker.touch()
