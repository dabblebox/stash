package action

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/service"
)

// Purge deletes files from a remote service.
func Purge(opt Options, dep Dep) (int, int, error) {

	//-------------------------------------
	//- Init Catalog
	//-------------------------------------
	c, err := catalog.Read(opt.Catalog)
	if err != nil {
		return len(c.Files), 0, err
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
		return len(c.Files), 0, fmt.Errorf("%s does not contain matching %s ", opt.Catalog, filter.Format(" or "))
	}

	deleted := 0

	for serviceKey, catalogFiles := range catalog.GroupByService(c.Filter(filter)) {

		fmt.Fprintf(dep.Stderr, "\n%s (deleting)\n\n", bold(service.Name(serviceKey)))

		remote, ok := service.Services[serviceKey]
		if !ok {
			return len(c.Files), 0, fmt.Errorf("service %s not found ", serviceKey)
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
			sf, err := cf.ToServiceModel(c.Context, key, remote, []byte{})
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			fmt.Fprintf(dep.Stderr, "- [%s]\n", filePathColor(sf.RemoteKey))

			if opt.Warn && len(cf.Keys) > 0 {
				filePathConfirm := ""
				prompt := &survey.Input{
					Help:    "Permanently delete the remote file. If a local copy does not exists, configuration will be lost.",
					Message: "Enter above remote key to confirm delete?",
				}
				if err := survey.AskOne(prompt, &filePathConfirm,
					survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr),
				); err != nil {
					return len(c.Files), deleted, err
				}

				if filePathConfirm != sf.RemoteKey {
					continue
				}
			}

			if err := remote.Purge(sf); err != nil {
				return len(c.Files), deleted, err
			}

			delete(c.Files, key)

			if err := cf.RecordState(c.Context); err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			deleted++

			if err := catalog.Save(opt.Catalog, c); err != nil {
				return len(c.Files), deleted, err
			}
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) deleted\n\n", deleted)

	if len(dep.Monitor.Errors) > 0 {
		return len(c.Files), deleted, errors.New("delete errors detected")
	}

	return len(c.Files), deleted, nil
}
