package action

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/output"
	"github.com/dabblebox/stash/component/service"
)

// GetOpt ...
type GetOpt struct {
	Options

	Output string
}

// DownloadedFile ..
type DownloadedFile struct {
	Service string

	Path   string
	Output string

	Data []byte
}

// Get downloads files from a remote service.
func Get(opt GetOpt, dep Dep) ([]DownloadedFile, error) {

	//-------------------------------------
	//- Init Catalog
	//-------------------------------------
	c, err := catalog.Read(opt.Catalog)
	if err != nil {
		return []DownloadedFile{}, err
	}

	c, err = catalog.ExpandEnv(c)
	if err != nil {
		return []DownloadedFile{}, err
	}

	//-------------------------------------
	//- Filter Files
	//-------------------------------------
	filter := catalog.NewGetFilter(opt.Files, opt.Tags, opt.Service)

	targetFiles := c.Filter(filter)

	//-------------------------------------
	//- Validate Request
	//-------------------------------------
	if len(targetFiles) == 0 {
		return []DownloadedFile{}, fmt.Errorf("%s does not contain matching %s ", opt.Catalog, filter.Format(" or "))
	}

	downloaded := []DownloadedFile{}

	for serviceKey, catalogFiles := range catalog.GroupByService(targetFiles) {

		fmt.Fprintf(dep.Stderr, "\n%s (downloading)\n\n", bold(service.Name(serviceKey)))

		remote, ok := service.Services[serviceKey]
		if !ok {
			dep.Monitor.Error(fmt.Errorf("service %s not found ", serviceKey))
			continue
		}

		if err := remote.PreHook(service.IO{
			Stdin:  dep.Stdin,
			Stdout: dep.Stdout,
			Stderr: dep.Stderr,
		}); err != nil {
			dep.Monitor.Error(fmt.Errorf("service %s failed to initialize: %s", serviceKey, err))
			continue
		}

		for key, cf := range catalogFiles {

			stashFile, err := cf.ToServiceModel(c.Context, key, remote, []byte{})
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			fmt.Fprintln(dep.Stderr, formatFileDownloadText(c.Context, opt.Output, cf.Path, stashFile.RemoteKey, cf.Service))

			result, err := remote.Download(stashFile, opt.Output)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			t, err := output.GetTransformer(opt.Output, cf.Type)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			d, err := t.Transform(result.Data)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			downloaded = append(downloaded, DownloadedFile{
				Service: cf.Service,
				Path:    cf.Path,
				Output:  opt.Output,
				Data:    d,
			})

			if opt.Output == output.TypeFile {
				if err := cf.RecordState(c.Context); err != nil {
					dep.Monitor.FileError(err)
					continue
				}
			}
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) downloaded\n\n", len(downloaded))

	if len(dep.Monitor.Errors) > 0 {
		return downloaded, errors.New("download errors detected")
	}

	return downloaded, nil
}

func formatFileDownloadText(context, outputType, path, remoteKey, service string) string {
	loc := "unknown"

	switch outputType {
	case output.TypeFile:
		loc = path
	case output.TypeTerraform:
		loc = file.Number(fmt.Sprintf("%s/%s.tf", filepath.Dir(path), service))
	default:
		loc = "output"
	}

	return fmt.Sprintf("- [%s] => [%s]", filePathColor(remoteKey), filePathColor(loc))
}
