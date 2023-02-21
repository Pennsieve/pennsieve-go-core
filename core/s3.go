package core

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3API interface {
	HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
}
