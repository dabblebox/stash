package s3

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dabblebox/stash/component/service/aws/policy"
)

func CreateBucket(bucket string, policy policy.Policy, tags map[string]string, svc *s3.S3) error {

	ctx := context.Background()

	if _, err := svc.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return err
	}

	if err := svc.WaitUntilBucketExistsWithContext(ctx,
		&s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		},
		request.WithWaiterDelay(request.ConstantWaiterDelay(10*time.Second)),
	); err != nil {
		return fmt.Errorf("tired of waiting: %v", err)
	}

	if _, err := svc.PutPublicAccessBlockWithContext(ctx, &s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	}); err != nil {
		return err
	}

	b, err := json.Marshal(policy)
	if err != nil {
		return err
	}

	if _, err := svc.PutBucketPolicyWithContext(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(string(b)),
	}); err != nil {
		return err
	}

	tagSet := []*s3.Tag{}
	for k, v := range tags {
		tagSet = append(tagSet, &s3.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	if _, err := svc.PutBucketTaggingWithContext(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucket),
		Tagging: &s3.Tagging{
			TagSet: tagSet,
		},
	}); err != nil {
		return err
	}

	serverConfig := &s3.ServerSideEncryptionConfiguration{
		Rules: []*s3.ServerSideEncryptionRule{{
			ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
				SSEAlgorithm: aws.String(s3.ServerSideEncryptionAwsKms),
			},
		}},
	}

	if _, err := svc.PutBucketEncryption(&s3.PutBucketEncryptionInput{
		Bucket:                            aws.String(bucket),
		ServerSideEncryptionConfiguration: serverConfig,
	}); err != nil {
		return err
	}

	if _, err := svc.PutBucketVersioning(&s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			Status: aws.String("Enabled"),
		},
	}); err != nil {
		return err
	}

	if _, err := svc.PutBucketAcl(&s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("private"),
	}); err != nil {
		return err
	}

	return nil
}
