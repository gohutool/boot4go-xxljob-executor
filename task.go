package executor4go

import (
	"context"
	"fmt"
	. "github.com/gohutool/boot4go-xxljob-executor/utils"
	"github.com/gohutool/log4go"
)

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : task.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 12:35
* 修改历史 : 1. [2022/4/10 12:35] 创建文件 by NST
*/

// Task 任务
type Task struct {
	Id        int64
	Name      string
	Ext       context.Context
	Param     *RunReq
	fn        TaskFunc
	Cancel    context.CancelFunc
	StartTime int64
	EndTime   int64
	//日志
	log log4go.Logger
}

// TaskFunc 任务执行函数
type TaskFunc func(cxt context.Context, param *RunReq) (string, error)

// Run 运行任务
func (t *Task) Run(callback func(code int64, msg string)) {
	// cancel 实现长时间任务的Kill功能， 在长任务的Task中要配合cxt.Done():进行强行中断
	// 可参考/examples/task/long_task.go
	defer func(cancel func()) {
		if err := recover(); err != nil {
			t.log.Error(t.Info()+" panic: %v", err)
			//debug.PrintStack() //堆栈跟踪
			callback(500, "task panic:"+fmt.Sprintf("%v", err))
			cancel()
		}
	}(t.Cancel)
	msg, err := t.fn(t.Ext, t.Param)
	if err == nil {
		callback(200, msg)
	} else {
		callback(500, "task exception:"+fmt.Sprintf("%v", err))
	}
}

// Info 任务信息
func (t *Task) Info() string {
	return "任务ID[" + Int64ToStr(t.Id) + "]任务名称[" + t.Name + "]参数：" + t.Param.ExecutorParams
}
