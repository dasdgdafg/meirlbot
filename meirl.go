package main

import (
	"github.com/dasdgdafg/ircFramework"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"
)

const server = "irc.rizon.net"
const port = "6697"
const nickname = "meirlBot"
const ident = "meirl"
const realname = "a bot to post pic of yourself irl"

var passwordBytes, _ = ioutil.ReadFile("password.txt")
var password = string(passwordBytes)

func main() {
	rand.Seed(time.Now().UnixNano())

	bot := ircFramework.IRCBot{Server: server,
		Port:           port,
		Nickname:       nickname,
		Ident:          ident,
		Realname:       realname,
		Password:       password,
		ListenToStdin:  true,
		MessageHandler: processPrivmsg,
	}
	bot.Run()
}

var cooldown = make(map[string]int) // map of "#channel nick" to cooldown
var cuteImage = CuteImage{}

func processPrivmsg(linesToSend chan<- string, nick string, channel string, msg string) {
	chanNick := channel + " " + nick
	if cuteImage.checkForMatch(msg) {
		// reply to the channel or to a pm
		sendTo := ""
		if channel[:1] == "#" {
			if cooldown[chanNick] == 0 {
				sendTo = channel
				cooldown[chanNick] = 5
				log.Println("cd for " + chanNick + " is " + strconv.Itoa(cooldown[chanNick]))
			} else {
				newMsg := "NOTICE " + nick + " " + ":too hayai"
				log.Println("sending notice: " + newMsg)
				linesToSend <- newMsg
			}
		} else if channel == nickname {
			// pm, reply to the nick
			sendTo = nick
		}

		if sendTo != "" {
			go sendImage(linesToSend, sendTo, msg, nick, cuteImage)
		}
	} else if cooldown[chanNick] != 0 {
		cooldown[chanNick] -= 1
		log.Println("cd for " + chanNick + " is " + strconv.Itoa(cooldown[chanNick]))
		if cooldown[chanNick] == 0 {
			delete(cooldown, chanNick)
		}
	}
}

func sendImage(linesToSend chan<- string, sendTo string, msg string, nick string, img CuteImage) {
	str, url := img.getImageForMessage(msg, nick)
	if url == "" {
		newMsg := "PRIVMSG " + sendTo + " " + ":couldn't find any images"
		log.Println("sending message: " + newMsg)
		linesToSend <- newMsg
	} else {
		newMsg := "PRIVMSG " + sendTo + " " + ":" + str + " " + url
		log.Println("sending image: " + newMsg)
		linesToSend <- newMsg
	}
}
