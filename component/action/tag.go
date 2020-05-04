package action

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dabblebox/stash/component/catalog"
	"github.com/dabblebox/stash/component/service"
	"github.com/gookit/color"
)

// TagOpt ...
type TagOpt struct {
	Options

	Add    []string
	Delete []string
}

// Tag deletes tracked files from the local folders.
func Tag(opt TagOpt, dep Dep) error {

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

	tagged := 0

	tags := color.FgCyan.Render

	for serviceKey, catalogFiles := range catalog.GroupByService(targetFiles) {

		fmt.Fprintf(dep.Stderr, "\n%s\n\n", bold(service.Name(serviceKey)))

		for key, cf := range catalogFiles {

			tagsChanged := false

			if (len(opt.Files) > 0 || len(opt.Service) > 0) && len(opt.Tags) > 0 {
				cf.Tags = []string{}

				for _, t := range opt.Tags {
					cf.AddTag(t)
					tagsChanged = true
				}
			}

			for _, t := range opt.Add {
				cf.AddTag(t)
				tagsChanged = true
			}

			for _, t := range opt.Delete {
				cf.RemoveTag(t)
				tagsChanged = true
			}

			c.Files[key] = cf

			if len(cf.Tags) > 0 {
				fmt.Fprintf(dep.Stderr, "- [%s]\n    tags: %s\n", filePathColor(cf.Path), tags(strings.Join(cf.Tags, ",")))
			} else {
				fmt.Fprintf(dep.Stderr, "- [%s] \n", filePathColor(cf.Path))
			}

			if tagsChanged {
				tagged++
			}
		}
	}

	fmt.Fprintf(dep.Stderr, "\n%d file(s) tagged\n\n", tagged)

	if len(dep.Monitor.Errors) > 0 {
		return errors.New("tagging errors detected")
	}

	if tagged > 0 {
		if err := catalog.Save(opt.Catalog, c); err != nil {
			return err
		}
	}

	return nil
}
