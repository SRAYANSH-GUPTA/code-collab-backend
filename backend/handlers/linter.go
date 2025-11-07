package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"codecollab/config"
	"codecollab/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

// InvokeLinter invokes the appropriate Lambda function based on the language
func InvokeLinter(language, code string, cfg *config.Config) ([]models.LintError, error) {
	// Get the appropriate Lambda ARN for the language
	lambdaARN, err := getLambdaARN(language, cfg)
	if err != nil {
		return nil, err
	}

	// If using mock Lambda (for testing), return mock data
	if cfg.UseMockLambda {
		return getMockLintErrors(language), nil
	}

	// Create AWS Lambda client
	lambdaClient, err := createLambdaClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Lambda client: %w", err)
	}

	// Prepare the request payload
	request := models.LambdaRequest{
		Language: language,
		Code:     code,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Invoke the Lambda function
	result, err := lambdaClient.Invoke(context.TODO(), &lambda.InvokeInput{
		FunctionName: aws.String(lambdaARN),
		Payload:      payload,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to invoke Lambda: %w", err)
	}

	// Check for Lambda function errors
	if result.FunctionError != nil {
		return nil, fmt.Errorf("Lambda function error: %s", *result.FunctionError)
	}

	// Parse the response
	var response models.LambdaResponse
	if err := json.Unmarshal(result.Payload, &response); err != nil {
		return nil, fmt.Errorf("failed to parse Lambda response: %w", err)
	}

	return response.Errors, nil
}

// getLambdaARN returns the Lambda ARN for the specified language
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

// createLambdaClient creates an AWS Lambda client with the provided configuration
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

// getMockLintErrors returns mock lint errors for testing purposes
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
