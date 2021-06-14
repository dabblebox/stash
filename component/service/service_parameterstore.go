package service

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/format"
	"github.com/dabblebox/stash/component/output"
	awskms "github.com/dabblebox/stash/component/service/aws/kms"
	"github.com/dabblebox/stash/component/service/aws/policy"
	"github.com/dabblebox/stash/component/service/aws/role"
	"github.com/dabblebox/stash/component/service/aws/terraform"
	"github.com/dabblebox/stash/component/service/aws/user"
	"github.com/dabblebox/stash/component/slice"
)

const (
	PSKMSKeyIDDefault = "aws/ssm"
)

// ParameterStoreService ...
type ParameterStoreService struct {
	session *session.Session

	io IO
}

func (s *ParameterStoreService) ensureSession() error {
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
func (s ParameterStoreService) Key() string {
	return "parameter-store"
}

// ObjectKey ...
func (s ParameterStoreService) ObjectKey(path string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9.-_/]`)

	return fmt.Sprintf("/%s", re.ReplaceAllString(path, "-"))
}

// SecurityRating ...
func (s ParameterStoreService) SecurityRating() int {
	return SecurityRatingMedium
}

// Compatible ...
func (s ParameterStoreService) Compatible(types []string) bool {
	compatible := map[string]bool{
		file.TypeEnv: true,
	}

	for _, t := range types {
		if _, ok := compatible[t]; !ok {
			return false
		}
	}

	return true
}

// PreHook ...
func (s *ParameterStoreService) PreHook(io IO) error {
	s.io = io

	return nil
}

// Sync ...
func (s ParameterStoreService) Sync(file File) (File, error) {
	params := map[string]string{}

	if len(file.Data) > 0 {
		p, err := dotenv.Parse(bytes.NewReader(file.Data))
		if err != nil {
			return file, fmt.Errorf("file invalid: %s", err)
		}
		params = p
	}

	if err := s.ensureSession(); err != nil {
		return file, err
	}

	if err := file.EnsureOption(Opt{
		Key:          KMSKeyIDOption,
		DefaultValue: PSKMSKeyIDDefault,
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

	svc := ssm.New(s.session)

	remoteParams := map[string]param{}

	for remoteKeyPath, trackedProps := range getKeyPaths(file) {

		rps, err := getRemoteParamsWithMetaData(remoteKeyPath, trackedProps, svc)
		if err != nil {
			return file, err
		}

		for k, v := range rps {
			remoteParams[k] = v
		}
	}

	modifiedParams := []param{}
	unsyncedParams := map[string]time.Time{}

	for name, value := range params {

		remoteKey, found := file.LookupRemoteKey(name)
		if !found {
			remoteKey = fmt.Sprintf("%s/%s", file.RemoteKey, name)
		}

		param := toParam(
			remoteKey,
			value,
			file.Options[KMSKeyIDOption])

		if remoteParam, ok := remoteParams[param.name]; ok {
			if param.Changed(remoteParam) {
				modifiedParams = append(modifiedParams, param)

				if remoteParam.lastModified.After(file.Synced) {
					unsyncedParams[remoteParam.name] = remoteParam.lastModified
				}
			}
		} else {
			modifiedParams = append(modifiedParams, param)
		}
	}

	if len(unsyncedParams) > 0 {
		overwrite := false
		prompt := &survey.Confirm{
			Message: "Remote data has changed since your last sync. Overwrite?",
			Help:    fmt.Sprintf("Modified Remote Keys%s", formatUnsynced(unsyncedParams)),
		}
		if err := survey.AskOne(prompt, &overwrite, survey.WithStdio(s.io.Stdin, s.io.Stdout, s.io.Stderr)); err != nil {
			return file, err
		}

		if !overwrite {
			return file, errors.New("user aborted sync")
		}
	}

	//------------------------------------------
	//- Update Modified Parameters
	//------------------------------------------
	for _, param := range modifiedParams {
		_, err := svc.PutParameter(&ssm.PutParameterInput{
			Name:      &param.name,
			Value:     &param.value,
			Overwrite: aws.Bool(true),
			Type:      aws.String(ssm.ParameterTypeSecureString),
			KeyId:     nilDefault(param.keyID, PSKMSKeyIDDefault),
		})
		if err != nil {
			return file, fmt.Errorf("%s: %s", param.name, err)
		}

		file.AddKey(param.name)
	}

	//------------------------------------------
	//- Delete Removed Parameters
	//------------------------------------------
	for _, remoteParam := range remoteParams {
		if _, found := params[filepath.Base(remoteParam.name)]; !found {
			if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
				Name: aws.String(remoteParam.name),
			}); err != nil {
				return file, err
			}

			file.RemoveKey(remoteParam.name)
		}
	}

	return file, nil
}

func toParam(name, value, keyID string) param {
	return param{
		name:  name,
		value: value,
		keyID: keyID,
		pType: ssm.ParameterTypeSecureString,
	}
}

func getRemoteParamsWithMetaData(remoteKey string, props []string, svc *ssm.SSM) (map[string]param, error) {

	parameters := map[string]param{}

	stashedParams, err := getRemoteParams(remoteKey, svc)
	if err != nil {
		return parameters, err
	}

	for _, sp := range stashedParams {
		if slice.In(filepath.Base(sp.name), props) {
			history, err := getParamHistory(svc, &sp.name, "", []*ssm.ParameterHistory{})
			if err != nil {
				return parameters, err
			}

			sp.keyID = findKMSKeyID(history)

			parameters[sp.name] = sp
		}
	}

	return parameters, nil
}

func findKMSKeyID(history []*ssm.ParameterHistory) (keyID string) {

	var maxVersion int64 = 0

	for _, ph := range history {

		if ph.Version != nil && ph.KeyId != nil && (*ph.Version) > maxVersion {
			maxVersion = *ph.Version
			keyID = strings.Replace(*ph.KeyId, "alias/", "", 1)
		}
	}

	return keyID
}

// The AWS api call, GetParameterHistory, does NOT make a call for every individual page in Parameter Store like the call, DescribeParameters, does.
// This is a performance work around suggested by the AWS support staff. As the number of Parameter Store parameters increases, the number of calls
// to DescribeParameters also increases making scailing difficult due to rate limits. GetParameterHistory usually takes one or two calls to return
// additional parameter data.
func getParamHistory(svc *ssm.SSM, name *string, nextToken string, params []*ssm.ParameterHistory) ([]*ssm.ParameterHistory, error) {

	input := &ssm.GetParameterHistoryInput{
		Name:       name,
		MaxResults: aws.Int64(50),
	}

	if len(nextToken) > 0 {
		input.SetNextToken(nextToken)
	}

	output, err := svc.GetParameterHistory(input)
	if err != nil {
		return nil, err
	}

	params = append(params, output.Parameters...)

	if output.NextToken == nil || len(*output.NextToken) == 0 {
		return params, nil
	}

	return getParamHistory(svc, name, *output.NextToken, params)
}

// The AWS api call for GetParametersByPath has higher rate limits than DescribeParameters. Using this function is more resiliant
// and should be used where possible. However, comparing KMS key ids requires using DescribeParameters; so, in a few cases, the
// call to describeParams is necesary.
func getRemoteParamsByPath(svc *ssm.SSM, startsWith string, nextToken string, params []*ssm.Parameter) ([]*ssm.Parameter, error) {

	input := &ssm.GetParametersByPathInput{
		Recursive:      aws.Bool(true),
		Path:           aws.String(startsWith),
		WithDecryption: aws.Bool(true),
		MaxResults:     aws.Int64(10),
	}

	if len(nextToken) > 0 {
		input.SetNextToken(nextToken)
	}

	output, err := svc.GetParametersByPath(input)
	if err != nil {
		return nil, err
	}

	params = append(params, output.Parameters...)

	if output.NextToken == nil || len(*output.NextToken) == 0 {
		return params, nil
	}

	return getRemoteParamsByPath(svc, startsWith, *output.NextToken, params)
}

func getKeyPaths(f File) map[string][]string {
	paths := map[string][]string{f.RemoteKey: []string{}}

	for _, k := range f.Keys {
		path := filepath.Dir(k)
		prop := filepath.Base(k)

		if props, found := paths[path]; found {
			props = append(props, prop)
			paths[path] = props
		} else {
			paths[path] = []string{prop}
		}
	}

	return paths
}

// Download ...
func (s ParameterStoreService) Download(file File, format string) (File, error) {
	if err := s.ensureSession(); err != nil {
		return file, err
	}

	svc := ssm.New(s.session)

	remoteParams := []param{}

	for remoteKeyPath, trackedProps := range getKeyPaths(file) {

		rps, err := getRemoteParams(remoteKeyPath, svc)
		if err != nil {
			return file, err
		}

		for _, rp := range rps {
			if slice.In(filepath.Base(rp.name), trackedProps) {
				remoteParams = append(remoteParams, rp)
			}
		}
	}

	if len(remoteParams) == 0 {
		return file, errors.New("parameters not found, verify parameters exist in current account")
	}

	paramMap := toParamMap(remoteParams)

	switch format {
	case output.TypeTerraform:
		dir := filepath.Dir(file.LocalPath)

		filePath := fmt.Sprintf("%s/kms.tf", dir)
		printFile(filePath, s.io.Stderr)

		if err := terraform.EnsureTFKMSFile(filePath, blankDefault(file.Options[KMSKeyIDOption], PSKMSKeyIDDefault)); err != nil {
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

		d, err := s.terraform(paramMap, file, s.io)
		file.Data = d
		return file, err
	case output.TypeECSTaskInjectJson:
		d, err := taskDefJsonTransformSecrets(paramMap)
		file.Data = d
		return file, err
	case output.TypeECSTaskInjectEnv:
		d, err := taskDefEnvTransformSecrets(paramMap)
		file.Data = d
		return file, err
	}

	var buffer bytes.Buffer
	for key, param := range paramMap {

		name := key[strings.LastIndex(key, "/")+1 : len(key)]

		if isString(param.Value) {
			buffer.WriteString(fmt.Sprintf("%s=\"%s\"\n", name, param))
		} else {
			buffer.WriteString(fmt.Sprintf("%s=%s\n", name, param))
		}
	}

	file.Data = buffer.Bytes()

	return file, nil
}

func (s ParameterStoreService) terraform(m map[string]value, file File, io IO) ([]byte, error) {

	var hcl bytes.Buffer
	w := bufio.NewWriter(&hcl)

	arns := []string{}
	for _, value := range m {
		arns = append(arns, value.ARN)
	}

	if len(arns) > 0 {
		arns = append(arns, filepath.Dir(arns[0]))
	}

	p, err := json.MarshalIndent(policy.New(policy.Statement{
		Effect:   "Allow",
		Action:   []string{"ssm:GetParameters", "ssm:GetParametersByPath"},
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
		Name:   format.TerraformResourceName(filepath.Base(file.RemoteKey)),
		Policy: string(p),
	}); err != nil {
		return []byte{}, err
	}

	w.Flush()

	return hcl.Bytes(), nil
}

func toParamMap(params []param) map[string]value {
	data := map[string]value{}

	for _, p := range params {
		data[p.name] = value{
			ARN:   p.arn,
			Value: p.value,
		}
	}

	return data
}

func getRemoteParams(remoteKey string, svc *ssm.SSM) ([]param, error) {
	parameters := []param{}

	remoteParams, err := getRemoteParamsByPath(svc, remoteKey, "", []*ssm.Parameter{})
	if err != nil {
		return nil, err
	}

	for _, sp := range remoteParams {
		parameters = append(parameters, param{
			name:         *sp.Name,
			value:        *sp.Value,
			pType:        *sp.Type,
			lastModified: *sp.LastModifiedDate,
			arn:          *sp.ARN,
		})
	}

	return parameters, nil
}

// Purge ...
func (s ParameterStoreService) Purge(file File) error {
	if err := s.ensureSession(); err != nil {
		return err
	}

	svc := ssm.New(s.session)

	remoteParams := []param{}

	for remoteKeyPath, trackedProps := range getKeyPaths(file) {

		rps, err := getRemoteParams(remoteKeyPath, svc)
		if err != nil {
			return err
		}

		for _, rp := range rps {
			if slice.In(filepath.Base(rp.name), trackedProps) {
				remoteParams = append(remoteParams, rp)
			}
		}
	}

	for _, p := range remoteParams {
		if _, err := svc.DeleteParameter(&ssm.DeleteParameterInput{
			Name: aws.String(p.name),
		}); err != nil {
			return fmt.Errorf("%s: %s", p.name, err)
		}

		file.RemoveKey(p.name)
	}

	return nil
}

func init() {
	s := new(ParameterStoreService)

	Services[s.Key()] = s
}

type param struct {
	key   string
	name  string
	value string

	pType string
	keyID string

	lastModified time.Time

	arn string
}

func (p param) Changed(rp param) bool {
	return p.value != rp.value || rp.pType != p.pType || rp.keyID != p.keyID
}
