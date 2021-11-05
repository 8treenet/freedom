package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/8treenet/freedom/freedom/template/crud"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	packageName = "po"
	//Dsn .
	Dsn = "root:123123@tcp(127.0.0.1:3306)/xxx?charset=utf8"
	//JSONFile .
	JSONFile = ""
	//OutObj .
	OutObj = "./domain/po"
	//OutFunc .
	OutFunc = "./adapter/repository"
	//Prefix .
	Prefix = ""
	//NewCRUDCmd .
	NewCRUDCmd = &cobra.Command{
		Use:   "new-po",
		Short: "Create the model code for the CRUD.",
		Long:  `Create the model code for the CRUD, You can view subcommands and customize builds`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			os.MkdirAll(OutObj, os.ModePerm)
			os.MkdirAll(OutFunc, os.ModePerm)
			tl := crud.PoDefContent()
			funTempl := crud.FunTemplate()
			list, err := GetStruct()
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
					"Name":       list[index].Name,
					"Content":    list[index].Content,
					"Time":       false,
					"SetMethods": list[index].SetMethods,
					"AddMethods": list[index].AddMethods,
					"Import":     "",
				}
				pdata["Import"] = "import (\n"
				if strings.Contains(list[index].Content, "time.Time") {
					pdata["Import"] = pdata["Import"].(string) + `"time"` + "\n"
				}

				if strings.Contains(list[index].Content, "datatypes.JSON") {
					pdata["Import"] = pdata["Import"].(string) + `"gorm.io/datatypes"` + "\n"
				}

				if len(list[index].AddMethods) > 0 {
					pdata["Import"] = pdata["Import"].(string) + `"gorm.io/gorm"` + "\n"
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
				fmt.Println(successString("Success [" + OutObj + "/" + list[index].TableRealName + ".go]"))

				if !generateBuild {
					generatePdata := map[string]interface{}{
						"PackagePath": path.Base(sysPath) + "/domain/po",
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
			fmt.Println(successString("Success [" + OutFunc + "/" + "generate.go]"))
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

// GetStruct .
func GetStruct() (list []crud.ObjectContent, e error) {
	defer func() {
		sort.Slice(list, func(i, j int) bool {
			return list[i].TableRealName > list[j].TableRealName
		})
	}()
	cmd := crud.NewGenerate()
	if Prefix != "" {
		cmd.SetPrefix(Prefix)
	}
	if Dsn != "" {
		list, e = cmd.Dsn(Dsn).RunDsn()
		return
	}
	if JSONFile != "" {
		list, e = cmd.Dsn(Dsn).RunDsn()
		return cmd.RunJSON(JSONFile)
	}

	e = errors.New("Wrong instruction")
	return
}

func init() {
	NewCRUDCmd.Flags().StringVarP(&Dsn, "dsn", "d", "", `The address of the data source "root:123123@tcp(127.0.0.1:3306)/xxx?charset=utf8"`)
	NewCRUDCmd.Flags().StringVarP(&JSONFile, "json", "j", "", `Table structure of JSON, "./domain/po/schema.json"`)
	NewCRUDCmd.Flags().StringVarP(&Prefix, "prefix", "p", "", `Ignore prefix`)

	AddCommand(NewCRUDCmd)
}

func successString(str string) string {
	return fmt.Sprintf("\x1b[0;%dm%s\x1b[0m", 32, str)
}
