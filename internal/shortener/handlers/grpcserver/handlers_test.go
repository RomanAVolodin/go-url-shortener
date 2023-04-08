package grpcserver

import (
	"context"
	"log"
	"net"
	"os"
	"testing"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"
	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/repositories"
	tLoc "github.com/RomanAVolodin/go-url-shortener/internal/shortener/tests"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"
	"github.com/stretchr/testify/assert"

	pb "github.com/RomanAVolodin/go-url-shortener/internal/shortener/proto"
	"google.golang.org/grpc"
)

var conn *grpc.ClientConn
var client pb.ShortenerClient

func TestMain(m *testing.M) {
	listen, err := net.Listen("tcp", ":3200")
	if err != nil {
		log.Fatal(err)
	}
	gRPCServer := grpc.NewServer(grpc.UnaryInterceptor(UnaryUserIDInterceptor))

	pb.RegisterShortenerServer(gRPCServer, &ShortenerGrpc{Shortener: &handlers.Shortener{Repo: &repositories.InMemoryRepository{
		Storage: map[string]entities.ShortURL{
			tLoc.ShortURLFixture.ID:         tLoc.ShortURLFixture,
			tLoc.ShortURLFixtureInactive.ID: tLoc.ShortURLFixtureInactive,
		},
	}}})
	go func() {
		if err := gRPCServer.Serve(listen); err != nil {
			log.Fatal(err)
		}
	}()

	conn, _ = grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	client = pb.NewShortenerClient(conn)

	code := m.Run()
	gRPCServer.GracefulStop()
	conn.Close()
	os.Exit(code)
}

func TestShortenerGrpc_CreateURL(t *testing.T) {
	resp, err := client.CreateURL(context.Background(), &pb.CreateShortURLSimpleRequest{
		Url: "https://mail.ru",
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_RetrieveURL(t *testing.T) {
	resp, err := client.RetrieveURL(context.Background(), &pb.RetrieveShortURLByIDRequest{
		Id: tLoc.ShortURLFixture.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_RetrieveURLShouldFailNotFound(t *testing.T) {
	_, err := client.RetrieveURL(context.Background(), &pb.RetrieveShortURLByIDRequest{
		Id: "fake_id",
	})
	assert.Error(t, err)
}

func TestShortenerGrpc_RetrieveURLShouldFailDeleted(t *testing.T) {
	_, err := client.RetrieveURL(context.Background(), &pb.RetrieveShortURLByIDRequest{
		Id: tLoc.ShortURLIDFixtureInactive,
	})
	assert.Error(t, err)
}

func TestShortenerGrpc_CreateMultipleURLs(t *testing.T) {
	urls := []*pb.CreateURLsWithCorrelationRequest{
		{
			CorrelationId: "12",
			OriginalUrl:   "https://mail.ru",
		},
	}
	resp, err := client.CreateMultipleURLs(
		context.Background(),
		&pb.CreateMultipleRequest{Urls: urls},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_GetUsersRecords(t *testing.T) {
	resp, err := client.GetUsersRecords(
		context.Background(),
		&pb.GetUsersRecordsRequest{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_DeleteRecords(t *testing.T) {
	ids := []string{
		"zhbZW9Gr8Ffex8e3NBnjd7",
		"gjTbAf2pYzyZF4vvg5YWHS",
		"nRSKAu77ejkSE7PXWsZoG8",
	}
	resp, err := client.DeleteRecords(
		context.Background(),
		&pb.DeleteRecordsRequest{Id: ids},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_GetServiceStats(t *testing.T) {
	resp, err := client.GetServiceStats(
		context.Background(),
		&pb.GetServiceStatsRequest{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestShortenerGrpc_PingDatabase(t *testing.T) {
	resp, err := client.PingDatabase(
		context.Background(),
		&pb.PingDbRequest{},
	)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
