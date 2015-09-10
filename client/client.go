package client

import (
  "fmt"
  "sync"
  
  "golang.org/x/net/websocket"
)
type PushoverClient struct {
  secret string
  syncing *sync.Mutex
  running *sync.Mutex
  
  UserKey string
  DeviceID string
  Messages chan *Message
}

const (
  DEVICE_NAME = "gopullover"
)

func (pc *PushoverClient) Sync(delete bool) {
  go func(client *PushoverClient,delete bool,sendto chan<- *Message) {
    pc.syncing.Lock()
    // sendto <- &Message{Title:"Yo"}
    mr,err := GetMessages(pc.secret, pc.DeviceID)
    if err == nil {
      var lastmessage int
      for _,m := range mr.Messages {
        pc.Messages <- &m
        lastmessage = m.Id
      }
      if delete {
        DeleteUpto(pc.secret, pc.DeviceID, lastmessage)
      }
    } else {
      fmt.Printf("error: %s\n", err)
    }
    pc.syncing.Unlock()
  }(pc,delete,pc.Messages)
}

func (pc *PushoverClient) RunRealtime() {
  go func(pc *PushoverClient) {
    pc.running.Lock()
    wsOrigin := "http://localhost"
    wsUrl := "wss://client.pushover.net/push"
    ws,err := websocket.Dial(wsUrl, "", wsOrigin)
    if err != nil {
      fmt.Printf("Websocket problem: %s", err)
    }
    loginString := fmt.Sprintf("login:%s:%s\n", pc.DeviceID, pc.secret)
    // fmt.Printf("Sending: [%s]\n",loginString)
    if _,err := ws.Write([]byte(loginString)); err != nil {
      fmt.Printf("failed to write: %s\n", err)
    }
    running := true
    for running {
      msg := make([]byte, 1)
      if _,err := ws.Read(msg); err != nil {
        fmt.Printf("WS problem: %s\n", err)
        running = false
      }
      switch rune(msg[0]) {
      case '#':
        // passthrough
      case 'E':
        fmt.Printf("Error!\n")
        running = false
      case '!':
        fmt.Printf("Triggering sync\n")
        pc.Sync(true)
      case 'R':
        fmt.Printf("Triggering reload\n")
        // TODO: reload
        running = false
      default:
        fmt.Printf("wat? [%s]\n", msg[0])
        running = false
      }
    }
    pc.running.Unlock()
  }(pc)
}

func CreateClient(email, password, deviceid string) (pc *PushoverClient, err error) {
  pc = &PushoverClient{
    syncing: &sync.Mutex{},
    running: &sync.Mutex{},
    Messages: make(chan *Message,10),
  }
  lr,err := Login(email,password)
  if err != nil {
    return
  }
  if !lr.OK() {
    err = fmt.Errorf("Pushover status: %d", lr.APIResponse.Status)
    return
  }
  pc.UserKey = lr.Id
  pc.secret = lr.Secret
  // login done, register the device
  if len(deviceid) == 0 {
    // no device id, better register and get one
    var dr RegisterResponse
    dr, err = RegisterDevice(pc.secret, DEVICE_NAME)
    if err != nil {
      return
    }
    if !lr.OK() {
      err = fmt.Errorf("Pushover status: %d", lr.APIResponse.Status)
      return
    } else {
      fmt.Printf("Device ID is [%s]", dr.Id)
      pc.DeviceID = dr.Id
    }
  } else {
    pc.DeviceID = deviceid
  }
  pc.Sync(true)
  pc.RunRealtime()
  return
}