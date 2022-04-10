package executor4go

import (
	"context"
	"encoding/json"
	"github.com/go-basic/ipv4"
	. "github.com/gohutool/boot4go-xxljob-executor/utils"
	"github.com/gohutool/log4go"
	"time"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : executor4go.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 11:24
* 修改历史 : 1. [2022/4/10 11:24] 创建文件 by NST
*/

const (
	EXECUTOR4G_VERSION = "executor4go-v1.0.0"
	EXECUTOR4G_MAJOR   = 1
	EXECUTOR4G_MINOR   = 0
	EXECUTOR4G_BUILD   = 0
)

/*****************  executor Engine  *********************/

// executor 执行器
type executor interface {
	// Init
	// 初始化执行器
	Init(opts ...Option)
	// Run 启动服务
	Run() error
	// Registry 注册执行器到调度中心
	Registry()
	// UnRegistry 从调度中心在注销执行器
	UnRegistry() error
	// SetLogHandler 日志handler
	SetLogHandler(handler LogHandler)
	// Stop 停止服务
	Stop()
	// RequestCallback 执行后回调请求给调度中心
	RequestCallback(task *Task, code int64, msg string)
	// RegTask
	// 注册任务
	RegTask(pattern string, task TaskFunc)
}

// NewExecutor 创建执行器
func NewExecutor(opts ...Option) executor {
	return newExecutor(opts...)
}

func newExecutor(opts ...Option) executor {
	//var exec executor

	options := newOptions(opts...)

	if options.executorImpl == nil {
		exec := &RestFulExecutor{}
		exec.opts = options
		return exec
	} else {
		return options.executorImpl
	}
}

type Executor struct {
	opts    Options
	address string
	regList *taskPool //注册任务列表
	runList *taskPool //正在执行任务列表
	log     log4go.Logger

	logHandler LogHandler //日志查询handler
}

func (e *Executor) Init(opts ...Option) {
	for _, o := range opts {
		o(&e.opts)
	}

	e.log = e.opts.logger

	e.regList = &taskPool{
		data: make(map[string]*Task),
	}
	e.runList = &taskPool{
		data: make(map[string]*Task),
	}

	e.address = e.opts.ExecutorIp + ":" + e.opts.ExecutorPort

	e.log.Info("%v %v", EXECUTOR4G_VERSION, e.address)

	go e.Registry()
}

func (e *Executor) SetLogHandler(logHandler LogHandler) {
	e.logHandler = logHandler
}

// RegTask 注册任务
func (e *Executor) RegTask(pattern string, task TaskFunc) {
	var t = &Task{}
	t.fn = task
	e.regList.Set(pattern, t)
	return
}

func (e *Executor) CancelTask(param *killReq) {
	task := e.runList.Get(Int64ToStr(param.JobID))
	task.Cancel()
	e.runList.Del(Int64ToStr(param.JobID))
}

func (e *Executor) RunTask(param *RunReq,
	onRunning func(task *Task, param *RunReq),
	onSuccess func(task *Task, param *RunReq)) (*Task, error) {

	//阻塞策略处理
	if e.runList.Exists(Int64ToStr(param.JobID)) {
		if param.ExecutorBlockStrategy == coverEarly { //覆盖之前调度
			oldTask := e.runList.Get(Int64ToStr(param.JobID))
			if oldTask != nil {
				oldTask.Cancel()
				e.runList.Del(Int64ToStr(oldTask.Id))
			}
		} else { //单机串行,丢弃后续调度 都进行阻塞
			oldTask := e.runList.Get(Int64ToStr(param.JobID))
			onRunning(oldTask, param)
			return oldTask, e.log.Error("任务[" + Int64ToStr(param.JobID) + "]已经在运行了:" + param.ExecutorHandler)
		}
	}

	cxt := context.Background()
	task := e.regList.Get(param.ExecutorHandler)
	if param.ExecutorTimeout > 0 {
		task.Ext, task.Cancel = context.WithTimeout(cxt, time.Duration(param.ExecutorTimeout)*time.Second)
	} else {
		task.Ext, task.Cancel = context.WithCancel(cxt)
	}
	task.Id = param.JobID
	task.Name = param.ExecutorHandler
	task.Param = param
	task.log = e.log

	e.runList.Set(Int64ToStr(task.Id), task)
	go task.Run(func(code int64, msg string) {
		e.RequestCallback(task, code, msg)
	})
	e.log.Info("任务[" + Int64ToStr(param.JobID) + "]开始执行:" + param.ExecutorHandler)
	onSuccess(task, param)
	return task, nil
}

/*****************  interface methods, will be implemented by One Executor's sub class  *********************/

// Run 运行执行器引擎
/**
在执行器引擎里必须处理来至于调度中心的一下命令
run
kill
log
beat
idleBeat
*/
func (e *Executor) Run() error {
	return nil
}

func (e *Executor) RequestCallback(task *Task, code int64, msg string) {

}

// Stop 停止服务
func (e *Executor) Stop() {
	e.UnRegistry()
}

// Registry 注册执行器到调度中心
func (e *Executor) Registry() {
}

// UnRegistry 从调度中心在注销执行器
func (e *Executor) UnRegistry() error {
	return nil
}

//通用响应
type res struct {
	Code int64       `json:"code"` // 200 表示正常、其他失败
	Msg  interface{} `json:"msg"`  // 错误提示消息
}

/*****************  request 参数  *********************/

// Registry 注册参数
type Registry struct {
	RegistryGroup string `json:"registryGroup"`
	RegistryKey   string `json:"registryKey"`
	RegistryValue string `json:"registryValue"`
}

//执行器执行完任务后，回调任务结果时使用
type call []*callElement

type callElement struct {
	LogID         int64          `json:"logId"`
	LogDateTim    int64          `json:"logDateTim"`
	ExecuteResult *ExecuteResult `json:"executeResult"`
	//以下是7.31版本 v2.3.0 Release所使用的字段
	HandleCode int    `json:"handleCode"` //200表示正常,500表示失败
	HandleMsg  string `json:"handleMsg"`
}

// ExecuteResult 任务执行结果 200 表示任务执行正常，500表示失败
type ExecuteResult struct {
	Code int64       `json:"code"`
	Msg  interface{} `json:"msg"`
}

/*****************  command 参数  *********************/

//阻塞处理策略
const (
	serialExecution = "SERIAL_EXECUTION" //单机串行
	discardLater    = "DISCARD_LATER"    //丢弃后续调度
	coverEarly      = "COVER_EARLY"      //覆盖之前调度
)

// RunReq 触发任务请求参数
type RunReq struct {
	JobID                 int64  `json:"jobId"`                 // 任务ID
	ExecutorHandler       string `json:"executorHandler"`       // 任务标识
	ExecutorParams        string `json:"executorParams"`        // 任务参数
	ExecutorBlockStrategy string `json:"executorBlockStrategy"` // 任务阻塞策略
	ExecutorTimeout       int64  `json:"executorTimeout"`       // 任务超时时间，单位秒，大于零时生效
	LogID                 int64  `json:"logId"`                 // 本次调度日志ID
	LogDateTime           int64  `json:"logDateTime"`           // 本次调度日志时间
	GlueType              string `json:"glueType"`              // 任务模式，可选值参考 com.xxl.job.core.glue.GlueTypeEnum
	GlueSource            string `json:"glueSource"`            // GLUE脚本代码
	GlueUpdatetime        int64  `json:"glueUpdatetime"`        // GLUE脚本更新时间，用于判定脚本是否变更以及是否需要刷新
	BroadcastIndex        int64  `json:"broadcastIndex"`        // 分片参数：当前分片
	BroadcastTotal        int64  `json:"broadcastTotal"`        // 分片参数：总分片
}

//终止任务请求参数
type killReq struct {
	JobID int64 `json:"jobId"` // 任务ID
}

//忙碌检测请求参数
type idleBeatReq struct {
	JobID int64 `json:"jobId"` // 任务ID
}

// LogReq 日志请求
type LogReq struct {
	LogDateTim  int64 `json:"logDateTim"`  // 本次调度日志时间
	LogID       int64 `json:"logId"`       // 本次调度日志ID
	FromLineNum int   `json:"fromLineNum"` // 日志开始行号，滚动加载日志
}

// LogRes 日志响应
type LogRes struct {
	Code    int64         `json:"code"`    // 200 表示正常、其他失败
	Msg     string        `json:"msg"`     // 错误提示消息
	Content LogResContent `json:"content"` // 日志响应内容
}

// LogResContent 日志响应内容
type LogResContent struct {
	FromLineNum int    `json:"fromLineNum"` // 本次请求，日志开始行数
	ToLineNum   int    `json:"toLineNum"`   // 本次请求，日志结束行号
	LogContent  string `json:"logContent"`  // 本次请求日志内容
	IsEnd       bool   `json:"isEnd"`       // 日志是否全部加载完
}

// LogHandler jobLogger处理器
type LogHandler func(req *LogReq) *LogRes

/*****************  Command response  *********************/

//执行任务回调
func returnCallBytes(req *RunReq, code int64, msg string) []byte {
	data := call{
		&callElement{
			LogID:      req.LogID,
			LogDateTim: req.LogDateTime,
			ExecuteResult: &ExecuteResult{
				Code: code,
				Msg:  msg,
			},
			HandleCode: int(code),
			HandleMsg:  msg,
		},
	}
	str, _ := json.Marshal(data)
	return str
}

//杀死任务返回
func returnKillBytes(req *killReq, code int64) []byte {
	msg := ""
	if code != 200 {
		msg = "Task kill err"
	}
	data := res{
		Code: code,
		Msg:  msg,
	}
	str, _ := json.Marshal(data)
	return str
}

//忙碌返回
func returnIdleBeatBytes(code int64) []byte {
	msg := ""
	if code != 200 {
		msg = "Task is busy"
	}
	data := res{
		Code: code,
		Msg:  msg,
	}
	str, _ := json.Marshal(data)
	return str
}

//通用返回
func returnGeneralBytes() []byte {
	data := &res{
		Code: 200,
		Msg:  "",
	}
	str, _ := json.Marshal(data)
	return str
}

/*****************  Option  *********************/

type Options struct {
	ServerAddr   string        `json:"server_addr"`   //调度中心地址
	AccessToken  string        `json:"access_token"`  //请求令牌
	Timeout      time.Duration `json:"timeout"`       //接口超时时间
	ExecutorIp   string        `json:"executor_ip"`   //本地(执行器)IP(可自行获取)
	ExecutorPort string        `json:"executor_port"` //本地(执行器)端口
	RegistryKey  string        `json:"registry_key"`  //执行器名称
	LogDir       string        `json:"log_dir"`       //日志目录

	logger       log4go.Logger //日志
	executorImpl executor
}

func newOptions(opts ...Option) Options {
	opt := Options{
		ExecutorIp:   ipv4.LocalIP(),
		ExecutorPort: DefaultExecutorPort,
		RegistryKey:  DefaultRegistryKey,
	}

	for _, o := range opts {
		o(&opt)
	}

	if &opt.logger == nil {
		opt.logger = log4go.LoggerManager.GetLogger("gohutool.executor4go")
	}

	return opt
}

type Option func(o *Options)

var (
	DefaultExecutorPort = "9999"
	DefaultRegistryKey  = "golang-jobs"
)

// ServerAddr 设置调度中心地址
func ServerAddr(addr string) Option {
	return func(o *Options) {
		o.ServerAddr = addr
	}
}

// AccessToken 请求令牌
func AccessToken(token string) Option {
	return func(o *Options) {
		o.AccessToken = token
	}
}

// ExecutorIp 设置执行器IP
func ExecutorIp(ip string) Option {
	return func(o *Options) {
		o.ExecutorIp = ip
	}
}

// ExecutorPort 设置执行器端口
func ExecutorPort(port string) Option {
	return func(o *Options) {
		o.ExecutorPort = port
	}
}

// RegistryKey 设置执行器标识
func RegistryKey(registryKey string) Option {
	return func(o *Options) {
		o.RegistryKey = registryKey
	}
}

// SetLogger 设置日志处理器
func SetLogger(l log4go.Logger) Option {
	return func(o *Options) {
		o.logger = l
	}
}
