package seedstore

import (
	"fmt"

	"github.com/spf13/viper"
)

type SeedStore struct {
	viper *viper.Viper
}

func New() (*SeedStore, error) {
	v := viper.New()
	v.AddConfigPath(".")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return &SeedStore{viper: v}, nil
}

func (s *SeedStore) Subscribe() string {
	fmt.Println("subscribed")
	return "subscribed"
}
