package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const functionDir = "../function"

type LambdaGolangProxyAPIDemoStackProps struct {
	awscdk.StackProps
}

func NewLambdaGolangProxyAPIDemoStack(scope constructs.Construct, id string, props *LambdaGolangProxyAPIDemoStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	table := awsdynamodb.NewTable(stack, jsii.String("dynamodb-table"),
		&awsdynamodb.TableProps{
			PartitionKey: &awsdynamodb.Attribute{
				Name: jsii.String("shortcode"),
				Type: awsdynamodb.AttributeType_STRING},
		})

	table.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)

	function := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("gin-go-lambda-function"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:     awslambda.Runtime_GO_1_X(),
			Environment: &map[string]*string{"TABLE_NAME": table.TableName()},
			Entry:       jsii.String(functionDir),
		})

	table.GrantReadWriteData(function)

	api := awsapigateway.NewLambdaRestApi(stack, jsii.String("lambda-rest-api"), &awsapigateway.LambdaRestApiProps{
		Handler: function,
	})

	app := api.Root().AddResource(jsii.String("app"), nil)
	app.AddMethod(jsii.String("POST"), nil, nil) // POST /app

	ops := app.AddResource(jsii.String("{shortcode}"), nil)
	ops.AddMethod(jsii.String("GET"), nil, nil)    // GET /app/{shortcode}
	ops.AddMethod(jsii.String("DELETE"), nil, nil) // DELETE /app/{shortcode}
	ops.AddMethod(jsii.String("PUT"), nil, nil)    // PUT /app/{shortcode}

	awscdk.NewCfnOutput(stack, jsii.String("api-gateway-endpoint"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("API-Gateway-Endpoint"),
			Value:      api.Url()})

	awscdk.NewCfnOutput(stack, jsii.String("dynamodb-table-name"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("dynamodb-table-name"),
			Value:      table.TableName()})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewLambdaGolangProxyAPIDemoStack(app, "LambdaGolangProxyAPIDemoStack", &LambdaGolangProxyAPIDemoStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

func env() *awscdk.Environment {
	return nil
}
