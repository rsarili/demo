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

const stackName = "iot-demo"

type StackOutputs struct {
	IotCoreEndpoint string
	DeviceEndpoint  string
}

func NewStackOutputs() StackOutputs {
	client := newCloudformationClient()
	fullStackName := getFullStackName()
	deviceEndpointUrlExportName := fullStackName + "-device-registration-endpoint"

	stackOutput, err := client.DescribeStacks(context.TODO(), &cloudformation.DescribeStacksInput{
		StackName: &fullStackName,
	})

	if err != nil {
		log.Fatalf("unable get stack output")
	}
	log.Println(deviceEndpointUrlExportName)

	var ApiGatewayUrl *string
	for _, export := range stackOutput.Stacks[0].Outputs {
		if export.ExportName != nil && *export.ExportName == deviceEndpointUrlExportName {
			ApiGatewayUrl = export.OutputValue
		}
	}
	if ApiGatewayUrl == nil {
		log.Fatalf("can not found api gateway url")
	}
	return StackOutputs{
		IotCoreEndpoint: "",
		DeviceEndpoint:  *ApiGatewayUrl,
	}
}

func TestServerError(t *testing.T) {
	stackOutput := NewStackOutputs()
	deviceId := rand.IntN(1000)
	requestBody := []byte(fmt.Sprintf(`{
		"deviceId": %d
		}`, deviceId))

	log.Printf("POST %s", stackOutput.DeviceEndpoint)
	log.Printf("deviceId: %d", deviceId)
	resp, err := http.Post(stackOutput.DeviceEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("failure")
	}
	assert.Equal(t, resp.StatusCode, http.StatusUnprocessableEntity)

	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	log.Printf("response status code: %d, body: %s", resp.StatusCode, string(responseBody))
}

func TestSuccessfulDeviceRegistration(t *testing.T) {
	stackOutput := NewStackOutputs()
	deviceId := rand.IntN(1000)
	requestBody := []byte(fmt.Sprintf(`{
		"deviceId": "%d",
		"deviceType": "sensor"
		}`, deviceId))

	log.Printf("POST %s", stackOutput.DeviceEndpoint)
	log.Printf("deviceId: %d", deviceId)
	resp, err := http.Post(stackOutput.DeviceEndpoint, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Fatalf("failure")
	}
	assert.Equal(t, resp.StatusCode, http.StatusCreated)

	defer resp.Body.Close()
	responseBody, _ := io.ReadAll(resp.Body)

	log.Printf("response status code: %d, body: %s", resp.StatusCode, string(responseBody))
}

func newCloudformationClient() *cloudformation.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	return cloudformation.NewFromConfig(cfg)
}

func getFullStackName() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("unable to get username, %v", err)
	}

	fullStackName := fmt.Sprintf("%s-%s", stackName, user.Username)
	log.Printf("stack name: %s", fullStackName)

	return fullStackName
}
