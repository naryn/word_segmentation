package main

import (
	"fmt"
	"strings"
	"log"
	"os"
	"net/http"
	"regexp"
	"utils"
	"utils/config"
	//"github.com/tim1020/godaemon"
	"models"
//	"github.com/yanyiwu/gojieba"
)


var (
	accLogger   *utils.AccessLogger
	WordSegmentation *models.WordSegmentation
)

//根据配置文件初始化日志对象
func init() {
	var err error
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("配置文件异常: %v\n", err)
			os.Exit(0)
		}
	}()

	models.Config, err = config.NewConfig("json", "conf/config.json")
	if err != nil {
		panic(err)
	}

	//系统日志
	sysLogFile, err := os.OpenFile(models.Config.String("log_sys_path"), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(sysLogFile)
	//访问日志
	accLogger = utils.NewAccLogger(models.Config.String("log_access_prefix"), ".log")

	WordSegmentation = models.NewWordSegmentation()
}

func main() {

	log.Printf("ApiServer %v start\n", models.Config.String("sys_srv_addr"))
	http.HandleFunc("/", handle)                       //设置访问的路由
	//err := godaemon.GracefulServe(config["addr"], nil) //设置监听的端口
	err := http.ListenAndServe(models.Config.String("sys_srv_addr"), nil)
	if err != nil {
		fmt.Println("Server Err: ", err)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	timer := utils.NewTimer()
	r.ParseForm()
	defer func() {
		if err := recover(); err != nil {
			log.Printf("handle catch panic: %v\n", err)
		}
		accLogger.LogTimer(path, r.FormValue("w"), timer)
	}()

	urlRegexp := regexp.MustCompile(`/?v?([\d|\.]*)?/([^/]+)\.html`)
	matches := urlRegexp.FindStringSubmatch(path)
	if len(matches) < 3 {
		fmt.Fprintf(w, "api not exist")
		return
	}
	module, method := parseUri(matches[2], r)

	s := strings.Join(r.Form["w"],"")
	result := ""
	//http://127.0.0.1:9092/app/web/v3.1/segmentation-cutAll.html?w=xxxx
	switch module + "-" + method {
	case "segmentation-cutAll":
		result = WordSegmentation.CutAll(s)
	case "segmentation-cut":
		result = WordSegmentation.Cut(s)
	case "segmentation-cutForSearch":
		result = WordSegmentation.CutForSearch(s)
	case "segmentation-tag":
		result = WordSegmentation.Tag(s)
	default:
		fmt.Fprint(w, "Api not exist")
		return
	}

	w.Header().Add("Cache-Control", "no-cache, private, max-age=0")
	w.Header().Add("Expires", "Mon, 26 Jul 2013 05:00:00 GMT")
	w.Header().Set("Content-Type", "text/html;charset=utf-8")

	fmt.Fprint(w, result)
}

//解析module、method、GET参数
func parseUri(uri string, r *http.Request) (module, method string) {
	urls := strings.Split(uri, "-")
	module = urls[0]
	method = "index"
	var queryUrls []string

	if len(urls)%2 == 0 {
		method = urls[1]
		queryUrls = urls[2:]
	} else {
		queryUrls = urls[1:]
	}

	queryUrlsLen := len(queryUrls)
	if queryUrlsLen > 0 {
		r.ParseForm()
		for i := 0; i < queryUrlsLen/2; i++ {
			r.Form.Add(queryUrls[i*2], queryUrls[i*2+1])
		}
	}
	return
}



