package tests

import (
	"context"
	"log"
	"os"
	"testing"

	_ "embed"

	usersvcv1 "github.com/mlukasik-dev/faceit-usersvc/gen/go/faceit/usersvc/v1"
	"github.com/mlukasik-dev/faceit-usersvc/internal/appconfig"
	"github.com/mlukasik-dev/faceit-usersvc/internal/controller"
	"github.com/mlukasik-dev/faceit-usersvc/internal/events"
	"github.com/mlukasik-dev/faceit-usersvc/internal/store"
	"go.uber.org/zap"
)

var (
	s        *store.Store
	ctr      usersvcv1.ServiceServer
	testData = struct {
		users []*store.User
	}{
		[]*store.User{
			{FirstName: "John", LastName: "Doe", Email: "john.doe7@gmail.com", Country: "UK"},
			{FirstName: "Jane", LastName: "Doe", Email: "jane.doe7@gmail.com", Country: "UK"},
		},
	}
)

//go:embed config.yaml
var configFile []byte

func TestMain(m *testing.M) {
	if err := appconfig.Init(configFile); err != nil {
		log.Fatal(err)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer logger.Sync()

	client, err := store.Connect(appconfig.AppConfig.Mongodb.URI)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	s = store.New(client)
	if err := s.CreateIndexes(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := s.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	if err := seedDB(); err != nil {
		log.Fatal(err)
	}

	e := events.New()

	ctr = controller.New(s, logger, e)

	os.Exit(m.Run())
}

func seedDB() error {
	var users []*store.User
	for _, u := range testData.users {
		user, err := s.CreateUser(context.Background(), u, "123456")
		if err != nil {
			return err
		}
		users = append(users, user)
	}
	testData.users = users
	return nil
}
