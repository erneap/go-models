package svcs

import (
	"context"
	"time"

	"github.com/erneap/go-models/config"
	"github.com/erneap/go-models/general"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddReport(app, rpttype, subtype, mimetype string, body []byte) *general.DBReport {
	now := time.Now().UTC()
	rpt := &general.DBReport{
		ID:            primitive.NewObjectID(),
		ReportDate:    now,
		Application:   app,
		ReportType:    rpttype,
		ReportSubType: subtype,
		MimeType:      mimetype,
	}
	rpt.SetDocument(body)

	rptCol := config.GetCollection(config.DB, "general", "reports")

	rptCol.InsertOne(context.TODO(), rpt)

	return rpt
}

func UpdateReport(id, mimetype string, body []byte) (*general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var rpt *general.DBReport
	err = rptCol.FindOne(context.TODO(), filter).Decode(&rpt)
	if err != nil {
		return nil, err
	}

	rpt.ReportDate = time.Now().UTC()
	rpt.MimeType = mimetype
	rpt.SetDocument(body)
	_, err = rptCol.ReplaceOne(context.TODO(), filter, rpt)
	if err != nil {
		return nil, err
	}
	return rpt, nil
}

func DeleteReport(id string) error {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": oId,
	}

	_, err = rptCol.DeleteOne(context.TODO(), filter)
	return err
}
