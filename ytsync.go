package main

import (
    "fmt"
    "encoding/json"
    "strconv"
)


type AppMsg interface {
}

type YTMsg struct {
    T int8
    M interface{}
}

type App interface {
    OnMsg(player *Player, msg string)
    OnConnected(player *Player)
    OnDisconnected(player *Player)
}

type YTSyncApp struct {
    videoId string
    playerStates map[*Player]int8
    playerTimes map[*Player]float64
    playerVideo map[*Player]string
    buffering bool
}
// IF -1, then it's a command
// if it's Player number, it is an event happeneing on someone else
// should display it
func (app *YTSyncApp) OnMsg(player *Player, msg string) {
    // convert
    var m Msg;
    err := json.Unmarshal([]byte(msg),&m)
    if err != nil {
        fmt.Println("Error in JSON", err.Error())
        return;
    }
    ytmsgbyte, _ := json.Marshal(m.M)
    var ytmsg YTMsg
    err = json.Unmarshal(ytmsgbyte,&ytmsg)
    if err != nil {
        fmt.Println("Error in JSON", err.Error())
        return;
    }
    // do something
    // change state, and store time, store videoId
    session := player.session
    players := session.players
    switch ytmsg.T {
    case -1:
        videoId, ok := ytmsg.M.(string)
        if videoId == app.playerVideo[player] {
            break
        }
        if ok {
            app.playerStates[player] = -1
            app.playerVideo[player] = videoId
            app.videoId = videoId
            for p := range players {
                if (p == player) {
                    continue
                }
                sendLoad(p, videoId)
            }
        }
    case 0:
        app.playerStates[player] = 0
        for p := range players {
            if (p == player) {
                continue
            }
            p.conn.SendTextMsg(string(msg))
        }
    case 1:
        time, ok := ytmsg.M.(float64)
        app.playerStates[player] = 1
        if ok {
            if time == 0 {
            } else {
                for p := range players {
                    if (p == player) {
                        continue
                    }
                    sendPlay(p)
                }
            }
        } else {
            fmt.Println("error")
        }
    case 2:
        // if case 2 without being in buffer mode, or seen a 3
        // just send everyone a pause
        time, ok := ytmsg.M.(float64)
        if app.buffering == false {
            if ok {
            // send pause message
                for p := range players {
                    if (p == player) {
                        continue
                    }
                    sendPause(p)
                }
                app.playerTimes[player] = time
                app.playerStates[player] = 2
            } else {
                fmt.Println("ERROR")
            }
        } else {
            if ok {
                app.playerTimes[player] = time
                app.playerStates[player] = 2
                doneBuffering := true
                for _, v := range app.playerStates {
                    if v == 3 {
                        doneBuffering = false
                        break
                    }
                }
                if doneBuffering {
                    fmt.Println("Everyone done buffering")
                    app.buffering = false
                    for p := range players {
                        sendPlay(p)
                    }
                }
                // search all state
            } else {
                fmt.Println("ERROR")
            }

        }
        // check everoyne else if their on the same 2 + time
        // (if starting) (no 3 event happened)
    case 3:
        // enter buffer mode
        app.playerStates[player] = 3
        if app.buffering {
            // send messages that they're buddering at certain time
        } else {
            app.buffering = true
            for p := range players{
                app.playerStates[p] = 3
            }
            time, ok := ytmsg.M.(float64)
            if ok {
            // send pause message to that specific time
            // (state buffering) -> wait until all receive pause messages 2 at
            // that time, then resume
                if time != 0 {
                    for p := range players {
                        if p != player {
                            sendPauseBuff(p, time)
                        }
                    }
                }
                app.playerTimes[player] = time
            } else {
                fmt.Println("ERROR");
            }
        }
    case 5:
        for p := range players {
            if (p == player) {
                continue
            }
            p.conn.SendTextMsg(string(msg))
        }
    }
    for k, v := range app.playerStates {
        fmt.Print("P:",k.id," S:", v,"|")
    }
    for p := range players {
        p.conn.SendTextMsg(string(msg))
    }
    fmt.Print("\n")
    fmt.Println("Buffer: ",app.buffering)
}

func sendPauseBuff(player *Player, time float64) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":4,\"M\":"+strconv.FormatFloat(time, 'f', -1, 64)+"}}"
    player.conn.SendTextMsg(msg)
}
func sendPlay(player *Player) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":1}}"
    player.conn.SendTextMsg(msg)
}
func sendLoad(player *Player, id string) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":-1,\"M\":\""+id+"\"}}"
    player.conn.SendTextMsg(msg)
}
func sendPause(player *Player) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":2}}"
    player.conn.SendTextMsg(msg)
}

func (app *YTSyncApp) OnConnected(player *Player) {
    //player.conn.SendTextMsg('{"P":-1,
    // send them -1
}
func (app *YTSyncApp) OnDisconnected(player *Player) {
    // delete player
    delete(app.playerStates, player)
    delete(app.playerTimes, player)
}

