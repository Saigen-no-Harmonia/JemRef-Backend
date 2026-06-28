package testutil

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

// SetupTestFirebase テスト時にfirebaseエミュレータに接続するための設定
func SetupTestFirebase() (*firebase.App, error) {
	os.Setenv("FIREBASE_AUTH_EMULATOR_HOST", "localhost:9099")

	opt := option.WithoutAuthentication()
	return firebase.NewApp(context.Background(), nil, opt)
}
