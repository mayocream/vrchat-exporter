package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mayocream/vrchat-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/spf13/pflag"
)

var (
	username            = pflag.String("username", "", "Username of the VRChat user")
	password            = pflag.String("password", "", "Password of the VRChat user")
	totp                = pflag.String("totp", "", "TOTP of the VRChat user")
	listen              = pflag.String("listen", ":8080", "Address to listen on for HTTP requests")
	pushgateway         = pflag.String("pushgateway", "", "Address of the Prometheus Pushgateway, e.g. https://prometheus-blocks-prod-us-central1.grafana.net/api/prom/push")
	pushgatewayUsername = pflag.String("pushgateway-username", "", "Username of the Prometheus Pushgateway")
	pushgatewayPassword = pflag.String("pushgateway-password", "", "Password of the Prometheus Pushgateway")
	interval            = pflag.Int64("interval", 60, "Interval in seconds to push metrics to the Pushgateway")
)

const (
	namespace = "vrchat"
)

type VRChatCollector struct {
	userOnlineDesc *prometheus.Desc
	client         *vrchat.Client
}

func NewVRChatCollector(client *vrchat.Client) *VRChatCollector {
	return &VRChatCollector{
		client: client,
		userOnlineDesc: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "user_online"),
			"Whether the user is online",
			[]string{"username", "location", "status", "status_desc"}, nil,
		),
	}
}

func (c *VRChatCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.userOnlineDesc
}

func (c *VRChatCollector) Collect(ch chan<- prometheus.Metric) {
	users, err := c.client.GetFriends(vrchat.GetFriendsParams{})
	if err != nil {
		log.Printf("Failed to get friends: %v", err)
		return
	}

	for _, user := range *users {
		ch <- prometheus.MustNewConstMetric(c.userOnlineDesc, prometheus.GaugeValue, 1, user.DisplayName, user.Location, string(user.Status), user.StatusDescription)
	}

	log.Printf("Collected %d users", len(*users))
}

func init() {
	pflag.Parse()
}

func main() {
	// Create a new VRChat client
	client := vrchat.NewClient("https://vrchat.com/api/1")
	err := client.Authenticate(*username, *password, *totp)
	if err != nil {
		log.Fatalf("Failed to authenticate: %v", err)
	}

	// Get the current user
	user, err := client.GetCurrentUser()
	if err != nil {
		log.Fatalf("Failed to get current user: %v", err)
	}
	log.Printf("Current user: %s", user.DisplayName)

	// Create and register the collector
	collector := NewVRChatCollector(client)
	// Register the collector
	prometheus.MustRegister(collector)

	// Push metrics to the Prometheus Pushgateway
	if *pushgateway != "" {
		for {
			err := push.New(*pushgateway, "vrchat_exporter").
				BasicAuth(*pushgatewayUsername, *pushgatewayPassword).
				Collector(collector).
				Push()
			if err != nil {
				log.Fatalf("Failed to push metrics: %v", err)
			}

			log.Printf("Metrics pushed to %s", *pushgateway)

			time.Sleep(time.Duration(*interval) * time.Second)
		}
	}

	// Start the HTTP server
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Listening on %s", *listen)
	err = http.ListenAndServe(*listen, nil)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
}
