package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dabblebox/stash/component/format"
	"github.com/dabblebox/stash/component/output"
	awskms "github.com/dabblebox/stash/component/service/aws/kms"
	"github.com/dabblebox/stash/component/service/aws/policy"
	"github.com/dabblebox/stash/component/service/aws/role"
	awsS3 "github.com/dabblebox/stash/component/service/aws/s3"
	"github.com/dabblebox/stash/component/service/aws/terraform"
	"github.com/dabblebox/stash/component/service/aws/user"
)

const (
	S3BucketOption    = "s3_bucket"
	S3KMSKeyIDDefault = "aws/s3"

	S3RoleOption = "iam_role"

	arnsOption = "iam_arns"
	arnsDesc   = "Add an additional IAM arns of any users or roles that needs access."
)

// S3Service ...
type S3Service struct {
	session *session.Session

	io IO
}

func (s *S3Service) ensureSession() error {
	if s.session != nil {
		return nil
	}

	sess, err := session.NewSession(&aws.Config{})
	if err != nil {
		return err
	}

	s.session = sess

	return nil
}

// Key ...
func (s *S3Service) Key() string {
	return "s3"
}

// ObjectKey ...
func (s *S3Service) ObjectKey(path string) string {
	return path
}

// Compatible ...
func (s *S3Service) Compatible(types []string) bool {
	return true
}

// SecurityRating ...
func (s *S3Service) SecurityRating() int {
	return SecurityRatingLow
}

// PreHook ...
func (s *S3Service) PreHook(io IO) error {
	s.io = io

	return nil
}

// Sync ...
func (s *S3Service) Sync(file File) (File, error) {
	if len(file.Data) == 0 {
		return file, s.Purge(file)
	}

	if err := s.ensureSession(); err != nil {
		return file, err
	}

	if err := file.EnsureOption(Opt{Key: S3BucketOption}, s.io); err != nil {
		return file, err
	}

	if err := file.EnsureOption(Opt{
		Key:          KMSKeyIDOption,
		DefaultValue: S3KMSKeyIDDefault,
		Description:  awskms.Prompt}, s.io); err != nil {
		return file, err
	}

	keyID := file.Options[KMSKeyIDOption]
	if strings.Contains(keyID, "alias/") {

		user, err := user.Get(user.Dep{
			Session: s.session,
			Stdin:   s.io.Stdin,
			Stdout:  s.io.Stdout,
			Stderr:  s.io.Stderr,
		})
		if err != nil {
			return file, err
		}

		policy := awskms.Policy([]string{user.ID}, []string{}, []string{}, user.AccountID)

		k, err := awskms.CreateKey("Created by Stash", keyID, policy, map[string]string{}, kms.New(s.session))
		if err != nil {
			return file, err
		}

		file.Options[KMSKeyIDOption] = k
	}

	bucket := file.Options[S3BucketOption]
	keyID = file.Options[KMSKeyIDOption]

	svc := s3.New(s.session)

	o, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &file.RemoteKey,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == s3.ErrCodeNoSuchBucket {

				create := false
				prompt := &survey.Confirm{
					Help: `Create an S3 bucket locked down to specific users with default KMS encryption enabled. 
  An optional Terraform script to manage the resource will be created in the working directory.`,
					Message: fmt.Sprintf("%s does not exist. Create?", bucket),
				}
				if err := survey.AskOne(prompt, &create, survey.WithStdio(s.io.Stdin, s.io.Stdout, s.io.Stderr)); err != nil {
					return file, err
				}
				if create {

					user, err := user.Get(user.Dep{
						Session: s.session,
						Stdin:   s.io.Stdin,
						Stdout:  s.io.Stdout,
						Stderr:  s.io.Stderr,
					})
					if err != nil {
						return file, err
					}

					bucketPolicy := awsS3.Policy(bucket, user.AccountID, []string{user.Arn}, []string{})

					if err := awsS3.CreateBucket(bucket, bucketPolicy, map[string]string{}, svc); err != nil {
						return file, err
					}
				}
			}
		} else {
			return file, err
		}
	} else {
		defer o.Body.Close()

		if (*o.LastModified).After(file.Synced) {
			overwrite := false
			prompt := &survey.Confirm{
				Message: "Remote data has changed since your last sync. Overwrite?",
			}
			if err := survey.AskOne(prompt, &overwrite, survey.WithStdio(s.io.Stdin, s.io.Stdout, s.io.Stderr)); err != nil {
				return file, err
			}

			if !overwrite {
				return file, errors.New("user aborted sync")
			}
		}
	}

	uploader := s3manager.NewUploader(s.session)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:               &bucket,
		Key:                  &file.RemoteKey,
		Body:                 bytes.NewReader(file.Data),
		ServerSideEncryption: aws.String("aws:kms"),
		SSEKMSKeyId:          nilDefault(keyID, S3KMSKeyIDDefault),
	})

	file.Keys = []string{file.RemoteKey}

	return file, err
}

