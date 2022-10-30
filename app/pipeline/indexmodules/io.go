package indexmodules

import (
	"bytes"
	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/storage/managedwriter"
	"cloud.google.com/go/bigquery/storage/managedwriter/adapt"
	"context"
	"errors"
	app "github.com/kislerdm/gomodanalysis/app/pipeline"
	"github.com/kislerdm/gomodanalysis/app/pipeline/indexmodules/model"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"log"
	"net/http"
	"os"
	"time"
	"unsafe"
)

// CfgWriter configurations for writer client.
type CfgWriter struct {
	projectID string
}

// Writer IO client to persist data.
type Writer interface {
	Store(ctx context.Context, data PersistenceFormat, path string) error
}

type bg struct {
	c *managedwriter.Client

	cfg CfgWriter
}

func (c bg) Store(ctx context.Context, data PersistenceFormat, path string) error {
	managedStream, err := c.c.NewManagedStream(ctx,
		managedwriter.WithDestinationTable("projects/"+c.cfg.projectID+"/"+path),
		managedwriter.WithSchemaDescriptor(data.Descriptor),
	)
	if err != nil {
		return err
	}

	result, err := managedStream.AppendRows(ctx, data.Data)
	if err != nil {
		return err
	}

	_, err = result.GetResult(ctx)
	return err
}

// NewConfigWriter initialises configuration.
func NewConfigWriter() (CfgWriter, error) {
	c := CfgWriter{
		projectID: os.Getenv("PROJECT_ID"),
	}

	if c.projectID == "" {
		return CfgWriter{}, errors.New("env variable PROJECT_ID must be set")
	}

	return c, nil
}

// NewWriter initialise IO client to write output.
func NewWriter(ctx context.Context, cfg CfgWriter) (Writer, error) {
	bgClient, err := managedwriter.NewClient(ctx, cfg.projectID)
	if err != nil {
		return nil, errors.New("error initialising bigquery client " + err.Error())
	}
	return &bg{
		c:   bgClient,
		cfg: cfg,
	}, nil
}

// RawData fetched results.
type RawData []byte

// PersistenceFormat data format to persist.
type PersistenceFormat struct {
	Data       [][]byte
	Descriptor *descriptorpb.DescriptorProto
}

