package main

import (
	. "github.com/gohutool/boot4go-xxljob-executor"
	"github.com/gohutool/log4go"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : example.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 15:29
* 修改历史 : 1. [2022/4/10 15:29] 创建文件 by NST
*/

func main() {
	log4go.LoggerManager.InitWithDefaultConfig()
	logger := log4go.LoggerManager.GetLogger("gohutool.executor4go.examples")

	exec := NewExecutor(
		ServerAddr("http://192.168.56.101:18080/xxl-job-admin"),
		AccessToken(""),            //请求令牌(默认为空)
		ExecutorIp("192.168.56.1"), //可自动获取
		ExecutorPort("9999"),       //默认9999（非必填）
		RegistryKey("executor4go"), //执行器名称
		SetLogger(logger),          //自定义日志
	)
	exec.Init()
	//设置日志查看handler
	exec.SetLogHandler(func(req *LogReq) *LogRes {
		return &LogRes{Code: 200, Msg: "", Content: LogResContent{
			FromLineNum: req.FromLineNum,
			ToLineNum:   2,
			LogContent:  "这个是自定义日志handler",
			IsEnd:       true,
		}}
	})
	//注册任务handler
	exec.RegTask("task.simple", SimpleTask)
	exec.RegTask("task.longTime", LongTimeTask)
	exec.RegTask("task.panic", PanicTask)
	exec.RegTask("task.exception", ExceptionTask)

	logger.Info(exec.Run())
}
