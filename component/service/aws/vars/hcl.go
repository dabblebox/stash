package vars

type HCLVarsModel struct {
	SAMLUsers []string
	IAMUsers  []string
	IAMRoles  []string
}

var HCLVarsTemplate = `
# PLEASE READ: List the SAML users that require full access to the configuration. It is CRITICAL that
# at least one valid user be listed here to prevent being completely locked out of AWS resources. Email 
# addresses are case sensitive and role users should be in the format: role-name/first.last@domain.com
configSAMLUsers = [{{range .SAMLUsers}}
  "{{ . }}"
{{end}}]

# List the IAM user names that require decrypt access to the configuration. This is useful when giving
# a server access to configuration when the server does not live in the AWS cloud and cannot use roles.
configIAMUsers = [{{range .IAMUsers}}
  "{{ . }}"
{{end}}]

# List the role names that require decrypt access to the configuration. This is useful when giving a 
# Fargate task or Lambda function access to configuration. If using a locked down S3 bucket to store 
# configs, the roles will also need to be added to S3 bucket policy.
configIAMRoles = [{{range .IAMRoles}}
  "{{ . }}"
{{end}}]`

var HCLVariablesTemplate = `
# This file gets and formats user and role information used by the KMS key and
# configuration storage services.

# This tags declaration can be removed when done elsewhere.
variable "tags" {
  type = map(string)
  default = {}
}

variable "configSAMLUsers" { 
  default = [] 
}

variable "configIAMUsers" { 
  default = [] 
}

variable "configIAMRoles" { 
  default = [] 
}

locals {
  configUserIds = flatten([
    for user in var.configSAMLUsers: [
       format("%s:%s", data.aws_iam_role.user_role[split("/", user)[0]].unique_id, split("/", user)[1])
    ]
  ])

  configUserArns = flatten([
    for user in var.configSAMLUsers: 
        format("arn:aws:sts::%s:assumed-role/%s", data.aws_caller_identity.current_config_user.account_id, user)
  ])

  configRoleArns = flatten([[
    for role in var.configIAMRoles:
       format("arn:aws:sts::%s:assumed-role/%s/*", data.aws_caller_identity.current_config_user.account_id, role)
  ],[
    for user in var.configIAMUsers:
       format("arn:aws:sts::%s:user/%s", data.aws_caller_identity.current_config_user.account_id, user)
  ]])

  configRoleIds = flatten([[
    for instance in data.aws_iam_role.role:
      format("%s:*",  instance.unique_id) 
  ],[
    for instance in data.aws_iam_user.user_name:
       format("%s",  instance.user_id) 
  ]])
}

data "aws_iam_role" "role" {
  for_each = toset(var.configIAMRoles)
  
  name = each.value
}

data "aws_iam_user" "user_name" {
  for_each = toset(var.configIAMUsers)
  
  user_name = each.value
}

data "aws_iam_role" "user_role" {
  for_each = toset(distinct(flatten([
    for instance in var.configSAMLUsers:
      split("/", instance)[0]
  ])))
  
  name = each.value
}

data "aws_caller_identity" "current_config_user" {}`
