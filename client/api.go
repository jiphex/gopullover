package client

import (
  "fmt"
  "net/http"
  "net/url"
  "encoding/json"
  "io/ioutil"
)

type APIResponse struct {
  Status int `json:"status"`
  Request string `json:"request"`
  Errors map[string][]string `json:"errors,omitempty"`
}

func (ar *APIResponse) OK() bool {
  return ar.Status == 1
}

type RegisterResponse struct {
  *APIResponse
  Id string `json:"id"`
}

type LoginResponse struct {
  *APIResponse
  Id string `json:"id"`
  Secret string `json:"secret"`
}

type Message struct {
  Id int `json:"id"`
  Umid int `json:"umid"`
  Title string `json:"title"`
  Message string `json:"message"`
  App string `json:"app"`
  AppID int `json:"aid"`
  IconURL string `json:"icon"`
  Timestamp uint64 `json:"date"`
  Priority int `json:"priority"`
  SoundURL string `json:"sound"`
  URL string `json:"url,omitempty"`
  URLTitle string `json:"url_title,omitempty"`
  Acked int `json:"acked,omitempty"`
  Receipt string `json:"receipt,omitempty"`
  HTML int `json:"html"`
}

type MessagesDevice struct {
  Encryption bool `json:"encryption_enabled"`
}

type MessagesUser struct {
  IOSLicense bool `json:"is_ios_licensed"`
  AndroidLicense bool `json:"is_android_licensed"`
  QuietHours bool `json:"quiet_hours"`
  DesktopLicense bool `json:"is_desktop_license"`
}

type MessagesResponse struct {
  *APIResponse
  Messages []Message `json:"messages"`
  Device MessagesDevice `json:"device"`
  User MessagesUser `json:"user"`
}

func Login(email, password string) (lr LoginResponse, err error) {
  pform := url.Values{}
  pform.Add("email",email)
  pform.Add("password",password)
  resp,err := http.PostForm("https://api.pushover.net/1/users/login.json", pform)
  if err != nil {
    return
  }
  if resp.StatusCode == 200 {
    dc := json.NewDecoder(resp.Body)
    dc.Decode(&lr)
    return
  } else {
    err = fmt.Errorf("Bad HTTP response from pushover: %d", resp.Status)
  }
  return
}

func RegisterDevice(secret, devicename string) (ar RegisterResponse, err error){
  pform := url.Values{}
  pform.Add("secret",secret)
  pform.Add("name",devicename)
  // https://pushover.net/api/client
  pform.Add("os","O")
  resp,err := http.PostForm("https://api.pushover.net/1/devices.json", pform)
  if err != nil {
    return
  }
  if resp.StatusCode == 200 {
    // io.Copy(os.Stdout, resp.Body)
    dc := json.NewDecoder(resp.Body)
    dc.Decode(&ar)
    return
  } else {
    err = fmt.Errorf("Bad Device Register response from pushover: %s", resp.Status)
  }
  return
}

func GetMessages(secret, deviceid string) (mr *MessagesResponse, err error) {
  pform := url.Values{}
  pform.Add("secret",secret)
  pform.Add("device_id",deviceid)
  // fmt.Printf("Requesting messages with %s\n", pform.Encode())
  resp,err := http.Get("https://api.pushover.net/1/messages.json?"+pform.Encode())
  if err != nil {
    return
  }
  if resp.StatusCode == 200 {
    var jbuf []byte
    jbuf,err = ioutil.ReadAll(resp.Body)
    if err != nil {
      fmt.Printf("Failed to read from body: %s", err)
      return
    }
    mr = &MessagesResponse{}
    jerr := json.Unmarshal(jbuf,mr)
    if jerr != nil {
      fmt.Printf("Bad unmarshal: [%s]\n", jerr)
      return
    }
    // fmt.Printf("M is %d %t %+v\n", mr.APIResponse.Status, mr.User.DesktopLicense, *mr)
  } else {
    b,_ := ioutil.ReadAll(resp.Body)
    err = fmt.Errorf("Bad Messages response from pushover: %s\n%s", resp.Status, b)
  }
  return
}

func DeleteUpto(secret, deviceid string, latest int) {
  pform := url.Values{}
  pform.Add("secret", secret)
  pform.Add("message", fmt.Sprintf("%d",latest))
  resp,err := http.PostForm(fmt.Sprintf("https://api.pushover.net/1/devices/%s/update_highest_message.json", deviceid), pform)
  if err != nil {
    fmt.Printf("Bad response from delete: %s", resp.Status)
  }
}