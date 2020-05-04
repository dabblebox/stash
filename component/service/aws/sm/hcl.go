package sm

type HCLModel struct {
	Arns map[string]string
}

var HCLTemplate = `# By default, secrets can be altered by other users with power user access. Use the 
# import commands to take control of each secret's policy. After applying the Terraform, 
# the secrets will be locked down to only specified users and roles and encrypted with
# a KMS key that is locked down also. 
#{{range $i, $a := .Arns }}
# $ terraform import aws_secretsmanager_secret.secret_{{ $i }} {{ $a }}{{end}}
{{range $i, $a := .Arns }}
resource "aws_secretsmanager_secret" "secret_{{ $i }}" {
	description = "Created by Stash"
	
	kms_key_id = aws_kms_key.config.key_id
	
	tags = var.tags
	
	policy = data.template_file.sm_policy.rendered
}
{{end}}

data "template_file" "sm_policy" {
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
				"secretsmanager:CancelRotateSecret",
				"secretsmanager:CreateSecret",
				"secretsmanager:DeleteSecret",
				"secretsmanager:PutSecretValue",
				"secretsmanager:RestoreSecret",
				"secretsmanager:RotateSecret",
				"secretsmanager:TagResource",
				"secretsmanager:UntagResource",
				"secretsmanager:UpdateSecret",
				"secretsmanager:UpdateSecretVersionStage"
			],
			"Resource": [
			{{$notLast := notLast .Arns}}{{range $i, $e := .Arns}}   "{{ $e }}"{{if call $notLast}},{{end}}
			{{end}}],
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
				"secretsmanager:DescribeSecret",
				"secretsmanager:List*",
				"secretsmanager:GetRandomPassword",
				"secretsmanager:GetSecretValue"
			],
			"Resource": [
			{{$notLast := notLast .Arns}}{{range $i, $e := .Arns}}   "{{ $e }}"{{if call $notLast}},{{end}}
			{{end}}],
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
				"secretsmanager:CancelRotateSecret",
				"secretsmanager:CreateSecret",
				"secretsmanager:DeleteSecret",
				"secretsmanager:PutSecretValue",
				"secretsmanager:RestoreSecret",
				"secretsmanager:RotateSecret",
				"secretsmanager:TagResource",
				"secretsmanager:UntagResource",
				"secretsmanager:UpdateSecret",
				"secretsmanager:UpdateSecretVersionStage"
			],
			"Resource": [
			{{$notLast := notLast .Arns}}{{range $i, $e := .Arns}}   "{{ $e }}"{{if call $notLast}},{{end}}
			{{end}}],
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
				"secretsmanager:DescribeSecret",
				"secretsmanager:List*",
				"secretsmanager:GetRandomPassword",
				"secretsmanager:GetSecretValue"
			],
			"Resource": [
			{{$notLast := notLast .Arns}}{{range $i, $e := .Arns}}   "{{ $e }}"{{if call $notLast}},{{end}}
			{{end}}],
			"Condition": {
				"StringLike": {
					"aws:userId": $${readPrincipals}
				}
			}
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
}`

func NotLast(m map[string]string) func() bool {
	i := 0

	return func() bool {
		i++
		return (i != len(m))
	}
}
