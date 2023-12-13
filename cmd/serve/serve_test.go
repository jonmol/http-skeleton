// integration tests for the API server
package serve

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/jonmol/http-skeleton/model"
	"github.com/jonmol/http-skeleton/server/dto"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

type testFunc func(context.Context)

func setupDBHTTP(t *testing.T) func(t *testing.T) {
	t.Helper()
	setDefaults()

	ctx, cancel := context.WithCancel(context.Background())
	if _, err := os.Stat(viper.GetString(FieldDBAddr)); err == nil {
		os.RemoveAll(viper.GetString(FieldDBAddr))
	}

	db := model.NewModel(ctx)
	if err := db.OpenBadger(ctx, viper.GetString(FieldDBAddr)); err != nil {
		t.Fatal("Failed to open badger", err)
	}
	if err := db.EnsureDB(ctx); err != nil {
		t.Fatal("Failed to setup badger", err)
	}
	stop := startAPIHTTP(db)

	return func(t *testing.T) {
		t.Helper()
		sctx, can := context.WithTimeout(ctx, 100*time.Millisecond)
		defer can()
		if err := stop.fun(sctx); err != nil {
			t.Error("Failed to shut down HTTP listener", err)
		}

		if err := db.TearDown(ctx); err != nil {
			t.Error("Failed to tear down the DB", err)
		}

		cancel()
	}
}

func TestIntegrationHello(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tearDown := setupDBHTTP(t)
	defer tearDown(t)

	r := require.New(t)
	client := http.Client{}

	type respType struct {
		Data  dto.OutputHello   `json:"data"`
		Meta  map[string]uint64 `json:"meta"`
		Error map[string]string `json:"error"`
	}

	tests := []struct {
		name   string
		input  string
		code   int
		gVal   uint64
		wVal   uint64
		output respType
	}{
		{name: "first success", input: "test", code: http.StatusOK, gVal: 1, wVal: 1, output: respType{Data: dto.OutputHello{Response: "Why hello there test"}, Meta: map[string]uint64{"total": 1, "thisWord": 1}}},
		{name: "second success", input: "test2", code: http.StatusOK, gVal: 2, wVal: 1, output: respType{Data: dto.OutputHello{Response: "Why hello there test2"}, Meta: map[string]uint64{"total": 2, "thisWord": 1}}},
		{name: "third success", input: "test", code: http.StatusOK, gVal: 3, wVal: 2, output: respType{Data: dto.OutputHello{Response: "Why hello there test"}, Meta: map[string]uint64{"total": 3, "thisWord": 2}}},
		// Sinec this is just an example not everything is tested here. rude, veryRude and no input should also be here
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, buildURL(fmt.Sprintf("hello?input=%s", test.input)), http.NoBody)
			if err != nil {
				t.Fatal("Failed to create a request", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal("Failed test because resp returned errror", err)
			}
			defer resp.Body.Close()

			decoder := json.NewDecoder(resp.Body)
			var respData respType
			if e := decoder.Decode(&respData); e != nil {
				t.Fatal("Failed to decode response", e)
			}

			r.Equal(http.StatusOK, resp.StatusCode)
			r.Equal(test.gVal, respData.Meta["total"])
			r.Equal(test.wVal, respData.Meta["thisWord"])
		})
	}
}

func buildURL(ep string) string {
	return fmt.Sprintf("http://localhost:%d/v1/myService/private/%s", viper.GetInt(FieldPort), ep)
}

func setDefaults() {
	for _, flag := range ConfigStructure.Durations {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range ConfigStructure.Strings {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range ConfigStructure.Ints {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range ConfigStructure.Bools {
		viper.SetDefault(flag.Name, flag.Def)
	}

	for _, flag := range ConfigStructure.StringArrays {
		viper.SetDefault(flag.Name, flag.Def)
	}
	// don't use default for the DB dir so that we can delete it
	viper.SetDefault(FieldDBAddr, path.Join(os.TempDir(), "my-app-integration-test"))
}
