package test

import (
	"context"
	"os"
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/openfga/openfga/pkg/logger"
	"github.com/openfga/openfga/pkg/server/commands"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/telemetry"
	"github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/typesystem"
	"github.com/stretchr/testify/require"
	openfgapb "go.buf.build/openfga/go/openfga/api/openfga/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	defaultResolveNodeLimit = 25
	gitHubTestDataFile      = "testdata/github.json" // relative to project root
)

func CheckQueryTest(t *testing.T, datastore storage.OpenFGADatastore) {
	var tests = []struct {
		name             string
		schemaVersion    string
		typeDefinitions  []*openfgapb.TypeDefinition
		tuples           []*openfgapb.TupleKey
		resolveNodeLimit uint32
		request          *openfgapb.CheckRequest
		err              error
		response         *openfgapb.CheckResponse
	}{
		{
			name: "ExecuteReturnsResolutionTooComplexErrorForComplexResolution",
			// state
			typeDefinitions: []*openfgapb.TypeDefinition{{
				Type: "repo",
				Relations: map[string]*openfgapb.Userset{
					"reader": {
						Userset: &openfgapb.Userset_This{},
					},
					"writer": {
						Userset: &openfgapb.Userset_ComputedUserset{
							ComputedUserset: &openfgapb.ObjectRelation{
								Relation: "reader",
							},
						}},
				},
			}},
			resolveNodeLimit: 2,
			// input
			request: &openfgapb.CheckRequest{
				TupleKey: tuple.NewTupleKey("repo:openfga/openfga", "writer", "someUser"),
			},
			// output
			err: serverErrors.AuthorizationModelResolutionTooComplex,
		},
		{
			name:          "ExecuteReturnsResolutionTooComplexErrorForComplexUnionResolution",
			schemaVersion: typesystem.SchemaVersion1_0,
			typeDefinitions: []*openfgapb.TypeDefinition{{
				Type: "repo",
				Relations: map[string]*openfgapb.Userset{
					"writer": {
						Userset: &openfgapb.Userset_This{},
					},
					"reader": {
						Userset: &openfgapb.Userset_Union{
							Union: &openfgapb.Usersets{
								Child: []*openfgapb.Userset{
									{
										Userset: &openfgapb.Userset_This{},
									},
									{
										Userset: &openfgapb.Userset_ComputedUserset{
											ComputedUserset: &openfgapb.ObjectRelation{
												Relation: "writer",
											},
										},
									},
								},
							},
						},
					},
				},
			}},
			resolveNodeLimit: 2,
			// input
			request: &openfgapb.CheckRequest{
				TupleKey: tuple.NewTupleKey("repo:openfga/openfga", "reader", "someUser"),
			},
			// output
			err: serverErrors.AuthorizationModelResolutionTooComplex,
		},
		{
			// NOTE: not a valid test from the API as cannot write the tuples
			//type org
			//	relations
			//		define viewer as self
			//		define can_view as viewer
			//type document
			//	relations
			//		define parent as self
			//		define viewer as viewer from parent
			name:             "Fails_if_expanding_the_computed_userset_of_a_tupleToUserset_rewrite",
			resolveNodeLimit: defaultResolveNodeLimit,
			schemaVersion:    typesystem.SchemaVersion1_0,
			typeDefinitions: []*openfgapb.TypeDefinition{
				{
					Type: "document",
					Relations: map[string]*openfgapb.Userset{
						"parent": typesystem.This(),
						"viewer": typesystem.TupleToUserset("parent", "viewer"),
					},
				},
				{
					Type: "org",
					Relations: map[string]*openfgapb.Userset{
						"viewer":   typesystem.This(),
						"can_view": typesystem.ComputedUserset("viewer"),
					},
				},
			},
			tuples: []*openfgapb.TupleKey{
				tuple.NewTupleKey("org:x", "viewer", "org:y"),
				tuple.NewTupleKey("document:1", "parent", "org:y#can_view"),
				tuple.NewTupleKey("document:1", "parent", "org:z#can_view"), //not relevant
			},
			request: &openfgapb.CheckRequest{
				TupleKey:         tuple.NewTupleKey("document:1", "viewer", "org:y"),
				ContextualTuples: &openfgapb.ContextualTupleKeys{},
			},
			response: &openfgapb.CheckResponse{Allowed: false},
		},
		{
			// NOTE: not a valid test from the API as cannot write the tuples
			//type org
			//	relations
			//		define viewer as self
			//		define can_view as viewer
			//type document
			//	relations
			//		define parent as self
			//		define viewer as viewer from parent
			name:             "Fails_if_expanding_the_computed_userset_of_a_tupleToUserset_rewrite",
			resolveNodeLimit: defaultResolveNodeLimit,
			schemaVersion:    typesystem.SchemaVersion1_0,
			typeDefinitions: []*openfgapb.TypeDefinition{
				{
					Type: "document",
					Relations: map[string]*openfgapb.Userset{
						"parent": typesystem.This(),
						"viewer": typesystem.Union(
							typesystem.This(),
							typesystem.TupleToUserset("parent", "viewer"),
						),
					},
				},
			},
			tuples: []*openfgapb.TupleKey{
				tuple.NewTupleKey("document:1", "parent", "document:2#viewer"),
				tuple.NewTupleKey("document:2", "viewer", "jon"),
			},
			request: &openfgapb.CheckRequest{
				TupleKey:         tuple.NewTupleKey("document:1", "viewer", "org:y"),
				ContextualTuples: &openfgapb.ContextualTupleKeys{},
			},
			response: &openfgapb.CheckResponse{Allowed: false},
		},
	}

	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()
	meter := telemetry.NewNoopMeter()
	logger := logger.NewNoopLogger()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := ulid.Make().String()
			model := &openfgapb.AuthorizationModel{
				Id:              ulid.Make().String(),
				SchemaVersion:   test.schemaVersion,
				TypeDefinitions: test.typeDefinitions,
			}

			err := datastore.WriteAuthorizationModel(ctx, store, model)
			require.NoError(t, err)

			if test.tuples != nil {
				err := datastore.Write(ctx, store, nil, test.tuples)
				require.NoError(t, err)
			}

			cmd := commands.NewCheckQuery(datastore, tracer, meter, logger, test.resolveNodeLimit)
			test.request.StoreId = store
			test.request.AuthorizationModelId = model.Id
			resp, gotErr := cmd.Execute(ctx, test.request)

			require.ErrorIs(t, gotErr, test.err)

			if test.response != nil {
				require.NoError(t, gotErr)

				require.Equal(t, test.response.Allowed, resp.Allowed)

				if test.response.Allowed {
					require.Equal(t, test.response.Resolution, resp.Resolution)
				}
			}
		})
	}
}

