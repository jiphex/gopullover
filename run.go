package main

import (
  "log"
  "flag"
  "github.com/jiphex/gopullover/client"
)

func main() {
  email := flag.String("email", "test@example.com", "Pushover login email address")
  password := flag.String("password", "example", "Pushover login password")
  settings := flag.String("config", "settings.json", "Settings file")
  flag.Parse()
  pc,err := client.CreateClient(*email, *password, *settings)
  if err != nil {
    log.Printf("Error is %s\n", err)
  }
  for {
    m := <- pc.Messages
    log.Printf("[#%s\n%s]",m.Title,m.Message)
  }
  log.Println("done %+v", pc)
}