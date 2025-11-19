package kvsrv

import (
	"testing"

	tester "6.5840/tester1"
	"github.com/EdsonPetry/kv-server/kvtest"
)

type TestKV struct {
	*kvtest.Test
	t        *testing.T
	reliable bool
}

func MakeTestKV(t *testing.T, reliable bool) *TestKV {
	// Set the visualization model for the tester framework
	// tester.SetVisualizationModel(models.KvModel)

	cfg := tester.MakeConfig(t, 1, reliable, StartKVServer)
	ts := &TestKV{
		t:        t,
		reliable: reliable,
	}
	ts.Test = kvtest.MakeTest(t, cfg, false, ts)
	return ts
}

func (ts *TestKV) MakeClerk() kvtest.IKVClerk {
	clnt := ts.MakeClient()
	ck := MakeClerk(clnt, tester.ServerName(tester.GRP0, 0))
	return &kvtest.TestClerk{IKVClerk: ck, Clnt: clnt}
}

func (ts *TestKV) DeleteClerk(ck kvtest.IKVClerk) {
	tck := ck.(*kvtest.TestClerk)
	ts.DeleteClient(tck.Clnt)
}
