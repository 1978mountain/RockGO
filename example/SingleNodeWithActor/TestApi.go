package main

import (
	"fmt"
	"github.com/zllangct/rockgo/actor"
	"github.com/zllangct/rockgo/cluster"
	"github.com/zllangct/rockgo/network"
	MessageProtocol "github.com/zllangct/rockgo/network/messageProtocol"
)

/*
协议接口组,需继承network.ApiBase 基类

api结构体中，符合条件的方法会被作为api对客户端服务
api方法的规则为：func 函数名（sess **network.Session, message *消息结构体）
*/

type TestApi struct {
	network.ApiBase
	nodeComponent *Cluster.NodeComponent
	actorProxy    *Actor.ActorProxyComponent
}

//使用协议接口时，需先初始化，初始化时需传入定义的消息号对应字典
//以及所需的消息序列化组件，可轻易切换为protobuf，msgpack等其他序列化工具
func NewTestApi() *TestApi {
	r := &TestApi{}
	r.Instance(r).SetMT2ID(Testid2mt).SetProtocol(&MessageProtocol.JsonProtocol{})
	return r
}

//协议接口 1  Hello
func (this *TestApi) Hello(sess *network.Session, message *TestMessage) {
	println(fmt.Sprintf("Hello,%s", message.Name))
	p, err := this.GetParent()
	if err == nil {
		println(fmt.Sprintf("this api parent:%s", p.Name()))
	}

	//回复方式一
	sess.Emit(1, []byte(fmt.Sprintf("hello client %s", message.Name)))
}

//协议接口 2 创建房间
func (this *TestApi) CreateRoom(sess *network.Session, message *TestCreateRoom) {
	errReply := func() {
		r := &CreateResult{
			Result: false,
		}
		this.Reply(sess, r)
	}
	//升级session为actor服务调用器
	serviceCaller, err := this.Upgrade(sess)
	if err != nil {
		errReply()
		return
	}
	//调用创建房间服务
	reply, err := serviceCaller.Call("room", Service_RoomMgr_NewRoom, sess.ID)
	if err != nil {
		errReply()
		return
	}
	//reply 创建房间结果反馈到客户端
	r := &CreateResult{
		Result: true,
		RoomID: reply[0].(int),
	}
	//回复方式二
	this.Reply(sess, r)
}

//获取actor proxy组件
func (this *TestApi) ActorProxy() (*Actor.ActorProxyComponent, error) {
	if this.actorProxy == nil {
		p, err := this.GetParent()
		if err != nil {
			return nil, err
		}
		err = p.Root().Find(&this.actorProxy)
		if err != nil {
			return nil, err
		}
	}
	return this.actorProxy, nil
}

//获取node组件
func (this *TestApi) NodeComponent() (*Cluster.NodeComponent, error) {
	if this.nodeComponent == nil {
		o, err := this.GetParent()
		if err != nil {
			return nil, err
		}
		err = o.Root().Find(&this.nodeComponent)
		if err != nil {
			return nil, err
		}
		return this.nodeComponent, nil
	}
	return this.nodeComponent, nil
}

func (this *TestApi) Upgrade(sess *network.Session) (*Actor.ActorServiceCaller, error) {
	proxy, err := this.ActorProxy()
	if err != nil {
		return nil, err
	}
	return Actor.NewActorServiceCallerFromSession(sess, proxy), nil
}
