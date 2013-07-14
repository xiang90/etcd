package main

import (
	"encoding/json"
	//"errors"
	"github.com/coreos/etcd/store"
	"github.com/coreos/go-raft"
	"time"
)

// A command represents an action to be taken on the replicated state machine.
type Command interface {
	CommandName() string
	Apply(server *raft.Server) (interface{}, error)
}

// Set command
type SetCommand struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	ExpireTime time.Time `json:"expireTime"`
}

// The name of the set command in the log
func (c *SetCommand) CommandName() string {
	return "set"
}

// Set the key-value pair
func (c *SetCommand) Apply(server *raft.Server) (interface{}, error) {
	return etcdStore.Set(c.Key, c.Value, c.ExpireTime, server.CommitIndex())
}

// TestAndSet command
type TestAndSetCommand struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	PrevValue  string    `json: prevValue`
	ExpireTime time.Time `json:"expireTime"`
}

// The name of the testAndSet command in the log
func (c *TestAndSetCommand) CommandName() string {
	return "testAndSet"
}

// Set the key-value pair if the current value of the key equals to the given prevValue
func (c *TestAndSetCommand) Apply(server *raft.Server) (interface{}, error) {
	return etcdStore.TestAndSet(c.Key, c.PrevValue, c.Value, c.ExpireTime, server.CommitIndex())
}

// Get command
type GetCommand struct {
	Key string `json:"key"`
}

// The name of the get command in the log
func (c *GetCommand) CommandName() string {
	return "get"
}

// Get the value of key
func (c *GetCommand) Apply(server *raft.Server) (interface{}, error) {
	return etcdStore.Get(c.Key)
}

// Delete command
type DeleteCommand struct {
	Key string `json:"key"`
}

// The name of the delete command in the log
func (c *DeleteCommand) CommandName() string {
	return "delete"
}

// Delete the key
func (c *DeleteCommand) Apply(server *raft.Server) (interface{}, error) {
	return etcdStore.Delete(c.Key, server.CommitIndex())
}

// Watch command
type WatchCommand struct {
	Key        string `json:"key"`
	SinceIndex uint64 `json:"sinceIndex"`
}

// The name of the watch command in the log
func (c *WatchCommand) CommandName() string {
	return "watch"
}

func (c *WatchCommand) Apply(server *raft.Server) (interface{}, error) {
	// create a new watcher
	watcher := store.CreateWatcher()

	// add to the watchers list
	etcdStore.AddWatcher(c.Key, watcher, c.SinceIndex)

	// wait for the notification for any changing
	res := <-watcher.C

	return json.Marshal(res)
}

// JoinCommand
type JoinCommand struct {
	Name string `json:"name"`
}

// The name of the join command in the log
func (c *JoinCommand) CommandName() string {
	return "join"
}

// Join a server to the cluster
func (c *JoinCommand) Apply(server *raft.Server) (interface{}, error) {
	err := server.AddPeer(c.Name)

	return []byte("join success"), err
}
