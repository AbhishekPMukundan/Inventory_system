package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/mycode/inventory-system/inventory/inventorypb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Inventory client")
	opts := grpc.WithInsecure()

	cc, err := grpc.Dial("localhost:50051", opts)

	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}

	defer cc.Close()

	c := inventorypb.NewInventoryServiceClient(cc)

	inventory := &inventorypb.Inventory{
		Barcode:     "90930823",
		Store:       "MishiPay Global",
		Name:        "Global",
		Category:    "T-shirt",
		Description: "Stupid tshirt",
		Image:       "Image url",
		Restricted:  false,
		Inventory:   10,
		Msrp:        "10.00",
		Price:       "9.99",
	}

	createInventoryResponse, err := c.CreateInventory(context.Background(), &inventorypb.CreateInventoryRequest{Inventory: inventory})
	if err != nil {
		log.Fatalf("Unexpected error %v:", err)
	}
	fmt.Printf("Inventory has been created : %v\n", createInventoryResponse)
	inventoryID := createInventoryResponse.GetInventory().GetId()

	//read Inventory
	fmt.Println("Reading inventory")

	_, err2 := c.ReadInventory(context.Background(), &inventorypb.ReadInventoryRequest{InventoryId: "12"})

	if err2 != nil {
		fmt.Printf("Error happened while reading: %v\n", err2)
	}

	readInventoryReq := &inventorypb.ReadInventoryRequest{InventoryId: inventoryID}
	readInventoryRes, readBloagErr := c.ReadInventory(context.Background(), readInventoryReq)

	if readBloagErr != nil {
		fmt.Printf("error happened while reading; %v\n", readBloagErr)
	}

	fmt.Printf("Inventory : %v", readInventoryRes)

	//Update Inventory
	newInventory := &inventorypb.Inventory{
		Id:          inventoryID,
		Barcode:     "90930823",
		Store:       "MishiPay Global",
		Name:        "Global",
		Category:    "T-shirt",
		Description: "Stupid tshirt",
		Image:       "Image url",
		Restricted:  false,
		Inventory:   10,
		Msrp:        "10.00",
		Price:       "9.99",
	}

	updateRes, updateErr := c.UpdateInventory(context.Background(), &inventorypb.UpdateInventoryRequest{Inventory: newInventory})
	if updateErr != nil {
		fmt.Printf("Error happened while updating: %v \n", updateErr)
	}
	fmt.Printf("Inventory was update: %v\n", updateRes)

	// delete Inventory
	deleteRes, deleteErr := c.DeleteInventory(context.Background(), &inventorypb.DeleteInventoryRequest{InventoryId: inventoryID})

	if deleteErr != nil {
		fmt.Printf("Error happened while deleting: %v \n", deleteErr)
	}
	fmt.Printf("INventory was deleted: %v \n", deleteRes)

	// list Inventory

	stream, err := c.ListInventory(context.Background(), &inventorypb.ListInventoryRequest{})
	if err != nil {
		log.Fatalf("error while calling List inventory RPC: %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happened: %v", err)
		}
		fmt.Println(res.GetInventory())
	}
}
