package remote
import "flag"
import "log"
import "../pipe"
import "github.com/Shopify/go-lua"

var service = flag.String("service", "127.0.0.1:8888", "")
var groupName = flag.String("group", "", "")
var nickName = flag.String("nick", "", "")
func Init() {
	flag.Parse()
	println("remote init")
        c:=make(chan *pipe.HelperInfo)
	d:=make(chan bool)
        go func() {
                for {
                        select {
                        case info:=<-c:
                                switch info.Cmd {
				case pipe.Shutdown:
					d<-true
                                case pipe.Request:
                                        log.Println("request")
					var s pipe.RequestCmd
					pipe.DecodeBytes(info.Bytes, &s)
					handleRequest(s)
                                }
                        }
                }
        }()
        client := pipe.NewInnerClient(*service, c)
        if client!=nil {
		info := &pipe.RemoteInfo{}
		info.Group = *groupName
		info.Nick = *nickName
               client.Send(pipe.RegRemote, info)
	       <-d
        } else {
		log.Println("dial fail")
	}
}

func handleRequest(s pipe.RequestCmd) {
	l := lua.NewState()
	lua.OpenLibraries(l)
	if err := lua.DoFile(l, "internal/init.lua"); err != nil {
		log.Println(err.Error())
		return //todo
	}
	str:=s.Cmd
	l.Global("m")
	l.PushString("unpack")
	l.Table(-2)
	l.PushString(str)
	l.Call(1,1)
	l.SetGlobal("RequestInfo")
	if _err := lua.DoFile(l, "logic_remote/handle.lua"); _err != nil { 
		log.Println(_err.Error())
		return //todo
	}
}
