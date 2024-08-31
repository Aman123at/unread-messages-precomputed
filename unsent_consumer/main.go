package main

import (
	"context"
	"log"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-redis/redis/v8"
)

var masterRedis *redis.Client
var auxReplica *redis.Client

func init() {
	// initialize master redis and its auxillary replica
	masterRedis = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	auxReplica = redis.NewClient(&redis.Options{
		Addr: "localhost:6380",
	})

	log.Println("Master and Replica are up")
}

func consumeMessages(workerID int, mywait *sync.WaitGroup) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          "worker-group",
	})
	if err != nil {
		log.Fatal(err)
	}

	// subscribe to messages event
	err = consumer.Subscribe("messages", nil)
	if err != nil {
		log.Fatal(err)
	}

	for {
		// start reading messages from consumer
		msg, err := consumer.ReadMessage(-1)
		if err == nil {
			log.Printf("Worker %d consumed message touser: %s, fromuser: %s\n", workerID, string(msg.Key), string(msg.Value))

			// check this user mapping in auxillary replica of master
			isFromuserPresent := auxReplica.SIsMember(context.Background(), string(msg.Key), string(msg.Value))

			if isFromuserPresent.Err() != nil {
				log.Printf("Error in checking user mapping : %v", isFromuserPresent.Err())
			} else {
				// if fromuser is not present then only add it to set
				if !isFromuserPresent.Val() {
					log.Println("NEW KEY FOR REDIS, ADDING INTO SET")

					// add this fromuser in set for respective key (touser)
					_ = masterRedis.SAdd(context.Background(), string(msg.Key), string(msg.Value))

					// add same in replica redis (temporarily), ideally auxillary replica should fetch data from master Asynchronously
					// auxillary replica is just to get rid of unnecessary writes to master Redis (redundent writes)
					_ = auxReplica.SAdd(context.Background(), string(msg.Key), string(msg.Value))
				}
			}
			mywait.Done()
		} else {
			log.Printf("Error while consuming message: %v\n", err)
			mywait.Done()
		}
	}
}

func main() {
	log.Println("Message Consumer")

	// number of workers == number of kafka partitions
	numWorkers := 4
	wg := sync.WaitGroup{}
	wg.Add(4)
	for i := 0; i < numWorkers; i++ {
		wg.Add(i + 1)
		go consumeMessages(i, &wg) // Start each worker
	}
	wg.Wait()
}
