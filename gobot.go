package main

import "flag"
import "fmt"
import "io/ioutil"
import "os"
import "os/exec"
import irc "github.com/thoj/go-ircevent"
import re "regexp"
import s "strings"

const nick = "GOBOT"

var password string
var host string
var realname string
var port string
var channel string

var nickPattern = re.MustCompile("([^!]+)!")
var cheesePattern = re.MustCompile("(?i)cheese")
var implementedPattern = re.MustCompile("(?i)implemented")
var equationPattern = re.MustCompile(".*= *$")
var expPattern = re.MustCompile("GOBOT:(.*)=")

func dispatch(e *irc.Event) {
	conn := e.Connection
	fmt.Println(e.Message())
	fmt.Println(cheesePattern.MatchString(e.Message()))
	switch {
	case cheesePattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: cheese is pretty good", e.Nick)
	case implementedPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: yes damnit, satisfied?", e.Nick)
	case equationPattern.MatchString(e.Message()):
		expression := expPattern.FindStringSubmatch(e.Message())[1]
		fmt.Println("expression:", expression)
		bc := exec.Command("bc")
		bcIn, err := bc.StdinPipe()
		if err != nil {
			return
		}
		bcOut, err := bc.StdoutPipe()
		if err != nil {
			return
		}

		bc.Start()
		bcIn.Write([]byte(expression))
		bcIn.Write([]byte("\n"))
		bcIn.Close()
		bcBytes, _ := ioutil.ReadAll(bcOut)
		bc.Wait()
		result := s.Trim(string(bcBytes), " \n\t")
		if result == "" {
			conn.Privmsgf(channel, "%s: I think your equation is wrong, based on Math... try without extra words.", e.Nick)
		} else {
			conn.Privmsgf(channel, "%s: %s, I think...", e.Nick, result)
		}

	default:
		conn.Privmsgf(channel, "%s: I'm not sure what you mean...", e.Nick)
	}
}

func main() {

	flag.StringVar(&password, "pass", "", "Server password.")
	flag.StringVar(&host, "server", "irc.baconbunny.com", "Server to connect to.")
	flag.StringVar(&realname, "rname", "petkovich", "Realname.")
	flag.StringVar(&port, "port", "6667", "Server port.")
	flag.StringVar(&channel, "chan", "#gentoo", "Channel to connect ")
	flag.Parse()

	os.Setenv("BC_LINE_LENGTH", "0")
	conn := irc.IRC("GOBOT", "GOBOT")
	conn.Password = password
	conn.Connect(fmt.Sprintf("%s:%s", host, port))
	conn.AddCallback("001", func(e *irc.Event) { conn.Join("#gentoo") })
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		if toGoBot, _ := re.MatchString("GOBOT:", e.Message()); toGoBot {
			dispatch(e)
		}
	})
	conn.Loop()
}
