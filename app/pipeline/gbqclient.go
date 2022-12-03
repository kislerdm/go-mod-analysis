package pipeline

import (
	"context"
	"errors"

	"cloud.google.com/go/bigquery"
	"cloud.google.com/go/bigquery/storage/managedwriter"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/descriptorpb"
)

// DataWriter definition of the data to persist.
type DataWriter interface {
	Data() [][]byte
	Descriptor() *descriptorpb.DescriptorProto
}

type DataReader [][]interface{}

func (d DataReader) NRows() int {
	return len(d)
}

func (d DataReader) NCols() int {
	if d.NRows() == 0 {
		return 0
	}
	return len(d[0])
}

// GBQClient IO client to communicate with GBQ.
type GBQClient interface {
	// Read reads the data using the query/path.
	Read(ctx context.Context, query string) (DataReader, error)

	// Write writes data to the specified path in persistence layer.
	Write(ctx context.Context, data DataWriter, path string) error

	Close() error
}

type gbq struct {
	ProjectID string

	w *managedwriter.Client
	r *bigquery.Client
}

func (c gbq) Close() error {
	if err := c.w.Close(); err != nil {
		return err
	}
	return c.r.Close()
}

func (c gbq) Read(ctx context.Context, query string) (DataReader, error) {
	q := c.r.Query(query)

	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		return nil, err
	}
	if err := status.Err(); err != nil {
		return nil, err
	}

	it, err := job.Read(ctx)
	if err != nil {
		return nil, err
	}

	var o DataReader
	for {
		var row []bigquery.Value
		err := it.Next(&row)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var tmp []interface{}
		for _, v := range row {
			tmp = append(tmp, v)
		}
		o = append(o, tmp)
	}

	return o, nil
}

func (c gbq) Write(ctx context.Context, data DataWriter, path string) error {
	managedStream, err := c.w.NewManagedStream(
		ctx,
		managedwriter.WithDestinationTable("projects/"+c.ProjectID+"/"+path),
		managedwriter.WithSchemaDescriptor(data.Descriptor()),
	)
	if err != nil {
		return err
	}

	result, err := managedStream.AppendRows(ctx, data.Data())
	if err != nil {
		return err
	}

	_, err = result.GetResult(ctx)
	return err
}

// NewGBQClient init GBQ client.
func NewGBQClient(ctx context.Context, projectID string) (GBQClient, error) {
	writer, err := managedwriter.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.New("error initialising bigquery client " + err.Error())
	}
	reader, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		return nil, errors.New("bigquery.NewClient: " + err.Error())
	}
	return &gbq{
		ProjectID: projectID,
		r:         reader,
		w:         writer,
	}, nil
}
