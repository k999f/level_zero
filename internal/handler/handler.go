package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"level_zero/internal/cache"
	"level_zero/internal/db"
	"level_zero/internal/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nats-io/stan.go"
)

func IndexHandler (w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFiles("web/index.html")
	if err != nil {
		log.Println("Parsing index page error: ", err)
		http.Error(w, "Index page not found", 404)
		return
	}
	err = ts.Execute(w, nil)
	if err != nil {
		log.Println("Executing index page error: ", err)
		http.Error(w, "Index page not found", 404)
	}
}

func OrderHandler(c cache.Cache) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
    	uid := vars["uid"]
		order, ok := c.GetCache(uid)

		ts, err := template.ParseFiles("web/order.html")
		if err != nil {
			log.Println("Parsing order page error: ", err)
			http.Error(w, "Order page not found", 404)
			return
		}
		if ok {
			err = ts.Execute(w, order)
			if err != nil {
				log.Println("Executing order page error: ", err)
				http.Error(w, "Order page not found", 404)
			}
		} else {
			fmt.Fprint(w, "Order not found")
		}
    }
}

func MessageHandler(cache cache.Cache, conn *sql.DB) stan.MsgHandler {
    return func(m *stan.Msg) {
		order := models.Order{}

		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			log.Print("Error marshaling order: trash received")
		} else {
			fmt.Println("Message handled, order UID: ", order.OrderUid)
			_, ok := cache.GetCache(order.OrderUid)
			if ok {
				fmt.Println("Order with UID ", order.OrderUid, " already exists")
			} else {
				err = db.InsertOrder(conn, order)
				if err != nil {
					log.Print("Inserting order error: ", err)
				}
				cache.AddCache(order)
				fmt.Println("Orders in cache: ", len(cache.Data))	
			}
		}
    }
}