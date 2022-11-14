package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RequestBody struct {
	Payload Payload `json:"payload"`
}

type Payload struct {
	Context   string `json:"context"`
	DeployURL string `json:"deploy_ssl_url"`
	ReviewURL string `json:"review_url"`
	Commit    string `json:"commit_ref"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {

	requestBody := RequestBody{}
	json.Unmarshal([]byte(request.Body), &requestBody)

	if requestBody.Payload.Context == "deploy-preview" {
		deployURL := requestBody.Payload.DeployURL
		sha := requestBody.Payload.Commit
		if sha == "" {
			sha = "not-available"
		}
		trigger := os.Getenv("TRIGGER_URL")
		requestURL := fmt.Sprintf("%s?sha=%s&environmentUrl=%s&environmentName=%s", sha, trigger, deployURL, "deploy-preview")
		fmt.Println(requestURL)
		res, err := http.Get(requestURL)
		if err != nil {
			fmt.Printf("error making http request: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("client: got response!\n")
		fmt.Printf("client: status code: %d\n", res.StatusCode)
	} else {
		fmt.Println("context " + requestBody.Payload.Context + " detected, skipping")
	}

	return &events.APIGatewayProxyResponse{
		StatusCode:      200,
		Body:            "Success",
		IsBase64Encoded: false,
	}, nil
}

func main() {
	lambda.Start(handler)
}
