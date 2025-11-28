package milvus

import (
	"context"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type MilvusMigration struct {
	client *milvusclient.Client
}

func NewMilvusMigration(client *milvusclient.Client) *MilvusMigration {
	return &MilvusMigration{
		client: client,
	}
}

func (m *MilvusMigration) Run(ctx context.Context) error {
	// create docs collection
	colName := "documents"
	schema := entity.NewSchema().WithDynamicFieldEnabled(true)

	schema.WithField(entity.NewField().WithName("id").
		WithIsAutoID(false).WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true))

	schema.WithField(
		entity.NewField().
			WithName("vector").
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(768),
	)

	schema.WithField(
		entity.NewField().
			WithName("metadata").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(65535),
	)

	indexOptions := []milvusclient.CreateIndexOption{
		milvusclient.NewCreateIndexOption(colName, "id", index.NewAutoIndex(entity.COSINE)),
		milvusclient.NewCreateIndexOption(colName, "vector", index.NewAutoIndex(entity.COSINE)),
	}

	exist, err := m.client.HasCollection(ctx, milvusclient.NewHasCollectionOption(colName))

	if err != nil {
		return err
	}

	if !exist {
		if err := m.client.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(colName, schema).WithIndexOptions(indexOptions...)); err != nil {
			return err
		}
	}

	return nil
}
