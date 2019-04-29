package radau

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	dg "github.com/icechair/discordgo"
)

//Radau is the Bot Monster
type Radau struct {
	BotToken    string
	ClientID    string
	Permissions int
	discord     *dg.Session
}

// NewRadau creates the Bot Instance
func NewRadau(botToken string, clientID string, permissions int) Radau {
	r := Radau{}
	r.BotToken = botToken
	r.Permissions = permissions
	r.ClientID = clientID
	return r
}

//Start registers a Wakeup handler to connect the Bot to the Discord Gateway
func (r Radau) Start(listen string) {
	err := loadSound()
	if err != nil {
		log.Fatalf("Error loading sound: %+v", err)
		return
	}
	http.HandleFunc("/wakeup", func(w http.ResponseWriter, req *http.Request) {
		if r.discord != nil {
			fmt.Fprintf(w, "im already awake!")
			return
		}
		d, err := dg.New("Bot " + r.BotToken)
		if err != nil {
			log.Fatal(err)
		}
		d.AddHandler(ready)
		d.AddHandler(messageCreate)
		d.AddHandler(guildCreate)

		err = d.Open()
		if err != nil {
			log.Fatal(err)
		}
		r.discord = d
		fmt.Fprintf(w, "woke!")
	})

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	go func() {
		sig := <-sigs
		log.Printf("caught signal: %+v\n", sig)
		if r.discord != nil {
			log.Printf("stopping discord")
			err := r.discord.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
		os.Exit(0)
	}()
	log.Fatal(http.ListenAndServe(listen, nil))
	log.Printf("started listening on %+v", listen)
}

func ready(s *dg.Session, event *dg.Ready) {
	log.Printf("ready %+v, %+v", s, event)
	s.UpdateStatus(0, "")
}

func messageCreate(s *dg.Session, m *dg.MessageCreate) {
	log.Printf("messageCreate %+v, %+v", s, m)
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, "!airhorn") {
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			log.Printf("coudn't find Channel: %+v", err)
			return
		}

		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			log.Printf("coudn't find Guild: %+v", err)
			return
		}
		_, _ = s.ChannelMessageSend(m.ChannelID, "let me try")
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = playSound(s, g.ID, vs.ChannelID)
				if err != nil {
					fmt.Println("Error playing sound:", err)
				}

				return
			}
		}
	}
}

func guildCreate(s *dg.Session, event *dg.GuildCreate) {
	log.Printf("guildCreate %+v, %+v", s, event)
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "HI! i'm ready")
		}
	}
}

var buffer = make([][]byte, 0)

func loadSound() error {

	file, err := os.Open("airhorn.dca")
	if err != nil {
		fmt.Println("Error opening dca file :", err)
		return err
	}

	var opuslen int16

	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opuslen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}

		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opuslen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)

		// Should not be any end of file errors
		if err != nil {
			fmt.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

func playSound(s *dg.Session, guildID, channelID string) (err error) {

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()

	return nil
}
