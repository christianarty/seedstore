package util

import (
	"bytes"
	"seedstore/types"
	"testing"

	"github.com/spf13/viper"
)

func TestRules(t *testing.T) {
	var jsonConfig = []byte(`
	{
	  "server":{
		"defaultCode":"V",
		"codeConditions":[
		  {
			"value":"bar",
			"operator":"in",
			"entity":"name",
			"code":"C"
		  }
		]
	  }
	}
	`)

	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	mqttMessage := types.MQTTMessage{
		Name:     "foobar",
		Hash:     "asadfasdfasdfasf",
		Location: "falasdfasdfala",
	}
	expectedCode := "C"
	code, err := GenerateCodeFromRules(mqttMessage)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Fatalf("Expected code %s, got %s", expectedCode, code)
	}

}

func TestRulesFail(t *testing.T) {
	var jsonConfig = []byte(`
	{
	  "server":{
		"defaultCode":"V",
		"codeConditions":[
		  {
			"value":"barfoo",
			"operator":"eq",
			"entity":"name",
			"code":"C"
		  }
		]
	  }
	}
	`)

	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	mqttMessage := types.MQTTMessage{
		Name:     "foobar",
		Hash:     "asadfasdfasdfasf",
		Location: "falasdfasdfala",
	}
	expectedCode := "V"
	code, err := GenerateCodeFromRules(mqttMessage)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Fatalf("Expected code %s, got %s", expectedCode, code)
	}
}

func TestRulesAnotherField(t *testing.T) {
	var jsonConfig = []byte(`
	{
	  "server":{
		"defaultCode":"V",
		"codeConditions":[
		  {
			"value":"barquot",
			"operator":"not",
			"entity":"hash",
			"code":"H"
		  }
		]
	  }
	}
	`)

	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	mqttMessage := types.MQTTMessage{
		Name:     "foobar",
		Hash:     "asadfasdfasdfasf",
		Location: "falasdfasdfala",
	}
	expectedCode := "H"
	code, err := GenerateCodeFromRules(mqttMessage)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Fatalf("Expected code %s, got %s", expectedCode, code)
	}
}

func TestRulesNoConditions(t *testing.T) {
	var jsonConfig = []byte(`
	{
	  "server":{
		"defaultCode":"V"
	  }
	}
	`)

	viper.SetConfigType("json")
	err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	mqttMessage := types.MQTTMessage{
		Name:     "foobar",
		Hash:     "asadfasdfasdfasf",
		Location: "falasdfasdfala",
	}
	expectedCode := "V"
	code, err := GenerateCodeFromRules(mqttMessage)
	if err != nil {
		t.Fatal(err)
	}
	if code != expectedCode {
		t.Fatalf("Expected code %s, got %s", expectedCode, code)
	}
}
