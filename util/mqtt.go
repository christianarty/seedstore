package util

import (
	"fmt"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"log"
	"log/slog"
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
			m := fmt.Sprintf("[MQTT] Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
			slog.Info(m)
		}
	}
	if handlers.OnConnectionLostHandler != nil {
		connectLostHandler = handlers.OnConnectionLostHandler
	} else {
		connectLostHandler = func(client mqtt.Client, err error) {
			slog.Error("[MQTT] Connect lost: ", err)
		}
	}

	if handlers.OnConnectHandler != nil {
		connectHandler = handlers.OnConnectHandler
	} else {
		connectHandler = func(client mqtt.Client) {
			slog.Info("[MQTT] Connected")
		}
	}

	broker := viper.GetString("mqtt.host")
	slog.Info("Connecting to", "broker", broker)
	username := viper.GetString("mqtt.user")
	password := viper.GetString("mqtt.password")
	port := viper.GetInt("mqtt.port")
	if port == 0 {
		slog.Warn("PORT var not set, using the default 1883")
		port = 1883
	}

	clientId := viper.GetString("mqtt.clientId")
	if clientId == "" {
		slog.Warn("CLIENT_ID var not set, generating new uuid as client id")
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
		log.Fatal(token.Error())
	}
	return client
}
