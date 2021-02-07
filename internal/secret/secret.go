package secret

import (
	"fmt"
	"os"
)

type Secrets struct {
	BBCURL string
	SkyURL string
}

func (s *Secrets) Load() error {
	for k, v := range map[string]*string{
		"BBC_URL": &s.BBCURL,
		"SKY_URL": &s.SkyURL,
	} {
		var ok bool
		if *v, ok = os.LookupEnv(k); !ok {
			return fmt.Errorf("missing env variable: %s", k)
		}
	}

	return nil
}
