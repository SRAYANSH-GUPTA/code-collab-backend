// check-lambda checks the configured AWS functions without running their code.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"codecollab/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Checking Lambda access (region=%s, backend mock mode=%t)\n", cfg.AWSRegion, cfg.UseMockLambda)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.AWSSessionToken)),
		awsconfig.WithRetryMaxAttempts(1),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Cannot initialize AWS configuration")
		os.Exit(1)
	}
	client := lambda.NewFromConfig(awsCfg, func(options *lambda.Options) {
		if cfg.LambdaEndpointURL != "" {
			options.BaseEndpoint = aws.String(cfg.LambdaEndpointURL)
		}
	})
	failed := false
	for _, fn := range []struct{ name, arn string }{
		{"typescript/javascript", cfg.LambdaARNTypeScript},
		{"python", cfg.LambdaARNPython},
		{"dart", cfg.LambdaARNDart},
		{"go", cfg.LambdaARNGo},
		{"cpp", cfg.LambdaARNCpp},
	} {
		if fn.arn == "" {
			fmt.Printf("%s: ARN missing\n", fn.name)
			failed = true
			continue
		}
		checkCtx, checkCancel := context.WithTimeout(ctx, 15*time.Second)
		_, err := client.Invoke(checkCtx, &lambda.InvokeInput{
			FunctionName: aws.String(fn.arn), InvocationType: types.InvocationTypeDryRun,
		})
		checkCancel()
		if err != nil {
			failed = true
			var apiErr smithy.APIError
			if errors.As(err, &apiErr) {
				fmt.Printf("%s: FAIL (%s)\n", fn.name, apiErr.ErrorCode())
			} else {
				fmt.Printf("%s: FAIL (transport/configuration error: %T)\n", fn.name, err)
			}
			continue
		}
		fmt.Printf("%s: OK (invoke permission verified; function code not executed)\n", fn.name)
	}
	if failed {
		os.Exit(1)
	}
}
