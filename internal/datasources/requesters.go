package datasources

import (
	"context"
	"fmt"

	"github.com/hairyhenderson/gomplate/v3/internal/config"
	"github.com/spf13/afero"
)

var requesters = map[string]requester{}

func init() {
	registerRequesters()
}

// registerRequesters registers the source-reader functions
func registerRequesters() {
	requesters["aws+smp"] = &awsSMPRequester{}
	requesters["aws+sm"] = &awsSecretsManagerRequester{}
	requesters["boltdb"] = &boltDBRequester{}

	c := &consulRequester{}
	requesters["consul"] = c
	requesters["consul+http"] = c
	requesters["consul+https"] = c

	requesters["env"] = &envRequester{}
	requesters["file"] = &fileRequester{afero.NewOsFs()}

	h := &httpRequester{}
	requesters["http"] = h
	requesters["https"] = h

	requesters["merge"] = &mergeRequester{
		ds: map[string]config.DataSource{},
	}

	requesters["stdin"] = &stdinRequester{}

	v := &vaultRequester{}
	requesters["vault"] = v
	requesters["vault+http"] = v
	requesters["vault+https"] = v

	b := &blobRequester{}
	requesters["s3"] = b
	requesters["gs"] = b

	g := &gitRequester{}
	requesters["git"] = g
	requesters["git+file"] = g
	requesters["git+http"] = g
	requesters["git+https"] = g
	requesters["git+ssh"] = g
}

func lookupRequester(ctx context.Context, scheme string) (requester, error) {
	if requester, ok := requesters[scheme]; ok {
		if err := requester.Initialize(ctx); err != nil {
			return nil, err
		}
		return requester, nil
	}
	return nil, fmt.Errorf("no requester found for scheme %s (not registered?)", scheme)
}
