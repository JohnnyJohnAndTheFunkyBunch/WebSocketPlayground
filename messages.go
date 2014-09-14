package main


const (
    OPEN_MSG = 0
    CLOSE_MSG = 1
    ID_MSG = 2
    STATE_MSG = 3
    LATENCY_MSG = 4
    CONTENT_MSG = 5
    APPLICATION_MSG = 6
)

type OpenMsg struct {
    P uint8 // player id
    T uint8 // type
    M string
}
type CloseMsg struct {
    P uint8 // player id
    T uint8 // type
}
type IdMsg struct {
    P uint8 // player id
    T uint8 // type
    S SessionJSON
}
type StateMsg struct {
    P uint8 // player id
    T uint8 // type
    S SessionJSON
}

type LatencyMsg struct {
    P uint8 // player id
    T uint8 // type
    L int64 // latency
}

// client request name
// server responds html using C
type ContentMsg struct {
    P uint8
    T uint8
    C string
}

type ApplicationMsg struct {
    P uint8
    T uint8
    M interface{}
}

type YoutubeMsg struct {
    T uint8
    M string
}

type SessionJSON struct {
    Players []PlayerJSON
    C string
}

type PlayerJSON struct {
    Id uint8
    Name string
}

