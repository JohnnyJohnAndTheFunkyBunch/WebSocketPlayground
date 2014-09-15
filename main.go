package main

import (
	"fmt"
	"github.com/JohnnyJohnAndTheFunkyBunch/simplewebsocket"
	"log"
	"os"
    "encoding/json"
    "strconv"
    "io/ioutil"
)

type Player struct {
    id uint8
    conn *websocket.Conn
    session *Session
    username string
}

type Session struct {
    id string
    players map[*Player]bool
    playerCounter uint8
    content string
    app App
}

type Msg struct {
    P uint8 // player id
    T uint8 // type
}

var sessionMap map[string]*Session
var playerMap map[*websocket.Conn]*Player
var contentMap map[string]string // name : htmlfile

// P: Player, Number, Connect/Disconnct
// S: Session
// A: Application // file xml with application aplpication numberS
// M: Number, Messages (application handles it)
// 0 : open
// 1 : close
// 2 : identify player id
// 3 : state of session (players, current game)
// 4 : Latency message
// 5 : change content

func onConnected(conn *websocket.Conn) {
    fmt.Println("Connected: ", conn.Connection())
    // size of map
    // player == size + 1
    playerMap[conn] = &Player{conn:conn,session:nil}
}
func onDisconnected(conn *websocket.Conn) {
    fmt.Println("Disconnected: ", conn.Connection())
    player := playerMap[conn]
    if player != nil {
        repInit := Msg{P:player.id,T:1}
        repByte, _ := json.Marshal(repInit)
        session := player.session
        for p := range session.players {
            if p.conn == conn {
                continue
            }
            p.conn.SendTextMsg(string(repByte))
        }
        delete(session.players, player)
    }
    // have a stack of reusable player id
    // send message to session saying disconnected, find session
}
func onMsg(conn *websocket.Conn, msg string) {
    var m Msg;
    fmt.Println(msg)
    err := json.Unmarshal([]byte(msg), &m)
    if err != nil {
        fmt.Println("Error in JSON", err.Error())
        return;
    }
    switch m.T {
    case 0:
        var r OpenMsg
        json.Unmarshal([]byte(msg), &r)
        if r.M == "" {
            fmt.Println("Not a session");
            return;
        }
        handleStartMsg(conn, r.M)
    case 1:
        // send disconnected
    case 5:
        var r ContentMsg
        json.Unmarshal([]byte(msg), &r)
        handleContentMsg(conn, r)
    case 6:
        player := playerMap[conn]
        session := player.session
        if session.app != nil {
            session.app.OnMsg(player, msg)
        }

        //players := session.players
        //for p := range players {
        //    if (p.conn == conn) {
        //        continue
        //    }
        //    p.conn.SendTextMsg(string(msg))
        //}
    }


}

func handleContentMsg(conn *websocket.Conn, msg ContentMsg) {
    player := playerMap[conn]
    session := player.session
    appName := msg.C
    session.content = msg.C
    players := session.players
    msg.C = contentMap[msg.C]
    for p := range players {
        msgByte, _ := json.Marshal(msg)
        p.conn.SendTextMsg(string(msgByte))
    }
    if appName == "youtube" {
        m := make(map[*Player]uint8)
        // this can be made more efficient by merfing with previous loo[
        for p := range players {
            m[p] = 0
        }
        fmt.Println("session")
        session.app = YTSyncApp{videoId:"",playerStates:m}
    }
}

func handleStartMsg(conn *websocket.Conn, sessionId string) {
    fmt.Println("Session: ", sessionId)
    session := sessionMap[sessionId]
    player := playerMap[conn]
    if session != nil {
        fmt.Println("Session exist")
        // insert player
        if player != nil {
            player.session = session
            player.id = session.playerCounter
            player.username = "Player" + strconv.Itoa(int(player.id))
            session.players[player] = true
            session.playerCounter += 1
        }
    } else {
        fmt.Println("Creating session")
        newPlayerMap := make(map[*Player]bool)
        session = &Session{id:sessionId,players:newPlayerMap,playerCounter:0,content:"main"}
        sessionMap[sessionId] = session
        // insert player
        if player != nil {
            player.session = session
            player.id = 0
            player.username = "Player0"
            session.players[player] = true
            session.playerCounter += 1
        }
    }
    // send messgae saying connected to server
    players := session.players
    rep := Msg{P:player.id,T:0}
    var playersJSON []PlayerJSON
    for p := range players {
        if p.conn == conn {
            continue
        }
        repByte, _ := json.Marshal(rep)
        p.conn.SendTextMsg(string(repByte))
        playersJSON = append(playersJSON, PlayerJSON{p.id, p.username})
    }
    // send message to new connection with state
    sessionJSON := SessionJSON{Players:playersJSON,C:contentMap[session.content]}
    repInit := IdMsg{P:player.id,T:2,S:sessionJSON}
    repByte, _ := json.Marshal(repInit)
    player.conn.SendTextMsg(string(repByte))
}

func handleEventMsg(conn *websocket.Conn, msg string) {
    // send to rest of connections
    player := playerMap[conn]
    session := player.session
    players := session.players
    for p := range players {
        if p.conn == conn {
            continue
        }
        p.conn.SendTextMsg(msg)
    }
}

func handleDisconnectMsg(conn *websocket.Conn) {
}

func onPong(conn *websocket.Conn) {
    player := playerMap[conn]
    session := player.session
    players := session.players
    msg := LatencyMsg{player.id, 4, int64(player.conn.Latency())}
    msgByte, _ := json.Marshal(msg)
    for p := range players {
        p.conn.SendTextMsg(string(msgByte))
    }
}

func initWebSocket() {
	logger := log.New(os.Stdout, "websocket: ", log.Ltime)
	ws := websocket.CreateServer("localhost", "8000", logger)
	ws.HandleConnected(onConnected)
	ws.HandleDisconnected(onDisconnected)
	ws.HandleTextMsg(onMsg)
	ws.HandlePong(onPong)
	ws.ListenAndServe()
}

func loadFiles() {
    youtube, err := ioutil.ReadFile("content/youtube.html")
    if err != nil {
        panic("Can't load files")
    }
    main, err := ioutil.ReadFile("content/main.html")
    if err != nil {
        panic("Can't load files")
    }
    contentMap["youtube"] = string(youtube)
    contentMap["main"] = string(main)
}

func main() {
    // init session handler
    sessionMap = make(map[string]*Session)
    playerMap = make(map[*websocket.Conn]*Player)
    contentMap = make(map[string]string)
    // init files
    loadFiles()
    // init websocket
    initWebSocket()
}
