package repository

import "go.etcd.io/bbolt"

type repository struct {
	db *bbolt.DB
}

type Repository interface {
	UserRepo() UserRepository
}

func New(db *bbolt.DB) *repository {
	return &repository{db: db}
}

func (r *repository) UserRepo() UserRepository {
	return &userRepository{db: r.db}
}
