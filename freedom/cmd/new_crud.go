package cmd

import (
	"os"
	"os/exec"
	"text/template"

	"github.com/8treenet/freedom/freedom/template/crud"
	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
)

var (
	packageName = "models"
	Conf        = "./cmd/conf/db.toml"
	Out         = "./models"
	NewCRUDCmd  = &cobra.Command{
		Use:   "new-crud",
		Short: "Create the model code for the CRUD.",
		Long:  `Create the model code for the CRUD, You can view subcommands and customize builds`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			tl := crud.CrudTemplate()
			s2 := crud.NewTable2Struct()
			dc := DBConf{}
			if err = configure(&dc, Conf); err != nil {
				return
			}
			list, err := s2.Dsn(dc.Addr).Run()
			if err != nil {
				return err
			}
			for index := 0; index < len(list); index++ {
				pdata := map[string]interface{}{
					"Name":    list[index].Name,
					"Content": list[index].Content,
				}

				var pf *os.File
				pf, err = os.Create(Out + "/" + list[index].TableRealName + ".go")
				if err != nil {
					return err
				}
				tmpl, err := template.New("").Parse(tl)
				if err = tmpl.Execute(pf, pdata); err != nil {
					return err
				}
			}
			exec.Command("gofmt", "-w", Out).Output()

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
	NewCRUDCmd.Flags().StringVarP(&Conf, "conf", "c", "./cmd/conf/db.toml", "mysql profile path")
	NewCRUDCmd.Flags().StringVarP(&Out, "out", "o", "./models", "The resulting model path")
	AddCommand(NewCRUDCmd)
}
