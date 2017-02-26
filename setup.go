package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	region = "ap-northeast-1"
	//ep     = "http://localhost:8000"
)

func createTable(svc *dynamodb.DynamoDB) {
	// パラメータ
	tableInputParams := []*dynamodb.CreateTableInput{
		{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("user_id"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("user_id"),
					KeyType:       aws.String("HASH"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			TableName: aws.String("users"),
		},
		{
			AttributeDefinitions: []*dynamodb.AttributeDefinition{
				{
					AttributeName: aws.String("coupon_id"),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*dynamodb.KeySchemaElement{
				{
					AttributeName: aws.String("coupon_id"),
					KeyType:       aws.String("HASH"),
				},
			},
			ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(1),
				WriteCapacityUnits: aws.Int64(1),
			},
			TableName: aws.String("coupons"),
		},
	}

	for _, p := range tableInputParams {
		resp, err := svc.CreateTable(p)
		if err != nil {
			fmt.Println(*p.TableName, err.Error())
			continue
		}

		fmt.Println(*p.TableName, "created at", resp.TableDescription.CreationDateTime)
	}

}

func main() {
	conf := &aws.Config{
		Region: &region,
		//Endpoint: &ep,
	}

	sess, err := session.NewSession(conf)
	if err != nil {
		panic(err)
	}

	svc := dynamodb.New(sess)
	output, err := svc.ListTables(nil)
	if err != nil {
		panic(err)
	}
	for _, tn := range output.TableNames {
		fmt.Println(*tn)
		// if *tn == "users" {
		// 	fmt.Println("users is created already")
		// 	return
		// }
	}
	createTable(svc)
}
