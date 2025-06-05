package networked

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os/user"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"

	"github.com/stretchr/testify/assert"
)

func TestSuccess(t *testing.T) {
	var ApiGatewayUrl string = GetApiEndpoint()
	
	requestBody := []byte(`{
		"title": "Post title",
		"body": "Post description",
		"userId": 1
		}`)
	log.Printf("POST %s", ApiGatewayUrl)
	resp, err := http.Post(ApiGatewayUrl + "todos", "application/json",bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("failure")
	}
	assert.Equal(t, resp.StatusCode, 201)

	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	log.Println(resp.StatusCode)
	log.Println(string(responseBody))
}

func TestServerError(t *testing.T) {
	var ApiGatewayUrl string = GetApiEndpoint()
	userId := rand.IntN(10000)
	requestBody := []byte(fmt.Sprintf(`{
		"title": "Post title",
		"body": "Post description",
		"userId": %d,
		"error": 500
		}`, userId))

	log.Printf("POST %s", ApiGatewayUrl)
	log.Printf("userId %d", userId)
	resp, err := http.Post(ApiGatewayUrl + "todos", "application/json",bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("failure")
	}
	assert.Equal(t, resp.StatusCode, 502)

	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	log.Println(resp.StatusCode)
	log.Println(string(responseBody))
}

func GetApiEndpoint() string {
	user, err := user.Current()
	log.Printf("username: %s", user.Username)
	if err != nil {
		log.Fatalf("unable to get username, %v", err)
	}

	stackName := "iot-demo-" + user.Username
	apiUrlExportName := "iot-demo-" + user.Username + "-api-endpoint"

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	client := cloudformation.NewFromConfig(cfg)

	stackOutput, err := client.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: &stackName,
	})

	if err != nil {
		log.Fatalf("unable get stack output")
	}

	var ApiGatewayUrl *string
	for _, export := range stackOutput.Stacks[0].Outputs {
		if export.ExportName != nil && *export.ExportName == apiUrlExportName {
			ApiGatewayUrl = export.OutputValue
		}
	}
	if ApiGatewayUrl == nil {
		log.Fatalf("can not found api gateway url")
	}
	return *ApiGatewayUrl
}
