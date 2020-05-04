package kms

import (
	"fmt"

	"github.com/dabblebox/stash/component/service/aws/policy"
)

var (
	writeActions = []string{
		"kms:UpdateKeyDescription",
		"kms:UpdateAlias",
		"kms:UntagResource",
		"kms:TagResource",
		"kms:ScheduleKeyDeletion",
		"kms:RevokeGrant",
		"kms:RetireGrant",
		"kms:ReEncryptTo",
		"kms:ReEncryptFrom",
		"kms:PutKeyPolicy",
		"kms:ImportKeyMaterial",
		"kms:GetParametersForImport",
		"kms:GetKeyRotationStatus",
		"kms:GenerateRandom",
		"kms:GenerateDataKeyWithoutPlaintext",
		"kms:GenerateDataKey",
		"kms:Encrypt",
		"kms:EnableKeyRotation",
		"kms:EnableKey",
		"kms:DisableKeyRotation",
		"kms:DisableKey",
		"kms:DeleteImportedKeyMaterial",
		"kms:DeleteAlias",
		"kms:CreateKey",
		"kms:CreateGrant",
		"kms:CreateAlias",
		"kms:CancelKeyDeletion",
	}

	readActions = []string{
		"kms:Decrypt",
	}
)

// Policy ...
func Policy(userIds, roleIds, auditArns []string, accountID string) policy.Policy {
	p := policy.Template("aws:userId", userIds, append(userIds, roleIds...), writeActions, readActions, "*")

	p.Statement = append(p.Statement,
		policy.Statement{
			Sid:    "Enable IAM User Permissions",
			Effect: "Allow",
			Principal: policy.Principal{
				AWS: fmt.Sprintf("arn:aws:iam::%s:root", accountID),
			},
			Action: []string{
				"kms:DescribeKey",
				"kms:GetKeyPolicy",
				"kms:list*",
			},
			Resource: "*",
		})

	return p
}
