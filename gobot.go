package main

import "flag"
import "fmt"
import "io/ioutil"
import "os"
import "os/exec"
import irc "github.com/thoj/go-ircevent"
import re "regexp"
import s "strings"
import "math/rand"

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
var languagePattern = re.MustCompile("(?i)(?:fuck|shit|damn|gay)")
var apologyPattern = re.MustCompile("(?i)(?:sorry|my bad)")
var helpPattern = re.MustCompile("(?i)(?:help|what.*you do)")
var thanksPattern = re.MustCompile("(?i)(?:thanks|cool|awesome)")
var coffeePattern = re.MustCompile("(?i)(?:coffee|caffeine|tea)")

var cannedResponse = []string{
	"%s: Sorry, what's that?",
	"%s: Uhhh, what?",
	"%s: Ask someone else for help, I'm busy.",
	"%s: I'm not sure what you mean...",
}
var helpMessage = "%s: I can do maths, that's basically it. Send me an equation followed by ="

func dispatch(e *irc.Event) {
	conn := e.Connection
	fmt.Println(e.Message())
	fmt.Println(cheesePattern.MatchString(e.Message()))
	switch {
	case cheesePattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: cheese is pretty good", e.Nick)
	case implementedPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: yes damnit, satisfied?", e.Nick)
	case helpPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, helpMessage, e.Nick)
	case equationPattern.MatchString(e.Message()):
		maths(e)
	case thanksPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: Glad I could maybe help somehow...", e.Nick)
	case languagePattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: Act like an adult, be more respectful", e.Nick)
	case apologyPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: It's ok, I forgive you.", e.Nick)
	case coffeePattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: Did you say coffee? Get one for JC.", e.Nick)
	default:
		conn.Privmsgf(channel, cannedResponse[rand.Intn(len(cannedResponse))], e.Nick)
	}
}

func maths(e *irc.Event) {
	conn := e.Connection
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
