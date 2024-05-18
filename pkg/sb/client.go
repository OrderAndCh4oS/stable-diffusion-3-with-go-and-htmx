package sb

import (
	"github.com/nedpals/supabase-go"
	"os"
)

var Client *supabase.Client

func InitSb() error {
	sbUrl := os.Getenv("SUPABASE_URL")
	sbSecret := os.Getenv("SUPABASE_SECRET")
	Client = supabase.CreateClient(sbUrl, sbSecret)

	return nil
}
