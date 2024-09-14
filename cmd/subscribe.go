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
	"log/slog"
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
		slog.Error("Json formatting error: " + err.Error())
		return
	}
	fmt.Println("[INFO] JSON Message:", string(msg.Payload()))
	queue <- jsonMsg
	go processEvent()
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

	sub(client, topic)

	<-keepAlive
	slog.Info("Ending the subscription...")
	client.Disconnect(250)
}

func sub(client mqtt.Client, topic string) {
	token := client.Subscribe(topic, 1, nil)
	token.Wait()
	slog.Info("Subscribed to topic: " + topic)
}

func processEvent() {
	item := <-queue
	msg := fmt.Sprintf("[INFO] Processing Name - \"%s\"", item.Name)
	fmt.Println(msg)
	code, err := util.ProcessEvent(item)
	if err != nil {
		slog.Error("Code processing error: " + err.Error())
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
		slog.Error("Command lftp does not exist")
		return
	}
	statusCode, err := runCommand(binPath, lftpArgsAsDir)
	if statusCode != 0 && err == nil {
		slog.Info("Retrying the command to clone as a file...")
		statusCode, err = runCommand(binPath, lftpArgsAsFile)
		if statusCode != 0 {
			slog.Error("Error trying to clone the file")
			if err != nil {
				slog.Error(err.Error())
			}
		} else {
			slog.Info("Successfully cloned the file")
		}
	}
	if statusCode == 0 {
		slog.Info("Successfully cloned the directory")
	}

}

func init() {
	rootCmd.AddCommand(subscribeCmd)

	subscribeCmd.Flags().StringP("topic", "t", "queue", "the MQTT topic to use for subscribing the message, should be same as publish")

}

func checkIfCommandExists(bin string) (path string, err error) {
	var fMsg string
	path, err = exec.LookPath(bin)
	if err != nil {
		fMsg = fmt.Sprintf("Command %s does not exist in PATH", bin)
		slog.Error(fMsg)
		return path, err
	}
	fMsg = fmt.Sprintf("'%s' exists, found at %s", bin, path)
	slog.Info(fMsg)
	return path, nil

}

func runCommand(binPath string, args string) (exitCode int, e error) {
	fullCmd := fmt.Sprintf("%s %s", binPath, args)
	cmd := exec.Command("bash", "-c", fullCmd)
	var stdout, stderr bytes.Buffer
	// Write to stdout/err but also capture it in a variable
	prefixWriterStdOut := NewPrefixWriter(os.Stdout, "[CMD] ")
	prefixWriterStdErr := NewPrefixWriter(os.Stderr, "[CMD-ERR] ")
	cmd.Stdout = io.MultiWriter(prefixWriterStdOut, &stdout)
	cmd.Stderr = io.MultiWriter(prefixWriterStdErr, &stderr)
	err := cmd.Run()
	if err != nil {
		switch e := err.(type) {
		case *exec.Error:
			slog.Error("The command failed executing: " + err.Error())
			return 126, err
		case *exec.ExitError:
			errCodeMsg := fmt.Sprintf("Exit Code: %d", e.ExitCode())
			slog.Error("The command executed, but an error happened")
			slog.Error(errCodeMsg)
			return e.ExitCode(), nil
		default:
			log.Fatal("[FATAL] Unexpected error executing your command,", err)
		}
	}
	return 0, nil
}

type PrefixWriter struct {
	w      io.Writer
	prefix string
}

func NewPrefixWriter(w io.Writer, prefix string) *PrefixWriter {
	return &PrefixWriter{w, prefix}
}

func (e PrefixWriter) Write(p []byte) (int, error) {
	prefix := []byte(e.prefix)
	n, err := e.w.Write(append(prefix, p...))
	if err != nil {
		return n, err
	}
	if n != len(p) {
		return n, io.ErrShortWrite
	}
	return len(p), nil
}
