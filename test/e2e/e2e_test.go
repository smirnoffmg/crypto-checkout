package e2e_test

import (
	"context"
	"crypto-checkout/internal/domain/invoice"
	"crypto-checkout/internal/domain/merchant"
	"crypto-checkout/internal/domain/payment"
	"crypto-checkout/internal/infrastructure/database"
	"crypto-checkout/internal/infrastructure/events"
	"crypto-checkout/internal/presentation/web"
	"crypto-checkout/pkg/config"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	postgresConnStr string
	kafkaBroker     string
	terminate       func()
	logger          *zap.Logger
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	logger = zap.NewNop()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("crypto_checkout_test"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.WithInitScripts(),
	)
	if err != nil {
		fmt.Printf("failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	// Get PostgreSQL connection string
	postgresConnStr, err = postgresContainer.ConnectionString(ctx)
	if err != nil {
		fmt.Printf("failed to get postgres connection string: %v\n", err)
		os.Exit(1)
	}

	// Start Redpanda (Kafka-compatible) container
	req := testcontainers.ContainerRequest{
		Image:        "docker.redpanda.com/redpandadata/redpanda:v24.1.5",
		ExposedPorts: []string{"9092/tcp"},
		Cmd: []string{
			"redpanda", "start",
			"--overprovisioned",
			"--smp", "1",
			"--memory", "512M",
			"--reserve-memory", "0M",
			"--node-id", "0",
			"--check=false",
			"--kafka-addr", "PLAINTEXT://0.0.0.0:9092",
			"--advertise-kafka-addr", "PLAINTEXT://localhost:9092",
		},
		WaitingFor: wait.ForLog("Successfully started Redpanda"),
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("failed to start Kafka: %v\n", err)
		os.Exit(1)
	}

	// Get broker host:port
	host, _ := kafkaC.Host(ctx)
	port, _ := kafkaC.MappedPort(ctx, "9092")
	kafkaBroker = fmt.Sprintf("%s:%s", host, port.Port())

	// Wait for containers to be ready
	time.Sleep(2 * time.Second)

	terminate = func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = postgresContainer.Terminate(ctx)
		_ = kafkaC.Terminate(ctx)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	terminate()

	os.Exit(code)
}

// StartTestApp starts the app with testcontainers on a random free port.
func StartTestApp(t *testing.T) string {
	t.Helper()

	// Pre-bind to :0 to get a free port
	lc := net.ListenConfig{}
	ln, err := lc.Listen(context.Background(), "tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to allocate port: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close() // Close immediately after getting the port

	var srv *http.Server
	app := fx.New(
		// Supply test configuration with real containers
		fx.Supply(&config.Config{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 0, // Will be set by testutil
			},
			Database: config.DatabaseConfig{
				URL: postgresConnStr,
			},
			Log: config.LogConfig{
				Level: "debug",
				Dir:   "logs",
			},
			Kafka: config.KafkaConfig{
				Brokers:            kafkaBroker,
				TopicDomainEvents:  "crypto-checkout.domain-events",
				TopicIntegrations:  "crypto-checkout.integrations",
				TopicNotifications: "crypto-checkout.notifications",
				TopicAnalytics:     "crypto-checkout.analytics",
			},
		}),
		// Supply test logger
		fx.Supply(logger),
		// Provide all dependencies
		database.Module,
		events.Module, // Use real events module for e2e tests
		invoice.Module,
		payment.Module,
		merchant.Module,
		web.Module,
		// Set Gin to test mode
		fx.Invoke(func() {
			gin.SetMode(gin.TestMode)
		}),
		// Decorate server with port
		fx.Decorate(func(s *http.Server) *http.Server {
			s.Addr = addr // use reserved port
			return s
		}),
		fx.Populate(&srv),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if startErr := app.Start(startCtx); startErr != nil {
		t.Fatalf("failed to start app: %v", startErr)
	}

	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer stopCancel()
		_ = app.Stop(stopCtx)
	})

	return "http://" + addr
}
