package main

import (
	"context"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiot"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iot"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"log"
)

type InfraStackProps struct {
	Stage string
	awscdk.StackProps
}

func NewInfraStack(scope constructs.Construct, props InfraStackProps) awscdk.Stack {
	stackName := "iot-demo-" + props.Stage
	stack := awscdk.NewStack(scope, &stackName, &props.StackProps)

	policyName := "IotDevicePolicy"
	awsiot.NewCfnPolicy(stack, jsii.String("IotCorePolicy"), &awsiot.CfnPolicyProps{
		PolicyDocument: map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []map[string]interface{}{
				{
					"Effect":   "Allow",
					"Action":   []string{"iot:Publish", "iot:Connect", "iot:Subscribe", "iot:Receive"},
					"Resource": []string{"*"},
				},
			},
		},
		PolicyName: &policyName,
	})

	certificates_table := awsdynamodb.NewTableV2(stack, jsii.String("CertificatesTable"), &awsdynamodb.TablePropsV2{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("deviceId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName:     jsii.String(*getResourceName(stackName, "CertificatesTable")),
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})
	idempotency_table := awsdynamodb.NewTableV2(stack, jsii.String("IdempotencyTable"), &awsdynamodb.TablePropsV2{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("id"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		TableName:           jsii.String(*getResourceName(stackName, "IdempotencyTable")),
		RemovalPolicy:       awscdk.RemovalPolicy_DESTROY,
		TimeToLiveAttribute: jsii.String("expiration"),
	})

	gateway_handler_function := awslambda.NewFunction(stack, jsii.String("GatewayFunction"), &awslambda.FunctionProps{
		Runtime:      awslambda.Runtime_PYTHON_3_13(),
		Handler:      jsii.String("gateway_handler.handler"),
		FunctionName: getResourceName(stackName, "gateway"),
		Code:         awslambda.Code_FromAsset(jsii.String("../src/lambda/hello_world/dist"), nil),
		Environment: &map[string]*string{
			"IOT_DEVICE_POLICY":  &policyName,
			"CERTIFICATES_TABLE": certificates_table.TableName(),
			"IDEMPOTENCY_TABLE":  idempotency_table.TableName(),
		},
	})

	certificates_table.GrantFullAccess(gateway_handler_function)
	idempotency_table.GrantFullAccess(gateway_handler_function)
	gateway_handler_function.Role().AddToPrincipalPolicy(awsiam.NewPolicyStatement(
		&awsiam.PolicyStatementProps{
			Effect:    awsiam.Effect_ALLOW,
			Actions:   &[]*string{jsii.String("iot:CreateKeysAndCertificate"), jsii.String("iot:AttachPolicy")},
			Resources: &[]*string{jsii.String("*")},
		},
	))

	rest_api := awsapigateway.NewRestApi(stack, jsii.String("RestApi"), &awsapigateway.RestApiProps{
		RestApiName: getResourceName(stackName, "api"),
		DeployOptions: &awsapigateway.StageOptions{
			StageName: jsii.String("v1"),
		},
	})

	certificates_resource := rest_api.Root().AddResource(
		jsii.String("certificates"), nil)
	certificates_resource.AddMethod(jsii.String("POST"),
		awsapigateway.NewLambdaIntegration(gateway_handler_function, nil),
		nil)
	certificates_resource.AddMethod(jsii.String("DELETE"),
		awsapigateway.NewLambdaIntegration(gateway_handler_function, nil),
		nil)

	widget := awscloudwatch.NewLogQueryWidget(&awscloudwatch.LogQueryWidgetProps{
		LogGroupNames: &[]*string{gateway_handler_function.LogGroup().LogGroupName()},
		QueryLines: &[]*string{
			jsii.String("filter level=\"ERROR\""),
			jsii.String("fields user_id, @timestamp, @message, @logStream, @log"),
			jsii.String("sort @timestamp desc"),
			jsii.String("limit 10000")},
		Width: jsii.Number(24),
	})

	awscloudwatch.NewDashboard(stack, jsii.String("Dashboard"), &awscloudwatch.DashboardProps{
		DashboardName: getResourceName(stackName, "dashboard"),
		Widgets:       &[]*[]awscloudwatch.IWidget{{widget}},
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApiEndpoint"), &awscdk.CfnOutputProps{
		Value:      jsii.String(*rest_api.Url() + "/certificates"),
		ExportName: getResourceName(stackName, "device-registration-endpoint"),
	})

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	iot_client := iot.NewFromConfig(cfg)
	res, err := iot_client.DescribeEndpoint(context.TODO(), &iot.DescribeEndpointInput{
		EndpointType: jsii.String("iot:Data-ATS"),
	})
	if err != nil {
		log.Fatalf("unable to get IoT endpoint, %v", err)
	}

	awscdk.NewCfnOutput(stack, jsii.String("IotEndpoint"), &awscdk.CfnOutputProps{
		Value:      res.EndpointAddress,
		ExportName: getResourceName(stackName, "iot-core-endpoint"),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewInfraStack(app, InfraStackProps{
		Stage: app.Node().TryGetContext(jsii.String("stage")).(string),
		StackProps: awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func getResourceName(stackName, lamdaName string) *string {
	return jsii.String(stackName + "-" + lamdaName)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
