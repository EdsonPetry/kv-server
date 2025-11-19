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
	ck   kvtest.IKVClerk
	name string
}

// MakeLock is used by the tester passes in a k/v clerk; your code can
// perform a Put or Get by calling lk.ck.Put() or lk.ck.Get().
// Use l as the key to store the "lock state" (you would have to decide
// precisely what the lock state is).
func MakeLock(ck kvtest.IKVClerk, l string) *Lock {
	lk := &Lock{ck: ck, name: l}
	lk.ck.Put(l, "", 0)
	return lk
}

func (lk *Lock) Acquire() {
	for {
		// get current lock state from server
		val, ver, err := lk.ck.Get(lk.name)

		// create the lock on the server
		if err == rpc.ErrNoKey {
			err := lk.ck.Put(lk.name, "locked", 0)
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
			err := lk.ck.Put(lk.name, "locked", ver)
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

// Release currently assumes all clients behave correctly (do not call Release unless they have acquired a lock)
func (lk *Lock) Release() {
	_, ver, _ := lk.ck.Get(lk.name)
	lk.ck.Put(lk.name, "", ver)
}
