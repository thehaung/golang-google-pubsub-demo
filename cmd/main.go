package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	conf "github.com/thehaung/golang-google-pubsub-demo/pkg/configs"
	"github.com/thehaung/golang-google-pubsub-demo/pkg/models"
	"html/template"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	topic *pubsub.Topic

	messagesMu sync.Mutex
	messages   []string
)

const maxMessages = 10

func main() {
	InitEnv()

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, conf.GetGoogleProjectId())

	if err != nil {
		log.Fatal(err)
	}

	defer func(client *pubsub.Client) {
		err := client.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(client)

	topicName := conf.GetDefaultPubSubTopic()
	topic = client.Topic(topicName)

	exists, err := topic.Exists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		log.Printf("Topic %v doesn't exist - creating it", topicName)
		_, err = client.CreateTopic(ctx, topicName)
		if err != nil {
			log.Fatal(err)
		}
	}

	http.HandleFunc("/", listHandler)
	http.HandleFunc("/pubsub/publish", publishHandler)
	http.HandleFunc("/pubsub/push", pushHandler)

	// for subscription
	err = pullMessages(conf.GetGoogleProjectId(), conf.GetDefaultSubscription())

	// subscribe fail then serve http server =)))
	if err != nil {
		port := conf.GetServerPort()
		if port == "" {
			port = "8080"
			log.Printf("Defaulting to port %s", port)
		}

		log.Printf("Listening on port %s", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal(err)
		}
	}

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	messagesMu.Lock()
	defer messagesMu.Unlock()

	if err := tmpl.Execute(w, messages); err != nil {
		log.Printf("Could not execute template: %v", err)
	}
}

func pullMessages(projectID, subID string) error {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer func(client *pubsub.Client) {
		err := client.Close()
		if err != nil {

		}
	}(client)

	sub := client.Subscription(subID)

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long-running operation.
	//ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	//defer cancel()
	log.Printf("Got message: %q\n", "a")

	var received int32
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		log.Printf("Got message: %q\n", string(msg.Data))

		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}
	log.Printf("Received %d messages\n", received)
	if err != nil {
		return err
	}

	return nil
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	msg := &models.PushRequest{}
	if err := json.NewDecoder(r.Body).Decode(msg); err != nil {
		http.Error(w, fmt.Sprintf("Could not decode body: %v", err), http.StatusBadRequest)
		return
	}

	messagesMu.Lock()
	defer messagesMu.Unlock()
	// Limit to ten.
	messages = append(messages, string(msg.Message.Data))
	if len(messages) > maxMessages {
		messages = messages[len(messages)-maxMessages:]
	}
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	msg := &pubsub.Message{
		Data: []byte(r.FormValue("payload")),
	}

	if _, err := topic.Publish(ctx, msg).Get(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Could not publish message: %v", err), 500)
		return
	}

	_, err := fmt.Fprint(w, "Message published.")
	if err != nil {
		return
	}
}

func InitEnv() {
	if err := godotenv.Load("../configs/.env"); err != nil {
		log.Fatal(fmt.Sprintf("Loading env error: %s", err))
	}
}

var tmpl = template.Must(template.New("").Parse(`<!DOCTYPE html>
<html>
  <head>
    <title>Pub/Sub</title>
  </head>
  <body>
    <div>
      <p>Last ten messages received by this instance:</p>
      <ul>
      {{ range . }}
          <li>{{ . }}</li>
      {{ end }}
      </ul>
    </div>
    <form method="post" action="/pubsub/publish">
      <textarea name="payload" placeholder="Enter message here"></textarea>
      <input type="submit">
    </form>
    <p>Note: if the application is running across multiple instances, each
      instance will have its own list of messages.</p>
  </body>
</html>`))
