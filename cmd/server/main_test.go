package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

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
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	_ "github.com/lib/pq"
)

var databaseURL string
var testJWTkey = "12345"
var testMasterKey = "1234567812345678"
var testUser = "test"
var testUserID = 1

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

type clients struct {
	user    proto.UserClient
	storage proto.StorageClient
}

func testServer(ctx context.Context) (clients, func()) {
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
		}), grpc.WithTransportCredentials(insecure.NewCredentials()))
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

	uClient := proto.NewUserClient(conn)
	sClient := proto.NewStorageClient(conn)

	return clients{
		user:    uClient,
		storage: sClient,
	}, closer
}

type RegisterExp struct {
	out bool
	err string
}

type RegisterCase struct {
	name string
	in   *proto.RegiserRequest
	exp  RegisterExp
}

func TestRegisterNewUser(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []RegisterCase{
		{
			name: "Must add new user",
			in: &proto.RegiserRequest{
				Login:    "test",
				Password: "test",
			},
			exp: RegisterExp{
				out: true,
				err: "",
			},
		},
		{
			name: "Must return error - user exist",
			in: &proto.RegiserRequest{
				Login:    "test",
				Password: "test",
			},
			exp: RegisterExp{
				out: false,
				err: "this user exists",
			},
		},
		{
			name: "Must return error - password incorrect",
			in: &proto.RegiserRequest{
				Login: "test",
			},
			exp: RegisterExp{
				out: false,
				err: "login or password incorrect",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := client.user.Register(ctx, tt.in)
			assert.NoError(t, err)

			if out.Error != "" {
				if tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}
			}

			if out != nil && tt.exp.out {
				assert.NotEmpty(t, out.Jwt)
			}
		})
	}
}

type LoginExp struct {
	out bool
	err string
}

type LoginCase struct {
	name string
	in   *proto.LoginRequest
	exp  LoginExp
}

func TestLoginUser(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []LoginCase{
		{
			name: "Must login",
			in: &proto.LoginRequest{
				Login:    "test",
				Password: "test",
			},
			exp: LoginExp{
				out: true,
				err: "",
			},
		},
		{
			name: "Must return error - password incorrect",
			in: &proto.LoginRequest{
				Login: "test",
			},
			exp: LoginExp{
				out: false,
				err: "login or password incorrect",
			},
		},
		{
			name: "Must return error - user not found",
			in: &proto.LoginRequest{
				Login:    "test1",
				Password: "test",
			},
			exp: LoginExp{
				out: false,
				err: "user not found",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := client.user.Login(ctx, tt.in)
			assert.NoError(t, err)

			if out.Error != "" {
				if tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}
			}

			if out != nil && tt.exp.out {
				assert.NotEmpty(t, out.Jwt)
			}
		})
	}
}

type WriteFileExp struct {
	out *proto.WriteRecordResponse
	err string
}

type WriteFileCase struct {
	authorization bool
	name          string
	in            *proto.WriteRecordRequest
	exp           WriteFileExp
	err           error
}

func TestWriteFileStorage(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []WriteFileCase{
		{
			authorization: false,
			name:          "Write file forbidden without authorization",
			in: &proto.WriteRecordRequest{
				Name: "test",
				Type: "fiile",
				Data: []byte("test string"),
			},
			exp: WriteFileExp{
				out: &proto.WriteRecordResponse{},
				err: "",
			},
			//nolint:lll // This legal size
			err: errors.New("rpc error: code = Unauthenticated desc = AuthFromMD has error: rpc error: code = Unauthenticated desc = Request unauthenticated with bearer"),
		},
		{
			authorization: true,
			name:          "Write text must be success",
			in: &proto.WriteRecordRequest{
				Name: "test",
				Type: "text",
				Data: []byte("test string"),
			},
			exp: WriteFileExp{
				out: &proto.WriteRecordResponse{},
				err: "",
			},
		},
		{
			authorization: true,
			name:          "Write file must be success",
			in: &proto.WriteRecordRequest{
				Name: "test2",
				Type: "file",
				Data: []byte("test string"),
			},
			exp: WriteFileExp{
				out: &proto.WriteRecordResponse{},
				err: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.authorization {
				tkn, err := getJWT(testJWTkey, testUserID, testUser)
				assert.NoError(t, err)

				md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", *tkn))
				ctx = metadata.NewOutgoingContext(context.Background(), md)
			}

			stream, err := client.storage.WriteRecord(ctx)
			assert.NoError(t, err)

			err = stream.Send(tt.in)
			assert.NoError(t, err)

			out, err := stream.CloseAndRecv()
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if out != nil && out.Error != "" {
				if tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}
			}
		})
	}
}

