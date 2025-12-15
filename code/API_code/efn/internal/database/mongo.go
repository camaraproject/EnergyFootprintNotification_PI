/*
Copyright (C) 2022-2025 Contributors | TIM S.p.A. to CAMARA a Series of LF Projects, LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package database

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/config"
	servererr "github.com/camaraproject/EnergyFootprintNotification_PI/code/API_code/efn/pkg/server/error"
)

var _ Interface = &mongoDB{}

type mongoDB struct {
	jobs    *mongo.Collection
	jobApps *mongo.Collection
}

// NewMongoDB creates a new MongoDB connection using the provided URI and database name.
func NewMongoDB(conf config.Database) (Interface, error) {
	clientOpts := options.Client().ApplyURI(conf.Uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}
	jobsColl := client.Database(conf.Name).Collection("jobs")
	jobAppsColl := client.Database(conf.Name).Collection("jobAppResults")

	// Create unique compound index on (jobId, appId) to prevent race condition duplicates
	// Use background context with timeout to avoid blocking startup
	ctx := context.Background()
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "jobId", Value: 1}, {Key: "appId", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("jobId_appId_unique"),
	}
	// Try to create index, ignore error if index already exists
	_, err = jobAppsColl.Indexes().CreateOne(ctx, indexModel)
	if err != nil && !mongo.IsDuplicateKeyError(err) {
		// If it's a duplicate key error during index creation, it means there are existing
		// duplicate documents. Log a warning but continue - the upserts will still work,
		// though some race conditions may still occur until duplicates are cleaned up.
		return nil, err
	}

	return &mongoDB{jobs: jobsColl, jobApps: jobAppsColl}, nil
}

func (m *mongoDB) CreateJob(ctx context.Context, r *Job) error {
	_, err := m.jobs.InsertOne(ctx, r)
	return err
}

func (m *mongoDB) GetJob(ctx context.Context, id string) (*Job, error) {
	var job Job
	if err := m.jobs.FindOne(ctx, bson.M{"_id": id}).Decode(&job); err != nil {
		return nil, err
	}
	return &job, nil
}

// TrySetCalculationTriggered performs an atomic set of calculationTriggered=true if not yet true.
// It returns true only if this invocation actually set the flag (i.e., caller is first and may emit the event).
func (m *mongoDB) TrySetCalculationTriggered(ctx context.Context, jobID string) (bool, error) {
	filter := bson.M{"_id": jobID, "calculationTriggered": bson.M{"$ne": true}}
	update := bson.M{"$set": bson.M{"calculationTriggered": true}}
	res, err := m.jobs.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

// TrySetNotificationSent performs an atomic set of notificationSent=true if not yet true.
// It returns true only if this invocation actually set the flag (i.e., caller is first and may send notification).
func (m *mongoDB) TrySetNotificationSent(ctx context.Context, jobID string) (bool, error) {
	filter := bson.M{"_id": jobID, "notificationSent": bson.M{"$ne": true}}
	update := bson.M{"$set": bson.M{"notificationSent": true}}
	res, err := m.jobs.UpdateOne(ctx, filter, update)
	if err != nil {
		return false, err
	}
	return res.MatchedCount == 1, nil
}

func (m *mongoDB) SetJobStatus(ctx context.Context, jobID string, status Status) error {
	res, err := m.jobs.UpdateOne(ctx, bson.M{"_id": jobID}, bson.M{"$set": bson.M{"State": status}})
	if res.MatchedCount == 0 {
		return servererr.NewNotFound("job with id '" + jobID + "'")
	}
	return err
}

// app.Consumption -> result.AppInstanceEnergyConsumption = 0.0.
// ne.COnsumption x ne -> result.networkElement.Add()

func (m *mongoDB) CreateOrUpdateApplicationResult(ctx context.Context, creationMetadata JobAppResultMetadata, appInstanceConsumption float64) error {
	filter := bson.M{
		"jobId": creationMetadata.JobID,
		"appId": creationMetadata.AppID,
	}

	// Use dotted notation to update only the appInstanceEnergyConsumption field
	// without replacing the entire result object
	appConsumptionPath := "result.appInstanceEnergyConsumption"

	update := bson.M{
		// things that should only be set on insert
		"$setOnInsert": bson.M{
			"jobId":            creationMetadata.JobID,
			"appId":            creationMetadata.AppID,
			"numberOfTotalNEs": creationMetadata.NumberOfTotalNEs,
		},
		// things we want to update every time
		"$set": bson.M{
			appConsumptionPath: appInstanceConsumption,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.jobApps.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *mongoDB) CreateOrUpdateNetworkElementResult(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, neResult NetworkElementResult) error {
	filter := bson.M{
		"jobId": creationMetadata.JobID,
		"appId": creationMetadata.AppID,
	}

	// build the dotted path for this NE
	nePath := "result.networkElements." + neInstanceID

	update := bson.M{
		// things that should only be set on insert
		"$setOnInsert": bson.M{
			"jobId":            creationMetadata.JobID,
			"appId":            creationMetadata.AppID,
			"numberOfTotalNEs": creationMetadata.NumberOfTotalNEs,
		},
		// things we want to update every time
		"$set": bson.M{
			nePath: bson.M{
				"energyConsumption":  neResult.EnergyConsumption,
				"appInstanceTraffic": neResult.AppInstanceTraffic,
				"totalTraffic":       neResult.TotalTraffic,
			},
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.jobApps.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *mongoDB) SetNetworkElementEnergy(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, energyConsumption float64) error {
	filter := bson.M{
		"jobId": creationMetadata.JobID,
		"appId": creationMetadata.AppID,
	}

	// build the dotted path for the energy field only
	energyPath := "result.networkElements." + neInstanceID + ".energyConsumption"

	update := bson.M{
		// things that should only be set on insert
		"$setOnInsert": bson.M{
			"jobId":            creationMetadata.JobID,
			"appId":            creationMetadata.AppID,
			"numberOfTotalNEs": creationMetadata.NumberOfTotalNEs,
		},
		// only update the energy field
		"$set": bson.M{
			energyPath: energyConsumption,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.jobApps.UpdateOne(ctx, filter, update, opts)
	return err
}

func (m *mongoDB) SetNetworkElementTraffic(ctx context.Context, creationMetadata JobAppResultMetadata, neInstanceID string, appInstanceTraffic, totalTraffic float64) error {
	filter := bson.M{
		"jobId": creationMetadata.JobID,
		"appId": creationMetadata.AppID,
	}

	// build the dotted paths for traffic volume fields only
	appTrafficPath := "result.networkElements." + neInstanceID + ".appInstanceTraffic"
	totalTrafficPath := "result.networkElements." + neInstanceID + ".totalTraffic"

	update := bson.M{
		// things that should only be set on insert
		"$setOnInsert": bson.M{
			"jobId":            creationMetadata.JobID,
			"appId":            creationMetadata.AppID,
			"numberOfTotalNEs": creationMetadata.NumberOfTotalNEs,
		},
		// only update the traffic volume fields
		"$set": bson.M{
			appTrafficPath:   appInstanceTraffic,
			totalTrafficPath: totalTraffic,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := m.jobApps.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetAllJobAppResults returns all JobAppResults for a specific JobID/requestID.
func (m *mongoDB) GetAllJobAppResults(ctx context.Context, jobID string) ([]JobAppResult, error) {
	var results []JobAppResult
	cursor, err := m.jobApps.Find(ctx, bson.M{"jobId": jobID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result JobAppResult
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

func (m *mongoDB) GetJobAppResult(ctx context.Context, jobID, appID string) (*JobAppResult, error) {
	var result JobAppResult
	err := m.jobApps.FindOne(ctx, bson.M{"jobId": jobID, "appId": appID}).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
