package db

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

type FirebaseRepositoryImpl struct {
	client *auth.Client
}

func NewFirebaseRepositoryImpl(client *auth.Client) *FirebaseRepositoryImpl {
	return &FirebaseRepositoryImpl{
		client: client,
	}
}

// DeleteUser FIrebaseユーザを削除する
func (r *FirebaseRepositoryImpl) DeleteUser(
	ctx context.Context,
	firebaseUid string,
) error {

	err := r.client.DeleteUser(ctx, firebaseUid)

	if err == nil {
		return nil
	}

	// すでに削除済みでもOKとする
	if auth.IsUserNotFound(err) {
		return nil
	}

	return err
}
