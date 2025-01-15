package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/Renal37/goph-keeper/internal/server/core/domain/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var maxMsgSize = 100000648
var errorResponseFinished = "response finished error: %w"
var errorEesponseReturn = "response return error: %w"

type Client struct {
	Conn  *grpc.ClientConn
	Token string
}

func NewClient(addr string, certPath string, token string) (*Client, error) {
	// Get TLS cert
	tlsCredentials, err := loadTLSCredentials(certPath)
	if err != nil {
		return nil, fmt.Errorf("cannot load TLS credentials: %w", err)
	}

	// Connect to gRPC server
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMsgSize), grpc.MaxCallSendMsgSize(maxMsgSize)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed start grpc server: %w", err)
	}

	return &Client{
		Conn:  conn,
		Token: token,
	}, nil
}

func (c Client) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return fmt.Errorf("failed close gRPC client: %w", err)
	}

	return nil
}

func (c Client) Register(login string, password string) (*proto.RegisterResponse, error) {
	// Create client
	client := proto.NewUserClient(c.Conn)
	resp, err := client.Register(context.Background(), &proto.RegiserRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

func (c Client) Login(login string, password string) (*proto.LoginResponse, error) {
	// Create client
	client := proto.NewUserClient(c.Conn)
	resp, err := client.Login(context.Background(), &proto.LoginRequest{
		Login:    login,
		Password: password,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

func (c Client) ReadAllFile() (*proto.ReadAllRecordResponse, error) {
	// Set authorization in gRPC metadata
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Create client
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.ReadAllRecord(ctx, &proto.ReadAllRecordRequest{})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

//nolint:dupl // This legal duplicate
func (c Client) ReadFile(id int32) (*proto.ReadRecordResponse, error) {
	// Set authorization in gRPC metadata
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Create client
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.ReadRecord(ctx, &proto.ReadRecordRequest{
		Id: id,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

func (c Client) WriteFile(typ string, name string, data string) (*proto.WriteRecordResponse, error) {
	// Set authorization in gRPC metadata
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Create client
	client := proto.NewStorageClient(c.Conn)
	stream, err := client.WriteRecord(ctx)
	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}

	var resp *proto.WriteRecordResponse
	switch typ {
	case "text":
		// Send the gRPC data
		err = stream.Send(&proto.WriteRecordRequest{Name: name, Data: []byte(data), Type: "text"})
		if err != nil {
			return nil, fmt.Errorf("stream send has error: %w", err)
		}

		// Close the stream and get a response
		resp, err = stream.CloseAndRecv()
		if err != nil {
			return nil, fmt.Errorf("closed stream has error: %w", err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
		}
	case "file":
		file, err := os.Open(data)
		if err != nil {
			return nil, fmt.Errorf("failed open file: %w", err)
		}

		fi, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed read stat file: %w", err)
		}

		if fi.Size() > int64(maxMsgSize) {
			return nil, fmt.Errorf("maximum file size should be less: %v bytes", maxMsgSize)
		}

		// Read the file in chunks and send
		chunkSize := 4096
		buf := make([]byte, chunkSize)
		for {
			n, err := file.Read(buf)
			if errors.Is(err, io.EOF) {
				// End of file, close the stream
				resp, err = stream.CloseAndRecv()
				if err != nil {
					return nil, fmt.Errorf("failed CloseAndRecv: %w", err)
				}
				if resp.Error != "" {
					return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
				}
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed read file: %w", err)
			}

			// Send a piece of data
			err = stream.Send(&proto.WriteRecordRequest{Name: name, Data: buf[:n], Type: "file"})
			if err != nil {
				return nil, fmt.Errorf("failed send stream: %w", err)
			}
		}

		err = file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed close file: %w", err)
		}
	}

	return resp, nil
}

//nolint:dupl // This legal duplicate
func (c Client) DeleteFile(id int32) (*proto.DeleteRecordResponse, error) {
	// Set authorization in gRPC metadata
	md := metadata.Pairs("authorization", fmt.Sprintf("bearer %s", c.Token))
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// Create client
	client := proto.NewStorageClient(c.Conn)
	resp, err := client.DeleteRecord(ctx, &proto.DeleteRecordRequest{
		Id: id,
	})

	if err != nil {
		return nil, fmt.Errorf(errorResponseFinished, err)
	}
	if resp.Error != "" {
		return nil, fmt.Errorf(errorEesponseReturn, resp.Error)
	}

	return resp, nil
}

// loadTLSCredentials loading certificates.
func loadTLSCredentials(cert string) (credentials.TransportCredentials, error) {
	// Load certificate of the CA who signed server's certificate
	pemServerCA, err := os.ReadFile(cert)
	if err != nil {
		return nil, fmt.Errorf("failde load file: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs:    certPool,
		MinVersion: tls.VersionTLS12,
	}

	return credentials.NewTLS(config), nil
}
