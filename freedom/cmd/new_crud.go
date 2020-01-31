package cmd

import (
	"bytes"
	"go/build"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/8treenet/freedom/freedom/template/crud"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	packageName = "objects"
	Conf        = "./application/conf/db.toml"
	OutObj      = "./application/objects"
	OutFunc     = "./adapter/repositorys"
	NewCRUDCmd  = &cobra.Command{
		Use:   "new-crud",
		Short: "Create the model code for the CRUD.",
		Long:  `Create the model code for the CRUD, You can view subcommands and customize builds`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tl := crud.CrudTemplate()
			funTempl := crud.FunTemplate()
			s2 := crud.NewTable2Struct()
			dc := DBConf{}
			if err = configure(&dc, Conf); err != nil {
				return
			}
			list, err := s2.Dsn(dc.Addr).Run()
			if err != nil {
				return err
			}
			generateBuffer := new(bytes.Buffer)
			tmpl, err := template.New("").Parse(crud.FunTemplatePackage())
			if err != nil {
				return err
			}

			sysPath, err := os.Getwd()
			if err != nil {
				return err
			}
			pkg, err := build.ImportDir(sysPath+"/application/objects", build.IgnoreVendor)
			if err != nil {
				return err
			}
			pdata := map[string]interface{}{
				"PackagePath": pkg.ImportPath,
			}
			if err = tmpl.Execute(generateBuffer, pdata); err != nil {
				return err
			}

			for index := 0; index < len(list); index++ {
				pdata := map[string]interface{}{
					"Name":    list[index].Name,
					"Content": list[index].Content,
					"Time":    false,
					"Fields":  list[index].Fields,
				}
				if strings.Contains(list[index].Content, "time.Time") {
					pdata["Time"] = true
				}

				var pf *os.File
				pf, err = os.Create(OutObj + "/" + list[index].TableRealName + ".go")
				if err != nil {
					return err
				}
				tmpl, err := template.New("").Parse(tl)
				if err != nil {
					return err
				}
				if err = tmpl.Execute(pf, pdata); err != nil {
					return err
				}

				tmpl, err = template.New("").Parse(funTempl)
				if err != nil {
					return err
				}
				if err = tmpl.Execute(generateBuffer, pdata); err != nil {
					return err
				}
			}
			ioutil.WriteFile(OutFunc+"/"+"generate.go", generateBuffer.Bytes(), 0755)
			exec.Command("gofmt", "-w", OutObj).Output()
			exec.Command("gofmt", "-w", OutFunc).Output()

			return nil
		}}
)

type DBConf struct {
	Addr string `toml:"addr"`
}

// configure .
func configure(obj interface{}, fileName string) error {
	_, err := toml.DecodeFile(fileName, obj)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	NewCRUDCmd.Flags().StringVarP(&Conf, "conf", "c", "./server/conf/db.toml", "mysql profile path")
	AddCommand(NewCRUDCmd)
}
