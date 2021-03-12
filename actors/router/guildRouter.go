package router

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/2Bot/2Bot-Discord-Bot/actormodels"

	"github.com/2Bot/2Bot-Discord-Bot/actors/supervisors"
	"github.com/AsynkronIT/protoactor-go/actor"
	"github.com/AsynkronIT/protoactor-go/router"
)

var (
	_ router.State        = &GuildRouter{}
	_ router.RouterConfig = &GuildRouter{}
	_ actor.Process       = &guildRouterProcess{}
	_ actor.Actor         = &guildRouteActor{}
)

// GuildRouter acts as the registry for all GuildRouter actors,
// responding to queries for GuildRouter PIDs.
type GuildRouter struct {
	system  *actor.ActorSystem
	spawner actor.SpawnerContext
	sender  actor.SenderContext
}

func NewGuildRouter(system *actor.ActorSystem) *actor.Props {
	guildRouter := &GuildRouter{
		system: system,
	}

	return (&actor.Props{}).WithSpawnFunc(func(actorSystem *actor.ActorSystem, id string, props *actor.Props, parentContext actor.SpawnerContext) (*actor.PID, error) {
		ref := &guildRouterProcess{
			actorSystem: actorSystem,
		}
		proxy, absent := actorSystem.ProcessRegistry.Add(ref, id)
		if !absent {
			return proxy, actor.ErrNameExists
		}

		var pc = *props
		pc.WithSpawnFunc(nil)
		ref.state = guildRouter.CreateRouterState()

		wg := &sync.WaitGroup{}
		wg.Add(1)
		ref.router, _ = actor.DefaultSpawner(actorSystem, id+"/router", actor.PropsFromProducer(func() actor.Actor {
			return &guildRouteActor{
				props:  &pc,
				config: guildRouter,
				state:  ref.state,
				wg:     wg,
			}
		}), parentContext)
		wg.Wait() // wait for routerActor to start

		ref.parent = parentContext.Self()
		return proxy, nil
	})
}

func (m *GuildRouter) RouteMessage(message interface{}) {
	//msg := message.(actormodels.GuildEnvelope)
	msg := actor.UnwrapEnvelopeMessage(message).(actormodels.GuildEnvelope)
	pid := m.getOrSpawn(msg.GuildID, true)
	m.sender.Send(pid, message)
}

func (m *GuildRouter) SetRoutees(routees *actor.PIDSet) {}

func (m *GuildRouter) GetRoutees() *actor.PIDSet {
	return nil
}

func (m *GuildRouter) SetSender(sender actor.SenderContext) {
	m.sender = sender
}

func (m *GuildRouter) RouterType() router.RouterType {
	return router.PoolRouterType
}

func (m *GuildRouter) OnStarted(context actor.Context, props *actor.Props, state router.State) {
	state.SetSender(context)
	m.spawner = context
}

func (m *GuildRouter) CreateRouterState() router.State {
	return m
}

func (m *GuildRouter) getOrSpawn(id string, spawn bool) *actor.PID {
	if pid, ok := m.system.ProcessRegistry.LocalPIDs.Get(id); ok {
		return pid.(*actor.PID)
	}

	if !spawn {
		return nil
	}

	props := actor.PropsFromProducer(func() actor.Actor {
		return &GuildCommandRouter{
			GuildID: id,
		}
	}).WithSupervisor(&supervisors.GuildSupervisor{})

	pid, err := m.spawner.SpawnNamed(props, id)
	log.Println("spawned actor", pid, err)
	return pid
}

type guildRouterProcess struct {
	parent      *actor.PID
	router      *actor.PID
	mu          sync.Mutex
	state       router.State
	watchers    actor.PIDSet
	stopping    int32
	actorSystem *actor.ActorSystem
}

func (p *guildRouterProcess) SendUserMessage(pid *actor.PID, message interface{}) {
	_, msg, _ := actor.UnwrapEnvelope(message)
	if _, ok := msg.(actormodels.GuildEnvelope); ok {
		p.state.RouteMessage(message)
	} else if _, ok := msg.(router.ManagementMessage); ok {
		r, _ := p.actorSystem.ProcessRegistry.Get(p.router)
		// Always send the original message to the router actor,
		// since if the message is enveloped, the sender need to get a response.
		r.SendUserMessage(pid, message)
	}
}

