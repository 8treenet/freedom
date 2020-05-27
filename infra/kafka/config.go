package kafka

import (
	"crypto/x509"
	"io/ioutil"
	"time"

	"github.com/Shopify/sarama"
)

type kafkaConf struct {
	Producers []producerConf `toml:"producer_clients"`
	Consumers []consumerConf `toml:"consumer_clients"`
	Consumer  struct {
		Open       bool   `toml:"open"`
		ProxyHTTP2 bool   `toml:"proxy_http2"`
		ProxyAddr  string `toml:"proxy_addr"`
	} `toml:"consumer"`
	Producer struct {
		Open bool `toml:"open"`
	} `toml:"producer"`
	Other map[string]interface{} `toml:"other"`
}

type producerConf struct {
	Servers []string `toml:"servers"`
	Name    string   `toml:"name"`

	// Username string   `toml:"topics"`
	// Password string   `toml:"topics"`
	// CertFile string   `toml:"topics"`
}

type consumerConf struct {
	Servers []string `toml:"servers"`
	GroupID string   `toml:"group_id"`
	// Username string   `toml:"topics"`
	// Password string   `toml:"topics"`
	// CertFile string   `toml:"topics"`
}

func newConsumerConfig(kc consumerConf) *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Consumer.Retry.Backoff = 500 * time.Millisecond

	// if kc.Username != "" && kc.Password != "" && kc.CertFile != "" {
	// 	config.Net.SASL.Enable = true
	// 	config.Net.SASL.User = kc.Username
	// 	config.Net.SASL.Password = kc.Password
	// 	config.Net.SASL.Handshake = true
	// 	config.Net.TLS.Enable = true
	// 	config.Net.TLS.Config = &tls.Config{
	// 		RootCAs:            clientCertPool(kc.CertFile),
	// 		InsecureSkipVerify: true,
	// 	}
	// }
	return config
}

func newProducerConfig(kc producerConf) *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V0_11_0_0
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true

	config.Producer.Retry.Max = 6
	config.Producer.Retry.Backoff = 500 * time.Millisecond

	// if kc.Username != "" && kc.Password != "" && kc.CertFile != "" {
	// 	config.Net.SASL.Enable = true
	// 	config.Net.SASL.User = kc.Username
	// 	config.Net.SASL.Password = kc.Password
	// 	config.Net.SASL.Handshake = true
	// 	config.Net.TLS.Enable = true
	// 	config.Net.TLS.Config = &tls.Config{
	// 		RootCAs:            clientCertPool(kc.CertFile),
	// 		InsecureSkipVerify: true,
	// 	}
	// }
	return config
}

func clientCertPool(filePath string) *x509.CertPool {
	certBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	clientCertPool := x509.NewCertPool()
	ok := clientCertPool.AppendCertsFromPEM(certBytes)
	if !ok {
		panic("kafka producer failed to parse root certificate")
	}
	return clientCertPool
}

var confCallBack func(config *sarama.Config, other map[string]interface{})

// SettingConfig .
func SettingConfig(confFunc func(config *sarama.Config, other map[string]interface{})) {
	confCallBack = confFunc
}
