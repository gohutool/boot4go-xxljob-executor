package boot4go_xxljob_executor

import (
	"encoding/json"
	. "github.com/gohutool/boot4go-xxljob-executor/utils"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : restfulExecutorImpl.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 13:05
* 修改历史 : 1. [2022/4/10 13:05] 创建文件 by NST
*/

type RestFulExecutor struct {
	Executor
	mu      sync.RWMutex
	handler LogHandler
}

func (e *RestFulExecutor) Run() (err error) {
	// 创建路由器
	mux := http.NewServeMux()
	// 设置路由规则
	mux.HandleFunc("/run", e.runTask)
	mux.HandleFunc("/kill", e.killTask)
	mux.HandleFunc("/log", e.taskLog)
	mux.HandleFunc("/beat", e.beat)
	mux.HandleFunc("/idleBeat", e.idleBeat)
	// 创建服务器
	server := &http.Server{
		Addr:         e.address,
		WriteTimeout: time.Second * 3,
		Handler:      mux,
	}
	// 监听端口并提供服务
	e.log.Info("Starting server at " + e.address)
	go server.ListenAndServe()
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	e.UnRegistry()
	return nil
}

func (e *RestFulExecutor) Stop() {
	e.UnRegistry()
}

// Registry
// 注册执行器到调度中心
func (e *RestFulExecutor) Registry() {

	t := time.NewTimer(time.Second * 0) //初始立即执行
	defer t.Stop()
	req := &Registry{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   e.opts.RegistryKey,
		RegistryValue: "http://" + e.address,
	}
	param, err := json.Marshal(req)
	if err != nil {
		e.log.Critical("执行器注册信息解析失败:" + err.Error())
	}
	for {
		<-t.C
		t.Reset(time.Second * time.Duration(20)) //20秒心跳防止过期
		func() {
			result, err := e.post("/api/registry", string(param))
			if err != nil {
				e.log.Error("执行器注册失败1:" + err.Error())
				return
			}
			defer result.Body.Close()
			body, err := ioutil.ReadAll(result.Body)
			if err != nil {
				e.log.Error("执行器注册失败2:" + err.Error())
				return
			}
			res := &res{}
			_ = json.Unmarshal(body, &res)
			if res.Code != 200 {
				e.log.Error("执行器注册失败3:" + string(body))
				return
			}
			e.log.Info("执行器注册成功:" + string(body))
		}()

	}
}

// UnRegistry
// 从调度中心在注销执行器
func (e *RestFulExecutor) UnRegistry() {
	t := time.NewTimer(time.Second * 0) //初始立即执行
	defer t.Stop()
	req := &Registry{
		RegistryGroup: "EXECUTOR",
		RegistryKey:   e.opts.RegistryKey,
		RegistryValue: "http://" + e.address,
	}
	param, err := json.Marshal(req)
	if err != nil {
		e.log.Error("执行器摘除失败:" + err.Error())
		return
	}
	res, err := e.post("/api/registryRemove", string(param))
	if err != nil {
		e.log.Error("执行器摘除失败:" + err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	e.log.Info("执行器摘除成功:" + string(body))
	_ = res.Body.Close()
}

// RequestCallback
// 回调任务列表
func (e *RestFulExecutor) RequestCallback(task *Task, code int64, msg string) {
	e.runList.Del(Int64ToStr(task.Id))
	res, err := e.post("/api/callback", string(returnCall(task.Param, code, msg)))
	if err != nil {
		e.log.Error("callback err : ", err.Error())
		return
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		e.log.Error("callback ReadAll err : ", err.Error())
		return
	}
	e.log.Info("任务回调成功:" + string(body))
}

/*****************  Primate method to correspond each command from admin center  *********************/

//post
func (e *RestFulExecutor) post(action, body string) (resp *http.Response, err error) {
	request, err := http.NewRequest("POST", e.opts.ServerAddr+action, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	request.Header.Set("XXL-JOB-ACCESS-TOKEN", e.opts.AccessToken)
	client := http.Client{
		Timeout: e.opts.Timeout,
	}
	return client.Do(request)
}

// runTask to route uri /run
// 运行一个任务
func (e *RestFulExecutor) runTask(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &RunReq{}
	err := json.Unmarshal(req, &param)
	if err != nil {
		_, _ = writer.Write(returnCall(param, 500, "params err"))
		e.log.Error("参数解析错误:" + string(req))
		return
	}
	e.log.Info("任务参数:%v", param)
	if !e.regList.Exists(param.ExecutorHandler) {
		_, _ = writer.Write(returnCall(param, 500, "Task not registered"))
		e.log.Error("任务[" + Int64ToStr(param.JobID) + "]没有注册:" + param.ExecutorHandler)
		return
	}

	e.RunTask(param, func(task *Task, param *RunReq) {
		_, _ = writer.Write(returnCall(param, 500, "There are tasks running"))
	}, func(task *Task, param *RunReq) {
		_, _ = writer.Write(returnGeneral())
	})
}

// killTask to route path /kill
// 删除一个任务
func (e *RestFulExecutor) killTask(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &killReq{}
	_ = json.Unmarshal(req, &param)
	if !e.runList.Exists(Int64ToStr(param.JobID)) {
		_, _ = writer.Write(returnKill(param, 500))
		e.log.Error("任务[" + Int64ToStr(param.JobID) + "]没有运行")
		return
	}
	e.CancelTask(param)
	_, _ = writer.Write(returnGeneral())
}

// taskLog to route path /log
// 任务日志
func (e *RestFulExecutor) taskLog(writer http.ResponseWriter, request *http.Request) {
	var res *LogRes
	data, err := ioutil.ReadAll(request.Body)
	req := &LogReq{}
	if err != nil {
		e.log.Error("日志请求失败:" + err.Error())
		reqErrLogHandler(writer, req, err)
		return
	}
	err = json.Unmarshal(data, &req)
	if err != nil {
		e.log.Error("日志请求解析失败:" + err.Error())
		reqErrLogHandler(writer, req, err)
		return
	}
	e.log.Info("日志请求参数:%+v", req)
	if e.logHandler != nil {
		res = e.logHandler(req)
	} else {
		res = defaultLogHandler(req)
	}
	str, _ := json.Marshal(res)
	_, _ = writer.Write(str)
}

// beat to route path /beat
// 心跳检测
func (e *RestFulExecutor) beat(writer http.ResponseWriter, request *http.Request) {
	e.log.Info("心跳检测")
	_, _ = writer.Write(returnGeneral())
}

// idleBeat to route path /idleBeat
// 忙碌检测
func (e *RestFulExecutor) idleBeat(writer http.ResponseWriter, request *http.Request) {
	e.mu.Lock()
	defer e.mu.Unlock()
	req, _ := ioutil.ReadAll(request.Body)
	param := &idleBeatReq{}
	err := json.Unmarshal(req, &param)
	if err != nil {
		_, _ = writer.Write(returnIdleBeat(500))
		e.log.Error("参数解析错误:" + string(req))
		return
	}
	if e.runList.Exists(Int64ToStr(param.JobID)) {
		_, _ = writer.Write(returnIdleBeat(500))
		e.log.Error("idleBeat任务[" + Int64ToStr(param.JobID) + "]正在运行")
		return
	}
	e.log.Info("忙碌检测任务参数:%v", param)
	_, _ = writer.Write(returnGeneral())
}

/*****************  Logging  *********************/

/**
用来日志查询，显示到xxl-job-admin后台
*/

//默认返回
func defaultLogHandler(req *LogReq) *LogRes {
	return &LogRes{Code: 200, Msg: "", Content: LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   2,
		LogContent:  "这是日志默认返回，说明没有设置LogHandler",
		IsEnd:       true,
	}}
}

//请求错误
func reqErrLogHandler(w http.ResponseWriter, req *LogReq, err error) {
	res := &LogRes{Code: 500, Msg: err.Error(), Content: LogResContent{
		FromLineNum: req.FromLineNum,
		ToLineNum:   0,
		LogContent:  err.Error(),
		IsEnd:       true,
	}}
	str, _ := json.Marshal(res)
	_, _ = w.Write(str)
}
