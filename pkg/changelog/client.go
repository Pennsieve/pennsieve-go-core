package changelog

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	client   sqs.Client
	queueUrl string
}

func NewClient(client sqs.Client, queueUrl string) *Client {
	return &Client{client: client, queueUrl: queueUrl}
}

func (c *Client) EmitEvents(ctx context.Context, params MessageParams) error {

	message, err := json.Marshal(params)
	if err != nil {
		return err
	}

	messageInput := sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(c.queueUrl),
	}

	log.Info(messageInput)

	c.client.SendMessage(ctx, &messageInput)

	return nil

}
