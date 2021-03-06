package actions

import (
	"encoding/json"
	"github.com/abiosoft/ishell"
	"github.com/acmacalister/skittles"
	"github.com/synw/terr"
	"github.com/synw/centcom"
	"github.com/synw/centcli/libcentcli/state"
)


func Stop() *ishell.Cmd {
	command := &ishell.Cmd{
        Name: 	"stop",
        Help: 	"Stop an action: ex: stop listen channel_name",
        Func: 	func(ctx *ishell.Context) {
        	if state.Server == nil {
        		ctx.Println("No server selected: try the use command:", skittles.BoldWhite("use"), "server1")
        		return
        	}
			if len(ctx.Args) != 2 {
				err := terr.Err("One argument is required: ex: stop listen channel_name")
				ctx.Println(err.Error())
				return
			}
			action := ctx.Args[0]
			if action == "listen" {
				channel := ctx.Args[1]
				// check
				var listening []string
				found := false
				for _, ch := range(state.Listening) {
					if ch == channel {
						// unsub
						state.Cli.Unsubscribe(channel)
						found = true
						// shutdown listen routine if it is the last channel
						if len(state.Listening) == 1 {
							close(state.ListenChan)
							
						}
					} else {
						listening = append(listening, channel)
					}
				}
				if found == false {
					msg := "Not listening to channel "+channel
					err := terr.Err(msg)
					ctx.Println(err.Error())
					return
				}
				// update state
				state.Listening = listening
			}
		},
	}
	return command
}

func Listen() *ishell.Cmd {
	command := &ishell.Cmd{
        Name: 	"listen",
        Help: 	"Listen to channels: ex: listen channel_name",
        Func: 	func(ctx *ishell.Context) {
        	if state.Server == nil {
        		ctx.Println("No server selected: try the use command:", skittles.BoldWhite("use"), "server1")
        		return
        	}
			if len(ctx.Args) != 1 {
				err := terr.Err("One argument is required: ex: listen channel_name")
				ctx.Println(err.Error())
				return
			}
			channel := ctx.Args[0]
			centcom.SetVerbosity(2)
			// subscribe
			err := state.Cli.Subscribe(channel)
			if err != nil {
				ctx.Println(err)
			}
			// listen
			if len(state.Listening) == 0 {
				c := make(chan struct{})
				go func() {
					for {
						select {
						case msg := <-state.Cli.Channels:
							ctx.Println("->", msg.Channel, ":", msg.Payload)
						case <-c:
							return
						}
					}
				}()
				state.ListenChan = c
			}
			state.Listening = append(state.Listening, channel)
			ctx.Println("Listening to channel", channel, "...")
		},
	}
	return command
}

func Publish() *ishell.Cmd {
	command := &ishell.Cmd{
        Name: 	"publish",
        Help: 	"Publish into a channel: ex: publish channel_name {'hello':'world','foo':'bar'} //note: use no space in your payload",
        Func: 	func(ctx *ishell.Context) {
        	if state.Server == nil {
        		ctx.Println("No server selected: try the use command:", skittles.BoldWhite("use"), "server1")
        		return
        	}
			if len(ctx.Args) != 2 {
				err := terr.Err("Two arguments are required: ex: publish channel_name {'hello':'world'}")
				ctx.Println(err.Error())
				return
			}
			channel := ctx.Args[0]
			payload := ctx.Args[1]
			dataBytes, err := json.Marshal(payload)
			if err != nil {
				trace := terr.New("cmd.actions.Publish", err)
				ctx.Println(trace.Formatc())
			}
			_, err = state.Cli.Http.Publish(channel, dataBytes)
			if err != nil {
				trace := terr.New("cmd.actions.Publish", err)
				ctx.Println(trace.Formatc())
			}
        },
 	}
 	return command
}
