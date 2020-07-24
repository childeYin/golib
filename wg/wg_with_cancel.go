/*
@Time : 24/07/2020 3:05 PM
@Author : zhangjun
@File : wg2
@Description:
@Run:
*/
package wg

import (
	"fmt"
	"sync"
	"time"
	"context"
)

type WaitGroupWrapperWithCancel struct {
	wg sync.WaitGroup
	cancel func()
	ctx context.Context
	funcName string
}

//build new waitGroup object
func NewWaitGroupWrapper(ctx context.Context, cancel func(), funcName string ) *WaitGroupWrapperWithCancel{
	return &WaitGroupWrapperWithCancel{
		cancel:cancel,
		ctx:ctx,
		wg:sync.WaitGroup{},
		funcName:funcName,
	}
}

func (this *WaitGroupWrapperWithCancel) Wrap(cb func()) {
	this.wg.Add(1)
	go func() {
		cb()
		this.wg.Done()
		// 取消机制
		select {
		case <-this.ctx.Done():
			fmt.Printf("go cancel %s", this.funcName)
		}
	}()
}

func (this *WaitGroupWrapperWithCancel) Wait() {
	this.wg.Wait()
}

// WaitTimeout is same as Wait except that it accepts timeout arguement.
// FIXME if timeout triggered, there will be goroutine leak.
func (this *WaitGroupWrapperWithCancel) WaitTimeout(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		this.wg.Wait()
	}()

	select {
	case <-c:
		return false // completed normally

	case <-time.After(timeout):
		//超时调用cancel
		this.cancel()
		return true // timed out
	}
}
