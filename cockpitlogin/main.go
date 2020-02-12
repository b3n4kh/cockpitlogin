package main

import (
    "crypto/rand"
    "encoding/base64"
    "encoding/json"
    "errors"
    "io/ioutil"
    "log"
    "net"
    "net/http"
    "os"
    "os/exec"
    "os/signal"
    "os/user"
    "strings"
    "unicode"
)

type Config struct {
    Admins []string `json:"admins"`
}

var Configuration Config

func loadConfig(configFile string) (Config, error) {
    var config Config
    content, _ := ioutil.ReadFile(configFile)
    json.Unmarshal([]byte(content), &config)
    return config, nil
}

func generateRandomBytes(n int) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }
    return b, nil
}

func sanitizeRef(s string) string {
    return strings.Map(
        func(r rune) rune {
            if r == '/' || unicode.IsLetter(r) {
                return r
            }
            return -1
        },
        s,
    )
}

func sanitizeUser(s string) string {
    return strings.Map(
        func(r rune) rune {
            if unicode.IsLetter(r) || unicode.IsDigit(r) {
                return r
            }
            return -1
        },
        s,
    )
}

func isAdmin(user string) bool {
    for _, admin := range Configuration.Admins {
        if user == admin {
            return true
        }
    }
    log.Printf("user: %s is not an Admin\n", user)
    return false
}

func generateRandomString(s int) (string, error) {
    b, err := generateRandomBytes(s)
    return base64.URLEncoding.EncodeToString(b), err
}

func setupSocket(socketPath string) (listener net.Listener) {
    os.Remove(socketPath)

    listener, err := net.Listen("unix", socketPath)
    if err != nil {
        log.Fatalf("Could not listen on %s: %v", socketPath, err)
        return
    }
    os.Chmod(socketPath, 0770)

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        <-c
        os.Remove(socketPath)
        os.Exit(0)
    }()
    return listener
}

func httpListener(listener net.Listener) {
    defer listener.Close()
    err := http.Serve(listener, nil)
    if err != nil {
        log.Fatalf("Could not start HTTP server: %v", err)
    }
}

func setPassword(username string) (passwd string, err error) {
    _, err = user.Lookup(username)
    if err != nil {
        return "", err
    }
    pass, err := generateRandomString(32)
    if err != nil {
        return "", err
    }
    // echo "new_password" | passwd --stdin user
    cmd := exec.Command("/usr/bin/sudo", "/usr/bin/passwd", "--stdin", username)
    cmd.Stdin = strings.NewReader(pass)
    err = cmd.Run()
    if err != nil {
        return "", err
    }
    return pass, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
    user := r.Header.Get("REMOTE-USER")
    protoheader := r.Header.Get("X-Forwarded-Proto")
    q := r.URL.Query()
    ref := q.Get("ref")
    impersonate := q.Get("impersonate")
    referrer := "/"

    if user == "" || protoheader == ""{
        w.WriteHeader(401)
        w.Write([]byte("REMOTE-USER or X-Forwarded-Proto Header not set"))
        return
    }

    if ref != "" {
        referrer += sanitizeRef(ref)
    }

    if impersonate != "" {
        impuser := sanitizeUser(impersonate)
        if isAdmin(user) {
            log.Printf("user: %s will impersonate: %s\n", user, impuser)
            user = impuser
        }
    }

    cockpituri := protoheader + "://" + r.Host

    password, err := setPassword(user)
    if err != nil {
        w.WriteHeader(401)
        log.Println(err.Error())
        w.Write([]byte("Could not set Password"))
        return
    }
    logindata := user + ":" + password
    cockpitCookie, csrf, err := getCookie(logindata, cockpituri)
    if err != nil {
        w.WriteHeader(401)
        log.Println(err.Error())
        w.Write([]byte("Could not get Cookie"))
        return
    }
    http.SetCookie(w, cockpitCookie)
    http.Redirect(w, r, cockpituri + referrer, 303)
    w.Write(csrf)
}

func getCookie(logindata string, host string) (cookie *http.Cookie, csrf []byte, err error) {
    client := &http.Client{}
    req, err := http.NewRequest("GET", host + "/cockpit/login", nil)
    if err != nil {
        return nil, nil, err
    }

    sEnc := base64.StdEncoding.EncodeToString([]byte(logindata))

    req.Header.Add("Authorization", "Basic "+sEnc)
    req.Header.Add("X-Authorize", "true")

    resp, err := client.Do(req)
    log.Printf("/cockpit/login: %d\n", resp.StatusCode)
    if resp.StatusCode != 200 {
        return nil, nil, errors.New("Couldn't get cookie")
    }

    body, err := ioutil.ReadAll(resp.Body)

    for _, cookie := range resp.Cookies() {
        if cookie.Name == "cockpit" {
            return cookie, body, nil
        }
    }
    return nil, nil, errors.New("Couldn't get cookie")
}

func main() {
    file := "/run/cockpitlogin/socket"
    confFile := "/etc/cockpitlogin/config.json"
    var err error

    Configuration, err = loadConfig(confFile)
    if err != nil {
        log.Fatalf("Could not read config %s: %v", confFile, err)
        return
    }

    listener := setupSocket(file)
    http.HandleFunc("/", handler)
    log.Printf("Start Listening on: %s", file)
    httpListener(listener)
}
