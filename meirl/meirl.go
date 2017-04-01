package main

import (
    "log"
    "io"
    "io/ioutil"
    "net"
    "bufio"
    "crypto/tls"
    "regexp"
    "time"
    "os"
    "strconv"
    "ircbot/state"
)

// const server = "chat.freenode.net"
const server = "irc.rizon.net"
const port = "6697"
const nickname = "testBot"
const ident = "test"
const realname = "testBot v 1.0"
var passwordBytes, _ = ioutil.ReadFile("password.txt")
var password = string(passwordBytes)

var pingRegex = regexp.MustCompile("^PING :(.*)$") // PING :message
var motdEndRegex = regexp.MustCompile("^[^ ]* (?:376|422)") // :server 376 nick :End of /MOTD command.
var vhostSetRegex = regexp.MustCompile("^[^ ]* 396 " + nickname) // :server 396 nick host :is now your visible host
var privmsgRegex = regexp.MustCompile("^:([^!]*)![^ ]* PRIVMSG ([^ ]*) :(.*)$") // :nick!ident@host PRIVMSG #channel :message
var inviteRegex = regexp.MustCompile("^([^ ]*) INVITE " + nickname + " :(.*)$") // :nick!ident@host INVITE nick :#channel
var kickRegex = regexp.MustCompile("^([^ ]*) KICK ([^ ]*) " + nickname + " :(.*)$") // :nick!ident@host KICK #channel nick :message
var banRegex = regexp.MustCompile("^[^ ]* 474 " + nickname + " ([^ ]*) :(.*)$") // :server 474 nick #channel :Cannot join channel (+b)

type dbOp struct {
    add bool
    channel string
}

func main() {
    log.Println("starting up");
    
    t := time.Now()
    ts := t.Format("Jan 2 2006 15-04-05 EDT")
    
    logFilename := "logs/" + ts + ".txt"
    f, err := os.Create(logFilename)
    if err != nil {
        log.Fatalln(err);
    }
    defer f.Close()
    // output to both stdout and the log file by default
    logWriter := io.MultiWriter(os.Stdout, f)
    log.SetOutput(logWriter)
    fileOnlyLogger := log.New(f, "", log.Flags())
    
    socket, err := tls.Dial("tcp", server + ":" + port, &tls.Config{})
    if err != nil {
        log.Fatalln(err);
    }
    defer socket.Close()
    log.Println("socket connected")
    
    errors := make(chan bool)
    linesToSend := make(chan string)
    dbWrites := make(chan dbOp)
    go readLines(socket, errors, linesToSend, fileOnlyLogger, dbWrites)
    go writeLines(socket, errors, linesToSend, fileOnlyLogger)
    go manualInput(linesToSend)
    go writeToDb(dbWrites)
    linesToSend <- "NICK " + nickname
    linesToSend <- "USER " + ident + " * 8 :" + realname
    <- errors
}

// ensure db writes are done sequentially
func writeToDb(dbWrites <-chan dbOp) {
    for {
        job := <-dbWrites
        if job.add {
            state.AddChannel(job.channel)
        } else {
            state.RemoveChannel(job.channel)
        }
    }
}

// listen to stdin and send any lines over the socket
func manualInput(linesToSend chan<- string) {
    reader := bufio.NewReader(os.Stdin)
    for {
        text, err := reader.ReadString('\n')
        if err != nil {
            log.Fatalln(err)
        }
        linesToSend <- text
    }
}

func readLines(socket net.Conn, errors chan<- bool, linesToSend chan<- string, logFile *log.Logger, dbWrites chan<- dbOp) {
    reader := bufio.NewReader(socket)
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            log.Fatalln(err)
            errors <- true
        }
        // remove the trailing \r\n
        line = line[:len(line)-2]
        logFile.Println(">>> " + line) // only print these to the log file, not to the default logger
        processLine(line, linesToSend, dbWrites)
    }
}

func writeLines(socket net.Conn, errors chan<- bool, linesToSend <-chan string, logFile *log.Logger) {
    writer := bufio.NewWriter(socket)
    for {
        line := <- linesToSend
        logFile.Println("<<< " + line) // only print these to the log file, not to the default logger
        _, err := writer.WriteString(line + "\n")
        if err != nil {
            log.Fatalln(err)
            errors <- true
        }
        // make sure it actually gets sent
        writer.Flush()
    }
}

func processLine(line string, linesToSend chan<- string, dbWrites chan<- dbOp) {
    if pingRegex.MatchString(line) {
        linesToSend <- "PONG :" + pingRegex.FindStringSubmatch(line)[1]
    } else if motdEndRegex.MatchString(line) {
        linesToSend <- "PRIVMSG nickserv :identify " + password
    } else if vhostSetRegex.MatchString(line) {
        log.Println("joining channels")
        go joinChannels(linesToSend)
    } else if privmsgRegex.MatchString(line) {
        matches := privmsgRegex.FindStringSubmatch(line)
        processPrivmsg(linesToSend, matches[1], matches[2], matches[3])
    } else if inviteRegex.MatchString(line) {
        matches := inviteRegex.FindStringSubmatch(line)
        log.Println("joining " + matches[2] + ", invited by " + matches[1])
        linesToSend <- "JOIN " + matches[2]
        dbWrites <- dbOp{true, matches[2]}
    } else if kickRegex.MatchString(line) {
        matches := kickRegex.FindStringSubmatch(line)
        log.Println("kicked from " + matches[2] + " by " + matches[1] + " because " + matches[3])
        dbWrites <- dbOp{false, matches[2]}
    } else if banRegex.MatchString(line) {
        matches := banRegex.FindStringSubmatch(line)
        log.Println("can't join " + matches[1] + " because " + matches[2])
        dbWrites <- dbOp{false, matches[2]}
    }
}

func joinChannels(linesToSend chan<- string) {
    for _, ch := range state.GetChannels() {
        linesToSend <- "JOIN " + ch
    }
}

var cooldown = make(map[string]int); // map of "#channel nick" to cooldown
var cuteImage = CuteImage{}

func processPrivmsg(linesToSend chan<- string, nick string, channel string, msg string) {
    chanNick := channel + " " + nick
    if cuteImage.checkForMatch(msg) {
        // reply to the channel or to a pm
        sendTo := ""
        if channel[:1] == "#" && cooldown[chanNick] == 0 {
            sendTo = channel
            cooldown[chanNick] = 5
            log.Println("cd for " + chanNick + " is " + strconv.Itoa(cooldown[chanNick]))
        } else if (channel == nickname) {
            // pm, reply to the nick
            sendTo = nick
        }
        
        if (sendTo != "") {
            go sendImage(linesToSend, sendTo, msg, cuteImage)
        }
    } else if (cooldown[chanNick] != 0) {
        cooldown[chanNick] -= 1
        log.Println("cd for " + chanNick + " is " + strconv.Itoa(cooldown[chanNick]))
        if cooldown[chanNick] == 0 {
            delete(cooldown, chanNick)
        }
    }
}

func sendImage(linesToSend chan<- string, sendTo string, msg string, img CuteImage) {
    str, url := img.getImageForMessage(msg);
    if url != "" {
        newMsg := "PRIVMSG " + sendTo + " " + ":" + str + " " + url
        log.Println("sending image: " + newMsg);
        linesToSend <- newMsg
    }
}
