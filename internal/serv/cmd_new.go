package serv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/spf13/cobra"
	"github.com/valyala/fasttemplate"
)

func cmdNew(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help() //nolint: errcheck
		os.Exit(1)
	}

	tmpl := newTempl(map[string]string{
		"app_name":      strings.Title(strings.Join(args, " ")),
		"app_name_slug": strings.ToLower(strings.Join(args, "_")),
	})

	// Create app folder and add relevant files

	name := args[0]
	appPath := path.Join("./", name)

	ifNotExists(appPath, func(p string) error {
		return os.Mkdir(p, os.ModePerm)
	})

	ifNotExists(path.Join(appPath, "Dockerfile"), func(p string) error {
		if v, err := tmpl.get("Dockerfile"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	ifNotExists(path.Join(appPath, "docker-compose.yml"), func(p string) error {
		if v, err := tmpl.get("docker-compose.yml"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	ifNotExists(path.Join(appPath, "cloudbuild.yaml"), func(p string) error {
		if v, err := tmpl.get("cloudbuild.yaml"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	// Create app config folder and add relevant files

	appConfigPath := path.Join(appPath, "config")

	ifNotExists(appConfigPath, func(p string) error {
		return os.Mkdir(p, os.ModePerm)
	})

	ifNotExists(path.Join(appConfigPath, "dev.yml"), func(p string) error {
		if v, err := tmpl.get("dev.yml"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	ifNotExists(path.Join(appConfigPath, "prod.yml"), func(p string) error {
		if v, err := tmpl.get("prod.yml"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	ifNotExists(path.Join(appConfigPath, "seed.js"), func(p string) error {
		if v, err := tmpl.get("seed.js"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	// Create app migrations folder and add relevant files

	appMigrationsPath := path.Join(appConfigPath, "migrations")

	ifNotExists(appMigrationsPath, func(p string) error {
		return os.Mkdir(p, os.ModePerm)
	})

	ifNotExists(path.Join(appMigrationsPath, "0_init.sql"), func(p string) error {
		if v, err := tmpl.get("0_init.sql"); err == nil {
			return ioutil.WriteFile(p, v, 0644)
		} else {
			return err
		}
	})

	log.Printf("INR app '%s' initialized", name)
}

type Templ struct {
	*rice.Box
	data map[string]string
}

func newTempl(data map[string]string) *Templ {
	return &Templ{rice.MustFindBox("./tmpl"), data}
}

func (t *Templ) get(name string) ([]byte, error) {
	v := t.MustString(name)
	b := bytes.Buffer{}
	tmpl := fasttemplate.New(v, "{%", "%}")

	_, err := tmpl.ExecuteFunc(&b, func(w io.Writer, tag string) (int, error) {
		if val, ok := t.data[strings.TrimSpace(tag)]; ok {
			return w.Write([]byte(val))
		}
		return 0, fmt.Errorf("unknown template variable '%s'", tag)
	})

	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func ifNotExists(filePath string, doFn func(string) error) {
	_, err := os.Stat(filePath)

	if err == nil {
		log.Printf("ERR create skipped '%s' exists", filePath)
		return
	}

	if !os.IsNotExist(err) {
		log.Fatalf("ERR unable to check if '%s' exists", filePath)
	}

	err = doFn(filePath)
	if err != nil {
		log.Fatalf("ERR unable to create '%s'", filePath)
	}

	log.Printf("INR created '%s'", filePath)
}
