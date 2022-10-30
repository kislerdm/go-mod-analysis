package moduleversions

import (
	"bytes"
	"encoding/json"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"golang.org/x/mod/semver"
	"net/http"
)

// Versions defines the struct for min and max versions.
type Versions struct {
	Min string `json:"version_min"`
	Max string `json:"version_max"`
}

// MinMaxVersions defines min and max semver of the package given array of its versions.
func MinMaxVersions(v []string) Versions {
	if len(v) == 0 {
		return Versions{}
	}

	semver.Sort(v)

	return Versions{
		Min: v[0],
		Max: v[len(v)-1],
	}
}

func newError(msg string) []byte {
	o, _ := json.Marshal(map[string]string{"errorMessage": msg})
	return o
}

type input struct {
	Calls [][]string `json:"calls"`
}

type output struct {
	Replies []Versions `json:"replies"`
}

func handler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" {
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(req.Body); err != nil {
			panic(err)
		}

		var in input
		if err := json.Unmarshal(buf.Bytes(), &in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write(newError("unsupported in: corrupt JSON"))
			return
		}

		o := output{Replies: make([]Versions, len(in.Calls))}

		for i, el := range in.Calls {
			o.Replies[i] = MinMaxVersions(el)
		}

		r, err := json.Marshal(o)
		if err != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(r)
	}
}

func init() {
	functions.HTTP("moduleversions", handler)
}
