package rpc

import (
	"errors"
	"sync"
	"time"
)

// 进程级全局服务发现：Init 创建一次，之后 NewClient("grpc://svc") 无需再手动创建 Discovery。
// grpc:// scheme 在进程内只注册一次（resolverOnce），因此全局实例在整个进程生命周期内持有，
// Close 之后不可再次 Init（重新 Init 也无法更换 scheme 归属）。
var (
	initMu       sync.Mutex
	globalDisc   *Discovery
	globalErr    error
	globalClosed bool
)

// Init 初始化进程级服务发现，之后 NewClient("grpc://svc") 直接可用（自动发现 + 故障转移）。
// 幂等：多次调用以第一次为准，后续调用直接返回首次结果（包括首次的错误）。
// 典型的消费方启动代码：
//
//	if err := rpc.Init([]string{"http://127.0.0.1:2379"}, 5*time.Second); err != nil {
//	    log.Fatal(err)
//	}
//	c, _ := rpc.NewClient("grpc://greeter")
func Init(endpoints []string, dialTimeout time.Duration) error {
	initMu.Lock()
	defer initMu.Unlock()
	if globalClosed {
		return errors.New("rpc: global discovery closed; cannot re-init in the same process")
	}
	if globalDisc != nil || globalErr != nil {
		return globalErr
	}
	globalDisc, globalErr = NewDiscovery(endpoints, dialTimeout)
	return globalErr
}

// Close 关闭 Init 创建的全局服务发现（幂等）。
// 注意：grpc:// scheme 归属不可回退，Close 后进程内不能再 Init/拨号 grpc:// 目标，
// 因此仅应在进程退出前调用。
func Close() error {
	initMu.Lock()
	defer initMu.Unlock()
	globalClosed = true
	if globalDisc == nil {
		return nil
	}
	err := globalDisc.Close()
	globalDisc = nil
	globalErr = nil
	return err
}
