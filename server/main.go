package main

import (
	"encoding/json"
	"fmt"
	"github.com/steve-wang/moses"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Param struct {
	Port int `json:"port"`
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	var param Param
	load := func(param *Param) (err error) {
		file := filepath.Join(wd, "moses_server.json")
		defer func() {
			if err != nil {
				err = fmt.Errorf("failed to load %s: %s", file, err)
			}
		}()
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		return json.NewDecoder(f).Decode(param)
	}
	if err := load(&param); err != nil {
		fmt.Println(err)
		return
	}
	if err := run(wd, &param); err != nil {
		fmt.Println(err)
		return
	}
}

func run(wd string, param *Param) error {
	userlist, err := NewUserList(wd)
	if err != nil {
		return err
	}
	acceptor := moses.NewPrivateAcceptor(userlist)
	srv := moses.NewServer(acceptor, &moses.DirectConnector{})
	if err := srv.Start(uint16(param.Port)); err != nil {
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

func loadUserList(file string) (_ []Auth, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to load %s: %s", file, err)
		}
	}()
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var users []Auth
	if err := json.NewDecoder(f).Decode(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func NewUserList(wd string) (*UserList, error) {
	file := filepath.Join(wd, "moses_userlist.json")
	users, err := loadUserList(file)
	if err != nil {
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
	return ok && pwd == password
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
	for {
		<-time.After(time.Minute)
		fi, err := os.Stat(p.file)
		if err != nil {
			continue
		}
		if t.After(fi.ModTime()) {
			continue
		}
		users, err := loadUserList(p.file)
		if err != nil {
			continue
		}
		p.update(users)
	}
}