type DataRow struct {
	Path      string `json:"path"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

func (v RawData) Decode() ([]DataRow, error) {
	// timestamp data length
	if len(v) < 30 {
		return nil, errors.New("faulty input: byte array is too short")
	}

	if v[0] != '{' {
		return nil, errors.New("faulty input: not a valid JSON")
	}

	var o []DataRow
	var temp []byte
	for i, b := range v {
		temp = append(temp, b)
		if b == '\n' || i == len(v)-1 {
			r, err := decodeRow(temp)
			if err != nil {
				return nil, err
			}
			o = append(o, r)
			temp = nil
			continue
		}
	}

	return o, nil
}

func decodeRow(vals []byte) (DataRow, error) {
	if vals[len(vals)-1] == '\n' {
		vals = vals[:len(vals)-1]
	}

	o := DataRow{}

	if vals[0] != '{' || vals[len(vals)-1] != '}' {
		return DataRow{}, errors.New("faulty input: not a valid JSON")
	}

	for i, b := range vals {
		switch b {
		case 'P':
			o.Path = extractStringVal(vals[i:], "Path")
		case 'V':
			o.Version = extractStringVal(vals[i:], "Version")
		case 'T':
			if vals[i+1] != 'i' {
				break
			}
			o.Timestamp = extractStringVal(vals[i:], "Timestamp")
		}
	}

	return o, nil
}

func extractStringVal(vals []byte, key string) string {
	l := len(key)
	if flag := vals[:l]; *(*string)(unsafe.Pointer(&flag)) == key {
		return extractStringBetweenDoubleQuotes(vals[l+2:])
	}
	return ""
}

func extractStringBetweenDoubleQuotes(v []byte) string {
	var temp []byte
	for i, el := range v {
		if el == '"' {
			if i == len(v)-1 || v[i+1] == ',' || v[i+1] == '}' {
				break
			}
			continue
		}
		temp = append(temp, el)
	}
	return *(*string)(unsafe.Pointer(&temp))
}

func convertToGBQTableFormat(v []DataRow) ([]*model.Index, error) {
	o := make([]*model.Index, len(v))
	for i, b := range v {
		ts, err := time.Parse(time.RFC3339Nano, b.Timestamp)
		if err != nil {
			return nil, errors.New("corrupt timestamp: " + err.Error())
		}
		o[i] = &model.Index{
			Path:      b.Path,
			Version:   b.Version,
			Timestamp: ts.UnixMicro(),
		}
	}
	return o, nil
}

func ConvertToStoreFormat(v []DataRow) (PersistenceFormat, error) {
	m := &model.Index{}
	descriptorProto, err := adapt.NormalizeDescriptor(m.ProtoReflect().Descriptor())
	if err != nil {
		return PersistenceFormat{}, err
	}

	rows, err := convertToGBQTableFormat(v)
	if err != nil {
		return PersistenceFormat{}, err
	}

	data := make([][]byte, len(rows))
	for k, v := range rows {
		b, err := proto.Marshal(v)
		if err != nil {
			return PersistenceFormat{}, err
		}
		data[k] = b
	}

	return PersistenceFormat{
		Data:       data,
		Descriptor: descriptorProto,
	}, err
}

type Reader interface {
	Fetch(query map[string]string) (RawData, error)
}

// CfgReader configurations for reader client.
type CfgReader struct {
	HttpClient *http.Client
	Backoff    *app.Backoff
	Verbose    bool
}

type clientReader struct {
	Cfg CfgReader
}

func (c clientReader) Fetch(query map[string]string) (RawData, error) {
	const baseURI = "https://index.golang.org/index?limit=2000"

	d, err := c.Cfg.Backoff.LinearDelay()
	if err != nil {
		return nil, err
	}

	if c.Cfg.Verbose && d != 0 {
		log.Printf("delay %v sec. and call", d.Seconds())
	}

	time.Sleep(d)

	url := baseURI
	for k, v := range query {
		url += "&" + k + "=" + v
	}

	resp, err := c.Cfg.HttpClient.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 209 {
		c.Cfg.Backoff.UpCounter()
		return c.Fetch(query)
	}

	c.Cfg.Backoff.Reset()
	if resp.ContentLength == 0 {
		return nil, nil
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func NewReader(cfg ...CfgReader) Reader {
	var c CfgReader
	if len(cfg) == 0 {
		c = CfgReader{}
	} else {
		c = cfg[0]
	}

	if c.HttpClient == nil {
		c.HttpClient = &http.Client{Timeout: 10 * time.Second}
	}

	if c.Backoff == nil {
		c.Backoff = &app.Backoff{MaxSteps: 5, MaxDelay: 10 * time.Second}
	}

	return &clientReader{c}
}

// GetLastPaginationIndex returns the last fetched page.
func GetLastPaginationIndex() string {
	const q = `SELECT FORMAT_TIMESTAMP('%FT%R:%E*SZ', MAX(timestamp), 'UTC') AS last_ts FROM ` +
		"`go-mod-analysis.raw.index`;"

	ctx := context.Background()
	c, err := bigquery.NewClient(ctx, os.Getenv("PROJECT_ID"))
	defer func() {
		_ = c.Close()
	}()

	if err != nil {
		return ""
	}

	it, err := c.Query(q).Read(ctx)
	if err != nil {
		return ""
	}

	var v []bigquery.Value
	err = it.Next(&v)
	if err != nil {
		return ""
	}
	if o, ok := v[0].(string); ok {
		return o
	}

	return ""
}
