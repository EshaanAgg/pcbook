package service

import (
	"errors"
	"fmt"
	"sync"

	"github.com/eshaanagg/pcbook/go/pb"
	"github.com/jinzhu/copier"
)

var ErrAlreadyExists = errors.New("record already exists in the store")

// Define in an as interface, which can be implemented by multiple stores - InMemory, Database etc
type LaptopStore interface {
	Save(laptop *pb.Laptop) error
	Find(id string) (*pb.Laptop, error)
}

type InMemoryLaptopStore struct {
	// To manage concurrency
	mutex sync.RWMutex
	// To store the data as key-value pairs
	data map[string]*pb.Laptop
}

func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

func (store *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if store.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return fmt.Errorf("cannot copy the laptop data: %w", err)
	}

	store.data[other.Id] = other
	return nil
}

func (store *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	laptop := store.data[id]
	if laptop == nil {
		return nil, nil
	}

	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy the laptop data: %w", err)
	}

	return other, nil

}
