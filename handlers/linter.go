package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"codecollab/config"
	"codecollab/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
)

func InvokeLinter(language, code string, cfg *config.Config) ([]models.LintError, error) {

	lambdaARN, err := getLambdaARN(language, cfg)
	if err != nil {
		return nil, err
	}

	if cfg.UseMockLambda {
		return getMockLintErrors(language), nil
	}

	lambdaClient, err := createLambdaClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	request := models.LambdaRequest{
		Language: language,
		Code:     code,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	result, err := invokeLambdaWithRetry(context.TODO(), lambdaClient, lambdaARN, payload)

	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	if result.FunctionError != nil {
		return nil, fmt.Errorf("Lambda function error: %s", *result.FunctionError)
	}

	var lambdaAPIResponse struct {
		StatusCode int    `json:"statusCode"`
		Body       string `json:"body"`
	}

	if err := json.Unmarshal(result.Payload, &lambdaAPIResponse); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda API response: %w", err)
	}

	if lambdaAPIResponse.StatusCode != 200 {
		return nil, fmt.Errorf("Lambda returned error status: %d, body: %s", lambdaAPIResponse.StatusCode, lambdaAPIResponse.Body)
	}

	var response models.LambdaResponse
	if err := json.Unmarshal([]byte(lambdaAPIResponse.Body), &response); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda response body: %w", err)
	}

	return response.Errors, nil
}

func invokeLambdaWithRetry(ctx context.Context, lambdaClient *lambda.Client, lambdaARN string, payload []byte) (*lambda.InvokeOutput, error) {
	const maxAttempts = 4

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
			FunctionName: aws.String(lambdaARN),
			Payload:      payload,
		})
		if err == nil {
			return result, nil
		}

		lastErr = err
		if !isLambdaInitializingError(err) || attempt == maxAttempts {
			break
		}

		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}

	return nil, lastErr
}

func isLambdaInitializingError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		if apiErr.ErrorCode() == "CodeArtifactUserPendingException" {
			return true
		}
	}

	return strings.Contains(err.Error(), "Lambda is initializing your function")
}

func getLambdaARN(language string, cfg *config.Config) (string, error) {
	var arn string

	switch language {
	case "typescript", "javascript":
		arn = cfg.LambdaARNTypeScript
	case "python":
		arn = cfg.LambdaARNPython
	case "dart":
		arn = cfg.LambdaARNDart
	case "go", "golang":
		arn = cfg.LambdaARNGo
	case "cpp", "c++":
		arn = cfg.LambdaARNCpp
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	if arn == "" {
		return "", fmt.Errorf("Lambda ARN not configured for language: %s", language)
	}

	return arn, nil
}

func createLambdaClient(cfg *config.Config) (*lambda.Client, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)

	if err != nil {
		return nil, err
	}

	return lambda.NewFromConfig(awsCfg), nil
}

func getMockLintErrors(language string) []models.LintError {
	return []models.LintError{
		{
			Line:     1,
			Column:   1,
			Message:  fmt.Sprintf("Mock error for %s (testing mode)", language),
			Severity: "warning",
			Length:   10,
		},
	}
}
