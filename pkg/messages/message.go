package messages

type SimpleMessage struct {
  OriginalName string
  RelayPeerAddr string
  Contents string
}

func InitSimpleMessage(
	senderName,
	relayAddr string,
	content string,
) *SimpleMessage {
  ret := SimpleMessage{OriginalName : senderName, RelayPeerAddr : relayAddr, Contents : content};
  return &ret;
}

type RumorMessage struct {
  Origin string
  ID uint32
  Text string
}

func InitRumorMessage(
	originName string,
	identifier uint32,
	txt string,
) *RumorMessage {
  ret := RumorMessage{Origin : originName, ID : identifier, Text : txt};
  return &ret;
}
