package role

type HCLModel struct {
	Name string

	Policy string
}

var HCLTemplate = `
# This Terraform example of a role policy grants access to the specified
# configuration.
#
# When using a Task Definition to inject configuration into a container, the 
# policy should be attached to a Task Definition execution role, and when using 
# an app inside a container or a lambda function to get config, the policy 
# should be attached to the app or lambda role.
resource "aws_iam_role_policy_attachment" "role_attachment_{{ .Name }}" {
	for_each = toset(var.configRoles)
	
	role       = each.value
	policy_arn = aws_iam_policy.role_policy_{{ .Name }}.arn
}

resource "aws_iam_policy" "role_policy_{{ .Name }}" {
	name   = "{{ .Name }}"
	policy = <<EOF
{{ .Policy }}
EOF
}`