// TestCheckQueryAuthorizationModelsVersioning ensures that Check is using the "auth model id" passed in as parameter to expand the usersets
func TestCheckQueryAuthorizationModelsVersioning(t *testing.T, datastore storage.OpenFGADatastore) {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()
	meter := telemetry.NewNoopMeter()
	logger := logger.NewNoopLogger()
	store := ulid.Make().String()

	oldModel := &openfgapb.AuthorizationModel{
		Id:            ulid.Make().String(),
		SchemaVersion: typesystem.SchemaVersion1_0,
		TypeDefinitions: []*openfgapb.TypeDefinition{
			{
				Type: "repo",
				Relations: map[string]*openfgapb.Userset{
					"owner": typesystem.This(),
					"editor": typesystem.Union(
						typesystem.This(),
						typesystem.ComputedUserset("owner"),
					),
				},
			},
		},
	}

	err := datastore.WriteAuthorizationModel(ctx, store, oldModel)
	require.NoError(t, err)

	updatedModel := &openfgapb.AuthorizationModel{
		Id:            ulid.Make().String(),
		SchemaVersion: typesystem.SchemaVersion1_0,
		TypeDefinitions: []*openfgapb.TypeDefinition{
			{
				Type: "repo",
				Relations: map[string]*openfgapb.Userset{
					"owner":  typesystem.This(),
					"editor": typesystem.This(),
				},
			},
		},
	}

	err = datastore.WriteAuthorizationModel(ctx, store, updatedModel)
	require.NoError(t, err)

	err = datastore.Write(ctx, store, []*openfgapb.TupleKey{}, []*openfgapb.TupleKey{{Object: "repo:openfgapb", Relation: "owner", User: "yenkel"}})
	require.NoError(t, err)

	oldResp, err := commands.NewCheckQuery(datastore, tracer, meter, logger, defaultResolveNodeLimit).Execute(ctx, &openfgapb.CheckRequest{
		StoreId:              store,
		AuthorizationModelId: oldModel.Id,
		TupleKey: &openfgapb.TupleKey{
			Object:   "repo:openfgapb",
			Relation: "editor",
			User:     "yenkel",
		},
	})
	require.NoError(t, err)
	require.True(t, oldResp.Allowed)

	updatedResp, err := commands.NewCheckQuery(datastore, tracer, meter, logger, defaultResolveNodeLimit).Execute(ctx, &openfgapb.CheckRequest{
		StoreId:              store,
		AuthorizationModelId: updatedModel.Id,
		TupleKey: &openfgapb.TupleKey{
			Object:   "repo:openfgapb",
			Relation: "editor",
			User:     "yenkel",
		},
	})
	require.NoError(t, err)
	require.False(t, updatedResp.Allowed)
}

