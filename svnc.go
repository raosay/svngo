package main

import (
	"crypto/tls"
	"github.com/PuerkitoBio/goquery"
	"github.com/deanishe/awgo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	TRUNK = "trunk/"
)

var (
	maxResults = 200
	wf         *aw.Workflow
	//// Icon for bookmark filetype
	//icon = &aw.Icon{
	//	Value: "com.apple.safari.bookmark",
	//	Type:  aw.IconTypeFileType,
	//}
	username,password,host string
)

func init(){
	wf = aw.New(aw.MaxResults(maxResults))
}


func run(){
	var c conf
	conf:=c.getConf()
	if conf.Password == "" ||conf.Username == "" {
		wf.NewWarningItem("还未设置svn账号密码","在根目录conf.yaml设置")
		wf.SendFeedback()
		return
	}
	if conf.Host == "" ||conf.Local == "" {
		wf.NewWarningItem("配置文件缺少下载路径","在根目录conf.yaml设置")
		wf.SendFeedback()
		return
	}
	username = conf.Username
	password = conf.Password
	host = conf.Host
	var query string
	var checkout string
	if args := wf.Args(); len(args)>0 {
		query = args[0]
	}
	if args := wf.Args(); len(args)>1 {
		checkout = args[1]
	}

	matchSvn := matchSvn(query)
	if len(checkout) == 0 {
		for _, svn := range matchSvn {
			item := wf.NewItem(svn).Subtitle(host+svn)
			item.Arg(host+svn+"/"+TRUNK).Autocomplete(svn +" -checkout").Valid(false)
		}
	}else if checkout == "-checkout" {
		var exactMatch string
		for _, svn := range matchSvn {
			if svn == query{
				exactMatch = svn
				break
			}
		}
		if len(exactMatch) <= 0 {
			wf.NewItem("项目名没有准确匹配项").Subtitle("无法指定下载项,重新匹配").
				Autocomplete(query).Valid(false)
		}
		checkoutItems(exactMatch)

	}
	wf.SendFeedback()
}

func main() {

	wf.Run(run)
}

func matchSvn(query string) []string{
	svnDir := httpDO(host)
	var matchProject []string
	for _, svn := range svnDir {
		if strings.Contains(svn,query) {
			matchProject = append(matchProject,svn)
		}
	}
	return matchProject


}

/******
 *使用前需要校验当前项目名完整
 */
func checkoutItems(projectName string){
	//trunk
	trunkUrl := host + projectName
	//branches
	branchUrl := host+projectName+"/branches/"
	branches := httpDO(host+projectName+"/branches/")
	//tag
	//tags := httpDO(HOST+projectName+"/branches/")
	wf.NewItem(projectName +"-trunk").Subtitle(trunkUrl).Arg(trunkUrl).Valid(true).
		Alt().Subtitle("下载到本地").Valid(true).Arg(projectName +" " +trunkUrl)

	//反转数组 倒序排列
	length := len(branches)
	for i := 0; i < length/2; i++ {
		temp := branches[length-1-i]
		branches[length-1-i] = branches[i]
		branches[i] = temp
	}

	for _, branch := range branches {
		wf.NewItem(branch).Subtitle("/branches/"+branch).Arg(branchUrl+branch).Valid(true).
			Alt().Subtitle("下载到本地").Valid(true).Arg(projectName +" " +"/branches/"+branch)
	}



}

func httpDO(url string) []string{
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{
		Transport: tr,
	}
	/****
	* request 对象包含SetBasicAuth方法逻辑如下
	 */
	//auth := USERNAME + ":" + PASSWORD
	//baseAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req ,_ := http.NewRequest("GET",url,nil)
	req.SetBasicAuth(username,password)
	resp , err := client.Do(req)
	if err != nil {
		log.Println("出现错误了")
		log.Println(err)
		panic(err)
	}
	defer resp.Body.Close()
	doc , err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	var returnStrings []string
	items := doc.Find("ul").Find("li")
	items.Each(func(i int,s *goquery.Selection ){
		a := s.Find("a")
		name := a.Text()
		if name != ".."{
			log.Println(name)
			name = strings.Replace(name,"/","",-1)
			returnStrings = append(returnStrings,name)
		}
	})

	return returnStrings
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



