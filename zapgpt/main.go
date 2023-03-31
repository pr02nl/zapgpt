package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Response struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int      `json:"created"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Index   int     `json:"index"`
	Message Message `json:"message"`
}

func GenerateGPTText(query string) (string, error) {
	req := Request{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "user",
				Content: query,
			},
		},
		MaxTokens: 150,
	}
	reqJson, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqJson))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer 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")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var response Response
	err = json.Unmarshal(respBody, &response)
	if err != nil {
		return "", err
	}
	return response.Choices[0].Message.Content, nil
}

func parseBase64EncodedBody(body string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return "", err
	}
	data, err := url.ParseQuery(string(decoded))
	if err != nil {
		return "", err
	}
	if data.Has("Body") {
		return data.Get("Body"), nil
	}
	return "", errors.New("no body found")
}

func process(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	result, err := parseBase64EncodedBody(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	gptResult, err := GenerateGPTText(result)
	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return events.APIGatewayProxyResponse{
		Body:       gptResult,
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(process)
}
