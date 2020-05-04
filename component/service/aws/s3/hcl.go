package s3

type HCLModel struct {
	Bucket string

	Tags map[string]string

	Policy string
}

var HCLTemplate = `# This Terraform is an example of what a secure S3 bucket and policy might look like to 
# allow access to configuration for specific users and roles. The policy's users and 
# roles in the JSON below should be modified to provide appropriate access. This file 
# can be deleted when using the AWS console to manage the S3 bucket instead of Terraform.
#
# Stash does NOT manage infrastructure. When finished with this S3 bucket, please destroy
# through Terraform or in the AWS console.
#
# To manage the S3 bucket through Terraform run the following commands.
#
# $ terraform import aws_s3_bucket.config {{ .Bucket }}
# $ terraform import aws_s3_bucket_public_access_block.config_public_bucket_policy {{ .Bucket }}

resource "aws_s3_bucket" "config" {
  bucket        = "{{ .Bucket }}"

  # These two properties may show changes are required during Terraform plam,
  # even when they are configured to the same values at the bucket.
  force_destroy = true
  acl           = "private"

  tags = var.tags

  versioning {
    enabled = true
    mfa_delete = false
  }

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        sse_algorithm     = "aws:kms"
      }
    }
  }
}

resource "aws_s3_bucket_policy" "config" {
  bucket = aws_s3_bucket.config.id

  policy = data.template_file.s3_policy.rendered
}

data "template_file" "s3_policy" {
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
            "s3:Delete*",
            "s3:Put*",
            "s3:Replicate*",
            "s3:Restore*"
         ],
         "Resource": [
            "arn:aws:s3:::{{ .Bucket }}",
            "arn:aws:s3:::{{ .Bucket }}/*"
         ],
         "Condition": {
            "StringNotLike": {
              "aws:arn": $${writePrincipals}
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
            "s3:GetObject",
            "s3:List*"
         ],
         "Resource": [
            "arn:aws:s3:::{{ .Bucket }}",
            "arn:aws:s3:::{{ .Bucket }}/*"
         ],
         "Condition": {
            "StringNotLike": {
              "aws:arn": $${readPrincipals}
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
            "s3:Delete*",
            "s3:Put*",
            "s3:Replicate*",
            "s3:Restore*"
         ],
         "Resource": [
            "arn:aws:s3:::{{ .Bucket }}",
            "arn:aws:s3:::{{ .Bucket }}/*"
         ],
         "Condition": {
            "StringLike": {
              "aws:arn": $${writePrincipals}
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
            "s3:GetObject",
            "s3:List*"
         ],
         "Resource": [
            "arn:aws:s3:::{{ .Bucket }}",
            "arn:aws:s3:::{{ .Bucket }}/*"
         ],
         "Condition": {
            "StringLike": {
               "aws:arn": $${readPrincipals}
            }
         }
      },
      {
         "Effect": "Allow",
         "Principal": "*",
         "Action": [
            "s3:GetBucketTagging",
            "s3:GetBucketAcl"
         ],
         "Resource": [
            "arn:aws:s3:::{{ .Bucket }}",
            "arn:aws:s3:::{{ .Bucket }}/*"
         ],
         "Condition": {
            "StringLike": {
               "aws:arn": "arn:aws:sts::${data.aws_caller_identity.current_config_user.account_id}:*"
            }
         }
      }
   ]
}
EOF

  vars = {
    writePrincipals = jsonencode(local.configUserArns)

    readPrincipals = jsonencode(
      concat(
        local.configUserArns,
        local.configRoleArns,
      ),
    )
  }
}

resource "aws_s3_bucket_public_access_block" "config_public_bucket_policy" {
  bucket = aws_s3_bucket.config.id

  block_public_acls   = true
  block_public_policy = true
  ignore_public_acls = true
  restrict_public_buckets = true
}`
