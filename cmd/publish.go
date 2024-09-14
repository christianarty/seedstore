package cmd

import (
	"Queue4DownloadGo/types"
	"Queue4DownloadGo/util"
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"log/slog"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a message to the MQTT server",
	Long: `Publish a formatted JSON message to the MQTT server.
	The accepted values are:
	- name = name of the torrent at hand
	- hash = the torrent hash for it
	- location = the location of the torrent at hand
`,
	Args: cobra.NoArgs,
	Run:  publish,
}

func publish(cmd *cobra.Command, args []string) {
	client := util.InitMQTTDefault()
	name, _ := cmd.Flags().GetString("name")
	hash, _ := cmd.Flags().GetString("hash")
	location, _ := cmd.Flags().GetString("location")
	category, _ := cmd.Flags().GetString("category")
	topic, _ := cmd.Flags().GetString("topic")
	message := types.MQTTMessage{
		Name:     name,
		Hash:     hash,
		Location: location,
		Category: category,
	}
	pub(client, topic, &message)
}

func init() {
	rootCmd.AddCommand(publishCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	publishCmd.Flags().StringP("name", "n", "", "the name of the torrent at hand")
	publishCmd.Flags().StringP("hash", "s", "", "the hash of the torrent at hand")
	publishCmd.Flags().StringP("location", "l", "", "the location of the torrent at hand")
	publishCmd.Flags().StringP("category", "c", "", "the category code for the torrent")
	publishCmd.Flags().StringP("topic", "t", "queue", "the MQTT topic to use for publishing the message")

}

func pub(client mqtt.Client, topic string, message *types.MQTTMessage) {
	msg, err := json.Marshal(message)
	if err != nil {
		slog.Error(err.Error())
	}
	token := client.Publish(topic, 0, false, msg)
	token.Wait()

}