var tuples = []*openfgapb.TupleKey{
	tuple.NewTupleKey("repo:openfga/openfga", "reader", "team:openfga#member"),
	tuple.NewTupleKey("team:openfga", "member", "github|iaco@openfga"),
}

// Used to avoid compiler optimizations (see https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
var checkResponse *openfgapb.CheckResponse //nolint

func BenchmarkCheckWithoutTrace(b *testing.B, datastore storage.OpenFGADatastore) {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()
	meter := telemetry.NewNoopMeter()
	logger := logger.NewNoopLogger()
	store := ulid.Make().String()

	data, err := os.ReadFile(gitHubTestDataFile)
	require.NoError(b, err)

	var gitHubTypeDefinitions openfgapb.WriteAuthorizationModelRequest
	err = protojson.Unmarshal(data, &gitHubTypeDefinitions)
	require.NoError(b, err)

	model := &openfgapb.AuthorizationModel{
		Id:              ulid.Make().String(),
		SchemaVersion:   typesystem.SchemaVersion1_0,
		TypeDefinitions: gitHubTypeDefinitions.GetTypeDefinitions(),
	}

	err = datastore.WriteAuthorizationModel(ctx, store, model)
	require.NoError(b, err)

	err = datastore.Write(ctx, store, []*openfgapb.TupleKey{}, tuples)
	require.NoError(b, err)

	checkQuery := commands.NewCheckQuery(datastore, tracer, meter, logger, defaultResolveNodeLimit)

	var r *openfgapb.CheckResponse

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ = checkQuery.Execute(ctx, &openfgapb.CheckRequest{
			StoreId:              store,
			AuthorizationModelId: model.Id,
			TupleKey: &openfgapb.TupleKey{
				Object:   "repo:openfga/openfga",
				Relation: "reader",
				User:     "github|iaco@openfga",
			},
		})
	}

	checkResponse = r
}

func BenchmarkCheckWithTrace(b *testing.B, datastore storage.OpenFGADatastore) {
	ctx := context.Background()
	tracer := telemetry.NewNoopTracer()
	meter := telemetry.NewNoopMeter()
	logger := logger.NewNoopLogger()
	store := ulid.Make().String()

	data, err := os.ReadFile(gitHubTestDataFile)
	require.NoError(b, err)

	var gitHubTypeDefinitions openfgapb.WriteAuthorizationModelRequest
	err = protojson.Unmarshal(data, &gitHubTypeDefinitions)
	require.NoError(b, err)

	model := &openfgapb.AuthorizationModel{
		Id:              ulid.Make().String(),
		SchemaVersion:   typesystem.SchemaVersion1_0,
		TypeDefinitions: gitHubTypeDefinitions.GetTypeDefinitions(),
	}

	err = datastore.WriteAuthorizationModel(ctx, store, model)
	require.NoError(b, err)

	err = datastore.Write(ctx, store, []*openfgapb.TupleKey{}, tuples)
	require.NoError(b, err)

	checkQuery := commands.NewCheckQuery(datastore, tracer, meter, logger, defaultResolveNodeLimit)

	var r *openfgapb.CheckResponse

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, _ = checkQuery.Execute(ctx, &openfgapb.CheckRequest{
			StoreId:              store,
			AuthorizationModelId: model.Id,
			TupleKey: &openfgapb.TupleKey{
				Object:   "repo:openfga/openfga",
				Relation: "reader",
				User:     "github|iaco@openfga",
			},
			Trace: true,
		})
	}

	checkResponse = r
}
