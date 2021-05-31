package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	tb "gopkg.in/tucnak/telebot.v2"
)

func webmToMp4(in string, out string) error {
	return ffmpeg.Input(in, nil).Output(out, nil).Run()
}

func controllerLog(controllerId string, m *tb.Message) {
	log.Printf("[LOG]: User: %s | Controller: %s ", m.Sender.Username, controllerId)
}

type BotConf struct {
	TelegramToken string
	Store         string
}

func getConfig() (BotConf, error) {
	err := godotenv.Load()
	if err != nil {
		log.Print("No .env getting from actual env")
	}

	return BotConf{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		Store:         os.Getenv("STORE"),
	}, nil
}

func checkFolders(store string, folder string) {
	if _, err := os.Stat(store + folder); os.IsNotExist(err) {
		err = os.Mkdir(store+folder, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	log.Printf("Initializing bot\n")
	botConf, confErr := getConfig()
	if confErr != nil {
		log.Fatal("No config")
	}

	checkFolders(botConf.Store, "webm")
	checkFolders(botConf.Store, "mp4")

	b, err := tb.NewBot(tb.Settings{
		Token:  botConf.TelegramToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Bot ready")
	b.Handle("/help", func(m *tb.Message) {
		controllerLog("/help", m)
		b.Send(m.Sender, "Trade deal, I get a webm, you get an mp4..")
	})

	b.Handle(tb.OnDocument, func(m *tb.Message) {
		controllerLog("OnDocument", m)
		if m.Document.MIME != "video/webm" {
			b.Send(m.Sender, "Please send me a webm to convert")
		}

		webmFilename := botConf.Store + "webm/" + m.Document.FileID + ".webm"
		mp4Filename := botConf.Store + "mp4/" + m.Document.FileID + ".mp4"
		messageFilename := strings.TrimSuffix(m.Document.FileName, ".webm") + ".mp4"

		b.Send(m.Sender, "Downloading...")
		b.Download(&m.Document.File, webmFilename)

		b.Send(m.Sender, "Converting...")
		ffErr := webmToMp4(webmFilename, mp4Filename)
		if ffErr != nil {
			b.Send(m.Sender, "Internal server error")
			log.Printf("webm: %s, mp4: %s", webmFilename, mp4Filename)
			log.Fatalf("FFmpeg error: %s", ffErr)
		}

		b.Send(m.Sender, "Done")
		mp4 := &tb.Video{File: tb.FromDisk(mp4Filename), FileName: messageFilename}

		b.Send(m.Sender, "Uploading mp4...")
		b.Send(m.Sender, mp4)
		os.Remove(webmFilename)
		os.Remove(mp4Filename)
	})

	b.Start()
}
