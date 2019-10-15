package packets

import (
  M "../messages"
  "fmt"
)

type ClientPacket struct {
  Simple *M.SimpleMessage
}

func InitClientPacket(
	senderName,
	relayAddr,
	content string,
) *ClientPacket {
  msg := M.InitSimpleMessage(senderName, relayAddr, content);
  ret := ClientPacket{ Simple : msg};
  return &ret;
}

func (p *ClientPacket) PrintMessage(){
  fmt.Println("CLIENT MESSAGE", p.Simple.Contents)
}
