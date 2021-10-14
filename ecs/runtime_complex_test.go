package ecs_test

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/zllangct/rockgo/3rd/assert"
	"github.com/zllangct/rockgo/3rd/iter"
	"github.com/zllangct/rockgo/ecs"
	"io/ioutil"
	"strings"
)

// Add remove child adds a new child every second.
// When it has 10 children, it removes itself.
type AddRemoveChild struct {
	ecs.ComponentBase
	parent  *ecs.Object
	count   int
	elapsed float32
}

func (c *AddRemoveChild) New() ecs.IComponent {
	return &AddRemoveChild{}
}

func (c *AddRemoveChild) Attach(parent *ecs.Object) {
	c.parent = parent
}

func (c *AddRemoveChild) Update(context *ecs.Context) {
	c.elapsed += context.DeltaTime
	if c.elapsed > 1.0 {
		c.count += 1
		if c.count >= 3 {
			parent := c.parent.Parent()
			if parent != nil {
				parent.RemoveObject(c.parent)
			}
		} else {
			child := ecs.NewObject(fmt.Sprintf("Child: %d", c.count))
			c.parent.AddObject(child)
		}
		c.elapsed = 0
	}
}

// DumpState dumps an object tree of the runtime every 1/2 seconds
type DumpState struct {
	ecs.ComponentBase
	elapsed float32
}

func (c *DumpState) New() ecs.IComponent {
	return &DumpState{}
}

func (c *DumpState) Update(context *ecs.Context) {
	c.elapsed += context.DeltaTime

	if c.elapsed >= 0.5 {
		c.elapsed = 0.0
		_ = context.Object.Root()
	}
}

type Hello struct {
	ecs.ComponentBase
}

func (this *Hello) Start(context *ecs.Context) {
	this.Hello("my name is zhaolei.")

}

func (this *Hello) Hello(str string) {
	sum := 0
	for i := 0; i < 10000; i++ {
		sum = sum + i
	}
	//println("sum:",sum,str)
}

func (this *Hello) Update(context *ecs.Context) {
	this.Hello(strconv.Itoa(1))
}
func TestLargeObjects(T *testing.T) {
	//====================== IComponent
	runtime := ecs.NewRuntime(ecs.Config{
		ThreadPoolSize: 50,
	})

	root := ecs.NewObject("root")
	runtime.Root().AddObject(root)
	for i := 0; i < 1000; i++ {
		o1 := ecs.NewObject(strconv.Itoa(i))
		o1.AddComponent(&Hello{})
		root.AddObject(o1)
	}

	t1 := time.Now()
	for i := 0; i < 1000; i++ {
		runtime.UpdateFrame()
	}
	elapsed1 := time.Since(t1)
	println("ecs:", elapsed1)

	//========================== traditional
	tasklist := make([]*Hello, 1000)
	for i := 0; i < 1000; i++ {
		tasklist = append(tasklist, &Hello{})
	}
	t2 := time.Now()
	wg := sync.WaitGroup{}

	for j := 0; j < 50; j++ {
		wg.Add(1)
		go func() {
			for i := 0; i < 20; i++ {
				tasklist[i].Hello(strconv.Itoa(i))
			}
			wg.Done()
		}()
	}
	wg.Wait()
	elapsed2 := time.Since(t2)
	println("traditional:", elapsed2)
}

func TestComplexSerialization(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		logger := log.New(os.Stdout, "Runtime: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.SetOutput(ioutil.Discard) // No output thanks

		runtime := ecs.NewRuntime(ecs.Config{
			ThreadPoolSize: 10})

		runtime.Factory().Register(&AddRemoveChild{})

		runtime.Root().AddComponent(&DumpState{elapsed: 11})

		o1 := ecs.NewObject("Container One")
		w1 := ecs.NewObject("Worker 1")
		w2 := ecs.NewObject("Worker 2")

		o2 := ecs.NewObject("Container Two")
		w3 := ecs.NewObject("Worker 3")
		w4 := ecs.NewObject("Worker 4")

		o3 := ecs.NewObject("Container Tree")
		//o4 := IComponent.NewObject("Container Four")

		o1.AddObject(w1)
		o1.AddObject(w2)

		o2.AddObject(w3)
		o2.AddObject(w4)

		o1.AddObject(o2)

		w4.AddObject(o3)
		o3.AddObject(w4)

		w1.AddComponent(&AddRemoveChild{})
		w2.AddComponent(&AddRemoveChild{})
		w3.AddComponent(&AddRemoveChild{})
		w4.AddComponent(&AddRemoveChild{})
		o3.AddComponent(&Hello{})
		runtime.Root().AddObject(o1)
		t := &DumpState{}
		err := runtime.Root().Find(t)
		if err != nil {
			println(err.Error())
		}

		runtime.UpdateFrame()

		// Serialize o2 as an object template
		marker, err := runtime.Root().FindObject("Container One")
		T.Assert(err == nil)

		prefab, err := runtime.Extract(marker)
		T.Assert(err == nil)

		prefab.Name = "Copy 1"
		instance1, err := runtime.Insert(prefab, runtime.Root())
		T.Assert(err == nil)
		T.Assert(instance1 != nil)

		prefab.Name = "Copy 2"
		instance2, err := runtime.Insert(prefab, runtime.Root())
		T.Assert(err == nil)
		T.Assert(instance2 != nil)

		all, err := iter.Collect(runtime.Root().Objects())
		T.Assert(err == nil)
		T.Assert(len(all) == 2)
	})
}

func TestComplexRuntime(T *testing.T) {
	assert.Test(T, func(T *assert.T) {
		logger := log.New(os.Stdout, "Runtime: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.SetOutput(ioutil.Discard) // No output thanks

		runtime := ecs.NewRuntime(ecs.Config{
			ThreadPoolSize: 10})

		runtime.Root().AddComponent(&DumpState{})

		o1 := ecs.NewObject("Container One")
		w1 := ecs.NewObject("Worker 1")
		w2 := ecs.NewObject("Worker 2")

		o2 := ecs.NewObject("Container Two")
		w3 := ecs.NewObject("Worker 3")
		w4 := ecs.NewObject("Worker 4")

		o3 := ecs.NewObject("Container Three")
		w5 := ecs.NewObject("Worker 5")
		w6 := ecs.NewObject("Worker 6")

		o1.AddObject(w1)
		o1.AddObject(w2)

		o2.AddObject(w3)
		o2.AddObject(w4)

		w4.AddObject(o3)
		o3.AddObject(w5)
		o3.AddObject(w6)

		w1.AddComponent(&AddRemoveChild{})
		w2.AddComponent(&AddRemoveChild{})
		w3.AddComponent(&AddRemoveChild{})
		w4.AddComponent(&AddRemoveChild{})
		w5.AddComponent(&AddRemoveChild{})
		w6.AddComponent(&AddRemoveChild{})

		runtime.Root().AddObject(o1)
		runtime.Root().AddObject(o2)

		for i := 0; i < 50; i++ {
			runtime.UpdateFrame()
		}

		expectedOutput := strings.Trim(`
object: Untitled (2 / 1)
! *github.com/zllangct/RockGO/Component_test.DumpState
   object: Container One (0 / 0)
   object: Container Two (0 / 0)`, " \n")

		actualOutput := strings.Trim(runtime.Root().Debug(), " \n")

		T.Assert(expectedOutput == actualOutput)
	})
}
