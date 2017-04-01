package state

import (
    "io/ioutil"
    "sync"
    "log"
    "strings"
)

const channelFileName = "state_channels.txt";

var lock = sync.Mutex{}

func AddChannel(channel string) {
    lock.Lock()
    defer lock.Unlock()
    channels := readData()
    added := false
    for _, ch := range channels {
        if ch == channel {
            log.Println("already added " + channel)
            added = true
            break
        }
    }
    if !added {
        log.Println("adding " + channel)
        channels = append(channels, channel)
        writeData(channels)
    }
}

func RemoveChannel(channel string) {
    lock.Lock()
    defer lock.Unlock()
    channels := readData()
    for i, ch := range channels {
        if ch == channel {
            log.Println("removing " + channel)
            channels = append(channels[:i], channels[i+1:]...)
            writeData(channels)
            return
        }
    }
    log.Println("don't know about " + channel)
}

func GetChannels() []string {
    lock.Lock()
    defer lock.Unlock()
    return readData()
}

func readData() []string {
    channelsBytes, err := ioutil.ReadFile(channelFileName)
    if err != nil {
        log.Println("failed to get channels")
        log.Println(err)
        return []string{}
    }
    channels := strings.Split(string(channelsBytes), "\n")
    return channels
}

func writeData(channels []string) {
    channelBytes := []byte(strings.Join(channels, "\n"))
    ioutil.WriteFile(channelFileName, channelBytes, 0666)
}
