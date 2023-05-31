package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

type bnResp struct {
	Price float64 `json:"price,string,omitempty"`
	Code  int64   `json:"code"`
}

type coin struct {
	id     int
	crypto string
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
	connStr := "user=postgres password= dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	defer db.Close()
	for update := range updates {
		if update.Message == nil { // If we got a message
			continue
		}
		command := strings.Split(update.Message.Text, " ")
		switch command[0] {
		case "ADD":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			result, err := db.Exec("insert into cryptocoin (crypto) values ($1)", command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			fmt.Println(result.LastInsertId())
			fmt.Println(result.RowsAffected())

		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			result, err := db.Exec("delete from cryptocoin where crypto = $1", command[1])
			fmt.Println(result.RowsAffected()) // количество удаленных строк
			if err != nil {
				panic(err)
			}
		case "START":
			msg := ""
			if len(command) != 1 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная команда"))
			}
			crypto, err := db.Query("select * from cryptocoin")
			if err != nil {
				panic(err)
			}
			coins := []coin{}
			for crypto.Next() {
				p := coin{}
				err := crypto.Scan(&p.id, &p.crypto)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(err)
				coins = append(coins, p)
			}
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			for _, p := range coins {
				price, _ := getPrice(p.crypto)
				msg += fmt.Sprintf("%s: %f\n", p.crypto, price)
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
		}
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
		err = errors.New("�������� ������")
	}
	price = jsonResp.Price

	return
}
