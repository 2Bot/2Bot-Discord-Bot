package actors

import (
	"fmt"

	"github.com/AsynkronIT/protoactor-go/router"

	routers "github.com/2Bot/2Bot-Discord-Bot/actors/router"
	"github.com/AsynkronIT/protoactor-go/actor"
)

type ActorSystem struct {
	*actor.ActorSystem
	GuildRouter   *actor.PID
	CommandParser *actor.PID
}

func NewActorSystem() *ActorSystem {
	system := actor.NewActorSystem()
	system.ProcessRegistry.Address = "2Bot"

	return &ActorSystem{
		ActorSystem:   system,
		GuildRouter:   spawnGuildManager(system),
		CommandParser: spawnCommandParsers(system),
	}
}

func guildManagerSupervisor(reason interface{}) actor.Directive {
	fmt.Printf("reason: %T %s", reason, reason)
	return actor.ResumeDirective
}

func spawnGuildManager(system *actor.ActorSystem) *actor.PID {
	props := routers.NewGuildRouter(system).WithSupervisor(actor.NewOneForOneStrategy(1, 1, guildManagerSupervisor))

	routerPID, _ := system.Root.SpawnNamed(props, "GuildManager")
	return routerPID
}

func spawnCommandParsers(system *actor.ActorSystem) *actor.PID {
	props := router.NewRoundRobinPool(10). /* WithProducer(nil). */ WithSupervisor(actor.NewRestartingStrategy()).WithFunc(func(ctx actor.Context) {

	})
	groupPID, _ := system.Root.SpawnNamed(props, "CommandParsers")
	return groupPID
}