func parseTags(input string) map[string]string {

	tags := map[string]string{}

	tagLines := strings.Split(input, "\n")
	for _, l := range tagLines {
		if len(strings.TrimSpace(l)) > 0 {
			pair := strings.Split(l, "=")
			tags[pair[0]] = pair[1]
		}
	}

	return tags
}

// Download ...
func (s *S3Service) Download(file File, format string) (File, error) {

	if err := s.ensureSession(); err != nil {
		return file, err
	}

	bucket := file.Options[S3BucketOption]

	svc := s3.New(s.session)

	switch format {
	case output.TypeTerraform:

		dir := filepath.Dir(file.LocalPath)

		filePath := fmt.Sprintf("%s/kms.tf", dir)
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFKMSFile(filePath, blankDefault(file.Options[KMSKeyIDOption], S3KMSKeyIDDefault)); err != nil {
			return file, err
		}

		filePath = fmt.Sprintf("%s/%s.variables.tf", dir, s.Key())
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFVariablesFile(filePath); err != nil {
			return file, err
		}

		filePath = fmt.Sprintf("%s/%s.auto.tfvars", dir, s.Key())
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFVarsFile(filePath, terraform.Dep{
			Session: s.session,
			Stdin:   s.io.Stdin,
			Stdout:  s.io.Stdout,
			Stderr:  s.io.Stderr,
		}); err != nil {
			return file, err
		}

		d, err := s.terraform(file, bucket, s.io)
		file.Data = d
		return file, err
	case output.TypeECSTaskInjectJson:
		type EnvFileFormat struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		}

		envFile := []EnvFileFormat{{
			Value: fmt.Sprintf("arn:aws:s3:::%s/%s", bucket, file.RemoteKey),
			Type:  "s3",
		}}

		b, err := json.MarshalIndent(envFile, "", "    ")

		file.Data = b
		return file, err
	case output.TypeECSTaskInjectEnv:

		var b bytes.Buffer

		_, err := b.WriteString(fmt.Sprintf("s3=arn:aws:s3:::%s/%s", bucket, file.RemoteKey))

		file.Data = b.Bytes()
		return file, err
	}

	o, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &file.RemoteKey,
	})
	if err != nil {
		return file, err
	}

	defer o.Body.Close()
	b, err := ioutil.ReadAll(o.Body)
	if err != nil {
		return file, err
	}

	file.Data = b

	return file, nil
}

func (s S3Service) terraform(file File, bucket string, io IO) ([]byte, error) {

	var hcl bytes.Buffer
	w := bufio.NewWriter(&hcl)

	p, err := json.MarshalIndent(policy.New(policy.Statement{
		Effect:   "Allow",
		Action:   []string{"s3:Get*"},
		Resource: []string{fmt.Sprintf("arn:aws:s3:::%s/%s", file.Options[S3BucketOption], file.RemoteKey)},
	}), "", "    ")
	if err != nil {
		return []byte{}, err
	}

	tmpl, err := template.New("role").Parse(role.HCLTemplate)
	if err != nil {
		return []byte{}, err
	}

	if err := tmpl.Execute(w, role.HCLModel{
		Name:   format.TerraformResourceName(filepath.Base(file.RemoteKey)),
		Policy: string(p),
	}); err != nil {
		return []byte{}, err
	}

	if _, err := w.WriteString("\n\n"); err != nil {
		return []byte{}, err
	}

	tmpl, err = template.New("s3").Parse(awsS3.HCLTemplate)
	if err != nil {
		return []byte{}, err
	}

	if err := tmpl.Execute(w, awsS3.HCLModel{
		Bucket: bucket,
	}); err != nil {
		return []byte{}, err
	}

	w.Flush()

	return hcl.Bytes(), nil
}

// Purge ...
func (s *S3Service) Purge(file File) error {
	if err := s.ensureSession(); err != nil {
		return err
	}

	bucket := file.Options[S3BucketOption]

	svc := s3.New(s.session)

	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &file.RemoteKey,
	})

	return err
}

func init() {
	s := new(S3Service)

	Services[s.Key()] = s
}
