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
        mux           *http.ServeMux
        listener      net.Listener
        httpServer    *http.Server
        staticDir     http.Dir
        staticHandler http.Handler
    )

    // TCP监听
    if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
        fmt.Println(err)
        return
    }

    // 路由
    mux = http.NewServeMux()
    mux.HandleFunc("/job/save", handleJobSave)
    mux.HandleFunc("/job/delete", handleJobDelete)
    mux.HandleFunc("/job/list", handleJobList)
    mux.HandleFunc("/job/kill", handleJobKill)

    // 前端页面处理
    // /index.html 这个路由匹配不上上面四个, 所以会匹配到下面的/, 通过StripPrefix把/去掉, 通过staticHandler, 前面加行路径
    staticDir = http.Dir(G_config.WebRoot)
    staticHandler = http.FileServer(staticDir)
    mux.Handle("/", http.StripPrefix("/", staticHandler)) // ./webroot/index.html

    // 创建HTTP服务
    httpServer = &http.Server{
        Handler:      mux,
        ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
        WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
    }

    // 赋值单例
    G_apiServer = &ApiServer{httpServer: httpServer}

    // 服务端拉起
    go httpServer.Serve(listener)

    return

}

func handleJobDelete(w http.ResponseWriter, r *http.Request) {
    var (
        err     error
        bytes   []byte
        jobName string
        oldJob  *common.Job
    )
    if err = r.ParseForm(); err != nil {
        goto ERR
    }

    jobName = r.PostForm.Get("name")

    //if err = json.Unmarshal([]byte(postJob), &job); err != nil {
    //	goto ERR
    //}

    if oldJob, err = G_jobMgr.DeleteJob(jobName); err != nil {
        goto ERR
    }

    if bytes, err = common.BuildResponse(0, "success", oldJob); err != nil {
        goto ERR
    }

    w.Write(bytes)

ERR:
    if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
        w.Write(bytes)
    }
}

// 保存Job
/*
* "job": {"name": "job", "command": "job", "cronExpr": "*"}
 */
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
        goto ERR
    }

    // 2, 取job字段
    postJob = r.PostForm.Get("job")

    // 3, 反序列化
    if err = json.Unmarshal([]byte(postJob), &job); err != nil {
        goto ERR
    }

    // 4, 保存
    if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
        goto ERR
    }

    // 5, 正常应答
    if bytes, err = common.BuildResponse(0, "success", oldJob); err != nil {
        goto ERR
    }
    w.Write(bytes)
    return

ERR:
    if bytes, err = common.BuildResponse(-1, err.Error(), nil); err != nil {
        w.Write(bytes)
    }

}

// 枚举job列表
func handleJobList(w http.ResponseWriter, r *http.Request) {
    var (
        jobList []*common.Job
        err     error
        bytes   []byte
    )
    if jobList, err = G_jobMgr.ListJob(); err != nil {
        goto ERR
    }

    if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
        w.Write(bytes)
    }

    return

ERR:
    if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
        w.Write(bytes)
    }

}

// 强制杀死任务
func handleJobKill(w http.ResponseWriter, r *http.Request) {
    var (
        err   error
        name  string
        bytes []byte
    )
    // 解析表单
    if err = r.ParseForm(); err != nil {
        goto ERR
    }

    // 任务名
    name = r.PostForm.Get("name")

    if err := G_jobMgr.KillJob(name); err != nil {
        goto ERR
    }

    // 正常应答
    if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
        w.Write(bytes)
    }
    return

ERR:
    if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
        w.Write(bytes)
    }
}
