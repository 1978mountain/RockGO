package main

import (
	"github.com/zllangct/rockgo/actor"
	"github.com/zllangct/rockgo/ecs"
	"sync"
)

type RoomManagerComponent struct {
	ecs.ComponentBase
	Actor.ActorBase
	locker     sync.RWMutex
	rooms      map[int]*RoomComponent
	increasing int //实际运用不这样,此处便宜行事
}

func (this *RoomManagerComponent) Initialize() error {
	//初始化actor
	this.ActorInit(this.Parent())
	return nil
}

func (this *RoomManagerComponent) Awake(ctx *ecs.Context) {
	//初始化房间
	this.rooms = make(map[int]*RoomComponent)
	//注册actor消息
	this.AddHandler(Service_RoomMgr_NewRoom, this.NewRoom, true)
}

var Service_RoomMgr_NewRoom = "NewRoom"

func (this *RoomManagerComponent) NewRoom(message *Actor.ActorMessageInfo) error {
	r := &RoomComponent{}
	_, err := this.Parent().AddNewbjectWithComponents([]ecs.IComponent{r})
	if err != nil {
		return err
	}

	this.locker.Lock()
	this.increasing++
	r.RoomID = this.increasing
	this.rooms[r.RoomID] = r
	this.locker.Unlock()

	return message.Reply(r.RoomID)
}
