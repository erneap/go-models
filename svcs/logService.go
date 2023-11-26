package svcs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/go-models/config"
	"github.com/erneap/go-models/employees"
	"github.com/erneap/go-models/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Crud Functions for Creating, Retrieving, updating and deleting authentication
// log database records

// CRUD Create Function
func CreateLogEntry(dt time.Time, app string, lvl logs.DebugLevel, msg string) error {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	// new log entry
	entry := &logs.LogEntry{
		ID:          primitive.NewObjectID(),
		DateTime:    dt,
		Application: app,
		Level:       lvl,
		Message:     msg,
	}

	_, err := logCol.InsertOne(context.TODO(), entry)
	return err
}

// CRUD Retrieve Functions - one, between dates for application, by application,
// and all records.
func GetLogEntry(id string) (*logs.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	logid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": logid,
	}

	var entry logs.LogEntry
	if err = logCol.FindOne(context.TODO(), filter).Decode(&entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func GetLogEntriesByApplication(app string) ([]logs.LogEntry, error) {
	var entries []logs.LogEntry

	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{
		"application": app,
	}

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return entries, err
	}

	if err = cursor.All(context.TODO(), &entries); err != nil {
		return entries, err
	}

	sort.Sort(logs.ByLogEntry(entries))
	return entries, nil
}

func GetLogEntriesByApplicationAndDates(app string, begin, end time.Time) ([]logs.LogEntry, error) {
	var entries []logs.LogEntry

	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{
		"application": app,
		"datetime":    bson.M{"$gte": begin, "$lt": end},
	}

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return entries, err
	}

	if err = cursor.All(context.TODO(), &entries); err != nil {
		return entries, err
	}

	sort.Sort(logs.ByLogEntry(entries))
	return entries, nil
}

func GetLogEntries(app string, begin, end time.Time) ([]logs.LogEntry, error) {
	var entries []logs.LogEntry

	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{}

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return entries, err
	}

	if err = cursor.All(context.TODO(), &entries); err != nil {
		return entries, err
	}

	sort.Sort(logs.ByLogEntry(entries))
	return entries, nil
}

// CRUD Update
func UpdateLogEntry(entry logs.LogEntry) error {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{
		"_id": entry.ID,
	}

	_, err := logCol.ReplaceOne(context.TODO(), filter, entry)
	return err
}

// CRUD Delete functions - delete one by id, delete before date, delete by
// application before date
func DeleteLogEntry(id string) error {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	logid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": logid,
	}

	_, err = logCol.DeleteOne(context.TODO(), filter)
	return err
}

func DeleteLogEntriesBeforeDate(dt time.Time) error {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{
		"datetime": bson.M{"lt": dt},
	}

	_, err := logCol.DeleteMany(context.TODO(), filter)
	return err
}

func DeleteLogEntriesByApplicationBeforeDate(app string, dt time.Time) error {
	logCol := config.GetCollection(config.DB, "authenticate", "logs")

	filter := bson.M{
		"application": app,
		"datetime":    bson.M{"lt": dt},
	}

	_, err := logCol.DeleteMany(context.TODO(), filter)
	return err
}

// miscellanous functions for log entry work
func AddLogEntry(app string, lvl logs.DebugLevel, msg string) {
	if config.LogLevel >= int(lvl) {
		CreateLogEntry(time.Now().UTC(), app, lvl, msg)
	}
}

func AddLogEntry2(portion, category, title, msg string, emp *employees.Employee) error {
	name := ""
	site := "General"
	if emp != nil {
		name = emp.Name.GetLastFirst()
		if !strings.EqualFold(portion, "authenticate") {
			site = emp.SiteID
		}
	}
	logBase := os.Getenv("LOG_DIR")
	chgDate := time.Now()
	logLevel, _ := strconv.Atoi(os.Getenv("LOGLEVEL"))

	if logLevel < 1 || !strings.EqualFold(category, "debug") {
		logPath := path.Join(logBase, site, portion)
		if err := os.MkdirAll(logPath, 0755); err != nil {
			return err
		}

		logPath = path.Join(logPath, fmt.Sprintf("%s-%d.log", portion,
			chgDate.Year()))

		logEntry := &logs.LogEntry2{
			EntryDate: chgDate,
			Category:  category,
			Title:     title,
			Message:   msg,
			Name:      name,
		}

		entry := fmt.Sprintf("%s\n", logEntry.ToString())

		f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := f.WriteString(entry); err != nil {
			return err
		}
	}
	return nil
}

func GetLogEntries2(portion string, year int, emp *employees.Employee) ([]logs.LogEntry2, error) {
	site := "General"
	if emp != nil && !strings.EqualFold(portion, "authenticate") {
		site = emp.SiteID
	}
	logBase := os.Getenv("LOG_DIR")
	if strings.TrimSpace(site) == "" {
		site = "General"
	}

	logPath := path.Join(logBase, site, portion)
	if err := os.MkdirAll(logPath, 0755); err != nil {
		return nil, err
	}

	logPath = path.Join(logPath, fmt.Sprintf("%s-%d.log", portion, year))

	if _, err := os.Stat(logPath); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%s does not exist", logPath)
	}

	content, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var entries []logs.LogEntry2

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			entry := &logs.LogEntry2{}
			entry.FromString(line)
			entries = append(entries, *entry)
		}
	}

	return entries, nil
}
