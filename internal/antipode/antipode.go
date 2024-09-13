package antipode

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
)

type Datastore_type interface {
	write(context.Context, string, string, AntiObj) error
	read(context.Context, string, string) (AntiObj, error)
	consume(context.Context, string, string, chan struct{}) (<-chan AntiObj, error)
	barrier(context.Context, []WriteIdentifier, string) error
}

type AntiObj struct {
	Version string
	Lineage []WriteIdentifier
}

type WriteIdentifier struct {
	Dtstid  string
	TableId string
	Key     string
	Version string
}

type contextKey string

func Write(ctx context.Context, datastoreType Datastore_type, datastore_ID string, table string, key string, value string) (context.Context, error) {

	//extract lineage from ctx
	lineage := ctx.Value(contextKey("lineage")).([]WriteIdentifier)

	if lineage == nil {
		err := fmt.Errorf("Lineage not found inside context")
		return ctx, err
	}

	//update lineage
	lineage = append(lineage, WriteIdentifier{Dtstid: datastore_ID, Key: key, Version: value, TableId: table})

	//initialize AntiObj
	obj := AntiObj{value, lineage}

	err := datastoreType.write(ctx, table, key, obj)

	if err != nil {
		return ctx, err
	}

	//update ctx with the updated lineage
	ctx = context.WithValue(ctx, contextKey("lineage"), lineage)

	return ctx, nil
}

func Read(ctx context.Context, datastoreType Datastore_type, table string, key string) (string, []WriteIdentifier, error) {

	obj, err := datastoreType.read(ctx, table, key)

	return obj.Version, obj.Lineage, err
}

func Consume(ctx context.Context, datastoreType Datastore_type, exchange string, key string, stop chan struct{}) (<-chan AntiObj, error) {

	return datastoreType.consume(ctx, exchange, key, stop)
}

func Barrier(ctx context.Context, datastoreType Datastore_type, datastore_ID string) error {
	//extract lineage from ctx
	lineage := ctx.Value(contextKey("lineage")).([]WriteIdentifier)

	if lineage == nil {
		err := fmt.Errorf("Lineage not found inside context")
		return err
	}

	return datastoreType.barrier(ctx, lineage, datastore_ID)
}

func Requeue(ctx context.Context, datastoreType Datastore_type, key string, obj AntiObj) error {
	return datastoreType.write(ctx, "", key, obj)
}

func Transfer(ctx context.Context, lineage []WriteIdentifier) (context.Context, error) {
	//extract lineage from ctx
	oldLineage := ctx.Value(contextKey("lineage")).([]WriteIdentifier)

	if oldLineage == nil {
		err := fmt.Errorf("Lineage not found inside context")
		return ctx, err
	}

	newLineage := append(oldLineage, lineage...)

	ctx = context.WithValue(ctx, contextKey("lineage"), newLineage)

	return ctx, nil
}

func GetLineage(ctx context.Context) ([]WriteIdentifier, error) {
	lineage := ctx.Value(contextKey("lineage"))

	if lineage == nil {
		err := fmt.Errorf("Lineage not found inside context")
		return []WriteIdentifier{}, err
	}

	return lineage.([]WriteIdentifier), nil
}

func InitCtx(ctx context.Context) context.Context {
	var lineage []WriteIdentifier = []WriteIdentifier{}

	ctx = context.WithValue(ctx, contextKey("lineage"), lineage)

	return ctx
}
