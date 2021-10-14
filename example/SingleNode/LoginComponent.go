package main

import (
	"errors"
	"github.com/zllangct/rockgo/cluster"
	"github.com/zllangct/rockgo/config"
	"github.com/zllangct/rockgo/ecs"
	"reflect"
	"sync"
	"time"
)

type LoginComponent struct {
	ecs.ComponentBase
	locker        sync.RWMutex
	nodeComponent *Cluster.NodeComponent
	players       sync.Map // [account,*PlayerInfo]
}

func (this *LoginComponent) GetRequire() map[*ecs.Object][]reflect.Type {
	requires := make(map[*ecs.Object][]reflect.Type)
	requires[this.Root()] = []reflect.Type{
		reflect.TypeOf(&config.ConfigComponent{}),
	}
	return requires
}

func (this *LoginComponent) Awake(ctx *ecs.Context) {
	err := this.Root().Find(&this.nodeComponent)
	if err != nil {
		panic(err)
	}
	//模拟已存在的用户
	this.players.Store("zllang1", &PlayerInfo{Account: "zllang1", Password: "123456", Name: "zhaolei1", Age: 11, Coin: 100, LastLoginTime: time.Now()})
	this.players.Store("zllang2", &PlayerInfo{Account: "zllang2", Password: "123456", Name: "zhaolei2", Age: 12, Coin: 200, LastLoginTime: time.Now()})
	this.players.Store("zllang3", &PlayerInfo{Account: "zllang3", Password: "123456", Name: "zhaolei3", Age: 13, Coin: 300, LastLoginTime: time.Now()})
	this.players.Store("zllang4", &PlayerInfo{Account: "zllang4", Password: "123456", Name: "zhaolei4", Age: 14, Coin: 400, LastLoginTime: time.Now()})
	this.players.Store("zllang5", &PlayerInfo{Account: "zllang5", Password: "123456", Name: "zhaolei5", Age: 15, Coin: 500, LastLoginTime: time.Now()})

	err = this.nodeComponent.Register(this)
	if err != nil {
		panic(err)
	}
}

var ErrLoginPlayerNotExist = errors.New("this player doesnt exist")

func (this *LoginComponent) Login(account string, reply *PlayerInfo) error {
	if p, ok := this.players.Load(account); ok {
		*reply = *(p.(*PlayerInfo))
		return nil
	} else {
		return ErrLoginPlayerNotExist
	}
}
