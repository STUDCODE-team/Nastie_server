package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var db *sql.DB

func ConnectToDatabase() {
	var err error
	db, err = sql.Open("postgres", "host=127.0.0.1 user=postgres password=postgres dbname=nastie port=5432")
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to database")
}

func CreateTables() {
	_, err := db.Exec(`CREATE TABLE users (id SERIAL PRIMARY KEY, tg_id INTEGER NOT NULL);
					CREATE TABLE tokens (user_id INTEGER REFERENCES users(id), token VARCHAR NOT NULL)`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Created tables")
}

func AddUser(tg_id int) {
	_, err := db.Exec("INSERT INTO users(tg_id) VALUES($1) ON CONFLICT DO NOTHING", tg_id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Authorized user:", tg_id)
}

func AddToken(id int, token string) {
	_, err := db.Exec("INSERT INTO tokens(user_id, token) SELECT id, $1 FROM users WHERE tg_id=$2", token, id)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("User", id, "add a token: ", token)
}

func GetHosts(id int) []string {
	// TODO: надо сделать map с данными хоста, чтобы надпись на кнопке была не токеном а ФИО
	var tokens []string
	var token string
	rows, err := db.Query("SELECT token FROM tokens WHERE user_id=(SELECT id FROM users WHERE tg_id=$1 LIMIT 1)", id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&token)
		if err != nil {
			log.Fatal(err)
		}
		tokens = append(tokens, token)
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(tokens)
	return tokens
}

func GetButtonHosts(id int) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton
	hosts := GetHosts(id)

	for _, v := range hosts {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(v, v))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons)
}

func GetHostLocation(token string) {
	// TODO: здесь надо бы получить геопозицю хоста и отослать клиенту

	fmt.Println(token)
}

func TelegramBot() {
	bot, _ := tgbotapi.NewBotAPI("5401428277:AAHAbVaEoKnvPR4IQr0slOu9x2jVLrtTa54")
	bot.Debug = false

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "start":
				AddUser(update.Message.From.ID)
				msg.ReplyMarkup = GetButtonHosts(update.Message.From.ID)
				msg.Text = "Добро пожаловать!"
				bot.Send(msg)
			case "token":
				token := update.Message.CommandArguments()
				if token == "" {
					msg.Text = "Используйте /token [token]"
					bot.Send(msg)
				} else {
					AddToken(update.Message.From.ID, update.Message.CommandArguments())
				}
			default:
				msg.Text = "Неизвестная команда"
				bot.Send(msg)
			}
		} else if update.CallbackQuery != nil {
			GetHostLocation(update.CallbackQuery.Data)
			bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		}
	}
}

func main() {
	ConnectToDatabase()
	defer db.Close()
	// CreateTables() // Если надо создать таблицы

	go TelegramBot()

	for {

	}
}
