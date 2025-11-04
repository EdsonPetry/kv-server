package kvsrv

import (
	"log"
	"sync"

	"github.com/EdsonPetry/kv-server/labrpc"
	"github.com/EdsonPetry/kv-server/rpc"
	"github.com/EdsonPetry/kv-server/tester"
)

const Debug = false

func DPrintf(format string, a ...any) (n int, err error) {
	if Debug {
		log.Printf(format, a...)
	}
	return n, err
}

type KVServer struct {
	mu sync.Mutex

	// Your definitions here.
}

func MakeKVServer() *KVServer {
	kv := &KVServer{}
	// TODO: To implement.
	return kv
}

// Get returns the value and version for args.Key, if args.Key
// exists. Otherwise, Get returns ErrNoKey.
func (kv *KVServer) Get(args *rpc.GetArgs, reply *rpc.GetReply) {
	// TODO: To implement.
}

// Put updates the value for a key if args.Version matches the version of
// the key on the server. If versions don't match, it returns ErrVersion.
// If the key doesn't exist, Put installs the value if the
// args.Version is 0, and returns ErrNoKey otherwise.
func (kv *KVServer) Put(args *rpc.PutArgs, reply *rpc.PutReply) {
	// TODO: To implement.
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
