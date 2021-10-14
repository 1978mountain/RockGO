package Cluster

import (
	"fmt"
	"github.com/zllangct/rockgo/config"
	"github.com/zllangct/rockgo/ecs"
	"github.com/zllangct/rockgo/logger"
	"github.com/zllangct/rockgo/rpc"
	"github.com/zllangct/rockgo/utils"
	"reflect"
	"strings"
	"sync"
	"time"
)

type ChildComponent struct {
	ecs.ComponentBase
	locker          sync.RWMutex
	localAddr       string
	rpcMaster       *rpc.TcpClient //master节点
	nodeComponent   *NodeComponent
	reportCollecter []func() (string, float32)
	close           bool
}

func (this *ChildComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Parent().Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}),
		reflect.TypeOf(&NodeComponent{}),
	}
	return requires
}

func (this *ChildComponent) Awake(ctx *ecs.Context) {
	err := this.Parent().Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}

	go this.ConnectToMaster()
	go this.DoReport()
}

func (this *ChildComponent) Destroy(ctx *ecs.Context) {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.close = true
	this.ReportClose(this.localAddr)
}

//上报节点信息
func (this *ChildComponent) DoReport() {
	utils.When(time.Millisecond*50,
		func() bool {
			this.locker.RLock()
			defer this.locker.RUnlock()

			return this.rpcMaster != nil
		},
		func() bool {
			this.locker.RLock()
			defer this.locker.RUnlock()
			return this.localAddr != ""
		})
	args := &NodeInfo{
		Address: this.localAddr,
		Role:    config.Config.ClusterConfig.Role,
		AppName: config.Config.ClusterConfig.AppName,
	}
	var reply bool
	var interval = time.Duration(config.Config.ClusterConfig.ReportInterval)
	for {
		reply = false
		this.locker.RLock()
		if this.close {
			this.locker.RUnlock()
			return
		}
		m := make(map[string]float32)
		for _, collector := range this.reportCollecter {
			f, d := collector()
			m[f] = d
		}
		args.Info = m
		this.locker.RUnlock()
		if this.rpcMaster != nil {
			err := this.rpcMaster.Call("MasterService.ReportNodeInfo", args, &reply)
			if err != nil {

			}
		}
		time.Sleep(time.Millisecond * interval)
	}
}

//增加上报信息
func (this *ChildComponent) AddReportInfo(field string, collectFunction func() (string, float32)) {
	this.locker.Lock()
	this.reportCollecter = append(this.reportCollecter, collectFunction)
	this.locker.Unlock()
}

//增加上报节点关闭
func (this *ChildComponent) ReportClose(addr string) {
	var reply bool
	if this.rpcMaster != nil {
		_ = this.rpcMaster.Call("MasterService.ReportNodeClose", addr, &reply)
	}
}

//连接到master
func (this *ChildComponent) ConnectToMaster() {
	addr := config.Config.ClusterConfig.MasterAddress
	callback := func(event string, data ...interface{}) {
		switch event {
		case "close":
			this.OnDropped()
		}
	}
	logger.Info(" Looking for master ......")
	var err error
	for {
		this.locker.Lock()
		this.rpcMaster, err = this.nodeComponent.ConnectToNode(addr, callback)
		if err == nil {
			ip := strings.Split(this.rpcMaster.LocalAddr(), ":")[0]
			port := strings.Split(config.Config.ClusterConfig.LocalAddress, ":")[1]
			this.localAddr = fmt.Sprintf("%s:%s", ip, port)
			this.locker.Unlock()
			break
		}
		this.locker.Unlock()
		time.Sleep(time.Millisecond * 500)
	}

	this.nodeComponent.Locker().Lock()
	this.nodeComponent.isOnline = true
	this.nodeComponent.Locker().Unlock()

	logger.Info(fmt.Sprintf("Connected to master [ %s ]", addr))
}

//当节点掉线
func (this *ChildComponent) OnDropped() {
	//重新连接 time.Now().Format("2006-01-02T 15:04:05")
	this.locker.RLock()
	if this.close {
		this.locker.RUnlock()
		return
	}
	this.locker.RUnlock()
	this.nodeComponent.Locker().Lock()
	this.nodeComponent.isOnline = false
	this.nodeComponent.Locker().Unlock()

	logger.Info("Disconnected from master")
	this.ConnectToMaster()
}
