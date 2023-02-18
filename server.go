package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var location chan string = make(chan string)
var need_location chan string = make(chan string)

func main() {
	//starting server
	go startServer()

	//starting bot
	go startBot()

	for {

	}
}

func startBot() {
	bot, _ := tgbotapi.NewBotAPI("5862822125:AAEI_yHy6LT4ythfDxbbALgO9Tainf4oOrw")
	bot.Debug = false
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)
	// В канал updates будут приходить все новые сообщения.
	for update := range updates {
		//обработать сообщение
		go func() {
			fmt.Print(update.Message.Text)
			need_location <- "need"
			newLocation := strings.Split(<-location, "-")
			lati, _ := strconv.ParseFloat(newLocation[0], 32)
			longi, _ := strconv.ParseFloat(newLocation[1], 32)
			msg := tgbotapi.NewLocation(update.Message.Chat.ID, longi, lati)
			bot.Send(msg)
		}()
	}
}

func startServer() {
	// server creation
	dstream, err := net.Listen("tcp", ":30391")
	if err != nil {
		return
	}
	defer dstream.Close()

	// handle new connections in a loop
	for {
		// accept new connection
		con, err := dstream.Accept()
		if err != nil {
			return
		}
		// procced connection above in separated virtual thread
		go handle(con)
	}
}

func handle(con net.Conn) {
	defer con.Close()
	// create new channel to send replies
	replyChan := make(chan string)
	// get new client requests in loop in new thread
	go func() {
		for {
			buf := make([]byte, 128)
			rlen, err := con.Read(buf) // get request
			//error check
			if err != nil {
				return
			}
			// send request pack to parse it via function
			go parseRequest(string(buf[:rlen]), replyChan)
		}
	}()

	//sending replies to client in the loop
	for {
		select {
		case <-need_location:
			con.Write([]byte("GIVEMELOCATION#"))
		case reply := <-replyChan:
			location <- reply
		}
	}
}

// Requests may come together
// so we need to split it to single ones
func parseRequest(request string, replyChan chan string) {
	//requests are separated with '#'
	requestList := strings.Split(request, "#")
	for _, singleRequest := range requestList {
		if singleRequest != "" {
			replyChan <- singleRequest
		}
	}
}
