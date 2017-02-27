package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	region = "ap-northeast-1"
)

func newService() (*dynamodb.DynamoDB, error) {
	conf := &aws.Config{
		Region: &region,
		//Endpoint: &ep,
	}

	sess, err := session.NewSession(conf)
	if err != nil {
		return nil, err
	}

	svc := dynamodb.New(sess)

	return svc, nil
}
