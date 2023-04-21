package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	"net/http"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/entities"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/middlewares"

	"github.com/RomanAVolodin/go-url-shortener/internal/shortener/handlers"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/RomanAVolodin/go-url-shortener/internal/shortener/proto"
	"github.com/google/uuid"
)

// ShortenerGrpc is struct based on Chi router with repository.
type ShortenerGrpc struct {
	*handlers.Shortener
	pb.UnimplementedShortenerServer
}

// CreateURL creates url.
func (s *ShortenerGrpc) CreateURL(
	ctx context.Context,
	in *pb.CreateShortURLSimpleRequest,
) (*pb.CreateShortURLSimpleResponse, error) {
	var response pb.CreateShortURLSimpleResponse
	url := in.Url
	userID := ctx.Value(middlewares.UserIDKey).(uuid.UUID)

	shortURL, statusCode, err := s.SaveToRepository(ctx, url, userID)
	if err != nil {
		if statusCode == http.StatusConflict {
			return nil, status.Errorf(codes.AlreadyExists, `URL %s already used`, url)
		}
		return nil, err
	}
	response.Result = shortURL.Short
	return &response, nil
}

// RetrieveURL fetches url.
func (s *ShortenerGrpc) RetrieveURL(
	ctx context.Context,
	in *pb.RetrieveShortURLByIDRequest,
) (*pb.RetrieveShortURLByIDResponse, error) {
	var response pb.RetrieveShortURLByIDResponse

	urlID := in.Id
	urlItem, exist, err := s.Repo.GetByID(ctx, urlID)

	if exist && err == nil {
		if !urlItem.IsActive {
			return nil, status.Errorf(codes.Unavailable, `URL %s has been deleted`, urlItem.ID)
		}
		response.OriginalUrl = urlItem.Original
		return &response, nil
	}
	if !exist || (err != nil && errors.Is(err, sql.ErrNoRows)) {
		return nil, status.Errorf(codes.NotFound, `URL was not found by id %s`, urlItem.ID)
	}
	return nil, err
}

// CreateMultipleURLs creates multiple urls.
func (s *ShortenerGrpc) CreateMultipleURLs(
	ctx context.Context,
	in *pb.CreateMultipleRequest,
) (*pb.CreateMultipleResponse, error) {
	var response pb.CreateMultipleResponse
	userID := ctx.Value(middlewares.UserIDKey).(uuid.UUID)

	var createDTOs = make([]entities.ShortURLWithCorrelationCreateDto, 0, len(in.Urls))
	for _, item := range in.Urls {
		createDTOs = append(createDTOs, entities.ShortURLWithCorrelationCreateDto{
			CorrelationID: item.CorrelationId,
			Original:      item.OriginalUrl,
		})
	}

	items, err := s.SaveMultipleToRepository(ctx, createDTOs, userID)

	if err != nil {
		return &response, err
	}

	var result []*pb.CreateURLsWithCorrelationResponse
	for _, item := range items {
		respDTO := pb.CreateURLsWithCorrelationResponse{CorrelationId: item.CorrelationID, ShortUrl: item.Short}
		result = append(result, &respDTO)
	}

	response.Result = result
	return &response, nil
}

// GetUsersRecords gets users urls.
func (s *ShortenerGrpc) GetUsersRecords(
	ctx context.Context,
	in *pb.GetUsersRecordsRequest,
) (*pb.GetUsersRecordsResponse, error) {
	var response pb.GetUsersRecordsResponse
	userID := ctx.Value(middlewares.UserIDKey).(uuid.UUID)

	records, err := s.Repo.GetByUserID(ctx, userID)
	if err != nil {
		return &response, err
	}
	var result []*pb.UrlResponse
	for _, item := range records {
		respDTO := pb.UrlResponse{ShortUrl: item.Short, OriginalUrl: item.Original}
		result = append(result, &respDTO)
	}

	response.Urls = result
	return &response, nil
}

// DeleteRecords deletes urls.
func (s *ShortenerGrpc) DeleteRecords(
	ctx context.Context,
	in *pb.DeleteRecordsRequest,
) (*pb.DeleteRecordsResponse, error) {
	var response pb.DeleteRecordsResponse
	userID := ctx.Value(middlewares.UserIDKey).(uuid.UUID)
	err := s.DeleteFromRepository(ctx, in.Id, userID)
	if err != nil {
		return &response, err
	}
	return &response, nil
}

// GetServiceStats gets stats.
func (s *ShortenerGrpc) GetServiceStats(
	ctx context.Context,
	in *pb.GetServiceStatsRequest,
) (*pb.GetServiceStatsResponse, error) {
	var response pb.GetServiceStatsResponse
	urlsAmount, errURL := s.Repo.GetOverallURLsAmount(ctx)
	usersAmount, errUser := s.Repo.GetOverallUsersAmount(ctx)
	if errURL != nil || errUser != nil {
		return &response, status.Error(codes.DataLoss, "Error occurred")
	}
	response.Urls = uint32(urlsAmount)
	response.Users = uint32(usersAmount)
	return &response, nil
}

// PingDatabase pings db.
func (s *ShortenerGrpc) PingDatabase(ctx context.Context, in *pb.PingDbRequest) (*pb.PingDbResponse, error) {
	var response pb.PingDbResponse
	if err := s.CheckDatabase(ctx); err == nil {
		return &response, nil
	}
	return &response, status.Error(codes.Unavailable, "Database unavailable")
}
