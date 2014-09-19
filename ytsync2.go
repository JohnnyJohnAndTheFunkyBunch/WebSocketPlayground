package main

import (
    "fmt"
    "encoding/json"
    "strconv"
    "strings"
)



type YTMsg2 struct {
    T int8
    M interface{}
}


type YTSyncApp2 struct {
    videoId string
    playerStates map[*Player]int8
    playerTimes map[*Player]float64
    playerVideo map[*Player]string
    buffering bool
    playBuffer bool
    loading bool
}
// IF -1, then it's a command
// if it's Player number, it is an event happeneing on someone else
// should display it
func (app *YTSyncApp2) OnMsg(player *Player, msg string) {
    // convert
    var m Msg;
    err := json.Unmarshal([]byte(msg),&m)
    if err != nil {
        fmt.Println("Error in JSON", err.Error())
        return;
    }
    ytmsgbyte, _ := json.Marshal(m.M)
    var ytmsg YTMsg2
    err = json.Unmarshal(ytmsgbyte,&ytmsg)
    if err != nil {
        fmt.Println("Error in JSON", err.Error())
        return;
    }
    // do something
    // change state, and store time, store videoId
    // 4 = state info
    session := player.session
    players := session.players
    S:
    switch ytmsg.T {
    case -1:
        videoId, ok := ytmsg.M.(string)
        if !ok {
            fmt.Println("Error parsing videoId")
            break
        }
        if videoId == app.videoId {
            break
        }
        app.playBuffer = true
        app.videoId = videoId
        app.loading = true
        app.buffering = true
        for p := range players {
            app.sendLoad(p, videoId)
        }
    case 0:
    case 1:
        _, ok := ytmsg.M.(float64)
        if !ok {
            break
        }
        // check if everyone gucci, aka not 3
        for _, v := range app.playerStates {
            if v == 3 {
                app.playBuffer = true
                break S
            }
        }
        for p := range players {
            app.sendPlay(p)
        }
    case 2:
        // if case 2 without being in buffer mode, or seen a 3
        // just send everyone a pause
        time, ok := ytmsg.M.(float64)
        if !ok {
            break
        }
        app.playBuffer = false
        for p := range players {
            app.sendPause(p, time)
        }
    case 3:

    case 4:
        stateMsg , ok := ytmsg.M.(string)
        if !ok {
            break
        }
        stateMsgArray := strings.Split(stateMsg, ":")
        state, err := strconv.ParseInt(stateMsgArray[0], 0, 8)
        time, err2 := strconv.ParseFloat(stateMsgArray[1], 64)
        if err != nil {
            fmt.Println("Error:",err.Error())
            break
        }
        if err2 != nil {
            fmt.Println("Error:",err2.Error())
            break
        }
        app.handleStateInfo(player, session, int8(state), time)
        /*
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
        } */
    case 5:
    case 6:
        time, ok := ytmsg.M.(float64)
        if !ok {
            break
        }
        app.playBuffer = true
        for p := range players {
            app.sendPause(p, time)
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


func (app *YTSyncApp2) OnConnected(player *Player) {
    //player.conn.SendTextMsg('{"P":-1,
    // send them -1
}
func (app *YTSyncApp2) OnDisconnected(player *Player) {
    // delete player
    delete(app.playerStates, player)
    delete(app.playerTimes, player)
}

func (app *YTSyncApp2) sendPlay(player *Player) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":1}}"
    player.conn.SendTextMsg(msg)
}
func (app *YTSyncApp2) sendPause(player *Player, time float64) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":2,\"M\":"+strconv.FormatFloat(time, 'f', -1, 64)+"}}"
    player.conn.SendTextMsg(msg)
}
func (app *YTSyncApp2)sendLoad(player *Player, id string) {
    msg := "{\"P\":-1,\"T\":6,\"M\":{\"T\":-1,\"M\":\""+id+"\"}}"
    player.conn.SendTextMsg(msg)
}

func (app *YTSyncApp2) handleStateInfo(player *Player, session *Session, state int8, time float64) {
    app.playerStates[player] = state
    switch state {
    case -1:
        app.buffering = true
    case 0:
    case 1:
        for _, s := range app.playerStates {
            if s == 3 || s == -1 {
                return
            }
        }
        app.buffering = false
    case 2:
        fmt.Println(app.buffering, app.playBuffer)
        if app.buffering && app.playBuffer{
            for _, s := range app.playerStates {
                if s == 3 || s == -1 {
                    return
                }
            }
            app.buffering = false
            for p := range session.players {
                app.sendPlay(p)
            }
        }
    case 3:
        app.buffering = true
        if !app.loading {
            for p := range session.players {
                app.sendPause(p, time)
            }
        }
    case 5:
    }
}
