package terraform

import (
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/dabblebox/stash/component/service/aws/user"
	"github.com/dabblebox/stash/component/service/aws/vars"
)

type Dep struct {
	Session *session.Session

	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

func EnsureTFVarsFile(filePath string, dep Dep) error {
	if _, err := os.Stat(filePath); err != nil {

		user, err := user.Get(user.Dep{
			Session: dep.Session,
			Stdin:   dep.Stdin,
			Stdout:  dep.Stdout,
			Stderr:  dep.Stderr,
		})
		if err != nil {
			return err
		}

		return CreateTFVarsFile(filePath, user.Name)
	}

	return nil
}

func CreateTFVarsFile(filePath, userName string) error {

	tmpl, err := template.New("vars").Parse(vars.HCLVarsTemplate)
	if err != nil {
		return err
	}

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	m := vars.HCLVarsModel{}

	const userPrefix = "user/"

	if strings.Contains(userName, userPrefix) {
		m.IAMUsers = []string{strings.TrimPrefix(userName, userPrefix)}
	} else {
		m.SAMLUsers = []string{userName}
	}

	if err := tmpl.Execute(f, m); err != nil {
		return err
	}

	return nil
}
