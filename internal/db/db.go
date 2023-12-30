package db

import (
	"database/sql"
	"fmt"
	"level_zero/config"
	"level_zero/internal/models"
	"log"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func InsertOrder(conn *sql.DB, order models.Order) (error) {
	err := conn.Ping()
	if err != nil {
		return err
	}

	var orderId int
	orderQuery := `INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	RETURNING id`
	err = conn.QueryRow(orderQuery, order.OrderUid, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId, order.DateCreated, order.OofShard).Scan(&orderId)
	if err != nil {
		log.Print("Order insertion error: ", err)
	}
	
	deliveryQuery := `INSERT INTO deliveries (id, name, phone, zip, city, address, region, email)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err = conn.Exec(deliveryQuery, orderId, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Adress, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		log.Print("Delivery insertion error")
		return err
	}

	paymentQuery := `INSERT INTO payments (id, transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err = conn.Exec(paymentQuery, orderId, order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank, order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		log.Print("Payment insertion error")
		return err
	}

	itemQuery := `INSERT INTO items (id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for _, v := range order.Items {
		_, err = conn.Exec(itemQuery, orderId, v.ChrtId, v.TrackNumber, v.Price, v.Rid, v.Name, v.Sale, v.Size, v.TotalPrice, v.NmId, v.Brand, v.Status)
		if err != nil {
			log.Print("Items insertion error")
			return err
		}
	}

	return nil
}

func getOrderItems (conn *sql.DB, id string) ([]models.Item, error) {
	itemQuery := `SELECT chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status
	FROM items
	WHERE id = $1`
	var items []models.Item
	rows, err := conn.Query(itemQuery, id)
	if err != nil {
		return items, err
	}
	defer rows.Close()

	for rows.Next() {
		item := models.Item{}

        err := rows.Scan(&item.ChrtId, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale, &item.Size, &item.TotalPrice, &item.NmId, &item.Brand, &item.Status)
		if err != nil {
			return items, err
		}

		items = append(items, item)
    }

	err = rows.Err()
	if err != nil {
		return items, err
	}

    return items, nil

}

func GetAllOrders (conn *sql.DB) ([]models.Order, error) {
	allOrdersQuery := `SELECT o.id, o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature, o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
	d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
	p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt, p.bank, p.delivery_cost, p.goods_total, p.custom_fee
	FROM orders AS o
	LEFT JOIN deliveries AS d ON o.id = d.id
	LEFT JOIN payments AS p ON o.id = p.id`
	
	rows, err := conn.Query(allOrdersQuery)
	if err != nil {
		log.Print("Error getting all orders: ", err)
	}
	defer rows.Close()

	var orders []models.Order

	for rows.Next() {
		order := models.Order{}
        delivery := models.Delivery{}
		payment := models.Payment{}
		var id int

        err := rows.Scan(&id, &order.OrderUid, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey, &order.SmId, &order.DateCreated, &order.OofShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Adress, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestId, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee)
		if err != nil {
			return orders, err
		}

		order.Delivery = delivery
		order.Payment = payment
		items, err := getOrderItems(conn, strconv.Itoa(id))
		order.Items = items
		orders = append(orders, order)
    }

	err = rows.Err()
	if err != nil {
		return orders, err
	}

    return orders, nil
}

func InitialzeDatabase(config config.Config, schemaScriptPath string) (*sql.DB, error) {
	dsn := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", config.DbUser, config.DbPassword, config.DbName)
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	schemaScript, err := os.ReadFile(schemaScriptPath)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(string(schemaScript))
	if err != nil {
		return nil, err
	}

	return conn, nil
}