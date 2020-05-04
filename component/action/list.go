package action

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/service"
	"github.com/gookit/color"
)

// List deletes tracked files from the local folders.
func List(opt Options, dep Dep) error {

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

	cataloged := 0

	tags := color.FgCyan.Render

	for serviceKey, catalogFiles := range catalog.GroupByServiceSlice(targetFiles) {

		fmt.Fprintf(dep.Stderr, "\n%s\n\n", bold(service.Name(serviceKey)))

		sort.Sort(catalog.ByPath(catalogFiles))

		for _, cf := range catalogFiles {

			if len(cf.Tags) > 0 {
				fmt.Fprintf(dep.Stderr, "- [%s]\n    tags: %s\n", filePathColor(cf.Path), tags(strings.Join(cf.Tags, ",")))
			} else {
				fmt.Fprintf(dep.Stderr, "- [%s] \n", filePathColor(cf.Path))
			}

			fmt.Fprintln(dep.Stderr, "    keys:")

			for _, k := range cf.Keys {
				fmt.Fprintf(dep.Stderr, "    - %s\n", fileTokenColor(k))
			}

			cataloged++
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) matched\n\n", cataloged)

	if len(dep.Monitor.Errors) > 0 {
		return errors.New("listing errors detected")
	}

	return nil
}
