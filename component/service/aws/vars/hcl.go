package vars

type HCLVarsModel struct {
	Users []string
	Roles []string
}

var HCLVarsTemplate = `
# List the role users that require access to the configuration. It is CRITICAL that at least one valid
# role user be listed here to prevent being completely locked out of AWS resources. Email addresses are 
# case sensitive and role users should be in the format: role-name/First.Last@domain.com
configUsers = [{{range .Users}}
  "{{ . }}"
{{end}}]

# List the role names that require decrypt access to the configuration. This is useful when giving a 
# Fargate task or Lambda function access to configuration. If using a locked down S3 bucket to store 
# configs, the roles will also need to be added to S3 bucket policy.
configRoles = [{{range .Roles}}
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

variable "configUsers" { 
  default = [] 
}

variable "configRoles" { 
  default = [] 
}

locals {
  configUserIds = flatten([
    for user in var.configUsers: [
       format("%s:%s", data.aws_iam_role.user_role[split("/", user)[0]].unique_id, split("/", user)[1])
    ]
  ])

  configUserArns = flatten([
    for user in var.configUsers: 
        format("arn:aws:sts::%s:assumed-role/%s", data.aws_caller_identity.current_config_user.account_id, user)
  ])

  configRoleArns = flatten([
    for role in var.configRoles:
       format("arn:aws:sts::%s:assumed-role/%s/*", data.aws_caller_identity.current_config_user.account_id, role)
  ])

  configRoleIds = flatten([
    for instance in data.aws_iam_role.role:
      format("%s:*",  instance.unique_id) 
  ])
}

data "aws_iam_role" "role" {
  for_each = toset(var.configRoles)
  
  name = each.value
}

data "aws_iam_role" "user_role" {
  for_each = toset(distinct(flatten([
    for instance in var.configUsers:
      split("/", instance)[0]
  ])))
  
  name = each.value
}

data "aws_caller_identity" "current_config_user" {}`
