package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"gopkg.in/yaml.v2"
)


var (

	host,local string
)

func main() {
	var c conf
	conf:=c.getConf()
	if conf.Local == ""  {
		log.Println("未配置本地配置文件。。。 ",)
		panic("未配置本地环境变量")
		return
	}
	host = conf.Host
	local = conf.Local
	arr := os.Args
	if len(arr) < 3 {
		log.Print("缺少参数。。。%d ",len(arr))
		panic(arr)
	}

	name := arr[1]
	filePathi := arr[2]
	path  := filepath.Join(local,name,filePathi)
	info ,err :=os.Stat(path)
	if err !=nil {
		os.MkdirAll(path,0777)
		os.Chmod(path,0777)
	}

	loadUrl := host+name+"/"+filePathi
	fmt.Print(loadUrl)
	//svn checkout
	var ch string
	if info != nil && info.Size() > 1{
		ch = "svn update "+path+"/"
	}else{
		ch = "svn checkout "+loadUrl+" "+path+"/"
	}
	cmd := exec.Command("/bin/bash","-c",ch)
	o ,err1 := cmd.Output()
	if err1 !=nil {
		fmt.Print(err1)
		return
	}

	resp := string(o)
	fmt.Print(resp)


}

type conf struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host string `yaml:"host"`
	Local string `yaml:"local"`
}
func (c *conf) getConf() *conf {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		log.Println(err.Error())
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Println(err.Error())
	}
	return c
}