func (p *guildRouterProcess) SendSystemMessage(pid *actor.PID, message interface{}) {
	switch msg := message.(type) {
	case *actor.Watch:
		if atomic.LoadInt32(&p.stopping) == 1 {
			if r, ok := p.actorSystem.ProcessRegistry.Get(msg.Watcher); ok {
				r.SendSystemMessage(msg.Watcher, &actor.Terminated{Who: pid})
			}
			return
		}
		p.mu.Lock()
		p.watchers.Add(msg.Watcher)
		p.mu.Unlock()

	case *actor.Unwatch:
		p.mu.Lock()
		p.watchers.Remove(msg.Watcher)
		p.mu.Unlock()

	case *actor.Stop:
		term := &actor.Terminated{Who: pid}
		p.mu.Lock()
		p.watchers.ForEach(func(_ int, other *actor.PID) {
			if !other.Equal(p.parent) {
				if r, ok := p.actorSystem.ProcessRegistry.Get(other); ok {
					r.SendSystemMessage(other, term)
				}
			}
		})
		// Notify parent
		if p.parent != nil {
			if r, ok := p.actorSystem.ProcessRegistry.Get(p.parent); ok {
				r.SendSystemMessage(p.parent, term)
			}
		}
		p.mu.Unlock()

	default:
		r, _ := p.actorSystem.ProcessRegistry.Get(p.router)
		r.SendSystemMessage(pid, message)

	}
}

func (p *guildRouterProcess) Stop(pid *actor.PID) {
	if atomic.SwapInt32(&p.stopping, 1) == 1 {
		return
	}

	p.actorSystem.Root.StopFuture(p.router).Wait()
	p.actorSystem.ProcessRegistry.Remove(pid)
	p.SendSystemMessage(pid, &actor.Stop{})
}

type guildRouteActor struct {
	props  *actor.Props
	config router.RouterConfig
	state  router.State
	wg     *sync.WaitGroup
}

func (a *guildRouteActor) Receive(context actor.Context) {
	log.Printf("router actor got %T %v\n", context.Message(), context.Message())
	switch m := context.Message().(type) {
	case *actor.Started:
		a.config.OnStarted(context, a.props, a.state)
		a.wg.Done()

	case *router.AddRoutee:
		r := a.state.GetRoutees()
		if r.Contains(m.PID) {
			return
		}
		context.Watch(m.PID)
		r.Add(m.PID)
		a.state.SetRoutees(r)

	case *router.RemoveRoutee:
		r := a.state.GetRoutees()
		if !r.Contains(m.PID) {
			return
		}

		context.Unwatch(m.PID)
		r.Remove(m.PID)
		a.state.SetRoutees(r)
		// sleep for 1ms before sending the poison pill
		// This is to give some time to the routee actor receive all
		// the messages. Specially due to the synchronization conditions in
		// consistent hash router, where a copy of hmc can be obtained before
		// the update and cause messages routed to a dead routee if there is no
		// delay. This is a best effort approach and 1ms seems to be acceptable
		// in terms of both delay it cause to the router actor and the time it
		// provides for the routee to receive messages before it dies.
		time.Sleep(time.Millisecond * 1)
		context.Send(m.PID, &actor.PoisonPill{})

	case *router.BroadcastMessage:
		msg := m.Message
		sender := context.Sender()
		a.state.GetRoutees().ForEach(func(i int, pid *actor.PID) {
			context.RequestWithCustomSender(pid, msg, sender)
		})

	case *router.GetRoutees:
		r := a.state.GetRoutees()
		routees := make([]*actor.PID, r.Len())
		r.ForEach(func(i int, pid *actor.PID) {
			routees[i] = pid
		})

		context.Respond(&router.Routees{routees})
	case *actor.Terminated:
		r := a.state.GetRoutees()
		if r.Remove(m.Who) {
			a.state.SetRoutees(r)
		}
	}
}
