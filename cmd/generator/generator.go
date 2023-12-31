package main

import (
	"encoding/json"
	"fmt"
	"level_zero/config"
	"level_zero/internal/models"
	"level_zero/internal/util"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"

	"github.com/nats-io/stan.go"
)

const (
	messagesNumber = 3
	messagesInterval = 3
	maxOrderItems = 3
)

func randomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	symbols := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	str := make([]rune, n)
	for i := range(str) {
		str[i] = symbols[rand.Intn(len(symbols))]
	}
	return string(str)
}

func randomOrder() models.Order {
	order := models.Order{}
	items := []models.Item{}
	delivery := models.Delivery{}
	payment := models.Payment{}

	order.OrderUid = strconv.Itoa(gofakeit.Number(1, 10000))
	order.TrackNumber = fmt.Sprint("TRACK", gofakeit.Number(1, 1000))
	order.Entry = fmt.Sprint("ENTRY", gofakeit.Number(1, 100))
	order.Locale = fmt.Sprint("LOCALE", gofakeit.Number(1, 100))
	order.InternalSignature = fmt.Sprint("SIGNATURE", gofakeit.Number(1, 100))
	order.CustomerId = strconv.Itoa(gofakeit.Number(1, 10000))
	order.DeliveryService = fmt.Sprint("Service number ", gofakeit.Number(1, 100))
	order.Shardkey = strconv.Itoa(gofakeit.Number(1, 100))
	order.SmId = gofakeit.Number(1, 1000)
	order.DateCreated = time.Now()
	order.OofShard = strconv.Itoa(gofakeit.Number(1, 100))

	delivery.Name = gofakeit.Name()
	delivery.Phone = gofakeit.Phone()
	delivery.Zip = gofakeit.Zip()
	delivery.City = gofakeit.City()
	delivery.Adress = fmt.Sprint("Street number ", gofakeit.Number(1, 100))
	delivery.Region = fmt.Sprint("Region number ", gofakeit.Number(1, 100))
	delivery.Email = gofakeit.Email()

	itemsAmount := gofakeit.Number(1, maxOrderItems)
	for i := 0; i < itemsAmount; i++ {
		item := models.Item{}
		item.ChrtId = gofakeit.Number(1, 10000)
		item.TrackNumber = fmt.Sprint("TRACK", gofakeit.Number(1, 1000))
		item.Price = gofakeit.Number(1, 10000)
		item.Rid = randomString(21)
		item.Name = fmt.Sprint("Item number ", gofakeit.Number(1, 1000))
		item.Sale = gofakeit.Number(1, 50)
		item.Size = strconv.Itoa(gofakeit.Number(10, 40))
		item.TotalPrice = int(float64(item.Price/100) * float64(item.Sale)) 
		item.NmId = gofakeit.Number(1, 10000)
		item.Brand = fmt.Sprint("Brand number ", gofakeit.Number(1, 1000))
		item.Status = gofakeit.Number(1, 300)
		items = append(items, item)
	}

	payment.Transaction = randomString(19)
	payment.RequestId = strconv.Itoa(gofakeit.Number(1, 1000))
	payment.Currency = gofakeit.CurrencyShort()
	payment.Provider = fmt.Sprint("PROVIDER", gofakeit.Number(1, 100))

	paymentAmount := 0
	for _, v := range items {
		paymentAmount += v.TotalPrice
	}

	payment.Amount = paymentAmount
	payment.PaymentDt = gofakeit.Number(1, 10000)
	payment.Bank = fmt.Sprint("Bank number ", gofakeit.Number(1, 100))
	payment.DeliveryCost = gofakeit.Number(500, 3000)
	payment.GoodsTotal = len(items)
	payment.CustomFee = int(float64(payment.Amount) * 0.15)

	order.Delivery = delivery
	order.Payment = payment
	order.Items = items
	return order
}

func main() {
	configPath := "config/config.json"
	config := config.InitialzeConfig(configPath)
	
	natsUrl := fmt.Sprintf("nats://%s", config.NatsUrl)
	nc, err := stan.Connect(config.NatsCluster, "generator", stan.NatsURL(natsUrl))
	defer nc.Close()

	if err != nil {
		log.Fatal("Nats connection error: ", err)
	}

	for i := 0; i < messagesNumber; i++ {
		randomOrder := randomOrder()
		message, err := json.Marshal(randomOrder)
		err = nc.Publish(config.NatsSubject, message)
		if err != nil {
			log.Print("Error while sending message", i, ": ", err)
		} else {
			fmt.Println("Order", randomOrder.OrderUid, "was sent")
			util.PrintOrder(randomOrder)
		}

		time.Sleep(messagesInterval * time.Second)
	}
	
	err = nc.Publish(config.NatsSubject, []byte("Some trash"))
	if err != nil {
		log.Print("Error while sending trash: ", err)
	} else {
		fmt.Println("Trash was sent")
	}
}