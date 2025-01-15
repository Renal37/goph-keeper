package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Renal37/goph-keeper/internal/agent/client"
	"github.com/Renal37/goph-keeper/internal/logger"
	handler "github.com/Renal37/goph-keeper/internal/server/adapters/handler/grpc"
	"github.com/Renal37/goph-keeper/internal/server/adapters/middleware"
	interceptors "github.com/Renal37/goph-keeper/internal/server/adapters/middleware/grpc"
	repository "github.com/Renal37/goph-keeper/internal/server/adapters/repository/pg"
	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"github.com/Renal37/goph-keeper/internal/server/core/services"
	"github.com/golang-jwt/jwt/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/lib/pq"
)

var databaseURL string
var testJWTkey = "12345"
var testMasterKey = "1234567812345678"
var testMaxMsgSize = 100000648

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16.1-alpine3.18",
		Env: []string{
			"POSTGRES_PASSWORD=test",
			"POSTGRES_USER=test",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseURL = fmt.Sprintf("postgres://test:test@%s?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseURL)

	// Tell docker to hard kill the container in 120 seconds
	err = resource.Expire(120)
	if err != nil {
		log.Fatalf("Expire resource has error: %s", err)
	}

	var sqlDB *sql.DB
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 20 * time.Second
	if err = pool.Retry(func() error {
		sqlDB, err = sql.Open("postgres", databaseURL)
		if err != nil {
			return fmt.Errorf("Connection has error: %w", err)
		}

		err = sqlDB.Ping()
		if err != nil {
			return fmt.Errorf("Ping has error: %w", err)
		}

		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Run tests
	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func testServer(ctx context.Context) (*client.Client, func()) {
	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	lg, err := logger.Init("error")
	if err != nil {
		log.Fatalln(err)
	}

	repo, err := repository.NewDB(context.Background(), lg, databaseURL)
	if err != nil {
		lg.Fatal(err.Error())
	}

	baseServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			selector.UnaryServerInterceptor(
				auth.UnaryServerInterceptor(interceptors.GetAuthenticator(testJWTkey)),
				selector.MatchFunc(interceptors.AuthMatcher),
			),
		),
		grpc.ChainStreamInterceptor(
			selector.StreamServerInterceptor(
				auth.StreamServerInterceptor(interceptors.GetAuthenticator(testJWTkey)),
				selector.MatchFunc(interceptors.AuthMatcher),
			),
		),
	)
	userSvc := services.NewUserService(repo)

	// Create user service
	proto.RegisterUserServer(baseServer, &handler.UserHandler{
		Svc:    *userSvc,
		Logger: lg,
		JWTkey: testJWTkey,
	})

	// Create storage service
	storageSvc := services.NewStorageService(repo)
	proto.RegisterStorageServer(baseServer, &handler.StorageHandler{
		Svc:       *storageSvc,
		Logger:    lg,
		MasterKey: testMasterKey,
	})

	go func() {
		if err := baseServer.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	conn, err := grpc.DialContext(ctx, "",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			//nolint:wrapcheck // This legal return
			return lis.Dial()
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(testMaxMsgSize), grpc.MaxCallSendMsgSize(testMaxMsgSize)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		baseServer.Stop()
	}

	token, err := getJWT(testJWTkey, 1, "test")
	if err != nil {
		log.Printf("error get jwt: %v", err)
	}

	return &client.Client{
		Conn:  conn,
		Token: *token,
	}, closer
}

func TestRegisterNewUser(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	r, err := cl.Register("test", "test")
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Jwt)
}

func TestLoginUser(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	r, err := cl.Login("test", "test")
	assert.NoError(t, err)
	assert.NotEmpty(t, r.Jwt)
}

func TestWriteText(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	_, err := cl.WriteFile("text", "test", "test")
	assert.NoError(t, err)
}

func TestWriteFile(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	_, err := cl.WriteFile("file", "test.zip", "../../assets/test.zip")
	assert.NoError(t, err)
}

func TestReadAllFile(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	r, err := cl.ReadAllFile()
	assert.NoError(t, err)
	assert.NotZero(t, len(r.Units))
}

func TestReadText(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	r, err := cl.ReadFile(1)
	assert.NoError(t, err)

	assert.NotEmpty(t, r.Data)
	assert.Equal(t, r.Type, "text")
}

func TestReadFile(t *testing.T) {
	ctx := context.Background()

	cl, closer := testServer(ctx)
	defer closer()

	r, err := cl.ReadFile(2)
	assert.NoError(t, err)

	assert.NotEmpty(t, r.Data)
	assert.Equal(t, r.Type, "file")
}

/* UTILS. */
func getJWT(jwtKey string, id int, login string) (*string, error) {
	var DefaultSession = 30
	var DefaultExpTime = time.Now().Add(time.Duration(DefaultSession) * time.Minute)

	claims := &middleware.JWTclaims{
		ID:    id,
		Login: login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(DefaultExpTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return nil, fmt.Errorf("failed signed jwt: %w", err)
	}

	return &tokenString, nil
}
