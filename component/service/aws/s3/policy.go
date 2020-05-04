package s3

import (
	"fmt"

	"github.com/dabblebox/stash/component/service/aws/policy"
)

var (
	writeActions = []string{
		"s3:Delete*",
		"s3:Put*",
		"s3:Replicate*",
		"s3:Restore*",
	}

	readActions = []string{
		"s3:GetObject",
		"s3:List*",
	}
)

// Policy ...
func Policy(bucket, accountID string, userArns, roleArns []string) policy.Policy {

	p := policy.Template("aws:arn", userArns, append(userArns, roleArns...), writeActions, readActions, []string{
		fmt.Sprintf("arn:aws:s3:::%s", bucket),
		fmt.Sprintf("arn:aws:s3:::%s/*", bucket),
	})

	p.Statement = append(p.Statement, policy.Statement{
		Effect:    "Allow",
		Principal: "*",
		Action:    []string{"s3:GetBucketTagging", "s3:GetBucketAcl"},
		Resource: []string{
			fmt.Sprintf("arn:aws:s3:::%s", bucket),
			fmt.Sprintf("arn:aws:s3:::%s/*", bucket),
		},
		Condition: &policy.Condition{
			StringLike: map[string]interface{}{
				"aws:arn": []string{fmt.Sprintf("arn:aws:sts::%s:*", accountID)},
			},
		},
	})

	return p
}
