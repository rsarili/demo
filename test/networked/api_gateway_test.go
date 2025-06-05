package networked

import (
	"context"
	"io"
	"log"
	"net/http"
	"os/user"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	var ApiGatewayUrl string = GetApiEndpoint()
	log.Printf("sending request to %s", ApiGatewayUrl)

	resp, err := http.Get(ApiGatewayUrl + "todos")
	if err != nil {
		log.Fatalf("failure")
	}
	assert.Equal(t, resp.StatusCode, 200)

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	log.Println(resp.StatusCode)
	log.Println(string(body))
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
