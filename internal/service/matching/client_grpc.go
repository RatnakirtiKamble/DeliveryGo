package matching

import (
	"context"

	pb "github.com/RatnakirtiKamble/DeliveryGO/internal/transport/grpc/matchingpb"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/domain"
)

type GRPCClient struct {
	client pb.MatchingServiceClient
}

func NewGRPCClient(
	client pb.MatchingServiceClient,
) *GRPCClient {
	return &GRPCClient{client: client}
}


func (c *GRPCClient) SelectBestPath(
	ctx context.Context,
	order *domain.Order,
	candidateIDs []string,
) (*domain.PathTemplate, int, error) {

	var candidates []*pb.PathCandidate
	for _, id := range candidateIDs {
		candidates = append(candidates, &pb.PathCandidate{
			PathId: id,
		})
	}

	resp, err := c.client.SelectBestPath(ctx, &pb.SelectBestPathRequest{
		Order: &pb.Order{
			Id: order.ID,
			Lat: order.Lat,
			Lon: order.Lon,
		},
		Candidates: candidates,
	})
	
	if err != nil {
		return nil, 0, err 
	}

	return &domain.PathTemplate{ID: resp.BestPathId},
		int(resp.EstimatedCost),
		nil
}