package main

import (
	"context"
	"ehids/user"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cilium/ebpf/rlimit"
)

func main() {

	// Allow the current process to lock memory for eBPF resources.
	if err := rlimit.RemoveMemlock(); err != nil {
		log.Fatal(err)
	}

	stopper := make(chan os.Signal, 1)
	signal.Notify(stopper, os.Interrupt, syscall.SIGTERM)
	ctx, cancelFun := context.WithCancel(context.TODO())

	logger := log.Default()
	logger.Println("https://github.com/ehids/ehids")
	logger.Printf("process pid: %d\n", os.Getpid())

	for k, module := range user.GetModules() {
		if module.Name() != "EBPFProbeBPFCall" {
			continue //模块启用临时开关
		}

		logger.Printf("start to run %s module", k)
		//初始化
		err := module.Init(ctx, logger)
		if err != nil {
			panic(err)
		}

		// 加载ebpf，挂载到hook点上，开始监听
		go func(module user.IModule) {
			err := module.Run()
			if err != nil {
				logger.Fatalf("%v", err)
			}
		}(module)
	}

	<-stopper
	cancelFun()

	logger.Println("Received signal, exiting program..")
	time.Sleep(time.Millisecond * 100)
}
