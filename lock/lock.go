// Package lock ...
package lock

import (
	"time"

	"github.com/EdsonPetry/kv-server/kvtest"
	"github.com/EdsonPetry/kv-server/rpc"
)

type Lock struct {
	// IKVClerk is a go interface for k/v clerks: the interface hides
	// the specific Clerk type of ck but promises that ck supports
	// Put and Get. The tester passes the clerk in when calling
	// MakeLock().
	ck       kvtest.IKVClerk
	name     string
	clientID string
}

// MakeLock is used by the tester passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
// Use l as the key to store the "lock state" (you would have to decide
// precisely what the lock state is).
func MakeLock(ck kvtest.IKVClerk, l string) *Lock {
	lk := &Lock{ck: ck, name: l, clientID: kvtest.RandValue(8)}
	lk.ck.Put(l, "", 0)
	return lk
}

func (lk *Lock) Acquire() {
	for {
		// get current lock state from server
		val, ver, err := lk.ck.Get(lk.name)

		// create the lock on the server
		if err == rpc.ErrNoKey {
			err := lk.ck.Put(lk.name, lk.clientID, 0)
			// another client created the lock
			if err == rpc.ErrVersion {

				time.Sleep(10 * time.Millisecond)
				continue // retry again
			}

			if err == rpc.OK {
				return
			}
		}

		// if lock is unlocked, try to claim it
		if val == "" {
			err := lk.ck.Put(lk.name, lk.clientID, ver)
			// if another client claimed it first, retry
			if err == rpc.ErrVersion {
				time.Sleep(10 * time.Millisecond)
				continue // retry again
			}

			// if claimed successfully, done
			if err == rpc.OK {
				return
			}
		} else {
			time.Sleep(10 * time.Millisecond)
			continue // retry again
		}
	}
}

func (lk *Lock) Release() {
	val, ver, err := lk.ck.Get(lk.name)
	if err == rpc.ErrNoKey {
		return // ignore for now
	}

	if val == lk.clientID {
		lk.ck.Put(lk.name, "", ver)
		// NOTE: decided not to handle ErrVersion edge case in case of concurrent
		// releases leading to different versions on the same client.
	}
}
