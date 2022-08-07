package configs

import "os"

func GetGoogleProjectId() string {
	return os.Getenv("GOOGLE_CLOUD_PROJECT_ID")
}

func GetDefaultPubSubTopic() string {
	return os.Getenv("PUBSUB_TOPIC")
}

func GetDefaultSubscription() string {
	return os.Getenv("PUBSUB_SUBSCRIPTION")
}
