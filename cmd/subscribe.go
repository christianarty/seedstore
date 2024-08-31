/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"Queue4DownloadGo/types"
	"Queue4DownloadGo/util"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		fmt.Println("error:", err)
	}
	fmt.Println("JSON Message:", string(msg.Payload()))
	queue <- jsonMsg
	go processEvent()
}

func subscribe(cmd *cobra.Command, args []string) {
	client := util.InitMQTTWithHandlers(onMessageReceived, nil, nil)
	keepAlive := make(chan os.Signal)
	signal.Notify(keepAlive, os.Interrupt, syscall.SIGTERM)

	sub(client, "topic/test")

	<-keepAlive
	fmt.Println("\nEnding the subscription...")
	client.Disconnect(250)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

func processEvent() {
	fmt.Println("processing!!!")
	item := <-queue
	code, err := util.ProcessEvent(item)
	if err != nil {
		fmt.Println("error:", err)
	}
	wg.Add(1)
	go initiateTransfer(code, item.Location)
	fmt.Println("Done!!!")
	wg.Wait()
}

type LFTP struct {
	Threads  int `mapstructure:"threads"`
	Segments int `mapstructure:"segments"`
}

type ClientCredentials struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}
type ClientRules struct {
	Client struct {
		CodeDestinations map[string]string `mapstructure:"codeDestinations"`
		LFTP             LFTP              `mapstructure:"lftp"`
		Credentials      ClientCredentials `mapstructure:"credentials"`
	} `mapstructure:"client"`
}

func initiateTransfer(code string, location string) {
	defer wg.Done()
	var clientRules ClientRules
	err := viper.Unmarshal(&clientRules)
	if err != nil {
		log.Fatal(err)
	}
	toPath, found := clientRules.Client.CodeDestinations[strings.ToLower(code)]
	if !found {
		log.Fatal("No code destination found")
	}
	checkLftpExists()
	username := clientRules.Client.Credentials.Username
	password := clientRules.Client.Credentials.Password
	host := viper.GetString("torrent.source")
	lftpArgsAsDir := fmt.Sprintf("lftp -u \"%s,%s\" sftp://%s/  -e \"set sftp:auto-confirm yes; lcd %s; mirror -c --parallel=%d --use-pget-n=%d '%s' ;quit\"",
		username, password, host, toPath, clientRules.Client.LFTP.Threads, clientRules.Client.LFTP.Segments, location)
	lftpArgsAsFile := fmt.Sprintf("lftp -u \"%s,%s\" sftp://%s/  -e \"set sftp:auto-confirm yes; lcd %s; pget -n %d '%s' ;quit\"",
		username, password, host, toPath, clientRules.Client.LFTP.Threads, location)
	out, err := exec.Command("bash", "-c", lftpArgsAsDir).CombinedOutput()
	var tryAsFile = false
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			fmt.Println("failed executing:", err)
		case *exec.ExitError:
			fmt.Println("command exit rc =", e.ExitCode())
			fmt.Println(string(e.Stderr))
			fmt.Println("Trying the command as a file")
			tryAsFile = true
		default:
			tryAsFile = true
			fmt.Println("error:", err)
		}
	} else {
		fmt.Println("lftp command successfully executed")
		fmt.Println(string(out))
	}

	if tryAsFile {
		fileCmd, err := exec.Command("bash", "-c", lftpArgsAsFile).CombinedOutput()
		if err != nil {
			switch e := err.(type) {
			case *exec.Error:
				fmt.Println("failed executing:", err)
			case *exec.ExitError:
				fmt.Println("command exit rc =", e.ExitCode())
				fmt.Println(string(e.Stderr))
			default:
				panic(err)
			}
		} else {
			fmt.Println("lftp command successfully executed")
			fmt.Println(string(fileCmd))
		}
	}

}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// subscribeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// subscribeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func checkLftpExists() {
	path, err := exec.LookPath("lftp")
	if err != nil {
		fmt.Printf("didn't find 'lftp' executable\n")
	} else {
		fmt.Printf("'lftp' executable is in '%s'\n", path)
	}
}
