package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mqtt/database"
	gpssocket "mqtt/gps-socket"
	"mqtt/helper"
	"mqtt/models"
	"mqtt/mqtt"
	"time"

	"github.com/joho/godotenv"
	"github.com/segmentio/kafka-go"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
)

type Message struct {
	ID      string      `json:"id"`
	Content interface{} `json:"content"`
}

const (
	KafkaBrokersBinary = "43.231.115.108:9092" //43.231.115.108 localhost
	KafkaBrokers = "103.50.206.3:9092" //43.231.115.108 localhost
	Topic        = "mq-gps-counter"
	Topic2        = "mq"
	Topic3        = "mq2"
)

var kafkaWriter *kafka.Writer
var kafkaWriter2 *kafka.Writer
var kafkaWriter3 *kafka.Writer

var DB *gorm.DB

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	database.MustConnect()
	initRoute()

	channel, err := mqtt.InitializeRabbitMQ()
	if err != nil {
		log.Fatal("Error initializing RabbitMQ:", err)
	}

	if err := initKafkaWriter1(); err != nil {
		log.Fatal("Error initializing Kafka writer:", err)
	}
	defer kafkaWriter.Close()

	go gpssocket.SocketInit()
	defer channel.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumeMessages(ctx, channel)
	helper.WaitForShutdown()
}

func consumeMessages(ctx context.Context, channel *amqp.Channel) {
	msgs, err := channel.Consume(
		"ict.source",
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Error consuming messages from queue:", err)
	}

	log.Println("Waiting for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down consumer...")
			return
		case msg := <-msgs:
			processMessage(msg)
		}
	}
}

var batch0202 []models.Content0202
var batch0230 []models.Content0230
var batch0800 []models.Content0800
var m map[string]string

func initRoute() {
	m = make(map[string]string)
	routes := make([]models.Route, 0)
	db := database.DB

	tx := db.Model(&models.Route{}).Order("id desc")
	routes = make([]models.Route, 0)

	err := tx.Select("*").Find(&routes).Error
	if err != nil {
		log.Fatal(err)
	}

	for _, route := range routes {
		m[route.RouteCode] = route.RouteAlias
	}
}

var batchSize = 1

