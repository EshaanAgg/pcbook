package sample

import (
	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/golang/protobuf/ptypes"
)

func NewKeyboard() *pb.Keyboard {
	keyboard := &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBoolean(),
	}

	return keyboard
}

func NewCPU() *pb.CPU {
	brand := randomCPUBrand()
	cores := randomInt(2, 8)
	minGhz := randomFloat64(2.0, 3.5)

	cpu := &pb.CPU{
		Brand:         brand,
		Name:          randomCPUName(brand),
		NumberCores:   uint32(cores),
		NumberThreads: uint32(randomInt(cores, 12)),
		MinGhz:        minGhz,
		MaxGhz:        randomFloat64(minGhz, 5.0),
	}
	return cpu
}

func NewGPU() *pb.GPU {
	brand := randomGPUBrand()
	minGhz := randomFloat64(1.0, 1.5)

	gpu := &pb.GPU{
		Brand:  brand,
		Name:   randomGPUName(brand),
		MinGhz: minGhz,
		MaxGhz: randomFloat64(minGhz, 2.0),
		Memory: &pb.Memory{
			Value: uint64(randomInt(2, 6)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	return gpu
}

func NewRAM() *pb.Memory {
	ram := &pb.Memory{
		Value: uint64(randomInt(4, 64)),
		Unit:  pb.Memory_GIGABYTE,
	}

	return ram
}

func NewSSD() *pb.Storage {
	ssd := &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(128, 1024)),
			Unit:  pb.Memory_GIGABYTE,
		},
	}

	return ssd
}

func NewHDD() *pb.Storage {
	hdd := &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: &pb.Memory{
			Value: uint64(randomInt(1, 6)),
			Unit:  pb.Memory_TERABYTE,
		},
	}

	return hdd
}

func NewScreen() *pb.Screen {
	screen := &pb.Screen{
		SizeInch:   randomFloat32(13, 17),
		Resolution: randomScreenResolution(),
		Panel:      randomScreenPanel(),
		Multitouch: randomBoolean(),
	}

	return screen
}

func NewLaptop() *pb.Laptop {
	brand := randomLaptopBrand()

	laptop := &pb.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     randomLaptopName(brand),
		Cpu:      NewCPU(),
		Ram:      NewRAM(),
		Gpus:     []*pb.GPU{NewGPU()},
		Storages: []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeyboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		UpdatedAt: ptypes.TimestampNow(),
	}

	return laptop
}
