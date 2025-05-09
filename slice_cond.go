package main

import (
    "fmt"
    "sync"
    "time"
    "sync/atomic"
)

var wg sync.WaitGroup

type signal struct{}

var stop_status int32 = 1

var consumer_count = 3

type slice_list struct {
    status int32
    lock *sync.Mutex
    cond *sync.Cond
    data []int
}

func new_init() *slice_list{
    list := &slice_list{status:0}

    (*list).lock=new(sync.Mutex)
    (*list).data=make([]int,0)
    (*list).cond=sync.NewCond((*list).lock)

    return list
}

func (list *slice_list) add(item int){

    for atomic.LoadInt32(&list.status) == stop_status{
        fmt.Printf("已停止\n")
        return
    }

    list.data = append(list.data,item)

    list.cond.Signal()
}

func (list *slice_list) list(number int,wg *sync.WaitGroup){
    defer func(){
        wg.Done()
        list.lock.Unlock()
    }()

    for{

        list.lock.Lock()

        if atomic.LoadInt32(&list.status) == stop_status{
            for len(list.data) ==0 {
                return
            }
        }

        fmt.Printf("协程:%+v执行\n",number)

        for len(list.data) ==0 {
            fmt.Printf("协程%d没有元素阻塞\n",number)
            list.cond.Wait()
        }

        fmt.Printf("切片列表:%+v\n",list.data)
        if len(list.data) > 0{
            data := list.data[0]
            list.data = list.data[1:]
            fmt.Printf("协程:%+v打印元素:%+v\n",number,data)
        }

        list.lock.Unlock()
    }
}

func (list *slice_list) stop(wg *sync.WaitGroup){
    atomic.StoreInt32(&(list).status, stop_status)

    wg.Wait()
    fmt.Printf("全部结束\n")
}

func main() {

    list := new_init()

    for i:=100;i < 120;i++{
        list.add(i)
    }

    wg.Add(consumer_count)
    for i:=0;i<consumer_count;i++{
        go list.list(i+1,&wg)
    }

    //list.stop(&wg)

    for{
        time.Sleep(time.Second*10)
    }
}
