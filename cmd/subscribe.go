package cmd

import (
	"Queue4DownloadGo/types"
	"Queue4DownloadGo/util"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// subscribeCmd represents the subscribe command
var subscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "subscribe to an mqtt topic and process events",
	Long: `Subscribe to the download mqtt topic and process events. If the event is 
	valid, then we lftp the file/directory down to the appropriate folder
`,
	Run: subscribe,
}
var wg sync.WaitGroup
var fullQueue util.ConcurrentQueue[types.MQTTMessage]
var ticker = time.NewTicker(200 * time.Millisecond)

func init() {
	rootCmd.AddCommand(subscribeCmd)
	subscribeCmd.Flags().StringP("topic", "t", "queue", "the MQTT topic to use for subscribing the message, should be same as publish")
}

func subscribe(cmd *cobra.Command, args []string) {
	client := util.InitMQTTWithHandlers(onMessageReceived, nil, nil)
	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		slog.Error("Topic flag is invalid:" + err.Error())
		return
	}
	// This is to keep the subscribe command running indefinitely until there is a signal to kill
	// AKA CTRL-C
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	slog.Info("Subscribed to topic: " + topic)
	go eventProcessor()
	<-keepAlive
	slog.Info("Ending the subscription...")
	client.Disconnect(250)
}

// onMessageReceived is a callback function that is called when a message is received on the MQTT topic that the client is subscribed to.
// It unmarshals the JSON payload of the message into a types.MQTTMessage struct, logs the payload, and enqueues the message in the fullQueue.
func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	// It is assumed that the message is json, so we should unmarshall it.
	var jsonMsg types.MQTTMessage

	err := json.Unmarshal(msg.Payload(), &jsonMsg)
	if err != nil {
		slog.Error("Json formatting error: " + err.Error())
		return
	}
	logJson := fmt.Sprintf("MQTT Payload: %s", string(msg.Payload()))
	slog.Info(logJson)
	fullQueue.Enqueue(jsonMsg)
}

// eventProcessor is a goroutine that runs on a timer and processes events from the fullQueue.
// It dequeues an event from the fullQueue and calls processEvent to handle the event.
// This function is responsible for the main event processing loop of the application.
func eventProcessor() {
	for range ticker.C {
		if !fullQueue.IsEmpty() {
			item := fullQueue.Dequeue()
			processEvent(item)
		}
	}

}

// processEvent is a function that processes an event from the fullQueue. It generates a code from the rules in the config, and initiates a transfer of the paylaod to the configured destination.
// The function first logs a message indicating the name of the event being processed. It then calls util.GenerateCodeFromRules to generate a code from the rules defined in your config. If there is an error generating the code, it logs an error message.
// The function then adds a new goroutine to the waitgroup (wg) and calls initiateTransfer to initiate the transfer of the payload to the configured destination. The function then waits for the transfer to complete before returning.
func processEvent(item types.MQTTMessage) {
	msg := fmt.Sprintf("Processing Name - \"%s\"", item.Name)
	slog.Info(msg)
	code, err := util.GenerateCodeFromRules(item)
	if err != nil {
		slog.Error("Code processing error: " + err.Error())
	}
	wg.Add(1)
	go initiateTransfer(item.Name, code, item.Location)
	wg.Wait()
}

func initiateTransfer(name string, code string, location string) {
	defer wg.Done()
	var config types.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatal(err)
	}
	codeDestinations := viper.GetStringMapString("client.codeDestinations")
	toPath, found := codeDestinations[strings.ToLower(code)]
	if !found {
		log.Fatal("No code destination found")
	}
	username := viper.GetString("client.credentials.username")
	password := viper.GetString("client.credentials.password")
	host := viper.GetString("torrent.source")
	lftpArgsAsDir := fmt.Sprintf("-u \"%s,%s\" sftp://%s/  -e \"set sftp:auto-confirm yes; lcd %s; mirror -c --parallel=%d --use-pget-n=%d '%s' ;quit\"",
		username, password, host, toPath, config.Client.LFTP.Threads, config.Client.LFTP.Segments, location)
	lftpArgsAsFile := fmt.Sprintf("-u \"%s,%s\" sftp://%s/  -e \"set sftp:auto-confirm yes; lcd %s; pget -n %d '%s' ;quit\"",
		username, password, host, toPath, config.Client.LFTP.Threads, location)
	binPath, err := util.CheckIfCommandExists("lftp")
	if err != nil {
		slog.Error("Command lftp does not exist")
		return
	}
	statusCode, err := util.RunCommand(binPath, lftpArgsAsDir)
	if statusCode != 0 {
		if err != nil {
			slog.Error("The directory failed to clone and there was an error: " + err.Error())
		} else {
			slog.Info("Retrying the command to clone as a file...")
			statusCode, err = util.RunCommand(binPath, lftpArgsAsFile)
			if statusCode != 0 {
				slog.Error("Error trying to clone the file: " + name)
				if err != nil {
					slog.Error(err.Error())
				}
			} else {
				slog.Info("Successfully cloned the file: " + name)
			}
		}
	} else {
		slog.Info("Successfully cloned the directory: " + name)
	}

}
