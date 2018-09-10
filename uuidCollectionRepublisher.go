package main

import (
	"fmt"
	"time"
)

type uuidCollectionRepublisher interface {
	RepublishUUIDFromCollection(uuid, tid string, collection targetSystem) (msg *okMsg, wasFound bool, err error)
}

type notifyingUCRepublisher struct {
	notifierClient    notifierClient
	nativeStoreClient nativeStoreClientInterface
	rateLimit         time.Duration
}

func newNotifyingUCRepublisher(notifierClient notifierClient, nativeStoreClient nativeStoreClientInterface, rateLimit time.Duration) *notifyingUCRepublisher {
	return &notifyingUCRepublisher{notifierClient, nativeStoreClient, rateLimit}
}

type okMsg struct {
	uuid                     string
	tid                      string
	collectionName           string
	collectionOriginSystemID string
	sizeBytes                int
	notifierAppName          string
	contentType              string
}

func (msg *okMsg) String() string {
	return fmt.Sprintf("sent for publish uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v contentType=%v", msg.uuid, msg.tid, msg.collectionName, msg.collectionOriginSystemID, msg.sizeBytes, msg.notifierAppName, msg.contentType)
}

func (r *notifyingUCRepublisher) RepublishUUIDFromCollection(uuid, tid string, collection targetSystem) (msg *okMsg, wasFound bool, err error) {
	start := time.Now()
	nativeContent, isFound, err := r.nativeStoreClient.GetNative(collection.name, uuid, tid)
	if err != nil {
		return nil, false, fmt.Errorf("error while fetching native content: %v", err)
	}
	if !isFound {
		return nil, false, nil
	}

	err = r.notifierClient.Notify(nativeContent, collection.notifierApp, collection.originSystemID, collection.contentType, uuid, tid)
	if err != nil {
		extendTimeToLength(start, r.rateLimit)
		return nil, true, fmt.Errorf("couldn't send to notifier uuid=%v tid=%v collection=%v originSystemId=%v size=%vB notifierApp=%v %v", uuid, tid, collection.name, collection.originSystemID, len(nativeContent), collection.notifierApp, err)
	}

	extendTimeToLength(start, r.rateLimit)
	return &okMsg{uuid, tid, collection.name, collection.originSystemID, len(nativeContent), collection.notifierApp, collection.contentType}, true, nil
}

func extendTimeToLength(start time.Time, length time.Duration) {
	time.Sleep(time.Duration(start.Add(length).UnixNano()-time.Now().UnixNano()) * time.Nanosecond)
}
