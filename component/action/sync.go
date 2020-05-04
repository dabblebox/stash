package action

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/file"
	"github.com/dabblebox/stash/component/search"
	"github.com/dabblebox/stash/component/service"
	"github.com/gookit/color"
)

var (
	bold           = color.Bold.Render
	filePathColor  = color.FgLightMagenta.Render
	fileTokenColor = color.FgCyan.Render
)

// SyncOpt ...
type SyncOpt struct {
	Options

	Context string
}

func (so *SyncOpt) addFiles(paths []string) {

	for _, p := range paths {
		add := true

		for _, fp := range so.Files {
			if p == fp {
				add = false
			}
		}

		if add {
			so.Files = append(so.Files, p)
		}
	}
}

// Sync syncs files with a remote service.
func Sync(opt SyncOpt, dep Dep) error {

	//-------------------------------------
	//- Init Catalog
	//-------------------------------------
	c, err := catalog.Init(opt.Context, opt.Catalog, catalog.InitDep{
		Stdin:  dep.Stdin,
		Stdout: dep.Stdout,
		Stderr: dep.Stderr,
	})
	if err != nil {
		return err
	}

	//-------------------------------------
	//- Search By Regex (when not found)
	//-------------------------------------
	userSearched := []string{}
	userSpecified := []string{}

	for _, fp := range opt.Files {
		if info, err := os.Stat(fp); err != nil || info.IsDir() {
			results, err := search.By(fp)
			if err != nil {
				return err
			}

			if len(results) == 0 {
				return fmt.Errorf("%s: files not found", fp)
			}

			userSearched = append(userSearched, results...)
		} else {
			userSpecified = append(userSpecified, fp)
		}
	}
	opt.Files = userSpecified

	if len(userSearched) > 0 {

		userSelected := []string{}
		confirm := &survey.MultiSelect{
			Help: `If the results are not expected, ensure the specified file exists. 
When using a regex, wrap the expression in double quotes or escape
backslashes. Example: ^.*\\.env$`,
			Message:  "Sync",
			Options:  userSearched,
			PageSize: 20,
		}
		if err := survey.AskOne(confirm, &userSelected,
			survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr),
		); err != nil {
			return err
		}

		if len(userSelected) == 0 {
			return nil
		}

		opt.addFiles(userSelected)
	}

	//-------------------------------------
	//- Catalog New Files
	//-------------------------------------
	for _, fp := range opt.Files {
		if _, found := c.GetFile(fp); !found {

			remote, supported := service.Services[opt.Service]
			if !supported {
				compatible := service.ListCompatible([]string{fp})

				sort.Sort(service.BySecurityRating(compatible))

				value := ""
				prompt := &survey.Select{
					Message: fmt.Sprintf("Stash [%s]", filePathColor(fp)),
					Options: service.ToKeys(compatible),
				}

				if d := c.LookupService(fp); len(d) > 0 {
					prompt.Default = d
				}

				if err := survey.AskOne(prompt, &value,
					survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr),
				); err != nil {
					return err
				}

				remote = service.Services[value]
			}

			if err := c.AddFile("", fp, remote.Key(), opt.Tags); err != nil {
				return err
			}
		}
	}

	//-------------------------------------
	//- Filter Files
	//-------------------------------------
	filter := catalog.NewPushFilter(opt.Files, opt.Tags, opt.Service)

	targetFiles := c.Filter(filter)

	//-------------------------------------
	//- Validate Request
	//-------------------------------------
	if !filter.Empty() && len(targetFiles) == 0 {
		return fmt.Errorf("%s does not contain matching %s ", opt.Catalog, filter.Format(" or "))
	}

	//-------------------------------------
	//- Save Catalog
	//--------------------------------------
	if err := catalog.Save(opt.Catalog, c); err != nil {
		return err
	}

	//-------------------------------------
	//- Sync Catalog
	//-------------------------------------
	synced := 0

	for serviceKey, catalogFiles := range catalog.GroupByService(targetFiles) {

		fmt.Fprintf(dep.Stderr, "\n%s (synchronizing)\n\n", bold(service.Name(serviceKey)))

		remote, ok := service.Services[serviceKey]
		if !ok {
			dep.Monitor.Error(fmt.Errorf("service %s not found", serviceKey))
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
			fmt.Fprintln(dep.Stderr, formatFileSyncText(c.Context, key, cf.Path, service.FormatObjectKey(c.Context, cf.Path, remote)))

			data, err := file.Read(cf.Path)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			stashFile, err := cf.ToServiceModel(c.Context, key, remote, data)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			result, err := remote.Sync(stashFile)
			if err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			c.MergeResults([]service.File{result})

			if err := cf.RecordState(c.Context); err != nil {
				dep.Monitor.FileError(err)
				continue
			}

			synced++

			if c.AutoClean {
				if err := os.Remove(cf.Path); err != nil {
					dep.Monitor.FileError(err)
					continue
				}
			}
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) synced\n\n", synced)

	//-------------------------------------
	//- Save Catalog
	//-------------------------------------
	if err := catalog.Save(opt.Catalog, c); err != nil {
		return err
	}

	if len(dep.Monitor.Errors) > 0 {
		return errors.New("sync errors detected")
	}

	return nil
}

func formatFileSyncText(context, key, path, remote string) string {
	if len(path) > 0 {
		return fmt.Sprintf("- [%s] => [%s]", filePathColor(path), filePathColor(remote))
	}

	return fmt.Sprintf("- [%s] => [%s]", filePathColor(key), filePathColor(remote))
}
