package client

import (
  "fmt"
  "sync"
  "os"
  "encoding/json"
  
  "golang.org/x/net/websocket"
)
type PushoverClient struct {
  *Settings
  syncing *sync.Mutex
  running *sync.Mutex
  Messages chan *Message
}

const (
  DEVICE_NAME = "gopullover"
)

type Settings struct {
  DeviceID string `json:"device_id,omitempty"`
  secret string `json:"secret,omitempty"`
  UserKey string `json:"userkey,omitempty"`
}
 
func (pc *PushoverClient) saveSettings(fn string) (err error) {
  var f *os.File
  f,err = os.OpenFile(fn, os.O_CREATE | os.O_RDWR, 0600)
  f.Truncate(0)
  var settings []byte
  settings,err = json.Marshal(pc.Settings)
  if err != nil {
    return
  }
  var fb []byte
  fb,err = json.Marshal(settings)
  if err != nil {
    return err
  }
  _,err = f.Write(fb)
  f.Close()
  return
}

func (pc *PushoverClient) loadSettings(fn string) (err error) {
  var f *os.File
  f,err = os.OpenFile(fn, os.O_CREATE | os.O_RDONLY, 0600)
  if err != nil {
    return
  }
  defer f.Close()
  sd := json.NewDecoder(f)
  for sd.More() {
    sd.Decode(&pc.Settings)
  }
  return
}

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

func CreateClient(email, password, settingsfile string) (pc *PushoverClient, err error) {
  pc = &PushoverClient{
    syncing: &sync.Mutex{},
    running: &sync.Mutex{},
    Messages: make(chan *Message,10),
  }
  if len(settingsfile) > 0 {
    err = pc.loadSettings(settingsfile)
    if err != nil {
      return
    }
  }
  if len(pc.Settings.secret) == 0 {
    var lr LoginResponse
    lr,err = Login(email,password) 
    if err != nil {
      return
    }
    if !lr.OK() {
      err = fmt.Errorf("Pushover status: %d", lr.APIResponse.Status)
      return
    }
    pc.Settings.UserKey = lr.Id
    pc.Settings.secret = lr.Secret
  }
  // login done, register the device
  if len(pc.Settings.DeviceID) == 0 {
    // no device id, better register and get one
    var dr RegisterResponse
    dr, err = RegisterDevice(pc.secret, DEVICE_NAME)
    if err != nil {
      return
    }
    if !dr.OK() {
      err = fmt.Errorf("Pushover status: %d", dr.APIResponse.Status)
      return
    } else {
      fmt.Printf("Device ID is [%s]", dr.Id)
      pc.DeviceID = dr.Id
    }
  }
  pc.Sync(true)
  pc.RunRealtime()
  return
}