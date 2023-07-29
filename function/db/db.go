package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var client *dynamodb.Client

// table - shortcode(S, primary key)

const longURLDynamoDBAttributeName = "longurl"
const shortCodeDynamoDBAttributeName = "shortcode"
const activeDynamoDBAttributeName = "active"

var ErrUrlNotFound = errors.New("url not found")
var ErrUrlNotActive = errors.New("url not active")

var table string

func init() {
	table = os.Getenv("TABLE_NAME")
	if table == "" {
		log.Fatal("missing environment variable TABLE_NAME")
	}
	fmt.Println("initializing ddb client for table", table)

	cfg, _ := config.LoadDefaultConfig(context.Background())
	client = dynamodb.NewFromConfig(cfg)

	fmt.Println("ddb client initialized")

}

func SaveURL(longurl string) (string, error) {
	shortCode := uuid.New().String()[:8]

	log.Println("short code", shortCode)

	item := make(map[string]types.AttributeValue)

	item[longURLDynamoDBAttributeName] = &types.AttributeValueMemberS{Value: longurl}
	item[shortCodeDynamoDBAttributeName] = &types.AttributeValueMemberS{Value: shortCode}
	item[activeDynamoDBAttributeName] = &types.AttributeValueMemberBOOL{Value: true}

	_, err := client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item})

	if err != nil {
		log.Println("dynamodb put item failed")
		return "", err
	}

	log.Printf("short url for %s - %s\n", longurl, shortCode)
	return shortCode, nil
}

func GetLongURL(shortCode string) (string, error) {

	op, err := client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			shortCodeDynamoDBAttributeName: &types.AttributeValueMemberS{Value: shortCode}}})

	if err != nil {
		log.Println("failed to get long url", err)
		return "", err
	}

	if op.Item == nil {
		return "", ErrUrlNotFound
	}

	activeAV := op.Item[activeDynamoDBAttributeName]
	active := activeAV.(*types.AttributeValueMemberBOOL).Value

	if !active {
		return "", ErrUrlNotActive
	}

	longurlAV := op.Item[longURLDynamoDBAttributeName]
	longurl := longurlAV.(*types.AttributeValueMemberS).Value

	log.Println("long url", longurl)

	return longurl, nil
}

// Update enables or disables a short url
func Update(shortCode string, status bool) error {

	update := expression.Set(expression.Name(activeDynamoDBAttributeName), expression.Value(status))
	updateExpression, _ := expression.NewBuilder().WithUpdate(update).Build()

	condition := expression.AttributeExists(expression.Name(shortCodeDynamoDBAttributeName))
	conditionExpression, _ := expression.NewBuilder().WithCondition(condition).Build()

	_, err := client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			shortCodeDynamoDBAttributeName: &types.AttributeValueMemberS{Value: shortCode}},
		UpdateExpression:          updateExpression.Update(),
		ExpressionAttributeNames:  updateExpression.Names(),
		ExpressionAttributeValues: updateExpression.Values(),
		ConditionExpression:       conditionExpression.Condition(),
	})

	if err != nil && strings.Contains(err.Error(), "ConditionalCheckFailedException") {
		return ErrUrlNotFound
	}

	return err
}

// Delete removes a short url from the table
func Delete(shortCode string) error {

	condition := expression.AttributeExists(expression.Name(shortCodeDynamoDBAttributeName))
	conditionExpression, _ := expression.NewBuilder().WithCondition(condition).Build()

	_, err := client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			shortCodeDynamoDBAttributeName: &types.AttributeValueMemberS{Value: shortCode}},
		ConditionExpression:       conditionExpression.Condition(),
		ExpressionAttributeNames:  conditionExpression.Names(),
		ExpressionAttributeValues: conditionExpression.Values()})

	if err != nil && strings.Contains(err.Error(), "ConditionalCheckFailedException") {
		return ErrUrlNotFound
	}

	return err
}
