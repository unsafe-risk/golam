package golam

import "os"

const (
	envLambdaServerPort = "_LAMBDA_SERVER_PORT"
	envLambdaRuntimeAPI = "AWS_LAMBDA_RUNTIME_API"
)

func isLambdaRuntime() bool {
	return os.Getenv(envLambdaServerPort) != "" || os.Getenv(envLambdaRuntimeAPI) != ""
}
