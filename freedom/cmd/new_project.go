package cmd

import (
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/8treenet/freedom/freedom/template/project"

	"github.com/spf13/cobra"
)

var (
	NewProjectCmd = &cobra.Command{
		Use:   "new-project [project_name]",
		Short: "New a microservice project based on freedom",
		Long:  `New project from freedom project template. `,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			sysPath, err := filepath.Abs(args[0])
			if err != nil {
				return
			}

			projectPath := strings.Replace(sysPath, build.Default.GOPATH+"/src/", "", 1)
			projectName := args[0]
			pdata := map[string]interface{}{
				"PackagePath": projectPath,
				"PackageName": projectName,
				"VersionNum":  versionNum,
			}
			if !strings.Contains(sysPath, build.Default.GOPATH) {
				pdata["PackagePath"] = projectName
			}

			mkdirAll(sysPath)
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
	os.MkdirAll(projectPath+"/adapter/controllers", os.ModePerm)
	os.MkdirAll(projectPath+"/adapter/repositorys", os.ModePerm)
	os.MkdirAll(projectPath+"/application", os.ModePerm)
	os.MkdirAll(projectPath+"/application/objects", os.ModePerm)
	os.MkdirAll(projectPath+"/application/aggregates", os.ModePerm)
	os.MkdirAll(projectPath+"/application/entitys", os.ModePerm)
	os.MkdirAll(projectPath+"/infra/config", os.ModePerm)
}
