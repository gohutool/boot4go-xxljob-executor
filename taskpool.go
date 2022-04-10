package executor4go

import "sync"

/**
* golang-sample源代码，版权归锦翰科技（深圳）有限公司所有。
* <p>
* 文件名称 : taskpool.go
* 文件路径 :
* 作者 : DavidLiu
× Email: david.liu@ginghan.com
*
* 创建日期 : 2022/4/10 12:46
* 修改历史 : 1. [2022/4/10 12:46] 创建文件 by NST
*/

//var TaskPool = &taskPool{data: make(map[string]*Task)}

// taskPool
//任务列表 [JobID]执行函数,并行执行时[+LogID]
type taskPool struct {
	mu   sync.RWMutex
	data map[string]*Task
}

// Set 设置数据
func (t *taskPool) Set(key string, val *Task) {
	t.mu.Lock()
	t.data[key] = val
	t.mu.Unlock()
}

// Get 获取数据
func (t *taskPool) Get(key string) *Task {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.data[key]
}

// GetAll 获取数据
func (t *taskPool) GetAll() map[string]*Task {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.data
}

// Del 设置数据
func (t *taskPool) Del(key string) {
	t.mu.Lock()
	delete(t.data, key)
	t.mu.Unlock()
}

// Len 长度
func (t *taskPool) Len() int {
	return len(t.data)
}

// Exists Key是否存在
func (t *taskPool) Exists(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.data[key]
	return ok
}
