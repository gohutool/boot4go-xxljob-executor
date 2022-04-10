# boot4go-xxljob-executor
golang for xxl-job executor support

很多公司java与go开发共存，java中有xxl-job做为任务调度引擎，为此也出现了go执行器(客户端)，使用起来比较简单：
# 支持
```	
1.执行器注册
2.耗时任务取消
3.任务注册，像写http.Handler一样方便
4.任务panic处理
5.阻塞策略处理
6.任务完成支持返回执行备注
7.任务超时取消 (单位：秒，0为不限制)
8.失败重试次数(在参数param中，目前由任务自行处理)
```

# Example
```
package main

import (
	"fmt"
	xxl "github.com/xxl-job/xxl-job-executor-go"
	"github.com/xxl-job/xxl-job-executor-go/example/task"
	"log"
)

func main() {
	exec := executor4go.NewExecutor(
		executor4go.ServerAddr("http://127.0.0.1/xxl-job-admin"),
		executor4go.AccessToken(""),            //请求令牌(默认为空)
		executor4go.ExecutorIp("127.0.0.1"),    //可自动获取
		executor4go.ExecutorPort("9999"),       //默认9999（非必填）
		executor4go.RegistryKey("golang-jobs"), //执行器名称
		executor4go.SetLogger(&logger{}),       //自定义日志
	)
	executor4go.Init()
	//设置日志查看handler
	executor4go.LogHandler(func(req *executor4go.LogReq) *executor4go.LogRes {
		return &executor4go.LogRes{Code: 200, Msg: "", Content: executor4go.LogResContent{
			FromLineNum: req.FromLineNum,
			ToLineNum:   2,
			LogContent:  "这个是自定义日志handler",
			IsEnd:       true,
		}}
	})
	//注册任务handler
	executor4go.RegTask("task.test", task.Test)
	executor4go.RegTask("task.test2", task.Test2)
	executor4go.RegTask("task.panic", task.Panic)
	log.Fatal(executor4go.Run())
}

//Logger接口实现
type logger struct{}

func (l *logger) Info(format string, a ...interface{}) {
	fmt.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}

func (l *logger) Error(format string, a ...interface{}) {
	log.Println(fmt.Sprintf("自定义日志 - "+format, a...))
}
```
# 示例项目
github.com/gohutool/boot4go-xxljob-executor/examples/

# xxl-job-admin配置
### 添加执行器
执行器管理->新增执行器,执行器列表如下：
```
AppName		名称		注册方式	OnLine 		机器地址 		操作
golang-jobs	golang执行器	自动注册 		查看 ( 1 ）   
```
查看->注册节点
```
http://127.0.0.1:9999
```
### 添加任务
任务管理->新增(注意，使用BEAN模式，JobHandler与RegTask名称一致)
```
1	测试panic	BEAN：task.panic	* 0 * * * ?	admin	STOP	
2	测试耗时任务	BEAN：task.test2	* * * * * ?	admin	STOP	
3	测试golang	BEAN：task.test		* * * * * ?	admin	STOP
```

