package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func main() {
	lambda.Start(Handler)
}

func init() {
	r := gin.Default()

	// create short code
	r.POST("/app", CreateShortURL)

	// access url
	r.GET("/app/:shortcode", GetShortURL)

	// delete short code
	r.DELETE("/app/:shortcode", DeleteShortURL)

	// update short code status
	r.PUT("/app/:shortcode", UpdateStatus)

	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return ginLambda.ProxyWithContext(ctx, req)
}
