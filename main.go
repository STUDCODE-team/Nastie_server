package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"

	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var db *sql.DB

// Это для кнопок с хостами
const (
	Follow = "follow"
	Remove = "remove"
)

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
	_, err := db.Exec(`CREATE TABLE users (id SERIAL PRIMARY KEY, tg_id INTEGER NOT NULL UNIQUE);
					CREATE TABLE tokens (user_id INTEGER REFERENCES users(id), token VARCHAR NOT NULL);
					CREATE TABLE hosts (host_id SERIAL PRIMARY KEY, token VARCHAR, name VARCHAR)`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Tables was created")
}

func AddUser(tg_id int) {
	// TOOD: поискать sql запрос для добавления несуществующей записи в таблицу
	var user string
	err := db.QueryRow("SELECT * FROM users WHERE tg_id=$1", tg_id).Scan(&user)
	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO users(tg_id) VALUES($1)", tg_id)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("New user:", tg_id)
	}
}

func AddToken(id int, token string) bool {
	var row string
	err := db.QueryRow("SELECT token FROM hosts WHERE token=$1", token).Scan(row)
	if err != sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO tokens(user_id, token) SELECT id, $1 FROM users WHERE tg_id=$2;", token, id)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("User", id, "add a token:", token)
		return true
	}
	return false
}

func RemoveToken(id int, token string) {
	_, err := db.Exec("DELETE FROM tokens WHERE user_id=(SELECT id FROM users WHERE tg_id=$1 LIMIT 1) AND token=$2", id, token)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("User", id, "delete token:", token)
}

func GetHosts(id int) map[string]string {
	// TODO: надо сделать map с данными хоста, чтобы надпись на кнопке была не токеном а ФИО
	tokens := make(map[string]string)
	var token, name string
	rows, err := db.Query(`SELECT t.token, h.name FROM tokens t JOIN hosts h ON t.token=h.token 
					WHERE t.user_id=(SELECT id FROM users WHERE tg_id=$1 LIMIT 1)`, id)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&token, &name)
		if err != nil {
			log.Fatal(err)
		}
		tokens[token] = name
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
	return tokens
}

func GetButtonHosts(id int, param string) tgbotapi.InlineKeyboardMarkup {
	var buttons []tgbotapi.InlineKeyboardButton
	hosts := GetHosts(id)

	for token, name := range hosts {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardButtonData(name, token+":"+param))
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons)
}

func GetHostLocation(token string) string {
	// TODO: здесь надо бы получить геопозицю хоста и отослать клиенту

	fmt.Println(token)

	return "Александр Плешаков в данный момент сосет хуй"
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
				if len(GetHosts(update.Message.From.ID)) == 0 {
					msg.Text = "Добро пожаловать!\n\nИспользуйте /add [token] для добавления отслеживаемых целей"
				} else {
					msg.Text = "С возвращением!"
					msg.ReplyMarkup = GetButtonHosts(update.Message.From.ID, Follow)
				}
				bot.Send(msg)
			case "add":
				token := update.Message.CommandArguments()
				if token == "" {
					msg.Text = "Используйте /add [token]"
				} else if !AddToken(update.Message.From.ID, update.Message.CommandArguments()) {
					msg.Text = "Такого токена не существует!"
				} else {
					msg.Text = "Токен успешно добавлен!"
				}
				bot.Send(msg)
			case "remove":
				msg.Text = "Выберите того, кого хотите удалить"
				msg.ReplyMarkup = GetButtonHosts(update.Message.From.ID, Remove)
				bot.Send(msg)
			default:
				msg.Text = "Неизвестная команда"
				bot.Send(msg)
			}
		} else if update.CallbackQuery != nil {
			params := strings.Split(update.CallbackQuery.Data, ":") // token:param
			if params[1] == Follow {
				// TOOD: здесь надо получить и вывести в чат местоположение хостаы
				var location string
				if update.CallbackQuery.From.ID == 4543937681 { // Саня попросил
					bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Ало блять, харе жмакать, видишь же не работает"))
				} else {
					location = GetHostLocation(params[0])
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, location)
					bot.Send(msg)
					bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
				}
			} else if params[1] == Remove {
				RemoveToken(update.CallbackQuery.From.ID, params[0])
				bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, "Удален!"))
			}
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
