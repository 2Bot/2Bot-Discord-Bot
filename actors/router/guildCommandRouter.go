package router

import (
	"fmt"
	"log"

	"github.com/2Bot/2Bot-Discord-Bot/actormodels"

	"github.com/AsynkronIT/protoactor-go/actor"
)

// GuildCommandRouter handles every command for a given guild
type GuildCommandRouter struct {
	GuildID string
}

func (g *GuildCommandRouter) Receive(context actor.Context) {
	log.Printf("%s GOT UNE MESSAGE %T %#v\n", context.Self(), context.Message(), context.Message())
	switch msg := context.Message().(type) {
	case actormodels.GuildEnvelope:
		context.Respond(fmt.Sprintf("hello world %s", msg.Message))
	}
}
