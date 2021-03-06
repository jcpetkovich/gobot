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
var expPattern = re.MustCompile(fmt.Sprintf("%s:(.*)=", nick))
var languagePattern = re.MustCompile("(?i)(?:fuck|shit|damn)")
var apologyPattern = re.MustCompile("(?i)(?:sorry|my bad)")
var helpPattern = re.MustCompile("(?i)(?:help|what.*you do)")
var thanksPattern = re.MustCompile("(?i)(?:thanks|cool|awesome)")
var coffeePattern = re.MustCompile("(?i)(?:coffee|caffeine|tea)")
var godPattern = re.MustCompile("(?i)(?:where|what).*god")

var cannedResponse = []string{
	"%s: Sorry, what's that?",
	"%s: Uhhh, what?",
	"%s: K, wait, what?",
	"%s: I'm not sure what you mean...",
	"%s: I am reasonably certain that I don't understand that.",
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
	case godPattern.MatchString(e.Message()):
		conn.Privmsgf(channel, "%s: God? What God?", e.Nick)
	default:
		conn.Privmsgf(channel, cannedResponse[rand.Intn(len(cannedResponse))], e.Nick)
	}
}

func maths(e *irc.Event) {
	conn := e.Connection
	expression := expPattern.FindStringSubmatch(e.Message())[1]
	wordPattern, _ := re.Compile("[a-zA-Z']+")
	expression = wordPattern.ReplaceAllString(expression, "")
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
		conn.Privmsgf(channel, "%s: %s, yup, pretty sure.", e.Nick, result)
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
	conn := irc.IRC(nick, nick)
	conn.Password = password
	conn.Connect(fmt.Sprintf("%s:%s", host, port))
	conn.AddCallback("001", func(e *irc.Event) { conn.Join("#gentoo") })
	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		if toGoBot, _ := re.MatchString(fmt.Sprintf("%s:", nick), e.Message()); toGoBot {
			dispatch(e)
		}
	})
	conn.Loop()
}
