package main

import (
	"crypto/ed25519"
	"encoding/base64"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/connctd/connector-go"
	"github.com/connctd/lora-connector/connhttp"
	"github.com/connctd/lora-connector/lorawan"
	"github.com/connctd/lora-connector/mysql"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	Version = "undefined"
)

func setDefaults() {
	viper.SetDefault("http.addr", ":8088")
	viper.SetDefault("log.level", logrus.InfoLevel.String())
}

func readConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	setDefaults()
	flag.Parse()

	viper.SetEnvPrefix("LORACONN")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logrus.WithError(err).Warn("Failed to read configuration file")
	}
	logLevel, err := logrus.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.WithError(err).Warn("Failed to parse log level from config, using debug level as fallback")
	} else {
		logrus.SetLevel(logLevel)
	}
}

func main() {
	readConfig()

	dsn := viper.GetString("db.dsn")
	if dsn == "" {
		logrus.Fatal("Failed to read database DSN from config, please set db.dsn value")
	}

	host := viper.GetString("http.host")
	if host == "" {
		logrus.Fatal("no http.host configured. Please set http.host to the public host the connector is reachable under")
	}

	apiClient, err := connector.NewClient(connector.DefaultOptions(), connector.DefaultLogger)
	if err != nil {
		logrus.WithError(err).Fatalln("Failed to setup connctd client")
	}

	db, err := mysql.NewDB(dsn, apiClient, host)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	base64PubKey := viper.GetString("pubkey")
	if base64PubKey == "" {
		logrus.Fatal("Failed to read base64 encoded public key from config, please set 'pubkey' to a valid value")
	}
	pubKeyBytes, err := base64.RawStdEncoding.DecodeString(base64PubKey)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to decode base64 encoded public key")
	}
	pubKey := ed25519.PublicKey(pubKeyBytes)

	r := mux.NewRouter()

	loraWANHandler := lorawan.NewLoRaWANHandler(apiClient, true, db)
	r.Path("/lorawan/{installationId}/{instanceId}").Methods(http.MethodPost, http.MethodPut).Handler(loraWANHandler)
	cr := r.PathPrefix("/connector").Subrouter()

	connhttp.NewConnectorHandler(cr, db, "", pubKey)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		server := &http.Server{
			Addr:    viper.GetString("http.addr"),
			Handler: r,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Error("HTTP server failed")
		}
	}()

	<-sigs
	logrus.Info("Shutting down")
	os.Exit(0)

}