func processMessage(msg amqp.Delivery) {
	newMessage, _ := helper.RemoveState2(string(msg.Body))

	var message Message

	for attempts := 1; attempts <= 3; attempts++ {
		err := json.Unmarshal([]byte(newMessage), &message)
		if err != nil {
			log.Printf("Error decoding message (attempt %d): %v\n", attempts, err)
			if attempts < 3 {
				time.Sleep(10 * time.Millisecond) // Wait before retrying
				continue
			} else {
				log.Println("Max retry attempts reached, skipping message")
				fmt.Println(newMessage)
				fmt.Println(string(msg.Body))

				var data models.FailedData

				data.Message = string(msg.Body)

				if err := database.DB.Create(data).Error; err != nil {
					database.DB.Rollback()
					log.Println("saving failed data error: ", err)
				}
				return
			}
		}
		break // Decoding successful, exit loop
	}

	switch content := message.Content.(type) {
	case string:
		// Content is already a string, continue processing
	case map[string]interface{}:
		// Convert map to JSON string
		contentBytes, err := json.Marshal(content)
		if err != nil {
			log.Println("Error converting content to string:", err)
			return
		}
		message.Content = string(contentBytes)
	default:
		log.Println("Unsupported content type")
		return
	}

	switch message.ID {
	/*
		case "0200":
			var content0200 content.Content0200
			err := json.Unmarshal([]byte(message.Content.(string)), &content0200)
			if err != nil {
				log.Println("Error decoding content for ID 0200:", err)
				return
			}
			content0200.State, err = helper.ParseStateData(stateData)
			if err != nil {
				log.Println("Error parsing state data:", err)
				return
			}

			data := models.Content0200{
				Alarm:     helper.ArrayToString(content0200.Alarm),
				Dir:       content0200.Dir,
				GPSTime:   content0200.GPSTime,
				Latitude:  content0200.Latitude,
				Longitude: content0200.Longitude,
				Mileage:   content0200.Mileage,
				OnlineNo:  content0200.OnlineNo,
				RouteCode: content0200.RouteCode,
				Speed:     content0200.Speed,
				StaIndex:  content0200.StaIndex,
				State:     helper.MapToString(content0200.State),
				UpDown:    content0200.UpDown,
			}
			// Add data to batchMessages
			batch0200 = append(batch0200, data)
	*/

	
		case "0200":
			go func() {
				var content0200 models.Content0200
				err := json.Unmarshal([]byte(message.Content.(string)), &content0200)
				if err != nil {
					log.Println("Error decoding content for ID 0200:", err)
					return
				}
				// content0200.State, err = helper.ParseStateData(stateData)
				if err != nil {
					log.Println("Error parsing state data:", err)
					return
				}

				RouteName, ok := m[content0200.RouteCode]
				if !ok {
					RouteName = "null"
					fmt.Println(content0200.RouteCode)
				}

				data := models.ImapKafka{
					OnlineNo:  helper.StringToUint(content0200.OnlineNo),
					Azimuth:   content0200.Dir,
					RouteCode: helper.StringToUint(content0200.RouteCode),
					RouteName: RouteName,
					Speed:     content0200.Speed,
					Longitude: content0200.Longitude,
					Latitude:  content0200.Latitude,
				}

				contentJSON, err := json.Marshal(data)
				if err != nil {
					log.Println("Error encoding content0200 to JSON:", err)
					return
				}
				fmt.Println(contentJSON)

				err = kafkaWriter2.WriteMessages(context.Background(), kafka.Message{
					Key:   []byte("0200"), // Assuming ID is unique and can be used as the key
					Value: contentJSON,
				})
				if err != nil {
					log.Println("Error writing content0200 to Kafka:", err)
					return
				}

				//log.Printf("Written message with ID 0200: %+v\n", content0200)
				log.Printf("Written message with ID 0200: %+v\n", data)

			}()
	
	case "0230":
		var content0230 models.Content0230

		contentStr, ok := message.Content.(string)
		if !ok {
			log.Println("Error: Content is not a string")
			return
		}

		err := json.Unmarshal([]byte(contentStr), &content0230)
		if err != nil {
			log.Println("Error decoding content for ID 0230:", err)
			return
		}
		batch0230 = append(batch0230, content0230)

		gpsDataJSON, err := json.Marshal(content0230)
		if err != nil {
			log.Println("Error encoding GPS data to JSON:", err)
		}
		// Process Content0230
		routeId := content0230.RouteCode
		roomName := routeId
		eventName := "gpsData"
		log.Printf("Content0230: %+v\n", content0230)
		log.Printf("Received GPS Data: %+v. Broadcasting to room: %s\n", routeId, roomName)

		gpssocket.SendMessageToUser(roomName, eventName, string(gpsDataJSON))

		RouteName, ok := m[content0230.RouteCode]
		if !ok {
			RouteName = "null"
		}

		data := models.ImapHeadKafka{
			OnlineNo:  helper.StringToUint(content0230.OnlineNo),
			Azimuth:   content0230.Dir,
			RouteCode: helper.StringToUint(content0230.RouteCode),
			RouteName: RouteName,
			Longitude: content0230.Longitude,
			Latitude:  content0230.Latitude,
			GetOn:     content0230.GetOn,
			GetOff:    content0230.GetOff,
			GPSTime:   content0230.GPSTime,
			Speed:     content0230.Speed,
		}
		contentJSON, err := json.Marshal(data)
		if err != nil {
			log.Println("Error encoding content0230 to JSON:", err)
			return
		}

		err = kafkaWriter.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte("0230"), // Assuming ID is unique and can be used as the key
			Value: contentJSON,
		})

		if err != nil {
			log.Println("Error writing content0230 to Kafka:", err)
			return
		}

		err = kafkaWriter3.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte("0230"), // Assuming ID is unique and can be used as the key
			Value: contentJSON,
		})

		if err != nil {
			log.Println("Error writing content0230 to Kafka:", err)
			return
		}

		//log.Printf("Written message with ID 0200: %+v\n", content0200)
		log.Printf("Written message with ID 0230: %+v\n", data)

	case "0202":
		var Content0202 models.Content0202

		contentStr, ok := message.Content.(string)
		if !ok {
			log.Println("Error: Content is not a string")
			return
		}

		err := json.Unmarshal([]byte(contentStr), &Content0202)
		if err != nil {
			log.Println("Error decoding content for ID 0202:", err)
		}
		batch0202 = append(batch0202, Content0202)

	case "0800":
		var content0800 models.Content0800

		contentStr, ok := message.Content.(string)
		if !ok {
			log.Println("Error: Content is not a string")
			return
		}

		err := json.Unmarshal([]byte(contentStr), &content0800)
		if err != nil {
			log.Println("error decodin content for ID 0800:", err)
		}
		log.Println(content0800)
		batch0800 = append((batch0800), content0800)

	default:
		log.Println("Unknown message ID:", message.ID)
		return
	}

	if len(batch0230) >= batchSize {
		if err := bulk0230(batch0230); err != nil {
			log.Printf("Error performing bulk insert: 0230 %v\n", err)
		}
		batch0230 = nil
	}

	if len(batch0202) >= batchSize {
		if err := bulk0202(batch0202); err != nil {
			log.Println("Error performing bulk insert: 0202 \n", err)
		}
		batch0202 = nil
	}

	if len(batch0800) >= batchSize {
		if err := bulk0800(batch0800); err != nil {
			log.Println("Error performing bulk insert: 0800 %v\n", err)
		}
		batch0800 = nil
	}
	//log.Println("\tThis operation took:", time.Since(startTime))
}

func bulk0230(messages []models.Content0230) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(messages).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func bulk0202(messages []models.Content0202) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(messages).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func bulk0800(messages []models.Content0800) error {
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Create(messages).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func initKafkaWriter1() error {
	kafkaWriter = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaBrokersBinary},
		Topic:   Topic,
	})
	kafkaWriter2 = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaBrokers},
		Topic:   Topic2,
	})
	kafkaWriter3 = kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{KafkaBrokers},
		Topic:   Topic3,
	})

	return nil
}
