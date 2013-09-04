package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/steve-wang/moses"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := work(); err != nil {
		fmt.Println(err)
	}
}

func Md5sum(str string) string {
	h := md5.New()
	io.WriteString(h, str)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func loadJson(file string, v interface{}) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func work() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	var param struct {
		Port int `json:"port"`
	}
	if err := loadJson(filepath.Join(wd, "moses_server.json"), &param); err != nil {
		return err
	}
	return run(wd, param.Port)
}

func run(wd string, port int) error {
	userlist, err := NewUserList(wd)
	if err != nil {
		return err
	}
	acceptor := moses.NewSOCK5Acceptor(userlist)
	srv := moses.NewServer(acceptor, &moses.DirectConnector{})
	if err := srv.Start(uint16(port)); err != nil {
		return err
	}
	srv.Serve()
	return nil
}

type UserList struct {
	mutex sync.Mutex
	file  string
	users map[string]string
}

type Auth struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

func NewUserList(wd string) (*UserList, error) {
	file := filepath.Join(wd, "moses_userlist.json")
	var users []Auth
	if err := loadJson(file, &users); err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for _, user := range users {
		m[user.User] = user.Password
	}
	p := &UserList{
		users: m,
		file:  file,
	}
	go p.run()
	return p, nil
}

func (p *UserList) Check(user, password string) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	pwd, ok := p.users[user]
	return ok && pwd == Md5sum(password)
}

func (p *UserList) update(users []Auth) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.users = make(map[string]string, len(users))
	for _, user := range users {
		p.users[user.User] = user.Password
	}
}

func (p *UserList) run() {
	t := time.Now()
	tick := time.Tick(time.Minute)
	var users []Auth
	for {
		<-tick
		fi, err := os.Stat(p.file)
		if err != nil {
			continue
		}
		if t.After(fi.ModTime()) {
			continue
		}
		if err := loadJson(p.file, &users); err != nil {
			continue
		}
		p.update(users)
		t = fi.ModTime()
	}
}
