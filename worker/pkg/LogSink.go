package pkg

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
	"unsafe"
)

type LogSink struct {
	client        *mongo.Client
	logCollection *mongo.Collection
	logChan       chan *JobLog
	commitChan    chan []interface{}
}

func InitLogSink() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(G_config.MongodbConfig.ConnectTimeout)*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(G_config.MongodbConfig.Uri))
	if err != nil {
		return err
	}

	G_logSink = &LogSink{
		client:        client,
		logCollection: client.Database(G_config.MongodbConfig.DB).Collection(G_config.MongodbConfig.Collection),
		logChan:       make(chan *JobLog, 1000),
		commitChan:    make(chan []interface{}, 1000),
	}

	return nil
}

func (l *LogSink) WriteLoop() {
	logList := make([]interface{}, 0)
	var timer *time.Timer

	for {
		select {
		case log := <-l.logChan:
			if len(logList) == 0 {
				timer = time.AfterFunc(1*time.Second, func() {
					l.commitChan <- logList
				})
			}

			logList = append(logList, log)

			if len(logList) > 100 {
				l.saveLogs(logList)
				logList = make([]interface{}, 0)
				if timer != nil {
					timer.Stop()
				}
			}
		case list := <-l.commitChan:
			if (*reflect.SliceHeader)(unsafe.Pointer(&list)).Data != (*reflect.SliceHeader)(unsafe.Pointer(&logList)).Data {
				continue
			}
			l.saveLogs(logList)
			logList = make([]interface{}, 0)
		}
	}
}

func (l *LogSink) saveLogs(batch []interface{}) {
	_, _ = l.logCollection.InsertMany(context.Background(), batch)
}

func (l *LogSink) Append(jobLog *JobLog) {
	select {
	case l.logChan <- jobLog:
	default:
	}
}
