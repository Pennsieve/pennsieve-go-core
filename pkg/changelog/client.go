package changelog

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Client struct {
	client   sqs.Client
	queueUrl string
}

func NewClient(client sqs.Client, queueUrl string) *Client {
	return &Client{client: client, queueUrl: queueUrl}
}

func (c *Client) EmitEvents(ctx context.Context, params Message) error {

	message, err := json.Marshal(params)
	if err != nil {
		return err
	}

	messageInput := sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(c.queueUrl),
	}

	_, err = c.client.SendMessage(ctx, &messageInput)
	if err != nil {
		return err
	}

	return nil

}
