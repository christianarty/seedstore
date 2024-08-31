package util

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type Handlers struct {
	OnReceivedMessageHandler mqtt.MessageHandler
	OnConnectionLostHandler  mqtt.ConnectionLostHandler
	OnConnectHandler         mqtt.OnConnectHandler
}

func InitMQTTDefault() mqtt.Client {
	return initMQTT(&Handlers{})
}

func InitMQTTWithHandlers(received mqtt.MessageHandler, connLost mqtt.ConnectionLostHandler, onConn mqtt.OnConnectHandler) mqtt.Client {
	return initMQTT(&Handlers{
		OnReceivedMessageHandler: received,
		OnConnectionLostHandler:  connLost,
		OnConnectHandler:         onConn,
	})
}

func initMQTT(handlers *Handlers) mqtt.Client {
	var messagePubHandler mqtt.MessageHandler
	var connectHandler mqtt.OnConnectHandler
	var connectLostHandler mqtt.ConnectionLostHandler

	if handlers.OnReceivedMessageHandler != nil {
		messagePubHandler = handlers.OnReceivedMessageHandler
	} else {
		messagePubHandler = func(client mqtt.Client, msg mqtt.Message) {
			fmt.Printf("[MQTT] Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
		}
	}
	if handlers.OnConnectionLostHandler != nil {
		connectLostHandler = handlers.OnConnectionLostHandler
	} else {
		connectLostHandler = func(client mqtt.Client, err error) {
			fmt.Printf("[MQTT] Connect lost: %v", err)
		}
	}

	if handlers.OnConnectHandler != nil {
		connectHandler = handlers.OnConnectHandler
	} else {
		connectHandler = func(client mqtt.Client) {
			fmt.Println("[MQTT] Connected")
		}
	}

	broker := viper.GetString("mqtt.host")
	fmt.Printf("Connecting to %s\n", broker)
	username := viper.GetString("mqtt.user")
	password := viper.GetString("mqtt.password")
	port := viper.GetInt("mqtt.port")
	if port == 0 {
		fmt.Println("[WARNING] PORT env not set, using the default 1883")
		port = 1883
	}

	clientId := viper.GetString("mqtt.clientId")
	if clientId == "" {
		fmt.Println("[WARNING] CLIENT_ID is not set in .env, generating new uuid as client id")
		clientId = uuid.NewString()
	}
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", broker, port))
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetAutoReconnect(true)
	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	return client
}
