package main

import (
  "log"
  "flag"
  "github.com/jiphex/gopullover/client"
)

func main() {
  email := flag.String("email", "test@example.com", "Pushover login email address")
  password := flag.String("password", "example", "Pushover login password")
  deviceid := flag.String("device", "AAAZZZ", "Pushover device ID")
  flag.Parse()
  pc,err := client.CreateClient(*email, *password, *deviceid)
  if err != nil {
    log.Println("Error is %s", err)
  }
  for {
    m := <- pc.Messages
    log.Printf("[#%s\n%s]",m.Title,m.Message)
  }
  log.Println("done %+v", pc)
}