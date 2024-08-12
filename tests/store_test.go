package test

import "projectmanager/internal/types"

type MockStore struct{}

func (m *MockStore) CreateUser() error {
	return nil
}

func (m *MockStore) CreateTask(t *types.Task) (*types.Task, error) {
	return &types.Task{}, nil
}

func (m *MockStore) GetTask(id string) (*types.Task, error) {
	return &types.Task{}, nil
}
