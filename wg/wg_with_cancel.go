/*
@Time : 24/07/2020 3:05 PM
@Author : zhangjun
@File : wg2
@Description:
@Run:
*/
package wg

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type WaitGroupWrapperWithCancel struct {
	wg       sync.WaitGroup
	cancel   func()
	ctx      context.Context
	funcName string
	stopCh   chan bool
}

//build new waitGroup object
func NewWaitGroupWrapperWithCancel(ctx context.Context, cancel func(), funcName string) *WaitGroupWrapperWithCancel {
	wgObject := &WaitGroupWrapperWithCancel{
		cancel:   cancel,
		ctx:      ctx,
		wg:       sync.WaitGroup{},
		funcName: funcName,
	}
	wgObject.stopCh = make(chan bool)
	return wgObject
}

func (this *WaitGroupWrapperWithCancel) Wrap(cb func()) {
	this.wg.Add(1)
	go func() {
		defer this.wg.Done()
		// 取消机制
		select {
		case <-this.ctx.Done():
			fmt.Println("WaitGroupWrapper timeout Wrap go cancel func is [", this.funcName, "]")
			return

		case stopCh := <-this.stopCh:
			fmt.Println("WaitGroupWrapper success Wrap go cancel func is [", this.funcName, stopCh, "]")
			return
		}

		f := func(cb func()) {
			cb()
			this.stopCh <- true
		}
		f(cb)

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
