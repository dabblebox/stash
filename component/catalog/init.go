package catalog

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

func initViperCatalog(catalog string) *viper.Viper {
	var v = viper.New()
	v.AutomaticEnv()

	v.SetEnvPrefix("stash")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigFile(catalog)
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetDefault("version", viper.GetString("version"))

	return v
}

// Read loads an existing catalog file.
func Read(catalogFile string) (Catalog, error) {

	v := initViperCatalog(catalogFile)

	if err := v.ReadInConfig(); err != nil {
		if _, notFound := err.(viper.ConfigFileNotFoundError); notFound || os.IsNotExist(err) {
			return Catalog{}, os.ErrNotExist
		}

		return Catalog{}, err
	}

	catalog := Catalog{Files: map[string]File{}}

	if err := v.Unmarshal(&catalog); err != nil {
		return Catalog{}, fmt.Errorf("Unable to decode into struct, %v", err)
	}

	return catalog, nil
}

// InitDep ...
type InitDep struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

// Init creates a new or loads an existing catalog file.
func Init(context, catalogFile string, dep InitDep) (Catalog, error) {

	v := initViperCatalog(catalogFile)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok && !os.IsNotExist(err) {
			return Catalog{}, err
		}

		autoClean := true

		if len(context) == 0 {

			prompt := &survey.Input{
				Message: "Context",
				Default: getParentFolder(),
				Help:    "Prefix for stashed data keys providing application or repository context.",
			}

			err := survey.AskOne(prompt, &context, survey.WithValidator(survey.Required), survey.WithValidator(func(val interface{}) error {
				if str, ok := val.(string); !ok {
					re := regexp.MustCompile(`^[a-z-]+$`)

					if !re.MatchString(str) {
						return errors.New("only lower case letters and hyphens")
					}
				}
				return nil
			}),
				survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr))
			if err != nil {
				return Catalog{}, err
			}

			confirm := &survey.Confirm{
				Help:    "Deleting local configuration files improves security by reducing the risk of locally stored secrets.",
				Message: "Delete local copy?",
				Default: true,
			}
			if err := survey.AskOne(confirm, &autoClean,
				survey.WithStdio(dep.Stdin, dep.Stdout, dep.Stderr)); err != nil {
				return Catalog{}, err
			}
		}

		v.Set("context", context)
		v.Set("clean", autoClean)
	}

	var catalog Catalog

	if err := v.Unmarshal(&catalog); err != nil {
		return Catalog{}, fmt.Errorf("Unable to decode into struct, %v", err)
	}

	return catalog, nil
}

func getParentFolder() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	directories := strings.Split(wd, "/")

	if len(directories) < 1 {
		return ""
	}

	return directories[len(directories)-1]
}
