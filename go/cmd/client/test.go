package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/eshaanagg/pcbook/go/sample"
)

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	// Create 10 random laptops
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	filter := &pb.Filter{
		MaxPriceUsd: 3200,
		MinCpuCores: 4,
		MinCpuGhz:   2.5,
		MinRam: &pb.Memory{
			Value: 8,
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	searchLaptop(laptopClient, filter)
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.Id, "../../images/sampleLaptop.jpg")
}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	laptopIds := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIds[i] = laptop.GetId()
		createLaptop(laptopClient, laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("Rate laptop (Y/N): ")
		var answer string
		fmt.Scan(&answer)

		if strings.ToUpper(answer) != "Y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := rateLaptop(laptopClient, laptopIds, scores)
		if err != nil {
			log.Fatalf("Cannot rate laptop: %v", err)
		}
	}
}

func RunAllTests(laptopClient pb.LaptopServiceClient) {
	testCreateLaptop(laptopClient)
	testSearchLaptop(laptopClient)
	testUploadImage(laptopClient)
	testRateLaptop(laptopClient)
}
