package matchinggrpc 

import (
	"context"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/service/matching"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
	pb "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/grpc/matchingpb"
)

type Server struct {
	pb.UnimplementedMatchingServiceServer
	svc *matching.Service
}

func NewServer(svc *matching.Service) *Server {
	return &Server{svc:svc}
}

func (s *Server) SelectBestPath(
	ctx context.Context,
	req *pb.SelectBestPathRequest,
) (*pb.SelectBestPathResponse, error) {

	order := &domain.Order{
		ID:  req.Order.Id,
		Lat: req.Order.Lat,
		Lon: req.Order.Lon, 
	}

	var candidateIDs []string 
	for _, c := range req.Candidates {
		candidateIDs = append(candidateIDs, c.PathId)
	}

	path, cost, err := s.svc.SelectBestPath(order, candidateIDs)
	if err != nil {
		return nil, err 
	}

	return &pb.SelectBestPathResponse{
		BestPathId: 		path.ID,
		EstimatedCost: int32(cost),
	}, nil
}