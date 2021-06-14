package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/format"
	"github.com/dabblebox/stash/component/output"
	awskms "github.com/dabblebox/stash/component/service/aws/kms"
	"github.com/dabblebox/stash/component/service/aws/policy"
	"github.com/dabblebox/stash/component/service/aws/role"
	"github.com/dabblebox/stash/component/service/aws/sm"
	"github.com/dabblebox/stash/component/service/aws/terraform"
	"github.com/dabblebox/stash/component/service/aws/user"
)

const (
	KMSKeyIDOption = "kms_key_id"

	SMKMSKeyIDDefault = "aws/secretsmanager"

	SMSecretsDescription = "Managed by Stash"

	SMSecretsOption   = "secrets"
	SMSecretsDefault  = "single"
	SMSecretsSingle   = "single"
	SMSecretsMultiple = "multiple"

	SMDelimiterOption = "group_delimiter"
)

var SMSecretsOptions = []string{
	SMSecretsSingle,
	SMSecretsMultiple,
}

// SecretsManagerService ...
type SecretsManagerService struct {
	session *session.Session

	options map[string]string

	io IO
}

func (s *SecretsManagerService) ensureSession() error {
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
func (s *SecretsManagerService) Key() string {
	return "secrets-manager"
}

// ObjectKey ...
func (s *SecretsManagerService) ObjectKey(path string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9/_+=.@-]`)

	return re.ReplaceAllString(path, "-")
}

// SecurityRating ...
func (s *SecretsManagerService) SecurityRating() int {
	return SecurityRatingHigh
}

// Compatible ...
func (s *SecretsManagerService) Compatible(types []string) bool {
	compatible := map[string]bool{
		file.TypeEnv:        true,
		file.TypeJSON:       true,
		file.TypeJS:         true,
		file.TypeTypeScript: true,
		file.TypeXML:        true,
		file.TypeCert:       true,
		file.TypeSQL:        true,
		file.TypeYML:        true,
		file.TypeYAML:       true,
		file.TypeMissing:    true, // id_rsa private keys
	}

	for _, t := range types {
		if _, ok := compatible[t]; !ok {
			return false
		}
	}

	return true
}

// PreHook ...
func (s *SecretsManagerService) PreHook(io IO) error {
	s.io = io

	return nil
}

// Sync ...
func (s *SecretsManagerService) Sync(file File) (File, error) {

	if err := s.ensureSession(); err != nil {
		return File{}, err
	}

	svc := secretsmanager.New(s.session)

	if file.SupportsParsing() {
		if err := file.EnsureOption(Opt{
			Key:          SMSecretsOption,
			DefaultValue: SMSecretsDefault,
			Items:        SMSecretsOptions}, s.io); err != nil {
			return file, err
		}
	}

	if err := file.EnsureOption(Opt{
		Key:          KMSKeyIDOption,
		DefaultValue: SMKMSKeyIDDefault,
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

	secrets := map[string]string{}
	if len(file.Data) > 0 {
		s, err := toMap(&file)
		if err != nil {
			return file, fmt.Errorf("file invalid: %s", err)
		}
		secrets = s
	}

	deletedSecrets := []string{}
	newSecrets := map[string]secret{}
	modifiedSecrets := map[string]secret{}
	unsyncedSecrets := map[string]time.Time{}

	// Get remote secrets
	for key, value := range secrets {
		localSecret := newSecret(key, value, file.Options[KMSKeyIDOption])

		output, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(key),
			VersionStage: aws.String("AWSCURRENT"),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case secretsmanager.ErrCodeInvalidRequestException:
					deletedSecrets = append(deletedSecrets, key)
					modifiedSecrets[key] = localSecret
					continue
				case secretsmanager.ErrCodeResourceNotFoundException:
					newSecrets[key] = localSecret
					continue
				default:
					return file, err
				}
			}

			return file, err
		}

		o, err := svc.DescribeSecret(&secretsmanager.DescribeSecretInput{
			SecretId: aws.String(key),
		})
		if err != nil {
			return file, err
		}

		remoteSecret := toSecret(o, *output.SecretString)

		if localSecret != remoteSecret {
			modifiedSecrets[key] = localSecret

			if (*o.LastChangedDate).After(file.Synced) {
				unsyncedSecrets[key] = *o.LastChangedDate
			}
		}
	}

	if len(unsyncedSecrets) > 0 {
		overwrite := false
		prompt := &survey.Confirm{
			Message: "Remote data has changed since your last sync. Overwrite?",
			Help:    fmt.Sprintf("Modified Remote Keys%s", formatUnsynced(unsyncedSecrets)),
		}
		if err := survey.AskOne(prompt, &overwrite, survey.WithStdio(s.io.Stdin, s.io.Stdout, s.io.Stderr)); err != nil {
			return file, err
		}

		if !overwrite {
			return file, errors.New("user aborted sync")
		}
	}

	// Restore deleted secrets
	for _, key := range deletedSecrets {
		if _, err := svc.RestoreSecret(&secretsmanager.RestoreSecretInput{
			SecretId: aws.String(key),
		}); err != nil {
			return file, err
		}
	}

	// Delete removed secrets
	remoteKeys := make([]string, len(file.Keys))
	copy(remoteKeys, file.Keys)

	for _, remoteKey := range remoteKeys {
		if _, found := secrets[remoteKey]; !found {
			if _, err := svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
				SecretId: aws.String(remoteKey),
			}); err != nil {
				return file, err
			}

			file.RemoveKey(remoteKey)
		}
	}

	// Create new secrets
	for key, secret := range newSecrets {
		if _, err := svc.CreateSecret(&secretsmanager.CreateSecretInput{
			Name:         aws.String(secret.key),
			SecretString: aws.String(secret.value),
			Description:  aws.String(SMSecretsDescription),
			KmsKeyId:     aws.String(blankDefault(secret.keyID, SMKMSKeyIDDefault)),
		}); err != nil {
			return file, err
		}

		file.AddKey(key)
	}

	// Update modified secrets
	for key, secret := range modifiedSecrets {
		if _, err := svc.UpdateSecret(&secretsmanager.UpdateSecretInput{
			SecretId:     aws.String(secret.key),
			SecretString: aws.String(secret.value),
			KmsKeyId:     aws.String(blankDefault(secret.keyID, SMKMSKeyIDDefault)),
		}); err != nil {
			return file, err
		}

		file.AddKey(key)
	}

	return file, nil
}

func formatUnsynced(secrets map[string]time.Time) string {
	var builder strings.Builder

	for k, v := range secrets {
		builder.WriteString(fmt.Sprintf("\n  ∆ %s modified %s", k, v.Local().Format("3:04 1/2/2006 ")))
	}

	return builder.String()
}

type value struct {
	ARN   string
	Value string
}

func (v value) String() string {
	return v.Value
}

// Download ...
func (s *SecretsManagerService) Download(file File, format string) (File, error) {
	if err := s.ensureSession(); err != nil {
		return File{}, err
	}

	svc := secretsmanager.New(s.session)

	m := map[string]value{}
	for _, remoteKey := range file.Keys {
		o, err := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(remoteKey),
			VersionStage: aws.String("AWSCURRENT"),
		})
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeResourceNotFoundException:
				return file, fmt.Errorf("%s: %s", remoteKey, err)
			default:
				return file, fmt.Errorf("%s: %s", remoteKey, err)
			}
		}

		m[remoteKey] = value{
			ARN:   *o.ARN,
			Value: *o.SecretString,
		}
	}

	if format == output.TypeTerraform {
		dir := filepath.Dir(file.LocalPath)

		filePath := fmt.Sprintf("%s/kms.tf", dir)
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFKMSFile(filePath, blankDefault(file.Options[KMSKeyIDOption], SMKMSKeyIDDefault)); err != nil {
			return file, err
		}

		filePath = fmt.Sprintf("%s/%s-policy.tf", dir, s.Key())
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFSMPolicyFile(filePath); err != nil {
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

		d, err := s.terraform(m, file, s.io)
		if err != nil {
			return file, err
		}

		file.Data = d
	} else {
		d, err := toData(m, file, format)
		if err != nil {
			return file, err
		}

		file.Data = d
	}

	return file, nil
}

func (s SecretsManagerService) terraform(m map[string]value, file File, io IO) ([]byte, error) {

	var hcl bytes.Buffer
	w := bufio.NewWriter(&hcl)

	arnsMap := map[string]string{}
	arns := []string{}

	for _, value := range m {
		key := format.TerraformResourceName(filepath.Base(value.ARN))

		arnsMap[key] = value.ARN
		arns = append(arns, value.ARN)
	}

	p, err := json.MarshalIndent(policy.New(policy.Statement{
		Effect:   "Allow",
		Action:   []string{"secretsmanager:GetSecretValue"},
		Resource: &arns,
	}), "", "    ")
	if err != nil {
		return []byte{}, err
	}

	tmpl, err := template.New("role").Parse(role.HCLTemplate)
	if err != nil {
		return []byte{}, err
	}

	if err := tmpl.Execute(w, role.HCLModel{
		Name:   format.TerraformResourceName(file.RemoteKey),
		Policy: string(p),
	}); err != nil {
		return []byte{}, err
	}

	w.WriteString("\n\n")

	tmpl, err = template.New("sm").Funcs(template.FuncMap{"notLast": sm.NotLast}).Parse(sm.HCLSecretsTemplate)
	if err != nil {
		return []byte{}, err
	}

	if err := tmpl.Execute(w, sm.HCLModel{
		Arns: arnsMap,
	}); err != nil {
		return []byte{}, err
	}

	w.Flush()

	return hcl.Bytes(), nil
}

// Purge ...
func (s *SecretsManagerService) Purge(file File) error {

	if err := s.ensureSession(); err != nil {
		return err
	}

	svc := secretsmanager.New(s.session)

	remoteKeys := make([]string, len(file.Keys))
	copy(remoteKeys, file.Keys)

	for _, remoteKey := range remoteKeys {
		if _, err := svc.DeleteSecret(&secretsmanager.DeleteSecretInput{
			SecretId: aws.String(remoteKey),
		}); err != nil {
			return err
		}

		file.RemoveKey(remoteKey)
	}

	return nil
}

func init() {
	s := &SecretsManagerService{
		options: map[string]string{
			KMSKeyIDOption:  SMKMSKeyIDDefault,
			SMSecretsOption: SMSecretsDefault,
		},
	}

	Services[s.Key()] = s
}

func blankDefault(value, defaultValue string) string {
	if value == defaultValue {
		return ""
	}

	return value
}

func nilDefault(value, defaultValue string) *string {
	if value == defaultValue {
		return nil
	}

	return &value
}

func optionDefault(f File, option, defaultValue string) string {
	if v, ok := f.Options[option]; ok {
		return v
	}

	return defaultValue
}

func toSecret(d *secretsmanager.DescribeSecretOutput, value string) secret {
	s := secret{
		key:   *d.Name,
		keyID: SMKMSKeyIDDefault,
		value: value,
	}

	if d.KmsKeyId != nil {
		s.keyID = *d.KmsKeyId
	}

	return s
}

func newSecret(key, value, keyID string) secret {
	return secret{
		key:   key,
		value: value,
		keyID: keyID,
	}
}

type secret struct {
	key   string
	value string

	keyID string
}

func (s *secret) SetValue(v string) {
	s.value = v
}

func taskDefJsonTransformSecrets(m map[string]value) ([]byte, error) {
	type SecretFormat struct {
		Name      string `json:"name"`
		ValueFrom string `json:"valueFrom"`
	}

	env := []SecretFormat{}
	for key, value := range m {
		env = append(env, SecretFormat{
			ValueFrom: value.ARN,
			Name:      filepath.Base(key),
		})
	}

	return json.MarshalIndent(env, "", "    ")
}

func taskDefEnvTransformSecrets(m map[string]value) ([]byte, error) {
	var b bytes.Buffer

	for key, value := range m {
		if _, err := b.WriteString(fmt.Sprintf("%s=%s\n", filepath.Base(key), value.ARN)); err != nil {
			return b.Bytes(), err
		}
	}

	return b.Bytes(), nil
}

func toData(m map[string]value, f File, format string) ([]byte, error) {

	switch format {
	case output.TypeECSTaskInjectJson:
		return taskDefJsonTransformSecrets(m)
	case output.TypeECSTaskInjectEnv:
		return taskDefEnvTransformSecrets(m)
	}

	switch f.Type {
	case file.TypeEnv:
		return envToData(m, f.Options[SMDelimiterOption])
	case file.TypeJSON:
		return jsonToData(m, f.Options[SMSecretsOption])
	default:
		for _, value := range m {
			return []byte(value.String()), nil
		}
	}

	return []byte{}, nil
}

func envToData(m map[string]value, delimiter string) ([]byte, error) {
	results := bytes.Buffer{}
	for remoteKey, value := range m {

		temp := make(map[string]interface{})
		if err := json.Unmarshal([]byte(value.String()), &temp); err != nil {
			return []byte{}, err
		}

		for tk, tv := range temp {
			keySuffix := filepath.Base(remoteKey)

			prop := tk
			if len(delimiter) > 0 && strings.ToUpper(keySuffix) != tk {
				prop = strings.Trim(fmt.Sprintf("%s%s%s", strings.ToUpper(keySuffix), delimiter, tk), delimiter)
			}

			switch value := tv.(type) {
			case string:
				results.WriteString(fmt.Sprintf("%s=\"%s\"\n", prop, value))
			case int:
				results.WriteString(fmt.Sprintf("%s=%d\n", prop, value))
			default:
				results.WriteString(fmt.Sprintf("%s=%v\n", prop, value))
			}
		}
	}

	return results.Bytes(), nil
}

func jsonToData(m map[string]value, secrets string) ([]byte, error) {
	if secrets != SMSecretsMultiple {
		for _, value := range m {
			object := json.RawMessage{}
			if err := json.Unmarshal([]byte(value.String()), &object); err != nil {
				return []byte{}, err
			}

			return json.MarshalIndent(object, "", "   ")
		}
	}

	jsonObject := map[string]json.RawMessage{}

	for remoteKey, remoteValue := range m {

		secretProps := map[string]json.RawMessage{}

		if err := json.Unmarshal([]byte(remoteValue.String()), &secretProps); err == nil {

			for propName, propValue := range secretProps {
				remoteKeySuffix := filepath.Base(remoteKey)

				// single object prop
				if propName == remoteKeySuffix {
					jsonObject[remoteKeySuffix] = []byte(propValue)
					continue
				}

				// subsequent object props
				if raw, ok := jsonObject[remoteKeySuffix]; ok {

					childProps := map[string]json.RawMessage{}
					if err := json.Unmarshal(raw, &childProps); err != nil {
						return []byte{}, err
					}
					childProps[propName] = propValue

					b, err := json.Marshal(childProps)
					if err != nil {
						return []byte{}, err
					}

					jsonObject[remoteKeySuffix] = b
					continue
				}

				// initial object prop
				b, err := json.Marshal(map[string]json.RawMessage{propName: propValue})
				if err != nil {
					return []byte{}, err
				}

				jsonObject[remoteKeySuffix] = b
			}
		}
	}

	return json.MarshalIndent(jsonObject, "", "   ")
}

func toMap(f *File) (map[string]string, error) {

	switch f.Type {
	case file.TypeEnv:
		if v, ok := f.Options[SMSecretsOption]; !ok || v == SMSecretsSingle {
			props, err := f.parseENV()
			if err != nil {
				return map[string]string{}, err
			}

			jsonProps, err := json.Marshal(props)
			if err != nil {
				return map[string]string{}, err
			}

			return map[string]string{f.RemoteKey: string(jsonProps)}, nil
		}

		return envToMap(f, f.Options[SMDelimiterOption])
	case file.TypeJSON:
		if v, ok := f.Options[SMSecretsOption]; !ok || v == SMSecretsSingle {
			if len(f.Keys) == 1 {
				return map[string]string{f.Keys[0]: string(f.Data)}, nil
			}

			return map[string]string{f.RemoteKey: string(f.Data)}, nil
		}

		return jsonToMap(f)
	}

	if len(f.Keys) == 1 {
		return map[string]string{f.Keys[0]: string(f.Data)}, nil
	}

	return map[string]string{f.RemoteKey: string(f.Data)}, nil
}

func envToMap(f *File, delimiter string) (map[string]string, error) {
	props, err := f.parseENV()
	if err != nil {
		return map[string]string{}, err
	}

	group := make(map[string]map[string]interface{})
	for k, v := range props {

		keySuffix := k
		prop := k

		if len(delimiter) > 0 {
			parts := strings.Split(k, delimiter)

			keySuffix = parts[0]

			prop = parts[0]
			if len(parts) > 1 {
				prop = strings.Join(parts[1:], delimiter)
			}
		}

		var value interface{}
		if b, err := strconv.ParseBool(v); err == nil {
			value = b
		} else if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			value = json.Number(v)
		} else if _, err := strconv.ParseFloat(v, 64); err == nil {
			value = v
		} else {
			value = v
		}

		g, found := group[keySuffix]
		if !found {
			g = make(map[string]interface{})
		}

		g[prop] = value

		group[keySuffix] = g
	}

	secrets := map[string]string{}
	for keySuffix, childProps := range group {
		b, err := json.Marshal(childProps)
		if err != nil {
			return map[string]string{}, err
		}

		key, found := getKey(keySuffix, f.Keys)
		if !found {
			key = fmt.Sprintf("%s/%s", f.RemoteKey, keySuffix)
		}

		secrets[key] = string(b)
	}

	return secrets, nil
}

func jsonToMap(f *File) (map[string]string, error) {
	lm, err := f.parseJSON()
	if err != nil {
		return map[string]string{}, err
	}

	m := map[string]string{}
	for k, v := range lm {
		key, found := getKey(k, f.Keys)
		if found {
			m[key] = v
		} else {
			m[fmt.Sprintf("%s/%s", f.RemoteKey, k)] = v
		}
	}

	return m, nil
}

func getKey(name string, keys []string) (string, bool) {
	for _, k := range keys {
		if name == filepath.Base(k) {
			return k, true
		}
	}

	return "", false
}
