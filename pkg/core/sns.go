package core

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type SnsAPI interface {
	PublishBatch(ctx context.Context, params *sns.PublishBatchInput, optFns ...func(*sns.Options)) (*sns.PublishBatchOutput, error)
}
