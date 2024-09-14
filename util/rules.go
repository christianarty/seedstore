package util

import (
	"reflect"
	"seedstore/types"
	"strings"

	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func GenerateCodeFromRules(message types.MQTTMessage) (string, error) {
	var serverRules types.Config
	err := viper.Unmarshal(&serverRules)
	if err != nil {
		return "", err
	}
	for _, condition := range serverRules.Server.CodeConditions {
		if evalExpression(condition, message) {
			return condition.Code, nil
		}
	}
	return serverRules.Server.DefaultCode, nil
}

func evalExpression(condition types.Rule, message types.MQTTMessage) bool {
	caseFormatter := cases.Title(language.English)
	r := reflect.ValueOf(message)
	field := reflect.Indirect(r).FieldByName(caseFormatter.String(condition.Entity))
	if field.String() == "" || field.String() == "<invalid value>" {
		return false
	}
	if condition.Operator == "=" || condition.Operator == "eq" {
		return field.String() == condition.Value
	}
	if condition.Operator == "contains" || condition.Operator == "in" {
		return strings.Contains(field.String(), condition.Value)
	}
	if condition.Operator == "!=" || condition.Operator == "not" {
		return field.String() != condition.Value
	}
	return false
}
