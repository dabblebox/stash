package user

import (
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var user User

type Dep struct {
	Session *session.Session

	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

type User struct {
	Name      string
	Arn       string
	ID        string
	AccountID string
}

func Get(dep Dep) (User, error) {

	if user == (User{}) {
		u, err := sts.New(dep.Session).GetCallerIdentity(&sts.GetCallerIdentityInput{})
		if err != nil {
			return User{}, err
		}

		a, err := arn.Parse(*u.Arn)
		if err != nil {
			return User{}, err
		}

		user = User{
			Name:      strings.TrimPrefix(a.Resource, "assumed-role/"),
			Arn:       *u.Arn,
			ID:        *u.UserId,
			AccountID: *u.Account,
		}
	}

	return user, nil
}
