package action

import (
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dabblebox/stash/component/dotenv"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/output"
	"github.com/dabblebox/stash/component/service"
	"github.com/dabblebox/stash/component/token"
)

// InjectOpt ...
type InjectOpt struct {
	Data  []byte
	Files []string

	Service string
	Output  string
}

// Inject ...
func Inject(opt InjectOpt, dep Dep) ([]DownloadedFile, error) {

	injected := []DownloadedFile{}

	remote, ok := service.Services[opt.Service]
	if !ok {
		return injected, fmt.Errorf("service %s not found ", opt.Service)
	}

	if err := remote.PreHook(service.IO{
		Stdin:  dep.Stdin,
		Stdout: dep.Stdout,
		Stderr: dep.Stderr,
	}); err != nil {
		return injected, fmt.Errorf("service %s failed to initialize: %s", opt.Service, err)
	}

	fmt.Fprintf(dep.Stderr, "\n%s (injecting)\n\n", bold(service.Name(opt.Service)))

	files := map[string][]byte{}

	if len(opt.Data) > 0 {
		files["InjectOpt.Data"] = opt.Data
	}

	for _, fp := range opt.Files {

		data, err := file.Read(fp)
		if err != nil {
			dep.Monitor.FileError(err)
			continue
		}

		files[fp] = data
	}

	for path, data := range files {

		fmt.Fprintf(dep.Stderr, "- [%s]\n", filePathColor(path))

		keys := token.Find(data)

		if len(keys) == 0 {
			dep.Monitor.FileWarn("tokens not found")
			continue
		}

		for fileKey, remoteKey := range keys {

			remoteFileType := ""
			if len(remoteKey.Field) > 0 {
				remoteFileType = file.TypeEnv
			}

			stashFile := service.File{
				RemoteKey: remoteKey.String(),
				Type:      remoteFileType,
				Keys:      []string{remoteKey.String()},
			}

			fmt.Fprintf(dep.Stderr, "  - ${%s} => (%s)\n", fileTokenColor(fileKey), fileTokenColor(remoteKey))

			result, err := remote.Download(stashFile, opt.Output)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			m := map[string]string{}

			if len(remoteKey.Field) > 0 {
				params, err := dotenv.Parse(bytes.NewReader(result.Data))
				if err != nil {
					return injected, err
				}

				v, found := params[remoteKey.Field]
				if !found {
					dep.Monitor.FileError(fmt.Errorf("field '%s' not found in '%s'", remoteKey.Field, remoteKey))
					continue
				}

				m[fileKey] = v
			} else {
				m[fileKey] = string(result.Data)
			}

			data = token.Replace(m, data)
		}

		t, err := output.GetTransformer(opt.Output, strings.TrimLeft(filepath.Ext(path), "."))
		if err != nil {
			dep.Monitor.FileError(err)
			continue
		}

		d, err := t.Transform(data)
		if err != nil {
			dep.Monitor.FileError(err)
			continue
		}

		injected = append(injected, DownloadedFile{
			Service: opt.Service,
			Path:    path,
			Output:  opt.Output,
			Data:    d,
		})
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) injected\n\n", len(injected))

	if len(dep.Monitor.Errors) > 0 {
		return injected, errors.New("inject errors detected")
	}

	return injected, nil
}
