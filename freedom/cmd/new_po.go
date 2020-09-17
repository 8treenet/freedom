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
	packageName = "po"
	// Conf .
	Conf = "./server/conf/db.toml"
	// OutObj .
	OutObj = "./domain/po"
	// OutFunc .
	OutFunc = "./adapter/repository"
	// NewCRUDCmd .
	NewCRUDCmd = &cobra.Command{
		Use:   "new-po",
		Short: "Create the model code for the CRUD.",
		Long:  `Create the model code for the CRUD, You can view subcommands and customize builds`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			os.MkdirAll(OutObj, os.ModePerm)
			os.MkdirAll(OutFunc, os.ModePerm)
			tl := crud.PoDefContent()
			funTempl := crud.FunTemplate()
			s2 := crud.NewTable2Struct()
			dc := dbConf{}
			if err = configure(&dc, Conf); err != nil {
				return
			}
			list, err := s2.Dsn(dc.Addr).Run()
			if err != nil {
				return err
			}
			generateBuffer := new(bytes.Buffer)
			generateTmpl, err := template.New("").Parse(crud.FunTemplatePackage())
			if err != nil {
				return err
			}

			sysPath, err := os.Getwd()
			generateBuild := false

			for index := 0; index < len(list); index++ {
				pdata := map[string]interface{}{
					"Name":         list[index].Name,
					"Content":      list[index].Content,
					"Time":         false,
					"Fields":       list[index].Fields,
					"NumberFields": list[index].NumberFields,
					"Import":       "",
				}
				pdata["Import"] = "import (\n"
				if strings.Contains(list[index].Content, "time.Time") {
					pdata["Import"] = pdata["Import"].(string) + `"time"` + "\n"
				}
				if len(list[index].NumberFields) > 0 {
					pdata["Import"] = pdata["Import"].(string) + `"github.com/jinzhu/gorm"` + "\n"
				}
				pdata["Import"] = pdata["Import"].(string) + ")"

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

				if !generateBuild {
					generatePkg, err := build.ImportDir(sysPath+"/domain/po", build.IgnoreVendor)
					if err != nil {
						return err
					}
					generatePdata := map[string]interface{}{
						"PackagePath": generatePkg.ImportPath,
					}
					if err = generateTmpl.Execute(generateBuffer, generatePdata); err != nil {
						return err
					}
					generateBuild = true
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

type dbConf struct {
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
