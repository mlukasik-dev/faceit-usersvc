// +build integration

package controller_test

import (
	"context"
	"log"
	"os"
	"testing"

	usersvcv1 "github.com/mlukasik-dev/usersvc/gen/go/usersvc/v1"
	"github.com/mlukasik-dev/usersvc/internal/appconfig"
	"github.com/mlukasik-dev/usersvc/internal/controller"
	"github.com/mlukasik-dev/usersvc/internal/events"
	"github.com/mlukasik-dev/usersvc/internal/store"
	"go.uber.org/zap"
)

var (
	s        *store.Store
	ctr      usersvcv1.ServiceServer
	l        *zap.Logger
	testData = struct {
		users []*store.User
	}{
		[]*store.User{
			{FirstName: "John", LastName: "Doe", Email: "john.doe@gmail.com", Country: "UK"},
			{FirstName: "Jane", LastName: "Doe", Email: "jane.doe@gmail.com", Country: "UK"},
			{FirstName: "Jan", LastName: "Kowalski", Email: "jan.kowalski@gmail.com", Country: "PL"},
		},
	}
)

var testConfig = `
mongodb:
  uri: ${MONGODB_URI:?uri was not provided}
`

func TestMain(m *testing.M) {
	if err := appconfig.Init([]byte(testConfig)); err != nil {
		log.Fatal(err)
	}

	var err error
	l, err = zap.NewProduction()
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()

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

	ctr = controller.New(s, l, e)

	code := m.Run()

	unseedDB()

	os.Exit(code)
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

func unseedDB() error {
	for _, u := range testData.users {
		err := s.DeleteUser(context.Background(), u.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
