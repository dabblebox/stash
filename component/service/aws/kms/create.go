package kms

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/dabblebox/stash/component/service/aws/policy"
)

func CreateKey(desc, alias string, policy policy.Policy, tags map[string]string, svc *kms.KMS) (string, error) {

	awsTags := []*kms.Tag{}
	for k, v := range tags {
		awsTags = append(awsTags, &kms.Tag{
			TagKey:   aws.String(k),
			TagValue: aws.String(v),
		})
	}

	p, err := json.MarshalIndent(policy, "", "   ")
	if err != nil {
		return "", err
	}

	k, err := svc.CreateKey(&kms.CreateKeyInput{
		Description: aws.String(desc),
		Tags:        awsTags,
		Policy:      aws.String(string(p)),
	})
	if err != nil {
		return "", err
	}

	if _, err := svc.CreateAlias(&kms.CreateAliasInput{
		AliasName:   aws.String(alias),
		TargetKeyId: k.KeyMetadata.KeyId,
	}); err != nil {
		return "", err
	}

	return *k.KeyMetadata.KeyId, nil
}
