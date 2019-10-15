package main

import (
  "fmt"
	"flag"
  P "../pkg/packets"
  "github.com/dedis/protobuf"
  "net"
)

func main() {
  var uiPort = flag.String(
    "UIPort",
    "8080",
    "port for the UI client (default \"8080\")",
  )

  var msg = flag.String(
    "msg",
    "",
    "message to be sent",
  )

  flag.Parse();

  packet := P.InitClientPacket("Client248115", "127.0.0.1:" + *uiPort, *msg);

  sendClientPacket(packet)
}

func sendClientPacket(packet *P.ClientPacket){

  destinationAddr, _ := net.ResolveUDPAddr(
		"udp4",
		packet.Simple.RelayPeerAddr,
	)

  //Open the connection
  conn, _ := net.DialUDP("udp4", nil, destinationAddr)

  buf, err := protobuf.Encode(packet)

  if err != nil {
    fmt.Println("Error (encoding):", err.Error(), "with the packet :", packet)
    return
  }

  //Try to send
  size , err := conn.Write(buf)

	if err != nil {
		fmt.Println("Error (sending) :", err.Error(), "size :", size)
	}

}
