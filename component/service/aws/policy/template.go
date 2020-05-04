package policy

// Template ...
func Template(conditionType string, writeUsers, readUsers interface{}, write, read []string, resources interface{}) Policy {

	denyWrite := Statement{
		Sid:    "DenyWriteToAllExceptSAMLUsers",
		Effect: "Deny",
		Principal: Principal{
			AWS: "*",
		},
		Action:   write,
		Resource: resources,
		Condition: &Condition{
			StringNotLike: map[string]interface{}{
				conditionType: writeUsers,
			},
		},
	}

	denyRead := Statement{
		Sid:    "DenyReadToAllExceptRoleAndSAMLUsers",
		Effect: "Deny",
		Principal: Principal{
			AWS: "*",
		},
		Action:   read,
		Resource: resources,
		Condition: &Condition{
			StringNotLike: map[string]interface{}{
				conditionType: readUsers,
			},
		},
	}

	allowWrite := Statement{
		Sid:    "AllowWriteToSAMLUsers",
		Effect: "Allow",
		Principal: Principal{
			AWS: "*",
		},
		Action:   write,
		Resource: resources,
		Condition: &Condition{
			StringLike: map[string]interface{}{
				conditionType: writeUsers,
			},
		},
	}

	allowRead := Statement{
		Sid:    "AllowReadRoleAndSAMLUsers",
		Effect: "Allow",
		Principal: Principal{
			AWS: "*",
		},
		Action:   read,
		Resource: resources,
		Condition: &Condition{
			StringLike: map[string]interface{}{
				conditionType: readUsers,
			},
		},
	}

	return New(denyWrite, denyRead, allowWrite, allowRead)
}
