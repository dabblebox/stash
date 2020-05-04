package action

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/service"
)

// Clean deletes tracked files from the local folders.
func Clean(opt Options, dep Dep) error {

	//-------------------------------------
	//- Init Catalog
	//-------------------------------------
	c, err := catalog.Read(opt.Catalog)
	if err != nil {
		return fmt.Errorf("%s: %s", opt.Catalog, err)
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
		return fmt.Errorf("%s does not contain matching %s ", opt.Catalog, filter.Format(" or "))
	}

	deleted := 0
	for serviceKey, catalogFiles := range catalog.GroupByServiceSlice(targetFiles) {

		fmt.Fprintf(dep.Stderr, "\n%s\n\n", bold(service.Name(serviceKey)))

		sort.Sort(catalog.ByPath(catalogFiles))

		for _, cf := range catalogFiles {
			if _, err := os.Stat(cf.Path); err == nil {
				fmt.Fprintf(dep.Stderr, "- [%s]\n", filePathColor(cf.Path))

				if err := os.Remove(cf.Path); err != nil {
					dep.Monitor.FileError(err)
					continue
				}

				deleted++
			}
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) deleted locally\n\n", deleted)

	if len(dep.Monitor.Errors) > 0 {
		return errors.New("cleaning errors detected")
	}

	return nil
}
