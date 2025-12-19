package worker

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"

	"github.com/RatnakirtiKamble/DeliveryGO/internal/app"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/postgres"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/store/redis"
	"github.com/RatnakirtiKamble/DeliveryGO/internal/util"
)

type ProvisionEvent struct {
	StoreID 	string 	`json:"store_id"`
	H3 			string 	`json:"h3"`
}

type PathProvisionWorker struct {
	pathStore *postgres.PathTemplateStore
	pathIndex *redis.PathIndex
}

func NewPathProvisionWorker(
	pathStore *postgres.PathTemplateStore,
	pathIndex *redis.PathIndex,
) *PathProvisionWorker {
	return &PathProvisionWorker{
		pathStore: pathStore,
		pathIndex: pathIndex,
	}
}

func (w *PathProvisionWorker) Handle(
	ctx context.Context,
	msg kafka.Message,
) error {
	log.Printf(
	"[path-provision] received event topic=%s payload=%s",
	msg.Topic,
	string(msg.Value),
	)

	cfg := app.LoadConfig()
	if msg.Topic != "path.provision.requested" {
		return nil 
	}

	var evt ProvisionEvent
	if err := json.Unmarshal(msg.Value, &evt); err != nil {
		return err 
	}

	store, ok := util.StoreLocations[evt.StoreID]

	if !ok {
		log.Println("unknown store:", evt.StoreID)
		return nil 
	}

	rawLat, rawLon := util.H3CellToLatLon(evt.H3)

	lat, lon, err := util.SnapToRoad(rawLat, rawLon, cfg.OSRMAddr)
	if err != nil{
		log.Printf(
			"[path-provision] no routable point near h3=%s err=%v",
			evt.H3,
			err,
		)
		return nil
	}

	eta, polyline, err := util.ComputeRoute(
		store.Lat, store.Lon,
		lat, lon,
		cfg.OSRMAddr,
	)

	if err != nil {
		log.Println("Error computing route: ", err)
		return err 
	}

	hash := sha1.Sum([]byte(evt.StoreID + ":" + evt.H3))
	pathID := hex.EncodeToString(hash[:])

	if err := w.pathStore.Insert(
		ctx,
		pathID,
		evt.StoreID,
		evt.H3,
		eta,
		polyline,
	); err != nil {
		return err
	}

	if err := w.pathIndex.AddPathToH3(
		ctx,
		evt.H3,
		pathID,
	); err != nil {
		return err
	}

	log.Printf(
		"[path-provision-worker] provisioned path=%s store=%s h3=%s eta=%d",
		pathID,
		evt.StoreID,
		evt.H3,
		eta,
	)

	return nil
}