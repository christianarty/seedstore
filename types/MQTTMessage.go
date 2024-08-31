package types

type MQTTMessage struct {
	Name     string `json:"name"`
	Hash     string `json:"hash"`
	Location string `json:"location"`
	Category string `json:"category"`
}
