package packets

import (
  S "../status"
)

type StatusPacket struct {
  Want []S.PeerStatus
}

func InitStatusPacket(
	want []S.PeerStatus,
) *StatusPacket {
  return &StatusPacket{Want: want}
}
