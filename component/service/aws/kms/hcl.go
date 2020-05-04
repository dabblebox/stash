package kms

type HCLModel struct {
	KMSKeyID string
}

var HCLTemplate = `
# Manage the KMS key through Terraform.
{{if .KMSKeyID }}# 
# $ terraform import aws_kms_key.config {{ .KMSKeyID }}{{end}}
resource "aws_kms_key" "config" {
  deletion_window_in_days = 7
  
  enable_key_rotation     = true
  
  tags = var.tags

  policy = data.template_file.config_policy.rendered
}

data "template_file" "config_policy" {
  template = <<EOF
  {
      "Version": "2012-10-17",
      "Statement": [
          {
              "Sid": "DenyWriteToAllExceptSAMLUsers",
              "Effect": "Deny",
              "Principal": {
                  "AWS": "*"
              },
              "Action": [
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
                  "kms:EnableKeyRotation",
                  "kms:EnableKey",
                  "kms:Encrypt",
                  "kms:DisableKeyRotation",
                  "kms:DisableKey",
                  "kms:DeleteImportedKeyMaterial",
                  "kms:DeleteAlias",
                  "kms:CreateKey",
                  "kms:CreateGrant",
                  "kms:CreateAlias",
                  "kms:CancelKeyDeletion"
              ],
              "Resource": "*",
              "Condition": {
                  "StringNotLike": {
                      "aws:userId": $${writePrincipals}
                  }
              }
          },
          {
              "Sid": "DenyReadToAllExceptRoleAndSAMLUsers",
              "Effect": "Deny",
              "Principal": {
                  "AWS": "*"
              },
              "Action": [
                  "kms:Decrypt"
              ],
              "Resource": "*",
              "Condition": {
                  "StringNotLike": {
                      "aws:userId": $${readPrincipals}
                  }
              }
          },
          {
              "Sid": "AllowWriteToSAMLUsers",
              "Effect": "Allow",
              "Principal": {
                  "AWS": "*"
              },
              "Action": [
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
                  "kms:EnableKeyRotation",
                  "kms:EnableKey",
                  "kms:Encrypt",
                  "kms:DisableKeyRotation",
                  "kms:DisableKey",
                  "kms:DeleteImportedKeyMaterial",
                  "kms:DeleteAlias",
                  "kms:CreateKey",
                  "kms:CreateGrant",
                  "kms:CreateAlias",
                  "kms:CancelKeyDeletion"
              ],
              "Resource": "*",
              "Condition": {
                  "StringLike": {
                      "aws:userId": $${writePrincipals}
                  }
              }
          },
          {
              "Sid": "AllowReadRoleAndSAMLUsers",
              "Effect": "Allow",
              "Principal": {
                  "AWS": "*"
              },
              "Action": [
                  "kms:Decrypt"
              ],
              "Resource": "*",
              "Condition": {
                  "StringLike": {
                      "aws:userId": $${readPrincipals}
                  }
              }
          },
          {
            "Sid": "Enable IAM User Permissions",
            "Effect": "Allow",
            "Principal": {
                "AWS": "arn:aws:iam::${data.aws_caller_identity.current_config_user.account_id}:root"
            },
            "Action": [
                "kms:DescribeKey",
                "kms:GetKeyPolicy",
                "kms:List*"
            ],
            "Resource": "*"
          }
      ]
  }
  EOF

  vars = {
    writePrincipals = jsonencode(local.configUserIds)

    readPrincipals = jsonencode(
      concat(
        local.configUserIds,
        local.configRoleIds,
      ),
    )
  }
}

output "kms_key_id" {
  value = aws_kms_key.config.key_id
}`
