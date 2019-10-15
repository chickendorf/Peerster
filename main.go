
package main

import (
  "fmt"
	"flag"
  G "./pkg/gossip"
  W "./pkg/web"
  "strings"
)

func main() {

  var uiPort = flag.String(
		"UIPort",
		"8080",
		"port for the UI client (default \"8080\")",
	)

  var gossipAddr = flag.String(
		"gossipAddr",
		"127.0.0.1:5000",
		"port for the gossiper (default \"127.0.0.1:5000\")",
	)

  var name = flag.String(
		"name",
		"gigi",
		"name of the gossiper",
	)

  var peersString = flag.String(
    "peers",
		"",
		"comma separated list of peers of the form ip:port",
	)

  var simple = flag.Bool(
		"simple",
		false,
		"run gossiper in simple broadcast mode",
	)

  var server = flag.Bool(
		"server",
		false,
		"run the web server",
	)

  var antiEntropy = flag.Int(
    "antiEntropy",
    10,
    "timeout in seconds for anti-entropy (defualt 10s)",
  )


  flag.Parse()

  var peerList []string
	if len(*peersString) > 0 {
		peerList = strings.Split(*peersString, ",")
	}

  /*fmt.Println("UIPort:", *uiPort);
  fmt.Println("gossipAddr:", *gossipAddr);
  fmt.Println("name:", *name);
  fmt.Println("peerList:", *peersList);
  fmt.Println("simple:", *simple);*/

  var mainGossip = G.InitGossiper(*name, *uiPort, *gossipAddr, peerList, *simple, *antiEntropy)

  if *server {
    fmt.Println("Run web")
    go W.RunServer(mainGossip);
  }

  mainGossip.ListenClient()

  fmt.Println("Exit")
}
