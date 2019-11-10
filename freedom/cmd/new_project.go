package cmd

import (
	"fmt"
	"go/build"
	"io/ioutil"
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
			gopath := os.Getenv("GOPATH")
			if !strings.Contains(sysPath, gopath+"/src/") {
				return fmt.Errorf("Project path must be within '%s'", gopath+"/src/")
			}

			projectPath := strings.Replace(sysPath, build.Default.GOPATH+"/src/", "", 1)
			projectName := args[0]
			pdata := map[string]interface{}{
				"PackagePath": projectPath,
				"PackageName": projectName,
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
			/*
				tp, mod, err := getTemplatePath(gopath)
				if err != nil {
					return
				}
				execcmd := exec.Command("cp", "-r", tp, sysPath)
				if _, err = execcmd.Output(); err != nil {
					return
				}
				if mod {
					//chmod -R 755 sysPath
					chmod := exec.Command("chmod", "-R", "755", sysPath)
					chmod.Output()
				}

				projectPath := strings.Replace(sysPath, build.Default.GOPATH+"/src/", "", 1)
				projectName := args[0]
				// fmt.Println("New project", sysPath, projectPath, projectName)
				pdata := map[string]interface{}{
					"PackagePath": projectPath,
					"PackageName": projectName,
				}

				fileList, err := getAllFile(sysPath, []string{})
				for index := 0; index < len(fileList); index++ {
					if !strings.Contains(fileList[index], ".template") {
						continue
					}
					text, err := ioutil.ReadFile(fileList[index])
					if err != nil {
						return err
					}

					var pf *os.File
					newFile := strings.Split(fileList[index], ".template")
					pf, err = os.Create(newFile[0])
					if err != nil {
						return err
					}

					tmpl, err := template.New(projectName).Parse(string(text))
					if err != nil {
						return err
					}

					if err = tmpl.Execute(pf, pdata); err != nil {
						return err
					}
					os.Remove(fileList[index])
				}
			*/
		},
	}
)

func init() {
	AddCommand(NewProjectCmd)
}

func getAllFile(pathname string, s []string) ([]string, error) {
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		fmt.Println("read dir fail:", err)
		return s, err
	}
	for _, fi := range rd {
		if fi.IsDir() {
			fullDir := pathname + "/" + fi.Name()
			s, err = getAllFile(fullDir, s)
			if err != nil {
				fmt.Println("read dir fail:", err)
				return s, err
			}
		} else {
			fullName := pathname + "/" + fi.Name()
			s = append(s, fullName)
		}
	}
	return s, nil
}

// getTemplatePath
func getTemplatePath(gopath string) (string, bool, error) {
	_, err := os.Stat(gopath + "/src/github.com/8treenet/freedom/freedom/template/project")
	if err == nil {
		return gopath + "/src/github.com/8treenet/freedom/freedom/template/project", false, nil
	}
	rds, err := ioutil.ReadDir(gopath + "/pkg/mod/github.com/8treenet")
	if err != nil {
		return "", false, nil
	}
	for index := 0; index < len(rds); index++ {
		if strings.Contains(rds[index].Name(), "freedom") {
			return gopath + "/pkg/mod/github.com/8treenet/" + rds[index].Name() + "/freedom/template/project", true, nil
		}
	}
	return "", false, fmt.Errorf("unknown error")
}

func mkdirAll(projectPath string) {
	os.MkdirAll(projectPath+"/cmd", os.ModePerm)
	os.MkdirAll(projectPath+"/cmd/conf", os.ModePerm)
	os.MkdirAll(projectPath+"/business/controllers", os.ModePerm)
	os.MkdirAll(projectPath+"/models", os.ModePerm)
	os.MkdirAll(projectPath+"/components/config", os.ModePerm)
	os.MkdirAll(projectPath+"/business/repositorys", os.ModePerm)
	os.MkdirAll(projectPath+"/business/services", os.ModePerm)
}
