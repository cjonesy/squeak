package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/marcusolsson/tui-go"

	"github.com/segmentio/kafka-go"
)

func getKafkaReader(topic string, brokers []string) *kafka.Reader {
	host, _ := os.Hostname()

	kafkaReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        host,
		Topic:          topic,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
	})

	return kafkaReader
}

func updateMessages(topic string, brokers []string, history *tui.Box, ui tui.UI) {

	reader := getKafkaReader(topic, brokers)

	for {
		msg, err := reader.ReadMessage(context.Background())

		if err != nil {
			log.Printf("ERROR: %v", err)

		}

		if msg.Value != nil {
			history.Append(tui.NewHBox(
				tui.NewLabel(msg.Time.Local().String()),
				tui.NewPadder(1, 0, tui.NewLabel(fmt.Sprintf("<%s>", "anon"))),
				tui.NewLabel(string(msg.Value)),
				tui.NewSpacer(),
			))
		}
		ui.Repaint()
	}
}

func sendMessage(topic string, brokers []string, message string) {
	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:        brokers,
		Topic:          topic,
	})

	msg := kafka.Message{
		Value: []byte(message),
	}

	err := kafkaWriter.WriteMessages(context.Background(), msg)
	if err != nil {
		log.Printf("ERROR: %v", err)

	}
}

func main() {
	args := os.Args[1:]
	broker := args[0]
	topic := args[1]

	brokers := []string{broker}

	history := tui.NewVBox()

	history.Append(tui.NewHBox(
		tui.NewLabel("Welcome to Squeak! Press 'Esc' to exit..."),
		tui.NewSpacer(),
	))

	historyScroll := tui.NewScrollArea(history)
	historyScroll.SetAutoscrollToBottom(true)

	historyBox := tui.NewVBox(historyScroll)
	historyBox.SetBorder(true)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(historyBox, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)

	input.OnSubmit(func(e *tui.Entry) {
		sendMessage(topic, brokers, e.Text())
		input.SetText("")
	})

	root := tui.NewHBox(chat)

	ui, err := tui.New(root)
	if err != nil {
		log.Fatal(err)
	}

	ui.SetKeybinding("Esc", func() { ui.Quit() })

	go updateMessages(topic, brokers, history, ui)

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}

}

