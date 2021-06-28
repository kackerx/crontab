package master

import (
	"encoding/json"
	"fmt"
	"github.com/kackerx/crontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiServer struct {
	httpServer *http.Server
}

var (
	// 单例
	G_apiServer *ApiServer
)

// 初始化服务
func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux
		listener   net.Listener
		httpServer *http.Server
	)

	// TCP监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		fmt.Println(err)
		return
	}

	// 创建HTTP服务
	httpServer = &http.Server{
		Handler:      mux,
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
	}

	// 路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)

	// 赋值单例
	G_apiServer = &ApiServer{httpServer: httpServer}

	// 服务端拉起
	go httpServer.Serve(listener)

	return

}

// 保存Job
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	// 任务保存到ETCD
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)
	// 1, 解析post表单
	if err = r.ParseForm(); err != nil {
		return
	}

	// 2, 取job字段
	postJob = r.PostForm.Get("job")

	// 3, 反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		return
	}

	// 4, 保存
	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		return
	}

	// 5, 正常应答
	if bytes, err = common.BuildResponse(0, "success", oldJob); err != nil {
		goto ERR
	}
	w.Write(bytes)

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err != nil {
		w.Write(bytes)
	}

}
