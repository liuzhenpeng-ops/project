package main

import "fmt"
import "time"
import "sync"
import "sync/atomic"
import "math/rand"

func main() {
        var max uint64 = 5
        var min uint64 = 1
        obj := NewPool(max,min)

        /*go func(){
            time.Sleep(2*time.Second)
            atomic.StoreUint64(&obj.status,STOP)
            obj.PoolStop()
        }()*/

        go func(){
            k :=0
            for{
                if atomic.LoadUint64(&obj.status) == STOP{
                    close(obj.task)
                    return
                }

                task := Task{
                    f:func(str string){
                        fmt.Printf("打印:%+v\n",str)
                    },
                    params:randStr(5),
                }
                if k == 1000{
                    time.Sleep(1*time.Second)
                    return
                }
                obj.TaskPut(&task)
                k++
            }
        }()

        for{
            fmt.Printf("协程总数:%+v\n",atomic.LoadUint64(&obj.working))
            time.Sleep(time.Second*2)
        }
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func randStr(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }

    return string(b)
}

type signal struct{}

type Task struct{
    f func(str string)
    params string
}

type Pool struct{
    task chan *Task
    max_caps uint64
    min_caps uint64
    working uint64
    is_stop chan signal
    status uint64
    mutex *sync.Mutex
}

const (
    STOP = iota
    RUNING
)

func NewPool(max_caps,min_caps uint64) *Pool{
    obj := new(Pool)

    if min_caps > max_caps{
        fmt.Printf("最小容量不能大于最大容量\n")
    }

    if max_caps < 1{
        max_caps = 1
    }

    if min_caps < 1{
        min_caps = 1
    }

    (*obj).task = make(chan *Task,max_caps)
    (*obj).is_stop = make(chan signal)
    (*obj).working =0
    (*obj).max_caps = max_caps
    (*obj).min_caps = min_caps
    (*obj).status = RUNING
    (*obj).mutex = new(sync.Mutex)

    return obj
}

func (p *Pool) PoolStop(){
    atomic.StoreUint64(&p.status,STOP)
    p.is_stop <- signal{}
}

func (p *Pool) TaskPut(task *Task){
    defer p.mutex.Unlock()

    p.mutex.Lock()

   if atomic.LoadUint64(&p.status) == STOP{
       return
   }

    if atomic.LoadUint64(&p.working) < p.max_caps{
        p.Run()
    }

    p.task<- task
}

func (p *Pool) Run(){
    atomic.AddUint64(&p.working, 1)
    go func(){

        idle_timer := time.Millisecond*100
        idle := false

        timer := time.NewTimer(idle_timer)

        for{
            select{
                case <-p.is_stop:
                    fmt.Printf("收到退出信号,退出\n")
                    goto exit
                default:
            }

            select{
            case <-timer.C:
                if idle && atomic.LoadUint64(&p.working) > 0 && atomic.LoadUint64(&p.working) > p.min_caps{
                    goto exit
                }

                idle = true
                timer.Reset(idle_timer)
            case data,ok:=<-p.task:
                if !ok{
                    fmt.Printf("关闭,退出")
                    goto exit
                }

                idle = false

                data.f(data.params)
            }
        }

        exit:
            atomic.AddUint64(&p.working, ^uint64(0))
            return
    }()
}
