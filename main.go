package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

type bnResp struct {
	Price float64 `json:"price,string,omitempty"`
	Code  int64   `json:"code"`
}

func main() {
	bot, err := tgbotapi.NewBotAPI("5690516392:AAFBZbqHzs76dcc2JDFkpsZrJmOqWfLfDJs")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	connStr := "user=postgres password=adletloh695 dbname=db sslmode=disable"

	for update := range updates {
		if update.Message == nil { // If we got a message
			continue
		}
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		price, _ := getPrice(update.Message.Text)
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		}
		defer db.Close()
		result, err := db.Exec("insert into crytpocoin (crypto) values ($1)", update.Message.Text)
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		}
		fmt.Println(result.LastInsertId()) // не поддерживается
		fmt.Println(result.RowsAffected()) // количество добавленных строк
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("%f", price))

		bot.Send(msg)
	}
}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		fmt.Println("Arman", err, "bot")
		return
	}
	if jsonResp.Code != 0 {
		err = errors.New("Неверный символ")
	}
	price = jsonResp.Price

	return
}
