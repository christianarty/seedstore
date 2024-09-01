package cmd

import (
	"Queue4DownloadGo/types"
	"Queue4DownloadGo/util"
	"bytes"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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
var queue = make(chan types.MQTTMessage, 2)

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	// It is assumed that the message is json, so we should unmarshall it.
	var jsonMsg types.MQTTMessage

	err := json.Unmarshal(msg.Payload(), &jsonMsg)
	if err != nil {
		fmt.Println("[ERR] Json formatting error:", err)
	}
	fmt.Println("[INFO] JSON Message:", string(msg.Payload()))
	queue <- jsonMsg
	go processEvent()
}

func subscribe(cmd *cobra.Command, args []string) {
	client := util.InitMQTTWithHandlers(onMessageReceived, nil, nil)
	topic, _ := cmd.Flags().GetString("topic")

	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	sub(client, topic)

	<-keepAlive
	fmt.Println("\n[INFO] Ending the subscription...")
	client.Disconnect(250)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("[INFO] Subscribed to topic: %s\n", topic)
}

func processEvent() {
	item := <-queue
	msg := fmt.Sprintf("[INFO] Processing Name - \"%s\"", item.Name)
	fmt.Println(msg)
	code, err := util.ProcessEvent(item)
	if err != nil {
		fmt.Println("[ERR] Code processing error:", err)
	}
	wg.Add(1)
	go initiateTransfer(code, item.Location)
	wg.Wait()
}

func initiateTransfer(code string, location string) {
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
	binPath, err := checkIfCommandExists("lftp")
	if err != nil {
		fmt.Println("[ERR] Command lftp does not exist")
		return
	}
	statusCode, err := runLftpCommand(binPath, lftpArgsAsDir)
	if statusCode != 0 && err == nil {
		fmt.Println("[INFO] Retrying the command to clone as a file...")
		statusCode, err = runLftpCommand(binPath, lftpArgsAsFile)
		if statusCode != 0 {
			fmt.Println("[ERR] Error trying to clone the file")
			if err != nil {
				fmt.Println("[ERR] ", err.Error())
			}
		}
	}

}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	subscribeCmd.Flags().StringP("topic", "t", "queue", "the MQTT topic to use for subscribing the message, should be same as publish")

}

func checkIfCommandExists(bin string) (path string, err error) {
	path, err = exec.LookPath(bin)
	if err != nil {
		fmt.Printf("[ERR] didn't find 'lftp' executable\n")
		return path, err
	} else {
		fmt.Printf("[INFO] 'lftp' executable is in '%s'\n", path)
		return path, nil
	}
}

func runLftpCommand(binPath string, args string) (exitCode int, e error) {
	fullCmd := fmt.Sprintf("%s %s", binPath, args)
	cmd := exec.Command("bash", "-c", fullCmd)
	var stdout, stderr bytes.Buffer
	// Write to stdout/err but also capture it in a variable
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdout)
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)
	err := cmd.Run()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("[ERR] The command failed executing: ", err)
			return 126, err
		case *exec.ExitError:
			fmt.Println("[ERR] The command executed, but an error happened")
			fmt.Println("[ERR] Exit Code: ", e.ExitCode())
			return e.ExitCode(), nil
		default:
			log.Fatal("[FATAL] Unexpected error executing your command,", err)
		}
	}
	return 0, nil
}
