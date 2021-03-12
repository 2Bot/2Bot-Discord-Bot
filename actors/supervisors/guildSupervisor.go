package supervisors

import (
	"github.com/AsynkronIT/protoactor-go/actor"
)

var _ actor.SupervisorStrategy = &GuildSupervisor{}
var _ actor.Actor = &GuildSupervisor{}

type GuildSupervisor struct{}

func (g *GuildSupervisor) Receive(context actor.Context) {
}

func (g *GuildSupervisor) HandleFailure(system *actor.ActorSystem, supervisor actor.Supervisor, child *actor.PID, rs *actor.RestartStatistics, reason interface{}, message interface{}) {
	supervisor.ResumeChildren(child)
}
