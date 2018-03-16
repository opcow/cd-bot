package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/opcow/go-ircevent"
)

var lastCD time.Time
var start = make(chan int)
var quit = make(chan bool)

// Config blah
type Config struct {
	Nick     string
	User     string
	Password string
	Server   string
	Port     string
	Channel  string
	Secret   string
}

func main() {

	//usr, err := user.Current()
	//if err != nil {
	//      log.Fatal(err)
	//}
	//cfile := usr.HomeDir + "/.cd-bot.cfg"
	cfile := ".cd-bot.cfg"

	var config Config
	if _, err := toml.DecodeFile(cfile, &config); err != nil {
		fmt.Println(err)
		return
	}

	if config.User == "" {
		config.User = config.Nick
	}

	if config.Port == "" {
		config.Port = "6667"
	}

	if strings.Contains(config.Server, ":") {
		config.Server = "[" + config.Server + "]"
	}
	config.Server = config.Server + ":" + config.Port

	ircobj := irc.IRC(config.Nick, config.User)
	ircobj.Password = config.Password
	err := ircobj.Connect(config.Server)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed connecting")
		return
	}
	//daemon.SdNotify(false, "READY=1")

	chanName := config.Channel
	seed := rand.NewSource(time.Now().Unix())
	rnd := rand.New(seed)

	ircobj.AddCallback("001", func(e *irc.Event) {
		ircobj.Join(chanName)
	})

	go cdPrinter(ircobj, chanName)
	ircobj.AddCallback("PRIVMSG", func(event *irc.Event) {

		if event.Message() == config.Secret+" bye" {
			quit <- true
			ircobj.Part(chanName)
			ircobj.Quit()
			os.Exit(0)
		}

		m := strings.Split(event.Message(), " ")
		if event.Arguments[0] != chanName {
			return
		}
		if m[0] == "!cd" {
			if len(m) > 1 {
				var count int
				_, err := fmt.Sscanf(m[1], "%v", &count)
				//count, err := strconv.Atoi(m[1])
				if err != nil {
					ircobj.Privmsg(chanName, idiot[rnd.Intn(len(idiot))])
				} else if count == 0 {
					ircobj.Privmsg(chanName, "You in a hurry?")
				} else {
					if count < -10 || count > 10 {
						ircobj.Privmsg(chanName, "The count must be a number from 1 to 10 (defaults to 5).")
					} else {
						start <- count
					}
				}
			} else {
				start <- 5
			}
		}

		// else if m[0] == "!si" {

		// 	info, err := valveqry.GetInfo("vps66848.vps.ovh.ca:27015")
		// 	if err != nil {
		// 		ircobj.Privmsg(chanName, error.Error(err))
		// 		return
		// 	}
		// 	ircobj.Privmsg(chanName, buildString(info))
		// }
		//event.Nick Contains the sender
		//event.Arguments[0] Contains the channel

	})
	ircobj.Loop()
}

// func buildString(si *valveqry.ServerInf) string {

// 	s := fmt.Sprintf("%s | protocol: %d | map: %s | game: %s | ver: %s | players: %d/%d | bots: %d | ",
// 		si.Name, si.Protocol, si.Map, si.Game, si.Version, si.Players, si.MaxPlayers, si.Bots)

// 	if si.Visibility == 0 {
// 		s += "public | "
// 	} else {
// 		s += "private | "
// 	}
// 	if si.Vac == 0 {
// 		s += "unsecured"
// 	} else {
// 		s += "VAC secured"
// 	}
// 	return s
// }

func cdPrinter(conn *irc.Connection, chanName string) {
	lastCD := time.Now().Add(time.Second * -20)
	for {
		select {
		case <-quit:
			return
		case i := <-start:
			if throttle(lastCD) {
				lastCD = time.Now()
				ticker := time.NewTicker(1 * time.Second)
				for range ticker.C {
					if i == 0 {
						break
					}
					conn.Privmsg(chanName, strconv.Itoa(i))
					if i < 0 {
						i++
					} else {
						i--
					}
				}
				ticker.Stop()
				conn.Privmsg(chanName, "\x0309Go!")
			} else {
				conn.Privmsg(chanName, "\x0304No!")
			}
		}
	}
}

func throttle(lastTime time.Time) bool {
	//t1 := time.Date(2006, 1, 1, 12, 23, 10, 0, time.UTC)
	if time.Now().Sub(lastTime).Seconds() < 20 {
		return false
	}
	return true
}
