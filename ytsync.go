package main


type AppMsg interface {
}

type App interface {
    OnMsg(player *Player, msg string)
    OnConnected(player *Player)
    OnDisconnected(player *Player)
}

type YTSyncApp struct {
    videoId string
    playerStates map[*Player]uint8
}

func (app YTSyncApp) OnMsg(player *Player, msg string) {
    session := player.session
    players := session.players
    for p := range players {
        if (p == player) {
            continue
        }
        p.conn.SendTextMsg(string(msg))
    }
}
func (app YTSyncApp) OnConnected(player *Player) {
}
func (app YTSyncApp) OnDisconnected(player *Player) {
}

