package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gorilla/mux"
)

type MessageBody struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

var producer *kafka.Producer

func init() {
	prod, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		log.Fatalf("Unavailable Kafka Producer : %v", err.Error())
	}
	producer = prod
	log.Println("Kafka producer created")
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello, client!",
	})
}

func produceMessage(mBody MessageBody) {
	// Send message with touser as the key
	key := []byte(mBody.To) // Partitioning key based on touser
	value := []byte(mBody.From)

	// Produce message
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &[]string{"messages"}[0], Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
	}, nil)

	if err != nil {
		log.Printf("Failed to produce message: %s\n", err)
	} else {
		log.Printf("Message send to kafka")
	}
}

func handleMessageSend(w http.ResponseWriter, r *http.Request) {
	var messageBody MessageBody
	decodeErr := json.NewDecoder(r.Body).Decode(&messageBody)
	if decodeErr != nil {
		http.Error(w, decodeErr.Error(), http.StatusBadRequest)
		return
	}
	if messageBody.From == "" || messageBody.To == "" || messageBody.Content == "" {
		http.Error(w, "from,to and content is required!", http.StatusBadRequest)
		return
	}

	// send message to kafka
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Message sent successfully!",
	})

	produceMessage(messageBody)
}

func main() {
	log.Println("EDGE server for Unread messages")
	r := mux.NewRouter()
	r.HandleFunc("/sent-msg", handleMessageSend).Methods("POST")
	r.HandleFunc("/", handleWelcome).Methods("GET")
	log.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
