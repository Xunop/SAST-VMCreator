package main

import "sync"

type TopicInfo struct {
	UserID   string
	RootID   string
	ParentID string
	ThreadID string
}

// Global map to store ongoing `/create_vm` topics
var activeTopics sync.Map
