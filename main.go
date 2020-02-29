package main

import (
	"discord-ncov/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type Config struct {
	Token     string `json:"token"`
	URI       string `json:"uri"`
	Command   string `json:"command"`
	ChannelID string `json:"ChannelId"`
	Interval  int    `json:"interval"`
	Mask      string `json:"mask"`
}

var c Config

var lastStatus model.Latest

func init() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened config.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &c)

	lastStatus = model.Latest{}
	lastStatus.Get("", c.URI)
}

func main() {
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + c.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	dg.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		discord.UpdateStatus(0, c.Command)
		guilds := discord.State.Guilds
		fmt.Println("Ready with", len(guilds), "guilds.")
	})

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	go heartBeat(dg)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func heartBeat(s *discordgo.Session) {
	result := model.Latest{}
	for range time.Tick(time.Minute * time.Duration(c.Interval)) {
		err := result.Get("", c.URI)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !lastStatus.Equals(result) {
			lastStatus = result
			s.ChannelMessageSend(c.ChannelID, formatMessage(result))
		}
	}
}

func formatMessage(result model.Latest) string {
	if result.Country == "" {
		return "No confirmed cases or country not found."
	}
	return fmt.Sprintf(c.Mask, result.Country, result.Confirmed, result.Deaths, result.Recovered)
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	content := m.Content
	PREFIX := c.Command
	if len(content) < len(PREFIX) {
		return
	}
	if content[:len(PREFIX)] != PREFIX {
		return
	}
	content = content[len(PREFIX):]
	args := strings.Fields(content)
	country := ""
	if len(args) > 0 {
		country = strings.Trim(content, " ")
	}
	result := model.Latest{}
	err := result.Get(country, c.URI)
	if err != nil {
		fmt.Println(err)
		return
	}
	s.ChannelMessageSend(m.ChannelID, formatMessage(result))
}
