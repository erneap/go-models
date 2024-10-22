package general

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	b64 "encoding/base64"
)

type ReportType struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Application string             `json:"application" bson:"application"`
	ReportType  string             `json:"reporttype" bson:"reporttype"`
	SubTypes    []string           `json:"subtypes,omitempty" bson:"subtypes,omitempty"`
}

type ByReportTypes []ReportType

func (c ByReportTypes) Len() int { return len(c) }
func (c ByReportTypes) Less(i, j int) bool {
	if c[i].Application == c[j].Application {
		return c[i].ReportType < c[j].ReportType
	}
	return c[i].Application < c[j].Application
}
func (c ByReportTypes) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type DBReport struct {
	ID            primitive.ObjectID `json:"id" bson:"_id"`
	ReportDate    time.Time          `json:"reportdate" bson:"reportdate"`
	ReportTypeID  primitive.ObjectID `json:"reporttypeid" bson:"reporttypeid"`
	ReportSubType string             `json:"subtype,omitempty" bson:"subtype,omitempty"`
	MimeType      string             `json:"mimetype" bson:"mimetype"`
	DocumentBody  string             `json:"docbody" bson:"docbody"`
}

type ByDBReports []DBReport

func (c ByDBReports) Len() int { return len(c) }
func (c ByDBReports) Less(i, j int) bool {
	if c[i].ReportTypeID == c[j].ReportTypeID {
		if c[i].ReportDate.Equal(c[j].ReportDate) {
			if c[i].ReportSubType != "" && c[j].ReportSubType != "" {
				return c[i].ReportSubType < c[j].ReportSubType
			}
		}
		return c[i].ReportDate.Before(c[j].ReportDate)
	}
	return c[i].ReportTypeID.Hex() < c[j].ReportTypeID.Hex()
}
func (c ByDBReports) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (r *DBReport) SetDocument(data []byte) {
	enc := b64.StdEncoding.EncodeToString(data)
	r.DocumentBody = enc
}

func (r DBReport) GetDocument() ([]byte, error) {
	uDec, err := b64.StdEncoding.DecodeString(r.DocumentBody)
	if err != nil {
		return nil, err
	}
	return uDec, nil
}
