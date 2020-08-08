package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/mycode/inventory-system/inventory/inventorypb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
)

var collection *mongo.Collection

type server struct {
}

type inventoryItem struct {
	ID          primitive.ObjectID `bson:"id,omitempty"`
	Barcode     string             `bson:"barcode,omitempty"`
	Store       string             `bson:"store,omitempty"`
	Name        string             `bson:"name,omitempty"`
	Category    string             `bson:"category,omitempty"`
	Description string             `bson:"description,omitempty"`
	Image       string             `bson:"image,omitempty"`
	Restricted  bool               `bson:"restricted,omitempty"`
	Inventory   int32              `bson:"inventory,omitempty"`
	Msrp        string             `bson:"msrp,omitempty"`
	Price       string             `bson:"price,omitempty"`
}

func (*server) CreateInventory(ctx context.Context, req *inventorypb.CreateInventoryRequest) (*inventorypb.CreateInventoryResponse, error) {
	inventory := req.GetInventory()
	data := inventoryItem{
		Barcode:     inventory.GetBarcode(),
		Store:       inventory.GetStore(),
		Name:        inventory.GetName(),
		Category:    inventory.GetCategory(),
		Description: inventory.GetDescription(),
		Image:       inventory.GetImage(),
		Restricted:  inventory.GetRestricted(),
		Inventory:   inventory.GetInventory(),
		Msrp:        inventory.GetMsrp(),
		Price:       inventory.GetPrice(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal Error %v", err),
		)
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)

	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("cannot convert OID"),
		)
	}

	return &inventorypb.CreateInventoryResponse{
		Inventory: &inventorypb.Inventory{
			Id:          oid.Hex(),
			Barcode:     inventory.GetBarcode(),
			Store:       inventory.GetStore(),
			Name:        inventory.GetName(),
			Category:    inventory.GetCategory(),
			Description: inventory.GetDescription(),
			Image:       inventory.GetImage(),
			Restricted:  inventory.GetRestricted(),
			Inventory:   inventory.GetInventory(),
			Msrp:        inventory.GetMsrp(),
			Price:       inventory.GetPrice(),
		},
	}, nil
}

func (*server) ReadInventory(ctx context.Context, req *inventorypb.ReadInventoryRequest) (*inventorypb.ReadInventoryResponse, error) {
	fmt.Println("Read INventory request")
	inventoryID := req.GetInventoryId()
	oid, err := primitive.ObjectIDFromHex(inventoryID)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	data := &inventoryItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find the item with the specified ID: %v", err),
		)
	}
	return &inventorypb.ReadInventoryResponse{
		Inventory: &inventorypb.Inventory{
			Id:          data.ID.Hex(),
			Barcode:     data.Barcode,
			Store:       data.Store,
			Name:        data.Name,
			Category:    data.Category,
			Description: data.Description,
			Image:       data.Image,
			Restricted:  data.Restricted,
			Inventory:   data.Inventory,
			Msrp:        data.Msrp,
			Price:       data.Price,
		},
	}, nil
}

func (*server) UpdateInventory(ctx context.Context, req *inventorypb.UpdateInventoryRequest) (*inventorypb.UpdateInventoryResponse, error) {
	fmt.Println("update inventory request")
	inventory := req.GetInventory()
	oid, err := primitive.ObjectIDFromHex(inventory.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}

	data := &inventoryItem{}
	filter := bson.M{"_id": oid}

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find the item with the specified ID: %v", err),
		)
	}

	data.Barcode = inventory.GetBarcode()
	data.Store = inventory.GetStore()
	data.Name = inventory.GetName()
	data.Category = inventory.GetCategory()
	data.Description = inventory.GetDescription()
	data.Image = inventory.GetImage()
	data.Restricted = inventory.GetRestricted()
	data.Inventory = inventory.GetInventory()
	data.Msrp = inventory.GetMsrp()
	data.Price = inventory.GetPrice()

	_, updateErr := collection.ReplaceOne(context.Background(), filter, data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in MongoDB: %v", updateErr),
		)
	}

	return &inventorypb.UpdateInventoryResponse{
		Inventory: &inventorypb.Inventory{
			Id:          data.ID.Hex(),
			Barcode:     data.Barcode,
			Store:       data.Store,
			Name:        data.Name,
			Category:    data.Category,
			Description: data.Description,
			Image:       data.Image,
			Restricted:  data.Restricted,
			Inventory:   data.Inventory,
			Msrp:        data.Msrp,
			Price:       data.Price,
		},
	}, nil
}

func (*server) DeleteInventory(ctx context.Context, req *inventorypb.DeleteInventoryRequest) (*inventorypb.DeleteInventoryResponse, error) {
	fmt.Println("Delete inventory request")

	oid, err := primitive.ObjectIDFromHex(req.GetInventoryId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse ID"),
		)
	}
	filter := bson.M{"_id": oid}

	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot delete object in MongoDB: %v", err),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot find object in MongoDB: %v", err),
		)
	}

	return &inventorypb.DeleteInventoryResponse{InventoryId: req.GetInventoryId()}, nil
}

func (*server) ListInventory(req *inventorypb.ListInventoryRequest, stream inventorypb.InventoryService_ListInventoryServer) error {
	fmt.Println("Listing all Items in inventory")

	cur, err := collection.Find(context.Background(), primitive.D{{}})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &inventoryItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding from MongoDB: %v", err),
			)
		}
		stream.Send(&inventorypb.ListInventoryResponse{
			Inventory: &inventorypb.Inventory{
				Id:          data.ID.Hex(),
				Barcode:     data.Barcode,
				Store:       data.Store,
				Name:        data.Name,
				Category:    data.Category,
				Description: data.Description,
				Image:       data.Image,
				Restricted:  data.Restricted,
				Inventory:   data.Inventory,
				Msrp:        data.Msrp,
				Price:       data.Price,
			},
		},
		)
	}

	if err := cur.Err(); err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", err),
		)
	}
	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Starting MongoBD")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://mongo:27017"))
	if err != nil {
		log.Fatal(err)
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inventory server started.....")

	collection = client.Database("inventorydb").Collection("Inventory")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}

	s := grpc.NewServer(opts...)
	inventorypb.RegisterInventoryServiceServer(s, &server{})
	reflection.Register(s)

	go func() {
		fmt.Println("Starting server.....")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	fmt.Println("stopping the server")
	s.Stop()
	fmt.Println("closing the listener")
	lis.Close()
	fmt.Println("Closing mongodb connection")
	client.Disconnect(context.TODO())
	fmt.Println("Stopping the listener")
}
