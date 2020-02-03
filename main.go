package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strings"

	"github.com/sethvargo/go-password/password"
)

func setupSocket(socketPath string) (listener net.Listener) {
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Could not listen on %s: %v", socketPath, err)
		return
	}
	os.Chmod(socketPath, 0775)

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
	u, err := user.Lookup(username)
	if err != nil {
		return "", err
	}
	fmt.Println(u.Name)
	pass, err := password.Generate(64, 10, 0, false, true)
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
	if user == "" {
		w.WriteHeader(401)
		w.Write([]byte("User not found"))
		return
	}
	password, err := setPassword(user)
	if err != nil {
		w.WriteHeader(401)
		fmt.Println(err.Error())
		w.Write([]byte("Could not set Password"))
		return
	}

	cockpitCookie, csrf, err := getCookie(user + ":" + password)
	if err != nil {
		w.WriteHeader(401)
		fmt.Println(err.Error())
		w.Write([]byte("Could not get Cookie"))
		return
	}
	http.SetCookie(w, cockpitCookie)
	http.Redirect(w, r, "http://localhost", 303)
	w.Write(csrf)
}

func getCookie(logindata string) (cookie *http.Cookie, csrf []byte, err error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost/cockpit/login", nil)

	if err != nil {
		log.Fatal(err)
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(logindata))

	req.Header.Add("Authorization", "Basic "+sEnc)
	req.Header.Add("X-Authorize", "true")

	resp, err := client.Do(req)
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("/cockpit/login: %d\n", resp.StatusCode)

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "cockpit" {
			return cookie, body, nil
		}
	}
	return nil, nil, errors.New("Couldn't get cookie")
}

func main() {
	file := "/run/cockpitlogin/socket"

	listener := setupSocket(file)
	http.HandleFunc("/", handler)
	log.Printf("Start Listening on: %s", file)
	httpListener(listener)
}
