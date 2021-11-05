package cmd

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"github.com/8treenet/freedom/freedom/template/project"

	"github.com/spf13/cobra"
)

var (
	// NewProjectCmd .
	NewProjectCmd = &cobra.Command{
		Use:   "new-project [project_name]",
		Short: "New a microservice project based on freedom",
		Long:  `New project from freedom project template. `,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) < 1 || args[0] == "" {
				return errors.New("[project_name] empty")
			}
			sysPath, err := filepath.Abs(args[0])
			if err != nil {
				return
			}
			mkdirAll(sysPath)

			projectName := args[0]
			pdata := map[string]interface{}{
				"PackagePath": projectName,
				"PackageName": projectName,
				"VersionNum":  versionNum,
			}

			m := project.FileContent()
			for filepath, content := range m {
				var pf *os.File
				pf, err = os.Create(sysPath + filepath)
				if err != nil {
					return err
				}
				tmpl, err := template.New(projectName).Parse(content)
				if err = tmpl.Execute(pf, pdata); err != nil {
					return err
				}
			}
			exec.Command("gofmt", "-w", sysPath).Output()
			return nil
		},
	}
)

func init() {
	AddCommand(NewProjectCmd)
}

func mkdirAll(projectPath string) {
	os.MkdirAll(projectPath+"/server", os.ModePerm)
	os.MkdirAll(projectPath+"/server/conf", os.ModePerm)
	os.MkdirAll(projectPath+"/adapter", os.ModePerm)
	os.MkdirAll(projectPath+"/adapter/controller", os.ModePerm)
	os.MkdirAll(projectPath+"/adapter/repository", os.ModePerm)
	os.MkdirAll(projectPath+"/domain", os.ModePerm)
	os.MkdirAll(projectPath+"/domain/aggregate", os.ModePerm)
	os.MkdirAll(projectPath+"/domain/entity", os.ModePerm)
	os.MkdirAll(projectPath+"/domain/vo", os.ModePerm)
	os.MkdirAll(projectPath+"/domain/po", os.ModePerm)
	os.MkdirAll(projectPath+"/domain/event", os.ModePerm)
	os.MkdirAll(projectPath+"/infra", os.ModePerm)
}
