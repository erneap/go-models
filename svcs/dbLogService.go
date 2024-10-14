package svcs

import (
	"context"
	"time"

	"github.com/erneap/go-models/config"
	"github.com/erneap/go-models/general"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CRUD Methods for this data collection
func CreateDBLogEntry(app, cat, title, name, msg string) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	// new log entry
	entry := &general.LogEntry{
		ID:          primitive.NewObjectID(),
		EntryDate:   time.Now().UTC(),
		Application: app,
		Category:    cat,
		Title:       title,
		Name:        name,
		Message:     msg,
	}

	_, err := logCol.InsertOne(context.TODO(), entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func UpdateDBLogEntry(id, cat, title, name, msg string) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var entry general.LogEntry
	err = logCol.FindOne(context.TODO(), filter).Decode(&entry)
	if err != nil {
		return nil, err
	}

	if cat != "" {
		entry.Category = cat
	}
	if title != "" {
		entry.Title = title
	}
	if name != "" {
		entry.Name = name
	}
	entry.Message = msg

	_, err = logCol.ReplaceOne(context.TODO(), filter, &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func DeleteDBLogEntry(id string) error {
	logCol := config.GetCollection(config.DB, "general", "logs")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": oId,
	}

	_, err = logCol.DeleteOne(context.TODO(), filter)
	return err
}
