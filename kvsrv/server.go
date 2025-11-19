package kvsrv

import (
	"log"
	"sync"

	"6.5840/labrpc"
	tester "6.5840/tester1"
	"github.com/EdsonPetry/kv-server/rpc"
)

const Debug = false

func DPrintf(format string, a ...any) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return n, err
}

type Entry struct {
	Value   string
	Version rpc.Tversion
}

type KVServer struct {
	mu  sync.Mutex
	kvs map[string]Entry
}

func MakeKVServer() *KVServer {
	kv := &KVServer{kvs: make(map[string]Entry)}
	return kv
}

// Get returns the value and version for args.Key, if args.Key
// exists. Otherwise, Get returns ErrNoKey.
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.kvs[args.Key]
	if !ok {
		reply.Err = rpc.ErrNoKey
		return
	}

	reply.Value = entry.Value
	reply.Version = entry.Version
	reply.Err = rpc.OK
}

// Put updates the value for a key if args.Version matches the version of
// the key on the server. If versions don't match, it returns ErrVersion.
// If the key doesn't exist, Put installs the value if the
// args.Version is 0, and returns ErrNoKey otherwise.
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	entry, ok := kv.kvs[args.Key]

	// key does not exist
	if !ok {
		if args.Version == 0 {
			newEntry := Entry{args.Value, args.Version + 1}
			kv.kvs[args.Key] = newEntry
			reply.Err = rpc.OK
		} else {
			reply.Err = rpc.ErrNoKey
		}
		return
	}

	// key exists, version mismatch
	if args.Version != entry.Version {
		reply.Err = rpc.ErrVersion
		return
	}

	// key exists, version match -> update value and increment version
	kv.kvs[args.Key] = Entry{args.Value, args.Version + 1}
	reply.Err = rpc.OK
}

func (kv *KVServer) Kill() {
	// Lab instructions say we can ignore Kill() for this lab
	// may be implemented in a future lab?
}

// NOTE: can ignore all arguments; they are for replicated KVservers

func StartKVServer(ends []*labrpc.ClientEnd, gid tester.Tgid, srv int, persister *tester.Persister) []tester.IService {
	kv := MakeKVServer()
	return []tester.IService{kv}
}