type ReadAllExp struct {
	out *proto.ReadAllRecordResponse
	err string
}

type ReadAllCase struct {
	authorization bool
	name          string
	in            *proto.ReadAllRecordRequest
	exp           ReadAllExp
	err           error
}

func TestReadAllStorage(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []ReadAllCase{
		{
			authorization: false,
			name:          "Read all file forbidden without authorization",
			in:            &proto.ReadAllRecordRequest{},
			exp: ReadAllExp{
				out: &proto.ReadAllRecordResponse{},
				err: "",
			},
			//nolint:lll // This legal size
			err: errors.New("rpc error: code = Unauthenticated desc = AuthFromMD has error: rpc error: code = Unauthenticated desc = Request unauthenticated with bearer"),
		},
		{
			authorization: true,
			name:          "Read all file must be success",
			in:            &proto.ReadAllRecordRequest{},
			exp: ReadAllExp{
				out: &proto.ReadAllRecordResponse{},
				err: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.authorization {
				tkn, err := getJWT(testJWTkey, testUserID, testUser)
				assert.NoError(t, err)

				md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", *tkn))
				ctx = metadata.NewOutgoingContext(context.Background(), md)
			}

			out, err := client.storage.ReadAllRecord(ctx, tt.in)
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if out != nil {
				if out.Error != "" && tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}

				assert.NotZero(t, len(out.Units))
			}
		})
	}
}

type ReadFileExp struct {
	out *proto.ReadRecordResponse
	err string
}

type ReadFileCase struct {
	authorization bool
	name          string
	in            *proto.ReadRecordRequest
	exp           ReadFileExp
	err           error
}

func TestReadFileStorage(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []ReadFileCase{
		{
			authorization: false,
			name:          "Read file forbidden without authorization",
			in:            &proto.ReadRecordRequest{},
			exp: ReadFileExp{
				out: &proto.ReadRecordResponse{},
				err: "",
			},
			//nolint:lll // This legal size
			err: errors.New("rpc error: code = Unauthenticated desc = AuthFromMD has error: rpc error: code = Unauthenticated desc = Request unauthenticated with bearer"),
		},
		{
			authorization: true,
			name:          "Read file must be success",
			in: &proto.ReadRecordRequest{
				Id: 1,
			},
			exp: ReadFileExp{
				out: &proto.ReadRecordResponse{},
				err: "",
			},
		},
		{
			authorization: true,
			name:          "Read file not exist",
			in: &proto.ReadRecordRequest{
				Id: 44,
			},
			exp: ReadFileExp{
				out: &proto.ReadRecordResponse{},
				err: "record not found",
			},
		},
	}

	//nolint:dupl // This legal dupl
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.authorization {
				tkn, err := getJWT(testJWTkey, testUserID, testUser)
				assert.NoError(t, err)

				md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", *tkn))
				ctx = metadata.NewOutgoingContext(context.Background(), md)
			}

			out, err := client.storage.ReadRecord(ctx, tt.in)
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if out != nil {
				if out.Error != "" && tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}

				assert.NotEmpty(t, out)
			}
		})
	}
}

type DeleteFileExp struct {
	out *proto.DeleteRecordResponse
	err string
}

type DeleteFileCase struct {
	authorization bool
	name          string
	in            *proto.DeleteRecordRequest
	exp           DeleteFileExp
	err           error
}

func TestDeleteFileStorage(t *testing.T) {
	ctx := context.Background()

	client, closer := testServer(ctx)
	defer closer()

	tests := []DeleteFileCase{
		{
			authorization: false,
			name:          "Delete file forbidden without authorization",
			in:            &proto.DeleteRecordRequest{},
			exp: DeleteFileExp{
				out: &proto.DeleteRecordResponse{},
				err: "",
			},
			//nolint:lll // This legal size
			err: errors.New("rpc error: code = Unauthenticated desc = AuthFromMD has error: rpc error: code = Unauthenticated desc = Request unauthenticated with bearer"),
		},
		{
			authorization: true,
			name:          "Delete file must be success",
			in: &proto.DeleteRecordRequest{
				Id: 1,
			},
			exp: DeleteFileExp{
				out: &proto.DeleteRecordResponse{},
				err: "",
			},
		},
	}

	//nolint:dupl // This legal dupl
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.authorization {
				tkn, err := getJWT(testJWTkey, testUserID, testUser)
				assert.NoError(t, err)

				md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", *tkn))
				ctx = metadata.NewOutgoingContext(context.Background(), md)
			}

			out, err := client.storage.DeleteRecord(ctx, tt.in)
			if tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.err, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if out != nil {
				if out.Error != "" && tt.exp.err != out.Error {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.exp.err, out.Error)
				}

				assert.NotEmpty(t, out)
			}
		})
	}
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
